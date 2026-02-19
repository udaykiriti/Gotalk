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
}

// ReadPump pumps messages from the websocket connection to the hub.
//
// The application runs ReadPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
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
		c.Conn.Close()
		// Ensure client is unregistered when WritePump exits
		c.Hub.Unregister <- c
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
	if !validNameRegex.MatchString(room) {
		log.Printf("Invalid room name: %s", room)
		conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "Invalid room name"))
		conn.Close()
		return
	}
	// Sanitize room name length
	if len(room) > 50 {
		room = room[:50]
	}

	username := strings.TrimSpace(r.URL.Query().Get("user"))
	if username == "" {
		username = "Anonymous"
	}
	if !validNameRegex.MatchString(username) {
		log.Printf("Invalid username: %s", username)
		conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "Invalid username"))
		conn.Close()
		return
	}
	// Sanitize username length
	if len(username) > 30 {
		username = username[:30]
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
	go client.WritePump()
	go client.ReadPump()
}
