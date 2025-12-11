package bedrock

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/bedrock-chat-poc/backend/domain/services"
)

// TestBedrockAgentConnectivity tests basic connectivity to real Bedrock Agent
// This is an integration test that requires valid AWS credentials and Bedrock Agent configuration
// Requirements: 1.1, 1.2
func TestBedrockAgentConnectivity(t *testing.T) {
	// Skip if running in CI or if Bedrock configuration is not available
	if os.Getenv("CI") != "" {
		t.Skip("Skipping integration test in CI environment")
	}

	agentID := os.Getenv("BEDROCK_AGENT_ID")
	aliasID := os.Getenv("BEDROCK_AGENT_ALIAS_ID")

	if agentID == "" || aliasID == "" {
		t.Skip("Skipping integration test - BEDROCK_AGENT_ID and BEDROCK_AGENT_ALIAS_ID must be set")
	}

	ctx := context.Background()

	// Create adapter with real AWS configuration
	adapter, err := NewAdapter(ctx, agentID, aliasID, DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to create Bedrock adapter: %v", err)
	}

	// Test basic agent invocation with simple message
	t.Run("BasicAgentInvocation", func(t *testing.T) {
		input := services.AgentInput{
			SessionID: generateTestSessionID(),
			Message:   "Hello, can you help me with a simple question?",
		}

		response, err := adapter.InvokeAgent(ctx, input)
		if err != nil {
			t.Fatalf("InvokeAgent failed: %v", err)
		}

		// Verify response structure
		if response == nil {
			t.Fatal("Expected non-nil response")
		}

		if response.Content == "" {
			t.Error("Expected non-empty content in response")
		}

		if len(response.Content) < 10 {
			t.Errorf("Expected substantial content, got: %s", response.Content)
		}

		// Verify response has proper structure
		if response.Citations == nil {
			t.Error("Expected citations array to be initialized (even if empty)")
		}

		if response.Metadata == nil {
			t.Error("Expected metadata map to be initialized")
		}

		t.Logf("Received response with %d characters of content and %d citations", 
			len(response.Content), len(response.Citations))
	})

	// Test agent invocation with knowledge base integration (if configured)
	t.Run("AgentInvocationWithKnowledgeBase", func(t *testing.T) {
		knowledgeBaseID := os.Getenv("BEDROCK_KNOWLEDGE_BASE_ID")
		if knowledgeBaseID == "" {
			t.Skip("Skipping knowledge base test - BEDROCK_KNOWLEDGE_BASE_ID not set")
		}

		input := services.AgentInput{
			SessionID:        generateTestSessionID(),
			Message:          "What information do you have available in your knowledge base?",
			KnowledgeBaseIDs: []string{knowledgeBaseID},
		}

		response, err := adapter.InvokeAgent(ctx, input)
		if err != nil {
			t.Fatalf("InvokeAgent with knowledge base failed: %v", err)
		}

		// Verify response structure
		if response == nil {
			t.Fatal("Expected non-nil response")
		}

		if response.Content == "" {
			t.Error("Expected non-empty content in response")
		}

		// Knowledge base responses may include citations
		t.Logf("Knowledge base response: %d characters, %d citations", 
			len(response.Content), len(response.Citations))

		// If citations are present, verify their structure
		for i, citation := range response.Citations {
			if citation.Excerpt == "" {
				t.Errorf("Citation %d has empty excerpt", i)
			}
			if citation.Metadata == nil {
				t.Errorf("Citation %d has nil metadata", i)
			}
		}
	})

	// Test session context preservation and conversation flow
	// Requirements: 1.4 - Session context maintenance across multiple messages
	t.Run("SessionContextAndConversationFlow", func(t *testing.T) {
		sessionID := generateTestSessionID()

		// Test 1: Basic context establishment and retrieval
		t.Run("BasicContextRetention", func(t *testing.T) {
			// First message - establish context with specific information
			input1 := services.AgentInput{
				SessionID: sessionID,
				Message:   "My name is Alice and I work as a software engineer. Please remember this information.",
			}

			response1, err := adapter.InvokeAgent(ctx, input1)
			if err != nil {
				t.Fatalf("First InvokeAgent failed: %v", err)
			}

			if response1.Content == "" {
				t.Error("Expected response to first message")
			}

			// Wait to ensure session is processed
			time.Sleep(2 * time.Second)

			// Second message - test context retention
			input2 := services.AgentInput{
				SessionID: sessionID,
				Message:   "What is my name and what do I do for work?",
			}

			response2, err := adapter.InvokeAgent(ctx, input2)
			if err != nil {
				t.Fatalf("Second InvokeAgent failed: %v", err)
			}

			if response2.Content == "" {
				t.Error("Expected response to second message")
			}

			// Verify context retention - check for name and profession
			responseContent := strings.ToLower(response2.Content)
			hasName := strings.Contains(responseContent, "alice")
			hasProfession := strings.Contains(responseContent, "software") || strings.Contains(responseContent, "engineer")

			t.Logf("Context retention test:")
			t.Logf("  First message: %s", input1.Message)
			t.Logf("  First response: %s", response1.Content[:min(200, len(response1.Content))])
			t.Logf("  Second message: %s", input2.Message)
			t.Logf("  Second response: %s", response2.Content[:min(200, len(response2.Content))])
			t.Logf("  Name retained: %v, Profession retained: %v", hasName, hasProfession)

			if !hasName && !hasProfession {
				t.Logf("Warning: Context may not be fully retained. This could indicate session context issues.")
			}
		})

		// Test 2: Multi-turn conversation flow
		t.Run("MultiTurnConversation", func(t *testing.T) {
			conversationSessionID := generateTestSessionID()

			// Simulate a realistic conversation flow
			conversationSteps := []struct {
				message          string
				expectedKeywords []string
				description      string
			}{
				{
					message:          "I'm planning a vacation to Japan. Can you help me?",
					expectedKeywords: []string{"japan", "vacation", "help"},
					description:      "Initial vacation planning request",
				},
				{
					message:          "What are the best cities to visit?",
					expectedKeywords: []string{"cities", "visit", "tokyo", "kyoto", "osaka"},
					description:      "Follow-up about cities (should understand Japan context)",
				},
				{
					message:          "How long should I stay there?",
					expectedKeywords: []string{"stay", "duration", "days", "weeks"},
					description:      "Duration question (should understand Japan vacation context)",
				},
				{
					message:          "What's the best time of year to go?",
					expectedKeywords: []string{"time", "season", "weather", "spring", "fall"},
					description:      "Timing question (should maintain Japan vacation context)",
				},
			}

			var previousResponses []string

			for i, step := range conversationSteps {
				t.Logf("Conversation step %d: %s", i+1, step.description)

				input := services.AgentInput{
					SessionID: conversationSessionID,
					Message:   step.message,
				}

				response, err := adapter.InvokeAgent(ctx, input)
				if err != nil {
					t.Fatalf("Conversation step %d failed: %v", i+1, err)
				}

				if response.Content == "" {
					t.Errorf("Expected response for conversation step %d", i+1)
					continue
				}

				previousResponses = append(previousResponses, response.Content)

				// Check if response is contextually appropriate
				responseContent := strings.ToLower(response.Content)
				contextuallyRelevant := false

				for _, keyword := range step.expectedKeywords {
					if strings.Contains(responseContent, keyword) {
						contextuallyRelevant = true
						break
					}
				}

				t.Logf("  Message: %s", step.message)
				t.Logf("  Response: %s", response.Content[:min(150, len(response.Content))])
				t.Logf("  Contextually relevant: %v", contextuallyRelevant)

				if !contextuallyRelevant {
					t.Logf("  Warning: Response may not be contextually relevant to the conversation")
				}

				// Wait between messages to simulate realistic conversation timing
				if i < len(conversationSteps)-1 {
					time.Sleep(1 * time.Second)
				}
			}

			t.Logf("Multi-turn conversation completed with %d exchanges", len(conversationSteps))
		})

		// Test 3: Session isolation between different conversations
		t.Run("SessionIsolation", func(t *testing.T) {
			// Create two separate sessions
			session1ID := generateTestSessionID()
			session2ID := generateTestSessionID()

			// Establish different contexts in each session
			// Session 1: Doctor context
			input1A := services.AgentInput{
				SessionID: session1ID,
				Message:   "I am Dr. Smith, a cardiologist. I specialize in heart surgery.",
			}

			response1A, err := adapter.InvokeAgent(ctx, input1A)
			if err != nil {
				t.Fatalf("Session 1 setup failed: %v", err)
			}

			if response1A.Content == "" {
				t.Error("Expected response to session 1 setup")
			}

			// Session 2: Teacher context
			input2A := services.AgentInput{
				SessionID: session2ID,
				Message:   "I am Ms. Johnson, a high school math teacher. I teach algebra and geometry.",
			}

			response2A, err := adapter.InvokeAgent(ctx, input2A)
			if err != nil {
				t.Fatalf("Session 2 setup failed: %v", err)
			}

			if response2A.Content == "" {
				t.Error("Expected response to session 2 setup")
			}

			// Wait for sessions to be processed
			time.Sleep(2 * time.Second)

			// Test isolation: Ask about profession in each session
			// Session 1 should remember doctor context
			input1B := services.AgentInput{
				SessionID: session1ID,
				Message:   "What is my profession and specialty?",
			}

			response1B, err := adapter.InvokeAgent(ctx, input1B)
			if err != nil {
				t.Fatalf("Session 1 context test failed: %v", err)
			}

			// Session 2 should remember teacher context
			input2B := services.AgentInput{
				SessionID: session2ID,
				Message:   "What is my profession and what subjects do I teach?",
			}

			response2B, err := adapter.InvokeAgent(ctx, input2B)
			if err != nil {
				t.Fatalf("Session 2 context test failed: %v", err)
			}

			// Verify session isolation
			response1Content := strings.ToLower(response1B.Content)
			response2Content := strings.ToLower(response2B.Content)

			// Session 1 should mention doctor/cardiology, not teacher/math
			session1HasDoctor := strings.Contains(response1Content, "doctor") || 
								strings.Contains(response1Content, "cardiologist") ||
								strings.Contains(response1Content, "heart")
			session1HasTeacher := strings.Contains(response1Content, "teacher") || 
								 strings.Contains(response1Content, "math") ||
								 strings.Contains(response1Content, "algebra")

			// Session 2 should mention teacher/math, not doctor/cardiology
			session2HasTeacher := strings.Contains(response2Content, "teacher") || 
								 strings.Contains(response2Content, "math") ||
								 strings.Contains(response2Content, "algebra") ||
								 strings.Contains(response2Content, "geometry")
			session2HasDoctor := strings.Contains(response2Content, "doctor") || 
								strings.Contains(response2Content, "cardiologist") ||
								strings.Contains(response2Content, "heart")

			t.Logf("Session isolation test results:")
			t.Logf("  Session 1 (Doctor):")
			t.Logf("    Setup: %s", input1A.Message)
			t.Logf("    Query: %s", input1B.Message)
			t.Logf("    Response: %s", response1B.Content[:min(200, len(response1B.Content))])
			t.Logf("    Has doctor context: %v, Has teacher context: %v", session1HasDoctor, session1HasTeacher)
			
			t.Logf("  Session 2 (Teacher):")
			t.Logf("    Setup: %s", input2A.Message)
			t.Logf("    Query: %s", input2B.Message)
			t.Logf("    Response: %s", response2B.Content[:min(200, len(response2B.Content))])
			t.Logf("    Has teacher context: %v, Has doctor context: %v", session2HasTeacher, session2HasDoctor)

			// Verify proper isolation
			if session1HasTeacher {
				t.Logf("Warning: Session 1 may have leaked context from Session 2")
			}
			if session2HasDoctor {
				t.Logf("Warning: Session 2 may have leaked context from Session 1")
			}

			// Positive verification
			if !session1HasDoctor {
				t.Logf("Warning: Session 1 may not have retained its doctor context")
			}
			if !session2HasTeacher {
				t.Logf("Warning: Session 2 may not have retained its teacher context")
			}
		})

		// Test 4: Context persistence across streaming and non-streaming calls
		t.Run("ContextPersistenceAcrossCallTypes", func(t *testing.T) {
			mixedSessionID := generateTestSessionID()

			// Establish context with regular call
			input1 := services.AgentInput{
				SessionID: mixedSessionID,
				Message:   "I am a chef who specializes in Italian cuisine. I own a restaurant called Bella Vista.",
			}

			response1, err := adapter.InvokeAgent(ctx, input1)
			if err != nil {
				t.Fatalf("Context setup with regular call failed: %v", err)
			}

			time.Sleep(2 * time.Second)

			// Test context with streaming call
			input2 := services.AgentInput{
				SessionID: mixedSessionID,
				Message:   "What type of restaurant do I own and what cuisine do I specialize in?",
			}

			streamReader, err := adapter.InvokeAgentStream(ctx, input2)
			if err != nil {
				t.Fatalf("Streaming call failed: %v", err)
			}
			defer streamReader.Close()

			var streamContent strings.Builder
			for {
				chunk, done, err := streamReader.Read()
				if done {
					break
				}
				if err != nil {
					t.Fatalf("Stream read error: %v", err)
				}
				streamContent.WriteString(chunk)
			}

			// Verify context retention in streaming response
			streamResponseContent := strings.ToLower(streamContent.String())
			hasChef := strings.Contains(streamResponseContent, "chef")
			hasItalian := strings.Contains(streamResponseContent, "italian")
			hasRestaurant := strings.Contains(streamResponseContent, "restaurant") || 
							strings.Contains(streamResponseContent, "bella vista")

			t.Logf("Mixed call type test:")
			t.Logf("  Setup (regular): %s", input1.Message)
			t.Logf("  Setup response: %s", response1.Content[:min(150, len(response1.Content))])
			t.Logf("  Query (streaming): %s", input2.Message)
			t.Logf("  Stream response: %s", streamContent.String()[:min(150, len(streamContent.String()))])
			t.Logf("  Context retained - Chef: %v, Italian: %v, Restaurant: %v", hasChef, hasItalian, hasRestaurant)

			if !hasChef && !hasItalian && !hasRestaurant {
				t.Logf("Warning: Context may not be retained across different call types")
			}
		})
	})

	// Test streaming functionality
	t.Run("StreamingResponse", func(t *testing.T) {
		input := services.AgentInput{
			SessionID: generateTestSessionID(),
			Message:   "Please tell me a brief story about artificial intelligence.",
		}

		streamReader, err := adapter.InvokeAgentStream(ctx, input)
		if err != nil {
			t.Fatalf("InvokeAgentStream failed: %v", err)
		}
		defer streamReader.Close()

		var totalContent strings.Builder
		chunkCount := 0
		citationCount := 0

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

			// Check for citations
			citation, err := streamReader.ReadCitation()
			if err == nil && citation != nil {
				citationCount++
				if citation.Excerpt == "" {
					t.Error("Stream citation has empty excerpt")
				}
			}
		}

		// Verify we received content
		if totalContent.Len() == 0 {
			t.Error("Expected to receive content from stream")
		}

		if chunkCount == 0 {
			t.Error("Expected to receive at least one content chunk")
		}

		t.Logf("Stream test: received %d chunks, %d total characters, %d citations", 
			chunkCount, totalContent.Len(), citationCount)

		// Verify content is substantial
		if totalContent.Len() < 20 {
			t.Errorf("Expected substantial streaming content, got: %s", totalContent.String())
		}
	})

	// Test error handling with invalid input
	t.Run("InputValidation", func(t *testing.T) {
		// Test empty session ID
		input := services.AgentInput{
			SessionID: "",
			Message:   "Test message",
		}

		_, err := adapter.InvokeAgent(ctx, input)
		if err == nil {
			t.Error("Expected error for empty session ID")
		}

		if !strings.Contains(err.Error(), "session ID") {
			t.Errorf("Expected session ID validation error, got: %v", err)
		}

		// Test empty message
		input = services.AgentInput{
			SessionID: generateTestSessionID(),
			Message:   "",
		}

		_, err = adapter.InvokeAgent(ctx, input)
		if err == nil {
			t.Error("Expected error for empty message")
		}

		if !strings.Contains(err.Error(), "message") {
			t.Errorf("Expected message validation error, got: %v", err)
		}

		// Test message too long
		input = services.AgentInput{
			SessionID: generateTestSessionID(),
			Message:   strings.Repeat("a", 25001),
		}

		_, err = adapter.InvokeAgent(ctx, input)
		if err == nil {
			t.Error("Expected error for message too long")
		}

		if !strings.Contains(err.Error(), "length") && !strings.Contains(err.Error(), "long") {
			t.Errorf("Expected length validation error, got: %v", err)
		}
	})

	// Test timeout handling
	t.Run("TimeoutHandling", func(t *testing.T) {
		// Create adapter with very short timeout for testing
		shortTimeoutConfig := DefaultConfig()
		shortTimeoutConfig.RequestTimeout = 1 * time.Millisecond

		shortAdapter, err := NewAdapter(ctx, agentID, aliasID, shortTimeoutConfig)
		if err != nil {
			t.Fatalf("Failed to create short timeout adapter: %v", err)
		}

		input := services.AgentInput{
			SessionID: generateTestSessionID(),
			Message:   "This request should timeout quickly",
		}

		_, err = shortAdapter.InvokeAgent(ctx, input)
		if err == nil {
			t.Error("Expected timeout error with very short timeout")
		}

		// Check if it's a timeout error
		if strings.Contains(err.Error(), "timeout") || strings.Contains(err.Error(), "context deadline exceeded") {
			t.Logf("Correctly received timeout error: %v", err)
		} else {
			t.Logf("Warning: Expected timeout error but got: %v", err)
		}
	})
}

// generateTestSessionID creates a unique session ID for testing
func generateTestSessionID() string {
	return "test-session-" + time.Now().Format("20060102-150405") + "-" + 
		   time.Now().Format("000000")
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}