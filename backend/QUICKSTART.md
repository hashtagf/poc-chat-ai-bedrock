# Backend Quick Start Guide

## Running the Server

### Option 1: Direct Go Run
```bash
cd backend
go run cmd/server/main.go
```

Server starts on `http://localhost:8080`

### Option 2: Build and Run
```bash
cd backend
make build
./bin/server
```

### Option 3: Using Docker
```bash
cd backend
docker build -t chat-backend .
docker run -p 8080:8080 chat-backend
```

## Testing the API

### 1. Health Check
```bash
curl http://localhost:8080/health
# Expected: OK
```

### 2. Create a Session
```bash
curl -X POST http://localhost:8080/api/sessions
# Returns: {"id":"uuid","created_at":"...","message_count":0}
```

### 3. Get Session Details
```bash
# Replace {session-id} with actual ID from step 2
curl http://localhost:8080/api/sessions/{session-id}
```

### 4. List All Sessions
```bash
curl http://localhost:8080/api/sessions
```

### 5. Test WebSocket Streaming

Using the provided test client:
```bash
# Build the client
go build -o bin/wsclient cmd/wsclient/main.go

# Run with your session ID
./bin/wsclient -session {session-id} -message "Hello, world!"
```

Or use the test script:
```bash
./test_api.sh
```

## Running Tests

```bash
# Run all tests
make test

# Run with coverage
go test ./... -cover

# Run specific package tests
go test ./interfaces/chat -v
go test ./infrastructure/repositories -v
```

## Development Workflow

1. Make code changes
2. Format code: `make fmt`
3. Check for issues: `make vet`
4. Run tests: `make test`
5. Build: `make build`

Or run all checks at once:
```bash
make check
```

## API Endpoints Summary

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check |
| POST | `/api/sessions` | Create new session |
| GET | `/api/sessions` | List all sessions |
| GET | `/api/sessions/{id}` | Get session details |
| WS | `/api/chat/stream` | WebSocket for streaming chat |

## WebSocket Message Format

### Client → Server
```json
{
  "session_id": "uuid",
  "content": "Your message text"
}
```

### Server → Client
```json
// Content chunk
{"type": "content", "content": "Response text..."}

// Citation (future)
{"type": "citation", "citation": {...}}

// Completion
{"type": "done"}

// Error
{"type": "error", "error": {"code": "ERROR_CODE", "message": "..."}}
```

## Validation Rules

### Message Content
- Cannot be empty or whitespace only
- Maximum length: 2000 characters
- Session ID must be valid UUID

### Error Codes
- `INVALID_REQUEST`: Invalid request parameters
- `SESSION_NOT_FOUND`: Session does not exist
- `SESSION_CREATE_FAILED`: Failed to create session
- `PROCESSING_FAILED`: Failed to process message
- `METHOD_NOT_ALLOWED`: HTTP method not allowed

## Next Steps

This backend currently provides:
- ✅ Session management (create, get, list)
- ✅ WebSocket streaming infrastructure
- ✅ Request validation
- ✅ Error handling
- ✅ CORS support

To integrate with Bedrock Agent Core:
1. Add AWS SDK for Go v2 dependencies
2. Implement BedrockAdapter in `infrastructure/bedrock/`
3. Create StreamProcessor for parsing Bedrock responses
4. Update WebSocket handler to use Bedrock adapter
5. Add citation extraction logic

See task 15 in the implementation plan for Bedrock integration details.
