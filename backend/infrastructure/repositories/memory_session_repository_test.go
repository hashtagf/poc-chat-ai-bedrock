package repositories

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/bedrock-chat-poc/backend/domain/entities"
)

func TestMemorySessionRepository_Create(t *testing.T) {
	repo := NewMemorySessionRepository()
	defer repo.Close()
	ctx := context.Background()

	session := &entities.Session{
		ID:           "test-id",
		CreatedAt:    time.Now(),
		MessageCount: 0,
	}

	err := repo.Create(ctx, session)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Verify session was created
	found, err := repo.FindByID(ctx, "test-id")
	if err != nil {
		t.Fatalf("Failed to find session: %v", err)
	}

	if found.ID != session.ID {
		t.Errorf("Expected ID %s, got %s", session.ID, found.ID)
	}
}

func TestMemorySessionRepository_Create_Duplicate(t *testing.T) {
	repo := NewMemorySessionRepository()
	defer repo.Close()
	ctx := context.Background()

	session := &entities.Session{
		ID:           "test-id",
		CreatedAt:    time.Now(),
		MessageCount: 0,
	}

	// Create first time
	if err := repo.Create(ctx, session); err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Try to create again with same ID
	err := repo.Create(ctx, session)
	if err == nil {
		t.Error("Expected error when creating duplicate session, got nil")
	}
}

func TestMemorySessionRepository_FindByID_NotFound(t *testing.T) {
	repo := NewMemorySessionRepository()
	defer repo.Close()
	ctx := context.Background()

	_, err := repo.FindByID(ctx, "nonexistent")
	if err == nil {
		t.Error("Expected error when finding nonexistent session, got nil")
	}
}

func TestMemorySessionRepository_List(t *testing.T) {
	repo := NewMemorySessionRepository()
	defer repo.Close()
	ctx := context.Background()

	// Create multiple sessions
	for i := 0; i < 3; i++ {
		session := &entities.Session{
			ID:           string(rune('a' + i)),
			CreatedAt:    time.Now(),
			MessageCount: i,
		}
		if err := repo.Create(ctx, session); err != nil {
			t.Fatalf("Failed to create session: %v", err)
		}
	}

	sessions, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("Failed to list sessions: %v", err)
	}

	if len(sessions) != 3 {
		t.Errorf("Expected 3 sessions, got %d", len(sessions))
	}
}

func TestMemorySessionRepository_Update(t *testing.T) {
	repo := NewMemorySessionRepository()
	defer repo.Close()
	ctx := context.Background()

	session := &entities.Session{
		ID:           "test-id",
		CreatedAt:    time.Now(),
		MessageCount: 0,
	}

	if err := repo.Create(ctx, session); err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Update session
	session.MessageCount = 5
	now := time.Now()
	session.LastMessageAt = &now

	if err := repo.Update(ctx, session); err != nil {
		t.Fatalf("Failed to update session: %v", err)
	}

	// Verify update
	found, err := repo.FindByID(ctx, "test-id")
	if err != nil {
		t.Fatalf("Failed to find session: %v", err)
	}

	if found.MessageCount != 5 {
		t.Errorf("Expected message count 5, got %d", found.MessageCount)
	}

	if found.LastMessageAt == nil {
		t.Error("Expected LastMessageAt to be set")
	}
}

func TestMemorySessionRepository_Update_NotFound(t *testing.T) {
	repo := NewMemorySessionRepository()
	defer repo.Close()
	ctx := context.Background()

	session := &entities.Session{
		ID:           "nonexistent",
		CreatedAt:    time.Now(),
		MessageCount: 0,
	}

	err := repo.Update(ctx, session)
	if err == nil {
		t.Error("Expected error when updating nonexistent session, got nil")
	}
}

func TestMemorySessionRepository_Delete(t *testing.T) {
	repo := NewMemorySessionRepository()
	defer repo.Close()
	ctx := context.Background()

	session := &entities.Session{
		ID:           "test-id",
		CreatedAt:    time.Now(),
		MessageCount: 0,
	}

	if err := repo.Create(ctx, session); err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Delete session
	if err := repo.Delete(ctx, "test-id"); err != nil {
		t.Fatalf("Failed to delete session: %v", err)
	}

	// Verify deletion
	_, err := repo.FindByID(ctx, "test-id")
	if err == nil {
		t.Error("Expected error when finding deleted session, got nil")
	}
}

