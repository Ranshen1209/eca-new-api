package openai

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/QuantumNous/new-api/common"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/gin-gonic/gin"
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

func TestCopyResponsesWebsocketClientHeadersForwardsSessionIdentity(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/v1/responses", nil)
	c.Request.Header.Set("Authorization", "Bearer user-token")
	c.Request.Header.Set("X-Client-Request-Id", "codex-turn-aaa")
	c.Request.Header.Set("Session-Id", "session-aaa")
	c.Request.Header.Set("X-Codex-Turn-State", "turn-state-aaa")
	c.Request.Header.Set("Thread-Id", "thread-aaa")
	c.Request.Header.Set("User-Agent", "codex-cli")
	c.Set(common.RequestIdKey, "eca-req-should-not-replace")

	header := http.Header{}
	header.Set("Authorization", "Bearer channel-key")
	copyResponsesWebsocketClientHeaders(c, &header)

	assert.Equal(t, "Bearer channel-key", header.Get("Authorization"))
	assert.Equal(t, "codex-turn-aaa", header.Get("X-Client-Request-Id"))
	assert.Equal(t, "session-aaa", header.Get("Session-Id"))
	assert.Equal(t, "turn-state-aaa", header.Get("X-Codex-Turn-State"))
	assert.Equal(t, "thread-aaa", header.Get("Thread-Id"))
	assert.Equal(t, "codex-cli", header.Get("User-Agent"))
}

func TestCopyResponsesWebsocketClientHeadersUsesRequestIdWhenSessionMissing(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/v1/responses", nil)
	c.Set(common.RequestIdKey, "eca-req-bbb")

	header := http.Header{}
	copyResponsesWebsocketClientHeaders(c, &header)

	assert.Equal(t, "eca-req-bbb", header.Get("X-Client-Request-Id"))
	assert.Empty(t, header.Get("Authorization"))
	assert.Empty(t, header.Get("Session-Id"))
}

func TestOpenaiResponsesWebsocketHandlerRejectsMissingWrappers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/v1/responses", nil)

	err, usage := OpenaiResponsesWebsocketHandler(c, &relaycommon.RelayInfo{})
	require.NotNil(t, err)
	assert.Nil(t, usage)
}
