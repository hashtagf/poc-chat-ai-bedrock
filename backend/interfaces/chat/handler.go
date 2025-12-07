package chat

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/bedrock-chat-poc/backend/domain/entities"
	"github.com/bedrock-chat-poc/backend/domain/repositories"
	"github.com/bedrock-chat-poc/backend/domain/services"
	"github.com/bedrock-chat-poc/backend/infrastructure/bedrock"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// Handler handles HTTP and WebSocket requests for the chat interface
type Handler struct {
	sessionRepo     repositories.SessionRepository
	bedrockService  services.BedrockService
	streamProcessor *bedrock.StreamProcessor
	upgrader        websocket.Upgrader
}

// HandlerConfig holds configuration for the handler
type HandlerConfig struct {
	ReadBufferSize  int
	WriteBufferSize int
}

// NewHandler creates a new chat handler with default configuration
func NewHandler(sessionRepo repositories.SessionRepository, bedrockService services.BedrockService, streamProcessor *bedrock.StreamProcessor) *Handler {
	return NewHandlerWithConfig(sessionRepo, bedrockService, streamProcessor, HandlerConfig{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	})
}

// NewHandlerWithConfig creates a new chat handler with custom configuration
func NewHandlerWithConfig(sessionRepo repositories.SessionRepository, bedrockService services.BedrockService, streamProcessor *bedrock.StreamProcessor, config HandlerConfig) *Handler {
	return &Handler{
		sessionRepo:     sessionRepo,
		bedrockService:  bedrockService,
		streamProcessor: streamProcessor,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  config.ReadBufferSize,
			WriteBufferSize: config.WriteBufferSize,
			CheckOrigin: func(r *http.Request) bool {
				// Allow all origins for POC - in production, restrict this
				return true
			},
		},
	}
}

// HandleCreateSession handles POST /api/sessions
func (h *Handler) HandleCreateSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	ctx := r.Context()

	// Create new session
	session := &entities.Session{
		ID:           uuid.New().String(),
		CreatedAt:    time.Now(),
		MessageCount: 0,
	}

	if err := h.sessionRepo.Create(ctx, session); err != nil {
		log.Printf("Failed to create session: %v", err)
		h.writeError(w, http.StatusInternalServerError, "SESSION_CREATE_FAILED", "Failed to create session")
		return
	}

	response := SessionResponse{
		ID:           session.ID,
		CreatedAt:    session.CreatedAt,
		MessageCount: session.MessageCount,
	}

	h.writeJSON(w, http.StatusCreated, response)
}

// HandleGetSession handles GET /api/sessions/{id}
func (h *Handler) HandleGetSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	// Extract session ID from URL path
	sessionID := strings.TrimPrefix(r.URL.Path, "/api/sessions/")
	if sessionID == "" || sessionID == "/api/sessions/" {
		h.writeError(w, http.StatusBadRequest, "INVALID_SESSION_ID", "Session ID is required")
		return
	}

	ctx := r.Context()
	session, err := h.sessionRepo.FindByID(ctx, sessionID)
	if err != nil {
		log.Printf("Failed to find session %s: %v", sessionID, err)
		h.writeError(w, http.StatusNotFound, "SESSION_NOT_FOUND", "Session not found")
		return
	}

	response := SessionResponse{
		ID:            session.ID,
		CreatedAt:     session.CreatedAt,
		LastMessageAt: session.LastMessageAt,
		MessageCount:  session.MessageCount,
	}

	h.writeJSON(w, http.StatusOK, response)
}

// HandleListSessions handles GET /api/sessions
func (h *Handler) HandleListSessions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	ctx := r.Context()
	sessions, err := h.sessionRepo.List(ctx)
	if err != nil {
		log.Printf("Failed to list sessions: %v", err)
		h.writeError(w, http.StatusInternalServerError, "SESSION_LIST_FAILED", "Failed to list sessions")
		return
	}

	responses := make([]SessionResponse, len(sessions))
	for i, session := range sessions {
		responses[i] = SessionResponse{
			ID:            session.ID,
			CreatedAt:     session.CreatedAt,
			LastMessageAt: session.LastMessageAt,
			MessageCount:  session.MessageCount,
		}
	}

	h.writeJSON(w, http.StatusOK, responses)
}

