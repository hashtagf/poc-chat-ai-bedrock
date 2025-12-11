# Design Document: Bedrock Agent Go Integration Verification

## Overview

This design document outlines a comprehensive verification and testing strategy for the Amazon Bedrock Agent integration with the Go backend application. The verification focuses on validating the existing Bedrock adapter implementation, ensuring robust error handling, and providing confidence that the integration works reliably across different environments and failure scenarios.

The design addresses the critical need to verify that the Bedrock Agent integration works correctly before production deployment, with special attention to IAM permissions, error handling, and DevOps concerns that have caused access denied errors in the past.

### Requirements Coverage

This design addresses all requirements from the requirements document:

- **Requirement 1** (Agent Invocation Verification): Addressed by Integration Test Suite (Section 3.1)
- **Requirement 2** (Input Validation Testing): Addressed by Unit Test Suite (Section 3.2)
- **Requirement 3** (Streaming Response Testing): Addressed by Stream Testing Framework (Section 3.3)
- **Requirement 4** (Error Handling Verification): Addressed by Error Simulation Framework (Section 3.4)
- **Requirement 5** (Retry Logic Testing): Addressed by Retry Testing Suite (Section 3.5)
- **Requirement 6** (Configuration Testing): Addressed by Configuration Validation (Section 3.6)
- **Requirement 7** (AWS SDK Integration): Addressed by AWS Integration Tests (Section 3.7)
- **Requirement 8** (Citation Processing): Addressed by Citation Testing Suite (Section 3.8)
- **Requirement 9** (Logging and Debugging): Addressed by Observability Testing (Section 3.9)
- **Requirement 10** (IAM and Access Control): Addressed by Permission Verification Suite (Section 3.10)
- **Requirement 11** (Environment Configuration): Addressed by Environment Testing Framework (Section 3.11)
- **Requirement 12** (Monitoring and Observability): Addressed by Metrics and Monitoring Tests (Section 3.12)
- **Requirement 13** (Domain Interface Integration): Addressed by Interface Compliance Tests (Section 3.13)

## Architecture

### High-Level Testing Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Test Orchestration Layer                  │
│  ┌────────────────────────────────────────────────────────┐ │
│  │  Test Suite Manager                                    │ │
│  │  - Environment Setup/Teardown                          │ │
│  │  - Test Data Management                                │ │
│  │  │  - Parallel Test Execution                          │ │
│  │  - Result Aggregation                                  │ │
│  └────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                    Integration Test Layer                    │
│  ┌────────────────────────────────────────────────────────┐ │
│  │  Real AWS Bedrock Integration                          │ │
│  │  - Live Agent Invocation                               │ │
│  │  - Knowledge Base Integration                           │ │
│  │  - VPC Endpoint Testing                                │ │
│  │  - Cross-Region Testing                                │ │
│  └────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                    Unit Test Layer                          │
│  ┌────────────────────────────────────────────────────────┐ │
│  │  Adapter Logic Testing                                 │ │
│  │  - Input Validation                                    │ │
│  │  - Error Transformation                                │ │
│  │  - Retry Logic                                         │ │
│  │  - Configuration Handling                              │ │
│  └────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                    Mock and Simulation Layer                 │
│  ┌────────────────────────────────────────────────────────┐ │
│  │  AWS SDK Mocking                                      │ │
│  │  - Error Simulation                                   │ │
│  │  - Rate Limiting Simulation                           │ │
│  │  - Network Failure Simulation                         │ │
│  │  - Response Streaming Simulation                      │ │
│  └────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                    Observability and Metrics                │
│  ┌────────────────────────────────────────────────────────┐ │
│  │  Test Metrics Collection                               │ │
│  │  - Performance Metrics                                 │ │
│  │  - Error Rate Tracking                                │ │
│  │  - Success Rate Monitoring                            │ │
│  │  - Latency Distribution                               │ │
│  └────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

### Test Environment Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Development Environment                   │
│  ┌────────────────────────────────────────────────────────┐ │
│  │  Local Testing                                         │ │
│  │  - Unit Tests with Mocks                               │ │
│  │  │  - Fast Feedback Loop                               │ │
│  │  - Configuration Validation                            │ │
│  │  - Static Analysis                                     │ │
│  └────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                    Integration Environment                   │
│  ┌────────────────────────────────────────────────────────┐ │
│  │  AWS Test Account                                      │ │
│  │  - Real Bedrock Agent                                  │ │
│  │  - Test Knowledge Base                                 │ │
│  │  - IAM Role Testing                                    │ │
│  │  - VPC Endpoint Testing                                │ │
│  └────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                    Production-like Environment               │
│  ┌────────────────────────────────────────────────────────┐ │
│  │  Staging AWS Account                                   │ │
│  │  - Production Configuration                            │ │
│  │  - VPC with Private Subnets                            │ │
│  │  - Production IAM Policies                             │ │
│  │  - Monitoring and Alerting                            │ │
│  └────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

