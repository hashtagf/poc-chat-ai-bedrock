package bedrock

import (
	"bytes"
	"context"
	"errors"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/bedrock-chat-poc/backend/domain/entities"
	"github.com/bedrock-chat-poc/backend/domain/services"
)

// TestStreamProcessorLogging tests logging functionality in the stream processor
// Requirements: 9.1, 9.2, 9.3 - Stream processing events must be logged with proper structure
func TestStreamProcessorLogging(t *testing.T) {
	t.Run("StreamProcessingLogging", testStreamProcessingLogging)
	t.Run("StreamErrorLogging", testStreamErrorLogging)
	t.Run("StreamTimeoutLogging", testStreamTimeoutLogging)
	t.Run("ResourceCleanupLogging", testResourceCleanupLogging)
}

// testStreamProcessingLogging verifies logging during normal stream processing
func testStreamProcessingLogging(t *testing.T) {
	// Capture log output
	var logBuffer bytes.Buffer
	originalOutput := log.Writer()
	log.SetOutput(&logBuffer)
	defer log.SetOutput(originalOutput)

	// Create test stream reader
	reader := &loggingMockStreamReader{
		chunks: []string{"Hello", " world", "!"},
		citations: []*entities.Citation{
			{
				SourceID:   "test-source-1",
				SourceName: "Test Document",
				Excerpt:    "Test excerpt",
			},
		},
	}

	// Create test writer
	writer := &testChunkWriter{}

	// Create stream processor
	processor := NewStreamProcessor(DefaultStreamProcessorConfig())

	// Process stream
	err := processor.ProcessStream(context.Background(), reader, writer)
	if err != nil {
		t.Fatalf("ProcessStream should not error: %v", err)
	}

	logOutput := logBuffer.String()

	// Verify stream completion logging
	if !strings.Contains(logOutput, "[StreamProcessor] Stream completed successfully") {
		t.Error("Log should contain stream completion entry")
	}

	// Verify that no error logs are present for successful processing
	if strings.Contains(logOutput, "[StreamProcessor] Error") {
		t.Error("Log should not contain error entries for successful processing")
	}
	if strings.Contains(logOutput, "[StreamProcessor] Failed") {
		t.Error("Log should not contain failure entries for successful processing")
	}

	t.Logf("✓ Stream processing logging verified - Found completion log")
}

// testStreamErrorLogging verifies error logging in stream processing
func testStreamErrorLogging(t *testing.T) {
	// Capture log output
	var logBuffer bytes.Buffer
	originalOutput := log.Writer()
	log.SetOutput(&logBuffer)
	defer log.SetOutput(originalOutput)

	// Create test stream reader that returns an error
	reader := &loggingMockStreamReader{
		chunks:      []string{"Hello"},
		shouldError: true,
		errorMsg:    "Test stream error",
	}

	// Create test writer
	writer := &testChunkWriter{}

	// Create stream processor
	processor := NewStreamProcessor(DefaultStreamProcessorConfig())

	// Process stream (should fail)
	err := processor.ProcessStream(context.Background(), reader, writer)
	if err == nil {
		t.Error("ProcessStream should return error")
	}

	logOutput := logBuffer.String()

	// Verify error logging
	if !strings.Contains(logOutput, "[StreamProcessor] Stream read error:") {
		t.Error("Log should contain stream read error entry")
	}
	if !strings.Contains(logOutput, "Test stream error") {
		t.Error("Log should contain specific error message")
	}

	t.Logf("✓ Stream error logging verified - Found error log entries")
}

