package chat

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/bedrock-chat-poc/backend/domain/entities"
	"github.com/bedrock-chat-poc/backend/domain/services"
	"github.com/bedrock-chat-poc/backend/infrastructure/bedrock"
	"github.com/bedrock-chat-poc/backend/infrastructure/repositories"
	"github.com/gorilla/websocket"
)

// TestWebSocketMessageSendingAndReceiving tests end-to-end message transmission
// Requirement 1.1: Message transmission to Bedrock Agent Core
func TestWebSocketMessageSendingAndReceiving(t *testing.T) {
	// Setup
	sessionRepo := repositories.NewMemorySessionRepository()
	streamProcessor := bedrock.NewStreamProcessor(bedrock.DefaultStreamProcessorConfig())
	handler := NewHandler(sessionRepo, nil, streamProcessor)

	// Create a test session
	session := &entities.Session{
		ID:           "test-session-ws-1",
		CreatedAt:    time.Now(),
		MessageCount: 0,
	}
	if err := sessionRepo.Create(context.Background(), session); err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(handler.HandleWebSocket))
	defer server.Close()

	// Convert http:// to ws://
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// Connect WebSocket client
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect WebSocket: %v", err)
	}
	defer ws.Close()

	// Send a message
	messageReq := MessageRequest{
		SessionID: "test-session-ws-1",
		Content:   "Hello, Bedrock!",
	}

	if err := ws.WriteJSON(messageReq); err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}

	// Read response chunks
	receivedChunks := []string{}
	done := false

	// Set read deadline
	ws.SetReadDeadline(time.Now().Add(5 * time.Second))

	for !done {
		var chunk StreamChunk
		if err := ws.ReadJSON(&chunk); err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				break
			}
			t.Fatalf("Failed to read chunk: %v", err)
		}

		switch chunk.Type {
		case "content":
			receivedChunks = append(receivedChunks, chunk.Content)
		case "done":
			done = true
		case "error":
			t.Fatalf("Received error chunk: %s - %s", chunk.Error.Code, chunk.Error.Message)
		}
	}

	// Verify we received content
	if len(receivedChunks) == 0 {
		t.Error("Expected to receive content chunks, got none")
	}

	// Verify session was updated
	updatedSession, err := sessionRepo.FindByID(context.Background(), "test-session-ws-1")
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

// TestWebSocketStreamingResponse tests streaming response handling
// Requirement 2.1: Real-time streaming response display
func TestWebSocketStreamingResponse(t *testing.T) {
	sessionRepo := repositories.NewMemorySessionRepository()
	streamProcessor := bedrock.NewStreamProcessor(bedrock.DefaultStreamProcessorConfig())
	handler := NewHandler(sessionRepo, nil, streamProcessor)

	// Create session
	session := &entities.Session{
		ID:           "test-session-ws-2",
		CreatedAt:    time.Now(),
		MessageCount: 0,
	}
	if err := sessionRepo.Create(context.Background(), session); err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(handler.HandleWebSocket))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// Connect
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer ws.Close()

	// Send message
	messageReq := MessageRequest{
		SessionID: "test-session-ws-2",
		Content:   "Test streaming response",
	}

	if err := ws.WriteJSON(messageReq); err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}

	// Track chunk arrival times to verify streaming
	chunkTimes := []time.Time{}
	receivedContent := ""
	done := false

	ws.SetReadDeadline(time.Now().Add(5 * time.Second))

	for !done {
		var chunk StreamChunk
		if err := ws.ReadJSON(&chunk); err != nil {
			break
		}

		switch chunk.Type {
		case "content":
			chunkTimes = append(chunkTimes, time.Now())
			receivedContent += chunk.Content
		case "done":
			done = true
		case "error":
			t.Fatalf("Received error: %s", chunk.Error.Message)
		}
	}

	// Verify we received multiple chunks (streaming behavior)
	if len(chunkTimes) < 2 {
		t.Errorf("Expected multiple chunks for streaming, got %d", len(chunkTimes))
	}

	// Verify chunks arrived over time (not all at once)
	if len(chunkTimes) >= 2 {
		timeDiff := chunkTimes[len(chunkTimes)-1].Sub(chunkTimes[0])
		if timeDiff < 50*time.Millisecond {
			t.Logf("Warning: Chunks arrived very quickly (%v), may not be true streaming", timeDiff)
		}
	}

	// Verify we received content
	if receivedContent == "" {
		t.Error("Expected to receive content")
	}
}

