package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/bedrock-chat-poc/backend/config"
	"github.com/bedrock-chat-poc/backend/infrastructure/bedrock"
	"github.com/bedrock-chat-poc/backend/infrastructure/repositories"
	"github.com/bedrock-chat-poc/backend/interfaces/chat"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Log startup information
	log.Printf("Starting chat backend server")
	log.Printf("Environment: %s", cfg.Environment)
	log.Printf("Server: %s:%s", cfg.Server.Host, cfg.Server.Port)
	log.Printf("AWS Region: %s", cfg.AWS.Region)
	log.Printf("Log Level: %s", cfg.Logging.Level)

	// Initialize dependencies
	sessionRepo := repositories.NewMemorySessionRepository()

	// Initialize Bedrock adapter
	var bedrockService *bedrock.Adapter
	if cfg.Bedrock.AgentID != "" && cfg.Bedrock.AgentAliasID != "" {
		bedrockConfig := bedrock.AdapterConfig{
			MaxRetries:     cfg.Bedrock.MaxRetries,
			InitialBackoff: cfg.Bedrock.InitialBackoff,
			MaxBackoff:     cfg.Bedrock.MaxBackoff,
			RequestTimeout: cfg.Bedrock.RequestTimeout,
		}

		bedrockService, err = bedrock.NewAdapter(context.Background(), cfg.Bedrock.AgentID, cfg.Bedrock.AgentAliasID, bedrockConfig)
		if err != nil {
			log.Printf("Warning: Failed to initialize Bedrock adapter: %v", err)
			log.Printf("Running in mock mode without Bedrock integration")
		} else {
			log.Printf("Bedrock adapter initialized")
			log.Printf("  Agent ID: %s", cfg.Bedrock.AgentID)
			log.Printf("  Alias ID: %s", cfg.Bedrock.AgentAliasID)
			if cfg.Bedrock.KnowledgeBaseID != "" {
				log.Printf("  Knowledge Base ID: %s", cfg.Bedrock.KnowledgeBaseID)
			}
			log.Printf("  Model ID: %s", cfg.Bedrock.ModelID)
			log.Printf("  Max Retries: %d", cfg.Bedrock.MaxRetries)
			log.Printf("  Request Timeout: %v", cfg.Bedrock.RequestTimeout)
		}
	} else {
		log.Printf("Bedrock configuration not set, running in mock mode")
		if cfg.IsProduction() {
			log.Fatalf("Bedrock configuration is required in production environment")
		}
	}

	// Initialize stream processor
	streamProcessorConfig := bedrock.StreamProcessorConfig{
		StreamTimeout: cfg.WebSocket.StreamTimeout,
		ChunkTimeout:  cfg.WebSocket.ChunkTimeout,
	}
	streamProcessor := bedrock.NewStreamProcessor(streamProcessorConfig)
	log.Printf("Stream processor initialized")
	log.Printf("  Stream Timeout: %v", cfg.WebSocket.StreamTimeout)
	log.Printf("  Chunk Timeout: %v", cfg.WebSocket.ChunkTimeout)

	// Initialize chat handler with WebSocket configuration
	chatHandler := chat.NewHandlerWithConfig(
		sessionRepo,
		bedrockService,
		streamProcessor,
		chat.HandlerConfig{
			ReadBufferSize:  cfg.WebSocket.ReadBufferSize,
			WriteBufferSize: cfg.WebSocket.WriteBufferSize,
		},
	)

	// Set up routes
	mux := http.NewServeMux()

	// Session management endpoints
	mux.HandleFunc("/api/sessions", func(w http.ResponseWriter, r *http.Request) {
		chat.SetCORSHeaders(w)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.Method == http.MethodPost {
			chatHandler.HandleCreateSession(w, r)
		} else if r.Method == http.MethodGet {
			chatHandler.HandleListSessions(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/sessions/", func(w http.ResponseWriter, r *http.Request) {
		chat.SetCORSHeaders(w)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		chatHandler.HandleGetSession(w, r)
	})

	// WebSocket endpoint for streaming chat
	mux.HandleFunc("/api/chat/stream", func(w http.ResponseWriter, r *http.Request) {
		chat.SetCORSHeaders(w)
		chatHandler.HandleWebSocket(w, r)
	})

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Configuration endpoint (development only)
	if cfg.IsDevelopment() {
		mux.HandleFunc("/api/config", func(w http.ResponseWriter, r *http.Request) {
			chat.SetCORSHeaders(w)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			// Return sanitized configuration (no credentials)
			fmt.Fprintf(w, `{
				"environment": "%s",
				"bedrock_configured": %t,
				"websocket_timeout": "%v",
				"session_timeout": "%v"
			}`, cfg.Environment, bedrockService != nil, cfg.WebSocket.Timeout, cfg.Session.Timeout)
		})
	}

	// Create server with timeouts
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server
	log.Printf("Server listening on %s", server.Addr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
