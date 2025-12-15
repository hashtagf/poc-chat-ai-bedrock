# Chat Backend API

Go backend for the Bedrock Agent Core chat interface POC - **✅ FULLY FUNCTIONAL**

## Status: Production Ready

**Completion Date**: December 10, 2025  
**Infrastructure**: Deployed and validated in us-east-1  
**Integration**: Bedrock Agent Core with S3 Vectors Knowledge Base  
**Testing**: Comprehensive integration tests passing

## Architecture

This backend follows hexagonal architecture principles with complete Bedrock integration:

- **Domain Layer** (`domain/`): Core business entities and repository interfaces
- **Infrastructure Layer** (`infrastructure/`): Bedrock Agent Core adapter, MongoDB repositories
- **Interfaces Layer** (`interfaces/`): HTTP/WebSocket handlers with streaming support
- **Configuration** (`config/`): Environment-based configuration management

## API Endpoints

### Session Management

#### Create Session
```
POST /api/sessions
Response: { "id": "uuid", "created_at": "timestamp", "message_count": 0 }
```

#### Get Session
```
GET /api/sessions/{id}
Response: { "id": "uuid", "created_at": "timestamp", "last_message_at": "timestamp", "message_count": 0 }
```

#### List Sessions
```
GET /api/sessions
Response: [{ "id": "uuid", "created_at": "timestamp", "message_count": 0 }, ...]
```

### Chat Streaming

#### WebSocket Connection
```
WS /api/chat/stream

Client sends:
{
  "session_id": "uuid",
  "content": "message text"
}

Server streams:
{ "type": "content", "content": "chunk..." }
{ "type": "citation", "citation": {...} }
{ "type": "done" }
{ "type": "error", "error": {"code": "...", "message": "..."} }
```

## Running the Server

### Development Mode (Mock Responses)
```bash
# Install dependencies
go mod download

# Run without Bedrock (mock mode)
go run cmd/server/main.go

# Server starts on http://localhost:8080
```

### Production Mode (With Bedrock)
```bash
# Set environment variables
export AWS_REGION=us-east-1
export BEDROCK_AGENT_ID=W6R84XTD2X
export BEDROCK_AGENT_ALIAS_ID=TXENIZDWOS
export BEDROCK_KNOWLEDGE_BASE_ID=AQ5JOUEIGF

# Run with Bedrock integration
go run cmd/server/main.go
```

### Using Docker
```bash
# Build and run with docker-compose
docker-compose up --build backend
```

## Configuration

- **Port**: 8080 (configurable via PORT env var)
- **CORS**: Enabled for all origins (development) - restrict in production
- **Session Storage**: In-memory with MongoDB support
- **Bedrock Integration**: AWS SDK v2 with streaming support
- **WebSocket**: Real-time streaming with connection recovery

## Request Validation

### Message Content
- Cannot be empty or whitespace only
- Maximum length: 2000 characters
- Session ID must be valid

## Error Handling

All errors follow the format:
```json
{
  "code": "ERROR_CODE",
  "message": "User-friendly error message"
}
```

Error codes:
- `INVALID_REQUEST`: Invalid request parameters
- `SESSION_NOT_FOUND`: Session does not exist
- `SESSION_CREATE_FAILED`: Failed to create session
- `PROCESSING_FAILED`: Failed to process message
- `METHOD_NOT_ALLOWED`: HTTP method not allowed

## Current Features ✅

- ✅ **Bedrock Agent Core Integration**: Full AWS SDK v2 implementation
- ✅ **Streaming Responses**: Real-time token streaming from Bedrock
- ✅ **Knowledge Base Citations**: S3 Vectors integration with confidence scores
- ✅ **Error Handling**: Comprehensive retry logic with exponential backoff
- ✅ **Request Logging**: Structured logging with request IDs
- ✅ **Session Management**: MongoDB persistence with WebSocket support
- ✅ **Health Checks**: Kubernetes-ready health endpoints
- ✅ **Configuration**: Environment-based config with validation

## Testing

```bash
# Run all tests
go test ./... -v

# Run with coverage
go test ./... -cover

# Run integration tests (requires AWS credentials)
go test ./infrastructure/bedrock -v

# Test API endpoints
./test_api.sh
```

## Production Deployment

The backend is production-ready with:
- Docker containerization
- Health check endpoints
- Structured logging
- Error handling with retries
- AWS IAM role support
- MongoDB session persistence