## Components and Interfaces

### 1. Integration Test Suite

**Purpose**: Validates end-to-end functionality with real AWS Bedrock services.

**Test Categories**:

**Basic Agent Invocation Tests**:
```go
func TestAgentInvocation_BasicMessage(t *testing.T) {
    // Test simple message without knowledge base
    input := services.AgentInput{
        SessionID: generateSessionID(),
        Message:   "Hello, how can you help me?",
    }
    
    response, err := adapter.InvokeAgent(ctx, input)
    assert.NoError(t, err)
    assert.NotEmpty(t, response.Content)
    assert.NotEmpty(t, response.RequestID)
}

func TestAgentInvocation_WithKnowledgeBase(t *testing.T) {
    // Test message with knowledge base integration
    input := services.AgentInput{
        SessionID:        generateSessionID(),
        Message:          "What information do you have about our products?",
        KnowledgeBaseIDs: []string{testKnowledgeBaseID},
    }
    
    response, err := adapter.InvokeAgent(ctx, input)
    assert.NoError(t, err)
    assert.NotEmpty(t, response.Content)
    // May include citations from knowledge base
}
```

**Session Context Tests**:
```go
func TestAgentInvocation_SessionContext(t *testing.T) {
    sessionID := generateSessionID()
    
    // First message
    response1, err := adapter.InvokeAgent(ctx, services.AgentInput{
        SessionID: sessionID,
        Message:   "My name is John. Remember this.",
    })
    assert.NoError(t, err)
    
    // Second message should maintain context
    response2, err := adapter.InvokeAgent(ctx, services.AgentInput{
        SessionID: sessionID,
        Message:   "What is my name?",
    })
    assert.NoError(t, err)
    assert.Contains(t, strings.ToLower(response2.Content), "john")
}
```

**Citation Processing Tests**:
```go
func TestAgentInvocation_CitationProcessing(t *testing.T) {
    input := services.AgentInput{
        SessionID:        generateSessionID(),
        Message:          "Tell me about our company policies",
        KnowledgeBaseIDs: []string{testKnowledgeBaseID},
    }
    
    response, err := adapter.InvokeAgent(ctx, input)
    assert.NoError(t, err)
    
    if len(response.Citations) > 0 {
        citation := response.Citations[0]
        assert.NotEmpty(t, citation.Excerpt)
        // Verify citation structure
        assert.NotNil(t, citation.Metadata)
    }
}
```

### 2. Unit Test Suite

**Purpose**: Tests adapter logic in isolation with mocked AWS SDK.

**Input Validation Tests**:
```go
func TestValidateInput_EmptySessionID(t *testing.T) {
    adapter := &Adapter{config: DefaultConfig()}
    
    input := services.AgentInput{
        SessionID: "",
        Message:   "Valid message",
    }
    
    err := adapter.validateInput(input)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "session ID is required")
}

func TestValidateInput_MessageTooLong(t *testing.T) {
    adapter := &Adapter{config: DefaultConfig()}
    
    input := services.AgentInput{
        SessionID: "valid-session",
        Message:   strings.Repeat("a", 25001), // Exceeds 25000 limit
    }
    
    err := adapter.validateInput(input)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "exceeds maximum length")
}
```

**Configuration Tests**:
```go
func TestNewAdapter_InvalidConfiguration(t *testing.T) {
    tests := []struct {
        name    string
        agentID string
        aliasID string
        wantErr string
    }{
        {
            name:    "empty agent ID",
            agentID: "",
            aliasID: "valid-alias",
            wantErr: "agentID is required",
        },
        {
            name:    "empty alias ID",
            agentID: "valid-agent",
            aliasID: "",
            wantErr: "aliasID is required",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            _, err := NewAdapter(ctx, tt.agentID, tt.aliasID, DefaultConfig())
            assert.Error(t, err)
            assert.Contains(t, err.Error(), tt.wantErr)
        })
    }
}
```

### 3. Stream Testing Framework

**Purpose**: Validates streaming response functionality.

