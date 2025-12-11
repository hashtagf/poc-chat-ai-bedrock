package bedrock

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/bedrock-chat-poc/backend/config"
	"github.com/bedrock-chat-poc/backend/domain/services"
)

// TestEnvironmentConfiguration_Development tests development environment setup
func TestEnvironmentConfiguration_Development(t *testing.T) {
	// Save original environment
	originalEnv := saveEnvironment()
	defer restoreEnvironment(originalEnv)

	// Set development environment variables
	setDevelopmentEnvironment()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load development configuration: %v", err)
	}

	// Validate development-specific settings
	if cfg.Environment != "development" {
		t.Errorf("Expected environment 'development', got '%s'", cfg.Environment)
	}
	if cfg.AWS.Region != "us-east-1" {
		t.Errorf("Expected AWS region 'us-east-1', got '%s'", cfg.AWS.Region)
	}
	if cfg.Bedrock.AgentID == "" {
		t.Error("Agent ID should be set in development")
	}
	if cfg.Bedrock.AgentAliasID == "" {
		t.Error("Agent alias ID should be set in development")
	}

	// Test adapter creation with development config
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	adapter, err := NewAdapter(ctx, cfg.Bedrock.AgentID, cfg.Bedrock.AgentAliasID, AdapterConfig{
		MaxRetries:     cfg.Bedrock.MaxRetries,
		InitialBackoff: cfg.Bedrock.InitialBackoff,
		MaxBackoff:     cfg.Bedrock.MaxBackoff,
		RequestTimeout: cfg.Bedrock.RequestTimeout,
	})

	if err != nil {
		// In development, we might not have valid AWS credentials
		// Check if it's a credential error vs configuration error
		if strings.Contains(err.Error(), "credentials") || strings.Contains(err.Error(), "authentication") {
			t.Logf("Development environment: AWS credentials not available, skipping adapter test: %v", err)
			return
		}
		t.Fatalf("Failed to create adapter in development: %v", err)
	}

	if adapter == nil {
		t.Error("Adapter should be created successfully in development")
	}

	// Test basic functionality if adapter was created
	input := services.AgentInput{
		SessionID: generateTestSessionID(),
		Message:   "Test development environment",
	}

	// This might fail due to credentials, but we're testing configuration
	_, err = adapter.InvokeAgent(ctx, input)
	if err != nil {
		t.Logf("Development environment: Agent invocation failed (expected if no AWS credentials): %v", err)
	}
}

// TestEnvironmentConfiguration_Staging tests staging environment connectivity
func TestEnvironmentConfiguration_Staging(t *testing.T) {
	if !isStagingEnvironment() {
		t.Skip("Skipping staging test - not in staging environment")
	}

	// Save original environment
	originalEnv := saveEnvironment()
	defer restoreEnvironment(originalEnv)

	// Set staging environment variables
	setStagingEnvironment()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load staging configuration: %v", err)
	}

	// Validate staging-specific settings
	if cfg.Environment != "staging" {
		t.Errorf("Expected environment 'staging', got '%s'", cfg.Environment)
	}
	if cfg.AWS.Region != "us-east-1" {
		t.Errorf("Expected AWS region 'us-east-1', got '%s'", cfg.AWS.Region)
	}
	if cfg.Bedrock.AgentID == "" {
		t.Error("Agent ID should be set in staging")
	}
	if cfg.Bedrock.AgentAliasID == "" {
		t.Error("Agent alias ID should be set in staging")
	}

	// Test adapter creation with staging config
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	adapter, err := NewAdapter(ctx, cfg.Bedrock.AgentID, cfg.Bedrock.AgentAliasID, AdapterConfig{
		MaxRetries:     cfg.Bedrock.MaxRetries,
		InitialBackoff: cfg.Bedrock.InitialBackoff,
		MaxBackoff:     cfg.Bedrock.MaxBackoff,
		RequestTimeout: cfg.Bedrock.RequestTimeout,
	})
	if err != nil {
		t.Fatalf("Failed to create adapter in staging: %v", err)
	}
	if adapter == nil {
		t.Error("Adapter should be created successfully in staging")
	}

	// Test connectivity with staging resources
	input := services.AgentInput{
		SessionID: generateTestSessionID(),
		Message:   "Test staging environment connectivity",
	}

	response, err := adapter.InvokeAgent(ctx, input)
	if err != nil {
		t.Errorf("Agent invocation should succeed in staging: %v", err)
	}
	if response != nil {
		if response.Content == "" {
			t.Error("Response should contain content")
		}
		if response.RequestID == "" {
			t.Error("Response should contain request ID")
		}
		t.Logf("Staging environment test successful - Response length: %d", len(response.Content))
	}
}

