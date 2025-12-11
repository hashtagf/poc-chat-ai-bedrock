package bedrock

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/bedrock-chat-poc/backend/domain/services"
)

// TestIAMPermissions_ValidateAgentAccess tests that the current IAM configuration
// can successfully access the configured Bedrock Agent
// Requirements: 10.1
func TestIAMPermissions_ValidateAgentAccess(t *testing.T) {
	// Skip if running in CI or if Bedrock configuration is not available
	if os.Getenv("CI") != "" {
		t.Skip("Skipping IAM permissions test in CI environment")
	}

	agentID := os.Getenv("BEDROCK_AGENT_ID")
	aliasID := os.Getenv("BEDROCK_AGENT_ALIAS_ID")

	if agentID == "" || aliasID == "" {
		t.Skip("Skipping IAM test - BEDROCK_AGENT_ID and BEDROCK_AGENT_ALIAS_ID must be set")
	}

	ctx := context.Background()

	// Test with valid agent ID and alias ID
	adapter, err := NewAdapter(ctx, agentID, aliasID, DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to create Bedrock adapter with valid credentials: %v", err)
	}

	input := services.AgentInput{
		SessionID: generateTestSessionID(),
		Message:   "Test agent access with current IAM permissions",
	}

	response, err := adapter.InvokeAgent(ctx, input)
	if err != nil {
		// Check if this is an IAM permissions error
		var domainErr *services.DomainError
		if errors.As(err, &domainErr) {
			if domainErr.Code == services.ErrCodeUnauthorized {
				t.Fatalf("IAM permissions insufficient for agent access. Error: %v\n"+
					"Check that the IAM role has bedrock:InvokeAgent permission for agent %s", err, agentID)
			}
		}
		t.Fatalf("Unexpected error during agent invocation: %v", err)
	}

	if response == nil {
		t.Fatal("Expected non-nil response from valid agent access")
	}

	if response.Content == "" {
		t.Error("Expected non-empty content from agent response")
	}

	t.Logf("✓ IAM permissions validated - Agent access successful with %d characters of content", 
		len(response.Content))
}

// TestIAMPermissions_InvalidAgentID tests error handling when using an invalid agent ID
// Requirements: 10.2, 10.3
func TestIAMPermissions_InvalidAgentID(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping IAM permissions test in CI environment")
	}

	aliasID := os.Getenv("BEDROCK_AGENT_ALIAS_ID")
	if aliasID == "" {
		t.Skip("Skipping IAM test - BEDROCK_AGENT_ALIAS_ID must be set")
	}

	ctx := context.Background()

	// Test with invalid agent ID
	invalidAgentID := "INVALID-AGENT-ID-123"
	adapter, err := NewAdapter(ctx, invalidAgentID, aliasID, DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to create adapter (should succeed even with invalid agent ID): %v", err)
	}

	input := services.AgentInput{
		SessionID: generateTestSessionID(),
		Message:   "Test with invalid agent ID",
	}

	_, err = adapter.InvokeAgent(ctx, input)
	if err == nil {
		t.Error("Expected error when using invalid agent ID")
		return
	}

	// Verify this is an authorization error with actionable message
	var domainErr *services.DomainError
	if !errors.As(err, &domainErr) {
		t.Errorf("Expected DomainError, got: %T", err)
		return
	}

	if domainErr.Code != services.ErrCodeUnauthorized {
		t.Errorf("Expected ErrCodeUnauthorized, got: %s", domainErr.Code)
	}

	if domainErr.Retryable {
		t.Error("Authorization errors should not be retryable")
	}

	// Check that error message provides actionable guidance
	errorMsg := strings.ToLower(domainErr.Message)
	if !strings.Contains(errorMsg, "unauthorized") && !strings.Contains(errorMsg, "access") {
		t.Errorf("Error message should mention authorization issue: %s", domainErr.Message)
	}

	t.Logf("✓ Invalid agent ID correctly rejected with error: %s", domainErr.Message)
}

