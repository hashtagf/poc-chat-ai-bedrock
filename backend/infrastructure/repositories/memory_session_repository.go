package repositories

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/bedrock-chat-poc/backend/domain/entities"
)

const (
	// SessionTimeout is the duration after which inactive sessions are considered expired
	SessionTimeout = 30 * time.Minute
)

// MemorySessionRepository implements SessionRepository with in-memory storage
type MemorySessionRepository struct {
	sessions        map[string]*entities.Session
	messageHistory  map[string][]*entities.Message // sessionID -> messages
	mu              sync.RWMutex
	cleanupInterval time.Duration
	stopCleanup     chan struct{}
}

// NewMemorySessionRepository creates a new in-memory session repository
func NewMemorySessionRepository() *MemorySessionRepository {
	repo := &MemorySessionRepository{
		sessions:        make(map[string]*entities.Session),
		messageHistory:  make(map[string][]*entities.Message),
		cleanupInterval: 5 * time.Minute, // Check for expired sessions every 5 minutes
		stopCleanup:     make(chan struct{}),
	}
	
	// Start background cleanup goroutine
	go repo.cleanupExpiredSessions()
	
	return repo
}

// Close stops the background cleanup goroutine
func (r *MemorySessionRepository) Close() {
	close(r.stopCleanup)
}

// Create stores a new session
func (r *MemorySessionRepository) Create(ctx context.Context, session *entities.Session) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.sessions[session.ID]; exists {
		return fmt.Errorf("session %s already exists", session.ID)
	}

	r.sessions[session.ID] = session
	return nil
}

// FindByID retrieves a session by ID
func (r *MemorySessionRepository) FindByID(ctx context.Context, id string) (*entities.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	session, exists := r.sessions[id]
	if !exists {
		return nil, fmt.Errorf("session %s not found", id)
	}

	return session, nil
}

// List returns all sessions
func (r *MemorySessionRepository) List(ctx context.Context) ([]*entities.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	sessions := make([]*entities.Session, 0, len(r.sessions))
	for _, session := range r.sessions {
		sessions = append(sessions, session)
	}

	return sessions, nil
}

// Update modifies an existing session
func (r *MemorySessionRepository) Update(ctx context.Context, session *entities.Session) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.sessions[session.ID]; !exists {
		return fmt.Errorf("session %s not found", session.ID)
	}

	r.sessions[session.ID] = session
	return nil
}

// Delete removes a session and its message history
func (r *MemorySessionRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.sessions[id]; !exists {
		return fmt.Errorf("session %s not found", id)
	}

	delete(r.sessions, id)
	delete(r.messageHistory, id)
	return nil
}

// AddMessage adds a message to a session's history
func (r *MemorySessionRepository) AddMessage(ctx context.Context, message *entities.Message) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	session, exists := r.sessions[message.SessionID]
	if !exists {
		return fmt.Errorf("session %s not found", message.SessionID)
	}

	// Add message to history
	r.messageHistory[message.SessionID] = append(r.messageHistory[message.SessionID], message)

	// Update session metadata
	session.MessageCount++
	session.LastMessageAt = &message.Timestamp

	return nil
}

// GetMessages retrieves all messages for a session
func (r *MemorySessionRepository) GetMessages(ctx context.Context, sessionID string) ([]*entities.Message, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if _, exists := r.sessions[sessionID]; !exists {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}

	messages := r.messageHistory[sessionID]
	if messages == nil {
		return []*entities.Message{}, nil
	}

	return messages, nil
}

// IsExpired checks if a session has exceeded the inactivity timeout
func (r *MemorySessionRepository) IsExpired(session *entities.Session) bool {
	var lastActivity time.Time
	if session.LastMessageAt != nil {
		lastActivity = *session.LastMessageAt
	} else {
		lastActivity = session.CreatedAt
	}

	return time.Since(lastActivity) > SessionTimeout
}

// cleanupExpiredSessions runs periodically to remove expired sessions
func (r *MemorySessionRepository) cleanupExpiredSessions() {
	ticker := time.NewTicker(r.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			r.removeExpiredSessions()
		case <-r.stopCleanup:
			return
		}
	}
}

// removeExpiredSessions removes all sessions that have exceeded the timeout
func (r *MemorySessionRepository) removeExpiredSessions() {
	r.mu.Lock()
	defer r.mu.Unlock()

	expiredIDs := []string{}
	for id, session := range r.sessions {
		if r.IsExpired(session) {
			expiredIDs = append(expiredIDs, id)
		}
	}

	for _, id := range expiredIDs {
		delete(r.sessions, id)
		delete(r.messageHistory, id)
	}
}
