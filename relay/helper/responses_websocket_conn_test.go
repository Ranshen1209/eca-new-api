package helper

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/relaykit/types"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

func TestResponsesWebsocketConnAnswersPingDuringSetup(t *testing.T) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		wrapped := NewResponsesWebsocketConn(conn)
		defer wrapped.Close(websocket.CloseNormalClosure, "")

		time.Sleep(300 * time.Millisecond)

		_, payload, err := wrapped.Read()
		if err != nil {
			return
		}
		_ = wrapped.Write(websocket.TextMessage, payload)
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	client, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer client.Close()

	pongCh := make(chan struct{}, 1)
	client.SetPongHandler(func(string) error {
		select {
		case pongCh <- struct{}{}:
		default:
		}
		return nil
	})
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			if _, _, readErr := client.ReadMessage(); readErr != nil {
				return
			}
		}
	}()

	require.NoError(t, client.WriteControl(websocket.PingMessage, []byte("setup"), time.Now().Add(time.Second)))
	select {
	case <-pongCh:
	case <-time.After(2 * time.Second):
		t.Fatal("server did not answer ping while setup was still running")
	}

	require.NoError(t, client.WriteMessage(websocket.TextMessage, []byte(`{"type":"response.create"}`)))
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("client read loop did not observe the echoed setup frame")
	}
}

func TestResponsesWebsocketClientCloseCode(t *testing.T) {
	code, reason := ResponsesWebsocketClientCloseCode(nil)
	require.Equal(t, websocket.CloseNormalClosure, code)
	require.Equal(t, "", reason)

	code, reason = ResponsesWebsocketClientCloseCode(&websocket.CloseError{Code: websocket.CloseServiceRestart, Text: ResponsesWebsocketHTTPReplayCloseReason})
	require.Equal(t, websocket.CloseServiceRestart, code)
	require.Equal(t, ResponsesWebsocketHTTPReplayCloseReason, reason)

	code, reason = ResponsesWebsocketClientCloseCode(&websocket.CloseError{Code: websocket.CloseAbnormalClosure, Text: "unexpected EOF"})
	require.Equal(t, websocket.CloseServiceRestart, code)
	require.Equal(t, ResponsesWebsocketHTTPReplayCloseReason, reason)

	code, reason = ResponsesWebsocketClientCloseCode(io.EOF)
	require.Equal(t, websocket.CloseServiceRestart, code)
	require.Equal(t, ResponsesWebsocketHTTPReplayCloseReason, reason)
}

func TestResponsesWebsocketFailedTurnPayloadsIncludeCompleted(t *testing.T) {
	payloads := ResponsesWebsocketFailedTurnPayloads(types.OpenAIError{
		Message: "upstream closed",
		Type:    "server_error",
		Code:    "upstream_closed",
	})
	require.Len(t, payloads, 3)

	var typesFound []string
	for _, payload := range payloads {
		var event map[string]any
		require.NoError(t, json.Unmarshal(payload, &event))
		typesFound = append(typesFound, event["type"].(string))
	}
	require.Equal(t, []string{"error", "response.failed", "response.completed"}, typesFound)

	var completed map[string]any
	require.NoError(t, json.Unmarshal(payloads[2], &completed))
	response := completed["response"].(map[string]any)
	require.Equal(t, "failed", response["status"])
}

func TestIsResponsesWebsocketCompletionEvent(t *testing.T) {
	require.True(t, IsResponsesWebsocketCompletionEvent("response.completed"))
	require.True(t, IsResponsesWebsocketCompletionEvent("response.failed"))
	require.False(t, IsResponsesWebsocketCompletionEvent("response.created"))
}