// TestEnvironmentConfiguration_Production tests production VPC endpoint configuration
func TestEnvironmentConfiguration_Production(t *testing.T) {
	if !isProductionEnvironment() {
		t.Skip("Skipping production test - not in production environment")
	}

	// Save original environment
	originalEnv := saveEnvironment()
	defer restoreEnvironment(originalEnv)

	// Set production environment variables
	setProductionEnvironment()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load production configuration: %v", err)
	}

	// Validate production-specific settings
	if cfg.Environment != "production" {
		t.Errorf("Expected environment 'production', got '%s'", cfg.Environment)
	}
	if cfg.AWS.Region != "us-east-1" {
		t.Errorf("Expected AWS region 'us-east-1', got '%s'", cfg.AWS.Region)
	}
	if cfg.Bedrock.AgentID == "" {
		t.Error("Agent ID is required in production")
	}
	if cfg.Bedrock.AgentAliasID == "" {
		t.Error("Agent alias ID is required in production")
	}

	// Test adapter creation with production config
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	adapter, err := NewAdapter(ctx, cfg.Bedrock.AgentID, cfg.Bedrock.AgentAliasID, AdapterConfig{
		MaxRetries:     cfg.Bedrock.MaxRetries,
		InitialBackoff: cfg.Bedrock.InitialBackoff,
		MaxBackoff:     cfg.Bedrock.MaxBackoff,
		RequestTimeout: cfg.Bedrock.RequestTimeout,
	})
	if err != nil {
		t.Fatalf("Failed to create adapter in production: %v", err)
	}
	if adapter == nil {
		t.Error("Adapter should be created successfully in production")
	}

	// Test VPC endpoint connectivity
	input := services.AgentInput{
		SessionID: generateTestSessionID(),
		Message:   "Test production VPC endpoint connectivity",
	}

	response, err := adapter.InvokeAgent(ctx, input)
	if err != nil {
		t.Errorf("Agent invocation should succeed in production: %v", err)
	}
	if response != nil {
		if response.Content == "" {
			t.Error("Response should contain content")
		}
		if response.RequestID == "" {
			t.Error("Response should contain request ID")
		}
		t.Logf("Production environment test successful - Response length: %d", len(response.Content))
	}
}

