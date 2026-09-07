package helper

import (
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/relaykit/types"

	"github.com/gorilla/websocket"
)

const (
	ContextKeyResponsesWebsocketClient = "responses_websocket_client"
	ContextKeyResponsesWebsocketTarget = "responses_websocket_target"

	// ResponsesWebsocketHTTPReplayCloseReason matches CLIProxyAPI / Codex:
	// close 1012 makes the official CLI fall back to HTTP POST / SSE.
	ResponsesWebsocketHTTPReplayCloseReason = "upstream requires HTTP replay"

	responsesWebsocketPingWait  = 5 * time.Second
	responsesWebsocketWriteWait = 60 * time.Second
	responsesWebsocketReadLimit = 32 << 20
)

type responsesWebsocketFrame struct {
	messageType int
	payload     []byte
}

// ResponsesWebsocketConn is a gorilla websocket with one reader, mutexed
// writers, and automatic pong replies. Codex and CPA send pings while the
// gateway is still selecting a channel or copying events.
type ResponsesWebsocketConn struct {
	conn      *websocket.Conn
	writeMu   sync.Mutex
	closeOnce sync.Once
	msgs      chan responsesWebsocketFrame
	readErr   error
	errOnce   sync.Once
}

func NewResponsesWebsocketConn(conn *websocket.Conn) *ResponsesWebsocketConn {
	if conn == nil {
		return nil
	}
	wrapped := &ResponsesWebsocketConn{
		conn: conn,
		msgs: make(chan responsesWebsocketFrame, 64),
	}
	conn.SetReadLimit(responsesWebsocketReadLimit)
	conn.SetPingHandler(func(appData string) error {
		wrapped.writeMu.Lock()
		defer wrapped.writeMu.Unlock()
		err := conn.WriteControl(websocket.PongMessage, []byte(appData), time.Now().Add(responsesWebsocketPingWait))
		if errors.Is(err, websocket.ErrCloseSent) {
			return nil
		}
		var netErr net.Error
		if errors.As(err, &netErr) && netErr.Timeout() {
			return nil
		}
		return err
	})
	go wrapped.readLoop()
	return wrapped
}

func (c *ResponsesWebsocketConn) readLoop() {
	defer close(c.msgs)
	for {
		messageType, payload, err := c.conn.ReadMessage()
		if err != nil {
			c.errOnce.Do(func() {
				c.readErr = err
			})
			return
		}
		if messageType != websocket.TextMessage && messageType != websocket.BinaryMessage {
			continue
		}
		c.msgs <- responsesWebsocketFrame{messageType: messageType, payload: payload}
	}
}

func (c *ResponsesWebsocketConn) Read() (int, []byte, error) {
	if c == nil {
		return 0, nil, errors.New("websocket connection is nil")
	}
	frame, ok := <-c.msgs
	if !ok {
		if c.readErr != nil {
			return 0, nil, c.readErr
		}
		return 0, nil, io.EOF
	}
	return frame.messageType, frame.payload, nil
}

func (c *ResponsesWebsocketConn) Write(messageType int, payload []byte) error {
	if c == nil || c.conn == nil {
		return errors.New("websocket connection is nil")
	}
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	if err := c.conn.SetWriteDeadline(time.Now().Add(responsesWebsocketWriteWait)); err != nil {
		return err
	}
	err := c.conn.WriteMessage(messageType, payload)
	_ = c.conn.SetWriteDeadline(time.Time{})
	return err
}

func (c *ResponsesWebsocketConn) Close(code int, reason string) {
	if c == nil || c.conn == nil {
		return
	}
	c.closeOnce.Do(func() {
		c.writeMu.Lock()
		_ = c.conn.WriteControl(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(code, reason),
			time.Now().Add(time.Second),
		)
		c.writeMu.Unlock()
		_ = c.conn.Close()
	})
}

// ResponsesWebsocketClientCloseCode maps an upstream websocket death onto a
// close code Codex can handle. Abnormal 1006/EOF becomes 1012 so the CLI
// retries over HTTP instead of reporting a protocol reset.
func ResponsesWebsocketClientCloseCode(err error) (int, string) {
	if err == nil {
		return websocket.CloseNormalClosure, ""
	}
	var closeErr *websocket.CloseError
	if errors.As(err, &closeErr) {
		switch closeErr.Code {
		case websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseServiceRestart, websocket.CloseMessageTooBig:
			return closeErr.Code, closeErr.Text
		}
	}
	return websocket.CloseServiceRestart, ResponsesWebsocketHTTPReplayCloseReason
}

func WriteResponsesWebsocketError(conn *ResponsesWebsocketConn, openaiError types.OpenAIError) {
	WriteResponsesWebsocketFailedTurn(conn, openaiError)
}

func IsResponsesWebsocketCompletionEvent(eventType string) bool {
	switch strings.TrimSpace(eventType) {
	case "response.completed", "response.done", "response.failed", "response.incomplete", "response.cancelled", "response.canceled":
		return true
	default:
		return false
	}
}

func ResponsesWebsocketFailedTurnPayloads(openaiError types.OpenAIError) [][]byte {
	message := strings.TrimSpace(openaiError.Message)
	if message == "" {
		message = "upstream websocket closed before response.completed"
	}
	errType := strings.TrimSpace(openaiError.Type)
	if errType == "" {
		errType = "server_error"
	}
	code := openaiError.Code
	if code == nil || strings.TrimSpace(fmt.Sprint(code)) == "" {
		code = "upstream_closed"
	}
	detail := map[string]any{
		"type":    errType,
		"code":    code,
		"message": message,
	}
	response := map[string]any{
		"status": "failed",
		"error":  detail,
		"output": []any{},
	}
	events := []map[string]any{
		{"type": "error", "error": openaiError},
		{"type": "response.failed", "response": response},
		{"type": "response.completed", "response": response},
	}
	payloads := make([][]byte, 0, len(events))
	for _, event := range events {
		payload, err := common.Marshal(event)
		if err != nil {
			continue
		}
		payloads = append(payloads, payload)
	}
	return payloads
}

func WriteResponsesWebsocketFailedTurn(conn *ResponsesWebsocketConn, openaiError types.OpenAIError) {
	if conn == nil {
		return
	}
	for _, payload := range ResponsesWebsocketFailedTurnPayloads(openaiError) {
		_ = conn.Write(websocket.TextMessage, payload)
	}
}

func ResponsesWebsocketConnFromContext(c interface{ Get(string) (any, bool) }, key string) *ResponsesWebsocketConn {
	if c == nil {
		return nil
	}
	value, ok := c.Get(key)
	if !ok {
		return nil
	}
	conn, _ := value.(*ResponsesWebsocketConn)
	return conn
}