**Stream Reader Tests**:
```go
func TestInvokeAgentStream_BasicStreaming(t *testing.T) {
    input := services.AgentInput{
        SessionID: generateSessionID(),
        Message:   "Tell me a story",
    }
    
    streamReader, err := adapter.InvokeAgentStream(ctx, input)
    assert.NoError(t, err)
    defer streamReader.Close()
    
    var content strings.Builder
    for {
        chunk, done, err := streamReader.Read()
        if done {
            break
        }
        assert.NoError(t, err)
        content.WriteString(chunk)
    }
    
    assert.NotEmpty(t, content.String())
}

func TestInvokeAgentStream_CitationHandling(t *testing.T) {
    input := services.AgentInput{
        SessionID:        generateSessionID(),
        Message:          "What are our company values?",
        KnowledgeBaseIDs: []string{testKnowledgeBaseID},
    }
    
    streamReader, err := adapter.InvokeAgentStream(ctx, input)
    assert.NoError(t, err)
    defer streamReader.Close()
    
    // Read content and check for citations
    for {
        chunk, done, err := streamReader.Read()
        if done {
            break
        }
        assert.NoError(t, err)
        
        // Check for citations
        citation, err := streamReader.ReadCitation()
        if err == nil && citation != nil {
            assert.NotEmpty(t, citation.Excerpt)
        }
    }
}
```

### 4. Error Simulation Framework

**Purpose**: Tests error handling with simulated AWS SDK errors.

**AWS Error Simulation**:
```go
func TestErrorHandling_ThrottlingException(t *testing.T) {
    // Mock AWS SDK to return ThrottlingException
    mockClient := &mockBedrockClient{
        invokeAgentFunc: func(ctx context.Context, input *bedrockagentruntime.InvokeAgentInput) (*bedrockagentruntime.InvokeAgentOutput, error) {
            return nil, &smithy.GenericAPIError{
                Code:    "ThrottlingException",
                Message: "Rate exceeded",
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
        Message:   "Test message",
    }
    
    _, err := adapter.InvokeAgent(ctx, input)
    
    var domainErr *services.DomainError
    assert.True(t, errors.As(err, &domainErr))
    assert.Equal(t, services.ErrCodeRateLimit, domainErr.Code)
    assert.True(t, domainErr.Retryable)
}

func TestErrorHandling_AccessDeniedException(t *testing.T) {
    mockClient := &mockBedrockClient{
        invokeAgentFunc: func(ctx context.Context, input *bedrockagentruntime.InvokeAgentInput) (*bedrockagentruntime.InvokeAgentOutput, error) {
            return nil, &smithy.GenericAPIError{
                Code:    "AccessDeniedException",
                Message: "User is not authorized to perform: bedrock:InvokeAgent",
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
        Message:   "Test message",
    }
    
    _, err := adapter.InvokeAgent(ctx, input)
    
    var domainErr *services.DomainError
    assert.True(t, errors.As(err, &domainErr))
    assert.Equal(t, services.ErrCodeUnauthorized, domainErr.Code)
    assert.False(t, domainErr.Retryable)
}
```

### 5. Retry Testing Suite

**Purpose**: Validates retry logic and exponential backoff.

**Retry Logic Tests**:
```go
func TestRetryLogic_ExponentialBackoff(t *testing.T) {
    adapter := &Adapter{
        config: AdapterConfig{
            MaxRetries:     3,
            InitialBackoff: 100 * time.Millisecond,
            MaxBackoff:     5 * time.Second,
        },
    }
    
    tests := []struct {
        attempt int
        want    time.Duration
    }{
        {1, 100 * time.Millisecond},
        {2, 200 * time.Millisecond},
        {3, 400 * time.Millisecond},
        {4, 800 * time.Millisecond},
    }
    
    for _, tt := range tests {
        got := adapter.calculateBackoff(tt.attempt)
        assert.Equal(t, tt.want, got)
    }
}

func TestRetryLogic_MaxRetriesRespected(t *testing.T) {
    callCount := 0
    mockClient := &mockBedrockClient{
        invokeAgentFunc: func(ctx context.Context, input *bedrockagentruntime.InvokeAgentInput) (*bedrockagentruntime.InvokeAgentOutput, error) {
            callCount++
            return nil, &smithy.GenericAPIError{
                Code:    "ThrottlingException",
                Message: "Rate exceeded",
            }
        },
    }
    
    adapter := &Adapter{
        client:  mockClient,
        agentID: "test-agent",
        aliasID: "test-alias",
        config: AdapterConfig{
            MaxRetries:     2,
            InitialBackoff: 1 * time.Millisecond, // Fast for testing
            MaxBackoff:     10 * time.Millisecond,
        },
    }
    
    input := services.AgentInput{
        SessionID: "test-session",
        Message:   "Test message",
    }
    
    _, err := adapter.InvokeAgent(ctx, input)
    
    assert.Error(t, err)
    assert.Equal(t, 3, callCount) // Initial call + 2 retries
}
```