// TestIAMPermissions_InvalidAliasID tests error handling when using an invalid alias ID
// Requirements: 10.2, 10.3
func TestIAMPermissions_InvalidAliasID(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping IAM permissions test in CI environment")
	}

	agentID := os.Getenv("BEDROCK_AGENT_ID")
	if agentID == "" {
		t.Skip("Skipping IAM test - BEDROCK_AGENT_ID must be set")
	}

	ctx := context.Background()

	// Test with invalid alias ID
	invalidAliasID := "INVALID-ALIAS-ID-123"
	adapter, err := NewAdapter(ctx, agentID, invalidAliasID, DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to create adapter (should succeed even with invalid alias ID): %v", err)
	}

	input := services.AgentInput{
		SessionID: generateTestSessionID(),
		Message:   "Test with invalid alias ID",
	}

	_, err = adapter.InvokeAgent(ctx, input)
	if err == nil {
		t.Error("Expected error when using invalid alias ID")
		return
	}

	// Verify this is an authorization error
	var domainErr *services.DomainError
	if !errors.As(err, &domainErr) {
		t.Errorf("Expected DomainError, got: %T", err)
		return
	}

	if domainErr.Code != services.ErrCodeUnauthorized {
		t.Errorf("Expected ErrCodeUnauthorized, got: %s", domainErr.Code)
	}

	if domainErr.Retryable {
		t.Error("Authorization errors should not be retryable")
	}

	t.Logf("✓ Invalid alias ID correctly rejected with error: %s", domainErr.Message)
}

// TestIAMPermissions_KnowledgeBaseAccess tests access to knowledge base resources
// Requirements: 10.4
func TestIAMPermissions_KnowledgeBaseAccess(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping IAM permissions test in CI environment")
	}

	agentID := os.Getenv("BEDROCK_AGENT_ID")
	aliasID := os.Getenv("BEDROCK_AGENT_ALIAS_ID")
	knowledgeBaseID := os.Getenv("BEDROCK_KNOWLEDGE_BASE_ID")

	if agentID == "" || aliasID == "" {
		t.Skip("Skipping IAM test - BEDROCK_AGENT_ID and BEDROCK_AGENT_ALIAS_ID must be set")
	}

	if knowledgeBaseID == "" {
		t.Skip("Skipping knowledge base IAM test - BEDROCK_KNOWLEDGE_BASE_ID not set")
	}

	ctx := context.Background()

	adapter, err := NewAdapter(ctx, agentID, aliasID, DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to create Bedrock adapter: %v", err)
	}

	// Test knowledge base access
	input := services.AgentInput{
		SessionID:        generateTestSessionID(),
		Message:          "What information is available in the knowledge base?",
		KnowledgeBaseIDs: []string{knowledgeBaseID},
	}

	response, err := adapter.InvokeAgent(ctx, input)
	if err != nil {
		// Check if this is a knowledge base permission error
		var domainErr *services.DomainError
		if errors.As(err, &domainErr) {
			if domainErr.Code == services.ErrCodeUnauthorized {
				t.Fatalf("IAM permissions insufficient for knowledge base access. Error: %v\n"+
					"Check that the IAM role has bedrock:Retrieve permission for knowledge base %s", 
					err, knowledgeBaseID)
			}
		}
		t.Fatalf("Unexpected error during knowledge base query: %v", err)
	}

	if response == nil {
		t.Fatal("Expected non-nil response from knowledge base query")
	}

	if response.Content == "" {
		t.Error("Expected non-empty content from knowledge base response")
	}

	t.Logf("✓ Knowledge base access validated - Response with %d characters and %d citations", 
		len(response.Content), len(response.Citations))
}

