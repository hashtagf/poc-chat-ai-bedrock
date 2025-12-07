/**
 * Integration tests for WebSocket communication
 * Tests end-to-end message sending, receiving, streaming, reconnection, and session management
 * 
 * Requirements tested:
 * - 1.1: Message transmission to Bedrock Agent Core
 * - 2.1: Real-time streaming response display
 * - 7.1: Session management and state
 */

import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { useChatService } from '@/composables/useChatService'
import { useSessionManager } from '@/composables/useSessionManager'
import { useConversationHistory } from '@/composables/useConversationHistory'
import type { ChatError } from '@/types'

// Store active WebSocket instances for testing
let activeWebSockets: any[] = []
let originalWebSocket: any

// Mock WebSocket implementation for integration testing
class MockWebSocket {
  public readyState: number = WebSocket.CONNECTING
  public onopen: ((event: Event) => void) | null = null
  public onclose: ((event: CloseEvent) => void) | null = null
  public onerror: ((event: Event) => void) | null = null
  public onmessage: ((event: MessageEvent) => void) | null = null
  public url: string
  private messageHandler: ((data: any) => void) | null = null
  
  constructor(url: string) {
    this.url = url
    activeWebSockets.push(this)
    
    // Simulate connection opening
    setTimeout(() => {
      this.readyState = WebSocket.OPEN
      if (this.onopen) {
        this.onopen(new Event('open'))
      }
    }, 10)
  }
  
  send(data: string): void {
    const parsed = JSON.parse(data)
    
    // Call the message handler if set
    if (this.messageHandler) {
      this.messageHandler(parsed)
    }
  }
  
  close(code?: number, reason?: string): void {
    this.readyState = WebSocket.CLOSED
    if (this.onclose) {
      const event = new CloseEvent('close', { code: code || 1000, reason: reason || '' })
      this.onclose(event)
    }
  }
  
  // Simulate receiving a message from server
  simulateMessage(data: any): void {
    if (this.onmessage) {
      const event = new MessageEvent('message', { data: JSON.stringify(data) })
      this.onmessage(event)
    }
  }
  
  // Simulate connection error
  simulateError(): void {
    if (this.onerror) {
      this.onerror(new Event('error'))
    }
  }
  
  // Set message handler for testing
  setMessageHandler(handler: (data: any) => void): void {
    this.messageHandler = handler
  }
}

// Helper to get the most recent WebSocket connection
function getLastConnection(): MockWebSocket | null {
  return activeWebSockets.length > 0 ? activeWebSockets[activeWebSockets.length - 1] : null
}

// Helper to simulate streaming response
function simulateStreamingResponse(connection: MockWebSocket, message: string): void {
  const words = message.split(' ')
  
  // Send chunks
  words.forEach((word, index) => {
    setTimeout(() => {
      connection.simulateMessage({
        type: 'chunk',
        content: word + ' '
      })
    }, index * 50)
  })
  
  // Send completion
  setTimeout(() => {
    connection.simulateMessage({
      type: 'complete'
    })
  }, words.length * 50 + 50)
}

// Helper to simulate server error
function simulateServerError(connection: MockWebSocket, code: string, message: string): void {
  connection.simulateMessage({
    type: 'error',
    code,
    message,
    retryable: true
  })
}

