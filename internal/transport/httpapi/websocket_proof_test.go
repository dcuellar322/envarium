package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	session "envarium.dev/envarium/internal/session/application"
	"envarium.dev/envarium/internal/transport/contract"
	"github.com/coder/websocket"
)

func TestWebSocketNegotiatesPublicProtocolWithoutEchoingProof(t *testing.T) {
	sessions := session.NewManager()
	boot, _ := sessions.IssueBootstrap()
	active, _ := sessions.Exchange(boot.Token)
	handler := withBrowserSecurity(sessions, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		connection, err := websocket.Accept(w, r, &websocket.AcceptOptions{Subprotocols: []string{contract.BrowserSocketProtocol}})
		if err != nil {
			return
		}
		defer func() { _ = connection.CloseNow() }()
		_ = connection.Write(r.Context(), websocket.MessageText, []byte("connected"))
	}))
	server := httptest.NewServer(handler)
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	connection, response, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(server.URL, "http")+"/ws/v1/events", &websocket.DialOptions{
		HTTPHeader: http.Header{"Origin": {server.URL}}, Subprotocols: []string{contract.BrowserSocketProtocol, "envarium.session." + active.CSRFToken},
	})
	if response != nil && response.Body != nil {
		defer func() { _ = response.Body.Close() }()
	}
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = connection.CloseNow() }()
	if connection.Subprotocol() != contract.BrowserSocketProtocol || strings.Contains(response.Header.Get("Sec-WebSocket-Protocol"), active.CSRFToken) {
		t.Fatal("secret negotiated instead of public protocol")
	}
	_, message, err := connection.Read(ctx)
	if err != nil || string(message) != "connected" {
		t.Fatalf("message=%q err=%v", message, err)
	}
}
