package bedrock

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/bedrockagentruntime"
	"github.com/aws/smithy-go"
	"github.com/bedrock-chat-poc/backend/domain/services"
)

// mockBedrockClient is a mock implementation of the Bedrock client for testing
type mockBedrockClient struct {
	invokeAgentFunc func(ctx context.Context, input *bedrockagentruntime.InvokeAgentInput) (*bedrockagentruntime.InvokeAgentOutput, error)
	callCount       int
}

func (m *mockBedrockClient) InvokeAgent(ctx context.Context, input *bedrockagentruntime.InvokeAgentInput, optFns ...func(*bedrockagentruntime.Options)) (*bedrockagentruntime.InvokeAgentOutput, error) {
	m.callCount++
	if m.invokeAgentFunc != nil {
		return m.invokeAgentFunc(ctx, input)
	}
	return nil, errors.New("mock not configured")
}

// TestTimeoutScenarios tests timeout handling and context cancellation
func TestTimeoutScenarios(t *testing.T) {
	tests := []struct {
		name           string
		timeout        time.Duration
		mockDelay      time.Duration
		expectTimeout  bool
		expectCancel   bool
		expectedCode   string
		expectedRetry  bool
	}{
		{
			name:          "request timeout",
			timeout:       100 * time.Millisecond,
			mockDelay:     200 * time.Millisecond,
			expectTimeout: true,
			expectedCode:  services.ErrCodeTimeout,
			expectedRetry: true,
		},
		{
			name:         "context cancellation",
			timeout:      1 * time.Second,
			mockDelay:    50 * time.Millisecond,
			expectCancel: true,
			expectedCode: services.ErrCodeNetworkError,
			expectedRetry: false,
		},
		{
			name:          "successful within timeout",
			timeout:       200 * time.Millisecond,
			mockDelay:     50 * time.Millisecond,
			expectTimeout: false,
			expectCancel:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockBedrockClient{
				invokeAgentFunc: func(ctx context.Context, input *bedrockagentruntime.InvokeAgentInput) (*bedrockagentruntime.InvokeAgentOutput, error) {
					// Simulate delay
					select {
					case <-time.After(tt.mockDelay):
						if tt.expectTimeout {
							return nil, context.DeadlineExceeded
						}
						// Return successful response
						return &bedrockagentruntime.InvokeAgentOutput{}, nil
					case <-ctx.Done():
						return nil, ctx.Err()
					}
				},
			}

			adapter := &Adapter{
				client:  mockClient,
				agentID: "test-agent",
				aliasID: "test-alias",
				config: AdapterConfig{
					MaxRetries:     0, // No retries for timeout tests
					InitialBackoff: 1 * time.Millisecond,
					MaxBackoff:     10 * time.Millisecond,
					RequestTimeout: tt.timeout,
				},
			}

			input := services.AgentInput{
				SessionID: "test-session",
				Message:   "Test message",
			}

			ctx := context.Background()
			if tt.expectCancel {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				// Cancel after a short delay
				go func() {
					time.Sleep(25 * time.Millisecond)
					cancel()
				}()
			}

			_, err := adapter.InvokeAgent(ctx, input)

			if !tt.expectTimeout && !tt.expectCancel {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
				return
			}

			if err == nil {
				t.Errorf("Expected error, got nil")
				return
			}

			var domainErr *services.DomainError
			if !errors.As(err, &domainErr) {
				t.Errorf("Expected DomainError, got: %T", err)
				return
			}

			if domainErr.Code != tt.expectedCode {
				t.Errorf("Expected error code %s, got %s", tt.expectedCode, domainErr.Code)
			}

			if domainErr.Retryable != tt.expectedRetry {
				t.Errorf("Expected retryable %v, got %v", tt.expectedRetry, domainErr.Retryable)
			}
		})
	}
}

