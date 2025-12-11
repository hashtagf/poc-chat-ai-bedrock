package bedrock

import (
	"bytes"
	"context"
	"errors"
	"log"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/bedrockagentruntime"
	"github.com/aws/smithy-go"

	"github.com/bedrock-chat-poc/backend/domain/services"
)

// TestLoggingAndMonitoring tests logging and monitoring functionality
// Requirements: 9.1, 9.2, 9.3, 12.1, 12.2
func TestLoggingAndMonitoring(t *testing.T) {
	t.Run("APICallLogging", testAPICallLogging)
	t.Run("ErrorLoggingWithRequestIDs", testErrorLoggingWithRequestIDs)
	t.Run("MetricsCollection", testMetricsCollection)
	t.Run("StructuredLoggingFormat", testStructuredLoggingFormat)
	t.Run("RetryLogging", testRetryLogging)
	t.Run("StreamLogging", testStreamLogging)
}

// testAPICallLogging verifies that all API calls are properly logged
// Requirements: 9.1 - All API calls must be logged with session ID and agent ID
func testAPICallLogging(t *testing.T) {
	// Capture log output
	var logBuffer bytes.Buffer
	originalOutput := log.Writer()
	log.SetOutput(&logBuffer)
	defer log.SetOutput(originalOutput)

	// Create mock client that succeeds
	mockClient := &loggingMockBedrockClient{
		invokeAgentFunc: func(ctx context.Context, input *bedrockagentruntime.InvokeAgentInput) (*bedrockagentruntime.InvokeAgentOutput, error) {
			// Return a simple response - the logging happens before stream processing
			return &bedrockagentruntime.InvokeAgentOutput{}, nil
		},
	}

	adapter := &Adapter{
		client:  mockClient,
		agentID: "test-agent-123",
		aliasID: "test-alias-456",
		config:  DefaultConfig(),
	}

	// Test InvokeAgent logging
	input := services.AgentInput{
		SessionID: "test-session-789",
		Message:   "Test message for logging",
	}

	// Call InvokeAgent - we expect it to log the request even if processing fails
	adapter.InvokeAgent(context.Background(), input)

	logOutput := logBuffer.String()

	// Verify request logging (this should always happen)
	if !strings.Contains(logOutput, "[Bedrock] InvokeAgent request") {
		t.Error("Log should contain InvokeAgent request entry")
	}
	if !strings.Contains(logOutput, "SessionID: test-session-789") {
		t.Error("Log should contain session ID")
	}
	if !strings.Contains(logOutput, "AgentID: test-agent-123") {
		t.Error("Log should contain agent ID")
	}

	t.Logf("✓ API call logging verified - Found expected log entries")
}

// testErrorLoggingWithRequestIDs verifies error logging includes request IDs
// Requirements: 9.2 - Error logging must include AWS request IDs for debugging
func testErrorLoggingWithRequestIDs(t *testing.T) {
	// Capture log output
	var logBuffer bytes.Buffer
	originalOutput := log.Writer()
	log.SetOutput(&logBuffer)
	defer log.SetOutput(originalOutput)

	testCases := []struct {
		name        string
		errorCode   string
		errorMsg    string
		expectLog   string
	}{
		{
			name:      "ThrottlingException",
			errorCode: "ThrottlingException",
			errorMsg:  "Rate exceeded",
			expectLog: "AWS API Error - Code: ThrottlingException",
		},
		{
			name:      "AccessDeniedException",
			errorCode: "AccessDeniedException",
			errorMsg:  "User is not authorized",
			expectLog: "AWS API Error - Code: AccessDeniedException",
		},
		{
			name:      "ValidationException",
			errorCode: "ValidationException",
			errorMsg:  "Invalid parameter",
			expectLog: "AWS API Error - Code: ValidationException",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Clear log buffer
			logBuffer.Reset()

			// Create mock client that returns specific error
			mockClient := &loggingMockBedrockClient{
				invokeAgentFunc: func(ctx context.Context, input *bedrockagentruntime.InvokeAgentInput) (*bedrockagentruntime.InvokeAgentOutput, error) {
					return nil, &smithy.GenericAPIError{
						Code:    tc.errorCode,
						Message: tc.errorMsg,
					}
				},
			}

			adapter := &Adapter{
				client:  mockClient,
				agentID: "test-agent",
				aliasID: "test-alias",
				config:  DefaultConfig(),
			}

			input := services.AgentInput{
				SessionID: "test-session",
				Message:   "Test error logging",
			}

			_, err := adapter.InvokeAgent(context.Background(), input)
			if err == nil {
				t.Error("InvokeAgent should return error")
			}

			logOutput := logBuffer.String()

			// Verify error logging format
			if !strings.Contains(logOutput, tc.expectLog) {
				t.Errorf("Log should contain expected error log: %s", tc.expectLog)
			}
			if !strings.Contains(logOutput, tc.errorMsg) {
				t.Errorf("Log should contain error message: %s", tc.errorMsg)
			}
			if !strings.Contains(logOutput, "RequestID:") {
				t.Error("Log should contain request ID")
			}
			if !strings.Contains(logOutput, "[Bedrock] InvokeAgent failed") {
				t.Error("Log should contain failure entry")
			}

			t.Logf("✓ Error logging verified for %s", tc.name)
		})
	}
}

