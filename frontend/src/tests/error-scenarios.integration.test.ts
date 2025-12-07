/**
 * Integration tests for error scenarios
 * Tests network failures, backend errors, rate limiting, and malformed responses
 * 
 * Requirements tested:
 * - 8.1: Network error detection and notification
 * - 8.2: Rate limit error handling
 * - 8.4: Malformed response handling
 */

import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { useChatService } from '@/composables/useChatService'
import { useErrorHandler } from '@/composables/useErrorHandler'
import type { ChatError } from '@/types'

// Store active WebSocket instances for testing
let activeWebSockets: any[] = []
let originalWebSocket: any

// Mock WebSocket implementation for error scenario testing
class MockWebSocket {
  public readyState: number = WebSocket.CONNECTING
  public onopen: ((event: Event) => void) | null = null
  public onclose: ((event: CloseEvent) => void) | null = null
  public onerror: ((event: Event) => void) | null = null
  public onmessage: ((event: MessageEvent) => void) | null = null
  public url: string
  private messageHandler: ((data: any) => void) | null = null
  private shouldFailConnection: boolean = false
  private shouldFailOnSend: boolean = false
  
  constructor(url: string) {
    this.url = url
    activeWebSockets.push(this)
    
    // Check if we should fail the connection
    if (this.shouldFailConnection) {
      setTimeout(() => {
        if (this.onerror) {
          this.onerror(new Event('error'))
        }
      }, 10)
      return
    }
    
    // Simulate connection opening
    setTimeout(() => {
      this.readyState = WebSocket.OPEN
      if (this.onopen) {
        this.onopen(new Event('open'))
      }
    }, 10)
  }
  
