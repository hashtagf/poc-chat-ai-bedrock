# Configuration Management

This package provides centralized configuration management for the chat backend application.

## Overview

Configuration is loaded from environment variables with sensible defaults. The package supports multiple environments (development, production, test) with environment-specific configuration files.

## Configuration Structure

```go
type Config struct {
    Environment string
    Server      ServerConfig
    AWS         AWSConfig
    Bedrock     BedrockConfig
    WebSocket   WebSocketConfig
    Session     SessionConfig
    Logging     LoggingConfig
}
```

## Usage

### Loading Configuration

```go
import "github.com/bedrock-chat-poc/backend/config"

func main() {
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("Failed to load configuration: %v", err)
    }

    // Use configuration
    if cfg.IsDevelopment() {
        log.Println("Running in development mode")
    }
}
```

### Environment Files

The package includes environment-specific configuration files:

- `development.env` - Development environment settings
- `production.env` - Production environment settings

To use an environment file:

```bash
# Load development configuration
export $(cat backend/config/development.env | xargs)

# Or use with docker-compose
docker-compose --env-file backend/config/development.env up
```

## Configuration Options

### Environment

- `ENVIRONMENT` - Application environment (development, production, test)
  - Default: `development`

### Server Configuration

- `SERVER_PORT` - HTTP server port
  - Default: `8080`
- `SERVER_HOST` - HTTP server host
  - Default: `0.0.0.0`

### AWS Configuration

- `AWS_REGION` - AWS region for Bedrock services
  - Default: `us-east-1`
  - Required: Yes
- `AWS_ACCESS_KEY_ID` - AWS access key (development only)
  - Default: empty
  - Note: Use IAM roles in production
- `AWS_SECRET_ACCESS_KEY` - AWS secret key (development only)
  - Default: empty
  - Note: Use IAM roles in production
- `AWS_SESSION_TOKEN` - AWS session token (optional)
  - Default: empty

### Bedrock Configuration

- `BEDROCK_AGENT_ID` - Bedrock Agent ID
  - Default: empty
  - Required: Yes (in production)
- `BEDROCK_AGENT_ALIAS_ID` - Bedrock Agent Alias ID
  - Default: empty
  - Required: Yes (in production)
- `BEDROCK_KNOWLEDGE_BASE_ID` - Knowledge Base ID
  - Default: empty
- `BEDROCK_MODEL_ID` - Model identifier
  - Default: `anthropic.claude-v2`
- `BEDROCK_MAX_RETRIES` - Maximum retry attempts for rate limits
  - Default: `3`
- `BEDROCK_INITIAL_BACKOFF` - Initial backoff duration for retries
  - Default: `1s`
- `BEDROCK_MAX_BACKOFF` - Maximum backoff duration
  - Default: `30s`
- `BEDROCK_REQUEST_TIMEOUT` - Request timeout duration
  - Default: `60s`

### WebSocket Configuration

- `WS_TIMEOUT` - WebSocket connection timeout
  - Default: `30s`
- `WS_BUFFER_SIZE` - WebSocket buffer size
  - Default: `8192`
- `WS_READ_BUFFER_SIZE` - WebSocket read buffer size
  - Default: `1024`
- `WS_WRITE_BUFFER_SIZE` - WebSocket write buffer size
  - Default: `1024`
- `WS_STREAM_TIMEOUT` - Maximum time for entire stream
  - Default: `5m`
- `WS_CHUNK_TIMEOUT` - Maximum time between chunks
  - Default: `30s`

### Session Configuration

- `SESSION_TIMEOUT` - Session inactivity timeout
  - Default: `30m`

### Logging Configuration

- `LOG_LEVEL` - Logging level (debug, info, warn, error)
  - Default: `info`
- `LOG_FORMAT` - Log format (text, json)
  - Default: `text`

## Validation

Configuration is automatically validated on load. The following validations are performed:

- Environment must be one of: development, production, test
- Server port must be specified
- AWS region must be specified
- In production: Bedrock agent ID and alias ID are required
- WebSocket timeout and buffer size must be positive
- Session timeout must be positive

## Best Practices

### Development

1. Use `development.env` as a template
2. Copy to `.env` and customize for your local setup
3. Never commit `.env` files with credentials
4. Use AWS CLI profiles or environment variables for credentials

### Production

1. Use IAM roles for AWS credentials - never hardcode keys
2. Set all required Bedrock configuration values
3. Use appropriate timeouts for production workloads
4. Enable JSON logging for better log aggregation
5. Set log level to `info` or `warn` to reduce noise

### Testing

1. Set `ENVIRONMENT=test` for test runs
2. Use mock Bedrock services when agent ID is not set
3. Reduce timeouts for faster test execution

## Security Considerations

- **Never commit credentials** to version control
- Use IAM roles in production environments
- Rotate credentials regularly
- Use AWS Secrets Manager for sensitive configuration in production
- Validate all configuration values before use
- Use principle of least privilege for IAM permissions

## Examples

### Development with Mock Bedrock

```bash
export ENVIRONMENT=development
export SERVER_PORT=8080
export AWS_REGION=us-east-1
# Leave BEDROCK_AGENT_ID empty to run in mock mode
```

### Production with Bedrock

```bash
export ENVIRONMENT=production
export SERVER_PORT=8080
export AWS_REGION=us-east-1
export BEDROCK_AGENT_ID=ABCDEFGHIJ
export BEDROCK_AGENT_ALIAS_ID=TSTALIASID
export BEDROCK_KNOWLEDGE_BASE_ID=KB123456
export LOG_LEVEL=info
export LOG_FORMAT=json
```

### Custom Timeouts

```bash
export WS_TIMEOUT=60s
export WS_STREAM_TIMEOUT=10m
export BEDROCK_REQUEST_TIMEOUT=120s
export SESSION_TIMEOUT=1h
```
