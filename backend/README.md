# Chat Backend API

Go backend for the Bedrock Agent Core chat interface POC.

## Architecture

This backend follows hexagonal architecture principles:

- **Domain Layer** (`domain/`): Core business entities and repository interfaces
- **Application Layer**: Use cases (to be implemented with Bedrock integration)
- **Infrastructure Layer** (`infrastructure/`): External adapters (repositories, AWS SDK)
- **Interfaces Layer** (`interfaces/`): HTTP/WebSocket handlers

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

```bash
# Install dependencies
go mod download

# Run the server
go run cmd/server/main.go

# Server starts on http://localhost:8080
```

## Configuration

- **Port**: 8080 (default)
- **CORS**: Enabled for all origins (POC only - restrict in production)
- **Session Storage**: In-memory (POC only - use persistent storage in production)

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

## Next Steps

- Integrate with AWS Bedrock Agent Core SDK
- Implement streaming processor for Bedrock responses
- Add citation extraction from knowledge base
- Implement retry logic with exponential backoff
- Add request/response logging with request IDs
- Add authentication and authorization