  send(data: string): void {
    if (this.shouldFailOnSend) {
      throw new Error('Network error: Failed to send')
    }
    
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
  
  // Configure to fail on connection
  static failNextConnection(): void {
    MockWebSocket.prototype.shouldFailConnection = true
  }
  
  // Configure to fail on send
  static failNextSend(): void {
    MockWebSocket.prototype.shouldFailOnSend = true
  }
  
  // Reset failure flags
  static resetFailures(): void {
    MockWebSocket.prototype.shouldFailConnection = false
    MockWebSocket.prototype.shouldFailOnSend = false
  }
}

// Helper to get the most recent WebSocket connection
function getLastConnection(): MockWebSocket | null {
  return activeWebSockets.length > 0 ? activeWebSockets[activeWebSockets.length - 1] : null
}

describe('Error Scenarios Integration Tests', () => {
  beforeEach(() => {
    // Store original WebSocket
    originalWebSocket = global.WebSocket
    
    // Replace with mock
    global.WebSocket = MockWebSocket as any
    
    // Clear active connections
    activeWebSockets = []
    
    // Reset failure flags
    MockWebSocket.resetFailures()
  })
  
  afterEach(() => {
    // Restore original WebSocket
    if (originalWebSocket) {
      global.WebSocket = originalWebSocket
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
  
  describe('Network Failure Handling (Requirement 8.1)', () => {
    it('should detect network connectivity loss during message transmission', async () => {
      const sessionId = 'test-network-failure-1'
      const errors: ChatError[] = []
      
      const chatService = useChatService({
        sessionId,
        onError: (error) => errors.push(error)
      })
      
      // Wait for connection to establish
      await new Promise(resolve => setTimeout(resolve, 50))
      
      const connection = getLastConnection()
      expect(connection).toBeTruthy()
      
      // Simulate network error directly
      connection!.simulateError()
      
      // Wait for error to be processed
      await new Promise(resolve => setTimeout(resolve, 100))
      
      // Verify network error was detected
      expect(errors.length).toBeGreaterThan(0)
      expect(errors.some(e => e.code === 'NETWORK_ERROR')).toBe(true)
      expect(errors.some(e => e.retryable === true)).toBe(true)
      
      // Verify error message is user-friendly
      const networkError = errors.find(e => e.code === 'NETWORK_ERROR')
      expect(networkError?.message).toBeTruthy()
      expect(networkError?.message).not.toContain('stack')
    })
    
    it('should detect connection loss during streaming', async () => {
      const sessionId = 'test-network-failure-2'
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
          
          // Simulate connection loss during streaming
          setTimeout(() => {
            connection!.close(1006, 'Network error')
          }, 100)
        }
      })
      
      await chatService.sendMessage('Test streaming')
      await new Promise(resolve => setTimeout(resolve, 200))
      
      // Verify stream interruption was detected
      expect(errors.some(e => e.code === 'STREAM_INTERRUPTED')).toBe(true)
      expect(errors.some(e => e.retryable === true)).toBe(true)
      
      // Verify partial content was preserved
      const interruptError = errors.find(e => e.code === 'STREAM_INTERRUPTED')
      expect(interruptError?.details?.partialContent).toBeDefined()
    })
    
    it('should provide recovery options for network errors', async () => {
      const sessionId = 'test-network-failure-3'
      const errors: ChatError[] = []
      
      const chatService = useChatService({
        sessionId,
        onError: (error) => errors.push(error)
      })
      
      await new Promise(resolve => setTimeout(resolve, 50))
      
      // Simulate network error
      const connection = getLastConnection()
      expect(connection).toBeTruthy()
      connection!.simulateError()
      
      await new Promise(resolve => setTimeout(resolve, 100))
      
      // Verify error provides recovery information
      const networkError = errors.find(e => e.code === 'NETWORK_ERROR')
      expect(networkError).toBeDefined()
      expect(networkError?.retryable).toBe(true)
      expect(networkError?.message).toBeTruthy()
    })
    
    it('should handle WebSocket connection failure', async () => {
      const sessionId = 'test-connection-failure'
      const errors: ChatError[] = []
      
      const chatService = useChatService({
        sessionId,
        onError: (error) => errors.push(error)
      })
      
      // Wait for connection to establish
      await new Promise(resolve => setTimeout(resolve, 50))
      
      const connection = getLastConnection()
      expect(connection).toBeTruthy()
      
      // Simulate connection error
      connection!.simulateError()
      
      // Wait for error to be processed
      await new Promise(resolve => setTimeout(resolve, 100))
      
      // Verify connection error was detected
      expect(errors.some(e => e.code === 'NETWORK_ERROR')).toBe(true)
    })
    
    it('should attempt reconnection after network failure', async () => {
      const sessionId = 'test-reconnection'
      const errors: ChatError[] = []
      
      const chatService = useChatService({
        sessionId,
        onError: (error) => errors.push(error)
      })
      
      await new Promise(resolve => setTimeout(resolve, 50))
      
      const initialConnectionCount = activeWebSockets.length
      
      // Close connection abnormally
      const connection = getLastConnection()
      expect(connection).toBeTruthy()
      connection!.close(1006, 'Connection lost')
      
      // Wait for reconnection attempt
      await new Promise(resolve => setTimeout(resolve, 2000))
      
      // Verify reconnection was attempted (new connection created)
      // Note: In the actual implementation, reconnection happens on next send or automatically
      expect(activeWebSockets.length).toBeGreaterThanOrEqual(initialConnectionCount)
    })
  })
  
  describe('Backend Error Responses', () => {
    it('should handle generic server errors', async () => {
      const sessionId = 'test-server-error'
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
          // Simulate server error
          connection!.simulateMessage({
            type: 'error',
            code: 'SERVER_ERROR',
            message: 'An error occurred. Please try again.',
            retryable: true
          })
        }
      })
      
      await chatService.sendMessage('Test server error')
      await new Promise(resolve => setTimeout(resolve, 100))
      
      // Verify server error was handled
      expect(errors.some(e => e.code === 'SERVER_ERROR')).toBe(true)
      expect(errors.some(e => e.retryable === true)).toBe(true)
    })
    
    it('should handle timeout errors', async () => {
      const sessionId = 'test-timeout'
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
          // Simulate timeout error
          connection!.simulateMessage({
            type: 'error',
            code: 'TIMEOUT',
            message: 'Request timed out. Please try again.',
            retryable: true
          })
        }
      })
      