### 6. Configuration Validation

**Purpose**: Tests adapter configuration and AWS SDK setup.

**AWS Configuration Tests**:
```go
func TestAWSConfiguration_RegionValidation(t *testing.T) {
    // Test with different AWS regions
    regions := []string{"us-east-1", "ap-southeast-1", "eu-west-1"}
    
    for _, region := range regions {
        t.Run(region, func(t *testing.T) {
            os.Setenv("AWS_REGION", region)
            defer os.Unsetenv("AWS_REGION")
            
            adapter, err := NewAdapter(ctx, "test-agent", "test-alias", DefaultConfig())
            if err != nil {
                t.Skipf("Skipping region %s due to configuration error: %v", region, err)
            }
            
            assert.NotNil(t, adapter)
            assert.NotNil(t, adapter.client)
        })
    }
}
```

### 7. AWS Integration Tests

**Purpose**: Tests real AWS SDK integration and VPC endpoint connectivity.

**VPC Endpoint Tests**:
```go
func TestVPCEndpoint_Connectivity(t *testing.T) {
    if !isVPCEnvironment() {
        t.Skip("Skipping VPC endpoint test - not in VPC environment")
    }
    
    // Test that requests go through VPC endpoint
    adapter, err := NewAdapter(ctx, testAgentID, testAliasID, DefaultConfig())
    assert.NoError(t, err)
    
    input := services.AgentInput{
        SessionID: generateSessionID(),
        Message:   "Test VPC connectivity",
    }
    
    response, err := adapter.InvokeAgent(ctx, input)
    assert.NoError(t, err)
    assert.NotEmpty(t, response.Content)
    
    // Verify traffic went through VPC endpoint (check logs or metrics)
}
```

### 8. Citation Testing Suite

**Purpose**: Validates citation parsing and conversion.

**Citation Conversion Tests**:
```go
func TestCitationConversion_CompleteMetadata(t *testing.T) {
    // Create mock AWS citation
    awsCitation := types.Citation{
        GeneratedResponsePart: &types.GeneratedResponsePart{
            TextResponsePart: &types.TextResponsePart{
                Text: aws.String("This is the excerpt from the document"),
            },
        },
        RetrievedReferences: []types.RetrievedReference{
            {
                Content: &types.RetrievalResultContent{
                    Text: aws.String("Source document content"),
                },
                Location: &types.RetrievalResultLocation{
                    S3Location: &types.RetrievalResultS3Location{
                        Uri: aws.String("s3://bucket/document.pdf"),
                    },
                },
                Metadata: map[string]interface{}{
                    "author": "John Doe",
                    "date":   "2024-01-01",
                },
            },
        },
    }
    
    adapter := &Adapter{}
    domainCitation := adapter.convertCitation(awsCitation)
    
    assert.Equal(t, "This is the excerpt from the document", domainCitation.Excerpt)
    assert.Equal(t, "Source document content", domainCitation.SourceName)
    assert.Equal(t, "s3://bucket/document.pdf", domainCitation.SourceID)
    assert.Equal(t, "s3://bucket/document.pdf", domainCitation.URL)
    assert.Equal(t, "John Doe", domainCitation.Metadata["author"])
    assert.Equal(t, "2024-01-01", domainCitation.Metadata["date"])
}
```

### 9. Observability Testing

**Purpose**: Validates logging, metrics, and debugging features.

**Logging Tests**:
```go
func TestLogging_RequestResponseLogging(t *testing.T) {
    var logBuffer bytes.Buffer
    log.SetOutput(&logBuffer)
    defer log.SetOutput(os.Stderr)
    
    adapter, err := NewAdapter(ctx, testAgentID, testAliasID, DefaultConfig())
    assert.NoError(t, err)
    
    input := services.AgentInput{
        SessionID: "test-session-123",
        Message:   "Test logging",
    }
    
    _, err = adapter.InvokeAgent(ctx, input)
    
    logOutput := logBuffer.String()
    assert.Contains(t, logOutput, "InvokeAgent request")
    assert.Contains(t, logOutput, "SessionID: test-session-123")
    assert.Contains(t, logOutput, testAgentID)
}
```

