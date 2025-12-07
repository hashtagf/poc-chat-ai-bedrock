# Bedrock Agent Core Chat UI - POC

A proof-of-concept chat interface for Amazon Bedrock Agent Core with knowledge base integration. This application demonstrates conversational AI capabilities with real-time streaming responses and knowledge base citations.

## Table of Contents

- [Features](#features)
- [Architecture](#architecture)
- [Tech Stack](#tech-stack)
- [Quick Start](#quick-start)
- [Project Structure](#project-structure)
- [Configuration](#configuration)
- [API Documentation](#api-documentation)
- [Component Documentation](#component-documentation)
- [Testing](#testing)
- [Development](#development)
- [Deployment](#deployment)
- [Troubleshooting](#troubleshooting)
- [Contributing](#contributing)

## Features

- **Real-time Streaming**: Incremental display of AI responses as they're generated
- **Knowledge Base Integration**: Citations from Bedrock knowledge bases with source information
- **Session Management**: Multiple conversation sessions with isolated histories
- **Error Handling**: Comprehensive error handling with user-friendly messages and retry logic
- **Accessibility**: WCAG AA compliant with keyboard navigation and screen reader support
- **Responsive Design**: Works across desktop, tablet, and mobile devices
- **Property-Based Testing**: Comprehensive test coverage with fast-check

## Architecture

This project follows **Hexagonal Architecture** (Ports & Adapters) principles:

```
┌─────────────────────────────────────────────────────────────┐
│                    Frontend (Vue 3)                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │   Message    │  │ Conversation │  │   Citation   │     │
│  │   Input      │  │   Display    │  │   Display    │     │
│  └──────────────┘  └──────────────┘  └──────────────┘     │
└────────────────────────────┬──────────────────────────────┘
                             │ HTTP/WebSocket
                    ┌────────▼────────┐
                    │   Backend (Go)  │
                    │  - API Handler  │
                    │  - Session Mgmt │
                    └────────┬────────┘
                             │
                    ┌────────▼────────┐
                    │  Bedrock Agent  │
                    │      Core       │
                    └─────────────────┘
```

### Architectural Layers

**Frontend (Vue 3)**
- **Components**: UI presentation layer
- **Composables**: Business logic and state management
- **Types**: TypeScript interfaces and type definitions

**Backend (Go)**
- **Domain**: Core business entities and repository interfaces
- **Infrastructure**: External adapters (Bedrock, repositories)
- **Interfaces**: HTTP/WebSocket handlers

**Dependencies Flow**: UI → Composables → Backend API → Bedrock Agent Core

## Tech Stack

- **Frontend**: Vue 3 (Composition API) + TypeScript + Tailwind CSS + Vite
- **Backend**: Go 1.21+ with standard library and AWS SDK v2
- **AI/ML**: Amazon Bedrock Agent Core with Knowledge Base integration
- **Testing**: Vitest (frontend), Go testing (backend), fast-check (property-based testing)
- **Infrastructure**: Docker, Docker Compose
- **Communication**: WebSocket for real-time streaming, REST for session management

## Quick Start

### Prerequisites

- **Go**: 1.21 or higher ([Download](https://go.dev/dl/))
- **Node.js**: 18 or higher ([Download](https://nodejs.org/))
- **Docker & Docker Compose**: Latest version (optional, for containerized deployment)
- **AWS Account**: With Bedrock access configured
- **AWS CLI**: Configured with credentials (optional, for local development)

## Project Structure

```
.
├── frontend/                      # Vue 3 frontend application
│   ├── src/
│   │   ├── components/           # Vue components
│   │   │   ├── ChatContainer.vue # Root chat component
│   │   │   ├── MessageInput.vue  # User input component
│   │   │   ├── MessageList.vue   # Message display
│   │   │   ├── MessageBubble.vue # Individual message
│   │   │   ├── CitationDisplay.vue # Knowledge base citations
│   │   │   └── ErrorDisplay.vue  # Error messages
│   │   ├── composables/          # Business logic
│   │   │   ├── useChatService.ts # WebSocket & messaging
│   │   │   ├── useConversationHistory.ts # Message storage
│   │   │   ├── useSessionManager.ts # Session management
│   │   │   └── useErrorHandler.ts # Error handling
│   │   ├── types/                # TypeScript types
│   │   │   └── index.ts          # Type definitions
│   │   ├── tests/                # Integration tests
│   │   ├── App.vue               # Root Vue component
│   │   └── main.ts               # Application entry
│   ├── Dockerfile                # Frontend container
│   ├── package.json              # Dependencies
│   └── vite.config.ts            # Vite configuration
├── backend/                       # Go backend application
│   ├── cmd/
│   │   ├── server/               # Main server application
│   │   │   └── main.go
│   │   └── wsclient/             # WebSocket test client
│   │       └── main.go
│   ├── config/                   # Configuration management
│   │   ├── config.go             # Config loader
│   │   ├── development.env       # Dev environment
│   │   └── production.env        # Prod environment
│   ├── domain/                   # Domain layer
│   │   ├── entities/             # Business entities
│   │   │   ├── message.go
│   │   │   └── session.go
│   │   ├── repositories/         # Repository interfaces
│   │   │   └── session_repository.go
│   │   └── services/             # Service interfaces
│   │       └── bedrock_service.go
│   ├── infrastructure/           # Infrastructure layer
│   │   ├── bedrock/              # Bedrock adapter
│   │   │   ├── adapter.go        # AWS SDK integration
│   │   │   ├── stream_processor.go # Stream handling
│   │   │   └── stream_reader.go  # Stream reader
│   │   └── repositories/         # Repository implementations
│   │       └── memory_session_repository.go
│   ├── interfaces/               # Interface layer
│   │   └── chat/                 # Chat handlers
│   │       ├── handler.go        # HTTP/WebSocket handlers
│   │       └── dto.go            # Data transfer objects
│   ├── docs/                     # Documentation
│   │   └── CONFIGURATION.md      # Configuration guide
│   ├── Dockerfile                # Backend container
│   ├── Makefile                  # Build commands
│   └── go.mod                    # Go dependencies
├── .kiro/                        # Kiro specs and steering
│   ├── specs/chat-ui/            # Feature specifications
│   │   ├── requirements.md       # Requirements document
│   │   ├── design.md             # Design document
│   │   └── tasks.md              # Implementation tasks
│   └── steering/                 # Development guidelines
│       ├── product.md            # Product context
│       ├── structure.md          # Architecture guidelines
│       └── tech.md               # Technology standards
├── docker-compose.yml            # Full stack orchestration
├── .env.example                  # Environment template
├── README.md                     # This file
└── DOCKER.md                     # Docker setup guide
```

### Option 1: Docker Compose (Recommended)

The easiest way to run the full stack:

```bash
# 1. Clone the repository
git clone <repository-url>
cd bedrock-chat-poc

# 2. Copy and configure environment variables
cp .env.example .env
# Edit .env with your AWS credentials and Bedrock configuration

# 3. Build and start all services
docker-compose up --build

# Access the application:
# - Frontend: http://localhost:5173
# - Backend API: http://localhost:8080
# - Health check: http://localhost:8080/health
```

For detailed Docker instructions, see [DOCKER.md](DOCKER.md).

### Option 2: Local Development

#### Step 1: Configure Environment

```bash
# Copy environment template
cp .env.example .env

# Edit .env with your configuration
# Minimum required for development:
# - AWS_REGION=us-east-1
# - BEDROCK_AGENT_ID=your_agent_id (optional for mock mode)
# - BEDROCK_AGENT_ALIAS_ID=your_alias_id (optional for mock mode)
```

#### Step 2: Start Backend

```bash
cd backend

# Install dependencies
go mod download

# Run tests (optional)
go test ./... -v

# Start server
go run cmd/server/main.go

# Server runs on http://localhost:8080
# Health check: curl http://localhost:8080/health
```

#### Step 3: Start Frontend

```bash
# In a new terminal
cd frontend

# Install dependencies
npm install

# Run tests (optional)
npm test

# Start development server
npm run dev

# Frontend runs on http://localhost:5173
```

#### Step 4: Verify Setup

Open your browser to `http://localhost:5173` and start chatting!

### Mock Mode (No AWS Required)

For development without AWS Bedrock:

```bash
# In .env, leave Bedrock IDs empty:
ENVIRONMENT=development
AWS_REGION=us-east-1
# BEDROCK_AGENT_ID=  (commented out or empty)
# BEDROCK_AGENT_ALIAS_ID=  (commented out or empty)

# The backend will run in mock mode with simulated responses
```

## API Documentation

### Base URL

- **Development**: `http://localhost:8080`
- **Production**: Configure via `SERVER_HOST` and `SERVER_PORT` environment variables

### Health Check

```bash
GET /health

Response: 200 OK
{
  "status": "healthy",
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### Session Management

#### Create Session

Creates a new conversation session with a unique identifier.

```bash
POST /api/sessions

Response: 201 Created
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "created_at": "2024-01-01T00:00:00Z",
  "message_count": 0
}
```

#### Get Session

Retrieves details for a specific session.

```bash
GET /api/sessions/{id}

Response: 200 OK
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "created_at": "2024-01-01T00:00:00Z",
  "last_message_at": "2024-01-01T00:01:00Z",
  "message_count": 5
}

Error: 404 Not Found
{
  "code": "SESSION_NOT_FOUND",
  "message": "Session not found"
}
```

#### List Sessions

Returns all active sessions.

```bash
GET /api/sessions

Response: 200 OK
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

### Chat Streaming (WebSocket)

#### Connect to WebSocket

```bash
ws://localhost:8080/api/chat/stream
```

#### Client Message Format

```json
{
  "session_id": "550e8400-e29b-41d4-a716-446655440000",
  "content": "What is Amazon Bedrock?"
}
```

**Validation Rules:**
- `session_id`: Required, must be a valid UUID
- `content`: Required, 1-2000 characters, cannot be whitespace-only

#### Server Response Format

The server streams multiple message types:

**Content Chunk** (streaming response text)
```json
{
  "type": "content",
  "content": "Amazon Bedrock is a fully managed service..."
}
```

**Citation** (knowledge base reference)
```json
{
  "type": "citation",
  "citation": {
    "source_id": "doc-123",
    "source_name": "AWS Documentation",
    "excerpt": "Bedrock provides access to foundation models...",
    "confidence": 0.95,
    "url": "https://docs.aws.amazon.com/bedrock/"
  }
}
```

**Completion** (stream finished successfully)
```json
{
  "type": "done"
}
```

**Error** (stream encountered an error)
```json
{
  "type": "error",
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "Service is temporarily busy. Please try again in 30 seconds.",
    "retryable": true
  }
}
```

#### Error Codes

| Code | Description | Retryable |
|------|-------------|-----------|
| `INVALID_REQUEST` | Invalid request parameters | No |
| `SESSION_NOT_FOUND` | Session does not exist | No |
| `RATE_LIMIT_EXCEEDED` | Bedrock rate limit hit | Yes |
| `SERVICE_ERROR` | Bedrock service error | Maybe |
| `NETWORK_ERROR` | Network connectivity issue | Yes |
| `TIMEOUT` | Request timed out | Yes |
| `MALFORMED_STREAM` | Stream parsing error | No |

### Example Usage

#### Using curl

```bash
# Create a session
SESSION_ID=$(curl -s -X POST http://localhost:8080/api/sessions | jq -r .id)
echo "Session ID: $SESSION_ID"

# Get session details
curl http://localhost:8080/api/sessions/$SESSION_ID | jq .

# List all sessions
curl http://localhost:8080/api/sessions | jq .
```

#### Using WebSocket Test Client

```bash
cd backend

# Build the test client
go build -o bin/wsclient cmd/wsclient/main.go

# Send a message
./bin/wsclient -session $SESSION_ID -message "Hello, Bedrock!"
```

#### Using JavaScript

```javascript
// Create session
const response = await fetch('http://localhost:8080/api/sessions', {
  method: 'POST'
});
const session = await response.json();

// Connect to WebSocket
const ws = new WebSocket('ws://localhost:8080/api/chat/stream');

ws.onopen = () => {
  ws.send(JSON.stringify({
    session_id: session.id,
    content: 'What is Amazon Bedrock?'
  }));
};

ws.onmessage = (event) => {
  const message = JSON.parse(event.data);
  
  switch (message.type) {
    case 'content':
      console.log('Content:', message.content);
      break;
    case 'citation':
      console.log('Citation:', message.citation);
      break;
    case 'done':
      console.log('Stream complete');
      ws.close();
      break;
    case 'error':
      console.error('Error:', message.error);
      break;
  }
};
```

## Component Documentation

### Frontend Components

#### ChatContainer

Root component that orchestrates all chat functionality.

**Props:**
- `sessionId?: string` - Optional initial session ID

**Events:**
- `@session-created` - Emitted when a new session is created
- `@error` - Emitted when an error occurs

**Usage:**
```vue
<ChatContainer 
  :session-id="currentSessionId"
  @session-created="handleSessionCreated"
  @error="handleError"
/>
```

#### MessageInput

User input component with validation and submission handling.

**Props:**
- `disabled: boolean` - Whether input is disabled (during streaming)
- `placeholder?: string` - Input placeholder text (default: "Type a message...")

**Events:**
- `@submit` - Emitted when user submits a message with `content: string`

**Features:**
- Enter key submission
- Automatic focus management
- Input validation (1-2000 characters, no whitespace-only)
- Loading indicator during submission

**Usage:**
```vue
<MessageInput 
  :disabled="isStreaming"
  placeholder="Ask me anything..."
  @submit="handleSubmit"
/>
```

#### MessageList

Scrollable container for displaying conversation history.

**Props:**
- `messages: Message[]` - Array of messages to display
- `isStreaming: boolean` - Whether a response is currently streaming
- `streamingContent?: string` - Current streaming message content

**Features:**
- Auto-scroll to latest message
- Maintains scroll position when user scrolls up
- Virtual scrolling for large conversations (>100 messages)
- Chronological ordering

**Usage:**
```vue
<MessageList 
  :messages="conversationHistory"
  :is-streaming="isStreaming"
  :streaming-content="currentStreamContent"
/>
```

#### MessageBubble

Individual message display with role-based styling.

**Props:**
- `message: Message` - Message object to display
- `role: 'user' | 'agent'` - Message sender role
- `content: string` - Message text content
- `timestamp: Date` - Message timestamp
- `citations?: Citation[]` - Optional knowledge base citations
- `status: 'sending' | 'sent' | 'error'` - Message status

**Features:**
- Role-based styling (user: right-aligned blue, agent: left-aligned gray)
- Timestamp formatting
- Status indicators
- Citation display integration
- Semantic HTML with ARIA labels

**Usage:**
```vue
<MessageBubble 
  :message="message"
  :role="message.role"
  :content="message.content"
  :timestamp="message.timestamp"
  :citations="message.citations"
  :status="message.status"
/>
```

#### CitationDisplay

Displays knowledge base citations with expandable details.

**Props:**
- `citations: Citation[]` - Array of citations to display
- `expanded?: boolean` - Whether citations are initially expanded

**Features:**
- Click/hover to reveal citation details
- Source name, excerpt, and confidence score display
- Visual distinction for multiple citations
- "No citations" indicator for general knowledge responses

**Usage:**
```vue
<CitationDisplay 
  :citations="message.citations"
  :expanded="false"
/>
```

#### ErrorDisplay

User-friendly error message display with retry functionality.

**Props:**
- `error: ChatError | null` - Error object to display
- `connectionStatus: 'connected' | 'disconnected' | 'reconnecting'` - Connection status

**Events:**
- `@retry` - Emitted when user clicks retry button
- `@dismiss` - Emitted when user dismisses error

**Features:**
- Auto-dismiss for non-critical errors (5 seconds)
- Retry button for retryable errors
- Connection status indicator
- Sanitized error messages

**Usage:**
```vue
<ErrorDisplay 
  :error="currentError"
  :connection-status="wsConnectionStatus"
  @retry="handleRetry"
  @dismiss="clearError"
/>
```

### Composables (Business Logic)

#### useChatService

Manages WebSocket connection and message transmission.

**Returns:**
```typescript
{
  sendMessage: (content: string) => Promise<void>
  streamingMessage: Ref<string>
  isStreaming: Ref<boolean>
  error: Ref<ChatError | null>
  clearError: () => void
  connectionStatus: Ref<'connected' | 'disconnected' | 'reconnecting'>
}
```

**Features:**
- WebSocket connection management
- Automatic reconnection with exponential backoff
- Message validation and transmission
- Streaming response processing
- Error handling and transformation

#### useConversationHistory

Manages message storage and retrieval.

**Returns:**
```typescript
{
  messages: Ref<Message[]>
  addMessage: (message: Message) => void
  clearHistory: () => void
  getMessageById: (id: string) => Message | undefined
}
```

**Features:**
- Reactive message array
- Chronological ordering
- Message lookup by ID
- History clearing

#### useSessionManager

Controls session lifecycle and identification.

**Returns:**
```typescript
{
  currentSessionId: Ref<string>
  createNewSession: () => Promise<string>
  loadSession: (sessionId: string) => Promise<void>
  sessionMetadata: Ref<SessionMetadata>
}
```

**Features:**
- Session creation with UUID generation
- Session switching
- Metadata tracking (created time, message count)
- Session isolation

#### useErrorHandler

Centralizes error handling and user messaging.

**Returns:**
```typescript
{
  error: Ref<ChatError | null>
  setError: (error: Error | ChatError) => void
  clearError: () => void
  transformError: (error: any) => ChatError
}
```

**Features:**
- Error sanitization (removes internal details)
- Error aggregation (multiple errors within time window)
- Retry logic with exponential backoff
- User-friendly error messages

### Type Definitions

#### Message
```typescript
interface Message {
  id: string                    // UUID v4
  role: 'user' | 'agent'
  content: string
  timestamp: Date
  citations?: Citation[]
  status: 'sending' | 'sent' | 'error'
  errorMessage?: string
}
```

#### Citation
```typescript
interface Citation {
  sourceId: string
  sourceName: string
  excerpt: string
  confidence?: number          // 0.0 to 1.0
  url?: string
  metadata?: Record<string, any>
}
```

#### ChatError
```typescript
interface ChatError {
  code: string
  message: string
  retryable: boolean
  details?: Record<string, any>
}
```

#### Session
```typescript
interface Session {
  id: string
  createdAt: Date
  lastMessageAt?: Date
  messageCount: number
}
```

### Backend Tests

The backend uses Go's built-in testing framework with table-driven tests.

```bash
cd backend

# Run all tests
go test ./... -v

# Run with coverage
go test ./... -cover

# Run with coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Run specific package
go test ./interfaces/chat -v

# Run specific test
go test ./interfaces/chat -run TestHandleCreateSession -v

# Run with race detection
go test ./... -race

# Using Makefile
make test              # Run all tests
make test-coverage     # Run with coverage
```

**Test Structure:**
- Unit tests: `*_test.go` files alongside source code
- Integration tests: `*_integration_test.go` files
- Table-driven tests for comprehensive coverage

**Example:**
```go
func TestMessageValidation(t *testing.T) {
    tests := []struct {
        name    string
        content string
        wantErr bool
    }{
        {"valid message", "Hello", false},
        {"empty message", "", true},
        {"whitespace only", "   ", true},
        {"too long", strings.Repeat("a", 2001), true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validateMessage(tt.content)
            if (err != nil) != tt.wantErr {
                t.Errorf("validateMessage() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### Frontend Tests

The frontend uses **Vitest** for unit testing and **fast-check** for property-based testing.

```bash
cd frontend

# Run all tests once
npm test

# Run in watch mode
npm run test:watch

# Run with coverage
npm run test:coverage

# Run specific test file
npm test -- MessageInput.test.ts

# Run property-based tests only
npm test -- --grep "Property"

# Run with UI
npm run test:ui
```

**Test Types:**

1. **Unit Tests** - Test individual components and composables
   - Located alongside source files: `*.test.ts`
   - Use Vue Test Utils for component testing
   - Mock external dependencies

2. **Property-Based Tests** - Test universal properties across many inputs
   - Use fast-check library
   - Run 100+ iterations per property
   - Tagged with property number from design doc

3. **Integration Tests** - Test component interactions
   - Located in `src/tests/`
   - Test WebSocket communication
   - Test error scenarios end-to-end

**Example Unit Test:**
```typescript
import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import MessageInput from './MessageInput.vue'

describe('MessageInput', () => {
  it('should prevent empty message submission', async () => {
    const wrapper = mount(MessageInput)
    const button = wrapper.find('button')
    
    expect(button.attributes('disabled')).toBeDefined()
  })
})
```

**Example Property Test:**
```typescript
import { test } from 'vitest'
import fc from 'fast-check'

// Feature: chat-ui, Property 2: Input validation prevents invalid submission
test('Property 2: whitespace-only messages are rejected', () => {
  fc.assert(
    fc.property(
      fc.stringOf(fc.constantFrom(' ', '\t', '\n', '\r')),
      (whitespace) => {
        const result = validateMessage(whitespace)
        return result === false
      }
    ),
    { numRuns: 100 }
  )
})
```

### Running All Tests

```bash
# Backend tests
cd backend && go test ./... -v

# Frontend tests
cd frontend && npm test

# Or use Docker
docker-compose run backend go test ./... -v
docker-compose run frontend npm test
```

### Test Coverage Goals

- **Backend**: >80% coverage for domain and infrastructure layers
- **Frontend**: >80% coverage for composables and components
- **Property Tests**: All 31 correctness properties implemented
- **Integration Tests**: All critical user flows covered

## Development

### Development Workflow

1. **Create a feature branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make changes and test**
   ```bash
   # Backend
   cd backend
   go test ./... -v
   
   # Frontend
   cd frontend
   npm test
   ```

3. **Run linters**
   ```bash
   # Backend
   cd backend
   gofmt -w .
   go vet ./...
   
   # Frontend
   cd frontend
   npm run lint
   ```

4. **Commit and push**
   ```bash
   git add .
   git commit -m "feat: add your feature"
   git push origin feature/your-feature-name
   ```

### Backend Development

The backend follows Go best practices and hexagonal architecture:

**Key Principles:**
- Explicit error handling with context
- Dependency injection via interfaces
- Table-driven tests for comprehensive coverage
- Context for cancellation/timeouts
- Domain-driven design

**Code Structure:**
```go
// Domain layer - pure business logic
type Session struct {
    ID        string
    CreatedAt time.Time
}

// Repository interface (port)
type SessionRepository interface {
    Create(ctx context.Context, session *Session) error
    Get(ctx context.Context, id string) (*Session, error)
}

// Handler (adapter)
func (h *Handler) HandleCreateSession(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    session := &entities.Session{
        ID:        uuid.New().String(),
        CreatedAt: time.Now(),
    }
    
    if err := h.sessionRepo.Create(ctx, session); err != nil {
        h.writeError(w, http.StatusInternalServerError, 
            "SESSION_CREATE_FAILED", "Failed to create session")
        return
    }
    
    h.writeJSON(w, http.StatusCreated, session)
}
```

**Adding a New Endpoint:**

1. Define domain entity in `domain/entities/`
2. Define repository interface in `domain/repositories/`
3. Implement repository in `infrastructure/repositories/`
4. Create handler in `interfaces/chat/`
5. Register route in `cmd/server/main.go`
6. Write tests for each layer

**Useful Commands:**
```bash
# Format code
gofmt -w .

# Run linter
go vet ./...

# Check for common mistakes
golint ./...

# Run tests with race detection
go test ./... -race

# Generate coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Frontend Development

The frontend uses Vue 3 Composition API with TypeScript and Tailwind CSS:

**Key Principles:**
- Composition API for reusable logic
- TypeScript for type safety
- Composables for business logic
- Components for presentation
- Tailwind utilities for styling

**Code Structure:**
```typescript
// Composable - business logic
export function useChatService() {
  const isStreaming = ref(false)
  const streamingMessage = ref('')
  const error = ref<ChatError | null>(null)
  
  const sendMessage = async (content: string) => {
    if (!content.trim()) {
      throw new Error('Message cannot be empty')
    }
    
    isStreaming.value = true
    // Send message via WebSocket
  }
  
  return {
    isStreaming,
    streamingMessage,
    error,
    sendMessage
  }
}

// Component - presentation
<script setup lang="ts">
import { useChatService } from '@/composables/useChatService'

const { isStreaming, sendMessage } = useChatService()

const handleSubmit = async (content: string) => {
  await sendMessage(content)
}
</script>

<template>
  <div class="flex flex-col h-screen">
    <MessageList :is-streaming="isStreaming" />
    <MessageInput @submit="handleSubmit" />
  </div>
</template>
```

**Adding a New Component:**

1. Create component file in `src/components/`
2. Define props and events with TypeScript
3. Extract business logic to composable if needed
4. Style with Tailwind utilities
5. Write unit tests in `*.test.ts`
6. Write property tests if applicable

**Useful Commands:**
```bash
# Type check
npm run type-check

# Lint
npm run lint

# Format
npm run format

# Build
npm run build

# Preview build
npm run preview
```

### Hot Reload

Both frontend and backend support hot reload during development:

**Backend:**
```bash
# Install air for hot reload
go install github.com/cosmtrek/air@latest

# Run with hot reload
cd backend
air
```

**Frontend:**
```bash
# Vite has built-in hot reload
cd frontend
npm run dev
```

### Debugging

**Backend:**
```bash
# Use delve debugger
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug server
cd backend
dlv debug cmd/server/main.go

# Set breakpoint
(dlv) break main.main
(dlv) continue
```

**Frontend:**
```bash
# Use browser DevTools
# Open browser console (F12)
# Set breakpoints in Sources tab

# Or use VS Code debugger
# Add launch configuration in .vscode/launch.json
```

### Code Style

**Backend (Go):**
- Follow [Effective Go](https://go.dev/doc/effective_go)
- Use `gofmt` for formatting
- Keep functions small and focused
- Document exported functions and types
- Use meaningful variable names

**Frontend (TypeScript/Vue):**
- Follow [Vue Style Guide](https://vuejs.org/style-guide/)
- Use TypeScript strict mode
- Prefer composition over inheritance
- Keep components small and focused
- Use Tailwind utilities, avoid custom CSS

## Deployment

### Production Checklist

Before deploying to production:

- [ ] Set `ENVIRONMENT=production` in environment variables
- [ ] Configure Bedrock Agent ID and Alias ID
- [ ] Use IAM roles for AWS credentials (never hardcode keys)
- [ ] Set up proper CORS restrictions in backend
- [ ] Enable HTTPS/TLS for all connections
- [ ] Configure proper logging (JSON format)
- [ ] Set up monitoring and alerting
- [ ] Implement rate limiting
- [ ] Add authentication and authorization
- [ ] Use persistent session storage (not in-memory)
- [ ] Configure proper timeouts and buffer sizes
- [ ] Set up backup and disaster recovery
- [ ] Review and test error handling
- [ ] Perform security audit
- [ ] Load test the application

### Docker Deployment

**Build production images:**
```bash
# Build backend
docker build -t chat-backend:latest ./backend

# Build frontend
docker build -t chat-frontend:latest ./frontend

# Or use docker-compose
docker-compose -f docker-compose.prod.yml build
```

**Run in production:**
```bash
# Set production environment variables
export ENVIRONMENT=production
export BEDROCK_AGENT_ID=your_agent_id
export BEDROCK_AGENT_ALIAS_ID=your_alias_id

# Start services
docker-compose -f docker-compose.prod.yml up -d

# Check health
curl http://localhost:8080/health
```

### AWS Deployment Options

#### Option 1: ECS (Elastic Container Service)

1. **Push images to ECR**
   ```bash
   # Authenticate to ECR
   aws ecr get-login-password --region us-east-1 | \
     docker login --username AWS --password-stdin <account-id>.dkr.ecr.us-east-1.amazonaws.com
   
   # Tag and push
   docker tag chat-backend:latest <account-id>.dkr.ecr.us-east-1.amazonaws.com/chat-backend:latest
   docker push <account-id>.dkr.ecr.us-east-1.amazonaws.com/chat-backend:latest
   ```

2. **Create ECS task definition**
   - Define container configurations
   - Set environment variables
   - Configure IAM role with Bedrock permissions
   - Set resource limits (CPU, memory)

3. **Create ECS service**
   - Configure load balancer
   - Set auto-scaling policies
   - Configure health checks

#### Option 2: EKS (Elastic Kubernetes Service)

1. **Create Kubernetes manifests**
   ```yaml
   # deployment.yaml
   apiVersion: apps/v1
   kind: Deployment
   metadata:
     name: chat-backend
   spec:
     replicas: 3
     template:
       spec:
         containers:
         - name: backend
           image: <account-id>.dkr.ecr.us-east-1.amazonaws.com/chat-backend:latest
           env:
           - name: ENVIRONMENT
             value: production
   ```

2. **Deploy to EKS**
   ```bash
   kubectl apply -f k8s/
   kubectl get pods
   kubectl get services
   ```

#### Option 3: EC2 with Docker Compose

1. **Launch EC2 instance**
   - Choose appropriate instance type
   - Attach IAM role with Bedrock permissions
   - Configure security groups (ports 80, 443, 8080)

2. **Install Docker and Docker Compose**
   ```bash
   sudo yum update -y
   sudo yum install docker -y
   sudo systemctl start docker
   sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
   sudo chmod +x /usr/local/bin/docker-compose
   ```

3. **Deploy application**
   ```bash
   git clone <repository-url>
   cd bedrock-chat-poc
   cp .env.example .env
   # Edit .env with production values
   docker-compose up -d
   ```

### Environment Variables for Production

```bash
# Application
ENVIRONMENT=production
SERVER_PORT=8080
SERVER_HOST=0.0.0.0

# AWS (use IAM roles, not keys)
AWS_REGION=us-east-1

# Bedrock
BEDROCK_AGENT_ID=your_production_agent_id
BEDROCK_AGENT_ALIAS_ID=your_production_alias_id
BEDROCK_KNOWLEDGE_BASE_ID=your_kb_id
BEDROCK_MAX_RETRIES=5
BEDROCK_REQUEST_TIMEOUT=120s

# WebSocket
WS_TIMEOUT=60s
WS_STREAM_TIMEOUT=10m

# Logging
LOG_LEVEL=info
LOG_FORMAT=json

# Frontend
VITE_API_URL=https://api.yourdomain.com
VITE_WS_URL=wss://api.yourdomain.com
```

### Monitoring

**CloudWatch Logs:**
```bash
# Configure log group
aws logs create-log-group --log-group-name /aws/chat-backend

# Stream logs
aws logs tail /aws/chat-backend --follow
```

**CloudWatch Metrics:**
- Request count
- Error rate
- Response time
- WebSocket connections
- Bedrock API calls
- Token usage

**Health Checks:**
```bash
# Backend health
curl https://api.yourdomain.com/health

# Expected response
{"status":"healthy","timestamp":"2024-01-01T00:00:00Z"}
```

### Scaling

**Horizontal Scaling:**
- Use load balancer (ALB/NLB)
- Scale ECS tasks or Kubernetes pods
- Configure auto-scaling based on CPU/memory

**Vertical Scaling:**
- Increase instance size
- Adjust resource limits in container definitions

**Database Scaling:**
- Use managed database service (RDS, DynamoDB)
- Implement connection pooling
- Add read replicas for read-heavy workloads

### Security Best Practices

1. **Use HTTPS/TLS** for all connections
2. **Implement authentication** (OAuth, JWT, etc.)
3. **Use IAM roles** instead of access keys
4. **Restrict CORS** to specific origins
5. **Validate all inputs** on backend
6. **Sanitize error messages** (no internal details)
7. **Enable rate limiting** to prevent abuse
8. **Use secrets manager** for sensitive configuration
9. **Regular security audits** and dependency updates
10. **Implement proper logging** without PII

## Configuration

The application uses environment variables for configuration with sensible defaults. See [backend/docs/CONFIGURATION.md](backend/docs/CONFIGURATION.md) for comprehensive configuration documentation.

### Quick Configuration

1. Copy the example environment file:
   ```bash
   cp .env.example .env
   ```

2. Edit `.env` with your configuration:
   ```bash
   # Minimal setup (mock mode)
   ENVIRONMENT=development
   AWS_REGION=us-east-1
   
   # With Bedrock integration
   BEDROCK_AGENT_ID=your_agent_id
   BEDROCK_AGENT_ALIAS_ID=your_alias_id
   ```

### Key Configuration Options

#### Backend Environment Variables

- `ENVIRONMENT`: Application environment (development, production, test)
- `SERVER_PORT`: Server port (default: 8080)
- `AWS_REGION`: AWS region for Bedrock (default: us-east-1)
- `BEDROCK_AGENT_ID`: Bedrock Agent ID (required in production)
- `BEDROCK_AGENT_ALIAS_ID`: Bedrock Agent Alias ID (required in production)
- `BEDROCK_KNOWLEDGE_BASE_ID`: Knowledge Base ID (optional)
- `WS_TIMEOUT`: WebSocket timeout (default: 30s)
- `WS_STREAM_TIMEOUT`: Stream timeout (default: 5m)
- `SESSION_TIMEOUT`: Session timeout (default: 30m)
- `LOG_LEVEL`: Logging level (default: info)

#### Frontend Environment Variables

- `VITE_API_URL`: Backend API URL (default: http://localhost:8080)
- `VITE_WS_URL`: WebSocket URL (default: ws://localhost:8080)

For complete configuration documentation, including:
- All available options
- Environment-specific setup
- AWS credentials configuration
- Bedrock configuration
- WebSocket tuning
- Logging configuration
- Best practices and troubleshooting

See: [backend/docs/CONFIGURATION.md](backend/docs/CONFIGURATION.md)

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

### Quick Contribution Guide

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Write tests
5. Run linters and tests
6. Submit a pull request

### Code of Conduct

- Be respectful and inclusive
- Follow coding standards
- Write clear commit messages
- Document your changes
- Test thoroughly

## License

This is a POC project for demonstration purposes. See LICENSE file for details.

## Acknowledgments

- **Amazon Bedrock** for AI/ML capabilities
- **Vue.js** team for the excellent framework
- **Go** community for robust tooling
- **fast-check** for property-based testing library

## Support

For issues and questions:

1. Check the [Troubleshooting](#troubleshooting) section
2. Review existing [GitHub Issues](https://github.com/your-org/bedrock-chat-poc/issues)
3. Create a new issue with detailed information

## Roadmap

### Phase 1 (Current - POC)
- [x] Basic chat interface
- [x] WebSocket streaming
- [x] Session management
- [x] Knowledge base citations
- [x] Error handling
- [x] Property-based testing

### Phase 2 (Future)
- [ ] Authentication and authorization
- [ ] Persistent session storage (DynamoDB)
- [ ] Message history pagination
- [ ] User preferences and settings
- [ ] Multi-user support
- [ ] Conversation export
- [ ] Advanced citation features

### Phase 3 (Future)
- [ ] Multi-modal input (voice, images)
- [ ] Conversation branching
- [ ] Message reactions and feedback
- [ ] Analytics and insights
- [ ] Admin dashboard
- [ ] API rate limiting
- [ ] Caching layer

## Related Documentation

- [Docker Setup Guide](DOCKER.md) - Detailed Docker and docker-compose instructions
- [Configuration Guide](backend/docs/CONFIGURATION.md) - Comprehensive configuration documentation
- [Bedrock Adapter](backend/infrastructure/bedrock/README.md) - Bedrock integration details
- [Backend API](backend/README.md) - Backend-specific documentation
- [Frontend](frontend/README.md) - Frontend-specific documentation
- [Requirements](. kiro/specs/chat-ui/requirements.md) - Feature requirements
- [Design Document](.kiro/specs/chat-ui/design.md) - System design and architecture
- [Implementation Tasks](.kiro/specs/chat-ui/tasks.md) - Development task list

## Troubleshooting

### Common Issues

#### Backend Won't Start

**Symptom:** Server fails to start or crashes immediately

**Solutions:**

1. **Port already in use**
   ```bash
   # Check what's using port 8080
   lsof -i :8080
   
   # Kill the process or change port
   export SERVER_PORT=8081
   go run cmd/server/main.go
   ```

2. **Go version too old**
   ```bash
   # Check Go version (requires 1.21+)
   go version
   
   # Update Go if needed
   # Download from https://go.dev/dl/
   ```

3. **Missing dependencies**
   ```bash
   cd backend
   go mod download
   go mod tidy
   ```

4. **AWS credentials not configured**
   ```bash
   # Check AWS credentials
   aws sts get-caller-identity
   
   # Configure AWS CLI
   aws configure
   
   # Or use environment variables
   export AWS_ACCESS_KEY_ID=your_key
   export AWS_SECRET_ACCESS_KEY=your_secret
   export AWS_REGION=us-east-1
   ```

5. **Invalid Bedrock configuration**
   ```bash
   # Run in mock mode (no Bedrock required)
   unset BEDROCK_AGENT_ID
   unset BEDROCK_AGENT_ALIAS_ID
   go run cmd/server/main.go
   ```

#### Frontend Won't Start

**Symptom:** Development server fails to start

**Solutions:**

1. **Port already in use**
   ```bash
   # Check what's using port 5173
   lsof -i :5173
   
   # Vite will automatically try next available port
   # Or specify a different port
   npm run dev -- --port 3000
   ```

2. **Node version too old**
   ```bash
   # Check Node version (requires 18+)
   node --version
   
   # Update Node if needed
   # Use nvm: nvm install 18
   ```

3. **Dependency issues**
   ```bash
   cd frontend
   
   # Clear cache and reinstall
   rm -rf node_modules package-lock.json
   npm install
   
   # Or use clean install
   npm ci
   ```

4. **Build errors**
   ```bash
   # Check for TypeScript errors
   npm run type-check
   
   # Clear Vite cache
   rm -rf node_modules/.vite
   npm run dev
   ```

#### WebSocket Connection Fails

**Symptom:** Frontend can't connect to backend WebSocket

**Solutions:**

1. **Backend not running**
   ```bash
   # Verify backend is running
   curl http://localhost:8080/health
   
   # Should return: {"status":"healthy"}
   ```

2. **CORS errors**
   ```bash
   # Check browser console for CORS errors
   # Backend allows all origins in development
   # In production, configure CORS in handler.go
   ```

3. **Invalid session ID**
   ```bash
   # Create a new session
   curl -X POST http://localhost:8080/api/sessions
   
   # Use the returned session ID
   ```

4. **WebSocket URL incorrect**
   ```bash
   # Check frontend environment variables
   # In .env or frontend/.env
   VITE_WS_URL=ws://localhost:8080
   
   # Restart frontend after changing
   ```

5. **Firewall blocking WebSocket**
   ```bash
   # Check firewall rules
   # Allow port 8080 for WebSocket connections
   ```

#### Bedrock Integration Issues

**Symptom:** Errors when sending messages to Bedrock

**Solutions:**

1. **Rate limit exceeded**
   ```
   Error: RATE_LIMIT_EXCEEDED
   
   # Wait and retry
   # Or increase retry configuration in .env:
   BEDROCK_MAX_RETRIES=5
   BEDROCK_MAX_BACKOFF=60s
   ```

2. **Invalid agent ID**
   ```
   Error: Agent not found
   
   # Verify agent ID in AWS Console
   # Update .env with correct ID:
   BEDROCK_AGENT_ID=your_correct_agent_id
   ```

3. **Insufficient IAM permissions**
   ```
   Error: AccessDeniedException
   
   # Add required permissions to IAM role/user:
   # - bedrock:InvokeAgent
   # - bedrock:InvokeAgentStream
   ```

4. **Knowledge base not found**
   ```
   Error: Knowledge base not found
   
   # Verify knowledge base ID
   # Or remove from configuration:
   unset BEDROCK_KNOWLEDGE_BASE_ID
   ```

5. **Timeout errors**
   ```
   Error: Request timed out
   
   # Increase timeout in .env:
   BEDROCK_REQUEST_TIMEOUT=120s
   WS_STREAM_TIMEOUT=10m
   ```

#### Docker Issues

**Symptom:** Docker containers fail to start or communicate

**Solutions:**

1. **Build failures**
   ```bash
   # Clean rebuild
   docker-compose down -v
   docker-compose build --no-cache
   docker-compose up
   ```

2. **Container crashes**
   ```bash
   # Check logs
   docker-compose logs backend
   docker-compose logs frontend
   
   # Check container status
   docker-compose ps
   ```

3. **Network issues**
   ```bash
   # Recreate network
   docker-compose down
   docker network prune
   docker-compose up
   ```

4. **Volume permission issues**
   ```bash
   # Fix permissions
   sudo chown -R $USER:$USER .
   ```

#### Test Failures

**Symptom:** Tests fail unexpectedly

**Solutions:**

1. **Backend tests fail**
   ```bash
   # Run with verbose output
   go test ./... -v
   
   # Run specific failing test
   go test ./path/to/package -run TestName -v
   
   # Check for race conditions
   go test ./... -race
   ```

2. **Frontend tests fail**
   ```bash
   # Clear test cache
   npm test -- --clearCache
   
   # Run with verbose output
   npm test -- --reporter=verbose
   
   # Run specific test
   npm test -- MessageInput.test.ts
   ```

3. **Property tests fail**
   ```bash
   # Property tests may fail due to randomness
   # Run multiple times to verify
   npm test -- --grep "Property" --repeat 3
   
   # Check the counterexample in test output
   # Fix the code or adjust the property
   ```

### Getting Help

If you're still experiencing issues:

1. **Check logs**
   - Backend: Console output or log files
   - Frontend: Browser console (F12)
   - Docker: `docker-compose logs`

2. **Enable debug logging**
   ```bash
   # Backend
   export LOG_LEVEL=debug
   
   # Frontend
   # Check browser console for detailed logs
   ```

3. **Verify configuration**
   ```bash
   # Backend configuration endpoint (development only)
   curl http://localhost:8080/api/config
   ```

4. **Check system resources**
   ```bash
   # Memory and CPU usage
   docker stats
   
   # Disk space
   df -h
   ```

5. **Review documentation**
   - [Configuration Guide](backend/docs/CONFIGURATION.md)
   - [Docker Setup](DOCKER.md)
   - [Bedrock Adapter](backend/infrastructure/bedrock/README.md)

### Known Limitations

- **Session Storage**: Currently in-memory only (POC). Sessions are lost on server restart.
- **Message History**: Limited to 500 messages per session for performance.
- **Concurrent Users**: Not optimized for high concurrency (POC).
- **Authentication**: No authentication implemented (POC).
- **CORS**: Allows all origins in development (restrict in production).

### Performance Tips

1. **Reduce WebSocket timeouts** for faster failure detection
2. **Increase buffer sizes** for high-throughput scenarios
3. **Enable virtual scrolling** for conversations with >100 messages
4. **Use connection pooling** for database connections (when implemented)
5. **Cache knowledge base queries** to reduce Bedrock API calls

## License

This is a POC project for demonstration purposes.
