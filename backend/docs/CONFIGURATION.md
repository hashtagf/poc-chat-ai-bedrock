# Configuration Management Guide

This guide explains how to configure the chat backend application for different environments.

## Table of Contents

- [Overview](#overview)
- [Quick Start](#quick-start)
- [Configuration Options](#configuration-options)
- [Environment-Specific Setup](#environment-specific-setup)
- [AWS Credentials](#aws-credentials)
- [Bedrock Configuration](#bedrock-configuration)
- [WebSocket Configuration](#websocket-configuration)
- [Logging Configuration](#logging-configuration)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)

## Overview

The application uses environment variables for configuration with sensible defaults. Configuration is centralized in the `config` package and validated on startup.

### Configuration Hierarchy

1. Environment variables (highest priority)
2. Default values (defined in code)

### Supported Environments

- `development` - Local development with optional mock mode
- `production` - Production deployment with full Bedrock integration
- `test` - Testing environment

## Quick Start

### Development Setup

1. Copy the example environment file:
   ```bash
   cp .env.example .env
   ```

2. Edit `.env` with your configuration:
   ```bash
   # Minimal development setup (mock mode)
   ENVIRONMENT=development
   AWS_REGION=ap-southeast-1
   ```

3. Run the server:
   ```bash
   go run cmd/server/main.go
   ```

### With Bedrock Integration

1. Set Bedrock credentials in `.env`:
   ```bash
   ENVIRONMENT=development
   AWS_REGION=ap-southeast-1
   BEDROCK_AGENT_ID=your_agent_id
   BEDROCK_AGENT_ALIAS_ID=your_alias_id
   ```

2. Configure AWS credentials (see [AWS Credentials](#aws-credentials))

3. Run the server:
   ```bash
   go run cmd/server/main.go
   ```

## Configuration Options

### Environment Variables

#### Server Configuration

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `ENVIRONMENT` | Application environment | `development` | No |
| `SERVER_PORT` | HTTP server port | `8080` | No |
| `SERVER_HOST` | HTTP server host | `0.0.0.0` | No |

#### AWS Configuration

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `AWS_REGION` | AWS region | `ap-southeast-1` | Yes |
| `AWS_ACCESS_KEY_ID` | AWS access key | - | No* |
| `AWS_SECRET_ACCESS_KEY` | AWS secret key | - | No* |
| `AWS_SESSION_TOKEN` | AWS session token | - | No |

*Use IAM roles in production instead of access keys

#### Bedrock Configuration

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `BEDROCK_AGENT_ID` | Bedrock Agent ID | - | Yes (prod) |
| `BEDROCK_AGENT_ALIAS_ID` | Bedrock Agent Alias ID | - | Yes (prod) |
| `BEDROCK_KNOWLEDGE_BASE_ID` | Knowledge Base ID | - | No |
| `BEDROCK_MODEL_ID` | Model identifier | `anthropic.claude-v2` | No |
| `BEDROCK_MAX_RETRIES` | Max retry attempts | `3` | No |
| `BEDROCK_INITIAL_BACKOFF` | Initial retry backoff | `1s` | No |
| `BEDROCK_MAX_BACKOFF` | Maximum retry backoff | `30s` | No |
| `BEDROCK_REQUEST_TIMEOUT` | Request timeout | `60s` | No |

#### WebSocket Configuration

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `WS_TIMEOUT` | Connection timeout | `30s` | No |
| `WS_BUFFER_SIZE` | Buffer size | `8192` | No |
| `WS_READ_BUFFER_SIZE` | Read buffer size | `1024` | No |
| `WS_WRITE_BUFFER_SIZE` | Write buffer size | `1024` | No |
| `WS_STREAM_TIMEOUT` | Stream timeout | `5m` | No |
| `WS_CHUNK_TIMEOUT` | Chunk timeout | `30s` | No |

#### Session Configuration

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `SESSION_TIMEOUT` | Session inactivity timeout | `30m` | No |

#### Logging Configuration

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `LOG_LEVEL` | Log level (debug, info, warn, error) | `info` | No |
| `LOG_FORMAT` | Log format (text, json) | `text` | No |

## Environment-Specific Setup

### Development

Use the provided `backend/config/development.env` as a template:

```bash
# Load development configuration
export $(cat backend/config/development.env | grep -v '^#' | xargs)

# Or source it
set -a
source backend/config/development.env
set +a

# Run the server
go run cmd/server/main.go
```

**Development Features:**
- Mock mode available (no Bedrock required)
- Debug logging enabled
- Configuration endpoint at `/api/config`
- Relaxed validation

### Production

Use the provided `backend/config/production.env` as a template:

```bash
# IMPORTANT: Customize production.env with your values
# Never commit production credentials to version control

# Load production configuration
export $(cat backend/config/production.env | grep -v '^#' | xargs)

# Run the server
./server
```

**Production Requirements:**
- Bedrock Agent ID and Alias ID are required
- Use IAM roles for AWS credentials
- JSON logging for log aggregation
- Strict validation
- No configuration endpoint

### Testing

```bash
export ENVIRONMENT=test
export AWS_REGION=ap-southeast-1
go test ./...
```

## AWS Credentials

### Development

**Option 1: AWS CLI Profile (Recommended)**
```bash
# Configure AWS CLI
aws configure

# The application will use default credentials
go run cmd/server/main.go
```

**Option 2: Environment Variables**
```bash
export AWS_ACCESS_KEY_ID=your_key
export AWS_SECRET_ACCESS_KEY=your_secret
go run cmd/server/main.go
```

**Option 3: IAM Role (EC2/ECS)**
```bash
# No configuration needed - uses instance role
go run cmd/server/main.go
```

### Production

**Always use IAM roles in production:**

1. Create an IAM role with Bedrock permissions:
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
         "Resource": "*"
       }
     ]
   }
   ```

2. Attach the role to your EC2 instance, ECS task, or Lambda function

3. The application will automatically use the role credentials

## Bedrock Configuration

### Finding Your Bedrock IDs

1. **Agent ID**: Found in the Bedrock console under "Agents"
   - Format: `ABCDEFGHIJ` (10 characters)

2. **Agent Alias ID**: Found in the agent's "Aliases" section
   - Format: `TSTALIASID` or similar
   - Use `TSTALIASID` for testing

3. **Knowledge Base ID**: Found in "Knowledge bases" section
   - Format: `KB123456` or similar

### Configuration Examples

**Minimal (Mock Mode):**
```bash
ENVIRONMENT=development
AWS_REGION=ap-southeast-1
# No Bedrock IDs - runs in mock mode
```

**With Bedrock:**
```bash
ENVIRONMENT=development
AWS_REGION=ap-southeast-1
BEDROCK_AGENT_ID=ABCDEFGHIJ
BEDROCK_AGENT_ALIAS_ID=TSTALIASID
```

**With Knowledge Base:**
```bash
ENVIRONMENT=production
AWS_REGION=ap-southeast-1
BEDROCK_AGENT_ID=ABCDEFGHIJ
BEDROCK_AGENT_ALIAS_ID=PRODALIASID
BEDROCK_KNOWLEDGE_BASE_ID=KB123456
```

### Retry Configuration

Adjust retry behavior for rate limits:

```bash
# Conservative (fewer retries, faster failure)
BEDROCK_MAX_RETRIES=2
BEDROCK_INITIAL_BACKOFF=500ms
BEDROCK_MAX_BACKOFF=10s

# Aggressive (more retries, longer wait)
BEDROCK_MAX_RETRIES=5
BEDROCK_INITIAL_BACKOFF=2s
BEDROCK_MAX_BACKOFF=60s
```

## WebSocket Configuration

### Timeout Configuration

```bash
# Short timeouts (faster failure detection)
WS_TIMEOUT=15s
WS_STREAM_TIMEOUT=2m
WS_CHUNK_TIMEOUT=10s

# Long timeouts (more patient)
WS_TIMEOUT=60s
WS_STREAM_TIMEOUT=10m
WS_CHUNK_TIMEOUT=60s
```

### Buffer Configuration

```bash
# Small buffers (memory constrained)
WS_BUFFER_SIZE=4096
WS_READ_BUFFER_SIZE=512
WS_WRITE_BUFFER_SIZE=512

# Large buffers (high throughput)
WS_BUFFER_SIZE=32768
WS_READ_BUFFER_SIZE=4096
WS_WRITE_BUFFER_SIZE=4096
```

## Logging Configuration

### Log Levels

- `debug` - Verbose logging for development
- `info` - Standard operational logging
- `warn` - Warning messages only
- `error` - Error messages only

### Log Formats

- `text` - Human-readable format for development
- `json` - Structured format for production log aggregation

### Examples

**Development:**
```bash
LOG_LEVEL=debug
LOG_FORMAT=text
```

**Production:**
```bash
LOG_LEVEL=info
LOG_FORMAT=json
```

## Best Practices

### Security

1. **Never commit credentials** to version control
   - Add `.env` to `.gitignore`
   - Use `.env.example` as a template

2. **Use IAM roles in production**
   - Never use access keys in production
   - Rotate credentials regularly

3. **Validate configuration on startup**
   - The application validates all configuration
   - Fails fast with clear error messages

4. **Use environment-specific files**
   - Keep development and production configs separate
   - Use different AWS accounts for environments

### Performance

1. **Tune timeouts for your use case**
   - Shorter timeouts for interactive applications
   - Longer timeouts for batch processing

2. **Adjust buffer sizes**
   - Larger buffers for high-throughput scenarios
   - Smaller buffers for memory-constrained environments

3. **Configure retry behavior**
   - More retries for unreliable networks
   - Fewer retries for cost-sensitive applications

### Monitoring

1. **Enable structured logging in production**
   ```bash
   LOG_FORMAT=json
   ```

2. **Set appropriate log levels**
   - Use `info` or `warn` in production
   - Use `debug` only for troubleshooting

3. **Monitor configuration endpoint** (development only)
   ```bash
   curl http://localhost:8080/api/config
   ```

## Troubleshooting

### Configuration Validation Errors

**Error: "invalid environment"**
- Ensure `ENVIRONMENT` is one of: `development`, `production`, `test`

**Error: "Bedrock agent ID is required in production"**
- Set `BEDROCK_AGENT_ID` and `BEDROCK_AGENT_ALIAS_ID` in production

**Error: "AWS region is required"**
- Set `AWS_REGION` environment variable

### Bedrock Connection Issues

**Error: "Failed to initialize Bedrock adapter"**
- Check AWS credentials are configured
- Verify IAM permissions for Bedrock
- Confirm agent ID and alias ID are correct

**Error: "Rate limit exceeded"**
- Increase `BEDROCK_MAX_RETRIES`
- Increase `BEDROCK_MAX_BACKOFF`
- Check Bedrock service quotas

### WebSocket Issues

**Error: "Stream timed out"**
- Increase `WS_STREAM_TIMEOUT`
- Check network connectivity
- Verify Bedrock service is responding

**Error: "Chunk timeout"**
- Increase `WS_CHUNK_TIMEOUT`
- Check for network latency issues

### Debugging

1. **Enable debug logging:**
   ```bash
   LOG_LEVEL=debug
   ```

2. **Check configuration at startup:**
   - Look for configuration log messages
   - Verify all values are loaded correctly

3. **Use configuration endpoint** (development):
   ```bash
   curl http://localhost:8080/api/config
   ```

4. **Test with mock mode:**
   ```bash
   unset BEDROCK_AGENT_ID
   unset BEDROCK_AGENT_ALIAS_ID
   go run cmd/server/main.go
   ```

## Additional Resources

- [AWS Bedrock Documentation](https://docs.aws.amazon.com/bedrock/)
- [AWS SDK for Go v2](https://aws.github.io/aws-sdk-go-v2/)
- [Go Environment Variables](https://pkg.go.dev/os#Getenv)
- [WebSocket Configuration](https://pkg.go.dev/github.com/gorilla/websocket)