// TestRateLimitingAndExponentialBackoff tests rate limiting scenarios and exponential backoff
func TestRateLimitingAndExponentialBackoff(t *testing.T) {
	tests := []struct {
		name              string
		errorCode         string
		maxRetries        int
		expectedCallCount int
		expectedRetryable bool
		expectedCode      string
	}{
		{
			name:              "throttling exception with retries",
			errorCode:         "ThrottlingException",
			maxRetries:        3,
			expectedCallCount: 4, // Initial + 3 retries
			expectedRetryable: true,
			expectedCode:      services.ErrCodeRateLimit,
		},
		{
			name:              "too many requests with retries",
			errorCode:         "TooManyRequestsException",
			maxRetries:        2,
			expectedCallCount: 3, // Initial + 2 retries
			expectedRetryable: true,
			expectedCode:      services.ErrCodeRateLimit,
		},
		{
			name:              "service unavailable with retries",
			errorCode:         "ServiceUnavailableException",
			maxRetries:        1,
			expectedCallCount: 2, // Initial + 1 retry
			expectedRetryable: true,
			expectedCode:      services.ErrCodeServiceError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockBedrockClient{
				invokeAgentFunc: func(ctx context.Context, input *bedrockagentruntime.InvokeAgentInput) (*bedrockagentruntime.InvokeAgentOutput, error) {
					return nil, &smithy.GenericAPIError{
						Code:    tt.errorCode,
						Message: fmt.Sprintf("Mock %s error", tt.errorCode),
					}
				},
			}

			adapter := &Adapter{
				client:  mockClient,
				agentID: "test-agent",
				aliasID: "test-alias",
				config: AdapterConfig{
					MaxRetries:     tt.maxRetries,
					InitialBackoff: 1 * time.Millisecond, // Fast for testing
					MaxBackoff:     10 * time.Millisecond,
					RequestTimeout: 5 * time.Second,
				},
			}

			input := services.AgentInput{
				SessionID: "test-session",
				Message:   "Test message",
			}

			start := time.Now()
			_, err := adapter.InvokeAgent(context.Background(), input)
			duration := time.Since(start)

			// Verify error occurred
			if err == nil {
				t.Errorf("Expected error, got nil")
				return
			}

			// Verify correct number of calls
			if mockClient.callCount != tt.expectedCallCount {
				t.Errorf("Expected %d calls, got %d", tt.expectedCallCount, mockClient.callCount)
			}

			// Verify error transformation
			var domainErr *services.DomainError
			if !errors.As(err, &domainErr) {
				t.Errorf("Expected DomainError, got: %T", err)
				return
			}

			if domainErr.Code != tt.expectedCode {
				t.Errorf("Expected error code %s, got %s", tt.expectedCode, domainErr.Code)
			}

			if domainErr.Retryable != tt.expectedRetryable {
				t.Errorf("Expected retryable %v, got %v", tt.expectedRetryable, domainErr.Retryable)
			}

			// Verify exponential backoff was applied (should take some time for retries)
			if tt.maxRetries > 0 {
				expectedMinDuration := time.Duration(tt.maxRetries) * time.Millisecond
				if duration < expectedMinDuration {
					t.Errorf("Expected duration >= %v for backoff, got %v", expectedMinDuration, duration)
				}
			}
		})
	}
}

