import { ref, type Ref } from 'vue'
import type { ChatService, ChatError } from '@/types'

/**
 * Composable for managing chat service with WebSocket connection
 * Handles message sending, streaming responses, and connection management
 */

// WebSocket connection states
type ConnectionStatus = 'disconnected' | 'connecting' | 'connected' | 'reconnecting'

// Configuration
const WEBSOCKET_URL = import.meta.env.VITE_WEBSOCKET_URL || 'ws://localhost:8080/ws'
const MAX_RECONNECT_ATTEMPTS = 5
const INITIAL_RECONNECT_DELAY = 1000 // 1 second
const MAX_RECONNECT_DELAY = 30000 // 30 seconds

export interface ChatServiceOptions {
  sessionId: string
  onMessageComplete?: (content: string) => void
  onError?: (error: ChatError) => void
}

export function useChatService(options: ChatServiceOptions): ChatService {
  const { sessionId, onMessageComplete, onError } = options

  // State
  const streamingMessage: Ref<string> = ref('')
  const isStreaming: Ref<boolean> = ref(false)
  const error: Ref<ChatError | null> = ref(null)
  const connectionStatus: Ref<ConnectionStatus> = ref('disconnected')
  
  // WebSocket instance
  let ws: WebSocket | null = null
  let reconnectAttempts = 0
  let reconnectTimeout: ReturnType<typeof setTimeout> | null = null

  /**
   * Calculate exponential backoff delay for reconnection
   */
  const getReconnectDelay = (): number => {
    const delay = INITIAL_RECONNECT_DELAY * Math.pow(2, reconnectAttempts)
    return Math.min(delay, MAX_RECONNECT_DELAY)
  }

  /**
   * Clear the current error state
   */
  const clearError = (): void => {
    error.value = null
  }

  /**
   * Handle WebSocket errors
   */
  const handleWebSocketError = (): void => {
    const chatError: ChatError = {
      code: 'NETWORK_ERROR',
      message: 'Unable to connect. Please check your network connection.',
      retryable: true,
      details: {
        timestamp: Date.now(),
        event: 'websocket_error'
      }
    }
    
    error.value = chatError
    
    if (onError) {
      onError(chatError)
    }
  }

  /**
   * Handle WebSocket connection close
   */
  const handleWebSocketClose = (event: CloseEvent): void => {
    connectionStatus.value = 'disconnected'
    
    // If streaming was in progress, mark it as interrupted
    if (isStreaming.value) {
      isStreaming.value = false
      
      const chatError: ChatError = {
        code: 'STREAM_INTERRUPTED',
        message: 'Response incomplete. Please try again.',
        retryable: true,
        details: {
          timestamp: Date.now(),
          partialContent: streamingMessage.value
        }
      }
      
      error.value = chatError
      
      if (onError) {
        onError(chatError)
      }
    }
    
    // Attempt reconnection if not a normal closure
    if (event.code !== 1000 && reconnectAttempts < MAX_RECONNECT_ATTEMPTS) {
      connectionStatus.value = 'reconnecting'
      const delay = getReconnectDelay()
      
      reconnectTimeout = setTimeout(() => {
        reconnectAttempts++
        connect()
      }, delay)
    } else if (reconnectAttempts >= MAX_RECONNECT_ATTEMPTS) {
      const chatError: ChatError = {
        code: 'CONNECTION_LOST',
        message: 'Connection lost. Please refresh the page.',
        retryable: false,
        details: {
          timestamp: Date.now(),
          attempts: reconnectAttempts
        }
      }
      
      error.value = chatError
      
      if (onError) {
        onError(chatError)
      }
    }
  }

  /**
   * Handle incoming WebSocket messages
   */
  const handleWebSocketMessage = (event: MessageEvent): void => {
    try {
      const data = JSON.parse(event.data)
      
      // Handle different message types
      if (data.type === 'chunk') {
        // Streaming chunk received
        isStreaming.value = true
        streamingMessage.value += data.content
      } else if (data.type === 'complete') {
        // Streaming complete
        const finalContent = streamingMessage.value
        isStreaming.value = false
        streamingMessage.value = ''
        
        // Reset reconnect attempts on successful completion
        reconnectAttempts = 0
        
        if (onMessageComplete) {
          onMessageComplete(finalContent)
        }
      } else if (data.type === 'error') {
        // Error from backend
        isStreaming.value = false
        
        const chatError: ChatError = {
          code: data.code || 'SERVER_ERROR',
          message: data.message || 'An error occurred. Please try again.',
          retryable: data.retryable !== false,
          details: {
            timestamp: Date.now(),
            ...data.details
          }
        }
        
        error.value = chatError
        
        if (onError) {
          onError(chatError)
        }
      }
    } catch (err) {
      // Malformed message
      const chatError: ChatError = {
        code: 'MALFORMED_RESPONSE',
        message: 'Received an invalid response. Please try again.',
        retryable: true,
        details: {
          timestamp: Date.now(),
          rawData: event.data
        }
      }
      
      error.value = chatError
      
      if (onError) {
        onError(chatError)
      }
    }
  }

  /**
   * Establish WebSocket connection
   */
  const connect = (): void => {
    if (ws && (ws.readyState === WebSocket.CONNECTING || ws.readyState === WebSocket.OPEN)) {
      return // Already connected or connecting
    }
    
    connectionStatus.value = 'connecting'
    
    try {
      ws = new WebSocket(`${WEBSOCKET_URL}?sessionId=${sessionId}`)
      
      ws.onopen = () => {
        connectionStatus.value = 'connected'
        reconnectAttempts = 0
        clearError()
      }
      
      ws.onerror = handleWebSocketError
      ws.onclose = handleWebSocketClose
      ws.onmessage = handleWebSocketMessage
    } catch (err) {
      const chatError: ChatError = {
        code: 'NETWORK_ERROR',
        message: 'Unable to connect. Please check your network connection.',
        retryable: true,
        details: {
          timestamp: Date.now(),
          error: err
        }
      }
      
      error.value = chatError
      
      if (onError) {
        onError(chatError)
      }
    }
  }

  /**
   * Disconnect WebSocket
   */
  const disconnect = (): void => {
    if (reconnectTimeout) {
      clearTimeout(reconnectTimeout)
      reconnectTimeout = null
    }
    
    if (ws) {
      ws.close(1000, 'Normal closure')
      ws = null
    }
    
    connectionStatus.value = 'disconnected'
  }

  /**
   * Send a message through the WebSocket connection
   * Validates input and handles connection state
   */
  const sendMessage = async (content: string): Promise<void> => {
    // Validate input
    const trimmedContent = content.trim()
    
    if (trimmedContent.length === 0) {
      const validationError: ChatError = {
        code: 'VALIDATION_ERROR',
        message: 'Message cannot be empty',
        retryable: false,
        details: {
          timestamp: Date.now()
        }
      }
      
      error.value = validationError
      
      if (onError) {
        onError(validationError)
      }
      
      throw new Error('Message cannot be empty')
    }
    
    if (trimmedContent.length > 2000) {
      const validationError: ChatError = {
        code: 'VALIDATION_ERROR',
        message: 'Message exceeds maximum length of 2000 characters',
        retryable: false,
        details: {
          timestamp: Date.now(),
          length: trimmedContent.length
        }
      }
      
      error.value = validationError
      
      if (onError) {
        onError(validationError)
      }
      
      throw new Error('Message too long')
    }
    
    // Check if already streaming
    if (isStreaming.value) {
      const streamingError: ChatError = {
        code: 'STREAMING_IN_PROGRESS',
        message: 'Please wait for the current response to complete',
        retryable: false,
        details: {
          timestamp: Date.now()
        }
      }
      
      error.value = streamingError
      
      if (onError) {
        onError(streamingError)
      }
      
      throw new Error('Streaming in progress')
    }
    
    // Ensure connection is established
    if (!ws || ws.readyState !== WebSocket.OPEN) {
      connect()
      
      // Wait for connection to be established
      await new Promise<void>((resolve, reject) => {
        const checkConnection = setInterval(() => {
          if (ws && ws.readyState === WebSocket.OPEN) {
            clearInterval(checkConnection)
            resolve()
          } else if (connectionStatus.value === 'disconnected' && reconnectAttempts >= MAX_RECONNECT_ATTEMPTS) {
            clearInterval(checkConnection)
            reject(new Error('Failed to establish connection'))
          }
        }, 100)
        
        // Timeout after 10 seconds
        setTimeout(() => {
          clearInterval(checkConnection)
          reject(new Error('Connection timeout'))
        }, 10000)
      })
    }
    
    // Send the message
    try {
      const messagePayload = {
        type: 'message',
        content: trimmedContent,
        sessionId,
        timestamp: Date.now()
      }
      
      ws!.send(JSON.stringify(messagePayload))
      
      // Clear any previous errors
      clearError()
      
      // Reset streaming state
      streamingMessage.value = ''
    } catch (err) {
      const sendError: ChatError = {
        code: 'NETWORK_ERROR',
        message: 'Failed to send message. Please try again.',
        retryable: true,
        details: {
          timestamp: Date.now(),
          error: err
        }
      }
      
      error.value = sendError
      
      if (onError) {
        onError(sendError)
      }
      
      throw err
    }
  }

  // Initialize connection
  connect()

  // Cleanup on unmount
  if (typeof window !== 'undefined') {
    window.addEventListener('beforeunload', disconnect)
  }

  return {
    sendMessage,
    streamingMessage,
    isStreaming,
    error,
    clearError
  }
}
