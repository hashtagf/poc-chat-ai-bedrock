package bedrock

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/bedrock-chat-poc/backend/domain/services"
	"github.com/gorilla/websocket"
)

// StreamProcessor handles processing of Bedrock streaming responses
// and forwards them to WebSocket connections
type StreamProcessor struct {
	streamTimeout time.Duration
	chunkTimeout  time.Duration
}

// StreamProcessorConfig holds configuration for the stream processor
type StreamProcessorConfig struct {
	// StreamTimeout is the maximum time to wait for the entire stream
	StreamTimeout time.Duration
	// ChunkTimeout is the maximum time to wait between chunks
	ChunkTimeout time.Duration
}

// DefaultStreamProcessorConfig returns default configuration
func DefaultStreamProcessorConfig() StreamProcessorConfig {
	return StreamProcessorConfig{
		StreamTimeout: 5 * time.Minute,
		ChunkTimeout:  30 * time.Second,
	}
}

// NewStreamProcessor creates a new stream processor
func NewStreamProcessor(config StreamProcessorConfig) *StreamProcessor {
	return &StreamProcessor{
		streamTimeout: config.StreamTimeout,
		chunkTimeout:  config.ChunkTimeout,
	}
}

// ChunkWriter defines the interface for writing chunks to a destination
type ChunkWriter interface {
	WriteContentChunk(content string) error
	WriteCitationChunk(citation CitationChunk) error
	WriteErrorChunk(code, message string) error
	WriteDoneChunk() error
}

