# API Documentation

## Chat API Reference

### Session Management

#### Create Session
```http
POST /api/sessions
Content-Type: application/json

{
  "user_id": "string"
}
```

**Response:**
```json
{
  "id": "uuid",
  "created_at": "timestamp",
  "message_count": 0
}
```

#### Get Session
```http
GET /api/sessions/{session_id}
```

**Response:**
```json
{
  "id": "uuid",
  "user_id": "string",
  "created_at": "timestamp",
  "message_count": number
}
```

### WebSocket Chat

#### Connection
```
WebSocket: ws://localhost:8080/api/chat
```

#### Send Message
```json
{
  "session_id": "uuid",
  "content": "Your message here",
  "timestamp": "ISO8601"
}
```

#### Receive Response
```json
{
  "type": "chunk",
  "content": "Response chunk",
  "session_id": "uuid",
  "timestamp": "ISO8601"
}
```

## Error Handling

### Error Response Format
```json
{
  "type": "error",
  "code": "ERROR_CODE",
  "message": "Human readable error message",
  "session_id": "uuid",
  "timestamp": "ISO8601"
}
```

### Common Error Codes
- `INVALID_INPUT` - Invalid request parameters
- `SESSION_NOT_FOUND` - Session does not exist
- `RATE_LIMIT_EXCEEDED` - Too many requests
- `SERVICE_ERROR` - Internal service error
- `TIMEOUT` - Request timeout
- `UNAUTHORIZED` - Authentication failed

## Rate Limits
- Maximum 100 requests per minute per session
- Maximum message length: 25,000 characters
- Session timeout: 30 minutes of inactivity