// TestValidateAllRequiredEnvironmentVariables tests all required environment variables
func TestValidateAllRequiredEnvironmentVariables(t *testing.T) {
	tests := []struct {
		name        string
		environment string
		setupFunc   func()
		required    []string
		optional    []string
	}{
		{
			name:        "Development Environment Variables",
			environment: "development",
			setupFunc:   setDevelopmentEnvironment,
			required: []string{
				"ENVIRONMENT",
				"SERVER_PORT",
				"SERVER_HOST",
				"AWS_REGION",
			},
			optional: []string{
				"BEDROCK_AGENT_ID",
				"BEDROCK_AGENT_ALIAS_ID",
				"BEDROCK_KNOWLEDGE_BASE_ID",
				"AWS_ACCESS_KEY_ID",
				"AWS_SECRET_ACCESS_KEY",
			},
		},
		{
			name:        "Staging Environment Variables",
			environment: "staging",
			setupFunc:   setStagingEnvironment,
			required: []string{
				"ENVIRONMENT",
				"SERVER_PORT",
				"SERVER_HOST",
				"AWS_REGION",
				"BEDROCK_AGENT_ID",
				"BEDROCK_AGENT_ALIAS_ID",
			},
			optional: []string{
				"BEDROCK_KNOWLEDGE_BASE_ID",
				"AWS_ACCESS_KEY_ID",
				"AWS_SECRET_ACCESS_KEY",
			},
		},
		{
			name:        "Production Environment Variables",
			environment: "production",
			setupFunc:   setProductionEnvironment,
			required: []string{
				"ENVIRONMENT",
				"SERVER_PORT",
				"SERVER_HOST",
				"AWS_REGION",
				"BEDROCK_AGENT_ID",
				"BEDROCK_AGENT_ALIAS_ID",
			},
			optional: []string{
				"BEDROCK_KNOWLEDGE_BASE_ID",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original environment
			originalEnv := saveEnvironment()
			defer restoreEnvironment(originalEnv)

			// Setup environment
			tt.setupFunc()

			// Load configuration
			cfg, err := config.Load()
			if err != nil {
				t.Fatalf("Configuration should load successfully: %v", err)
			}

			// Validate environment matches
			if cfg.Environment != tt.environment {
				t.Errorf("Expected environment '%s', got '%s'", tt.environment, cfg.Environment)
			}

			// Check required variables are set
			for _, envVar := range tt.required {
				value := os.Getenv(envVar)
				if value == "" {
					t.Errorf("Required environment variable %s should be set", envVar)
				}
			}

			// Check optional variables (log if missing but don't fail)
			for _, envVar := range tt.optional {
				value := os.Getenv(envVar)
				if value == "" {
					t.Logf("Optional environment variable %s is not set", envVar)
				}
			}

			// Validate configuration structure
			if cfg.Server.Port == "" {
				t.Error("Server port should be configured")
			}
			if cfg.Server.Host == "" {
				t.Error("Server host should be configured")
			}
			if cfg.AWS.Region == "" {
				t.Error("AWS region should be configured")
			}

			// Environment-specific validations
			switch tt.environment {
			case "production":
				if cfg.Bedrock.AgentID == "" {
					t.Error("Bedrock agent ID is required in production")
				}
				if cfg.Bedrock.AgentAliasID == "" {
					t.Error("Bedrock agent alias ID is required in production")
				}
			case "staging":
				if cfg.Bedrock.AgentID == "" {
					t.Error("Bedrock agent ID should be set in staging")
				}
				if cfg.Bedrock.AgentAliasID == "" {
					t.Error("Bedrock agent alias ID should be set in staging")
				}
			}

			// Validate timeout and retry configurations
			if cfg.Bedrock.MaxRetries <= 0 {
				t.Error("Max retries should be positive")
			}
			if cfg.Bedrock.InitialBackoff <= 0 {
				t.Error("Initial backoff should be positive")
			}
			if cfg.Bedrock.MaxBackoff <= 0 {
				t.Error("Max backoff should be positive")
			}
			if cfg.Bedrock.RequestTimeout <= 0 {
				t.Error("Request timeout should be positive")
			}
		})
	}
}

// TestEnvironmentConfiguration_InvalidConfiguration tests invalid configurations
func TestEnvironmentConfiguration_InvalidConfiguration(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func()
		wantError string
	}{
		{
			name: "Invalid Environment",
			setupFunc: func() {
				os.Setenv("ENVIRONMENT", "invalid")
				os.Setenv("SERVER_PORT", "8080")
				os.Setenv("AWS_REGION", "us-east-1")
			},
			wantError: "invalid environment",
		},
		{
			name: "Production Missing Agent ID",
			setupFunc: func() {
				os.Setenv("ENVIRONMENT", "production")
				os.Setenv("SERVER_PORT", "8080")
				os.Setenv("AWS_REGION", "us-east-1")
				os.Unsetenv("BEDROCK_AGENT_ID")
			},
			wantError: "Bedrock agent ID is required in production",
		},
		{
			name: "Staging Missing Agent ID",
			setupFunc: func() {
				os.Setenv("ENVIRONMENT", "staging")
				os.Setenv("SERVER_PORT", "8080")
				os.Setenv("AWS_REGION", "us-east-1")
				os.Unsetenv("BEDROCK_AGENT_ID")
			},
			wantError: "Bedrock agent ID is required in staging",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original environment
			originalEnv := saveEnvironment()
			defer restoreEnvironment(originalEnv)

			// Clear environment
			clearEnvironment()

			// Setup test environment
			tt.setupFunc()

			// Load configuration - should fail
			_, err := config.Load()
			if err == nil {
				t.Error("Configuration loading should fail")
			} else if !strings.Contains(err.Error(), tt.wantError) {
				t.Errorf("Error should contain '%s', got: %v", tt.wantError, err)
			}
		})
	}
}

// Helper functions for environment management

func saveEnvironment() map[string]string {
	env := make(map[string]string)
	envVars := []string{
		"ENVIRONMENT", "SERVER_PORT", "SERVER_HOST", "AWS_REGION",
		"AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY", "AWS_SESSION_TOKEN",
		"BEDROCK_AGENT_ID", "BEDROCK_AGENT_ALIAS_ID", "BEDROCK_KNOWLEDGE_BASE_ID",
		"BEDROCK_MODEL_ID", "BEDROCK_MAX_RETRIES", "BEDROCK_INITIAL_BACKOFF",
		"BEDROCK_MAX_BACKOFF", "BEDROCK_REQUEST_TIMEOUT",
	}

	for _, key := range envVars {
		env[key] = os.Getenv(key)
	}
	return env
}