// TestWebSocketValidation tests input validation
func TestWebSocketValidation(t *testing.T) {
	sessionRepo := repositories.NewMemorySessionRepository()
	streamProcessor := bedrock.NewStreamProcessor(bedrock.DefaultStreamProcessorConfig())
	handler := NewHandler(sessionRepo, nil, streamProcessor)

	// Create session
	session := &entities.Session{
		ID:           "test-session-ws-3",
		CreatedAt:    time.Now(),
		MessageCount: 0,
	}
	if err := sessionRepo.Create(context.Background(), session); err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(handler.HandleWebSocket))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	tests := []struct {
		name        string
		request     MessageRequest
		expectError bool
		errorCode   string
	}{
		{
			name: "empty content",
			request: MessageRequest{
				SessionID: "test-session-ws-3",
				Content:   "",
			},
			expectError: true,
			errorCode:   "INVALID_REQUEST",
		},
		{
			name: "whitespace only content",
			request: MessageRequest{
				SessionID: "test-session-ws-3",
				Content:   "   \t\n  ",
			},
			expectError: true,
			errorCode:   "INVALID_REQUEST",
		},
		{
			name: "content too long",
			request: MessageRequest{
				SessionID: "test-session-ws-3",
				Content:   strings.Repeat("a", 2001),
			},
			expectError: true,
			errorCode:   "INVALID_REQUEST",
		},
		{
			name: "missing session ID",
			request: MessageRequest{
				SessionID: "",
				Content:   "Hello",
			},
			expectError: true,
			errorCode:   "INVALID_REQUEST",
		},
		{
			name: "valid request",
			request: MessageRequest{
				SessionID: "test-session-ws-3",
				Content:   "Valid message",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Connect
			ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
			if err != nil {
				t.Fatalf("Failed to connect: %v", err)
			}
			defer ws.Close()

			// Send request
			if err := ws.WriteJSON(tt.request); err != nil {
				t.Fatalf("Failed to send message: %v", err)
			}

			// Read response
			ws.SetReadDeadline(time.Now().Add(2 * time.Second))

			var chunk StreamChunk
			if err := ws.ReadJSON(&chunk); err != nil {
				if !tt.expectError {
					t.Fatalf("Failed to read response: %v", err)
				}
				return
			}

			if tt.expectError {
				if chunk.Type != "error" {
					t.Errorf("Expected error chunk, got type: %s", chunk.Type)
				}
				if chunk.Error != nil && chunk.Error.Code != tt.errorCode {
					t.Errorf("Expected error code %s, got %s", tt.errorCode, chunk.Error.Code)
				}
			} else {
				if chunk.Type == "error" {
					t.Errorf("Unexpected error: %s - %s", chunk.Error.Code, chunk.Error.Message)
				}
			}
		})
	}
}

// TestWebSocketSessionNotFound tests session validation
// Requirement 7.1: Session management
func TestWebSocketSessionNotFound(t *testing.T) {
	sessionRepo := repositories.NewMemorySessionRepository()
	streamProcessor := bedrock.NewStreamProcessor(bedrock.DefaultStreamProcessorConfig())
	handler := NewHandler(sessionRepo, nil, streamProcessor)

	server := httptest.NewServer(http.HandlerFunc(handler.HandleWebSocket))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// Connect
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer ws.Close()

	// Send message with non-existent session
	messageReq := MessageRequest{
		SessionID: "nonexistent-session",
		Content:   "Hello",
	}

	if err := ws.WriteJSON(messageReq); err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}

	// Read response
	ws.SetReadDeadline(time.Now().Add(2 * time.Second))

	var chunk StreamChunk
	if err := ws.ReadJSON(&chunk); err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	// Verify error response
	if chunk.Type != "error" {
		t.Errorf("Expected error chunk, got type: %s", chunk.Type)
	}

	if chunk.Error == nil {
		t.Fatal("Expected error details")
	}

	if chunk.Error.Code != "SESSION_NOT_FOUND" {
		t.Errorf("Expected error code SESSION_NOT_FOUND, got %s", chunk.Error.Code)
	}
}

