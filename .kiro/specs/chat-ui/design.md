# Chat UI Design Document

## Overview

The Chat UI is a Vue 3-based conversational interface that enables users to interact with Amazon Bedrock Agent Core. The design follows hexagonal architecture principles, separating the UI layer from business logic and infrastructure concerns. The interface supports real-time streaming responses, conversation history management, and knowledge base citation display.

The system is built as a single-page application using Vue 3 Composition API with Tailwind CSS for styling. It communicates with a Go backend that handles Bedrock Agent Core integration, ensuring clean separation between presentation and business logic layers.

## Architecture

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        Chat UI (Vue 3)                       │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │   Message    │  │ Conversation │  │   Citation   │     │
│  │   Input      │  │   Display    │  │   Display    │     │
│  └──────────────┘  └──────────────┘  └──────────────┘     │
│         │                  │                  │             │
│         └──────────────────┴──────────────────┘             │
│                            │                                │
│                   ┌────────▼────────┐                       │
│                   │  Chat Service   │                       │
│                   │   (Composable)  │                       │
│                   └────────┬────────┘                       │
└────────────────────────────┼──────────────────────────────┘
                             │ HTTP/WebSocket
                    ┌────────▼────────┐
                    │   API Gateway   │
                    │   (Go Backend)  │
                    └────────┬────────┘
                             │
          ┌──────────────────┼──────────────────┐
          │                  │                  │
    ┌─────▼─────┐    ┌──────▼──────┐   ┌──────▼──────┐
    │  Session  │    │   Bedrock   │   │  Knowledge  │
    │  Manager  │    │    Agent    │   │    Base     │
    └───────────┘    │    Core     │   │   Service   │
                     └─────────────┘   └─────────────┘
```

### Component Responsibilities

**Frontend (Vue 3)**
- **ChatContainer**: Root component managing layout and state coordination
- **MessageInput**: Handles user input, validation, and submission
- **MessageList**: Displays conversation history with auto-scrolling
- **MessageBubble**: Renders individual messages with styling and metadata
- **CitationDisplay**: Shows knowledge base citations and source information
- **LoadingIndicator**: Provides visual feedback during processing
- **ErrorDisplay**: Shows user-friendly error messages

**Composables (Business Logic)**
- **useChatService**: Manages message sending, receiving, and streaming
- **useConversationHistory**: Handles message storage and retrieval
- **useSessionManager**: Controls session lifecycle and identification
- **useErrorHandler**: Centralizes error handling and user messaging

**Backend (Go)**
- **ChatHandler**: HTTP/WebSocket endpoint for message exchange
- **BedrockAdapter**: Interfaces with AWS Bedrock Agent Core SDK
- **StreamProcessor**: Handles streaming response parsing and forwarding
- **SessionRepository**: Persists and retrieves conversation sessions

## Components and Interfaces

### Frontend Components

#### ChatContainer Component
```typescript
interface ChatContainerProps {
  sessionId?: string;
  initialMessages?: Message[];
}

interface ChatContainerEmits {
  sessionCreated: (sessionId: string) => void;
  error: (error: ChatError) => void;
}
```

#### MessageInput Component
```typescript
interface MessageInputProps {
  disabled: boolean;
  placeholder?: string;
}

interface MessageInputEmits {
  submit: (content: string) => void;
}

// Validation rules
- Minimum length: 1 character (after trimming)
- Maximum length: 2000 characters
- No whitespace-only messages
```

#### MessageList Component
```typescript
interface MessageListProps {
  messages: Message[];
  isStreaming: boolean;
  streamingContent?: string;
}

// Auto-scroll behavior
- Scroll to bottom on new message
- Maintain scroll position when user scrolls up
- Resume auto-scroll when user scrolls to bottom
```

#### CitationDisplay Component
```typescript
interface CitationDisplayProps {
  citations: Citation[];
  expanded?: boolean;
}

