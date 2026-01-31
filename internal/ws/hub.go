package ws

import (
	"encoding/json"
	"log"

	"gotalk/internal/models"
)

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients by room.
	Rooms map[string]map[*Client]bool

	// Inbound messages from the clients.
	Broadcast chan models.Message

	// Register requests from the clients.
	Register chan *Client

	// Unregister requests from clients.
	Unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		Broadcast:  make(chan models.Message),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Rooms:      make(map[string]map[*Client]bool),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			// Register client to room
			if _, ok := h.Rooms[client.Room]; !ok {
				h.Rooms[client.Room] = make(map[*Client]bool)
			}
			h.Rooms[client.Room][client] = true
			log.Printf("Client '%s' registered to room: %s", client.Username, client.Room)

			// Broadcast "User Joined" notification
			h.broadcastToRoom(models.Message{
				Type:    models.TypeNotification,
				Room:    client.Room,
				User:    "System",
				Content: client.Username + " joined the room",
			})

		case client := <-h.Unregister:
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

					// Clean up empty rooms
					if len(clients) == 0 {
						delete(h.Rooms, client.Room)
					}
				}
			}

		case message := <-h.Broadcast:
			h.broadcastToRoom(message)
		}
	}
}

// broadcastToRoom sends a message to all clients in the specified room
func (h *Hub) broadcastToRoom(msg models.Message) {
	// Validation: Don't broadcast empty messages
	if msg.Content == "" {
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
