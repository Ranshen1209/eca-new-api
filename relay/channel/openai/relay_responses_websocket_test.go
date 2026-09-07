package openai

import (
	"testing"

	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResponsesWebsocketRequestModel(t *testing.T) {
	assert.Equal(t, "gpt-6-astra", ResponsesWebsocketRequestModel([]byte(`{"type":"response.create","model":"gpt-6-astra"}`)))
	assert.Equal(t, "gpt-6-astra", ResponsesWebsocketRequestModel([]byte(`{"type":"response.created","response":{"model":"gpt-6-astra"}}`)))
	assert.Equal(t, "", ResponsesWebsocketRequestModel([]byte(`{"type":"response.create","input":[]}`)))
}

func TestRewriteResponsesWebsocketUpstreamModel(t *testing.T) {
	got := RewriteResponsesWebsocketUpstreamModel([]byte(`{"type":"response.create","model":"alias"}`), "gpt-6-astra")
	assert.Equal(t, "gpt-6-astra", ResponsesWebsocketRequestModel(got))

	unchanged := RewriteResponsesWebsocketUpstreamModel([]byte(`{"type":"response.create","input":[]}`), "gpt-6-astra")
	assert.Equal(t, "", ResponsesWebsocketRequestModel(unchanged))
}

func TestAccumulateResponsesWebsocketUsage(t *testing.T) {
	info := &relaycommon.RelayInfo{
		ResponsesUsageInfo: &relaycommon.ResponsesUsageInfo{
			BuiltInTools: map[string]*relaycommon.BuildInToolInfo{},
		},
	}
	usage := &dto.Usage{}
	imageCounter := &relaycommon.ImageGenerationCallCounter{}
	imageCommitted := false

	accumulateResponsesWebsocketUsage(info, usage, dto.ResponsesStreamResponse{
		Type: "response.completed",
		Response: &dto.OpenAIResponsesResponse{
			Usage: &dto.Usage{
				InputTokens:  11,
				OutputTokens: 7,
				TotalTokens:  18,
				InputTokensDetails: &dto.InputTokenDetails{
					CachedTokens: 3,
				},
			},
		},
	}, imageCounter, &imageCommitted)

	require.Equal(t, 11, usage.PromptTokens)
	require.Equal(t, 7, usage.CompletionTokens)
	require.Equal(t, 18, usage.TotalTokens)
	require.Equal(t, 3, usage.PromptTokensDetails.CachedTokens)
	assert.True(t, imageCommitted)
}