// testStreamTimeoutLogging verifies timeout logging
func testStreamTimeoutLogging(t *testing.T) {
	// Capture log output
	var logBuffer bytes.Buffer
	originalOutput := log.Writer()
	log.SetOutput(&logBuffer)
	defer log.SetOutput(originalOutput)

	// Create test stream reader that hangs
	reader := &loggingMockStreamReader{
		chunks:    []string{"Hello"},
		hangAfter: 1, // Hang after first chunk
	}

	// Create test writer
	writer := &testChunkWriter{}

	// Create stream processor with short timeout
	config := StreamProcessorConfig{
		StreamTimeout: 100 * time.Millisecond,
		ChunkTimeout:  50 * time.Millisecond,
	}
	processor := NewStreamProcessor(config)

	// Process stream (should timeout)
	err := processor.ProcessStream(context.Background(), reader, writer)
	if err == nil {
		t.Error("ProcessStream should return timeout error")
	}

	logOutput := logBuffer.String()

	// Verify timeout logging (could be chunk timeout or stream timeout)
	if !strings.Contains(logOutput, "[StreamProcessor] Stream timeout exceeded") && 
	   !strings.Contains(logOutput, "[StreamProcessor] Chunk timeout") {
		t.Errorf("Log should contain timeout entry. Actual log: %s", logOutput)
	}

	var domainErr *services.DomainError
	if !errors.As(err, &domainErr) {
		t.Error("Error should be a domain error")
	} else if domainErr.Code != services.ErrCodeTimeout {
		t.Errorf("Expected timeout error code, got %s", domainErr.Code)
	}

	t.Logf("✓ Stream timeout logging verified - Found timeout log entries")
}

// testResourceCleanupLogging verifies resource cleanup logging
func testResourceCleanupLogging(t *testing.T) {
	// Capture log output
	var logBuffer bytes.Buffer
	originalOutput := log.Writer()
	log.SetOutput(&logBuffer)
	defer log.SetOutput(originalOutput)

	// Create test stream reader that fails to close
	reader := &loggingMockStreamReader{
		chunks:     []string{"Hello", "world"},
		closeError: errors.New("Failed to close stream"),
	}

	// Create test writer
	writer := &testChunkWriter{}

	// Create stream processor
	processor := NewStreamProcessor(DefaultStreamProcessorConfig())

	// Process stream
	err := processor.ProcessStream(context.Background(), reader, writer)
	if err != nil {
		t.Fatalf("ProcessStream should succeed despite close error: %v", err)
	}

	logOutput := logBuffer.String()

	// Verify cleanup error logging
	if !strings.Contains(logOutput, "[StreamProcessor] Error closing stream:") {
		t.Error("Log should contain stream close error entry")
	}
	if !strings.Contains(logOutput, "Failed to close stream") {
		t.Error("Log should contain specific close error message")
	}

	t.Logf("✓ Resource cleanup logging verified - Found cleanup error log")
}

// loggingMockStreamReader for testing stream processor logging
type loggingMockStreamReader struct {
	chunks      []string
	citations   []*entities.Citation
	currentIdx  int
	shouldError bool
	errorMsg    string
	hangAfter   int
	closeError  error
}

func (m *loggingMockStreamReader) Read() (string, bool, error) {
	if m.shouldError && m.currentIdx > 0 {
		return "", false, errors.New(m.errorMsg)
	}

	if m.hangAfter > 0 && m.currentIdx >= m.hangAfter {
		// Simulate hanging by sleeping longer than timeout
		time.Sleep(200 * time.Millisecond)
	}

	if m.currentIdx >= len(m.chunks) {
		return "", true, nil
	}

	chunk := m.chunks[m.currentIdx]
	m.currentIdx++
	return chunk, false, nil
}

func (m *loggingMockStreamReader) ReadCitation() (*entities.Citation, error) {
	if len(m.citations) > 0 && m.currentIdx <= len(m.citations) {
		return m.citations[0], nil
	}
	return nil, nil
}

func (m *loggingMockStreamReader) Close() error {
	return m.closeError
}

// testChunkWriter for testing
type testChunkWriter struct {
	contentChunks  []string
	citationChunks []CitationChunk
	errorChunks    []errorChunk
	doneReceived   bool
}

type errorChunk struct {
	code    string
	message string
}

func (w *testChunkWriter) WriteContentChunk(content string) error {
	w.contentChunks = append(w.contentChunks, content)
	return nil
}

func (w *testChunkWriter) WriteCitationChunk(citation CitationChunk) error {
	w.citationChunks = append(w.citationChunks, citation)
	return nil
}

func (w *testChunkWriter) WriteErrorChunk(code, message string) error {
	w.errorChunks = append(w.errorChunks, errorChunk{code: code, message: message})
	return nil
}

func (w *testChunkWriter) WriteDoneChunk() error {
	w.doneReceived = true
	return nil
}