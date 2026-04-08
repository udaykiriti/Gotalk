package ws

import (
	"encoding/json"
	"log"
	"strings"
	"sync"

	"gotalk/internal/models"
	"gotalk/internal/storage"
)

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients by room.
	Rooms map[string]map[*Client]bool
	mu    sync.RWMutex

	// Inbound messages from the clients.
	Broadcast chan models.Message

	// Register requests from the clients.
	Register chan *Client

	// Unregister requests from clients.
	Unregister chan *Client
	
	// Terminate the hub.
	Quit chan struct{}

	// Persistence layer
	Store *storage.Store
}

func NewHub(store *storage.Store) *Hub {
	return &Hub{
		Broadcast:  make(chan models.Message),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Rooms:      make(map[string]map[*Client]bool),
		Quit:       make(chan struct{}),
		Store:      store,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			// Register client to room
			if _, ok := h.Rooms[client.Room]; !ok {
				h.Rooms[client.Room] = make(map[*Client]bool)
			}
			h.Rooms[client.Room][client] = true
			h.mu.Unlock()
			log.Printf("Client '%s' registered to room: %s", client.Username, client.Room)

			// 1. Send chat history to the new client
			if h.Store != nil {
				history, err := h.Store.GetRoomHistory(client.Room, 50)
				if err != nil {
					log.Printf("Error fetching history for room %s: %v", client.Room, err)
				} else {
					for _, msg := range history {
						bytes, _ := json.Marshal(msg)
						client.Send <- bytes
					}
				}
			}

			// 2. Broadcast "User Joined" notification
			h.broadcastToRoom(models.Message{
				Type:    models.TypeNotification,
				Room:    client.Room,
				User:    "System",
				Content: client.Username + " joined the room",
			})

			// Broadcast updated user list
			h.broadcastUserList(client.Room)

		case client := <-h.Unregister:
			h.mu.Lock()
			if clients, ok := h.Rooms[client.Room]; ok {
				if _, ok := clients[client]; ok {
					delete(clients, client)
					close(client.Send)
					log.Printf("Client '%s' unregistered from room: %s", client.Username, client.Room)

					// Broadcast "User Left" notification
					h.broadcastToRoom(models.Message{
						Type:    models.TypeNotification,
						Room:    client.Room,
						User:    "System",
						Content: client.Username + " left the room",
					})

					// Broadcast updated user list
					if len(clients) > 0 {
						h.broadcastUserList(client.Room)
					}

					// Clean up empty rooms
					if len(clients) == 0 {
						delete(h.Rooms, client.Room)
					}
				}
			}
			h.mu.Unlock()

		case message := <-h.Broadcast:
			// Persist message if it's a regular chat message
			if h.Store != nil && message.Type == models.TypeMessage {
				_ = h.Store.SaveMessage(message)
			}
			h.broadcastToRoom(message)

		case <-h.Quit:
			log.Println("Shutting down Hub...")
			// Close all clients in all rooms
			for room, clients := range h.Rooms {
				for client := range clients {
					delete(clients, client)
					close(client.Send)
					client.Conn.Close()
				}
				delete(h.Rooms, room)
			}
			return
		}
	}
}

// Stop shuts down the hub
func (h *Hub) Stop() {
	close(h.Quit)
}

// GetActiveRooms returns a list of active room names and their user counts
func (h *Hub) GetActiveRooms() map[string]int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	rooms := make(map[string]int)
	for room, clients := range h.Rooms {
		rooms[room] = len(clients)
	}
	return rooms
}

// broadcastToRoom sends a message to all clients in the specified room
func (h *Hub) broadcastToRoom(msg models.Message) {
	// Validation: Don't broadcast empty or whitespace-only messages
	if strings.TrimSpace(msg.Content) == "" {
		return
	}

	clients, ok := h.Rooms[msg.Room]
	if !ok {
		return
	}

	// Marshaling once for efficiency
	bytes, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}

	for client := range clients {
		select {
		case client.Send <- bytes:
		default:
			close(client.Send)
			delete(clients, client)
		}
	}
}

// broadcastUserList sends the list of active users in a room to all clients in that room
func (h *Hub) broadcastUserList(room string) {
	clients, ok := h.Rooms[room]
	if !ok {
		return
	}

	var users []string
	for client := range clients {
		users = append(users, client.Username)
	}

	msg := models.Message{
		Type:    models.TypeUserList,
		Room:    room,
		Users:   users,
		Content: "user_list_update", // Dummy content to pass validation
	}
	h.broadcastToRoom(msg)
}
