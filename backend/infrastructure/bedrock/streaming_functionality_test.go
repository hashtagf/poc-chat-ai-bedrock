package bedrock

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/bedrockagentruntime"
	"github.com/aws/smithy-go"

	"github.com/bedrock-chat-poc/backend/domain/entities"
	"github.com/bedrock-chat-poc/backend/domain/services"
)

// TestStreamingFunctionality_BasicStreaming tests basic streaming response functionality with real AWS
// Requirements: 3.1 - Verify streaming responses work correctly
func TestStreamingFunctionality_BasicStreaming(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping streaming functionality test in CI environment")
	}

	agentID := os.Getenv("BEDROCK_AGENT_ID")
	aliasID := os.Getenv("BEDROCK_AGENT_ALIAS_ID")

	if agentID == "" || aliasID == "" {
		t.Skip("Skipping streaming test - BEDROCK_AGENT_ID and BEDROCK_AGENT_ALIAS_ID must be set")
	}

	ctx := context.Background()

	adapter, err := NewAdapter(ctx, agentID, aliasID, DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to create Bedrock adapter: %v", err)
	}

	input := services.AgentInput{
		SessionID: generateTestSessionID(),
		Message:   "Please provide a brief response about AI in one sentence.",
	}

	streamReader, err := adapter.InvokeAgentStream(ctx, input)
	if err != nil {
		t.Fatalf("InvokeAgentStream failed: %v", err)
	}
	defer streamReader.Close()

	// Read all chunks
	var content strings.Builder
	chunkCount := 0
	startTime := time.Now()

	for {
		chunk, done, err := streamReader.Read()
		if done {
			break
		}
		if err != nil {
			t.Fatalf("Stream read error: %v", err)
		}

		if chunk != "" {
			content.WriteString(chunk)
			chunkCount++
			t.Logf("Received chunk %d: %q", chunkCount, chunk)
		}
	}

	duration := time.Since(startTime)

	// Verify streaming worked correctly
	if chunkCount == 0 {
		t.Error("Expected to receive at least one chunk")
	}

	if content.Len() == 0 {
		t.Error("Expected to receive content from stream")
	}

	t.Logf("✓ Streaming test completed: %d chunks, %d characters, duration: %v", 
		chunkCount, content.Len(), duration)
}

// TestStreamingFunctionality_StreamCompletion tests stream completion handling
// Requirements: 3.2 - Test stream completion and error handling
func TestStreamingFunctionality_StreamCompletion(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping streaming functionality test in CI environment")
	}

	agentID := os.Getenv("BEDROCK_AGENT_ID")
	aliasID := os.Getenv("BEDROCK_AGENT_ALIAS_ID")

	if agentID == "" || aliasID == "" {
		t.Skip("Skipping streaming test - BEDROCK_AGENT_ID and BEDROCK_AGENT_ALIAS_ID must be set")
	}

	ctx := context.Background()

	adapter, err := NewAdapter(ctx, agentID, aliasID, DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to create Bedrock adapter: %v", err)
	}

	tests := []struct {
		name    string
		message string
	}{
		{
			name:    "short response",
			message: "Say 'Hello'",
		},
		{
			name:    "longer response",
			message: "Please tell me a brief story about technology in 2-3 sentences.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := services.AgentInput{
				SessionID: generateTestSessionID(),
				Message:   tt.message,
			}

			streamReader, err := adapter.InvokeAgentStream(ctx, input)
			if err != nil {
				t.Fatalf("InvokeAgentStream failed: %v", err)
			}
			defer streamReader.Close()

			chunkCount := 0
			var content strings.Builder
			completed := false

			for {
				chunk, done, err := streamReader.Read()
				if done {
					completed = true
					break
				}
				if err != nil {
					t.Fatalf("Stream read error: %v", err)
				}

				if chunk != "" {
					content.WriteString(chunk)
					chunkCount++
				}
			}

			// Verify stream completed properly
			if !completed {
				t.Error("Stream did not complete properly")
			}

			if content.Len() == 0 {
				t.Error("Expected to receive content from stream")
			}

			t.Logf("✓ Stream completion test: %d chunks, %d characters", chunkCount, content.Len())
		})
	}
}