// HandleWebSocket handles WebSocket connections for streaming chat
func (h *Handler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}
	defer conn.Close()

	log.Printf("WebSocket connection established")

	// Handle messages in a loop
	for {
		var req MessageRequest
		err := conn.ReadJSON(&req)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Validate request
		if err := h.validateMessageRequest(&req); err != nil {
			h.sendErrorChunk(conn, "INVALID_REQUEST", err.Error())
			continue
		}

		// Verify session exists
		ctx := context.Background()
		session, err := h.sessionRepo.FindByID(ctx, req.SessionID)
		if err != nil {
			h.sendErrorChunk(conn, "SESSION_NOT_FOUND", "Session not found")
			continue
		}

		// Process message and stream response
		if err := h.processMessage(ctx, conn, session, &req); err != nil {
			log.Printf("Failed to process message: %v", err)
			h.sendErrorChunk(conn, "PROCESSING_FAILED", "Failed to process message")
		}
	}
}

// processMessage processes a message and streams the response
func (h *Handler) processMessage(ctx context.Context, conn *websocket.Conn, session *entities.Session, req *MessageRequest) error {
	// Update session
	now := time.Now()
	session.LastMessageAt = &now
	session.MessageCount++
	if err := h.sessionRepo.Update(ctx, session); err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	// Check if Bedrock service is available
	if h.bedrockService == nil {
		// Mock mode - simulate streaming response
		return h.processMockMessage(ctx, conn, req)
	}

	// Create agent input
	input := services.AgentInput{
		SessionID: req.SessionID,
		Message:   req.Content,
	}

	// Invoke Bedrock agent with streaming
	streamReader, err := h.bedrockService.InvokeAgentStream(ctx, input)
	if err != nil {
		log.Printf("Failed to invoke Bedrock agent: %v", err)
		
		// Transform error to user-friendly message
		var domainErr *services.DomainError
		if errors.As(err, &domainErr) {
			h.sendErrorChunk(conn, domainErr.Code, domainErr.Message)
		} else {
			h.sendErrorChunk(conn, services.ErrCodeServiceError, "Failed to process message")
		}
		return err
	}

	// Create WebSocket chunk writer
	writer := bedrock.NewWebSocketChunkWriter(conn)

	// Process the stream
	if err := h.streamProcessor.ProcessStream(ctx, streamReader, writer); err != nil {
		log.Printf("Failed to process stream: %v", err)
		return err
	}

	return nil
}

// processMockMessage simulates a streaming response for testing without Bedrock
func (h *Handler) processMockMessage(ctx context.Context, conn *websocket.Conn, req *MessageRequest) error {
	// Simulate streaming response chunks
	responseText := fmt.Sprintf("Echo: %s", req.Content)
	words := strings.Fields(responseText)

	for _, word := range words {
		chunk := StreamChunk{
			Type:    "content",
			Content: word + " ",
		}
		if err := conn.WriteJSON(chunk); err != nil {
			return fmt.Errorf("failed to write chunk: %w", err)
		}
		time.Sleep(100 * time.Millisecond) // Simulate streaming delay
	}

	// Send done signal
	doneChunk := StreamChunk{
		Type: "done",
	}
	if err := conn.WriteJSON(doneChunk); err != nil {
		return fmt.Errorf("failed to write done chunk: %w", err)
	}

	return nil
}

// validateMessageRequest validates the message request
func (h *Handler) validateMessageRequest(req *MessageRequest) error {
	if req.SessionID == "" {
		return fmt.Errorf("session_id is required")
	}

	content := strings.TrimSpace(req.Content)
	if content == "" {
		return fmt.Errorf("content cannot be empty or whitespace only")
	}

	if len(req.Content) > 2000 {
		return fmt.Errorf("content exceeds maximum length of 2000 characters")
	}

	return nil
}

// sendErrorChunk sends an error chunk over WebSocket
func (h *Handler) sendErrorChunk(conn *websocket.Conn, code, message string) {
	chunk := StreamChunk{
		Type: "error",
		Error: &ErrorResponse{
			Code:    code,
			Message: message,
		},
	}
	if err := conn.WriteJSON(chunk); err != nil {
		log.Printf("Failed to send error chunk: %v", err)
	}
}

// writeJSON writes a JSON response
func (h *Handler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Failed to encode JSON response: %v", err)
	}
}

// writeError writes an error response
func (h *Handler) writeError(w http.ResponseWriter, status int, code, message string) {
	response := ErrorResponse{
		Code:    code,
		Message: message,
	}
	h.writeJSON(w, status, response)
}

// SetCORSHeaders sets CORS headers for the response
func SetCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}