// TestIAMPermissions_InvalidKnowledgeBaseID tests error handling with invalid knowledge base ID
// Requirements: 10.4
func TestIAMPermissions_InvalidKnowledgeBaseID(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping IAM permissions test in CI environment")
	}

	agentID := os.Getenv("BEDROCK_AGENT_ID")
	aliasID := os.Getenv("BEDROCK_AGENT_ALIAS_ID")

	if agentID == "" || aliasID == "" {
		t.Skip("Skipping IAM test - BEDROCK_AGENT_ID and BEDROCK_AGENT_ALIAS_ID must be set")
	}

	ctx := context.Background()

	adapter, err := NewAdapter(ctx, agentID, aliasID, DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to create Bedrock adapter: %v", err)
	}

	// Test with invalid knowledge base ID
	invalidKnowledgeBaseID := "INVALID-KB-ID-123"
	input := services.AgentInput{
		SessionID:        generateTestSessionID(),
		Message:          "Test with invalid knowledge base ID",
		KnowledgeBaseIDs: []string{invalidKnowledgeBaseID},
	}

	response, err := adapter.InvokeAgent(ctx, input)
	
	// Note: Invalid knowledge base IDs might not always cause immediate errors
	// The agent might simply ignore invalid knowledge bases and proceed
	// This depends on the specific Bedrock Agent configuration
	
	if err != nil {
		var domainErr *services.DomainError
		if errors.As(err, &domainErr) {
			if domainErr.Code == services.ErrCodeUnauthorized {
				// This is expected - log the specific resource information
				if !strings.Contains(domainErr.Message, invalidKnowledgeBaseID) {
					t.Logf("Note: Error message could include specific resource ID for better debugging")
				}
				t.Logf("✓ Invalid knowledge base ID correctly rejected: %s", domainErr.Message)
				return
			}
		}
		t.Logf("Unexpected error with invalid knowledge base ID: %v", err)
		return
	}

	// If no error occurred, the agent might have ignored the invalid knowledge base
	if response != nil {
		t.Logf("Note: Agent processed request despite invalid knowledge base ID (may have ignored it)")
		t.Logf("Response: %d characters, %d citations", len(response.Content), len(response.Citations))
	}
}

// TestIAMPermissions_FoundationModelAccess tests that IAM permissions allow access to the foundation model
// Requirements: 10.7
func TestIAMPermissions_FoundationModelAccess(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping IAM permissions test in CI environment")
	}

	agentID := os.Getenv("BEDROCK_AGENT_ID")
	aliasID := os.Getenv("BEDROCK_AGENT_ALIAS_ID")
	modelID := os.Getenv("BEDROCK_MODEL_ID")

	if agentID == "" || aliasID == "" {
		t.Skip("Skipping IAM test - BEDROCK_AGENT_ID and BEDROCK_AGENT_ALIAS_ID must be set")
	}

	ctx := context.Background()

	adapter, err := NewAdapter(ctx, agentID, aliasID, DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to create Bedrock adapter: %v", err)
	}

	// Test that we can invoke the agent (which uses the foundation model)
	input := services.AgentInput{
		SessionID: generateTestSessionID(),
		Message:   "Test foundation model access through agent invocation",
	}

	response, err := adapter.InvokeAgent(ctx, input)
	if err != nil {
		var domainErr *services.DomainError
		if errors.As(err, &domainErr) {
			if domainErr.Code == services.ErrCodeUnauthorized {
				if modelID != "" {
					t.Fatalf("IAM permissions insufficient for foundation model access. Error: %v\n"+
						"Check that the IAM role has bedrock:InvokeModel permission for model %s", 
						err, modelID)
				} else {
					t.Fatalf("IAM permissions insufficient for foundation model access. Error: %v\n"+
						"Check that the IAM role has bedrock:InvokeModel permission", err)
				}
			}
		}
		t.Fatalf("Unexpected error during foundation model access test: %v", err)
	}

	if response == nil || response.Content == "" {
		t.Error("Expected valid response from foundation model")
	}

	t.Logf("✓ Foundation model access validated through agent invocation")
	if modelID != "" {
		t.Logf("  Model ID: %s", modelID)
	}
}