// CitationChunk represents a citation to be sent over the wire
type CitationChunk struct {
	SourceID   string                 `json:"source_id"`
	SourceName string                 `json:"source_name"`
	Excerpt    string                 `json:"excerpt"`
	Confidence float64                `json:"confidence,omitempty"`
	URL        string                 `json:"url,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// WebSocketChunkWriter implements ChunkWriter for WebSocket connections
type WebSocketChunkWriter struct {
	conn *websocket.Conn
}

// NewWebSocketChunkWriter creates a new WebSocket chunk writer
func NewWebSocketChunkWriter(conn *websocket.Conn) *WebSocketChunkWriter {
	return &WebSocketChunkWriter{conn: conn}
}

// WriteContentChunk writes a content chunk to the WebSocket
func (w *WebSocketChunkWriter) WriteContentChunk(content string) error {
	chunk := map[string]interface{}{
		"type":    "content",
		"content": content,
	}
	return w.conn.WriteJSON(chunk)
}

// WriteCitationChunk writes a citation chunk to the WebSocket
func (w *WebSocketChunkWriter) WriteCitationChunk(citation CitationChunk) error {
	chunk := map[string]interface{}{
		"type":     "citation",
		"citation": citation,
	}
	return w.conn.WriteJSON(chunk)
}

// WriteErrorChunk writes an error chunk to the WebSocket
func (w *WebSocketChunkWriter) WriteErrorChunk(code, message string) error {
	chunk := map[string]interface{}{
		"type": "error",
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	}
	return w.conn.WriteJSON(chunk)
}

// WriteDoneChunk writes a done chunk to the WebSocket
func (w *WebSocketChunkWriter) WriteDoneChunk() error {
	chunk := map[string]interface{}{
		"type": "done",
	}
	return w.conn.WriteJSON(chunk)
}

// ProcessStream processes a streaming response and forwards chunks to the writer
func (sp *StreamProcessor) ProcessStream(ctx context.Context, reader services.StreamReader, writer ChunkWriter) error {
	// Create context with overall stream timeout
	streamCtx, cancel := context.WithTimeout(ctx, sp.streamTimeout)
	defer cancel()

	// Ensure stream is closed when done
	defer func() {
		if err := reader.Close(); err != nil {
			log.Printf("[StreamProcessor] Error closing stream: %v", err)
		}
	}()

	// Track if we've received any content
	receivedContent := false

	// Process chunks in a loop
	for {
		// Check if context is cancelled
		select {
		case <-streamCtx.Done():
			if streamCtx.Err() == context.DeadlineExceeded {
				log.Printf("[StreamProcessor] Stream timeout exceeded")
				if err := writer.WriteErrorChunk(services.ErrCodeTimeout, "Stream timed out"); err != nil {
					log.Printf("[StreamProcessor] Failed to write timeout error: %v", err)
				}
				return &services.DomainError{
					Code:      services.ErrCodeTimeout,
					Message:   "Stream processing timed out",
					Retryable: false,
					Cause:     streamCtx.Err(),
				}
			}
			return streamCtx.Err()
		default:
		}

		// Read next chunk with timeout
		chunkCtx, chunkCancel := context.WithTimeout(streamCtx, sp.chunkTimeout)
		
		chunk, done, err := sp.readChunkWithTimeout(chunkCtx, reader)
		chunkCancel()

		// Handle errors
		if err != nil {
			// Check if it's a timeout waiting for chunk
			if errors.Is(err, context.DeadlineExceeded) {
				log.Printf("[StreamProcessor] Chunk timeout - no data received within %v", sp.chunkTimeout)
				
				// If we've received some content, treat as stalled stream
				if receivedContent {
					if writeErr := writer.WriteErrorChunk(services.ErrCodeTimeout, "Stream stalled"); writeErr != nil {
						log.Printf("[StreamProcessor] Failed to write stall error: %v", writeErr)
					}
					return &services.DomainError{
						Code:      services.ErrCodeTimeout,
						Message:   "Stream stalled - no data received",
						Retryable: false,
						Cause:     err,
					}
				}
			}

			// Handle malformed stream errors
			var domainErr *services.DomainError
			if errors.As(err, &domainErr) {
				if domainErr.Code == services.ErrCodeMalformedStream {
					log.Printf("[StreamProcessor] Malformed stream chunk: %v", err)
					// Try to continue processing - don't fail the entire stream
					continue
				}
			}

			// For other errors, write error chunk and return
			log.Printf("[StreamProcessor] Stream read error: %v", err)
			if writeErr := writer.WriteErrorChunk(services.ErrCodeServiceError, "Error reading stream"); writeErr != nil {
				log.Printf("[StreamProcessor] Failed to write error chunk: %v", writeErr)
			}
			return err
		}

		// If done, break the loop
		if done {
			log.Printf("[StreamProcessor] Stream completed successfully")
			break
		}

		// Process the chunk
		if chunk != "" {
			receivedContent = true
			if err := writer.WriteContentChunk(chunk); err != nil {
				log.Printf("[StreamProcessor] Failed to write content chunk: %v", err)
				return fmt.Errorf("failed to write content chunk: %w", err)
			}
		}

		// Check for citations
		citation, err := reader.ReadCitation()
		if err != nil {
			log.Printf("[StreamProcessor] Error reading citation: %v", err)
			// Don't fail the stream for citation errors, just log
			continue
		}

		if citation != nil {
			citationChunk := CitationChunk{
				SourceID:   citation.SourceID,
				SourceName: citation.SourceName,
				Excerpt:    citation.Excerpt,
				Confidence: citation.Confidence,
				URL:        citation.URL,
				Metadata:   citation.Metadata,
			}

			if err := writer.WriteCitationChunk(citationChunk); err != nil {
				log.Printf("[StreamProcessor] Failed to write citation chunk: %v", err)
				// Don't fail the stream for citation write errors
			}
		}
	}

	// Send done signal
	if err := writer.WriteDoneChunk(); err != nil {
		log.Printf("[StreamProcessor] Failed to write done chunk: %v", err)
		return fmt.Errorf("failed to write done chunk: %w", err)
	}

	return nil
}

// readChunkWithTimeout reads a chunk with a timeout
func (sp *StreamProcessor) readChunkWithTimeout(ctx context.Context, reader services.StreamReader) (string, bool, error) {
	type result struct {
		chunk string
		done  bool
		err   error
	}

	resultChan := make(chan result, 1)

	// Read in a goroutine
	go func() {
		chunk, done, err := reader.Read()
		resultChan <- result{chunk: chunk, done: done, err: err}
	}()

	// Wait for result or timeout
	select {
	case <-ctx.Done():
		return "", false, ctx.Err()
	case res := <-resultChan:
		return res.chunk, res.done, res.err
	}
}

// ValidateChunk validates a chunk for malformed content
// Returns an error if the chunk is malformed
func ValidateChunk(chunk string) error {
	// Basic validation - check for null bytes or other invalid characters
	for i, r := range chunk {
		if r == 0 {
			return fmt.Errorf("chunk contains null byte at position %d", i)
		}
		// Check for invalid UTF-8 sequences (replacement character)
		if r == '\uFFFD' {
			return fmt.Errorf("chunk contains invalid UTF-8 at position %d", i)
		}
	}
	return nil
}
