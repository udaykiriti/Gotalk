package ws

import (
	"bytes"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"gotalk/internal/models"
	"regexp"
	"sync"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
	// Validation regex: Only alphanumeric, hyphens, underscores
	validNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	Hub *Hub

	// The websocket connection.
	Conn *websocket.Conn

	// Buffered channel of outbound messages.
	Send chan []byte

	// Room the client belongs to
	Room string

	// Username of the client
	Username string

	// Ensure unregister is called only once
	closeOnce sync.Once
}

// Close closes the client connection and sends an unregister signal to the hub.
func (c *Client) Close() {
	c.closeOnce.Do(func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	})
}

// ReadPump pumps messages from the websocket connection to the hub.
//
// The application runs ReadPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) ReadPump() {
	defer func() {
		c.Close()
	}()
	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error { c.Conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))

		// Skip empty messages
		if len(message) == 0 {
			continue
		}

		// Broadcast structured message
		c.Hub.Broadcast <- models.Message{
			Type:    models.TypeMessage,
			Room:    c.Room,
			User:    c.Username,
			Content: string(message),
		}
	}
}

// WritePump pumps messages from the hub to the websocket connection.
//
// A goroutine running WritePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Close()
	}()
	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			if _, err := w.Write(message); err != nil {
				w.Close()
				return
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// ServeWs handles websocket requests from the peer.
func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	// 1. Upgrade connection first
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Upgrade error: %v", err)
		return
	}

	// 2. Extract and validate parameters
	room := strings.TrimSpace(r.URL.Query().Get("room"))
	if room == "" {
		room = "general"
	}
	if !validNameRegex.MatchString(room) || len(room) > 50 {
		log.Printf("Invalid room name: %s", room)
		conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(4000, "Invalid room name"), time.Now().Add(time.Second))
		conn.Close()
		return
	}

	username := strings.TrimSpace(r.URL.Query().Get("user"))
	if username == "" {
		username = "Anonymous"
	}
	if !validNameRegex.MatchString(username) || len(username) > 30 {
		log.Printf("Invalid username: %s", username)
		conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(4001, "Invalid username"), time.Now().Add(time.Second))
		conn.Close()
		return
	}

	// 3. Register client
	client := &Client{
		Hub:      hub,
		Conn:     conn,
		Send:     make(chan []byte, 256),
		Room:     room,
		Username: username,
	}
	client.Hub.Register <- client

	log.Printf("New connection: User='%s' Room='%s' RemoteAddr='%s'", username, room, r.RemoteAddr)

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.ReadPump()
	go client.WritePump()
}
