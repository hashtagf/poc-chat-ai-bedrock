package chat

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bedrock-chat-poc/backend/domain/entities"
	"github.com/bedrock-chat-poc/backend/infrastructure/bedrock"
	"github.com/bedrock-chat-poc/backend/infrastructure/repositories"
)

func TestHandleCreateSession(t *testing.T) {
	sessionRepo := repositories.NewMemorySessionRepository()
	streamProcessor := bedrock.NewStreamProcessor(bedrock.DefaultStreamProcessorConfig())
	handler := NewHandler(sessionRepo, nil, streamProcessor)

	req := httptest.NewRequest(http.MethodPost, "/api/sessions", nil)
	w := httptest.NewRecorder()

	handler.HandleCreateSession(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}

	var response SessionResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.ID == "" {
		t.Error("Expected session ID to be set")
	}

	if response.MessageCount != 0 {
		t.Errorf("Expected message count 0, got %d", response.MessageCount)
	}
}

func TestHandleGetSession(t *testing.T) {
	sessionRepo := repositories.NewMemorySessionRepository()
	streamProcessor := bedrock.NewStreamProcessor(bedrock.DefaultStreamProcessorConfig())
	handler := NewHandler(sessionRepo, nil, streamProcessor)

	// Create a session first
	session := &entities.Session{
		ID:           "test-session-id",
		CreatedAt:    time.Now(),
		MessageCount: 5,
	}
	if err := sessionRepo.Create(context.Background(), session); err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/sessions/test-session-id", nil)
	w := httptest.NewRecorder()

	handler.HandleGetSession(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response SessionResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.ID != "test-session-id" {
		t.Errorf("Expected session ID 'test-session-id', got '%s'", response.ID)
	}

	if response.MessageCount != 5 {
		t.Errorf("Expected message count 5, got %d", response.MessageCount)
	}
}

func TestHandleGetSession_NotFound(t *testing.T) {
	sessionRepo := repositories.NewMemorySessionRepository()
	streamProcessor := bedrock.NewStreamProcessor(bedrock.DefaultStreamProcessorConfig())
	handler := NewHandler(sessionRepo, nil, streamProcessor)

	req := httptest.NewRequest(http.MethodGet, "/api/sessions/nonexistent", nil)
	w := httptest.NewRecorder()

	handler.HandleGetSession(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}

	var response ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Code != "SESSION_NOT_FOUND" {
		t.Errorf("Expected error code 'SESSION_NOT_FOUND', got '%s'", response.Code)
	}
}

func TestHandleListSessions(t *testing.T) {
	sessionRepo := repositories.NewMemorySessionRepository()
	streamProcessor := bedrock.NewStreamProcessor(bedrock.DefaultStreamProcessorConfig())
	handler := NewHandler(sessionRepo, nil, streamProcessor)

	// Create multiple sessions
	for i := 0; i < 3; i++ {
		session := &entities.Session{
			ID:           string(rune('a' + i)),
			CreatedAt:    time.Now(),
			MessageCount: i,
		}
		if err := sessionRepo.Create(context.Background(), session); err != nil {
			t.Fatalf("Failed to create session: %v", err)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/api/sessions", nil)
	w := httptest.NewRecorder()

	handler.HandleListSessions(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response []SessionResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(response) != 3 {
		t.Errorf("Expected 3 sessions, got %d", len(response))
	}
}

func TestValidateMessageRequest(t *testing.T) {
	streamProcessor := bedrock.NewStreamProcessor(bedrock.DefaultStreamProcessorConfig())
	handler := NewHandler(nil, nil, streamProcessor)

	tests := []struct {
		name    string
		req     MessageRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: MessageRequest{
				SessionID: "test-session",
				Content:   "Hello, world!",
			},
			wantErr: false,
		},
		{
			name: "empty session ID",
			req: MessageRequest{
				SessionID: "",
				Content:   "Hello",
			},
			wantErr: true,
		},
		{
			name: "empty content",
			req: MessageRequest{
				SessionID: "test-session",
				Content:   "",
			},
			wantErr: true,
		},
		{
			name: "whitespace only content",
			req: MessageRequest{
				SessionID: "test-session",
				Content:   "   \t\n  ",
			},
			wantErr: true,
		},
		{
			name: "content too long",
			req: MessageRequest{
				SessionID: "test-session",
				Content:   string(make([]byte, 2001)),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.validateMessageRequest(&tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateMessageRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHandleCreateSession_MethodNotAllowed(t *testing.T) {
	sessionRepo := repositories.NewMemorySessionRepository()
	streamProcessor := bedrock.NewStreamProcessor(bedrock.DefaultStreamProcessorConfig())
	handler := NewHandler(sessionRepo, nil, streamProcessor)

	req := httptest.NewRequest(http.MethodGet, "/api/sessions", nil)
	w := httptest.NewRecorder()

	handler.HandleCreateSession(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

func TestSetCORSHeaders(t *testing.T) {
	w := httptest.NewRecorder()
	SetCORSHeaders(w)

	if origin := w.Header().Get("Access-Control-Allow-Origin"); origin != "*" {
		t.Errorf("Expected Access-Control-Allow-Origin '*', got '%s'", origin)
	}

	if methods := w.Header().Get("Access-Control-Allow-Methods"); methods == "" {
		t.Error("Expected Access-Control-Allow-Methods to be set")
	}

	if headers := w.Header().Get("Access-Control-Allow-Headers"); headers == "" {
		t.Error("Expected Access-Control-Allow-Headers to be set")
	}
}
