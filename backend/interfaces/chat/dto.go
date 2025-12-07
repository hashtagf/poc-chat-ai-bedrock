package chat

import "time"

// MessageRequest represents an incoming message from the client
type MessageRequest struct {
	SessionID string `json:"session_id"`
	Content   string `json:"content"`
}

// MessageResponse represents a message response to the client
type MessageResponse struct {
	MessageID string             `json:"message_id"`
	Content   string             `json:"content"`
	Citations []CitationResponse `json:"citations,omitempty"`
	Timestamp time.Time          `json:"timestamp"`
}

// CitationResponse represents a citation in the response
type CitationResponse struct {
	SourceID   string                 `json:"source_id"`
	SourceName string                 `json:"source_name"`
	Excerpt    string                 `json:"excerpt"`
	Confidence float64                `json:"confidence,omitempty"`
	URL        string                 `json:"url,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// SessionCreateRequest represents a request to create a new session
type SessionCreateRequest struct {
	// Empty for now, can add initial configuration later
}

// SessionResponse represents a session response
type SessionResponse struct {
	ID            string     `json:"id"`
	CreatedAt     time.Time  `json:"created_at"`
	LastMessageAt *time.Time `json:"last_message_at,omitempty"`
	MessageCount  int        `json:"message_count"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// StreamChunk represents a chunk of streaming data
type StreamChunk struct {
	Type     string            `json:"type"` // "content", "citation", "error", "done"
	Content  string            `json:"content,omitempty"`
	Citation *CitationResponse `json:"citation,omitempty"`
	Error    *ErrorResponse    `json:"error,omitempty"`
}