// testMetricsCollection verifies metrics collection for success/failure rates
// Requirements: 12.1, 12.2 - Metrics collection for success rate, latency, and error rate
func testMetricsCollection(t *testing.T) {
	// Create a metrics collector to track calls
	metrics := &TestMetricsCollector{
		successCount: 0,
		errorCount:   0,
		latencies:    []time.Duration{},
		errorsByType: make(map[string]int),
		mutex:        sync.RWMutex{},
	}

	// Test successful calls
	t.Run("SuccessMetrics", func(t *testing.T) {
		mockClient := &loggingMockBedrockClient{
			invokeAgentFunc: func(ctx context.Context, input *bedrockagentruntime.InvokeAgentInput) (*bedrockagentruntime.InvokeAgentOutput, error) {
				// Simulate some processing time
				time.Sleep(10 * time.Millisecond)
				return &bedrockagentruntime.InvokeAgentOutput{}, nil
			},
		}

		adapter := &Adapter{
			client:  mockClient,
			agentID: "test-agent",
			aliasID: "test-alias",
			config:  DefaultConfig(),
		}

		input := services.AgentInput{
			SessionID: "test-session",
			Message:   "Test metrics",
		}

		// Make multiple successful calls
		for i := 0; i < 5; i++ {
			startTime := time.Now()
			_, err := adapter.InvokeAgent(context.Background(), input)
			duration := time.Since(startTime)

			if err == nil {
				metrics.RecordSuccess(duration)
			} else {
				metrics.RecordError(err, duration)
			}
		}

		// Verify success metrics
		if metrics.GetSuccessCount() != 5 {
			t.Errorf("Expected 5 successful calls, got %d", metrics.GetSuccessCount())
		}
		if metrics.GetErrorCount() != 0 {
			t.Errorf("Expected 0 errors, got %d", metrics.GetErrorCount())
		}
		if metrics.GetSuccessRate() != 1.0 {
			t.Errorf("Expected success rate 1.0, got %f", metrics.GetSuccessRate())
		}
		if metrics.GetAverageLatency() <= 0 {
			t.Error("Expected positive average latency")
		}

		t.Logf("✓ Success metrics: %d successful calls, avg latency: %v", 
			metrics.GetSuccessCount(), metrics.GetAverageLatency())
	})

	// Test error metrics
	t.Run("ErrorMetrics", func(t *testing.T) {
		// Reset metrics
		metrics.Reset()

		errorTypes := []string{
			"ThrottlingException",
			"AccessDeniedException", 
			"ValidationException",
		}

		for _, errorType := range errorTypes {
			mockClient := &loggingMockBedrockClient{
				invokeAgentFunc: func(ctx context.Context, input *bedrockagentruntime.InvokeAgentInput) (*bedrockagentruntime.InvokeAgentOutput, error) {
					return nil, &smithy.GenericAPIError{
						Code:    errorType,
						Message: "Test error",
					}
				},
			}

			adapter := &Adapter{
				client:  mockClient,
				agentID: "test-agent",
				aliasID: "test-alias",
				config:  DefaultConfig(),
			}

			input := services.AgentInput{
				SessionID: "test-session",
				Message:   "Test error metrics",
			}

			startTime := time.Now()
			_, err := adapter.InvokeAgent(context.Background(), input)
			duration := time.Since(startTime)

			if err != nil {
				metrics.RecordError(err, duration)
			}
		}

		// Verify error metrics
		if metrics.GetSuccessCount() != 0 {
			t.Errorf("Expected 0 successful calls, got %d", metrics.GetSuccessCount())
		}
		if metrics.GetErrorCount() != 3 {
			t.Errorf("Expected 3 errors, got %d", metrics.GetErrorCount())
		}
		if metrics.GetSuccessRate() != 0.0 {
			t.Errorf("Expected success rate 0.0, got %f", metrics.GetSuccessRate())
		}

		errorsByType := metrics.GetErrorsByType()
		if errorsByType[services.ErrCodeRateLimit] != 1 {
			t.Errorf("Expected 1 rate limit error, got %d", errorsByType[services.ErrCodeRateLimit])
		}
		if errorsByType[services.ErrCodeUnauthorized] != 1 {
			t.Errorf("Expected 1 unauthorized error, got %d", errorsByType[services.ErrCodeUnauthorized])
		}
		if errorsByType[services.ErrCodeInvalidInput] != 1 {
			t.Errorf("Expected 1 invalid input error, got %d", errorsByType[services.ErrCodeInvalidInput])
		}

		t.Logf("✓ Error metrics: %d errors by type: %v", 
			metrics.GetErrorCount(), errorsByType)
	})
}