func TestMemorySessionRepository_Delete_NotFound(t *testing.T) {
	repo := NewMemorySessionRepository()
	defer repo.Close()
	ctx := context.Background()

	err := repo.Delete(ctx, "nonexistent")
	if err == nil {
		t.Error("Expected error when deleting nonexistent session, got nil")
	}
}

func TestMemorySessionRepository_Concurrent(t *testing.T) {
	repo := NewMemorySessionRepository()
	defer repo.Close()
	ctx := context.Background()

	// Test concurrent writes
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			session := &entities.Session{
				ID:           string(rune('a' + id)),
				CreatedAt:    time.Now(),
				MessageCount: 0,
			}
			if err := repo.Create(ctx, session); err != nil {
				t.Errorf("Failed to create session: %v", err)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all sessions were created
	sessions, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("Failed to list sessions: %v", err)
	}

	if len(sessions) != 10 {
		t.Errorf("Expected 10 sessions, got %d", len(sessions))
	}
}

func TestMemorySessionRepository_AddMessage(t *testing.T) {
	repo := NewMemorySessionRepository()
	defer repo.Close()
	ctx := context.Background()

	// Create a session
	session := &entities.Session{
		ID:           "test-session",
		CreatedAt:    time.Now(),
		MessageCount: 0,
	}
	if err := repo.Create(ctx, session); err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Add a message
	message := &entities.Message{
		ID:        "msg-1",
		SessionID: "test-session",
		Role:      entities.RoleUser,
		Content:   "Hello",
		Timestamp: time.Now(),
		Status:    entities.StatusSent,
	}

	if err := repo.AddMessage(ctx, message); err != nil {
		t.Fatalf("Failed to add message: %v", err)
	}

	// Verify session was updated
	updatedSession, err := repo.FindByID(ctx, "test-session")
	if err != nil {
		t.Fatalf("Failed to find session: %v", err)
	}

	if updatedSession.MessageCount != 1 {
		t.Errorf("Expected message count 1, got %d", updatedSession.MessageCount)
	}

	if updatedSession.LastMessageAt == nil {
		t.Error("Expected LastMessageAt to be set")
	}
}

func TestMemorySessionRepository_AddMessage_SessionNotFound(t *testing.T) {
	repo := NewMemorySessionRepository()
	defer repo.Close()
	ctx := context.Background()

	message := &entities.Message{
		ID:        "msg-1",
		SessionID: "nonexistent",
		Role:      entities.RoleUser,
		Content:   "Hello",
		Timestamp: time.Now(),
		Status:    entities.StatusSent,
	}

	err := repo.AddMessage(ctx, message)
	if err == nil {
		t.Error("Expected error when adding message to nonexistent session, got nil")
	}
}

func TestMemorySessionRepository_GetMessages(t *testing.T) {
	repo := NewMemorySessionRepository()
	defer repo.Close()
	ctx := context.Background()

	// Create a session
	session := &entities.Session{
		ID:           "test-session",
		CreatedAt:    time.Now(),
		MessageCount: 0,
	}
	if err := repo.Create(ctx, session); err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Add multiple messages
	for i := 0; i < 3; i++ {
		message := &entities.Message{
			ID:        fmt.Sprintf("msg-%d", i),
			SessionID: "test-session",
			Role:      entities.RoleUser,
			Content:   fmt.Sprintf("Message %d", i),
			Timestamp: time.Now(),
			Status:    entities.StatusSent,
		}
		if err := repo.AddMessage(ctx, message); err != nil {
			t.Fatalf("Failed to add message: %v", err)
		}
	}

	// Get messages
	messages, err := repo.GetMessages(ctx, "test-session")
	if err != nil {
		t.Fatalf("Failed to get messages: %v", err)
	}

	if len(messages) != 3 {
		t.Errorf("Expected 3 messages, got %d", len(messages))
	}
}

func TestMemorySessionRepository_GetMessages_EmptyHistory(t *testing.T) {
	repo := NewMemorySessionRepository()
	defer repo.Close()
	ctx := context.Background()

	// Create a session without messages
	session := &entities.Session{
		ID:           "test-session",
		CreatedAt:    time.Now(),
		MessageCount: 0,
	}
	if err := repo.Create(ctx, session); err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Get messages
	messages, err := repo.GetMessages(ctx, "test-session")
	if err != nil {
		t.Fatalf("Failed to get messages: %v", err)
	}

	if len(messages) != 0 {
		t.Errorf("Expected 0 messages, got %d", len(messages))
	}
}