interface Citation {
  sourceId: string;
  sourceName: string;
  excerpt: string;
  confidence?: number;
  url?: string;
}
```

### Composables (Business Logic Layer)

#### useChatService
```typescript
interface ChatService {
  sendMessage(content: string): Promise<void>;
  streamingMessage: Ref<string>;
  isStreaming: Ref<boolean>;
  error: Ref<ChatError | null>;
  clearError(): void;
}

// Responsibilities:
// - Establish WebSocket connection for streaming
// - Send messages to backend
// - Process streaming responses
// - Handle connection errors and retries
```

#### useConversationHistory
```typescript
interface ConversationHistory {
  messages: Ref<Message[]>;
  addMessage(message: Message): void;
  clearHistory(): void;
  getMessageById(id: string): Message | undefined;
}

interface Message {
  id: string;
  role: 'user' | 'agent';
  content: string;
  timestamp: Date;
  citations?: Citation[];
  status: 'sending' | 'sent' | 'error';
}
```

#### useSessionManager
```typescript
interface SessionManager {
  currentSessionId: Ref<string>;
  createNewSession(): Promise<string>;
  loadSession(sessionId: string): Promise<void>;
  sessionMetadata: Ref<SessionMetadata>;
}

interface SessionMetadata {
  id: string;
  createdAt: Date;
  messageCount: number;
}
```

### Backend Interfaces

#### ChatHandler (HTTP/WebSocket)
```go
type ChatHandler interface {
    HandleMessage(ctx context.Context, req MessageRequest) (*MessageResponse, error)
    HandleStream(ctx context.Context, req MessageRequest, stream StreamWriter) error
}

type MessageRequest struct {
    SessionID string `json:"session_id"`
    Content   string `json:"content"`
}

type MessageResponse struct {
    MessageID  string     `json:"message_id"`
    Content    string     `json:"content"`
    Citations  []Citation `json:"citations,omitempty"`
    Timestamp  time.Time  `json:"timestamp"`
}

type StreamWriter interface {
    WriteChunk(chunk string) error
    WriteCitation(citation Citation) error
    WriteError(err error) error
    Close() error
}
```

#### BedrockAdapter (Infrastructure)
```go
type BedrockAdapter interface {
    InvokeAgent(ctx context.Context, input AgentInput) (*AgentResponse, error)
    InvokeAgentStream(ctx context.Context, input AgentInput) (StreamReader, error)
}

type AgentInput struct {
    SessionID        string
    Message          string
    KnowledgeBaseIDs []string
}

type AgentResponse struct {
    Content   string
    Citations []Citation
    Metadata  map[string]interface{}
}

type StreamReader interface {
    Read() (chunk string, done bool, err error)
    Close() error
}
```

## Data Models

### Frontend Models

```typescript
// Message model
interface Message {
  id: string;                    // UUID v4
  role: 'user' | 'agent';
  content: string;
  timestamp: Date;
  citations?: Citation[];
  status: 'sending' | 'sent' | 'error';
  errorMessage?: string;
}

// Citation model
interface Citation {
  sourceId: string;
  sourceName: string;
  excerpt: string;
  confidence?: number;          // 0.0 to 1.0
  url?: string;
  metadata?: Record<string, any>;
}

// Session model
interface Session {
  id: string;
  createdAt: Date;
  lastMessageAt?: Date;
  messageCount: number;
}

// Error model
interface ChatError {
  code: string;
  message: string;
  retryable: boolean;
  details?: Record<string, any>;
}
```

### Backend Models

```go
// Domain entities
type Message struct {
    ID        string
    SessionID string
    Role      MessageRole
    Content   string
    Timestamp time.Time
    Citations []Citation
    Status    MessageStatus
}

type MessageRole string
const (
    RoleUser  MessageRole = "user"
    RoleAgent MessageRole = "agent"
)

type MessageStatus string
const (
    StatusSending MessageStatus = "sending"
    StatusSent    MessageStatus = "sent"
    StatusError   MessageStatus = "error"
)

