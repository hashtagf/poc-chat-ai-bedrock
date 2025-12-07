# WebSocket Integration Tests

This directory contains comprehensive integration tests for the WebSocket communication layer of the chat UI.

## Test Coverage

### Frontend Tests (`websocket.integration.test.ts`)

The frontend integration tests cover:

1. **Message Sending and Receiving** (Requirement 1.1)
   - End-to-end message transmission through WebSocket
   - Empty message validation
   - Whitespace-only message validation
   - Message length validation (max 2000 characters)

2. **Streaming Response Handling** (Requirement 2.1)
   - Incremental display of streaming responses
   - Streaming completion state transitions
   - Partial content preservation on errors
   - Input blocking during active streaming

3. **Connection Interruption and Reconnection**
   - Detection of connection interruptions during streaming
   - Automatic reconnection attempts with exponential backoff
   - Graceful handling of network errors

4. **Session Management Integration** (Requirement 7.1)
   - Session state maintenance across multiple messages
   - Session isolation between different sessions
   - Session switching functionality
   - Session metadata tracking

5. **Error Scenarios**
   - Malformed server response handling
   - Backend error response handling (rate limits, etc.)

### Backend Tests (`websocket_integration_test.go`)

The backend integration tests cover:

1. **Message Sending and Receiving** (Requirement 1.1)
   - WebSocket connection establishment
   - Message transmission and response streaming
   - Session state updates

2. **Streaming Response Handling** (Requirement 2.1)
   - Real-time chunk delivery
   - Streaming completion signals
   - Multiple chunks over time

3. **Input Validation**
   - Empty content rejection
   - Whitespace-only content rejection
   - Content length validation
   - Session ID validation

4. **Session Management** (Requirement 7.1)
   - Session existence validation
   - Multiple messages per session
   - Session message count tracking

5. **Connection Management**
   - Graceful connection closure
   - Concurrent connections handling
   - Malformed JSON handling

6. **Bedrock Service Integration**
   - Mock Bedrock service responses
   - Error handling from Bedrock service
   - Rate limit error propagation

## Running the Tests

### Frontend Tests
```bash
cd frontend
npm test -- websocket.integration.test.ts
```

### Backend Tests
```bash
cd backend
go test -v ./interfaces/chat/... -run TestWebSocket
```

## Test Architecture

### Frontend Mock WebSocket
The frontend tests use a custom `MockWebSocket` implementation that:
- Simulates connection establishment
- Allows programmatic message injection
- Supports error simulation
- Tracks active connections

### Backend Test Server
The backend tests use:
- `httptest.NewServer` for HTTP server simulation
- Real WebSocket connections via `gorilla/websocket`
- Mock Bedrock service for controlled responses
- In-memory session repository

## Key Test Patterns

1. **Async Testing**: All tests properly handle asynchronous WebSocket operations with appropriate timeouts
2. **State Verification**: Tests verify both immediate effects and persistent state changes
3. **Error Injection**: Tests systematically inject errors at various points to verify error handling
4. **Isolation**: Each test creates its own session and connection to ensure independence

## Requirements Validated

- ✅ **1.1**: Message transmission to Bedrock Agent Core
- ✅ **2.1**: Real-time streaming response display
- ✅ **7.1**: Session management and state

## Notes

- Frontend tests use a 50ms delay for connection establishment simulation
- Backend tests use real WebSocket connections for more accurate integration testing
- All tests include proper cleanup to prevent resource leaks
- Tests are designed to be deterministic and repeatable
