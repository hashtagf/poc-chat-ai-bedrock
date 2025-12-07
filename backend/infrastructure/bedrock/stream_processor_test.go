package bedrock

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/bedrock-chat-poc/backend/domain/entities"
	"github.com/bedrock-chat-poc/backend/domain/services"
)

// mockStreamReader implements services.StreamReader for testing
type mockStreamReader struct {
	chunks    []string
	citations []*entities.Citation
	errors    []error
	index     int
	closed    bool
	hangAfter int // Hang after this many reads (-1 = never hang)
}

func (m *mockStreamReader) Read() (string, bool, error) {
	// Check if we should hang
	if m.hangAfter >= 0 && m.index >= m.hangAfter {
		// Simulate hanging by blocking forever
		select {}
	}

	if m.index >= len(m.chunks) {
		return "", true, nil
	}

	chunk := m.chunks[m.index]
	var err error
	if m.index < len(m.errors) {
		err = m.errors[m.index]
	}
	m.index++

	if err != nil {
		return "", false, err
	}

	return chunk, false, nil
}

func (m *mockStreamReader) ReadCitation() (*entities.Citation, error) {
	if len(m.citations) == 0 {
		return nil, nil
	}
	citation := m.citations[0]
	m.citations = m.citations[1:]
	return citation, nil
}

func (m *mockStreamReader) Close() error {
	m.closed = true
	return nil
}

// mockChunkWriter implements ChunkWriter for testing
type mockChunkWriter struct {
	contentChunks  []string
	citationChunks []CitationChunk
	errorChunks    []struct{ code, message string }
	doneWritten    bool
}

func (m *mockChunkWriter) WriteContentChunk(content string) error {
	m.contentChunks = append(m.contentChunks, content)
	return nil
}

func (m *mockChunkWriter) WriteCitationChunk(citation CitationChunk) error {
	m.citationChunks = append(m.citationChunks, citation)
	return nil
}

func (m *mockChunkWriter) WriteErrorChunk(code, message string) error {
	m.errorChunks = append(m.errorChunks, struct{ code, message string }{code, message})
	return nil
}

func (m *mockChunkWriter) WriteDoneChunk() error {
	m.doneWritten = true
	return nil
}

func TestStreamProcessor_ProcessStream_Success(t *testing.T) {
	// Create mock reader with chunks
	reader := &mockStreamReader{
		chunks:    []string{"Hello ", "world", "!"},
		hangAfter: -1,
	}

	writer := &mockChunkWriter{}

	config := StreamProcessorConfig{
		StreamTimeout: 1 * time.Second,
		ChunkTimeout:  500 * time.Millisecond,
	}
	processor := NewStreamProcessor(config)

	ctx := context.Background()
	err := processor.ProcessStream(ctx, reader, writer)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify all chunks were written
	if len(writer.contentChunks) != 3 {
		t.Errorf("Expected 3 content chunks, got %d", len(writer.contentChunks))
	}

	expectedChunks := []string{"Hello ", "world", "!"}
	for i, expected := range expectedChunks {
		if i >= len(writer.contentChunks) {
			t.Errorf("Missing chunk at index %d", i)
			continue
		}
		if writer.contentChunks[i] != expected {
			t.Errorf("Chunk %d: expected %q, got %q", i, expected, writer.contentChunks[i])
		}
	}

	// Verify done was written
	if !writer.doneWritten {
		t.Error("Expected done chunk to be written")
	}

	// Verify stream was closed
	if !reader.closed {
		t.Error("Expected stream to be closed")
	}
}

func TestStreamProcessor_ProcessStream_WithCitations(t *testing.T) {
	citation := &entities.Citation{
		SourceID:   "source-1",
		SourceName: "Test Source",
		Excerpt:    "Test excerpt",
		Confidence: 0.95,
		URL:        "https://example.com",
		Metadata:   map[string]interface{}{"key": "value"},
	}

	reader := &mockStreamReader{
		chunks:    []string{"Content with citation"},
		citations: []*entities.Citation{citation},
		hangAfter: -1,
	}

	writer := &mockChunkWriter{}

	config := DefaultStreamProcessorConfig()
	processor := NewStreamProcessor(config)

	ctx := context.Background()
	err := processor.ProcessStream(ctx, reader, writer)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify citation was written
	if len(writer.citationChunks) != 1 {
		t.Fatalf("Expected 1 citation chunk, got %d", len(writer.citationChunks))
	}

	citationChunk := writer.citationChunks[0]
	if citationChunk.SourceID != citation.SourceID {
		t.Errorf("Expected SourceID %q, got %q", citation.SourceID, citationChunk.SourceID)
	}
	if citationChunk.SourceName != citation.SourceName {
		t.Errorf("Expected SourceName %q, got %q", citation.SourceName, citationChunk.SourceName)
	}
	if citationChunk.Confidence != citation.Confidence {
		t.Errorf("Expected Confidence %f, got %f", citation.Confidence, citationChunk.Confidence)
	}
}

func TestStreamProcessor_ProcessStream_MalformedChunk(t *testing.T) {
	// Create reader that returns a malformed stream error
	malformedErr := &services.DomainError{
		Code:    services.ErrCodeMalformedStream,
		Message: "Malformed chunk",
	}

	reader := &mockStreamReader{
		chunks:    []string{"Good chunk", "", "Another good chunk"},
		errors:    []error{nil, malformedErr, nil},
		hangAfter: -1,
	}

	writer := &mockChunkWriter{}

	config := DefaultStreamProcessorConfig()
	processor := NewStreamProcessor(config)

	ctx := context.Background()
	err := processor.ProcessStream(ctx, reader, writer)

	// Should complete successfully, skipping the malformed chunk
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Should have written the good chunks
	if len(writer.contentChunks) != 2 {
		t.Errorf("Expected 2 content chunks, got %d", len(writer.contentChunks))
	}

	if writer.contentChunks[0] != "Good chunk" {
		t.Errorf("Expected first chunk to be 'Good chunk', got %q", writer.contentChunks[0])
	}
	if writer.contentChunks[1] != "Another good chunk" {
		t.Errorf("Expected second chunk to be 'Another good chunk', got %q", writer.contentChunks[1])
	}
}

