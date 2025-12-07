# Bedrock Adapter

This package provides an adapter implementation for Amazon Bedrock Agent Core, following hexagonal architecture principles.

## Overview

The Bedrock adapter implements the `BedrockService` interface defined in the domain layer, providing integration with AWS Bedrock Agent Core using AWS SDK v2.

## Architecture

```
Domain Layer (Port)
    ↓
BedrockService Interface
    ↓
Infrastructure Layer (Adapter)
    ↓
AWS SDK v2 (Bedrock Agent Runtime)
```

## Features

- **AWS SDK v2 Integration**: Uses the latest AWS SDK for Go v2
- **Retry Logic**: Exponential backoff for rate limits and transient errors
- **Error Transformation**: Converts AWS SDK errors to domain-specific errors
- **Request Logging**: Logs all API calls with request IDs for debugging
- **Streaming Support**: Handles both complete and streaming responses
- **Context Support**: Respects context cancellation and timeouts

## Usage

### Creating an Adapter

```go
import (
    "context"
    "github.com/bedrock-chat-poc/backend/infrastructure/bedrock"
)

func main() {
    ctx := context.Background()
    
    // Create adapter with default configuration
    adapter, err := bedrock.NewAdapter(
        ctx,
        "your-agent-id",
        "your-alias-id",
        bedrock.DefaultConfig(),
    )
    if err != nil {
        log.Fatal(err)
    }
    
    // Use adapter...
}
```

### Configuration

The adapter supports the following configuration options:

```go
type AdapterConfig struct {
    MaxRetries     int           // Maximum retry attempts (default: 3)
    InitialBackoff time.Duration // Initial backoff duration (default: 1s)
    MaxBackoff     time.Duration // Maximum backoff duration (default: 30s)
    RequestTimeout time.Duration // Request timeout (default: 60s)
}
```

### Invoking the Agent (Complete Response)

```go
input := services.AgentInput{
    SessionID: "session-123",
    Message:   "What is the weather today?",
    KnowledgeBaseIDs: []string{"kb-id-1"},
}

response, err := adapter.InvokeAgent(ctx, input)
if err != nil {
    // Handle error
    var domainErr *services.DomainError
    if errors.As(err, &domainErr) {
        log.Printf("Error code: %s, Retryable: %v", domainErr.Code, domainErr.Retryable)
    }
    return
}

log.Printf("Response: %s", response.Content)
log.Printf("Citations: %d", len(response.Citations))
```

### Streaming Responses

```go
stream, err := adapter.InvokeAgentStream(ctx, input)
if err != nil {
    log.Fatal(err)
}
defer stream.Close()

for {
    chunk, done, err := stream.Read()
    if err != nil {
        log.Printf("Stream error: %v", err)
        break
    }
    
    if done {
        break
    }
    
    // Process chunk
    fmt.Print(chunk)
    
    // Check for citations
    if citation, err := stream.ReadCitation(); err == nil && citation != nil {
        log.Printf("Citation: %s", citation.SourceName)
    }
}
```

## Error Handling

The adapter transforms AWS SDK errors into domain-specific errors with the following codes:

- `RATE_LIMIT_EXCEEDED`: Rate limit errors (retryable)
- `INVALID_INPUT`: Validation errors (not retryable)
- `SERVICE_ERROR`: Service errors (may be retryable)
- `NETWORK_ERROR`: Network errors (not retryable)
- `TIMEOUT`: Request timeout (retryable)
- `UNAUTHORIZED`: Authentication/authorization errors (not retryable)
- `MALFORMED_STREAM`: Stream parsing errors (not retryable)

### Retry Logic

The adapter automatically retries requests for the following error types:
- `ThrottlingException`
- `TooManyRequestsException`
- `ServiceUnavailableException`

Retries use exponential backoff with the following formula:
```
backoff = InitialBackoff * 2^(attempt-1)
```

Capped at `MaxBackoff` duration.

## AWS Configuration

The adapter uses the AWS SDK's default credential chain:
1. Environment variables (`AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`)
2. Shared credentials file (`~/.aws/credentials`)
3. IAM role (recommended for production)

### Required IAM Permissions

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "bedrock:InvokeAgent",
        "bedrock:InvokeAgentStream"
      ],
      "Resource": [
        "arn:aws:bedrock:*:*:agent/*",
        "arn:aws:bedrock:*:*:agent-alias/*"
      ]
    }
  ]
}
```

## Logging

The adapter logs the following information:
- Request initiation (SessionID, AgentID)
- Retry attempts with backoff duration
- Stream events (chunks, traces)
- Errors with request IDs
- Response completion (content length, citation count)

All logs are prefixed with `[Bedrock]` for easy filtering.

## Testing

Run tests with:
```bash
go test ./infrastructure/bedrock/... -v
```

The test suite includes:
- Input validation tests
- Backoff calculation tests
- Error transformation tests
- Configuration tests

## Implementation Notes

### Citation Extraction

Citations are extracted from the Bedrock response and converted to domain entities. The adapter handles:
- Text excerpts from generated responses
- Source references from knowledge bases
- S3 location URIs
- Metadata from retrieved references

### Stream Processing

The streaming implementation uses Go channels to process events from the Bedrock event stream. The adapter:
- Buffers content chunks for efficient reading
- Extracts citations as they arrive
- Handles trace events for debugging
- Properly closes streams on completion or error

### Context Handling

The adapter respects context cancellation and timeouts:
- Request-level timeouts via `RequestTimeout` configuration
- Context cancellation during retries
- Stream cancellation during reading

## Future Enhancements

Potential improvements for production use:
- Request/response caching
- Metrics collection (latency, error rates)
- Circuit breaker pattern for fault tolerance
- Request ID propagation for distributed tracing
- Knowledge base query optimization