### 10. Permission Verification Suite

**Purpose**: Tests IAM permissions and access control scenarios.

**IAM Permission Tests**:
```go
func TestIAMPermissions_ValidateAgentAccess(t *testing.T) {
    // Test with valid agent ID
    adapter, err := NewAdapter(ctx, testAgentID, testAliasID, DefaultConfig())
    assert.NoError(t, err)
    
    input := services.AgentInput{
        SessionID: generateSessionID(),
        Message:   "Test agent access",
    }
    
    response, err := adapter.InvokeAgent(ctx, input)
    assert.NoError(t, err)
    assert.NotEmpty(t, response.Content)
}

func TestIAMPermissions_InvalidAgentID(t *testing.T) {
    // Test with invalid agent ID
    adapter, err := NewAdapter(ctx, "invalid-agent-id", "invalid-alias", DefaultConfig())
    assert.NoError(t, err) // Constructor should succeed
    
    input := services.AgentInput{
        SessionID: generateSessionID(),
        Message:   "Test invalid agent",
    }
    
    _, err = adapter.InvokeAgent(ctx, input)
    assert.Error(t, err)
    
    var domainErr *services.DomainError
    assert.True(t, errors.As(err, &domainErr))
    assert.Equal(t, services.ErrCodeUnauthorized, domainErr.Code)
}

func TestIAMPermissions_KnowledgeBaseAccess(t *testing.T) {
    adapter, err := NewAdapter(ctx, testAgentID, testAliasID, DefaultConfig())
    assert.NoError(t, err)
    
    input := services.AgentInput{
        SessionID:        generateSessionID(),
        Message:          "Test knowledge base access",
        KnowledgeBaseIDs: []string{testKnowledgeBaseID},
    }
    
    response, err := adapter.InvokeAgent(ctx, input)
    assert.NoError(t, err)
    assert.NotEmpty(t, response.Content)
}
```

### 11. Environment Testing Framework

**Purpose**: Validates configuration across different deployment environments.

**Environment Configuration Tests**:
```go
func TestEnvironmentConfiguration_Development(t *testing.T) {
    // Load development configuration
    config := loadEnvironmentConfig("dev")
    
    adapter, err := NewAdapter(ctx, config.AgentID, config.AliasID, config.AdapterConfig)
    assert.NoError(t, err)
    
    // Test basic functionality
    input := services.AgentInput{
        SessionID: generateSessionID(),
        Message:   "Test dev environment",
    }
    
    response, err := adapter.InvokeAgent(ctx, input)
    assert.NoError(t, err)
    assert.NotEmpty(t, response.Content)
}

func TestEnvironmentConfiguration_Production(t *testing.T) {
    if !isProductionEnvironment() {
        t.Skip("Skipping production test - not in production environment")
    }
    
    // Load production configuration
    config := loadEnvironmentConfig("prod")
    
    adapter, err := NewAdapter(ctx, config.AgentID, config.AliasID, config.AdapterConfig)
    assert.NoError(t, err)
    
    // Test with production settings
    input := services.AgentInput{
        SessionID: generateSessionID(),
        Message:   "Test production environment",
    }
    
    response, err := adapter.InvokeAgent(ctx, input)
    assert.NoError(t, err)
    assert.NotEmpty(t, response.Content)
    
    // Verify VPC endpoint usage in production
    assert.True(t, isUsingVPCEndpoint())
}
```

### 12. Metrics and Monitoring Tests

**Purpose**: Validates metrics collection and monitoring capabilities.

**Metrics Collection Tests**:
```go
func TestMetrics_SuccessRateTracking(t *testing.T) {
    metricsCollector := NewTestMetricsCollector()
    
    adapter, err := NewAdapter(ctx, testAgentID, testAliasID, DefaultConfig())
    assert.NoError(t, err)
    
    // Perform multiple requests
    for i := 0; i < 10; i++ {
        input := services.AgentInput{
            SessionID: generateSessionID(),
            Message:   fmt.Sprintf("Test message %d", i),
        }
        
        startTime := time.Now()
        response, err := adapter.InvokeAgent(ctx, input)
        duration := time.Since(startTime)
        
        if err == nil {
            metricsCollector.RecordSuccess(duration)
        } else {
            metricsCollector.RecordError(err, duration)
        }
    }
    
    // Verify metrics
    assert.True(t, metricsCollector.GetSuccessRate() > 0.8) // 80% success rate
    assert.True(t, metricsCollector.GetAverageLatency() < 5*time.Second)
}
```