type Citation struct {
    SourceID   string
    SourceName string
    Excerpt    string
    Confidence float64
    URL        string
    Metadata   map[string]interface{}
}

type Session struct {
    ID            string
    CreatedAt     time.Time
    LastMessageAt *time.Time
    MessageCount  int
}
```


## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system—essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*

### Property 1: Message transmission for valid input
*For any* valid message content (non-empty, non-whitespace), when the user submits the message, the Chat UI should transmit it to the backend and add it to the conversation history.
**Validates: Requirements 1.1**

### Property 2: Input validation prevents invalid submission
*For any* string that is empty or contains only whitespace characters, the Chat UI should prevent message submission and the message should not be transmitted to the backend.
**Validates: Requirements 1.4, 8.3**

### Property 3: UI state during message processing
*For any* message being sent or response being generated, the Chat UI should disable the input field and display a loading indicator until processing completes.
**Validates: Requirements 1.2, 4.1**

### Property 4: Input field reset after successful send
*For any* successfully sent message, the Chat UI should clear the input field content and restore focus to the input field.
**Validates: Requirements 1.5**

### Property 5: Streaming response incremental display
*For any* streaming response, as each chunk arrives, the Chat UI should append it to the current message without waiting for the complete response.
**Validates: Requirements 2.1**

### Property 6: Streaming completion state transition
*For any* streaming response that completes successfully, the Chat UI should mark the message as complete and re-enable user input.
**Validates: Requirements 2.3**

### Property 7: Streaming error preservation
*For any* streaming response that fails mid-stream, the Chat UI should preserve the partial content received and display an error indicator.
**Validates: Requirements 2.4**

### Property 8: Input blocking during streaming
*For any* active streaming response, the Chat UI should prevent new message submission until the stream completes or fails.
**Validates: Requirements 2.5**

### Property 9: Chronological message ordering
*For any* set of messages in a conversation, the Chat UI should display them in chronological order based on their timestamps.
**Validates: Requirements 3.1**

### Property 10: Auto-scroll on new message
*For any* new message added to the conversation, if the user is at or near the bottom of the message list, the Chat UI should automatically scroll to show the new message.
**Validates: Requirements 3.2**

### Property 11: Timestamp display
*For any* message displayed in the conversation, the rendered output should include the message's timestamp.
**Validates: Requirements 3.4**

### Property 12: Error message sanitization
*For any* error that occurs (transmission, response generation, network), the Chat UI should display a user-friendly error message that does not expose internal details such as stack traces, AWS SDK errors, or system paths.
**Validates: Requirements 4.2**

### Property 13: Connection status indication
*For any* state where the connection to the backend is unavailable, the Chat UI should display a connection status indicator.
**Validates: Requirements 4.4**

### Property 14: Citation visual distinction
*For any* agent message that includes knowledge base citations, the Chat UI should render citation indicators that are visually distinct from regular message content.
**Validates: Requirements 4.5, 9.1**

### Property 15: Keyboard submission support
*For any* state where the input field is enabled, pressing the Enter key should submit the message (equivalent to clicking the send button).
**Validates: Requirements 5.2**

### Property 16: Semantic HTML structure
*For any* rendered Chat UI component, the HTML output should use semantic elements (e.g., `<main>`, `<article>`, `<button>`) rather than generic `<div>` elements for primary structural components.
**Validates: Requirements 5.3**

### Property 17: Reduced motion compliance
*For any* user with prefers-reduced-motion enabled, the Chat UI should disable or minimize animations and transitions.
**Validates: Requirements 5.4**

### Property 18: Color contrast compliance
*For any* text content displayed in the Chat UI, the color contrast ratio between text and background should meet WCAG AA standards (minimum 4.5:1 for normal text).
**Validates: Requirements 5.5**

### Property 19: Infrastructure error transformation
*For any* error originating from the infrastructure layer (AWS SDK, network), the error received by the UI layer should be a domain-specific error type, not a raw infrastructure error.
**Validates: Requirements 6.3**

### Property 20: Session reset clears history
*For any* active session with existing messages, creating a new session should clear the conversation history and generate a unique session identifier.
**Validates: Requirements 7.1, 7.2**

### Property 21: Session isolation
*For any* two different sessions, messages from one session should not appear in the conversation history of the other session.
**Validates: Requirements 7.3**

### Property 22: Session metadata display
*For any* active session, the Chat UI should display either the session identifier or creation timestamp.
**Validates: Requirements 7.4**

### Property 23: New session input focus
*For any* newly created session, the input field should receive focus automatically.
**Validates: Requirements 7.5**

### Property 24: Network error detection and notification
*For any* network connectivity loss during message transmission or response streaming, the Chat UI should detect the condition and display an error message with recovery options.
**Validates: Requirements 8.1**

### Property 25: Rate limit error handling
*For any* rate limit error from the Bedrock API, the Chat UI should display a message indicating temporary unavailability.
**Validates: Requirements 8.2**

### Property 26: Malformed response handling
*For any* empty or malformed response from the agent, the Chat UI should handle it gracefully without crashing and display appropriate user messaging.
**Validates: Requirements 8.4**

### Property 27: Error aggregation
*For any* sequence of multiple errors occurring within a short time window, the Chat UI should aggregate or summarize the errors rather than displaying each individually.
**Validates: Requirements 8.5**

### Property 28: Citation interaction reveals details
*For any* citation indicator in a message, when the user interacts with it (click/hover), the Chat UI should reveal the source information including source name, excerpt, and confidence score if available.
**Validates: Requirements 9.2**

### Property 29: Multi-citation distinction
*For any* message with multiple citations from different knowledge base sources, each citation should be distinguishable from the others in the UI.
**Validates: Requirements 9.3**

### Property 30: Non-cited response indication
*For any* agent response that does not include knowledge base citations, the Chat UI should indicate that the response is based on general knowledge rather than specific sources.
**Validates: Requirements 9.4**

### Property 31: Citation metadata display
*For any* citation that includes confidence scores or relevance indicators, the Chat UI should display this metadata alongside the citation.
**Validates: Requirements 9.5**

## Error Handling

### Error Categories

**Network Errors**
- Connection timeout: Display "Unable to connect. Please check your network connection."
- Connection lost: Display "Connection lost. Attempting to reconnect..."
- WebSocket closed: Attempt automatic reconnection with exponential backoff (1s, 2s, 4s, 8s, max 30s)

**Validation Errors**
- Empty message: Prevent submission, no error message needed (button disabled)
- Message too long: Display "Message exceeds maximum length of 2000 characters"
- Invalid characters: Display "Message contains invalid characters"

**Backend Errors**
- Rate limit (429): Display "Service is temporarily busy. Please try again in [X] seconds."
- Server error (500): Display "An error occurred. Please try again."
- Timeout: Display "Request timed out. Please try again."
- Invalid session: Clear session and prompt user to start new session

**Streaming Errors**
- Stream interrupted: Preserve partial content, display "Response incomplete. Please try again."
- Malformed chunk: Skip chunk, continue processing, log error
- Stream timeout: Finalize partial content, display "Response timed out"

**Bedrock-Specific Errors**
- Knowledge base unavailable: Display "Knowledge base temporarily unavailable. Responses may be limited."
- Agent unavailable: Display "AI agent is currently unavailable. Please try again later."
- Invalid input: Display "Unable to process message. Please rephrase and try again."

### Error Recovery Strategies

**Automatic Recovery**
- WebSocket reconnection: Exponential backoff up to 5 attempts
- Failed message retry: Store failed message, offer manual retry button
- Stream recovery: Attempt to resume stream if connection restored within 10 seconds

**User-Initiated Recovery**
- Retry button: Available for all retryable errors
- Clear error: Dismiss error message and return to normal state
- New session: Option to start fresh if session is corrupted

**Error Logging**
- Log all errors to console in development
- Send error telemetry to backend in production (without PII)
- Include error context: timestamp, session ID, message ID, error code

### Error State Management

```typescript
interface ErrorState {
  code: string;
  message: string;
  retryable: boolean;
  retryCount: number;
  timestamp: Date;
  context?: Record<string, any>;
}

