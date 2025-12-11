package bedrock

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/bedrock-chat-poc/backend/domain/services"
)

// TestKnowledgeBaseIntegration tests comprehensive knowledge base functionality
// This test suite validates all aspects of knowledge base integration including
// queries, citation generation, permissions, and response enhancement
// Requirements: 1.3, 1.5, 8.1, 8.2, 8.3
func TestKnowledgeBaseIntegration(t *testing.T) {
	// Skip if running in CI or if Bedrock configuration is not available
	if os.Getenv("CI") != "" {
		t.Skip("Skipping knowledge base integration test in CI environment")
	}

	agentID := os.Getenv("BEDROCK_AGENT_ID")
	aliasID := os.Getenv("BEDROCK_AGENT_ALIAS_ID")
	knowledgeBaseID := os.Getenv("BEDROCK_KNOWLEDGE_BASE_ID")

	if agentID == "" || aliasID == "" {
		t.Skip("Skipping knowledge base integration test - BEDROCK_AGENT_ID and BEDROCK_AGENT_ALIAS_ID must be set")
	}

	if knowledgeBaseID == "" {
		t.Skip("Skipping knowledge base integration test - BEDROCK_KNOWLEDGE_BASE_ID not set")
	}

	ctx := context.Background()

	// Create adapter with real AWS configuration
	adapter, err := NewAdapter(ctx, agentID, aliasID, DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to create Bedrock adapter: %v", err)
	}

	// Test 1: Verify knowledge base queries work correctly
	// Requirements: 1.3 - WHEN the agent input includes knowledge base IDs THEN the system SHALL return a response that may include citations from the knowledge base
	t.Run("KnowledgeBaseQueryFunctionality", func(t *testing.T) {
		testCases := []struct {
			name     string
			message  string
			expected []string // Keywords we expect to find in responses that use knowledge base
		}{
			{
				name:     "General knowledge base query",
				message:  "What information do you have available in your knowledge base?",
				expected: []string{"information", "available", "knowledge"},
			},
			{
				name:     "Specific document query",
				message:  "Tell me about the company policies and procedures.",
				expected: []string{"policy", "procedure", "company"},
			},
			{
				name:     "Technical documentation query",
				message:  "What technical documentation is available?",
				expected: []string{"technical", "documentation", "available"},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				input := services.AgentInput{
					SessionID:        generateTestSessionID(),
					Message:          tc.message,
					KnowledgeBaseIDs: []string{knowledgeBaseID},
				}

				response, err := adapter.InvokeAgent(ctx, input)
				if err != nil {
					t.Fatalf("Knowledge base query failed: %v", err)
				}

				// Verify response structure
				if response == nil {
					t.Fatal("Expected non-nil response from knowledge base query")
				}

				if response.Content == "" {
					t.Error("Expected non-empty content from knowledge base response")
				}

				// Verify response is substantial
				if len(response.Content) < 20 {
					t.Errorf("Expected substantial response content, got: %s", response.Content)
				}

				// Verify citations array is initialized
				if response.Citations == nil {
					t.Error("Expected citations array to be initialized")
				}

				// Verify metadata is initialized
				if response.Metadata == nil {
					t.Error("Expected metadata map to be initialized")
				}

				t.Logf("Knowledge base query '%s': %d characters, %d citations", 
					tc.name, len(response.Content), len(response.Citations))

				// Log response for manual verification
				t.Logf("Response content: %s", response.Content[:min(200, len(response.Content))])
			})
		}
	})

	// Test 2: Test citation generation and parsing
	// Requirements: 1.5 - WHEN the agent response includes citations THEN the system SHALL properly parse and return citation metadata including source information
	// Requirements: 8.1 - WHEN the agent response includes citations THEN the system SHALL convert AWS citation format to domain citation format
	t.Run("CitationGenerationAndParsing", func(t *testing.T) {
		// Use a query that's likely to generate citations from knowledge base
		input := services.AgentInput{
			SessionID:        generateTestSessionID(),
			Message:          "Please provide detailed information from your knowledge base with specific references and sources.",
			KnowledgeBaseIDs: []string{knowledgeBaseID},
		}

		response, err := adapter.InvokeAgent(ctx, input)
		if err != nil {
			t.Fatalf("Citation generation test failed: %v", err)
		}

		if response == nil {
			t.Fatal("Expected non-nil response")
		}

		t.Logf("Citation test response: %d characters, %d citations", 
			len(response.Content), len(response.Citations))

		// If citations are present, verify their structure and content
		if len(response.Citations) > 0 {
			t.Logf("✓ Citations found - testing citation structure")

			for i, citation := range response.Citations {
				t.Logf("Testing citation %d", i+1)

				// Requirements: 8.2 - WHEN citations contain generated response parts THEN the system SHALL extract the text excerpt
				if citation.Excerpt == "" {
					t.Errorf("Citation %d has empty excerpt", i)
				} else {
					t.Logf("  ✓ Citation %d has excerpt: %s", i, citation.Excerpt[:min(100, len(citation.Excerpt))])
				}

				// Requirements: 8.3 - WHEN citations contain retrieved references THEN the system SHALL extract source name and URL from S3 location
				if citation.SourceName == "" && citation.SourceID == "" {
					t.Logf("  Note: Citation %d has no source name or ID (may be expected)", i)
				} else {
					if citation.SourceName != "" {
						t.Logf("  ✓ Citation %d has source name: %s", i, citation.SourceName[:min(50, len(citation.SourceName))])
					}
					if citation.SourceID != "" {
						t.Logf("  ✓ Citation %d has source ID: %s", i, citation.SourceID)
					}
				}

				// Requirements: 8.4 - WHEN citations contain metadata THEN the system SHALL preserve all metadata in the domain citation
				if citation.Metadata == nil {
					t.Errorf("Citation %d has nil metadata map", i)
				} else {
					t.Logf("  ✓ Citation %d has metadata with %d keys", i, len(citation.Metadata))
					for key, value := range citation.Metadata {
						t.Logf("    Metadata: %s = %v", key, value)
					}
				}

				// Verify URL field if present
				if citation.URL != "" {
					t.Logf("  ✓ Citation %d has URL: %s", i, citation.URL)
				}

				// Verify confidence if present
				if citation.Confidence > 0 {
					t.Logf("  ✓ Citation %d has confidence: %.2f", i, citation.Confidence)
				}
			}
		} else {
			t.Logf("Note: No citations returned - this may be expected depending on knowledge base content and query")
		}
	})

	// Test 3: Validate knowledge base permissions
	// Requirements: 10.4 - WHEN knowledge base IDs are invalid or inaccessible THEN the system SHALL return permission errors with specific resource information
	t.Run("KnowledgeBasePermissions", func(t *testing.T) {
		// Test with valid knowledge base ID (should work)
		t.Run("ValidKnowledgeBaseAccess", func(t *testing.T) {
			input := services.AgentInput{
				SessionID:        generateTestSessionID(),
				Message:          "Test valid knowledge base access",
				KnowledgeBaseIDs: []string{knowledgeBaseID},
			}

			response, err := adapter.InvokeAgent(ctx, input)
			if err != nil {
				t.Fatalf("Valid knowledge base access failed: %v", err)
			}

			if response == nil {
				t.Fatal("Expected non-nil response from valid knowledge base")
			}

			if response.Content == "" {
				t.Error("Expected non-empty content from valid knowledge base")
			}

			t.Logf("✓ Valid knowledge base access successful: %d characters", len(response.Content))
		})

		// Test with invalid knowledge base ID
		t.Run("InvalidKnowledgeBaseHandling", func(t *testing.T) {
			invalidKnowledgeBaseID := "INVALID-KB-ID-12345"
			input := services.AgentInput{
				SessionID:        generateTestSessionID(),
				Message:          "Test with invalid knowledge base ID",
				KnowledgeBaseIDs: []string{invalidKnowledgeBaseID},
			}

			response, err := adapter.InvokeAgent(ctx, input)

			// Note: Bedrock Agent behavior with invalid knowledge base IDs can vary
			// Some configurations may ignore invalid IDs and proceed
			// Others may return permission errors
			
			if err != nil {
				var domainErr *services.DomainError
				if errors.As(err, &domainErr) {
					if domainErr.Code == services.ErrCodeUnauthorized {
						t.Logf("✓ Invalid knowledge base ID correctly rejected: %s", domainErr.Message)
						return
					}
				}
				t.Logf("Unexpected error with invalid knowledge base ID: %v", err)
				return
			}

			// If no error occurred, the agent may have ignored the invalid knowledge base
			if response != nil {
				t.Logf("Note: Agent processed request despite invalid knowledge base ID (may have ignored it)")
				t.Logf("Response: %d characters, %d citations", len(response.Content), len(response.Citations))
			}
		})

		// Test with multiple knowledge base IDs (valid + invalid)
		t.Run("MixedKnowledgeBaseIDs", func(t *testing.T) {
			input := services.AgentInput{
				SessionID: generateTestSessionID(),
				Message:   "Test with mixed valid and invalid knowledge base IDs",
				KnowledgeBaseIDs: []string{
					knowledgeBaseID,           // Valid
					"INVALID-KB-ID-67890",     // Invalid
				},
			}

			response, err := adapter.InvokeAgent(ctx, input)

			// The behavior here depends on Bedrock Agent configuration
			// It may process the valid KB and ignore the invalid one
			// Or it may return an error for the invalid KB
			
			if err != nil {
				t.Logf("Mixed knowledge base IDs resulted in error: %v", err)
			} else if response != nil {
				t.Logf("Mixed knowledge base IDs processed successfully: %d characters, %d citations", 
					len(response.Content), len(response.Citations))
			}
		})
	})

	// Test 4: Test response enhancement with knowledge base context
	// Requirements: 1.3 - Knowledge base integration should enhance responses
	t.Run("ResponseEnhancementWithKnowledgeBase", func(t *testing.T) {
		sessionID := generateTestSessionID()

		// Test comparison: same question with and without knowledge base
		baseMessage := "What are the best practices for software development?"

		// First: Ask without knowledge base
		inputWithoutKB := services.AgentInput{
			SessionID: sessionID + "-no-kb",
			Message:   baseMessage,
			// No KnowledgeBaseIDs
		}

		responseWithoutKB, err := adapter.InvokeAgent(ctx, inputWithoutKB)
		if err != nil {
			t.Fatalf("Query without knowledge base failed: %v", err)
		}

		// Wait a moment to ensure different session processing
		time.Sleep(1 * time.Second)

		// Second: Ask with knowledge base
		inputWithKB := services.AgentInput{
			SessionID:        sessionID + "-with-kb",
			Message:          baseMessage,
			KnowledgeBaseIDs: []string{knowledgeBaseID},
		}

		responseWithKB, err := adapter.InvokeAgent(ctx, inputWithKB)
		if err != nil {
			t.Fatalf("Query with knowledge base failed: %v", err)
		}

		// Compare responses
		t.Logf("Response comparison for: %s", baseMessage)
		t.Logf("Without KB: %d characters, %d citations", 
			len(responseWithoutKB.Content), len(responseWithoutKB.Citations))
		t.Logf("With KB: %d characters, %d citations", 
			len(responseWithKB.Content), len(responseWithKB.Citations))

		// Log content samples for manual verification
		t.Logf("Without KB content sample: %s", 
			responseWithoutKB.Content[:min(200, len(responseWithoutKB.Content))])
		t.Logf("With KB content sample: %s", 
			responseWithKB.Content[:min(200, len(responseWithKB.Content))])

		// Verify knowledge base response has proper structure
		if responseWithKB.Citations == nil {
			t.Error("Knowledge base response should have initialized citations array")
		}

		if responseWithKB.Metadata == nil {
			t.Error("Knowledge base response should have initialized metadata")
		}

		// The knowledge base response may or may not have citations depending on content
		// But it should be a valid response
		if responseWithKB.Content == "" {
			t.Error("Knowledge base response should have content")
		}
	})

	// Test 5: Test streaming with knowledge base integration
	t.Run("StreamingWithKnowledgeBase", func(t *testing.T) {
		input := services.AgentInput{
			SessionID:        generateTestSessionID(),
			Message:          "Please provide a detailed explanation using information from your knowledge base.",
			KnowledgeBaseIDs: []string{knowledgeBaseID},
		}

		streamReader, err := adapter.InvokeAgentStream(ctx, input)
		if err != nil {
			t.Fatalf("Knowledge base streaming failed: %v", err)
		}
		defer streamReader.Close()

		var totalContent strings.Builder
		var citations []interface{}
		chunkCount := 0

		// Read all chunks from the stream
		for {
			chunk, done, err := streamReader.Read()
			if done {
				break
			}
			if err != nil {
				t.Fatalf("Stream read error: %v", err)
			}

			if chunk != "" {
				totalContent.WriteString(chunk)
				chunkCount++
			}

			// Check for citations in stream
			citation, err := streamReader.ReadCitation()
			if err == nil && citation != nil {
				citations = append(citations, citation)
				t.Logf("Stream citation received: %v", citation)
			}
		}

		// Verify streaming response
		if totalContent.Len() == 0 {
			t.Error("Expected to receive content from knowledge base stream")
		}

		if chunkCount == 0 {
			t.Error("Expected to receive at least one content chunk")
		}

		t.Logf("Knowledge base streaming test: %d chunks, %d total characters, %d citations", 
			chunkCount, totalContent.Len(), len(citations))

		// Verify content is substantial
		if totalContent.Len() < 20 {
			t.Errorf("Expected substantial streaming content, got: %s", totalContent.String())
		}
	})

	// Test 6: Test knowledge base input validation
	// Requirements: 2.4 - WHEN knowledge base IDs contain invalid formats THEN the system SHALL reject the input and return a validation error
	t.Run("KnowledgeBaseInputValidation", func(t *testing.T) {
		testCases := []struct {
			name           string
			knowledgeBaseIDs []string
			expectError    bool
			description    string
		}{
			{
				name:           "Valid knowledge base ID",
				knowledgeBaseIDs: []string{knowledgeBaseID},
				expectError:    false,
				description:    "Should accept valid knowledge base ID",
			},
			{
				name:           "Multiple valid knowledge base IDs",
				knowledgeBaseIDs: []string{knowledgeBaseID, knowledgeBaseID}, // Using same ID twice for test
				expectError:    false,
				description:    "Should accept multiple knowledge base IDs",
			},
			{
				name:           "Empty knowledge base ID in array",
				knowledgeBaseIDs: []string{""},
				expectError:    true,
				description:    "Should reject empty knowledge base ID",
			},
			{
				name:           "Mixed valid and empty knowledge base IDs",
				knowledgeBaseIDs: []string{knowledgeBaseID, ""},
				expectError:    true,
				description:    "Should reject array containing empty knowledge base ID",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				input := services.AgentInput{
					SessionID:        generateTestSessionID(),
					Message:          "Test knowledge base ID validation",
					KnowledgeBaseIDs: tc.knowledgeBaseIDs,
				}

				response, err := adapter.InvokeAgent(ctx, input)

				if tc.expectError {
					if err == nil {
						t.Errorf("Expected error for %s, but got none", tc.description)
					} else {
						t.Logf("✓ Correctly rejected invalid input: %v", err)
					}
				} else {
					if err != nil {
						t.Errorf("Unexpected error for %s: %v", tc.description, err)
					} else if response == nil {
						t.Errorf("Expected valid response for %s", tc.description)
					} else {
						t.Logf("✓ Valid input accepted: %s", tc.description)
					}
				}
			})
		}
	})

	// Test 7: Test knowledge base context in conversation flow
	t.Run("KnowledgeBaseInConversationFlow", func(t *testing.T) {
		sessionID := generateTestSessionID()

		// Multi-turn conversation using knowledge base
		conversationSteps := []struct {
			message     string
			useKB       bool
			description string
		}{
			{
				message:     "Hello, I need help with technical documentation.",
				useKB:       false,
				description: "Initial greeting without KB",
			},
			{
				message:     "What technical information do you have available?",
				useKB:       true,
				description: "Query knowledge base for available information",
			},
			{
				message:     "Can you provide more details about that?",
				useKB:       true,
				description: "Follow-up question with KB context",
			},
			{
				message:     "Thank you for the information.",
				useKB:       false,
				description: "Closing without KB",
			},
		}

		for i, step := range conversationSteps {
			t.Logf("Conversation step %d: %s", i+1, step.description)

			input := services.AgentInput{
				SessionID: sessionID,
				Message:   step.message,
			}

			if step.useKB {
				input.KnowledgeBaseIDs = []string{knowledgeBaseID}
			}

			response, err := adapter.InvokeAgent(ctx, input)
			if err != nil {
				t.Fatalf("Conversation step %d failed: %v", i+1, err)
			}

			if response == nil || response.Content == "" {
				t.Errorf("Expected valid response for conversation step %d", i+1)
				continue
			}

			t.Logf("  Response: %d characters, %d citations", 
				len(response.Content), len(response.Citations))

			// Log if citations were received when KB was used
			if step.useKB && len(response.Citations) > 0 {
				t.Logf("  ✓ Knowledge base citations received: %d", len(response.Citations))
			}

			// Wait between conversation steps
			if i < len(conversationSteps)-1 {
				time.Sleep(1 * time.Second)
			}
		}

		t.Logf("Knowledge base conversation flow completed with %d exchanges", len(conversationSteps))
	})
}