### 13. Interface Compliance Tests

**Purpose**: Validates that the adapter correctly implements the domain service interface.

**Interface Implementation Tests**:
```go
func TestInterfaceCompliance_BedrockService(t *testing.T) {
    adapter, err := NewAdapter(ctx, testAgentID, testAliasID, DefaultConfig())
    assert.NoError(t, err)
    
    // Verify adapter implements BedrockService interface
    var _ services.BedrockService = adapter
    
    // Test InvokeAgent method
    input := services.AgentInput{
        SessionID: generateSessionID(),
        Message:   "Test interface compliance",
    }
    
    response, err := adapter.InvokeAgent(ctx, input)
    assert.NoError(t, err)
    assert.IsType(t, &services.AgentResponse{}, response)
    
    // Test InvokeAgentStream method
    streamReader, err := adapter.InvokeAgentStream(ctx, input)
    assert.NoError(t, err)
    assert.Implements(t, (*services.StreamReader)(nil), streamReader)
    defer streamReader.Close()
}

func TestInterfaceCompliance_DomainErrors(t *testing.T) {
    // Test that all errors returned are domain errors
    mockClient := &mockBedrockClient{
        invokeAgentFunc: func(ctx context.Context, input *bedrockagentruntime.InvokeAgentInput) (*bedrockagentruntime.InvokeAgentOutput, error) {
            return nil, errors.New("generic error")
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
        Message:   "Test error handling",
    }
    
    _, err := adapter.InvokeAgent(ctx, input)
    
    var domainErr *services.DomainError
    assert.True(t, errors.As(err, &domainErr))
    assert.NotEmpty(t, domainErr.Code)
    assert.NotEmpty(t, domainErr.Message)
}
```

## Data Models

### Test Configuration

**Environment-Specific Test Configuration**:
```go
type TestConfig struct {
    AgentID           string
    AliasID           string
    KnowledgeBaseID   string
    Region            string
    VPCEnabled        bool
    AdapterConfig     AdapterConfig
    TestTimeout       time.Duration
    MaxConcurrency    int
}

type EnvironmentConfig struct {
    Development TestConfig
    Staging     TestConfig
    Production  TestConfig
}
```

**Test Data Models**:
```go
type TestCase struct {
    Name        string
    Input       services.AgentInput
    Expected    ExpectedResult
    ShouldError bool
    ErrorCode   string
}

type ExpectedResult struct {
    ContentNotEmpty    bool
    CitationsExpected  bool
    MetadataKeys       []string
    MinContentLength   int
    MaxResponseTime    time.Duration
}

type TestMetrics struct {
    TotalRequests    int
    SuccessfulRequests int
    FailedRequests   int
    AverageLatency   time.Duration
    ErrorsByType     map[string]int
}
```

### Mock Interfaces

**AWS SDK Mocking**:
```go
type MockBedrockClient interface {
    InvokeAgent(ctx context.Context, input *bedrockagentruntime.InvokeAgentInput) (*bedrockagentruntime.InvokeAgentOutput, error)
}

type MockEventStream interface {
    Events() <-chan types.ResponseStreamEvent
    Err() error
}

type ErrorSimulator struct {
    ErrorType     string
    ErrorMessage  string
    RetryCount    int
    DelayBetween  time.Duration
}
```

## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system—essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*

Now I'll analyze the acceptance criteria for testability using the prework tool:

Based on the prework analysis, I'll now define the key correctness properties that should hold universally across all valid executions:

### Property 1: Valid Input Response Completeness
*For any* valid agent input, the system should return a complete response containing non-empty content, a request ID, and properly structured citations array (empty or populated).
**Validates: Requirements 1.1, 1.2, 13.3**

### Property 2: Session Context Preservation
*For any* sequence of messages with the same session ID, the system should maintain conversation context such that later messages can reference information from earlier messages in the session.
**Validates: Requirements 1.4**

### Property 3: Input Validation Consistency
*For any* input with invalid knowledge base ID formats, the system should reject the input and return a validation error with appropriate error code.
**Validates: Requirements 2.4**

### Property 4: Stream Interface Compliance
*For any* streaming request, the system should return a StreamReader that implements all required methods (Read, ReadCitation, Close) and properly signals completion.
**Validates: Requirements 3.1, 3.4, 3.6, 13.4**