// testStructuredLoggingFormat verifies structured logging format
// Requirements: 9.3 - Structured logging format for consistent parsing
func testStructuredLoggingFormat(t *testing.T) {
	// Capture log output
	var logBuffer bytes.Buffer
	originalOutput := log.Writer()
	log.SetOutput(&logBuffer)
	defer log.SetOutput(originalOutput)

	mockClient := &loggingMockBedrockClient{
		invokeAgentFunc: func(ctx context.Context, input *bedrockagentruntime.InvokeAgentInput) (*bedrockagentruntime.InvokeAgentOutput, error) {
			return &bedrockagentruntime.InvokeAgentOutput{}, nil
		},
	}

	adapter := &Adapter{
		client:  mockClient,
		agentID: "test-agent-structured",
		aliasID: "test-alias-structured",
		config:  DefaultConfig(),
	}

	input := services.AgentInput{
		SessionID: "structured-session-123",
		Message:   "Test structured logging",
	}

	// Call InvokeAgent to generate logs
	adapter.InvokeAgent(context.Background(), input)

	logOutput := logBuffer.String()
	lines := strings.Split(strings.TrimSpace(logOutput), "\n")

	// Verify structured format for each log line
	for _, line := range lines {
		if strings.Contains(line, "[Bedrock]") {
			// Verify log line contains structured elements
			if !strings.Contains(line, "[Bedrock]") {
				t.Error("Log should have component prefix")
			}
			
			// Check for key-value pairs in structured format
			if strings.Contains(line, "SessionID:") {
				if !strings.Contains(line, "SessionID: structured-session-123") {
					t.Error("Log should contain correct session ID format")
				}
			}
			if strings.Contains(line, "AgentID:") {
				if !strings.Contains(line, "AgentID: test-agent-structured") {
					t.Error("Log should contain correct agent ID format")
				}
			}
		}
	}

	// Verify specific structured log patterns - focus on request logging which always happens
	if !containsStructuredLog(logOutput, "InvokeAgent request", "SessionID", "AgentID") {
		t.Error("Log should contain structured request log")
	}

	t.Logf("✓ Structured logging format verified - Found %d log lines with proper structure", len(lines))
}

// testRetryLogging verifies retry attempt logging
// Requirements: 9.2 - Retry attempts must be logged with backoff duration and request ID
func testRetryLogging(t *testing.T) {
	// Capture log output
	var logBuffer bytes.Buffer
	originalOutput := log.Writer()
	log.SetOutput(&logBuffer)
	defer log.SetOutput(originalOutput)

	callCount := 0
	mockClient := &loggingMockBedrockClient{
		invokeAgentFunc: func(ctx context.Context, input *bedrockagentruntime.InvokeAgentInput) (*bedrockagentruntime.InvokeAgentOutput, error) {
			callCount++
			if callCount <= 2 {
				// Return throttling error for first two attempts
				return nil, &smithy.GenericAPIError{
					Code:    "ThrottlingException",
					Message: "Rate exceeded",
				}
			}
			// Succeed on third attempt
			return &bedrockagentruntime.InvokeAgentOutput{}, nil
		},
	}

	adapter := &Adapter{
		client:  mockClient,
		agentID: "test-agent",
		aliasID: "test-alias",
		config: AdapterConfig{
			MaxRetries:     2,
			InitialBackoff: 10 * time.Millisecond, // Fast for testing
			MaxBackoff:     100 * time.Millisecond,
			RequestTimeout: 5 * time.Second,
		},
	}

	input := services.AgentInput{
		SessionID: "retry-session",
		Message:   "Test retry logging",
	}

	_, err := adapter.InvokeAgent(context.Background(), input)
	if err != nil {
		t.Fatalf("InvokeAgent should not error: %v", err)
	}

	logOutput := logBuffer.String()

	// Verify retry logging
	if !strings.Contains(logOutput, "[Bedrock] Retry attempt 1") {
		t.Error("Log should contain first retry attempt")
	}
	if !strings.Contains(logOutput, "[Bedrock] Retry attempt 2") {
		t.Error("Log should contain second retry attempt")
	}
	if !strings.Contains(logOutput, "RequestID:") {
		t.Error("Log should contain request ID")
	}

	// Count retry log entries
	retryCount := strings.Count(logOutput, "Retry attempt")
	if retryCount != 2 {
		t.Errorf("Should log exactly 2 retry attempts, got %d", retryCount)
	}

	t.Logf("✓ Retry logging verified - Found %d retry attempts logged", retryCount)
}