func restoreEnvironment(env map[string]string) {
	for key, value := range env {
		if value == "" {
			os.Unsetenv(key)
		} else {
			os.Setenv(key, value)
		}
	}
}

func clearEnvironment() {
	envVars := []string{
		"ENVIRONMENT", "SERVER_PORT", "SERVER_HOST", "AWS_REGION",
		"AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY", "AWS_SESSION_TOKEN",
		"BEDROCK_AGENT_ID", "BEDROCK_AGENT_ALIAS_ID", "BEDROCK_KNOWLEDGE_BASE_ID",
		"BEDROCK_MODEL_ID", "BEDROCK_MAX_RETRIES", "BEDROCK_INITIAL_BACKOFF",
		"BEDROCK_MAX_BACKOFF", "BEDROCK_REQUEST_TIMEOUT",
	}

	for _, key := range envVars {
		os.Unsetenv(key)
	}
}

func setDevelopmentEnvironment() {
	os.Setenv("ENVIRONMENT", "development")
	os.Setenv("SERVER_PORT", "8080")
	os.Setenv("SERVER_HOST", "0.0.0.0")
	os.Setenv("AWS_REGION", "us-east-1")
	os.Setenv("BEDROCK_AGENT_ID", "W6R84XTD2X")
	os.Setenv("BEDROCK_AGENT_ALIAS_ID", "TXENIZDWOS")
	os.Setenv("BEDROCK_KNOWLEDGE_BASE_ID", "AQ5JOUEIGF")
	os.Setenv("BEDROCK_MODEL_ID", "us.amazon.nova-2-lite-v1:0")
	os.Setenv("BEDROCK_MAX_RETRIES", "3")
	os.Setenv("BEDROCK_INITIAL_BACKOFF", "1s")
	os.Setenv("BEDROCK_MAX_BACKOFF", "30s")
	os.Setenv("BEDROCK_REQUEST_TIMEOUT", "60s")
}

func setStagingEnvironment() {
	os.Setenv("ENVIRONMENT", "staging")
	os.Setenv("SERVER_PORT", "8080")
	os.Setenv("SERVER_HOST", "0.0.0.0")
	os.Setenv("AWS_REGION", "us-east-1")
	os.Setenv("BEDROCK_AGENT_ID", "staging-agent-id")
	os.Setenv("BEDROCK_AGENT_ALIAS_ID", "staging-alias-id")
	os.Setenv("BEDROCK_KNOWLEDGE_BASE_ID", "staging-kb-id")
	os.Setenv("BEDROCK_MODEL_ID", "anthropic.claude-v2")
	os.Setenv("BEDROCK_MAX_RETRIES", "3")
	os.Setenv("BEDROCK_INITIAL_BACKOFF", "1s")
	os.Setenv("BEDROCK_MAX_BACKOFF", "30s")
	os.Setenv("BEDROCK_REQUEST_TIMEOUT", "60s")
}

func setProductionEnvironment() {
	os.Setenv("ENVIRONMENT", "production")
	os.Setenv("SERVER_PORT", "8080")
	os.Setenv("SERVER_HOST", "0.0.0.0")
	os.Setenv("AWS_REGION", "us-east-1")
	os.Setenv("BEDROCK_AGENT_ID", "prod-agent-id")
	os.Setenv("BEDROCK_AGENT_ALIAS_ID", "prod-alias-id")
	os.Setenv("BEDROCK_KNOWLEDGE_BASE_ID", "prod-kb-id")
	os.Setenv("BEDROCK_MODEL_ID", "anthropic.claude-v2")
	os.Setenv("BEDROCK_MAX_RETRIES", "5")
	os.Setenv("BEDROCK_INITIAL_BACKOFF", "2s")
	os.Setenv("BEDROCK_MAX_BACKOFF", "60s")
	os.Setenv("BEDROCK_REQUEST_TIMEOUT", "120s")
}

func isStagingEnvironment() bool {
	return os.Getenv("ENVIRONMENT") == "staging" || 
		   os.Getenv("TEST_ENVIRONMENT") == "staging"
}

func isProductionEnvironment() bool {
	return os.Getenv("ENVIRONMENT") == "production" || 
		   os.Getenv("TEST_ENVIRONMENT") == "production"
}

