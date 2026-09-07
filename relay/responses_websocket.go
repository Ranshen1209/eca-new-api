package relay

import (
	"fmt"

	"github.com/QuantumNous/new-api/relay/channel/openai"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/relay/helper"
	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/QuantumNous/new-api/relaykit/types"
	"github.com/QuantumNous/new-api/service"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func ResponsesWebsocketHelper(c *gin.Context, info *relaycommon.RelayInfo, firstClientMessage []byte) (newAPIError *types.NewAPIError) {
	info.InitChannelMeta(c)

	adaptor := GetAdaptor(info.ApiType)
	if adaptor == nil {
		return types.NewError(fmt.Errorf("invalid api type: %d", info.ApiType), types.ErrorCodeInvalidApiType, types.ErrOptionWithSkipRetry())
	}
	adaptor.Init(info)

	if err := helper.ModelMappedHelper(c, info, info.Request); err != nil {
		return types.NewError(err, types.ErrorCodeChannelModelMappedError, types.ErrOptionWithSkipRetry())
	}
	if info.IsModelMapped && info.UpstreamModelName != "" {
		firstClientMessage = openai.RewriteResponsesWebsocketUpstreamModel(firstClientMessage, info.UpstreamModelName)
	}

	statusCodeMappingStr := c.GetString("status_code_mapping")
	resp, err := adaptor.DoRequest(c, info, nil)
	if err != nil {
		return types.NewError(err, types.ErrorCodeDoRequestFailed)
	}
	if resp == nil {
		return types.NewError(fmt.Errorf("invalid websocket connection"), types.ErrorCodeDoRequestFailed)
	}

	target, ok := resp.(*websocket.Conn)
	if !ok {
		return types.NewError(fmt.Errorf("invalid websocket connection"), types.ErrorCodeDoRequestFailed)
	}
	info.TargetWs = target
	defer info.TargetWs.Close()

	if err := helper.WssString(c, info.TargetWs, string(firstClientMessage)); err != nil {
		return types.NewError(err, types.ErrorCodeDoRequestFailed)
	}

	usage, newAPIError := adaptor.DoResponse(c, nil, info)
	if newAPIError != nil {
		service.ResetStatusCode(newAPIError, statusCodeMappingStr)
		return newAPIError
	}

	usageDto, ok := usage.(*dto.Usage)
	if !ok || usageDto == nil {
		usageDto = &dto.Usage{}
	}
	service.PostTextConsumeQuota(c, info, usageDto, nil)
	return nil
}