// Error display rules:
// - Show only the most recent error
// - Auto-dismiss non-critical errors after 5 seconds
// - Keep critical errors visible until user dismisses
// - Aggregate multiple similar errors within 10 seconds
```

## Testing Strategy

### Unit Testing

The Chat UI will use **Vitest** as the testing framework with **Vue Test Utils** for component testing. Unit tests will focus on:

**Component Behavior**
- MessageInput: Validation logic, submission handling, focus management
- MessageList: Message rendering, scroll behavior, timestamp formatting
- CitationDisplay: Citation rendering, interaction handling, metadata display
- ErrorDisplay: Error message formatting, retry button functionality

**Composable Logic**
- useChatService: Message sending, streaming processing, error handling
- useConversationHistory: Message storage, retrieval, ordering
- useSessionManager: Session creation, switching, metadata management
- useErrorHandler: Error transformation, aggregation, retry logic

**Example Unit Tests**
- Verify empty input prevents submission
- Verify Enter key triggers message send
- Verify messages are ordered by timestamp
- Verify error messages don't expose internal details
- Verify session switch clears previous messages

### Property-Based Testing

The Chat UI will use **fast-check** for property-based testing in TypeScript. Each property-based test will:
- Run a minimum of 100 iterations
- Be tagged with a comment referencing the design document property
- Use the format: `// Feature: chat-ui, Property X: [property text]`

