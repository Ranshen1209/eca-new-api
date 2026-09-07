package controller

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/i18n"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/middleware"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/relay"
	"github.com/QuantumNous/new-api/relay/channel/openai"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/relay/helper"
	"github.com/QuantumNous/new-api/relaykit/types"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/setting/ratio_setting"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var responsesWebsocketUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func RelayResponsesWebsocket(c *gin.Context) {
	requestId := c.GetString(common.RequestIdKey)
	ws, err := responsesWebsocketUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": types.OpenAIError{
				Message: "WebSocket upgrade failed",
				Type:    "invalid_request_error",
			},
		})
		return
	}
	clientConn := helper.NewResponsesWebsocketConn(ws)
	c.Set(helper.ContextKeyResponsesWebsocketClient, clientConn)
	defer clientConn.Close(websocket.CloseNormalClosure, "")

	var newAPIError *types.NewAPIError
	defer func() {
		if newAPIError != nil {
			logger.LogError(c, fmt.Sprintf("relay error: %s", common.LocalLogPreview(newAPIError.Error())))
			newAPIError.SetMessage(common.MessageWithRequestId(newAPIError.Error(), requestId))
			helper.WriteResponsesWebsocketError(clientConn, newAPIError.ToOpenAIError())
			clientConn.Close(websocket.CloseInternalServerErr, "")
		}
	}()

	_, firstMessage, err := clientConn.Read()
	if err != nil {
		newAPIError = types.NewError(fmt.Errorf("failed to read websocket request"), types.ErrorCodeInvalidRequest)
		return
	}

	modelName := openai.ResponsesWebsocketRequestModel(firstMessage)
	if modelName == "" {
		modelName = strings.TrimSpace(c.Query("model"))
	}
	if modelName == "" {
		newAPIError = types.NewError(errors.New("model is required"), types.ErrorCodeInvalidRequest)
		return
	}

	if limitErr := enforceResponsesWebsocketModelAccess(c, modelName); limitErr != nil {
		newAPIError = limitErr
		return
	}

	common.SetContextKey(c, constant.ContextKeyOriginalModel, modelName)
	common.SetContextKey(c, constant.ContextKeyRequestStartTime, time.Now())

	channel, channelErr := selectResponsesWebsocketChannel(c, modelName)
	if channelErr != nil {
		newAPIError = channelErr
		return
	}
	if setupErr := middleware.SetupContextForSelectedChannel(c, channel, modelName); setupErr != nil {
		newAPIError = setupErr
		return
	}
	addUsedChannel(c, channel.Id)

	relayInfo := relaycommon.GenRelayInfoResponsesWebsocket(c, ws)
	meta := fastTokenCountMetaForPricing(relayInfo.Request)
	tokens, err := service.EstimateRequestToken(c, meta, relayInfo)
	if err != nil {
		newAPIError = types.NewError(err, types.ErrorCodeCountTokenFailed)
		return
	}
	relayInfo.SetEstimatePromptTokens(tokens)

	priceData, err := helper.ModelPriceHelper(c, relayInfo, tokens, meta)
	if err != nil {
		newAPIError = types.NewError(err, types.ErrorCodeModelPriceError, types.ErrOptionWithStatusCode(http.StatusBadRequest))
		return
	}
	if priceData.FreeModel {
		logger.LogInfo(c, fmt.Sprintf("模型 %s 免费，跳过预扣费", relayInfo.OriginModelName))
	} else {
		newAPIError = service.PreConsumeBilling(c, priceData.QuotaToPreConsume, relayInfo)
		if newAPIError != nil {
			return
		}
	}

	defer func() {
		if newAPIError != nil {
			newAPIError = service.NormalizeViolationFeeError(newAPIError)
			if relayInfo.Billing != nil {
				relayInfo.Billing.Refund(c)
			}
			service.ChargeViolationFeeIfNeeded(c, relayInfo, newAPIError)
		}
	}()

	if billingErr := service.PrepareTieredBillingForSelectedGroup(c, relayInfo); billingErr != nil {
		newAPIError = billingErr
		return
	}

	newAPIError = relay.ResponsesWebsocketHelper(c, relayInfo, firstMessage)
	if newAPIError == nil {
		service.RecordChannelAffinity(c, channel.Id)
	}
}