// TestStreamingFunctionality_CitationProcessing tests citation processing in streams
// Requirements: 3.3 - Validate citation processing in streams
func TestStreamingFunctionality_CitationProcessing(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping streaming functionality test in CI environment")
	}

	agentID := os.Getenv("BEDROCK_AGENT_ID")
	aliasID := os.Getenv("BEDROCK_AGENT_ALIAS_ID")
	knowledgeBaseID := os.Getenv("BEDROCK_KNOWLEDGE_BASE_ID")

	if agentID == "" || aliasID == "" {
		t.Skip("Skipping streaming citation test - BEDROCK_AGENT_ID and BEDROCK_AGENT_ALIAS_ID must be set")
	}

	ctx := context.Background()

	adapter, err := NewAdapter(ctx, agentID, aliasID, DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to create Bedrock adapter: %v", err)
	}

	input := services.AgentInput{
		SessionID: generateTestSessionID(),
		Message:   "What information do you have about our company or products?",
	}

	// Add knowledge base ID if available
	if knowledgeBaseID != "" {
		input.KnowledgeBaseIDs = []string{knowledgeBaseID}
	}

	streamReader, err := adapter.InvokeAgentStream(ctx, input)
	if err != nil {
		t.Fatalf("InvokeAgentStream failed: %v", err)
	}
	defer streamReader.Close()

	// Read content and citations
	var citations []*entities.Citation
	contentReceived := false
	chunkCount := 0

	for {
		chunk, done, err := streamReader.Read()
		if done {
			break
		}
		if err != nil {
			t.Fatalf("Stream read error: %v", err)
		}

		if chunk != "" {
			contentReceived = true
			chunkCount++
		}

		// Check for citations
		citation, err := streamReader.ReadCitation()
		if err != nil {
			t.Fatalf("Citation read error: %v", err)
		}
		if citation != nil {
			citations = append(citations, citation)
			t.Logf("Found citation: Excerpt=%q, SourceName=%q, URL=%q", 
				citation.Excerpt, citation.SourceName, citation.URL)
		}
	}

	// Verify content was received
	if !contentReceived {
		t.Error("Expected to receive content")
	}

	// Verify citation processing works (may or may not have citations depending on knowledge base)
	t.Logf("✓ Citation processing test: %d chunks, %d citations", chunkCount, len(citations))

	// If citations were found, verify their structure
	for i, citation := range citations {
		if citation.Excerpt == "" {
			t.Errorf("Citation %d has empty excerpt", i)
		}
		
		// Verify metadata is properly initialized
		if citation.Metadata == nil {
			t.Errorf("Citation %d has nil metadata", i)
		}
	}
}

// TestStreamingFunctionality_ResourceCleanup tests resource cleanup on stream close
// Requirements: 3.6 - Test resource cleanup on stream close
func TestStreamingFunctionality_ResourceCleanup(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping streaming functionality test in CI environment")
	}

	agentID := os.Getenv("BEDROCK_AGENT_ID")
	aliasID := os.Getenv("BEDROCK_AGENT_ALIAS_ID")

	if agentID == "" || aliasID == "" {
		t.Skip("Skipping streaming test - BEDROCK_AGENT_ID and BEDROCK_AGENT_ALIAS_ID must be set")
	}

	ctx := context.Background()

	adapter, err := NewAdapter(ctx, agentID, aliasID, DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to create Bedrock adapter: %v", err)
	}

	input := services.AgentInput{
		SessionID: generateTestSessionID(),
		Message:   "Say hello",
	}

	streamReader, err := adapter.InvokeAgentStream(ctx, input)
	if err != nil {
		t.Fatalf("InvokeAgentStream failed: %v", err)
	}

	// Read one chunk
	chunk, done, err := streamReader.Read()
	if err != nil {
		t.Fatalf("Stream read error: %v", err)
	}
	
	receivedContent := chunk != "" || done

	// Close the stream early (before reading all content)
	err = streamReader.Close()
	if err != nil {
		t.Errorf("Stream close error: %v", err)
	}

	// Verify stream is marked as done after close
	_, done, err = streamReader.Read()
	if !done {
		t.Error("Expected stream to be done after close")
	}

	// Multiple closes should not cause errors
	err = streamReader.Close()
	if err != nil {
		t.Errorf("Multiple close error: %v", err)
	}

	t.Logf("✓ Resource cleanup test completed, received content: %v", receivedContent)
}

