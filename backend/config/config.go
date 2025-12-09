package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all application configuration
type Config struct {
	Environment string
	Server      ServerConfig
	AWS         AWSConfig
	Bedrock     BedrockConfig
	WebSocket   WebSocketConfig
	Session     SessionConfig
	Logging     LoggingConfig
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port string
	Host string
}

// AWSConfig holds AWS configuration
type AWSConfig struct {
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
}

// BedrockConfig holds Bedrock Agent Core configuration
type BedrockConfig struct {
	AgentID          string
	AgentAliasID     string
	KnowledgeBaseID  string
	ModelID          string
	MaxRetries       int
	InitialBackoff   time.Duration
	MaxBackoff       time.Duration
	RequestTimeout   time.Duration
}

// WebSocketConfig holds WebSocket configuration
type WebSocketConfig struct {
	Timeout          time.Duration
	BufferSize       int
	ReadBufferSize   int
	WriteBufferSize  int
	StreamTimeout    time.Duration
	ChunkTimeout     time.Duration
}

// SessionConfig holds session configuration
type SessionConfig struct {
	Timeout time.Duration
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string
	Format string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		Environment: getEnv("ENVIRONMENT", "development"),
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
		},
		AWS: AWSConfig{
			Region:          getEnv("AWS_REGION", "ap-southeast-1"),
			AccessKeyID:     getEnv("AWS_ACCESS_KEY_ID", ""),
			SecretAccessKey: getEnv("AWS_SECRET_ACCESS_KEY", ""),
			SessionToken:    getEnv("AWS_SESSION_TOKEN", ""),
		},
		Bedrock: BedrockConfig{
			AgentID:          getEnv("BEDROCK_AGENT_ID", ""),
			AgentAliasID:     getEnv("BEDROCK_AGENT_ALIAS_ID", ""),
			KnowledgeBaseID:  getEnv("BEDROCK_KNOWLEDGE_BASE_ID", ""),
			ModelID:          getEnv("BEDROCK_MODEL_ID", "anthropic.claude-v2"),
			MaxRetries:       getEnvAsInt("BEDROCK_MAX_RETRIES", 3),
			InitialBackoff:   getEnvAsDuration("BEDROCK_INITIAL_BACKOFF", 1*time.Second),
			MaxBackoff:       getEnvAsDuration("BEDROCK_MAX_BACKOFF", 30*time.Second),
			RequestTimeout:   getEnvAsDuration("BEDROCK_REQUEST_TIMEOUT", 60*time.Second),
		},
		WebSocket: WebSocketConfig{
			Timeout:         getEnvAsDuration("WS_TIMEOUT", 30*time.Second),
			BufferSize:      getEnvAsInt("WS_BUFFER_SIZE", 8192),
			ReadBufferSize:  getEnvAsInt("WS_READ_BUFFER_SIZE", 1024),
			WriteBufferSize: getEnvAsInt("WS_WRITE_BUFFER_SIZE", 1024),
			StreamTimeout:   getEnvAsDuration("WS_STREAM_TIMEOUT", 5*time.Minute),
			ChunkTimeout:    getEnvAsDuration("WS_CHUNK_TIMEOUT", 30*time.Second),
		},
		Session: SessionConfig{
			Timeout: getEnvAsDuration("SESSION_TIMEOUT", 30*time.Minute),
		},
		Logging: LoggingConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "text"),
		},
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate environment
	if c.Environment != "development" && c.Environment != "production" && c.Environment != "test" {
		return fmt.Errorf("invalid environment: %s (must be development, production, or test)", c.Environment)
	}

	// Validate server configuration
	if c.Server.Port == "" {
		return fmt.Errorf("server port is required")
	}

	// Validate AWS region
	if c.AWS.Region == "" {
		return fmt.Errorf("AWS region is required")
	}

	// Validate Bedrock configuration (only in production)
	if c.Environment == "production" {
		if c.Bedrock.AgentID == "" {
			return fmt.Errorf("Bedrock agent ID is required in production")
		}
		if c.Bedrock.AgentAliasID == "" {
			return fmt.Errorf("Bedrock agent alias ID is required in production")
		}
	}

	// Validate WebSocket configuration
	if c.WebSocket.Timeout <= 0 {
		return fmt.Errorf("WebSocket timeout must be positive")
	}
	if c.WebSocket.BufferSize <= 0 {
		return fmt.Errorf("WebSocket buffer size must be positive")
	}

	// Validate session timeout
	if c.Session.Timeout <= 0 {
		return fmt.Errorf("session timeout must be positive")
	}

	return nil
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

// IsTest returns true if running in test mode
func (c *Config) IsTest() bool {
	return c.Environment == "test"
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getEnvAsInt gets an environment variable as an integer with a default value
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}

// getEnvAsDuration gets an environment variable as a duration with a default value
func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := time.ParseDuration(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}