func enforceResponsesWebsocketModelAccess(c *gin.Context, modelName string) *types.NewAPIError {
	if !common.GetContextKeyBool(c, constant.ContextKeyTokenModelLimitEnabled) {
		return nil
	}
	raw, ok := common.GetContextKey(c, constant.ContextKeyTokenModelLimit)
	if !ok {
		return types.NewError(errors.New(i18n.T(c, i18n.MsgDistributorTokenNoModelAccess)), types.ErrorCodeChannelModelMappedError, types.ErrOptionWithStatusCode(http.StatusForbidden))
	}
	tokenModelLimit, ok := raw.(map[string]bool)
	if !ok {
		tokenModelLimit = map[string]bool{}
	}
	matchName := ratio_setting.FormatMatchingModelName(modelName)
	if _, allowed := tokenModelLimit[matchName]; !allowed {
		return types.NewError(fmt.Errorf("%s", i18n.T(c, i18n.MsgDistributorTokenModelForbidden, map[string]any{"Model": modelName})), types.ErrorCodeChannelModelMappedError, types.ErrOptionWithStatusCode(http.StatusForbidden))
	}
	return nil
}

func selectResponsesWebsocketChannel(c *gin.Context, modelName string) (*model.Channel, *types.NewAPIError) {
	if channelId, ok := common.GetContextKey(c, constant.ContextKeyTokenSpecificChannelId); ok {
		id, err := strconv.Atoi(channelId.(string))
		if err != nil {
			return nil, types.NewError(errors.New(i18n.T(c, i18n.MsgDistributorInvalidChannelId)), types.ErrorCodeGetChannelFailed, types.ErrOptionWithStatusCode(http.StatusBadRequest))
		}
		channel, err := model.GetChannelById(id, true)
		if err != nil {
			return nil, types.NewError(errors.New(i18n.T(c, i18n.MsgDistributorInvalidChannelId)), types.ErrorCodeGetChannelFailed, types.ErrOptionWithStatusCode(http.StatusBadRequest))
		}
		if channel.Status != common.ChannelStatusEnabled {
			return nil, types.NewError(errors.New(i18n.T(c, i18n.MsgDistributorChannelDisabled)), types.ErrorCodeGetChannelFailed, types.ErrOptionWithStatusCode(http.StatusForbidden))
		}
		return channel, nil
	}

	usingGroup := common.GetContextKeyString(c, constant.ContextKeyUsingGroup)
	if preferredChannelID, found := service.GetPreferredChannelByAffinity(c, modelName, usingGroup); found {
		preferred, err := model.CacheGetChannel(preferredChannelID)
		if err == nil && preferred != nil && preferred.Status == common.ChannelStatusEnabled {
			if usingGroup == "auto" {
				userGroup := common.GetContextKeyString(c, constant.ContextKeyUserGroup)
				for _, group := range service.GetRequestAutoGroups(c, userGroup) {
					if model.IsChannelEnabledForGroupModel(group, modelName, preferred.Id) {
						common.SetContextKey(c, constant.ContextKeyAutoGroup, group)
						service.MarkChannelAffinityUsed(c, group, preferred.Id)
						return preferred, nil
					}
				}
			} else if model.IsChannelEnabledForGroupModel(usingGroup, modelName, preferred.Id) {
				service.MarkChannelAffinityUsed(c, usingGroup, preferred.Id)
				return preferred, nil
			}
		}
	}

	channel, selectGroup, err := service.CacheGetRandomSatisfiedChannel(&service.RetryParam{
		Ctx:         c,
		ModelName:   modelName,
		TokenGroup:  usingGroup,
		RequestPath: c.Request.URL.Path,
		Retry:       common.GetPointer(0),
	})
	if err != nil {
		showGroup := usingGroup
		if usingGroup == "auto" {
			showGroup = fmt.Sprintf("auto(%s)", selectGroup)
		}
		return nil, types.NewError(fmt.Errorf("%s", i18n.T(c, i18n.MsgDistributorGetChannelFailed, map[string]any{"Group": showGroup, "Model": modelName, "Error": err.Error()})), types.ErrorCodeGetChannelFailed, types.ErrOptionWithStatusCode(http.StatusServiceUnavailable))
	}
	if channel == nil {
		return nil, types.NewError(fmt.Errorf("%s", i18n.T(c, i18n.MsgDistributorNoAvailableChannel, map[string]any{"Group": usingGroup, "Model": modelName})), types.ErrorCodeGetChannelFailed, types.ErrOptionWithStatusCode(http.StatusServiceUnavailable))
	}
	return channel, nil
}