// TestStreamingFunctionality_ErrorHandling tests error transformation in streaming
// Requirements: 3.2 - Test stream completion and error handling
func TestStreamingFunctionality_ErrorHandling(t *testing.T) {
	tests := []struct {
		name            string
		setupError      error
		expectedCode    string
		expectRetryable bool
	}{
		{
			name: "throttling error",
			setupError: &smithy.GenericAPIError{
				Code:    "ThrottlingException",
				Message: "Rate exceeded",
			},
			expectedCode:    services.ErrCodeRateLimit,
			expectRetryable: true,
		},
		{
			name: "access denied error",
			setupError: &smithy.GenericAPIError{
				Code:    "AccessDeniedException",
				Message: "Access denied",
			},
			expectedCode:    services.ErrCodeUnauthorized,
			expectRetryable: false,
		},
		{
			name: "validation error",
			setupError: &smithy.GenericAPIError{
				Code:    "ValidationException",
				Message: "Invalid parameters",
			},
			expectedCode:    services.ErrCodeInvalidInput,
			expectRetryable: false,
		},
		{
			name: "service unavailable",
			setupError: &smithy.GenericAPIError{
				Code:    "ServiceUnavailableException",
				Message: "Service unavailable",
			},
			expectedCode:    services.ErrCodeServiceError,
			expectRetryable: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock client that returns the error
			mockClient := &mockBedrockClient{
				invokeAgentFunc: func(ctx context.Context, input *bedrockagentruntime.InvokeAgentInput) (*bedrockagentruntime.InvokeAgentOutput, error) {
					return nil, tt.setupError
				},
			}

			adapter := &Adapter{
				client:  mockClient,
				agentID: "test-agent",
				aliasID: "test-alias",
				config: AdapterConfig{
					MaxRetries:     1,
					InitialBackoff: 1 * time.Millisecond,
					MaxBackoff:     10 * time.Millisecond,
					RequestTimeout: 5 * time.Second,
				},
			}

			input := services.AgentInput{
				SessionID: "test-session",
				Message:   "Test error handling",
			}

			_, err := adapter.InvokeAgentStream(context.Background(), input)

			// Should fail during setup
			if err == nil {
				t.Fatal("Expected setup error, got nil")
			}

			var domainErr *services.DomainError
			if !errors.As(err, &domainErr) {
				t.Fatalf("Expected DomainError, got: %T", err)
			}

			if domainErr.Code != tt.expectedCode {
				t.Errorf("Expected error code %q, got %q", tt.expectedCode, domainErr.Code)
			}

			if domainErr.Retryable != tt.expectRetryable {
				t.Errorf("Expected retryable %v, got %v", tt.expectRetryable, domainErr.Retryable)
			}

			t.Logf("✓ Error handling test: %s -> %s (retryable: %v)", 
				tt.name, domainErr.Code, domainErr.Retryable)
		})
	}
}

// TestStreamingFunctionality_ContextCancellation tests context cancellation during streaming
// Requirements: 3.2 - Test stream completion and error handling
func TestStreamingFunctionality_ContextCancellation(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping streaming functionality test in CI environment")
	}

	agentID := os.Getenv("BEDROCK_AGENT_ID")
	aliasID := os.Getenv("BEDROCK_AGENT_ALIAS_ID")

	if agentID == "" || aliasID == "" {
		t.Skip("Skipping streaming test - BEDROCK_AGENT_ID and BEDROCK_AGENT_ALIAS_ID must be set")
	}

	adapter, err := NewAdapter(context.Background(), agentID, aliasID, DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to create Bedrock adapter: %v", err)
	}

	input := services.AgentInput{
		SessionID: generateTestSessionID(),
		Message:   "Please tell me a long story about technology",
	}

	// Create context with short timeout to simulate cancellation
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	streamReader, err := adapter.InvokeAgentStream(ctx, input)
	if err != nil {
		// If the context times out during setup, that's also valid
		if errors.Is(err, context.DeadlineExceeded) {
			t.Logf("✓ Context cancellation test: timeout during setup")
			return
		}
		t.Fatalf("InvokeAgentStream failed: %v", err)
	}
	defer streamReader.Close()

	// Try to read - should eventually be cancelled due to timeout
	for {
		_, done, err := streamReader.Read()
		if done {
			t.Logf("✓ Context cancellation test: stream completed before timeout")
			return
		}
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
				t.Logf("✓ Context cancellation test: properly cancelled with %v", err)
				return
			}
			t.Fatalf("Unexpected error: %v", err)
		}
	}
}

// TestStreamingFunctionality_InputValidation tests input validation for streaming
// Requirements: 3.2 - Test stream completion and error handling
func TestStreamingFunctionality_InputValidation(t *testing.T) {
	// Test validation logic directly
	adapter := &Adapter{
		agentID: "test-agent",
		aliasID: "test-alias",
		config:  DefaultConfig(),
	}

	tests := []struct {
		name    string
		input   services.AgentInput
		wantErr bool
	}{
		{
			name: "empty session ID",
			input: services.AgentInput{
				SessionID: "",
				Message:   "Test message",
			},
			wantErr: true,
		},
		{
			name: "empty message",
			input: services.AgentInput{
				SessionID: "test-session",
				Message:   "",
			},
			wantErr: true,
		},
		{
			name: "message too long",
			input: services.AgentInput{
				SessionID: "test-session",
				Message:   string(make([]byte, 26000)),
			},
			wantErr: true,
		},
		{
			name: "valid input",
			input: services.AgentInput{
				SessionID: "test-session",
				Message:   "Valid message",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test validation directly without calling AWS
			err := adapter.validateInput(tt.input)
			
			if tt.wantErr {
				if err == nil {
					t.Error("Expected validation error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected validation error for valid input: %v", err)
				}
			}

			t.Logf("✓ Input validation test: %s", tt.name)
		})
	}
}

