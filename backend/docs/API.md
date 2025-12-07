# API Documentation

Complete API reference for the Bedrock Chat UI backend.

## Table of Contents

- [Overview](#overview)
- [Authentication](#authentication)
- [Base URL](#base-url)
- [Response Format](#response-format)
- [Error Handling](#error-handling)
- [Rate Limiting](#rate-limiting)
- [Endpoints](#endpoints)
  - [Health Check](#health-check)
  - [Session Management](#session-management)
  - [Chat Streaming](#chat-streaming)
- [WebSocket Protocol](#websocket-protocol)
- [Error Codes](#error-codes)
- [Examples](#examples)

## Overview

The Bedrock Chat UI API provides RESTful endpoints for session management and WebSocket endpoints for real-time chat streaming with Amazon Bedrock Agent Core.

**API Version:** 1.0  
**Protocol:** HTTP/1.1, WebSocket  
**Content-Type:** application/json

## Authentication

**Current Status:** No authentication (POC)

**Future:** Will implement JWT-based authentication:
```
Authorization: Bearer <token>
```

## Base URL

**Development:**
```
http://localhost:8080
```

**Production:**
```
https://api.yourdomain.com
```

## Response Format

### Success Response

```json
{
  "data": { ... },
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### Error Response

```json
{
  "code": "ERROR_CODE",
  "message": "User-friendly error message",
  "timestamp": "2024-01-01T00:00:00Z"
}
```

## Error Handling

### HTTP Status Codes

| Status Code | Description |
|-------------|-------------|
| 200 | OK - Request succeeded |
| 201 | Created - Resource created successfully |
| 400 | Bad Request - Invalid request parameters |
| 404 | Not Found - Resource not found |
| 405 | Method Not Allowed - HTTP method not supported |
| 429 | Too Many Requests - Rate limit exceeded |
| 500 | Internal Server Error - Server error occurred |
| 503 | Service Unavailable - Service temporarily unavailable |

### Error Response Format

All errors follow a consistent format:

```json
{
  "code": "ERROR_CODE",
  "message": "Human-readable error message",
  "details": {
    "field": "Additional context (optional)"
  },
  "timestamp": "2024-01-01T00:00:00Z"
}
```

## Rate Limiting

**Current Status:** No rate limiting (POC)

**Future Implementation:**
- 100 requests per minute per IP
- 1000 requests per hour per user
- Headers: `X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Reset`

## Endpoints

### Health Check

Check if the API is running and healthy.

#### Request

```
GET /health
```

#### Response

**Status:** 200 OK

```json
{
  "status": "healthy",
  "timestamp": "2024-01-01T00:00:00Z",
  "version": "1.0.0"
}
```

#### Example

```bash
curl http://localhost:8080/health
```

---

### Session Management

#### Create Session

Create a new conversation session.

**Endpoint:** `POST /api/sessions`

**Request Body:** None

**Response:**

**Status:** 201 Created

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "created_at": "2024-01-01T00:00:00Z",
  "message_count": 0
}
```

**Errors:**

| Status | Code | Description |
|--------|------|-------------|
| 500 | SESSION_CREATE_FAILED | Failed to create session |

**Example:**

```bash
curl -X POST http://localhost:8080/api/sessions
```

```javascript
const response = await fetch('http://localhost:8080/api/sessions', {
  method: 'POST'
});
const session = await response.json();
console.log('Session ID:', session.id);
```

---

#### Get Session

Retrieve details for a specific session.

**Endpoint:** `GET /api/sessions/{id}`

**Path Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| id | string | Session UUID |

**Response:**

**Status:** 200 OK

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "created_at": "2024-01-01T00:00:00Z",
  "last_message_at": "2024-01-01T00:05:00Z",
  "message_count": 5
}
```

**Errors:**

| Status | Code | Description |
|--------|------|-------------|
| 404 | SESSION_NOT_FOUND | Session does not exist |

**Example:**

```bash
curl http://localhost:8080/api/sessions/550e8400-e29b-41d4-a716-446655440000
```

```javascript
const sessionId = '550e8400-e29b-41d4-a716-446655440000';
const response = await fetch(`http://localhost:8080/api/sessions/${sessionId}`);
const session = await response.json();
```

---

#### List Sessions

Retrieve all active sessions.

**Endpoint:** `GET /api/sessions`

**Query Parameters:**

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| limit | integer | 100 | Maximum number of sessions to return |
| offset | integer | 0 | Number of sessions to skip |

**Response:**

**Status:** 200 OK

```json
[
  {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "created_at": "2024-01-01T00:00:00Z",
    "message_count": 5
  },
  {
    "id": "660e8400-e29b-41d4-a716-446655440001",
    "created_at": "2024-01-01T01:00:00Z",
    "message_count": 2
  }
]
```

**Example:**

```bash
curl http://localhost:8080/api/sessions
curl "http://localhost:8080/api/sessions?limit=10&offset=0"
```

```javascript
const response = await fetch('http://localhost:8080/api/sessions');
const sessions = await response.json();
```

---

### Chat Streaming

#### WebSocket Connection

Establish a WebSocket connection for real-time chat streaming.

**Endpoint:** `ws://localhost:8080/api/chat/stream`

**Protocol:** WebSocket

**Connection:**

```javascript
const ws = new WebSocket('ws://localhost:8080/api/chat/stream');

ws.onopen = () => {
  console.log('Connected');
};

ws.onmessage = (event) => {
  const message = JSON.parse(event.data);
  console.log('Received:', message);
};

ws.onerror = (error) => {
  console.error('Error:', error);
};

ws.onclose = () => {
  console.log('Disconnected');
};
```

---

## WebSocket Protocol

### Client Messages

#### Send Message

Send a user message to the agent.

**Format:**

```json
{
  "session_id": "550e8400-e29b-41d4-a716-446655440000",
  "content": "What is Amazon Bedrock?"
}
```

**Fields:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| session_id | string | Yes | Valid session UUID |
| content | string | Yes | Message content (1-2000 characters) |

**Validation Rules:**

- `session_id`: Must be a valid UUID format
- `content`: 
  - Minimum length: 1 character (after trimming)
  - Maximum length: 2000 characters
  - Cannot be whitespace-only

**Example:**

```javascript
ws.send(JSON.stringify({
  session_id: '550e8400-e29b-41d4-a716-446655440000',
  content: 'What is Amazon Bedrock?'
}));
```

---

### Server Messages

The server streams multiple message types during a conversation.

#### Content Chunk

Streaming response text from the agent.

**Format:**

```json
{
  "type": "content",
  "content": "Amazon Bedrock is a fully managed service..."
}
```

**Fields:**

| Field | Type | Description |
|-------|------|-------------|
| type | string | Always "content" |
| content | string | Text chunk from agent response |

**Notes:**
- Multiple content messages may be sent for a single response
- Content should be appended to build the complete response
- Content is streamed in real-time as generated

---

#### Citation

Knowledge base citation reference.

**Format:**

```json
{
  "type": "citation",
  "citation": {
    "source_id": "doc-123",
    "source_name": "AWS Documentation",
    "excerpt": "Bedrock provides access to foundation models...",
    "confidence": 0.95,
    "url": "https://docs.aws.amazon.com/bedrock/",
    "metadata": {
      "page": 5,
      "section": "Overview"
    }
  }
}
```

**Fields:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| type | string | Yes | Always "citation" |
| citation.source_id | string | Yes | Unique source identifier |
| citation.source_name | string | Yes | Human-readable source name |
| citation.excerpt | string | Yes | Relevant text excerpt |
| citation.confidence | number | No | Confidence score (0.0-1.0) |
| citation.url | string | No | Source URL |
| citation.metadata | object | No | Additional metadata |

**Notes:**
- Citations may arrive during or after content streaming
- Multiple citations may be sent for a single response
- Citations should be associated with the current message

---

#### Completion

Indicates the response stream has completed successfully.

**Format:**

```json
{
  "type": "done"
}
```

**Fields:**

| Field | Type | Description |
|-------|------|-------------|
| type | string | Always "done" |

**Notes:**
- Sent after all content and citations have been streamed
- Indicates the agent has finished responding
- Client should re-enable input after receiving this message

---

#### Error

Indicates an error occurred during processing.

**Format:**

```json
{
  "type": "error",
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "Service is temporarily busy. Please try again in 30 seconds.",
    "retryable": true,
    "details": {
      "retry_after": 30
    }
  }
}
```

**Fields:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| type | string | Yes | Always "error" |
| error.code | string | Yes | Error code (see Error Codes) |
| error.message | string | Yes | User-friendly error message |
| error.retryable | boolean | Yes | Whether the request can be retried |
| error.details | object | No | Additional error context |

**Notes:**
- Error messages are sanitized (no internal details)
- Check `retryable` field to determine if retry is appropriate
- Connection may remain open after error (depends on error type)

---

## Error Codes

### Client Errors (4xx)

| Code | HTTP Status | Description | Retryable |
|------|-------------|-------------|-----------|
| INVALID_REQUEST | 400 | Invalid request parameters | No |
| SESSION_NOT_FOUND | 404 | Session does not exist | No |
| INVALID_SESSION_ID | 400 | Session ID format is invalid | No |
| INVALID_MESSAGE_CONTENT | 400 | Message content is invalid | No |
| MESSAGE_TOO_LONG | 400 | Message exceeds maximum length | No |
| EMPTY_MESSAGE | 400 | Message is empty or whitespace-only | No |

### Server Errors (5xx)

| Code | HTTP Status | Description | Retryable |
|------|-------------|-------------|-----------|
| SESSION_CREATE_FAILED | 500 | Failed to create session | Yes |
| PROCESSING_FAILED | 500 | Failed to process message | Yes |
| INTERNAL_ERROR | 500 | Internal server error | Yes |

### Bedrock Errors

| Code | Description | Retryable |
|------|-------------|-----------|
| RATE_LIMIT_EXCEEDED | Bedrock rate limit hit | Yes |
| SERVICE_ERROR | Bedrock service error | Maybe |
| NETWORK_ERROR | Network connectivity issue | Yes |
| TIMEOUT | Request timed out | Yes |
| MALFORMED_STREAM | Stream parsing error | No |
| UNAUTHORIZED | Authentication/authorization failed | No |
| INVALID_INPUT | Invalid input to Bedrock | No |

---

## Examples

### Complete Chat Flow

```javascript
// 1. Create a session
const sessionResponse = await fetch('http://localhost:8080/api/sessions', {
  method: 'POST'
});
const session = await sessionResponse.json();
console.log('Session created:', session.id);

// 2. Connect to WebSocket
const ws = new WebSocket('ws://localhost:8080/api/chat/stream');

let currentMessage = '';
let citations = [];

ws.onopen = () => {
  console.log('WebSocket connected');
  
  // 3. Send a message
  ws.send(JSON.stringify({
    session_id: session.id,
    content: 'What is Amazon Bedrock?'
  }));
};

ws.onmessage = (event) => {
  const message = JSON.parse(event.data);
  
  switch (message.type) {
    case 'content':
      // Append content chunk
      currentMessage += message.content;
      console.log('Current message:', currentMessage);
      break;
      
    case 'citation':
      // Store citation
      citations.push(message.citation);
      console.log('Citation received:', message.citation.source_name);
      break;
      
    case 'done':
      // Response complete
      console.log('Final message:', currentMessage);
      console.log('Citations:', citations);
      
      // Reset for next message
      currentMessage = '';
      citations = [];
      break;
      
    case 'error':
      // Handle error
      console.error('Error:', message.error.message);
      
      if (message.error.retryable) {
        console.log('Error is retryable');
        // Implement retry logic
      }
      break;
  }
};

ws.onerror = (error) => {
  console.error('WebSocket error:', error);
};

ws.onclose = () => {
  console.log('WebSocket disconnected');
};

// 4. Send another message (after receiving 'done')
setTimeout(() => {
  ws.send(JSON.stringify({
    session_id: session.id,
    content: 'Tell me more about its features'
  }));
}, 5000);

// 5. Close connection when done
setTimeout(() => {
  ws.close();
}, 30000);
```

### Error Handling

```javascript
ws.onmessage = (event) => {
  const message = JSON.parse(event.data);
  
  if (message.type === 'error') {
    const { code, message: errorMessage, retryable, details } = message.error;
    
    switch (code) {
      case 'RATE_LIMIT_EXCEEDED':
        const retryAfter = details?.retry_after || 30;
        console.log(`Rate limited. Retry after ${retryAfter} seconds`);
        setTimeout(() => retryMessage(), retryAfter * 1000);
        break;
        
      case 'SESSION_NOT_FOUND':
        console.error('Session expired. Creating new session...');
        createNewSession();
        break;
        
      case 'NETWORK_ERROR':
        if (retryable) {
          console.log('Network error. Retrying...');
          retryWithBackoff();
        }
        break;
        
      default:
        console.error(`Error: ${errorMessage}`);
        if (retryable) {
          console.log('Retrying...');
          retryMessage();
        }
    }
  }
};
```

### Reconnection Logic

```javascript
let reconnectAttempts = 0;
const maxReconnectAttempts = 5;

function connect() {
  const ws = new WebSocket('ws://localhost:8080/api/chat/stream');
  
  ws.onopen = () => {
    console.log('Connected');
    reconnectAttempts = 0;
  };
  
  ws.onclose = () => {
    console.log('Disconnected');
    
    if (reconnectAttempts < maxReconnectAttempts) {
      const backoff = Math.min(1000 * Math.pow(2, reconnectAttempts), 30000);
      console.log(`Reconnecting in ${backoff}ms...`);
      
      setTimeout(() => {
        reconnectAttempts++;
        connect();
      }, backoff);
    } else {
      console.error('Max reconnection attempts reached');
    }
  };
  
  return ws;
}

const ws = connect();
```

### Using with TypeScript

```typescript
interface SessionResponse {
  id: string;
  created_at: string;
  message_count: number;
}

interface ClientMessage {
  session_id: string;
  content: string;
}

interface ServerMessage {
  type: 'content' | 'citation' | 'done' | 'error';
  content?: string;
  citation?: Citation;
  error?: ChatError;
}

interface Citation {
  source_id: string;
  source_name: string;
  excerpt: string;
  confidence?: number;
  url?: string;
  metadata?: Record<string, any>;
}

interface ChatError {
  code: string;
  message: string;
  retryable: boolean;
  details?: Record<string, any>;
}

// Create session
const response = await fetch('http://localhost:8080/api/sessions', {
  method: 'POST'
});
const session: SessionResponse = await response.json();

// Connect WebSocket
const ws = new WebSocket('ws://localhost:8080/api/chat/stream');

ws.onmessage = (event: MessageEvent) => {
  const message: ServerMessage = JSON.parse(event.data);
  
  switch (message.type) {
    case 'content':
      console.log(message.content);
      break;
    case 'citation':
      console.log(message.citation);
      break;
    case 'done':
      console.log('Complete');
      break;
    case 'error':
      console.error(message.error);
      break;
  }
};

// Send message
const clientMessage: ClientMessage = {
  session_id: session.id,
  content: 'Hello'
};
ws.send(JSON.stringify(clientMessage));
```

---

## Testing

### Using curl

```bash
# Health check
curl http://localhost:8080/health

# Create session
SESSION_ID=$(curl -s -X POST http://localhost:8080/api/sessions | jq -r .id)
echo "Session ID: $SESSION_ID"

# Get session
curl http://localhost:8080/api/sessions/$SESSION_ID | jq .

# List sessions
curl http://localhost:8080/api/sessions | jq .
```

### Using WebSocket Test Client

```bash
cd backend

# Build test client
go build -o bin/wsclient cmd/wsclient/main.go

# Send message
./bin/wsclient -session $SESSION_ID -message "What is Amazon Bedrock?"
```

### Using Postman

1. Import the API collection (if available)
2. Set base URL variable: `http://localhost:8080`
3. Create a session using POST request
4. Use WebSocket request to test streaming

---

## Changelog

### Version 1.0.0 (Current)

- Initial API release
- Session management endpoints
- WebSocket streaming support
- Bedrock Agent Core integration
- Knowledge base citations

### Future Versions

- Authentication and authorization
- Rate limiting
- Pagination for session list
- Message history endpoints
- User preferences
- Analytics endpoints

---

## Support

For API issues or questions:

1. Check this documentation
2. Review [Configuration Guide](CONFIGURATION.md)
3. Check [Troubleshooting](../README.md#troubleshooting)
4. Create a GitHub issue

---

## Related Documentation

- [Main README](../../README.md)
- [Configuration Guide](CONFIGURATION.md)
- [Bedrock Adapter](../infrastructure/bedrock/README.md)
- [Docker Setup](../../DOCKER.md)