func TestMemorySessionRepository_GetMessages_SessionNotFound(t *testing.T) {
	repo := NewMemorySessionRepository()
	defer repo.Close()
	ctx := context.Background()

	_, err := repo.GetMessages(ctx, "nonexistent")
	if err == nil {
		t.Error("Expected error when getting messages for nonexistent session, got nil")
	}
}

func TestMemorySessionRepository_IsExpired(t *testing.T) {
	repo := NewMemorySessionRepository()
	defer repo.Close()

	// Test session that just expired
	expiredTime := time.Now().Add(-31 * time.Minute)
	expiredSession := &entities.Session{
		ID:            "expired",
		CreatedAt:     expiredTime,
		LastMessageAt: &expiredTime,
		MessageCount:  1,
	}

	if !repo.IsExpired(expiredSession) {
		t.Error("Expected session to be expired")
	}

	// Test active session
	activeSession := &entities.Session{
		ID:           "active",
		CreatedAt:    time.Now(),
		MessageCount: 0,
	}

	if repo.IsExpired(activeSession) {
		t.Error("Expected session to be active")
	}

	// Test session with recent activity
	recentTime := time.Now().Add(-5 * time.Minute)
	recentSession := &entities.Session{
		ID:            "recent",
		CreatedAt:     time.Now().Add(-1 * time.Hour),
		LastMessageAt: &recentTime,
		MessageCount:  5,
	}

	if repo.IsExpired(recentSession) {
		t.Error("Expected session with recent activity to be active")
	}
}

func TestMemorySessionRepository_Delete_WithMessages(t *testing.T) {
	repo := NewMemorySessionRepository()
	defer repo.Close()
	ctx := context.Background()

	// Create session with messages
	session := &entities.Session{
		ID:           "test-session",
		CreatedAt:    time.Now(),
		MessageCount: 0,
	}
	if err := repo.Create(ctx, session); err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Add messages
	message := &entities.Message{
		ID:        "msg-1",
		SessionID: "test-session",
		Role:      entities.RoleUser,
		Content:   "Hello",
		Timestamp: time.Now(),
		Status:    entities.StatusSent,
	}
	if err := repo.AddMessage(ctx, message); err != nil {
		t.Fatalf("Failed to add message: %v", err)
	}

	// Delete session
	if err := repo.Delete(ctx, "test-session"); err != nil {
		t.Fatalf("Failed to delete session: %v", err)
	}

	// Verify session and messages are deleted
	_, err := repo.FindByID(ctx, "test-session")
	if err == nil {
		t.Error("Expected error when finding deleted session, got nil")
	}

	_, err = repo.GetMessages(ctx, "test-session")
	if err == nil {
		t.Error("Expected error when getting messages for deleted session, got nil")
	}
}

func TestMemorySessionRepository_CleanupExpiredSessions(t *testing.T) {
	repo := NewMemorySessionRepository()
	repo.cleanupInterval = 100 * time.Millisecond // Speed up for testing
	defer repo.Close()
	ctx := context.Background()

	// Create an expired session
	expiredTime := time.Now().Add(-31 * time.Minute)
	expiredSession := &entities.Session{
		ID:            "expired",
		CreatedAt:     expiredTime,
		LastMessageAt: &expiredTime,
		MessageCount:  1,
	}
	if err := repo.Create(ctx, expiredSession); err != nil {
		t.Fatalf("Failed to create expired session: %v", err)
	}

	// Create an active session
	activeSession := &entities.Session{
		ID:           "active",
		CreatedAt:    time.Now(),
		MessageCount: 0,
	}
	if err := repo.Create(ctx, activeSession); err != nil {
		t.Fatalf("Failed to create active session: %v", err)
	}

	// Wait for cleanup to run
	time.Sleep(200 * time.Millisecond)

	// Verify expired session was removed
	_, err := repo.FindByID(ctx, "expired")
	if err == nil {
		t.Error("Expected expired session to be removed")
	}

	// Verify active session still exists
	_, err = repo.FindByID(ctx, "active")
	if err != nil {
		t.Errorf("Expected active session to exist: %v", err)
	}
}
