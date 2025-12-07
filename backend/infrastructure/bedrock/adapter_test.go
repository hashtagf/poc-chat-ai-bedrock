package bedrock

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/bedrock-chat-poc/backend/domain/services"
)

func TestValidateInput(t *testing.T) {
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
			name: "valid input",
			input: services.AgentInput{
				SessionID: "session-123",
				Message:   "Hello, world!",
			},
			wantErr: false,
		},
		{
			name: "empty session ID",
			input: services.AgentInput{
				SessionID: "",
				Message:   "Hello, world!",
			},
			wantErr: true,
		},
		{
			name: "empty message",
			input: services.AgentInput{
				SessionID: "session-123",
				Message:   "",
			},
			wantErr: true,
		},
		{
			name: "message too long",
			input: services.AgentInput{
				SessionID: "session-123",
				Message:   string(make([]byte, 26000)),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := adapter.validateInput(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateInput() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCalculateBackoff(t *testing.T) {
	adapter := &Adapter{
		config: AdapterConfig{
			InitialBackoff: 1 * time.Second,
			MaxBackoff:     30 * time.Second,
		},
	}

	tests := []struct {
		name    string
		attempt int
		want    time.Duration
	}{
		{
			name:    "first retry",
			attempt: 1,
			want:    1 * time.Second,
		},
		{
			name:    "second retry",
			attempt: 2,
			want:    2 * time.Second,
		},
		{
			name:    "third retry",
			attempt: 3,
			want:    4 * time.Second,
		},
		{
			name:    "fourth retry",
			attempt: 4,
			want:    8 * time.Second,
		},
		{
			name:    "max backoff",
			attempt: 10,
			want:    30 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := adapter.calculateBackoff(tt.attempt)
			if got != tt.want {
				t.Errorf("calculateBackoff() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsRetryable(t *testing.T) {
	adapter := &Adapter{}

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
		{
			name: "context deadline exceeded",
			err:  context.DeadlineExceeded,
			want: false,
		},
		{
			name: "context canceled",
			err:  context.Canceled,
			want: false,
		},
		{
			name: "generic error",
			err:  errors.New("generic error"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := adapter.isRetryable(tt.err)
			if got != tt.want {
				t.Errorf("isRetryable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTransformError(t *testing.T) {
	adapter := &Adapter{}

	tests := []struct {
		name         string
		err          error
		wantCode     string
		wantRetryable bool
	}{
		{
			name:         "nil error",
			err:          nil,
			wantCode:     "",
			wantRetryable: false,
		},
		{
			name:         "context deadline exceeded",
			err:          context.DeadlineExceeded,
			wantCode:     services.ErrCodeTimeout,
			wantRetryable: true,
		},
		{
			name:         "context canceled",
			err:          context.Canceled,
			wantCode:     services.ErrCodeNetworkError,
			wantRetryable: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := adapter.transformError(tt.err, "test-request-id")
			
			if tt.err == nil {
				if got != nil {
					t.Errorf("transformError() = %v, want nil", got)
				}
				return
			}

			var domainErr *services.DomainError
			if !errors.As(got, &domainErr) {
				t.Errorf("transformError() did not return DomainError")
				return
			}

			if domainErr.Code != tt.wantCode {
				t.Errorf("transformError() code = %v, want %v", domainErr.Code, tt.wantCode)
			}

			if domainErr.Retryable != tt.wantRetryable {
				t.Errorf("transformError() retryable = %v, want %v", domainErr.Retryable, tt.wantRetryable)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.MaxRetries != 3 {
		t.Errorf("DefaultConfig() MaxRetries = %v, want 3", cfg.MaxRetries)
	}

	if cfg.InitialBackoff != 1*time.Second {
		t.Errorf("DefaultConfig() InitialBackoff = %v, want 1s", cfg.InitialBackoff)
	}

	if cfg.MaxBackoff != 30*time.Second {
		t.Errorf("DefaultConfig() MaxBackoff = %v, want 30s", cfg.MaxBackoff)
	}

	if cfg.RequestTimeout != 60*time.Second {
		t.Errorf("DefaultConfig() RequestTimeout = %v, want 60s", cfg.RequestTimeout)
	}
}

func TestNewAdapter_Validation(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		agentID string
		aliasID string
		wantErr bool
	}{
		{
			name:    "empty agent ID",
			agentID: "",
			aliasID: "test-alias",
			wantErr: true,
		},
		{
			name:    "empty alias ID",
			agentID: "test-agent",
			aliasID: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewAdapter(ctx, tt.agentID, tt.aliasID, DefaultConfig())
			if (err != nil) != tt.wantErr {
				t.Errorf("NewAdapter() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