// TestAccessDeniedErrorTransformation tests access denied error handling
func TestAccessDeniedErrorTransformation(t *testing.T) {
	tests := []struct {
		name              string
		errorCode         string
		expectedCode      string
		expectedRetryable bool
		expectedMessage   string
	}{
		{
			name:              "access denied exception",
			errorCode:         "AccessDeniedException",
			expectedCode:      services.ErrCodeUnauthorized,
			expectedRetryable: false,
			expectedMessage:   "Unauthorized access to Bedrock service",
		},
		{
			name:              "unauthorized exception",
			errorCode:         "UnauthorizedException",
			expectedCode:      services.ErrCodeUnauthorized,
			expectedRetryable: false,
			expectedMessage:   "Unauthorized access to Bedrock service",
		},
		{
			name:              "validation exception",
			errorCode:         "ValidationException",
			expectedCode:      services.ErrCodeInvalidInput,
			expectedRetryable: false,
			expectedMessage:   "Invalid input parameters",
		},
		{
			name:              "invalid parameter exception",
			errorCode:         "InvalidParameterException",
			expectedCode:      services.ErrCodeInvalidInput,
			expectedRetryable: false,
			expectedMessage:   "Invalid input parameters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockBedrockClient{
				invokeAgentFunc: func(ctx context.Context, input *bedrockagentruntime.InvokeAgentInput) (*bedrockagentruntime.InvokeAgentOutput, error) {
					return nil, &smithy.GenericAPIError{
						Code:    tt.errorCode,
						Message: fmt.Sprintf("Mock %s error", tt.errorCode),
					}
				},
			}

			adapter := &Adapter{
				client:  mockClient,
				agentID: "test-agent",
				aliasID: "test-alias",
				config: AdapterConfig{
					MaxRetries:     2, // Should not retry non-retryable errors
					InitialBackoff: 1 * time.Millisecond,
					MaxBackoff:     10 * time.Millisecond,
					RequestTimeout: 5 * time.Second,
				},
			}

			input := services.AgentInput{
				SessionID: "test-session",
				Message:   "Test message",
			}

			_, err := adapter.InvokeAgent(context.Background(), input)

			// Verify error occurred
			if err == nil {
				t.Errorf("Expected error, got nil")
				return
			}

			// Verify only one call was made (no retries for non-retryable errors)
			if mockClient.callCount != 1 {
				t.Errorf("Expected 1 call (no retries), got %d", mockClient.callCount)
			}

			// Verify error transformation
			var domainErr *services.DomainError
			if !errors.As(err, &domainErr) {
				t.Errorf("Expected DomainError, got: %T", err)
				return
			}

			if domainErr.Code != tt.expectedCode {
				t.Errorf("Expected error code %s, got %s", tt.expectedCode, domainErr.Code)
			}

			if domainErr.Retryable != tt.expectedRetryable {
				t.Errorf("Expected retryable %v, got %v", tt.expectedRetryable, domainErr.Retryable)
			}

			if domainErr.Message != tt.expectedMessage {
				t.Errorf("Expected message %s, got %s", tt.expectedMessage, domainErr.Message)
			}
		})
	}
}

// TestRetryLimitsRespected tests that retry limits are properly respected
func TestRetryLimitsRespected(t *testing.T) {
	tests := []struct {
		name              string
		maxRetries        int
		expectedCallCount int
	}{
		{
			name:              "no retries",
			maxRetries:        0,
			expectedCallCount: 1,
		},
		{
			name:              "one retry",
			maxRetries:        1,
			expectedCallCount: 2,
		},
		{
			name:              "three retries",
			maxRetries:        3,
			expectedCallCount: 4,
		},
		{
			name:              "five retries",
			maxRetries:        5,
			expectedCallCount: 6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockBedrockClient{
				invokeAgentFunc: func(ctx context.Context, input *bedrockagentruntime.InvokeAgentInput) (*bedrockagentruntime.InvokeAgentOutput, error) {
					// Always return a retryable error
					return nil, &smithy.GenericAPIError{
						Code:    "ThrottlingException",
						Message: "Rate limit exceeded",
					}
				},
			}

			adapter := &Adapter{
				client:  mockClient,
				agentID: "test-agent",
				aliasID: "test-alias",
				config: AdapterConfig{
					MaxRetries:     tt.maxRetries,
					InitialBackoff: 1 * time.Millisecond, // Fast for testing
					MaxBackoff:     10 * time.Millisecond,
					RequestTimeout: 5 * time.Second,
				},
			}

			input := services.AgentInput{
				SessionID: "test-session",
				Message:   "Test message",
			}

			_, err := adapter.InvokeAgent(context.Background(), input)

			// Verify error occurred (since we always return an error)
			if err == nil {
				t.Errorf("Expected error, got nil")
			}

			// Verify exact number of calls
			if mockClient.callCount != tt.expectedCallCount {
				t.Errorf("Expected exactly %d calls, got %d", tt.expectedCallCount, mockClient.callCount)
			}
		})
	}
}

// TestExponentialBackoffCalculation tests the exponential backoff calculation
func TestExponentialBackoffCalculation(t *testing.T) {
	adapter := &Adapter{
		config: AdapterConfig{
			InitialBackoff: 100 * time.Millisecond,
			MaxBackoff:     5 * time.Second,
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
			want:    100 * time.Millisecond,
		},
		{
			name:    "second retry",
			attempt: 2,
			want:    200 * time.Millisecond,
		},
		{
			name:    "third retry",
			attempt: 3,
			want:    400 * time.Millisecond,
		},
		{
			name:    "fourth retry",
			attempt: 4,
			want:    800 * time.Millisecond,
		},
		{
			name:    "fifth retry",
			attempt: 5,
			want:    1600 * time.Millisecond,
		},
		{
			name:    "sixth retry",
			attempt: 6,
			want:    3200 * time.Millisecond,
		},
		{
			name:    "seventh retry (capped at max)",
			attempt: 7,
			want:    5 * time.Second,
		},
		{
			name:    "tenth retry (still capped)",
			attempt: 10,
			want:    5 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := adapter.calculateBackoff(tt.attempt)
			if got != tt.want {
				t.Errorf("calculateBackoff(%d) = %v, want %v", tt.attempt, got, tt.want)
			}
		})
	}
}

