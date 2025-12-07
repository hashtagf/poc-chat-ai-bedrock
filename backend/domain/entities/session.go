package entities

import "time"

// Session represents a conversation session
type Session struct {
	ID            string
	CreatedAt     time.Time
	LastMessageAt *time.Time
	MessageCount  int
}