**Property Test Coverage**

Each correctness property listed above will be implemented as a property-based test. Key property tests include:

**Input Validation Properties**
- Property 2: Generate random whitespace strings, verify all are rejected
- Property 15: Generate random valid messages, verify Enter key submits each

**Message Handling Properties**
- Property 1: Generate random valid messages, verify each is transmitted
- Property 9: Generate random message sets, verify chronological ordering
- Property 11: Generate random messages, verify timestamps in rendered output

**Streaming Properties**
- Property 5: Generate random chunk sequences, verify incremental display
- Property 7: Generate random partial responses, verify preservation on error
- Property 8: Generate random streaming states, verify input blocking

**Session Properties**
- Property 20: Generate random sessions with messages, verify reset clears all
- Property 21: Generate random multi-session scenarios, verify isolation

**Error Handling Properties**
- Property 12: Generate random error types, verify no internal details exposed
- Property 19: Generate random infrastructure errors, verify transformation
- Property 27: Generate random error sequences, verify aggregation

**Citation Properties**
- Property 14: Generate random messages with citations, verify visual distinction
- Property 29: Generate random multi-citation messages, verify distinction
- Property 31: Generate random citations with metadata, verify display

### Integration Testing

Integration tests will verify the interaction between frontend and backend:

**WebSocket Communication**
- Establish connection and send messages
- Receive and process streaming responses
- Handle connection interruptions and reconnection
- Verify message ordering across network boundary

**Session Management**
- Create sessions and verify backend persistence
- Switch sessions and verify correct history loading
- Verify session isolation across multiple clients

**Error Scenarios**
- Simulate backend errors and verify UI handling
- Simulate network failures and verify recovery
- Simulate rate limiting and verify backoff behavior

### Test Configuration

```typescript
// vitest.config.ts
export default defineConfig({
  test: {
    environment: 'jsdom',
    coverage: {
      provider: 'v8',
      reporter: ['text', 'json', 'html'],
      exclude: ['**/*.spec.ts', '**/*.test.ts', '**/types/**']
    },
    setupFiles: ['./tests/setup.ts']
  }
});

// fast-check configuration
const fc = require('fast-check');

// Minimum 100 iterations for all property tests
const propertyTestConfig = {
  numRuns: 100,
  verbose: true
};
```