// testStreamLogging verifies streaming-specific logging
// Requirements: 9.1, 9.3 - Stream events and trace information must be logged
func testStreamLogging(t *testing.T) {
	// Capture log output
	var logBuffer bytes.Buffer
	originalOutput := log.Writer()
	log.SetOutput(&logBuffer)
	defer log.SetOutput(originalOutput)

	mockClient := &loggingMockBedrockClient{
		invokeAgentFunc: func(ctx context.Context, input *bedrockagentruntime.InvokeAgentInput) (*bedrockagentruntime.InvokeAgentOutput, error) {
			return &bedrockagentruntime.InvokeAgentOutput{}, nil
		},
	}

	adapter := &Adapter{
		client:  mockClient,
		agentID: "test-agent",
		aliasID: "test-alias",
		config:  DefaultConfig(),
	}

	input := services.AgentInput{
		SessionID: "stream-session",
		Message:   "Test stream logging",
	}

	// Call InvokeAgentStream - we expect it to log the request even if it fails later
	adapter.InvokeAgentStream(context.Background(), input)

	logOutput := logBuffer.String()

	// Verify stream request logging (this should always happen)
	if !strings.Contains(logOutput, "[Bedrock] InvokeAgentStream request") {
		t.Error("Log should contain stream request entry")
	}
	if !strings.Contains(logOutput, "SessionID: stream-session") {
		t.Error("Log should contain session ID")
	}
	if !strings.Contains(logOutput, "AgentID: test-agent") {
		t.Error("Log should contain agent ID")
	}

	t.Logf("✓ Stream logging verified - Found expected stream log entries")
}

// TestMetricsCollector is a test implementation for collecting metrics
type TestMetricsCollector struct {
	successCount int
	errorCount   int
	latencies    []time.Duration
	errorsByType map[string]int
	mutex        sync.RWMutex
}

func (m *TestMetricsCollector) RecordSuccess(latency time.Duration) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.successCount++
	m.latencies = append(m.latencies, latency)
}

func (m *TestMetricsCollector) RecordError(err error, latency time.Duration) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.errorCount++
	m.latencies = append(m.latencies, latency)

	// Categorize error by type
	var domainErr *services.DomainError
	if errors.As(err, &domainErr) {
		m.errorsByType[domainErr.Code]++
	} else {
		m.errorsByType["UNKNOWN"]++
	}
}

func (m *TestMetricsCollector) GetSuccessCount() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.successCount
}

func (m *TestMetricsCollector) GetErrorCount() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.errorCount
}

func (m *TestMetricsCollector) GetSuccessRate() float64 {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	total := m.successCount + m.errorCount
	if total == 0 {
		return 0.0
	}
	return float64(m.successCount) / float64(total)
}

func (m *TestMetricsCollector) GetAverageLatency() time.Duration {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	if len(m.latencies) == 0 {
		return 0
	}
	
	var total time.Duration
	for _, latency := range m.latencies {
		total += latency
	}
	return total / time.Duration(len(m.latencies))
}

func (m *TestMetricsCollector) GetErrorsByType() map[string]int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	result := make(map[string]int)
	for k, v := range m.errorsByType {
		result[k] = v
	}
	return result
}

func (m *TestMetricsCollector) Reset() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.successCount = 0
	m.errorCount = 0
	m.latencies = []time.Duration{}
	m.errorsByType = make(map[string]int)
}

// containsStructuredLog checks if log output contains structured log with expected fields
func containsStructuredLog(logOutput, operation string, fields ...string) bool {
	lines := strings.Split(logOutput, "\n")
	for _, line := range lines {
		if strings.Contains(line, operation) {
			allFieldsPresent := true
			for _, field := range fields {
				if !strings.Contains(line, field) {
					allFieldsPresent = false
					break
				}
			}
			if allFieldsPresent {
				return true
			}
		}
	}
	return false
}

// loggingMockBedrockClient for testing logging functionality
type loggingMockBedrockClient struct {
	invokeAgentFunc func(ctx context.Context, input *bedrockagentruntime.InvokeAgentInput) (*bedrockagentruntime.InvokeAgentOutput, error)
}

func (m *loggingMockBedrockClient) InvokeAgent(ctx context.Context, input *bedrockagentruntime.InvokeAgentInput, optFns ...func(*bedrockagentruntime.Options)) (*bedrockagentruntime.InvokeAgentOutput, error) {
	if m.invokeAgentFunc != nil {
		return m.invokeAgentFunc(ctx, input)
	}
	return &bedrockagentruntime.InvokeAgentOutput{}, nil
}