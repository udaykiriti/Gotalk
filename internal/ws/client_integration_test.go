package ws

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"gotalk/internal/models"
)

func TestWebSocketMessagesAreSingleJSONFrames(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ServeWs(hub, w, r)
	})

	l, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Skipf("skipping websocket integration test: local listen unavailable: %v", err)
	}
	srv := &http.Server{Handler: mux}
	go func() {
		_ = srv.Serve(l)
	}()
	defer func() {
		_ = srv.Close()
	}()

	wsBase := "ws://" + l.Addr().String()
	room := "room-single-frame"

	receiver, err := dialWS(wsBase, room, "receiver")
	if err != nil {
		t.Skipf("skipping websocket integration test: dial receiver failed: %v", err)
	}
	defer receiver.Close()

	sender, err := dialWS(wsBase, room, "sender")
	if err != nil {
		t.Skipf("skipping websocket integration test: dial sender failed: %v", err)
	}
	defer sender.Close()

	if err := sender.WriteMessage(websocket.TextMessage, []byte("first")); err != nil {
		t.Fatalf("write first: %v", err)
	}
	if err := sender.WriteMessage(websocket.TextMessage, []byte("second")); err != nil {
		t.Fatalf("write second: %v", err)
	}

	var got []string
	deadline := time.Now().Add(4 * time.Second)
	for len(got) < 2 && time.Now().Before(deadline) {
		msg, err := readMessage(receiver, 1500*time.Millisecond)
		if err != nil {
			t.Fatalf("read/parse frame failed: %v", err)
		}

		if msg.Type == models.TypeMessage && msg.User == "sender" {
			got = append(got, msg.Content)
		}
	}

	if len(got) != 2 {
		t.Fatalf("expected 2 chat messages from sender, got %d (%v)", len(got), got)
	}
	if got[0] != "first" || got[1] != "second" {
		t.Fatalf("unexpected message order/content: %v", got)
	}
}

func dialWS(base, room, user string) (*websocket.Conn, error) {
	u, err := url.Parse(base + "/ws")
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("room", room)
	q.Set("user", user)
	u.RawQuery = q.Encode()

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	return conn, err
}

func readMessage(conn *websocket.Conn, timeout time.Duration) (models.Message, error) {
	_ = conn.SetReadDeadline(time.Now().Add(timeout))
	_, payload, err := conn.ReadMessage()
	if err != nil {
		return models.Message{}, err
	}

	var msg models.Message
	if err := json.Unmarshal(payload, &msg); err != nil {
		return models.Message{}, fmt.Errorf("invalid JSON frame: %w; payload=%q", err, string(payload))
	}
	return msg, nil
}
