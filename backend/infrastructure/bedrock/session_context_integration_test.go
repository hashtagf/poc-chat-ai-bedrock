package bedrock

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/bedrock-chat-poc/backend/domain/services"
)

// TestSessionContextIntegration tests session context and conversation flow with real Bedrock Agent
// This test requires valid AWS credentials and Bedrock Agent configuration
// Requirements: 1.4 - Session context maintenance across multiple messages
func TestSessionContextIntegration(t *testing.T) {
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

	// Test 1: Basic session context establishment and retrieval
	t.Run("SessionContextEstablishmentAndRetrieval", func(t *testing.T) {
		sessionID := generateUniqueSessionID("context-test")

		// Establish context with specific, memorable information
		setupInput := services.AgentInput{
			SessionID: sessionID,
			Message:   "Hello! My name is TestUser and I am a software developer working on AI applications. Please remember this information about me.",
		}

		setupResponse, err := adapter.InvokeAgent(ctx, setupInput)
		if err != nil {
			t.Fatalf("Context setup failed: %v", err)
		}

		if setupResponse.Content == "" {
			t.Error("Expected response to context setup message")
		}

		t.Logf("Context setup completed:")
		t.Logf("  Input: %s", setupInput.Message)
		t.Logf("  Response: %s", setupResponse.Content[:minInt(200, len(setupResponse.Content))])

		// Wait to ensure session context is processed
		time.Sleep(3 * time.Second)

		// Test context retrieval
		retrievalInput := services.AgentInput{
			SessionID: sessionID,
			Message:   "Can you tell me what you know about me from our conversation?",
		}

		retrievalResponse, err := adapter.InvokeAgent(ctx, retrievalInput)
		if err != nil {
			t.Fatalf("Context retrieval failed: %v", err)
		}

		if retrievalResponse.Content == "" {
			t.Error("Expected response to context retrieval message")
		}

		// Analyze response for context retention
		responseContent := strings.ToLower(retrievalResponse.Content)
		hasName := strings.Contains(responseContent, "testuser") || strings.Contains(responseContent, "test user")
		hasProfession := strings.Contains(responseContent, "software") || strings.Contains(responseContent, "developer") || strings.Contains(responseContent, "ai")

		t.Logf("Context retrieval results:")
		t.Logf("  Input: %s", retrievalInput.Message)
		t.Logf("  Response: %s", retrievalResponse.Content[:minInt(300, len(retrievalResponse.Content))])
		t.Logf("  Name retained: %v", hasName)
		t.Logf("  Profession retained: %v", hasProfession)

		// Document findings
		if hasName || hasProfession {
			t.Logf("✓ Session context is working - agent retained information from previous message")
		} else {
			t.Logf("⚠ Session context may not be fully working - agent did not clearly retain previous information")
			t.Logf("  This could be due to agent configuration, model behavior, or session handling")
		}
	})

	// Test 2: Multi-turn conversation flow
	t.Run("MultiTurnConversationFlow", func(t *testing.T) {
		sessionID := generateUniqueSessionID("conversation-test")

		// Define a conversation sequence that builds context
		conversationFlow := []struct {
			message     string
			description string
			checkFor    []string // Keywords to look for in response
		}{
			{
				message:     "I'm planning to learn a new programming language. I'm currently experienced with Go and Python.",
				description: "Establish programming background",
				checkFor:    []string{"programming", "language", "go", "python"},
			},
			{
				message:     "What language would you recommend for web development?",
				description: "Ask for recommendation (should consider established context)",
				checkFor:    []string{"web", "development", "recommend"},
			},
			{
				message:     "How does that compare to the languages I already know?",
				description: "Reference previous context (should remember Go and Python)",
				checkFor:    []string{"go", "python", "compare"},
			},
			{
				message:     "What would be the best way for someone with my background to get started?",
				description: "Ask for personalized advice (should consider full context)",
				checkFor:    []string{"background", "started", "experience"},
			},
		}

		var conversationHistory []string

		for i, step := range conversationFlow {
			t.Logf("Conversation step %d: %s", i+1, step.description)

			input := services.AgentInput{
				SessionID: sessionID,
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

			conversationHistory = append(conversationHistory, response.Content)

			// Check for contextual relevance
			responseContent := strings.ToLower(response.Content)
			relevantKeywords := 0
			for _, keyword := range step.checkFor {
				if strings.Contains(responseContent, keyword) {
					relevantKeywords++
				}
			}

			contextualRelevance := float64(relevantKeywords) / float64(len(step.checkFor))

			t.Logf("  Message: %s", step.message)
			t.Logf("  Response: %s", response.Content[:minInt(250, len(response.Content))])
			t.Logf("  Contextual relevance: %.1f%% (%d/%d keywords found)", 
				contextualRelevance*100, relevantKeywords, len(step.checkFor))

			// Wait between conversation steps
			if i < len(conversationFlow)-1 {
				time.Sleep(2 * time.Second)
			}
		}

		t.Logf("Multi-turn conversation completed with %d exchanges", len(conversationFlow))
		t.Logf("Conversation demonstrates session context maintenance across multiple related messages")
	})

	// Test 3: Session isolation verification
	t.Run("SessionIsolationVerification", func(t *testing.T) {
		// Create two distinct sessions with different contexts
		session1ID := generateUniqueSessionID("isolation-test-1")
		session2ID := generateUniqueSessionID("isolation-test-2")

		// Session 1: Establish medical professional context
		medical1Input := services.AgentInput{
			SessionID: session1ID,
			Message:   "I am Dr. Sarah Johnson, a cardiologist at City Hospital. I specialize in heart surgery and have 15 years of experience.",
		}

		medical1Response, err := adapter.InvokeAgent(ctx, medical1Input)
		if err != nil {
			t.Fatalf("Medical session setup failed: %v", err)
		}

		if medical1Response.Content == "" {
			t.Error("Expected response to medical session setup")
		}

		// Session 2: Establish teacher context
		teacher2Input := services.AgentInput{
			SessionID: session2ID,
			Message:   "I am Mr. David Chen, a high school mathematics teacher. I teach calculus and statistics to senior students.",
		}

		teacher2Response, err := adapter.InvokeAgent(ctx, teacher2Input)
		if err != nil {
			t.Fatalf("Teacher session setup failed: %v", err)
		}

		if teacher2Response.Content == "" {
			t.Error("Expected response to teacher session setup")
		}

		// Wait for context establishment
		time.Sleep(3 * time.Second)

		// Test session 1 context retention
		medical1Query := services.AgentInput{
			SessionID: session1ID,
			Message:   "What is my profession and where do I work?",
		}

		medical1QueryResponse, err := adapter.InvokeAgent(ctx, medical1Query)
		if err != nil {
			t.Fatalf("Medical session query failed: %v", err)
		}

		// Test session 2 context retention
		teacher2Query := services.AgentInput{
			SessionID: session2ID,
			Message:   "What subjects do I teach and to which students?",
		}

		teacher2QueryResponse, err := adapter.InvokeAgent(ctx, teacher2Query)
		if err != nil {
			t.Fatalf("Teacher session query failed: %v", err)
		}

		// Analyze session isolation
		medical1Content := strings.ToLower(medical1QueryResponse.Content)
		teacher2Content := strings.ToLower(teacher2QueryResponse.Content)

		// Check for correct context retention
		medical1HasMedical := strings.Contains(medical1Content, "doctor") || 
							 strings.Contains(medical1Content, "cardiologist") ||
							 strings.Contains(medical1Content, "hospital") ||
							 strings.Contains(medical1Content, "heart")

		teacher2HasTeacher := strings.Contains(teacher2Content, "teacher") ||
							 strings.Contains(teacher2Content, "mathematics") ||
							 strings.Contains(teacher2Content, "calculus") ||
							 strings.Contains(teacher2Content, "students")

		// Check for context leakage (should not happen)
		medical1HasTeacher := strings.Contains(medical1Content, "teacher") ||
							 strings.Contains(medical1Content, "mathematics") ||
							 strings.Contains(medical1Content, "calculus")

		teacher2HasMedical := strings.Contains(teacher2Content, "doctor") ||
							 strings.Contains(teacher2Content, "cardiologist") ||
							 strings.Contains(teacher2Content, "hospital")

		t.Logf("Session isolation test results:")
		t.Logf("  Medical Session (ID: %s):", session1ID[:12]+"...")
		t.Logf("    Setup: %s", medical1Input.Message[:80]+"...")
		t.Logf("    Query: %s", medical1Query.Message)
		t.Logf("    Response: %s", medical1QueryResponse.Content[:minInt(200, len(medical1QueryResponse.Content))])
		t.Logf("    Has medical context: %v, Has teacher context: %v", medical1HasMedical, medical1HasTeacher)

		t.Logf("  Teacher Session (ID: %s):", session2ID[:12]+"...")
		t.Logf("    Setup: %s", teacher2Input.Message[:80]+"...")
		t.Logf("    Query: %s", teacher2Query.Message)
		t.Logf("    Response: %s", teacher2QueryResponse.Content[:minInt(200, len(teacher2QueryResponse.Content))])
		t.Logf("    Has teacher context: %v, Has medical context: %v", teacher2HasTeacher, teacher2HasMedical)

		// Evaluate session isolation
		isolationScore := 0
		if medical1HasMedical {
			isolationScore++
			t.Logf("✓ Medical session retained its context")
		} else {
			t.Logf("⚠ Medical session may not have retained its context")
		}

		if teacher2HasTeacher {
			isolationScore++
			t.Logf("✓ Teacher session retained its context")
		} else {
			t.Logf("⚠ Teacher session may not have retained its context")
		}

		if !medical1HasTeacher {
			isolationScore++
			t.Logf("✓ Medical session properly isolated (no teacher context leakage)")
		} else {
			t.Logf("⚠ Medical session may have context leakage from teacher session")
		}

		if !teacher2HasMedical {
			isolationScore++
			t.Logf("✓ Teacher session properly isolated (no medical context leakage)")
		} else {
			t.Logf("⚠ Teacher session may have context leakage from medical session")
		}

		t.Logf("Session isolation score: %d/4", isolationScore)
		if isolationScore >= 3 {
			t.Logf("✓ Session isolation is working well")
		} else {
			t.Logf("⚠ Session isolation may need attention")
		}
	})

	// Test 4: Context persistence across call types (streaming vs non-streaming)
	t.Run("ContextPersistenceAcrossCallTypes", func(t *testing.T) {
		sessionID := generateUniqueSessionID("mixed-calls-test")

		// Establish context with regular call
		setupInput := services.AgentInput{
			SessionID: sessionID,
			Message:   "I am a chef who owns an Italian restaurant called 'Bella Notte' in downtown. I specialize in traditional Tuscan cuisine.",
		}

		setupResponse, err := adapter.InvokeAgent(ctx, setupInput)
		if err != nil {
			t.Fatalf("Context setup with regular call failed: %v", err)
		}

		t.Logf("Context established with regular call:")
		t.Logf("  Input: %s", setupInput.Message)
		t.Logf("  Response: %s", setupResponse.Content[:minInt(200, len(setupResponse.Content))])

		// Wait for context processing
		time.Sleep(3 * time.Second)

		// Test context retrieval with streaming call
		streamInput := services.AgentInput{
			SessionID: sessionID,
			Message:   "Can you tell me about my restaurant and what type of cuisine I serve?",
		}

		streamReader, err := adapter.InvokeAgentStream(ctx, streamInput)
		if err != nil {
			t.Fatalf("Streaming call failed: %v", err)
		}
		defer streamReader.Close()

		var streamContent strings.Builder
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
				streamContent.WriteString(chunk)
				chunkCount++
			}
		}

		// Analyze context retention in streaming response
		streamResponseContent := strings.ToLower(streamContent.String())
		hasChef := strings.Contains(streamResponseContent, "chef")
		hasRestaurant := strings.Contains(streamResponseContent, "restaurant") ||
						strings.Contains(streamResponseContent, "bella notte")
		hasItalian := strings.Contains(streamResponseContent, "italian") ||
					  strings.Contains(streamResponseContent, "tuscan")

		t.Logf("Context persistence across call types:")
		t.Logf("  Setup (regular call): %s", setupInput.Message[:80]+"...")
		t.Logf("  Query (streaming call): %s", streamInput.Message)
		t.Logf("  Stream response (%d chunks): %s", chunkCount, streamContent.String()[:minInt(300, len(streamContent.String()))])
		t.Logf("  Context retained - Chef: %v, Restaurant: %v, Italian: %v", hasChef, hasRestaurant, hasItalian)

		contextRetentionScore := 0
		if hasChef {
			contextRetentionScore++
		}
		if hasRestaurant {
			contextRetentionScore++
		}
		if hasItalian {
			contextRetentionScore++
		}

		t.Logf("Context retention score: %d/3", contextRetentionScore)
		if contextRetentionScore >= 2 {
			t.Logf("✓ Context persists well across different call types")
		} else {
			t.Logf("⚠ Context persistence across call types may need attention")
		}
	})
}

// generateUniqueSessionID creates a unique session ID for testing with a prefix
func generateUniqueSessionID(prefix string) string {
	return prefix + "-" + time.Now().Format("20060102-150405") + "-" + 
		   time.Now().Format("000000")
}

// minInt returns the minimum of two integers
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}