// TestWebSocketMultipleMessages tests sending multiple messages in sequence
func TestWebSocketMultipleMessages(t *testing.T) {
	sessionRepo := repositories.NewMemorySessionRepository()
	streamProcessor := bedrock.NewStreamProcessor(bedrock.DefaultStreamProcessorConfig())
	handler := NewHandler(sessionRepo, nil, streamProcessor)

	// Create session
	session := &entities.Session{
		ID:           "test-session-ws-4",
		CreatedAt:    time.Now(),
		MessageCount: 0,
	}
	if err := sessionRepo.Create(context.Background(), session); err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(handler.HandleWebSocket))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// Connect
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer ws.Close()

	messages := []string{"First message", "Second message", "Third message"}

	for i, msg := range messages {
		// Send message
		messageReq := MessageRequest{
			SessionID: "test-session-ws-4",
			Content:   msg,
		}

		if err := ws.WriteJSON(messageReq); err != nil {
			t.Fatalf("Failed to send message %d: %v", i+1, err)
		}

		// Read response until done
		done := false
		ws.SetReadDeadline(time.Now().Add(5 * time.Second))

		for !done {
			var chunk StreamChunk
			if err := ws.ReadJSON(&chunk); err != nil {
				t.Fatalf("Failed to read chunk for message %d: %v", i+1, err)
			}

			if chunk.Type == "done" {
				done = true
			} else if chunk.Type == "error" {
				t.Fatalf("Received error for message %d: %s", i+1, chunk.Error.Message)
			}
		}
	}

	// Verify session message count
	updatedSession, err := sessionRepo.FindByID(context.Background(), "test-session-ws-4")
	if err != nil {
		t.Fatalf("Failed to find session: %v", err)
	}

	if updatedSession.MessageCount != len(messages) {
		t.Errorf("Expected message count %d, got %d", len(messages), updatedSession.MessageCount)
	}
}

// TestWebSocketConnectionClose tests graceful connection closure
func TestWebSocketConnectionClose(t *testing.T) {
	sessionRepo := repositories.NewMemorySessionRepository()
	streamProcessor := bedrock.NewStreamProcessor(bedrock.DefaultStreamProcessorConfig())
	handler := NewHandler(sessionRepo, nil, streamProcessor)

	// Create session
	session := &entities.Session{
		ID:           "test-session-ws-5",
		CreatedAt:    time.Now(),
		MessageCount: 0,
	}
	if err := sessionRepo.Create(context.Background(), session); err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(handler.HandleWebSocket))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// Connect
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	// Send a message
	messageReq := MessageRequest{
		SessionID: "test-session-ws-5",
		Content:   "Test message",
	}

	if err := ws.WriteJSON(messageReq); err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}

	// Close connection immediately
	if err := ws.Close(); err != nil {
		t.Fatalf("Failed to close connection: %v", err)
	}

	// Verify no panic or error on server side
	// (If server crashes, test will fail)
	time.Sleep(100 * time.Millisecond)
}

