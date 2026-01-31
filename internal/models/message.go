package models

// MessageType distinguishes between chat messages and system notifications
type MessageType string

const (
	TypeMessage      MessageType = "message"
	TypeNotification MessageType = "notification"
)

// Message represents a message to be broadcasted to a specific room.
// It is structured to support JSON serialization for the frontend.
type Message struct {
	Type    MessageType `json:"type"`
	Room    string      `json:"room"`
	User    string      `json:"user"`
	Content string      `json:"content"`
}