// TestStreamingErrorHandling tests error handling in streaming scenarios
func TestStreamingErrorHandling(t *testing.T) {
	tests := []struct {
		name              string
		errorCode         string
		expectedCode      string
		expectedRetryable bool
	}{
		{
			name:              "streaming throttling error",
			errorCode:         "ThrottlingException",
			expectedCode:      services.ErrCodeRateLimit,
			expectedRetryable: true,
		},
		{
			name:              "streaming access denied",
			errorCode:         "AccessDeniedException",
			expectedCode:      services.ErrCodeUnauthorized,
			expectedRetryable: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockBedrockClient{
				invokeAgentFunc: func(ctx context.Context, input *bedrockagentruntime.InvokeAgentInput) (*bedrockagentruntime.InvokeAgentOutput, error) {
					return nil, &smithy.GenericAPIError{
						Code:    tt.errorCode,
						Message: fmt.Sprintf("Mock %s error", tt.errorCode),
					}
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
				Message:   "Test streaming message",
			}

			_, err := adapter.InvokeAgentStream(context.Background(), input)

			// Verify error occurred
			if err == nil {
				t.Errorf("Expected error, got nil")
				return
			}

			// Verify error transformation
			var domainErr *services.DomainError
			if !errors.As(err, &domainErr) {
				t.Errorf("Expected DomainError, got: %T", err)
				return
			}

			if domainErr.Code != tt.expectedCode {
				t.Errorf("Expected error code %s, got %s", tt.expectedCode, domainErr.Code)
			}

			if domainErr.Retryable != tt.expectedRetryable {
				t.Errorf("Expected retryable %v, got %v", tt.expectedRetryable, domainErr.Retryable)
			}
		})
	}
}

// TestContextCancellationDuringRetry tests context cancellation during retry attempts
func TestContextCancellationDuringRetry(t *testing.T) {
	mockClient := &mockBedrockClient{
		invokeAgentFunc: func(ctx context.Context, input *bedrockagentruntime.InvokeAgentInput) (*bedrockagentruntime.InvokeAgentOutput, error) {
			// Always return a retryable error to trigger retries
			return nil, &smithy.GenericAPIError{
				Code:    "ThrottlingException",
				Message: "Rate limit exceeded",
			}
		},
	}

	adapter := &Adapter{
		client:  mockClient,
		agentID: "test-agent",
		aliasID: "test-alias",
		config: AdapterConfig{
			MaxRetries:     5,
			InitialBackoff: 100 * time.Millisecond, // Longer backoff to allow cancellation
			MaxBackoff:     1 * time.Second,
			RequestTimeout: 10 * time.Second,
		},
	}

	input := services.AgentInput{
		SessionID: "test-session",
		Message:   "Test message",
	}

	ctx, cancel := context.WithCancel(context.Background())
	
	// Cancel context after a short delay to interrupt retry logic
	go func() {
		time.Sleep(150 * time.Millisecond) // Cancel during first retry backoff
		cancel()
	}()

	_, err := adapter.InvokeAgent(ctx, input)

	// Verify error occurred
	if err == nil {
		t.Errorf("Expected error due to context cancellation, got nil")
		return
	}

	// Verify error is context cancellation
	var domainErr *services.DomainError
	if !errors.As(err, &domainErr) {
		t.Errorf("Expected DomainError, got: %T", err)
		return
	}

	if domainErr.Code != services.ErrCodeNetworkError {
		t.Errorf("Expected error code %s, got %s", services.ErrCodeNetworkError, domainErr.Code)
	}

	if domainErr.Retryable != false {
		t.Errorf("Expected retryable false for context cancellation, got %v", domainErr.Retryable)
	}

	// Verify that not all retries were attempted (should be interrupted)
	if mockClient.callCount >= 6 { // Should be less than max retries + 1
		t.Errorf("Expected fewer calls due to cancellation, got %d", mockClient.callCount)
	}
}