### Property 5: Error Transformation Consistency
*For any* AWS SDK error, the system should transform it into a domain error with appropriate error code, message, and retryable flag based on the error type.
**Validates: Requirements 4.2, 4.3, 4.5, 4.6, 13.2**

### Property 6: Retry Logic Compliance
*For any* retryable error, the system should retry up to the configured maximum with exponential backoff that respects the maximum backoff duration.
**Validates: Requirements 5.1, 5.2, 5.3**

### Property 7: Configuration Application
*For any* custom configuration provided to the adapter, the system should use those values for retries, timeouts, and backoff calculations.
**Validates: Requirements 6.4**

### Property 8: AWS Integration Consistency
*For any* API call made to AWS Bedrock, the system should use the correct region and endpoint, extract request IDs when available, and log request details.
**Validates: Requirements 7.2, 7.4, 7.5**

### Property 9: Citation Conversion Completeness
*For any* AWS citation in the response, the system should convert it to domain citation format preserving all available metadata, excerpts, and source information.
**Validates: Requirements 8.1, 8.2, 8.3, 8.4**

### Property 10: Logging Consistency
*For any* agent invocation, error occurrence, or retry attempt, the system should log structured information with appropriate details for debugging and monitoring.
**Validates: Requirements 9.1, 9.2, 9.3, 9.4**

### Property 11: Permission Error Clarity
*For any* IAM permission issue, invalid resource ID, or access restriction, the system should return clear error messages with specific resource information and actionable guidance.
**Validates: Requirements 10.2, 10.3, 10.4, 10.7**

### Property 12: Environment Resource Validation
*For any* deployment environment, the system should validate that all required resources (agent IDs, knowledge base IDs, VPC endpoints) are accessible and properly configured.
**Validates: Requirements 11.1, 11.2, 11.3, 11.5**

### Property 13: Metrics and Monitoring Emission
*For any* Bedrock API call, error occurrence, or performance issue, the system should emit appropriate metrics and structured logs for monitoring and alerting.
**Validates: Requirements 12.1, 12.2, 12.3, 12.5**

### Property 14: Domain Abstraction Preservation
*For any* interaction between the adapter and domain layer, the system should never expose AWS SDK types and should work entirely through domain interfaces and types.
**Validates: Requirements 13.5**

## Error Handling

### AWS SDK Error Scenarios

**Authentication and Authorization Errors**:
- `AccessDeniedException`: Transform to `ErrCodeUnauthorized` with clear message about IAM permissions
- `UnauthorizedException`: Transform to `ErrCodeUnauthorized` with specific resource information
- Invalid agent/alias IDs: Transform to `ErrCodeUnauthorized` with resource-specific guidance

**Rate Limiting and Throttling**:
- `ThrottlingException`: Transform to `ErrCodeRateLimit` with retryable flag set to true
- `TooManyRequestsException`: Transform to `ErrCodeRateLimit` with exponential backoff
- Implement jitter in backoff calculations to avoid thundering herd

**Service Availability Errors**:
- `ServiceUnavailableException`: Transform to `ErrCodeServiceError` with retryable flag
- `InternalServerException`: Transform to `ErrCodeServiceError` with retryable flag
- Network timeouts: Transform to `ErrCodeTimeout` with retryable flag

**Validation Errors**:
- `ValidationException`: Transform to `ErrCodeInvalidInput` with non-retryable flag
- `InvalidParameterException`: Transform to `ErrCodeInvalidInput` with parameter details
- Input validation failures: Return validation errors before making API calls

### Context and Timeout Handling

**Context Cancellation**:
- `context.Canceled`: Transform to `ErrCodeNetworkError` with non-retryable flag
- Propagate cancellation to streaming operations
- Clean up resources when context is canceled

**Timeout Management**:
- `context.DeadlineExceeded`: Transform to `ErrCodeTimeout` with retryable flag
- Configure appropriate timeouts for different operation types
- Allow timeout configuration per environment

### Streaming Error Handling

**Stream Processing Errors**:
- Malformed stream events: Transform to `ErrCodeMalformedStream`
- Stream interruption: Attempt to recover or return partial results
- Citation parsing errors: Log errors but continue processing content

**Resource Cleanup**:
- Ensure streams are properly closed on errors
- Release resources even when errors occur
- Implement timeout for stream operations

### VPC Endpoint Error Handling