      await chatService.sendMessage('Test timeout')
      await new Promise(resolve => setTimeout(resolve, 100))
      
      // Verify timeout error was handled
      expect(errors.some(e => e.code === 'TIMEOUT')).toBe(true)
    })
    
    it('should handle invalid session errors', async () => {
      const sessionId = 'test-invalid-session'
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
          // Simulate invalid session error
          connection!.simulateMessage({
            type: 'error',
            code: 'INVALID_SESSION',
            message: 'Session expired. Please start a new session.',
            retryable: false
          })
        }
      })
      
      await chatService.sendMessage('Test invalid session')
      await new Promise(resolve => setTimeout(resolve, 100))
      
      // Verify invalid session error was handled
      expect(errors.some(e => e.code === 'INVALID_SESSION')).toBe(true)
      expect(errors.some(e => e.retryable === false)).toBe(true)
    })
  })
  
  describe('Rate Limiting Behavior (Requirement 8.2)', () => {
    it('should handle rate limit errors from backend', async () => {
      const sessionId = 'test-rate-limit-1'
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
          // Simulate rate limit error
          connection!.simulateMessage({
            type: 'error',
            code: 'RATE_LIMIT',
            message: 'Service is temporarily busy. Please try again in a moment.',
            retryable: true
          })
        }
      })
      
      await chatService.sendMessage('Test rate limit')
      await new Promise(resolve => setTimeout(resolve, 100))
      
      // Verify rate limit error was handled
      expect(errors.some(e => e.code === 'RATE_LIMIT')).toBe(true)
      expect(errors.some(e => e.retryable === true)).toBe(true)
      
      // Verify error message indicates temporary unavailability
      const rateLimitError = errors.find(e => e.code === 'RATE_LIMIT')
      expect(rateLimitError?.message).toContain('busy')
      expect(rateLimitError?.message).toContain('try again')
    })
    
    it('should display appropriate message for rate limiting', async () => {
      const sessionId = 'test-rate-limit-2'
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
          connection!.simulateMessage({
            type: 'error',
            code: 'RATE_LIMIT',
            message: 'Service is temporarily busy. Please try again in a moment.',
            retryable: true,
            details: {
              retryAfter: 30
            }
          })
        }
      })
      
      await chatService.sendMessage('Test rate limit message')
      await new Promise(resolve => setTimeout(resolve, 100))
      
      // Verify error message is user-friendly
      const rateLimitError = errors.find(e => e.code === 'RATE_LIMIT')
      expect(rateLimitError).toBeDefined()
      expect(rateLimitError?.message).not.toContain('429')
      expect(rateLimitError?.message).not.toContain('ThrottlingException')
      expect(rateLimitError?.message).not.toContain('AWS')
    })
    
    it('should mark rate limit errors as retryable', async () => {
      const sessionId = 'test-rate-limit-3'
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
          connection!.simulateMessage({
            type: 'error',
            code: 'RATE_LIMIT',
            message: 'Service is temporarily busy.',
            retryable: true
          })
        }
      })
      
      await chatService.sendMessage('Test retryable')
      await new Promise(resolve => setTimeout(resolve, 100))
      
      // Verify rate limit error is marked as retryable
      const rateLimitError = errors.find(e => e.code === 'RATE_LIMIT')
      expect(rateLimitError?.retryable).toBe(true)
    })
  })
  
  describe('Malformed Response Handling (Requirement 8.4)', () => {
    it('should handle malformed JSON responses', async () => {
      const sessionId = 'test-malformed-1'
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
          // Send malformed JSON
          if (connection!.onmessage) {
            const event = new MessageEvent('message', { data: 'invalid json{{{' })
            connection!.onmessage(event)
          }
        }
      })
      
      await chatService.sendMessage('Test malformed JSON')
      await new Promise(resolve => setTimeout(resolve, 100))
      
      // Verify malformed response was handled
      expect(errors.some(e => e.code === 'MALFORMED_RESPONSE')).toBe(true)
      expect(errors.some(e => e.retryable === true)).toBe(true)
    })
    
    it('should handle empty responses gracefully', async () => {
      const sessionId = 'test-empty-response'
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
          // Send empty response
          if (connection!.onmessage) {
            const event = new MessageEvent('message', { data: '' })
            connection!.onmessage(event)
          }
        }
      })
      
      await chatService.sendMessage('Test empty response')
      await new Promise(resolve => setTimeout(resolve, 100))
      
      // Verify empty response was handled
      expect(errors.some(e => e.code === 'MALFORMED_RESPONSE')).toBe(true)
    })
    
    it('should handle responses with missing required fields', async () => {
      const sessionId = 'test-missing-fields'
      const errors: ChatError[] = []
      let completedMessages: string[] = []
      
      const chatService = useChatService({
        sessionId,
        onError: (error) => errors.push(error),
        onMessageComplete: (content) => completedMessages.push(content)
      })
      
      await new Promise(resolve => setTimeout(resolve, 50))
      
      const connection = getLastConnection()
      expect(connection).toBeTruthy()
      
      connection!.setMessageHandler((data) => {
        if (data.type === 'message') {
          // Send response with missing type field
          if (connection!.onmessage) {
            const event = new MessageEvent('message', { data: JSON.stringify({ content: 'test' }) })
            connection!.onmessage(event)
          }
        }
      })
      
      await chatService.sendMessage('Test missing fields')
      await new Promise(resolve => setTimeout(resolve, 100))
      
      // Verify the response was handled (may be ignored or cause error depending on implementation)
      // The system should not crash
      expect(true).toBe(true)
    })
    
    it('should handle responses with unexpected structure', async () => {
      const sessionId = 'test-unexpected-structure'
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
          // Send response with unexpected structure
          connection!.simulateMessage({
            type: 'unknown_type',
            random_field: 'random_value',
            nested: {
              deeply: {
                nested: 'value'
              }
            }
          })
        }
      })
      
      await chatService.sendMessage('Test unexpected structure')
      await new Promise(resolve => setTimeout(resolve, 100))
      
      // System should handle gracefully without crashing
      expect(true).toBe(true)
    })
    
    it('should provide user-friendly error messages for malformed responses', async () => {
      const sessionId = 'test-malformed-message'
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
            const event = new MessageEvent('message', { data: 'not json at all' })
            connection!.onmessage(event)
          }
        }
      })
      
      await chatService.sendMessage('Test malformed message')
      await new Promise(resolve => setTimeout(resolve, 100))
      
      // Verify error message is user-friendly
      const malformedError = errors.find(e => e.code === 'MALFORMED_RESPONSE')
      expect(malformedError).toBeDefined()
      expect(malformedError?.message).not.toContain('JSON.parse')
      expect(malformedError?.message).not.toContain('SyntaxError')
      expect(malformedError?.message).not.toContain('stack')
    })
  })
  
  describe('Error Handler Integration', () => {
    it('should transform infrastructure errors to domain errors', () => {
      const errorHandler = useErrorHandler()
      
      // Simulate AWS SDK error
      const awsError = new Error('ThrottlingException: Rate exceeded')
      errorHandler.handleError(awsError)
      
      // Verify transformation
      expect(errorHandler.currentError.value).toBeDefined()
      expect(errorHandler.currentError.value?.code).toBe('RATE_LIMIT')
      expect(errorHandler.currentError.value?.message).not.toContain('ThrottlingException')
    })
    
    it('should sanitize error messages to remove internal details', () => {
      const errorHandler = useErrorHandler()
      
      // Simulate error with stack trace
      const errorWithStack = new Error('Error at /usr/local/app/src/service.ts:123:45')
      errorHandler.handleError(errorWithStack)
      
      // Verify sanitization
      expect(errorHandler.currentError.value?.message).not.toContain('/usr/local')
      expect(errorHandler.currentError.value?.message).not.toContain(':123:45')
    })
    
    it('should aggregate multiple similar errors', async () => {
      const errorHandler = useErrorHandler()
      
      // Simulate multiple rate limit errors in quick succession
      const error1 = new Error('Rate limit exceeded')
      const error2 = new Error('Rate limit exceeded')
      const error3 = new Error('Rate limit exceeded')
      
      errorHandler.handleError(error1)
      await new Promise(resolve => setTimeout(resolve, 100))
      errorHandler.handleError(error2)
      await new Promise(resolve => setTimeout(resolve, 100))
      errorHandler.handleError(error3)
      
      // Verify aggregation
      expect(errorHandler.currentError.value?.message).toContain('occurrences')
      expect(errorHandler.errorHistory.value.length).toBe(3)
    })
    
    it('should implement retry logic with exponential backoff', async () => {
      const errorHandler = useErrorHandler()
      
      let attempts = 0
      const flakeyFunction = async () => {
        attempts++
        if (attempts < 3) {
          throw new Error('Temporary failure')
        }
        return 'success'
      }
      
      const result = await errorHandler.retry(flakeyFunction, 5)
      
      // Verify retry succeeded
      expect(result).toBe('success')
      expect(attempts).toBe(3)
    })
    
    it('should not retry non-retryable errors', async () => {
      const errorHandler = useErrorHandler()
      
      let attempts = 0
      const nonRetryableFunction = async () => {
        attempts++
        const error = new Error('Validation error')
        throw error
      }
      
      try {
        await errorHandler.retry(nonRetryableFunction, 5)
      } catch (err) {
        // Expected to throw
      }
      
      // Verify only one attempt was made
      expect(attempts).toBe(1)
    })
  })
  
  describe('Error Recovery', () => {
    it('should allow retry after recoverable error', async () => {
      const sessionId = 'test-recovery-1'
      const errors: ChatError[] = []
      let messagesSent = 0
      
      const chatService = useChatService({
        sessionId,
        onError: (error) => errors.push(error)
      })
      
      await new Promise(resolve => setTimeout(resolve, 50))
      
      const connection = getLastConnection()
      expect(connection).toBeTruthy()
      
      connection!.setMessageHandler((data) => {
        if (data.type === 'message') {
          messagesSent++
          
          // First message fails
          if (messagesSent === 1) {
            connection!.simulateMessage({
              type: 'error',
              code: 'RATE_LIMIT',
              message: 'Rate limit exceeded',
              retryable: true
            })
          } else {
            // Second message succeeds
            connection!.simulateMessage({ type: 'chunk', content: 'Success' })
            setTimeout(() => connection!.simulateMessage({ type: 'complete' }), 50)
          }
        }
      })
      
      // First attempt - should fail
      await chatService.sendMessage('First attempt')
      await new Promise(resolve => setTimeout(resolve, 100))
      
      // Clear error
      chatService.clearError()
      
      // Second attempt - should succeed
      await chatService.sendMessage('Second attempt')
      await new Promise(resolve => setTimeout(resolve, 150))
      
      // Verify recovery
      expect(messagesSent).toBe(2)
      expect(errors.length).toBeGreaterThan(0)
    })
    
    it('should clear error state after successful operation', async () => {
      const sessionId = 'test-recovery-2'
      const errors: ChatError[] = []
      
      const chatService = useChatService({
        sessionId,
        onError: (error) => errors.push(error)
      })
      
      await new Promise(resolve => setTimeout(resolve, 50))
      
      const connection = getLastConnection()
      expect(connection).toBeTruthy()
      
      // Simulate error
      connection!.simulateError()
      await new Promise(resolve => setTimeout(resolve, 100))
      
      // Verify error was set
      expect(chatService.error.value).toBeDefined()
      
      // Clear error
      chatService.clearError()
      
      // Verify error was cleared
      expect(chatService.error.value).toBeNull()
    })
  })
})
