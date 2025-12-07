package entities

import "time"

// MessageRole represents the role of the message sender
type MessageRole string

const (
	RoleUser  MessageRole = "user"
	RoleAgent MessageRole = "agent"
)

// MessageStatus represents the current status of a message
type MessageStatus string

const (
	StatusSending MessageStatus = "sending"
	StatusSent    MessageStatus = "sent"
	StatusError   MessageStatus = "error"
)

// Message represents a single message in a conversation
type Message struct {
	ID        string
	SessionID string
	Role      MessageRole
	Content   string
	Timestamp time.Time
	Citations []Citation
	Status    MessageStatus
}

// Citation represents a knowledge base citation
type Citation struct {
	SourceID   string
	SourceName string
	Excerpt    string
	Confidence float64
	URL        string
	Metadata   map[string]interface{}
}