**Network Connectivity**:
- VPC endpoint unreachable: Transform to `ErrCodeNetworkError` with VPC-specific message
- Security group restrictions: Provide actionable error messages about required ports
- DNS resolution failures: Include VPC endpoint configuration guidance

**Cross-Account Access**:
- Trust relationship issues: Provide specific guidance about IAM role configuration
- Resource policy restrictions: Include information about required resource policies
- Cross-region access: Validate region availability and configuration

## Testing Strategy

### Dual Testing Approach

The testing strategy employs both unit testing and property-based testing to provide comprehensive coverage:

**Unit Tests**:
- Test specific examples and edge cases
- Validate error transformation logic
- Test configuration handling
- Verify interface implementations
- Mock AWS SDK for isolated testing

**Property-Based Tests**:
- Validate universal properties across all inputs
- Test with generated data to find edge cases
- Verify retry logic with simulated failures
- Test streaming behavior with various response patterns
- Use **Go's testing/quick package** for property-based testing
- Configure each property test to run a minimum of **100 iterations**

### Property-Based Testing Implementation

**Testing Framework**: Use Go's built-in `testing/quick` package for property-based testing.

**Test Configuration**: Each property-based test must run a minimum of 100 iterations to ensure adequate coverage of the input space.

**Property Test Tagging**: Each property-based test must include a comment with the exact format:
`// **Feature: bedrock-agent-verification, Property {number}: {property_text}**`

**Example Property Test**:
```go
func TestProperty1_ValidInputResponseCompleteness(t *testing.T) {
    // **Feature: bedrock-agent-verification, Property 1: Valid Input Response Completeness**
    
    adapter, err := NewAdapter(ctx, testAgentID, testAliasID, DefaultConfig())
    require.NoError(t, err)
    
    property := func(sessionID string, message string) bool {
        // Generate valid input
        if sessionID == "" {
            sessionID = generateValidSessionID()
        }
        if message == "" || len(message) > 25000 {
            message = generateValidMessage()
        }
        
        input := services.AgentInput{
            SessionID: sessionID,
            Message:   message,
        }
        
        response, err := adapter.InvokeAgent(ctx, input)
        if err != nil {
            return false
        }
        
        // Verify response completeness
        return response.Content != "" &&
               response.RequestID != "" &&
               response.Citations != nil // Should be non-nil (empty or populated)
    }
    
    config := &quick.Config{MaxCount: 100}
    if err := quick.Check(property, config); err != nil {
        t.Errorf("Property violation: %v", err)
    }
}
```

### Integration Testing Strategy

**Test Environment Setup**:
- Use dedicated AWS test account with Bedrock resources
- Create test agents and knowledge bases for integration testing
- Configure IAM roles with appropriate permissions for testing
- Set up VPC endpoints in staging environment for VPC testing

**Test Data Management**:
- Generate test documents for knowledge base testing
- Create test conversation scenarios for session context testing
- Prepare error simulation scenarios for resilience testing
- Maintain test data cleanup procedures

**Parallel Test Execution**:
- Run unit tests in parallel for fast feedback
- Serialize integration tests to avoid resource conflicts
- Use test isolation to prevent cross-test interference
- Implement test timeouts to prevent hanging tests

### Performance and Load Testing

**Latency Testing**:
- Measure response times under normal conditions
- Test performance with large messages (up to 25000 characters)
- Validate streaming performance with incremental content delivery
- Monitor performance across different AWS regions

**Concurrency Testing**:
- Test multiple concurrent requests with different session IDs
- Validate thread safety of the adapter implementation
- Test connection pooling and resource management
- Verify proper cleanup under concurrent load

**Rate Limiting Testing**:
- Simulate rate limiting scenarios
- Validate exponential backoff behavior under load
- Test recovery after rate limiting periods
- Verify metrics collection during rate limiting

### Monitoring and Observability Testing

**Metrics Validation**:
- Verify success rate metrics are accurate
- Test latency distribution tracking
- Validate error rate categorization
- Test alert generation for threshold breaches

**Log Structure Testing**:
- Verify structured logging format consistency
- Test log correlation with request IDs
- Validate sensitive data redaction in logs
- Test log aggregation and searchability

**Debugging Support Testing**:
- Verify trace information is captured correctly
- Test error context preservation through the call stack
- Validate debugging information in production-safe format
- Test integration with monitoring systems

This comprehensive testing strategy ensures that the Bedrock Agent integration is thoroughly validated across all scenarios, providing confidence for production deployment while maintaining the ability to diagnose and resolve issues quickly.