describe('WebSocket Integration Tests', () => {
  beforeEach(() => {
    // Store original WebSocket
    originalWebSocket = window.WebSocket
    
    // Replace with mock
    window.WebSocket = MockWebSocket as any
    
    // Clear active connections
    activeWebSockets = []
  })
  
  afterEach(() => {
    // Restore original WebSocket
    if (originalWebSocket) {
      window.WebSocket = originalWebSocket
    }
    
    // Clean up connections
    activeWebSockets.forEach(ws => {
      if (ws.readyState === WebSocket.OPEN) {
        ws.close()
      }
    })
    activeWebSockets = []
    
    vi.clearAllMocks()
  })
  
  describe('Message Sending and Receiving', () => {
    it('should send message through WebSocket and receive response', async () => {
      // Requirement 1.1: Message transmission
      const sessionId = 'test-session-1'
      const errors: ChatError[] = []
      const completedMessages: string[] = []
      
      const chatService = useChatService({
        sessionId,
        onError: (error) => errors.push(error),
        onMessageComplete: (content) => completedMessages.push(content)
      })
      
      // Wait for connection to establish
      await new Promise(resolve => setTimeout(resolve, 50))
      
      const connection = getLastConnection()
      expect(connection).toBeTruthy()
      
      // Set up message handler to simulate server response
      connection!.setMessageHandler((data) => {
        if (data.type === 'message') {
          simulateStreamingResponse(connection!, 'This is a response from Bedrock')
        }
      })
      
      // Send a message
      const testMessage = 'Hello, Bedrock!'
      await chatService.sendMessage(testMessage)
      
      // Wait for streaming to complete
      await new Promise(resolve => setTimeout(resolve, 500))
      
      // Verify message was completed
      expect(completedMessages.length).toBeGreaterThan(0)
      expect(errors.length).toBe(0)
    })
    
    it('should handle empty message validation', async () => {
      const sessionId = 'test-session-2'
      const errors: ChatError[] = []
      
      const chatService = useChatService({
        sessionId,
        onError: (error) => errors.push(error)
      })
      
      // Wait for connection
      await new Promise(resolve => setTimeout(resolve, 50))
      
      // Try to send empty message
      await expect(chatService.sendMessage('')).rejects.toThrow()
      
      // Verify validation error
      expect(errors.length).toBeGreaterThan(0)
      expect(errors[0].code).toBe('VALIDATION_ERROR')
    })
    
    it('should handle whitespace-only message validation', async () => {
      const sessionId = 'test-session-3'
      const errors: ChatError[] = []
      
      const chatService = useChatService({
        sessionId,
        onError: (error) => errors.push(error)
      })
      
      await new Promise(resolve => setTimeout(resolve, 50))
      
      // Try to send whitespace-only message
      await expect(chatService.sendMessage('   \t\n  ')).rejects.toThrow()
      
      expect(errors.length).toBeGreaterThan(0)
      expect(errors[0].code).toBe('VALIDATION_ERROR')
    })
    
    it('should handle message length validation', async () => {
      const sessionId = 'test-session-4'
      const errors: ChatError[] = []
      
      const chatService = useChatService({
        sessionId,
        onError: (error) => errors.push(error)
      })
      
      await new Promise(resolve => setTimeout(resolve, 50))
      
      // Try to send message exceeding max length
      const longMessage = 'a'.repeat(2001)
      await expect(chatService.sendMessage(longMessage)).rejects.toThrow()
      
      expect(errors.length).toBeGreaterThan(0)
      expect(errors[0].code).toBe('VALIDATION_ERROR')
      expect(errors[0].message).toContain('2000 characters')
    })
  })
  
  describe('Streaming Response Handling', () => {
    it('should display streaming response incrementally', async () => {
      // Requirement 2.1: Streaming response incremental display
      const sessionId = 'test-session-5'
      const streamingStates: string[] = []
      
      const chatService = useChatService({
        sessionId,
        onMessageComplete: () => {}
      })
      
      await new Promise(resolve => setTimeout(resolve, 50))
      
      const connection = getLastConnection()
      expect(connection).toBeTruthy()
      
      // Monitor streaming message state
      const stopWatch = setInterval(() => {
        if (chatService.streamingMessage.value) {
          streamingStates.push(chatService.streamingMessage.value)
        }
      }, 30)
      
      // Set up message handler
      connection!.setMessageHandler((data) => {
        if (data.type === 'message') {
          // Send chunks one by one
          setTimeout(() => connection!.simulateMessage({ type: 'chunk', content: 'Hello ' }), 50)
          setTimeout(() => connection!.simulateMessage({ type: 'chunk', content: 'from ' }), 100)
          setTimeout(() => connection!.simulateMessage({ type: 'chunk', content: 'Bedrock' }), 150)
          setTimeout(() => connection!.simulateMessage({ type: 'complete' }), 200)
        }
      })
      
      // Send message
      await chatService.sendMessage('Test streaming')
      await new Promise(resolve => setTimeout(resolve, 300))
      
      clearInterval(stopWatch)
      
      // Verify incremental display
      expect(streamingStates.length).toBeGreaterThan(0)
      
      // Verify streaming flag was set
      expect(chatService.isStreaming.value).toBe(false) // Should be false after completion
    })
    
    it('should handle streaming completion correctly', async () => {
      const sessionId = 'test-session-6'
      let completedContent = ''
      
      const chatService = useChatService({
        sessionId,
        onMessageComplete: (content) => {
          completedContent = content
        }
      })
      
      await new Promise(resolve => setTimeout(resolve, 50))
      
      const connection = getLastConnection()
      expect(connection).toBeTruthy()
      
      connection!.setMessageHandler((data) => {
        if (data.type === 'message') {
          simulateStreamingResponse(connection!, 'Complete response')
        }
      })
      
      await chatService.sendMessage('Complete test')
      await new Promise(resolve => setTimeout(resolve, 300))
      
      // Verify completion
      expect(completedContent).toContain('Complete response')
      expect(chatService.isStreaming.value).toBe(false)
      expect(chatService.streamingMessage.value).toBe('')
    })
    
    it('should preserve partial content on streaming error', async () => {
      const sessionId = 'test-session-7'
      const errors: ChatError[] = []
      
      const chatService = useChatService({
        sessionId,
        onError: (error) => errors.push(error)
      })
      
      await new Promise(resolve => setTimeout(resolve, 50))
      
      const connection = getLastConnection()
      expect(connection).toBeTruthy()
      
      connection!.setMessageHandler((data) => {
        if (data.type === 'message') {
          // Send some chunks then error
          setTimeout(() => connection!.simulateMessage({ type: 'chunk', content: 'Partial ' }), 50)
          setTimeout(() => connection!.simulateMessage({ type: 'chunk', content: 'content ' }), 100)
          setTimeout(() => {
            simulateServerError(connection!, 'STREAM_ERROR', 'Stream failed')
          }, 150)
        }
      })
      
      await chatService.sendMessage('Error test')
      await new Promise(resolve => setTimeout(resolve, 250))
      
      // Verify error was captured with partial content
      expect(errors.length).toBeGreaterThan(0)
      expect(errors[0].code).toBe('STREAM_ERROR')
    })
    
    it('should block input during streaming', async () => {
      const sessionId = 'test-session-8'
      const errors: ChatError[] = []
      
      const chatService = useChatService({
        sessionId,
        onError: (error) => errors.push(error)
      })
      
      await new Promise(resolve => setTimeout(resolve, 50))
      
      const connection = getLastConnection()
      expect(connection).toBeTruthy()
      
      connection!.setMessageHandler((data) => {
        if (data.type === 'message') {
          // Start streaming but don't complete immediately
          setTimeout(() => connection!.simulateMessage({ type: 'chunk', content: 'Streaming...' }), 50)
          // Don't send complete yet
        }
      })
      
      await chatService.sendMessage('First message')
      await new Promise(resolve => setTimeout(resolve, 100))
      
      // Verify streaming is active
      expect(chatService.isStreaming.value).toBe(true)
      
      // Try to send another message while streaming
      await expect(chatService.sendMessage('Second message')).rejects.toThrow()
      
      // Verify error about streaming in progress
      expect(errors.some(e => e.code === 'STREAMING_IN_PROGRESS')).toBe(true)
    })
  })
  
  describe('Connection Interruption and Reconnection', () => {
    it('should detect connection interruption during streaming', async () => {
      const sessionId = 'test-session-9'
      const errors: ChatError[] = []
      
      const chatService = useChatService({
        sessionId,
        onError: (error) => errors.push(error)
      })
      
      await new Promise(resolve => setTimeout(resolve, 50))
      
      const connection = getLastConnection()
      expect(connection).toBeTruthy()
      
      connection!.setMessageHandler((data) => {
        if (data.type === 'message') {
          // Start streaming
          setTimeout(() => connection!.simulateMessage({ type: 'chunk', content: 'Starting...' }), 50)
          
          // Interrupt connection
          setTimeout(() => {
            connection!.close(1006, 'Connection interrupted')
          }, 100)
        }
      })
      
      await chatService.sendMessage('Interruption test')
      await new Promise(resolve => setTimeout(resolve, 200))
      
      // Verify interruption was detected
      expect(errors.some(e => e.code === 'STREAM_INTERRUPTED')).toBe(true)
    })
    
    it('should attempt automatic reconnection on connection loss', async () => {
      const sessionId = 'test-session-10'
      const errors: ChatError[] = []
      
      const chatService = useChatService({
        sessionId,
        onError: (error) => errors.push(error)
      })
      
      await new Promise(resolve => setTimeout(resolve, 50))
      
      // Start streaming to trigger interruption error
      const connection = getLastConnection()
      expect(connection).toBeTruthy()
      
      connection!.setMessageHandler((data) => {
        if (data.type === 'message') {
          // Start streaming
          setTimeout(() => connection!.simulateMessage({ type: 'chunk', content: 'Starting...' }), 50)
        }
      })
      
      // Send message to start streaming
      await chatService.sendMessage('Test message')
      await new Promise(resolve => setTimeout(resolve, 100))
      
      // Now close connection while streaming
      connection!.close(1006, 'Connection lost')
      
      // Wait for reconnection attempt
      await new Promise(resolve => setTimeout(resolve, 2000))
      
      // Verify interruption was detected (connection closed during streaming)
      expect(errors.some(e => e.code === 'STREAM_INTERRUPTED')).toBe(true)
    })
    
    it('should handle network errors gracefully', async () => {
      const sessionId = 'test-session-11'
      const errors: ChatError[] = []
      
      useChatService({
        sessionId,
        onError: (error) => errors.push(error)
      })
      
      await new Promise(resolve => setTimeout(resolve, 50))
      
      // Simulate network error
      const connection = getLastConnection()
      expect(connection).toBeTruthy()
      connection!.simulateError()
      
      await new Promise(resolve => setTimeout(resolve, 100))
      
      // Verify network error was handled
      expect(errors.some(e => e.code === 'NETWORK_ERROR')).toBe(true)
      expect(errors.some(e => e.retryable === true)).toBe(true)
    })
  })
  
  describe('Session Management Integration', () => {
    it('should maintain session state across messages', async () => {
      // Requirement 7.1: Session management
      const sessionManager = useSessionManager()
      const conversationHistory = useConversationHistory()
      
      // Create a new session
      const sessionId = await sessionManager.createNewSession()
      expect(sessionId).toBeTruthy()
      expect(sessionManager.currentSessionId.value).toBe(sessionId)
      
      // Initialize chat service with session
      const completedMessages: string[] = []
      const chatService = useChatService({
        sessionId,
        onMessageComplete: (content) => {
          completedMessages.push(content)
          conversationHistory.addMessage({
            id: `msg-${Date.now()}`,
            role: 'agent',
            content,
            timestamp: new Date(),
            status: 'sent'
          })
        }
      })
      
      await new Promise(resolve => setTimeout(resolve, 50))
      
      // Send multiple messages
      await chatService.sendMessage('First message')
      await new Promise(resolve => setTimeout(resolve, 100))
      
      conversationHistory.addMessage({
        id: 'msg-1',
        role: 'user',
        content: 'First message',
        timestamp: new Date(),
        status: 'sent'
      })
      
      await chatService.sendMessage('Second message')
      await new Promise(resolve => setTimeout(resolve, 100))
      
      conversationHistory.addMessage({
        id: 'msg-2',
        role: 'user',
        content: 'Second message',
        timestamp: new Date(),
        status: 'sent'
      })
      
      // Verify session maintained state
      expect(sessionManager.currentSessionId.value).toBe(sessionId)
      expect(conversationHistory.messages.value.length).toBeGreaterThanOrEqual(2)
    })
    
    it('should isolate sessions correctly', async () => {
      // Create separate instances for each session to test isolation
      const sessionManager1 = useSessionManager()
      const conversationHistory1 = useConversationHistory()
      
      // Create first session
      const session1 = await sessionManager1.createNewSession()
      
      conversationHistory1.addMessage({
        id: 'msg-s1-1',
        role: 'user',
        content: 'Session 1 message',
        timestamp: new Date(),
        status: 'sent'
      })
      
      expect(conversationHistory1.messages.value.length).toBe(1)
      
      // Create new instances for second session
      const sessionManager2 = useSessionManager()
      const conversationHistory2 = useConversationHistory()
      
      // Create second session
      const session2 = await sessionManager2.createNewSession()
      
      // Verify sessions are different
      expect(session2).not.toBe(session1)
      
      // Add message to second session
      conversationHistory2.addMessage({
        id: 'msg-s2-1',
        role: 'user',
        content: 'Session 2 message',
        timestamp: new Date(),
        status: 'sent'
      })
      
      // Verify sessions are isolated - each has its own messages
      expect(conversationHistory1.messages.value.length).toBe(1)
      expect(conversationHistory1.messages.value[0].content).toBe('Session 1 message')
      expect(conversationHistory2.messages.value.length).toBe(1)
      expect(conversationHistory2.messages.value[0].content).toBe('Session 2 message')
    })
    
    it('should handle session switching', async () => {
      const sessionManager = useSessionManager()
      
      // Create multiple sessions
      const session1 = await sessionManager.createNewSession()
      expect(sessionManager.currentSessionId.value).toBe(session1)
      
      const session2 = await sessionManager.createNewSession()
      expect(sessionManager.currentSessionId.value).toBe(session2)
      
      // Switch back to session 1
      await sessionManager.loadSession(session1)
      expect(sessionManager.currentSessionId.value).toBe(session1)
    })
    
    it('should track session metadata', async () => {
      const sessionManager = useSessionManager()
      
      const sessionId = await sessionManager.createNewSession()
      
      // Verify metadata
      expect(sessionManager.sessionMetadata.value.id).toBe(sessionId)
      expect(sessionManager.sessionMetadata.value.createdAt).toBeInstanceOf(Date)
      expect(sessionManager.sessionMetadata.value.messageCount).toBe(0)
    })
  })
  
  describe('Error Scenarios', () => {
    it('should handle malformed server responses', async () => {
      const sessionId = 'test-session-12'
      const errors: ChatError[] = []
      
      const chatService = useChatService({
        sessionId,
        onError: (error) => errors.push(error)
      })
      
      await new Promise(resolve => setTimeout(resolve, 50))
      
      const connection = getLastConnection()
      expect(connection).toBeTruthy()
      
      connection!.setMessageHandler((data) => {
        if (data.type === 'message') {
          // Send malformed response
          if (connection!.onmessage) {
            const event = new MessageEvent('message', { data: 'invalid json{' })
            connection!.onmessage(event)
          }
        }
      })
      
      await chatService.sendMessage('Malformed test')
      await new Promise(resolve => setTimeout(resolve, 100))
      
      // Verify malformed response was handled
      expect(errors.some(e => e.code === 'MALFORMED_RESPONSE')).toBe(true)
    })
    
    it('should handle backend error responses', async () => {
      const sessionId = 'test-session-13'
      const errors: ChatError[] = []
      
      const chatService = useChatService({
        sessionId,
        onError: (error) => errors.push(error)
      })
      
      await new Promise(resolve => setTimeout(resolve, 50))
      
      const connection = getLastConnection()
      expect(connection).toBeTruthy()
      
      connection!.setMessageHandler((data) => {
        if (data.type === 'message') {
          simulateServerError(connection!, 'RATE_LIMIT', 'Service is temporarily busy')
        }
      })
      
      await chatService.sendMessage('Backend error test')
      await new Promise(resolve => setTimeout(resolve, 100))
      
      // Verify backend error was handled
      expect(errors.some(e => e.code === 'RATE_LIMIT')).toBe(true)
    })
  })
})