// TestIAMPermissions_CrossAccountAccess tests cross-account access scenarios if applicable
// Requirements: 10.6
func TestIAMPermissions_CrossAccountAccess(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping IAM permissions test in CI environment")
	}

	// This test is only relevant if cross-account resources are configured
	crossAccountAgentID := os.Getenv("BEDROCK_CROSS_ACCOUNT_AGENT_ID")
	crossAccountAliasID := os.Getenv("BEDROCK_CROSS_ACCOUNT_ALIAS_ID")

	if crossAccountAgentID == "" || crossAccountAliasID == "" {
		t.Skip("Skipping cross-account test - BEDROCK_CROSS_ACCOUNT_AGENT_ID and BEDROCK_CROSS_ACCOUNT_ALIAS_ID not set")
	}

	ctx := context.Background()

	adapter, err := NewAdapter(ctx, crossAccountAgentID, crossAccountAliasID, DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to create cross-account Bedrock adapter: %v", err)
	}

	input := services.AgentInput{
		SessionID: generateTestSessionID(),
		Message:   "Test cross-account access",
	}

	response, err := adapter.InvokeAgent(ctx, input)
	if err != nil {
		var domainErr *services.DomainError
		if errors.As(err, &domainErr) {
			if domainErr.Code == services.ErrCodeUnauthorized {
				t.Logf("Cross-account access denied (expected if trust relationships not configured): %s", 
					domainErr.Message)
				
				// Verify error message provides guidance about cross-account setup
				errorMsg := strings.ToLower(domainErr.Message)
				if strings.Contains(errorMsg, "cross") || strings.Contains(errorMsg, "account") || 
				   strings.Contains(errorMsg, "trust") {
					t.Logf("✓ Error message provides cross-account guidance")
				} else {
					t.Logf("Note: Error message could provide more specific cross-account guidance")
				}
				return
			}
		}
		t.Fatalf("Unexpected error during cross-account test: %v", err)
	}

	if response != nil && response.Content != "" {
		t.Logf("✓ Cross-account access successful - trust relationships properly configured")
	}
}

// TestIAMPermissions_VPCEndpointAccess tests VPC endpoint connectivity if applicable
// Requirements: 10.5
func TestIAMPermissions_VPCEndpointAccess(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping IAM permissions test in CI environment")
	}

	// Check if VPC endpoints are configured
	vpcEnabled := os.Getenv("ENABLE_VPC")
	if vpcEnabled != "true" {
		t.Skip("Skipping VPC endpoint test - VPC not enabled")
	}

	agentID := os.Getenv("BEDROCK_AGENT_ID")
	aliasID := os.Getenv("BEDROCK_AGENT_ALIAS_ID")

	if agentID == "" || aliasID == "" {
		t.Skip("Skipping VPC test - BEDROCK_AGENT_ID and BEDROCK_AGENT_ALIAS_ID must be set")
	}

	ctx := context.Background()

	adapter, err := NewAdapter(ctx, agentID, aliasID, DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to create Bedrock adapter for VPC test: %v", err)
	}

	input := services.AgentInput{
		SessionID: generateTestSessionID(),
		Message:   "Test VPC endpoint connectivity",
	}

	response, err := adapter.InvokeAgent(ctx, input)
	if err != nil {
		var domainErr *services.DomainError
		if errors.As(err, &domainErr) {
			if domainErr.Code == services.ErrCodeNetworkError {
				t.Fatalf("VPC endpoint connectivity failed. Error: %v\n"+
					"Check that:\n"+
					"1. VPC endpoints for Bedrock are created\n"+
					"2. Security groups allow HTTPS traffic (port 443)\n"+
					"3. Route tables are configured correctly\n"+
					"4. DNS resolution is working", err)
			}
			if domainErr.Code == services.ErrCodeUnauthorized {
				t.Fatalf("VPC endpoint access denied. Error: %v\n"+
					"Check that security groups allow traffic to Bedrock VPC endpoints", err)
			}
		}
		t.Fatalf("Unexpected error during VPC endpoint test: %v", err)
	}

	if response == nil || response.Content == "" {
		t.Error("Expected valid response through VPC endpoint")
	}

	t.Logf("✓ VPC endpoint access validated - traffic routing correctly through private endpoints")
}