### Test Utilities

**Generators for Property Tests**
```typescript
// Generate valid message content
const validMessageArb = fc.string({ minLength: 1, maxLength: 2000 })
  .filter(s => s.trim().length > 0);

// Generate whitespace-only strings
const whitespaceArb = fc.stringOf(fc.constantFrom(' ', '\t', '\n', '\r'));

// Generate message sequences
const messageSequenceArb = fc.array(
  fc.record({
    role: fc.constantFrom('user', 'agent'),
    content: validMessageArb,
    timestamp: fc.date()
  })
);

// Generate citations
const citationArb = fc.record({
  sourceId: fc.uuid(),
  sourceName: fc.string(),
  excerpt: fc.string({ maxLength: 200 }),
  confidence: fc.option(fc.float({ min: 0, max: 1 })),
  url: fc.option(fc.webUrl())
});
```

**Mock Services**
```typescript
// Mock chat service for component testing
const createMockChatService = () => ({
  sendMessage: vi.fn(),
  streamingMessage: ref(''),
  isStreaming: ref(false),
  error: ref(null),
  clearError: vi.fn()
});

// Mock WebSocket for integration testing
const createMockWebSocket = () => ({
  send: vi.fn(),
  close: vi.fn(),
  addEventListener: vi.fn(),
  removeEventListener: vi.fn()
});
```

## Implementation Notes

### Technology Choices

**Frontend Framework**: Vue 3 with Composition API
- Reactive state management with `ref` and `reactive`
- Composables for reusable business logic
- TypeScript for type safety

**Styling**: Tailwind CSS
- Utility-first approach for rapid development
- Consistent design system
- Responsive design utilities

**Communication**: WebSocket for streaming
- Real-time bidirectional communication
- Efficient for streaming responses
- Automatic reconnection handling

**Testing**: Vitest + fast-check
- Fast unit test execution
- Property-based testing for comprehensive coverage
- Vue Test Utils for component testing

### Performance Considerations

**Message Rendering**
- Virtual scrolling for conversations with >100 messages
- Debounce scroll events (100ms)
- Lazy load message timestamps (format on demand)

**Streaming Optimization**
- Batch chunk updates (every 50ms) to reduce re-renders
- Use `requestAnimationFrame` for smooth scrolling
- Throttle typing indicator updates

**Memory Management**
- Limit in-memory message history to 500 messages
- Implement message pagination for older messages
- Clear streaming buffers after message completion

### Security Considerations

**Input Sanitization**
- Escape HTML in user messages to prevent XSS
- Validate message length before transmission
- Strip potentially dangerous characters

**Error Message Safety**
- Never expose stack traces to users
- Sanitize error messages from backend
- Log detailed errors server-side only

**Session Security**
- Use cryptographically secure session IDs (UUID v4)
- Validate session IDs on backend
- Implement session timeout (30 minutes of inactivity)

### Accessibility Features

**Keyboard Navigation**
- Tab order: input field → send button → message list → citations
- Enter to send message
- Escape to clear input or dismiss errors
- Arrow keys for message history navigation (future enhancement)

**Screen Reader Support**
- ARIA labels for all interactive elements
- ARIA live regions for streaming messages and errors
- Semantic HTML structure
- Alt text for loading indicators

**Visual Accessibility**
- WCAG AA color contrast (4.5:1 minimum)
- Focus indicators on all interactive elements
- Reduced motion support
- Scalable text (supports browser zoom)

### Future Enhancements

**Phase 2 Features** (not in current scope)
- Message editing and deletion
- Conversation search
- Export conversation history
- Multi-modal input (voice, images)
- Conversation branching
- Message reactions
- Collaborative sessions (multiple users)

**Performance Optimizations** (if needed)
- Service worker for offline support
- IndexedDB for local message persistence
- Message compression for large conversations
- CDN for static assets
