package config

import (
	"os"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	// Save original environment
	originalEnv := make(map[string]string)
	envVars := []string{
		"ENVIRONMENT", "SERVER_PORT", "AWS_REGION",
		"BEDROCK_AGENT_ID", "BEDROCK_AGENT_ALIAS_ID",
		"WS_TIMEOUT", "SESSION_TIMEOUT",
	}
	for _, key := range envVars {
		originalEnv[key] = os.Getenv(key)
	}

	// Restore environment after test
	defer func() {
		for key, value := range originalEnv {
			if value == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, value)
			}
		}
	}()

	tests := []struct {
		name    string
		envVars map[string]string
		wantErr bool
	}{
		{
			name: "valid development configuration",
			envVars: map[string]string{
				"ENVIRONMENT": "development",
				"SERVER_PORT": "8080",
				"AWS_REGION":  "us-east-1",
			},
			wantErr: false,
		},
		{
			name: "valid production configuration",
			envVars: map[string]string{
				"ENVIRONMENT":            "production",
				"SERVER_PORT":            "8080",
				"AWS_REGION":             "us-east-1",
				"BEDROCK_AGENT_ID":       "test-agent-id",
				"BEDROCK_AGENT_ALIAS_ID": "test-alias-id",
			},
			wantErr: false,
		},
		{
			name: "missing bedrock config in production",
			envVars: map[string]string{
				"ENVIRONMENT": "production",
				"SERVER_PORT": "8080",
				"AWS_REGION":  "us-east-1",
			},
			wantErr: true,
		},
		{
			name: "invalid environment",
			envVars: map[string]string{
				"ENVIRONMENT": "invalid",
				"SERVER_PORT": "8080",
				"AWS_REGION":  "us-east-1",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			for _, key := range envVars {
				os.Unsetenv(key)
			}

			// Set test environment
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			cfg, err := Load()
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && cfg == nil {
				t.Error("Load() returned nil config without error")
			}
		})
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid development config",
			config: &Config{
				Environment: "development",
				Server: ServerConfig{
					Port: "8080",
					Host: "0.0.0.0",
				},
				AWS: AWSConfig{
					Region: "us-east-1",
				},
				WebSocket: WebSocketConfig{
					Timeout:    30 * time.Second,
					BufferSize: 8192,
				},
				Session: SessionConfig{
					Timeout: 30 * time.Minute,
				},
			},
			wantErr: false,
		},
		{
			name: "invalid environment",
			config: &Config{
				Environment: "invalid",
				Server: ServerConfig{
					Port: "8080",
				},
				AWS: AWSConfig{
					Region: "us-east-1",
				},
				WebSocket: WebSocketConfig{
					Timeout:    30 * time.Second,
					BufferSize: 8192,
				},
				Session: SessionConfig{
					Timeout: 30 * time.Minute,
				},
			},
			wantErr: true,
		},
		{
			name: "missing server port",
			config: &Config{
				Environment: "development",
				Server: ServerConfig{
					Port: "",
				},
				AWS: AWSConfig{
					Region: "us-east-1",
				},
				WebSocket: WebSocketConfig{
					Timeout:    30 * time.Second,
					BufferSize: 8192,
				},
				Session: SessionConfig{
					Timeout: 30 * time.Minute,
				},
			},
			wantErr: true,
		},
		{
			name: "missing AWS region",
			config: &Config{
				Environment: "development",
				Server: ServerConfig{
					Port: "8080",
				},
				AWS: AWSConfig{
					Region: "",
				},
				WebSocket: WebSocketConfig{
					Timeout:    30 * time.Second,
					BufferSize: 8192,
				},
				Session: SessionConfig{
					Timeout: 30 * time.Minute,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid websocket timeout",
			config: &Config{
				Environment: "development",
				Server: ServerConfig{
					Port: "8080",
				},
				AWS: AWSConfig{
					Region: "us-east-1",
				},
				WebSocket: WebSocketConfig{
					Timeout:    0,
					BufferSize: 8192,
				},
				Session: SessionConfig{
					Timeout: 30 * time.Minute,
				},
			},
			wantErr: true,
		},
		{
			name: "production without bedrock config",
			config: &Config{
				Environment: "production",
				Server: ServerConfig{
					Port: "8080",
				},
				AWS: AWSConfig{
					Region: "us-east-1",
				},
				Bedrock: BedrockConfig{
					AgentID:      "",
					AgentAliasID: "",
				},
				WebSocket: WebSocketConfig{
					Timeout:    30 * time.Second,
					BufferSize: 8192,
				},
				Session: SessionConfig{
					Timeout: 30 * time.Minute,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfig_EnvironmentChecks(t *testing.T) {
	tests := []struct {
		name        string
		environment string
		isDev       bool
		isProd      bool
		isTest      bool
	}{
		{
			name:        "development",
			environment: "development",
			isDev:       true,
			isProd:      false,
			isTest:      false,
		},
		{
			name:        "production",
			environment: "production",
			isDev:       false,
			isProd:      true,
			isTest:      false,
		},
		{
			name:        "test",
			environment: "test",
			isDev:       false,
			isProd:      false,
			isTest:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{Environment: tt.environment}

			if got := cfg.IsDevelopment(); got != tt.isDev {
				t.Errorf("IsDevelopment() = %v, want %v", got, tt.isDev)
			}
			if got := cfg.IsProduction(); got != tt.isProd {
				t.Errorf("IsProduction() = %v, want %v", got, tt.isProd)
			}
			if got := cfg.IsTest(); got != tt.isTest {
				t.Errorf("IsTest() = %v, want %v", got, tt.isTest)
			}
		})
	}
}

func TestGetEnvAsDuration(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		value        string
		defaultValue time.Duration
		want         time.Duration
	}{
		{
			name:         "valid duration",
			key:          "TEST_DURATION",
			value:        "30s",
			defaultValue: 10 * time.Second,
			want:         30 * time.Second,
		},
		{
			name:         "empty value uses default",
			key:          "TEST_DURATION",
			value:        "",
			defaultValue: 10 * time.Second,
			want:         10 * time.Second,
		},
		{
			name:         "invalid value uses default",
			key:          "TEST_DURATION",
			value:        "invalid",
			defaultValue: 10 * time.Second,
			want:         10 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value != "" {
				os.Setenv(tt.key, tt.value)
				defer os.Unsetenv(tt.key)
			}

			got := getEnvAsDuration(tt.key, tt.defaultValue)
			if got != tt.want {
				t.Errorf("getEnvAsDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetEnvAsInt(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		value        string
		defaultValue int
		want         int
	}{
		{
			name:         "valid integer",
			key:          "TEST_INT",
			value:        "42",
			defaultValue: 10,
			want:         42,
		},
		{
			name:         "empty value uses default",
			key:          "TEST_INT",
			value:        "",
			defaultValue: 10,
			want:         10,
		},
		{
			name:         "invalid value uses default",
			key:          "TEST_INT",
			value:        "invalid",
			defaultValue: 10,
			want:         10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value != "" {
				os.Setenv(tt.key, tt.value)
				defer os.Unsetenv(tt.key)
			}

			got := getEnvAsInt(tt.key, tt.defaultValue)
			if got != tt.want {
				t.Errorf("getEnvAsInt() = %v, want %v", got, tt.want)
			}
		})
	}
}