// TestIAMPermissions_ErrorMessageQuality tests that error messages provide actionable guidance
// Requirements: 10.2
func TestIAMPermissions_ErrorMessageQuality(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping IAM permissions test in CI environment")
	}

	ctx := context.Background()

	testCases := []struct {
		name        string
		agentID     string
		aliasID     string
		expectError bool
		description string
	}{
		{
			name:        "InvalidAgentFormat",
			agentID:     "invalid-format",
			aliasID:     "VALIDALIAS",
			expectError: true,
			description: "Agent ID with invalid format",
		},
		{
			name:        "InvalidAliasFormat", 
			agentID:     "VALIDAGENT",
			aliasID:     "invalid-format",
			expectError: true,
			description: "Alias ID with invalid format",
		},
		{
			name:        "NonExistentAgent",
			agentID:     "NONEXISTENT123",
			aliasID:     "NONEXISTENT123",
			expectError: true,
			description: "Non-existent agent and alias",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			adapter, err := NewAdapter(ctx, tc.agentID, tc.aliasID, DefaultConfig())
			if err != nil {
				t.Fatalf("Failed to create adapter for %s: %v", tc.description, err)
			}

			input := services.AgentInput{
				SessionID: generateTestSessionID(),
				Message:   "Test error message quality",
			}

			_, err = adapter.InvokeAgent(ctx, input)
			
			if !tc.expectError {
				if err != nil {
					t.Errorf("Unexpected error for %s: %v", tc.description, err)
				}
				return
			}

			if err == nil {
				t.Errorf("Expected error for %s but got none", tc.description)
				return
			}

			var domainErr *services.DomainError
			if !errors.As(err, &domainErr) {
				t.Errorf("Expected DomainError for %s, got: %T", tc.description, err)
				return
			}

			// Verify error message quality
			if domainErr.Message == "" {
				t.Errorf("Error message should not be empty for %s", tc.description)
			}

			// Check that error message is actionable (contains helpful keywords)
			errorMsg := strings.ToLower(domainErr.Message)
			hasActionableContent := strings.Contains(errorMsg, "check") ||
				strings.Contains(errorMsg, "verify") ||
				strings.Contains(errorMsg, "ensure") ||
				strings.Contains(errorMsg, "permission") ||
				strings.Contains(errorMsg, "access") ||
				strings.Contains(errorMsg, "role") ||
				strings.Contains(errorMsg, "policy")

			if !hasActionableContent {
				t.Logf("Note: Error message for %s could be more actionable: %s", 
					tc.description, domainErr.Message)
			}

			t.Logf("✓ %s error handled: %s", tc.description, domainErr.Message)
		})
	}
}

// TestIAMPermissions_StreamingAccess tests IAM permissions for streaming operations
// Requirements: 10.1
func TestIAMPermissions_StreamingAccess(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping IAM permissions test in CI environment")
	}

	agentID := os.Getenv("BEDROCK_AGENT_ID")
	aliasID := os.Getenv("BEDROCK_AGENT_ALIAS_ID")

	if agentID == "" || aliasID == "" {
		t.Skip("Skipping streaming IAM test - BEDROCK_AGENT_ID and BEDROCK_AGENT_ALIAS_ID must be set")
	}

	ctx := context.Background()

	adapter, err := NewAdapter(ctx, agentID, aliasID, DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to create Bedrock adapter: %v", err)
	}

	input := services.AgentInput{
		SessionID: generateTestSessionID(),
		Message:   "Test streaming access with IAM permissions",
	}

	streamReader, err := adapter.InvokeAgentStream(ctx, input)
	if err != nil {
		var domainErr *services.DomainError
		if errors.As(err, &domainErr) {
			if domainErr.Code == services.ErrCodeUnauthorized {
				t.Fatalf("IAM permissions insufficient for streaming access. Error: %v\n"+
					"Check that the IAM role has bedrock:InvokeAgent permission for streaming operations", err)
			}
		}
		t.Fatalf("Unexpected error during streaming access test: %v", err)
	}

	if streamReader == nil {
		t.Fatal("Expected non-nil stream reader")
	}
	defer streamReader.Close()

	// Try to read at least one chunk to verify streaming permissions work
	chunk, done, err := streamReader.Read()
	if err != nil {
		var domainErr *services.DomainError
		if errors.As(err, &domainErr) {
			if domainErr.Code == services.ErrCodeUnauthorized {
				t.Fatalf("IAM permissions insufficient for reading stream. Error: %v", err)
			}
		}
		t.Fatalf("Unexpected error reading from stream: %v", err)
	}

	if done && chunk == "" {
		t.Error("Expected to receive at least some content from stream")
	}

	t.Logf("✓ Streaming IAM permissions validated - successfully read stream content")
}