// TestWebSocketConcurrentConnections tests multiple concurrent WebSocket connections
func TestWebSocketConcurrentConnections(t *testing.T) {
	sessionRepo := repositories.NewMemorySessionRepository()
	streamProcessor := bedrock.NewStreamProcessor(bedrock.DefaultStreamProcessorConfig())
	handler := NewHandler(sessionRepo, nil, streamProcessor)

	server := httptest.NewServer(http.HandlerFunc(handler.HandleWebSocket))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// Create multiple sessions and connections
	numConnections := 5
	done := make(chan bool, numConnections)

	for i := 0; i < numConnections; i++ {
		sessionID := string(rune('a' + i))

		// Create session
		session := &entities.Session{
			ID:           sessionID,
			CreatedAt:    time.Now(),
			MessageCount: 0,
		}
		if err := sessionRepo.Create(context.Background(), session); err != nil {
			t.Fatalf("Failed to create session %s: %v", sessionID, err)
		}

		go func(sid string) {
			defer func() { done <- true }()

			// Connect
			ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
			if err != nil {
				t.Errorf("Failed to connect for session %s: %v", sid, err)
				return
			}
			defer ws.Close()

			// Send message
			messageReq := MessageRequest{
				SessionID: sid,
				Content:   "Concurrent test message",
			}

			if err := ws.WriteJSON(messageReq); err != nil {
				t.Errorf("Failed to send message for session %s: %v", sid, err)
				return
			}

			// Read response
			ws.SetReadDeadline(time.Now().Add(5 * time.Second))
			receivedDone := false

			for !receivedDone {
				var chunk StreamChunk
				if err := ws.ReadJSON(&chunk); err != nil {
					break
				}

				if chunk.Type == "done" {
					receivedDone = true
				} else if chunk.Type == "error" {
					t.Errorf("Received error for session %s: %s", sid, chunk.Error.Message)
					return
				}
			}
		}(sessionID)
	}

	// Wait for all connections to complete
	for i := 0; i < numConnections; i++ {
		select {
		case <-done:
			// Connection completed
		case <-time.After(10 * time.Second):
			t.Fatal("Timeout waiting for concurrent connections")
		}
	}

	// Verify all sessions were updated
	for i := 0; i < numConnections; i++ {
		sessionID := string(rune('a' + i))
		session, err := sessionRepo.FindByID(context.Background(), sessionID)
		if err != nil {
			t.Errorf("Failed to find session %s: %v", sessionID, err)
			continue
		}

		if session.MessageCount != 1 {
			t.Errorf("Expected message count 1 for session %s, got %d", sessionID, session.MessageCount)
		}
	}
}

// TestWebSocketMalformedJSON tests handling of malformed JSON
func TestWebSocketMalformedJSON(t *testing.T) {
	sessionRepo := repositories.NewMemorySessionRepository()
	streamProcessor := bedrock.NewStreamProcessor(bedrock.DefaultStreamProcessorConfig())
	handler := NewHandler(sessionRepo, nil, streamProcessor)

	server := httptest.NewServer(http.HandlerFunc(handler.HandleWebSocket))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// Connect
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer ws.Close()

	// Send malformed JSON
	if err := ws.WriteMessage(websocket.TextMessage, []byte("invalid json{")); err != nil {
		t.Fatalf("Failed to send malformed message: %v", err)
	}

	// Connection should close due to malformed JSON
	ws.SetReadDeadline(time.Now().Add(2 * time.Second))

	_, _, err = ws.ReadMessage()
	if err == nil {
		t.Error("Expected connection to close after malformed JSON")
	}
}

// MockBedrockService for testing with Bedrock integration
type MockBedrockService struct {
	shouldError bool
	errorCode   string
	errorMsg    string
}

func (m *MockBedrockService) InvokeAgent(ctx context.Context, input services.AgentInput) (*services.AgentResponse, error) {
	if m.shouldError {
		return nil, &services.DomainError{
			Code:    m.errorCode,
			Message: m.errorMsg,
		}
	}

	return &services.AgentResponse{
		Content:   "Mock response",
		Citations: []entities.Citation{},
		Metadata:  map[string]interface{}{},
	}, nil
}

