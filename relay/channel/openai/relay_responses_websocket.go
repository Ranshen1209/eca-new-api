package openai

import (
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/logger"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/relay/helper"
	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/QuantumNous/new-api/relaykit/types"
	"github.com/QuantumNous/new-api/setting/operation_setting"

	"github.com/bytedance/gopkg/util/gopool"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

var responsesWebsocketClientHeaders = []string{
	"Version",
	"OpenAI-Beta",
	"ChatGPT-Account-ID",
	"ChatGPT-Account-Id",
}

func copyResponsesWebsocketClientHeaders(c *gin.Context, header *http.Header) {
	if c == nil || c.Request == nil || header == nil {
		return
	}
	codexHeaders := operation_setting.CodexCliPassThroughHeaders()
	names := make([]string, 0, len(responsesWebsocketClientHeaders)+len(codexHeaders))
	names = append(names, responsesWebsocketClientHeaders...)
	names = append(names, codexHeaders...)
	seen := make(map[string]struct{}, len(names))
	for _, name := range names {
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		value := strings.TrimSpace(c.Request.Header.Get(name))
		if value != "" {
			header.Set(name, value)
		}
	}
	if strings.TrimSpace(header.Get("X-Client-Request-Id")) == "" {
		if requestID := strings.TrimSpace(c.GetString(common.RequestIdKey)); requestID != "" {
			header.Set("X-Client-Request-Id", requestID)
		}
	}
}

// ResponsesWebsocketRequestModel reads the model from a Codex/OpenAI
// Responses WebSocket client frame. Official clients send it on response.create.
func ResponsesWebsocketRequestModel(payload []byte) string {
	if model := strings.TrimSpace(gjson.GetBytes(payload, "model").String()); model != "" {
		return model
	}
	return strings.TrimSpace(gjson.GetBytes(payload, "response.model").String())
}

func RewriteResponsesWebsocketUpstreamModel(payload []byte, upstreamModel string) []byte {
	if strings.TrimSpace(upstreamModel) == "" {
		return payload
	}
	if strings.TrimSpace(gjson.GetBytes(payload, "model").String()) == "" {
		return payload
	}
	updated, err := sjson.SetBytes(payload, "model", upstreamModel)
	if err != nil {
		return payload
	}
	return updated
}

func accumulateResponsesWebsocketUsage(info *relaycommon.RelayInfo, usage *dto.Usage, event dto.ResponsesStreamResponse, imageCounter *relaycommon.ImageGenerationCallCounter, imageCommitted *bool) {
	if usage == nil {
		return
	}
	switch event.Type {
	case "response.completed", "response.done":
		if event.Response != nil {
			if event.Response.Usage != nil {
				addResponsesWebsocketUsage(usage, event.Response.Usage)
			}
			if imageCommitted != nil && !*imageCommitted && imageCounter != nil {
				if relaycommon.IsNonBillableResponsesStatus(event.Response.Status) {
					imageCounter.Reset()
					imageCounter.Commit(info)
					*imageCommitted = true
				} else {
					for i := range event.Response.Output {
						idx := i
						imageCounter.Observe(&event.Response.Output[i], &idx)
					}
					imageCounter.Commit(info)
					*imageCommitted = true
				}
			}
		} else if imageCommitted != nil && !*imageCommitted && imageCounter != nil {
			imageCounter.Commit(info)
			*imageCommitted = true
		}
	case "response.failed", "response.incomplete", "response.cancelled", "response.canceled":
		if imageCommitted != nil && !*imageCommitted && imageCounter != nil {
			imageCounter.Reset()
			imageCounter.Commit(info)
			*imageCommitted = true
		}
	case dto.ResponsesOutputTypeItemDone:
		if event.Item == nil || info == nil {
			return
		}
		switch event.Item.Type {
		case dto.BuildInCallWebSearchCall:
			info.CountBillableToolCall(dto.BuildInCallWebSearchCall, "")
		case dto.BuildInCallFileSearchCall:
			info.CountBillableToolCall(dto.BuildInCallFileSearchCall, "")
		case dto.BuildInCallFunctionCall:
			info.CountBillableToolCall(dto.BuildInCallFunctionCall, event.Item.Name)
		case dto.ResponsesOutputTypeImageGenerationCall:
			if imageCommitted != nil && !*imageCommitted && imageCounter != nil {
				imageCounter.Observe(event.Item, event.OutputIndex)
			}
		}
	}
}

func addResponsesWebsocketUsage(total *dto.Usage, part *dto.Usage) {
	if total == nil || part == nil {
		return
	}
	prompt := part.PromptTokens
	if prompt == 0 {
		prompt = part.InputTokens
	}
	completion := part.CompletionTokens
	if completion == 0 {
		completion = part.OutputTokens
	}
	total.PromptTokens += prompt
	total.CompletionTokens += completion
	if part.TotalTokens != 0 {
		total.TotalTokens += part.TotalTokens
	} else {
		total.TotalTokens += prompt + completion
	}
	total.PromptTokensDetails.CachedTokens += part.PromptTokensDetails.CachedTokens
	total.PromptTokensDetails.CacheWriteTokens += part.PromptTokensDetails.CacheWriteTokens
	if part.InputTokensDetails != nil {
		total.PromptTokensDetails.CachedTokens += part.InputTokensDetails.CachedTokens
		total.PromptTokensDetails.CacheWriteTokens += part.InputTokensDetails.CacheWriteTokens
	}
}

type responsesWebsocketCopyErr struct {
	side string
	err  error
}

func OpenaiResponsesWebsocketHandler(c *gin.Context, info *relaycommon.RelayInfo) (*types.NewAPIError, *dto.Usage) {
	if info == nil {
		return types.NewError(fmt.Errorf("invalid websocket connection"), types.ErrorCodeBadResponse), nil
	}
	// Use the wrappers already bound to this request. Wrapping info.ClientWs /
	// info.TargetWs again would start a second readLoop on the same gorilla
	// connection and split frames across two local sessions.
	clientConn := helper.ResponsesWebsocketConnFromContext(c, helper.ContextKeyResponsesWebsocketClient)
	targetConn := helper.ResponsesWebsocketConnFromContext(c, helper.ContextKeyResponsesWebsocketTarget)
	if clientConn == nil || targetConn == nil {
		return types.NewError(fmt.Errorf("invalid websocket connection"), types.ErrorCodeBadResponse), nil
	}

	info.IsStream = true
	errChan := make(chan responsesWebsocketCopyErr, 2)
	usage := &dto.Usage{}
	imageCounter := &relaycommon.ImageGenerationCallCounter{}
	imageCommitted := false
	var completed atomic.Bool

	gopool.Go(func() {
		defer func() {
			if r := recover(); r != nil {
				errChan <- responsesWebsocketCopyErr{side: "client", err: fmt.Errorf("panic in client reader: %v", r)}
			}
		}()
		for {
			messageType, message, err := clientConn.Read()
			if err != nil {
				errChan <- responsesWebsocketCopyErr{side: "client", err: err}
				return
			}
			if info.IsModelMapped && info.UpstreamModelName != "" {
				message = RewriteResponsesWebsocketUpstreamModel(message, info.UpstreamModelName)
			}
			if err := targetConn.Write(messageType, message); err != nil {
				errChan <- responsesWebsocketCopyErr{side: "target", err: fmt.Errorf("error writing to target: %v", err)}
				return
			}
		}
	})

	gopool.Go(func() {
		defer func() {
			if r := recover(); r != nil {
				errChan <- responsesWebsocketCopyErr{side: "target", err: fmt.Errorf("panic in target reader: %v", r)}
			}
		}()
		for {
			messageType, message, err := targetConn.Read()
			if err != nil {
				errChan <- responsesWebsocketCopyErr{side: "target", err: err}
				return
			}
			info.SetFirstResponseTime()
			var streamResponse dto.ResponsesStreamResponse
			if unmarshalErr := common.Unmarshal(message, &streamResponse); unmarshalErr == nil && streamResponse.Type != "" {
				if helper.IsResponsesWebsocketCompletionEvent(streamResponse.Type) {
					completed.Store(true)
				}
				accumulateResponsesWebsocketUsage(info, usage, streamResponse, imageCounter, &imageCommitted)
			}
			if err := clientConn.Write(messageType, message); err != nil {
				errChan <- responsesWebsocketCopyErr{side: "client", err: fmt.Errorf("error writing to client: %v", err)}
				return
			}
		}
	})

	result := <-errChan
	if result.side == "target" {
		if !websocket.IsCloseError(result.err, websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseNoStatusReceived) {
			logger.LogError(c, "responses websocket error: "+result.err.Error())
		}
		if !completed.Load() {
			helper.WriteResponsesWebsocketFailedTurn(clientConn, types.OpenAIError{
				Message: "upstream websocket closed before response.completed",
				Type:    "server_error",
				Code:    "upstream_closed",
			})
		}
		clientConn.Close(websocket.CloseNormalClosure, "")
	}

	if usage.TotalTokens == 0 && (usage.PromptTokens != 0 || usage.CompletionTokens != 0) {
		usage.TotalTokens = usage.PromptTokens + usage.CompletionTokens
	}
	if !imageCommitted {
		imageCounter.Commit(info)
	}
	return nil, usage
}