func TestStreamProcessor_ProcessStream_StreamTimeout(t *testing.T) {
	// Create reader that hangs immediately
	reader := &mockStreamReader{
		chunks:    []string{},
		hangAfter: 0, // Hang on first read
	}

	writer := &mockChunkWriter{}

	config := StreamProcessorConfig{
		StreamTimeout: 100 * time.Millisecond,
		ChunkTimeout:  50 * time.Millisecond,
	}
	processor := NewStreamProcessor(config)

	ctx := context.Background()
	err := processor.ProcessStream(ctx, reader, writer)

	// Should timeout
	if err == nil {
		t.Fatal("Expected timeout error, got nil")
	}

	// When no content is received, it returns the raw context error
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Expected context.DeadlineExceeded, got: %v", err)
	}

	// Should have written error chunk
	if len(writer.errorChunks) == 0 {
		t.Error("Expected error chunk to be written")
	}
}

func TestStreamProcessor_ProcessStream_ChunkTimeout(t *testing.T) {
	// Create a reader that simulates a stalled stream
	reader := &mockStreamReader{
		chunks:    []string{"First chunk"},
		hangAfter: 1, // Hang after first chunk
	}

	writer := &mockChunkWriter{}

	config := StreamProcessorConfig{
		StreamTimeout: 1 * time.Second,
		ChunkTimeout:  100 * time.Millisecond,
	}
	processor := NewStreamProcessor(config)

	ctx := context.Background()
	
	// This should timeout waiting for the second chunk
	err := processor.ProcessStream(ctx, reader, writer)

	if err == nil {
		t.Fatal("Expected timeout error, got nil")
	}

	// Should have written the first chunk
	if len(writer.contentChunks) != 1 {
		t.Errorf("Expected 1 content chunk, got %d", len(writer.contentChunks))
	}

	// Should have written error chunk for stalled stream
	if len(writer.errorChunks) == 0 {
		t.Error("Expected error chunk to be written for stalled stream")
	}
}

func TestStreamProcessor_ProcessStream_ContextCancellation(t *testing.T) {
	reader := &mockStreamReader{
		chunks:    []string{"Chunk 1", "Chunk 2", "Chunk 3"},
		hangAfter: -1,
	}

	writer := &mockChunkWriter{}

	config := DefaultStreamProcessorConfig()
	processor := NewStreamProcessor(config)

	// Create context that we'll cancel
	ctx, cancel := context.WithCancel(context.Background())
	
	// Cancel immediately
	cancel()

	err := processor.ProcessStream(ctx, reader, writer)

	// Should return context cancelled error
	if err == nil {
		t.Fatal("Expected context cancelled error, got nil")
	}

	if !errors.Is(err, context.Canceled) {
		t.Errorf("Expected context.Canceled error, got: %v", err)
	}
}

func TestStreamProcessor_ProcessStream_EmptyStream(t *testing.T) {
	// Create reader with no chunks
	reader := &mockStreamReader{
		chunks:    []string{},
		hangAfter: -1,
	}

	writer := &mockChunkWriter{}

	config := DefaultStreamProcessorConfig()
	processor := NewStreamProcessor(config)

	ctx := context.Background()
	err := processor.ProcessStream(ctx, reader, writer)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Should have written done chunk
	if !writer.doneWritten {
		t.Error("Expected done chunk to be written")
	}

	// Should not have written any content
	if len(writer.contentChunks) != 0 {
		t.Errorf("Expected 0 content chunks, got %d", len(writer.contentChunks))
	}
}

func TestStreamProcessor_ProcessStream_NonRetryableError(t *testing.T) {
	// Create reader that returns a non-retryable error
	serviceErr := &services.DomainError{
		Code:      services.ErrCodeServiceError,
		Message:   "Service error",
		Retryable: false,
	}

	reader := &mockStreamReader{
		chunks:    []string{""},
		errors:    []error{serviceErr},
		hangAfter: -1,
	}

	writer := &mockChunkWriter{}

	config := DefaultStreamProcessorConfig()
	processor := NewStreamProcessor(config)

	ctx := context.Background()
	err := processor.ProcessStream(ctx, reader, writer)

	// Should return the error
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	// Should have written error chunk
	if len(writer.errorChunks) == 0 {
		t.Error("Expected error chunk to be written")
	}
}

func TestValidateChunk(t *testing.T) {
	tests := []struct {
		name      string
		chunk     string
		wantError bool
	}{
		{
			name:      "valid chunk",
			chunk:     "Hello, world!",
			wantError: false,
		},
		{
			name:      "valid unicode",
			chunk:     "Hello 世界 🌍",
			wantError: false,
		},
		{
			name:      "empty chunk",
			chunk:     "",
			wantError: false,
		},
		{
			name:      "chunk with newlines",
			chunk:     "Line 1\nLine 2\n",
			wantError: false,
		},
		{
			name:      "chunk with null byte",
			chunk:     "Hello\x00World",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateChunk(tt.chunk)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateChunk() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}