func (m *MockBedrockService) InvokeAgentStream(ctx context.Context, input services.AgentInput) (services.StreamReader, error) {
	if m.shouldError {
		return nil, &services.DomainError{
			Code:    m.errorCode,
			Message: m.errorMsg,
		}
	}

	return &MockStreamReader{
		chunks: []string{"Mock ", "streaming ", "response"},
		index:  0,
	}, nil
}

type MockStreamReader struct {
	chunks []string
	index  int
}

func (m *MockStreamReader) Read() (chunk string, done bool, err error) {
	if m.index >= len(m.chunks) {
		return "", true, nil
	}

	chunk = m.chunks[m.index]
	m.index++
	return chunk, false, nil
}

func (m *MockStreamReader) ReadCitation() (*entities.Citation, error) {
	// No citations in mock
	return nil, nil
}

func (m *MockStreamReader) Close() error {
	return nil
}

// TestWebSocketWithBedrockService tests integration with Bedrock service
func TestWebSocketWithBedrockService(t *testing.T) {
	sessionRepo := repositories.NewMemorySessionRepository()
	streamProcessor := bedrock.NewStreamProcessor(bedrock.DefaultStreamProcessorConfig())
	mockBedrock := &MockBedrockService{}
	handler := NewHandler(sessionRepo, mockBedrock, streamProcessor)

	// Create session
	session := &entities.Session{
		ID:           "test-session-ws-6",
		CreatedAt:    time.Now(),
		MessageCount: 0,
	}
	if err := sessionRepo.Create(context.Background(), session); err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(handler.HandleWebSocket))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// Connect
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer ws.Close()

	// Send message
	messageReq := MessageRequest{
		SessionID: "test-session-ws-6",
		Content:   "Test with Bedrock",
	}

	if err := ws.WriteJSON(messageReq); err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}

	// Read response
	receivedContent := ""
	done := false
	ws.SetReadDeadline(time.Now().Add(5 * time.Second))

	for !done {
		var chunk StreamChunk
		if err := ws.ReadJSON(&chunk); err != nil {
			break
		}

		switch chunk.Type {
		case "content":
			receivedContent += chunk.Content
		case "done":
			done = true
		case "error":
			t.Fatalf("Received error: %s", chunk.Error.Message)
		}
	}

	// Verify we received the mock response
	if receivedContent != "Mock streaming response" {
		t.Errorf("Expected 'Mock streaming response', got '%s'", receivedContent)
	}
}

// TestWebSocketBedrockError tests error handling from Bedrock service
func TestWebSocketBedrockError(t *testing.T) {
	sessionRepo := repositories.NewMemorySessionRepository()
	streamProcessor := bedrock.NewStreamProcessor(bedrock.DefaultStreamProcessorConfig())
	mockBedrock := &MockBedrockService{
		shouldError: true,
		errorCode:   services.ErrCodeRateLimit,
		errorMsg:    "Rate limit exceeded",
	}
	handler := NewHandler(sessionRepo, mockBedrock, streamProcessor)

	// Create session
	session := &entities.Session{
		ID:           "test-session-ws-7",
		CreatedAt:    time.Now(),
		MessageCount: 0,
	}
	if err := sessionRepo.Create(context.Background(), session); err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(handler.HandleWebSocket))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// Connect
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer ws.Close()

	// Send message
	messageReq := MessageRequest{
		SessionID: "test-session-ws-7",
		Content:   "Test Bedrock error",
	}

	if err := ws.WriteJSON(messageReq); err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}

	// Read response
	ws.SetReadDeadline(time.Now().Add(2 * time.Second))

	var chunk StreamChunk
	if err := ws.ReadJSON(&chunk); err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	// Verify error response
	if chunk.Type != "error" {
		t.Errorf("Expected error chunk, got type: %s", chunk.Type)
	}

	if chunk.Error == nil {
		t.Fatal("Expected error details")
	}

	if chunk.Error.Code != services.ErrCodeRateLimit {
		t.Errorf("Expected error code %s, got %s", services.ErrCodeRateLimit, chunk.Error.Code)
	}
}
