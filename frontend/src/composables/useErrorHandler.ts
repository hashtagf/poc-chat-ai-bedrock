import { ref, type Ref } from 'vue'
import type { ChatError } from '@/types'

/**
 * Error handler composable for managing error state and transformations
 * Handles error sanitization, aggregation, and retry logic
 */

export interface ErrorHandler {
  currentError: Ref<ChatError | null>
  errorHistory: Ref<ChatError[]>
  handleError(error: unknown): void
  clearError(): void
  retry<T>(fn: () => Promise<T>, maxRetries?: number): Promise<T>
}

// Time window for error aggregation (10 seconds)
const ERROR_AGGREGATION_WINDOW_MS = 10000

// Error message mappings for different error types
const ERROR_MESSAGES: Record<string, string> = {
  NETWORK_ERROR: 'Unable to connect. Please check your network connection.',
  CONNECTION_LOST: 'Connection lost. Attempting to reconnect...',
  RATE_LIMIT: 'Service is temporarily busy. Please try again in a moment.',
  SERVER_ERROR: 'An error occurred. Please try again.',
  TIMEOUT: 'Request timed out. Please try again.',
  INVALID_SESSION: 'Session expired. Please start a new session.',
  VALIDATION_ERROR: 'Invalid input. Please check your message and try again.',
  STREAM_INTERRUPTED: 'Response incomplete. Please try again.',
  KNOWLEDGE_BASE_UNAVAILABLE: 'Knowledge base temporarily unavailable. Responses may be limited.',
  AGENT_UNAVAILABLE: 'AI agent is currently unavailable. Please try again later.',
  INVALID_INPUT: 'Unable to process message. Please rephrase and try again.',
  MALFORMED_RESPONSE: 'Received an invalid response. Please try again.',
  UNKNOWN_ERROR: 'An unexpected error occurred. Please try again.'
}

/**
 * Sanitize error messages to remove internal details
 * Prevents exposure of stack traces, AWS SDK errors, or system paths
 */
function sanitizeErrorMessage(error: unknown): string {
  if (typeof error === 'string') {
    // Remove common internal patterns
    return error
      .replace(/at\s+.*\(.*:\d+:\d+\)/g, '') // Stack trace lines
      .replace(/\/[^\s]+\/[^\s]+/g, '') // File paths
      .replace(/Error:\s*/g, '')
      .trim()
  }

  if (error instanceof Error) {
    // Extract just the message, no stack trace
    return error.message
      .replace(/at\s+.*\(.*:\d+:\d+\)/g, '')
      .replace(/\/[^\s]+\/[^\s]+/g, '')
      .trim()
  }

  return 'An error occurred'
}

/**
 * Transform infrastructure errors to domain errors
 * Maps AWS SDK and network errors to user-friendly ChatError objects
 */
function transformError(error: unknown): ChatError {
  // Handle ChatError objects that are already transformed
  if (isChatError(error)) {
    return error
  }

  // Handle Error objects
  if (error instanceof Error) {
    const message = error.message.toLowerCase()
    
    // Network errors
    if (message.includes('network') || message.includes('fetch')) {
      return {
        code: 'NETWORK_ERROR',
        message: ERROR_MESSAGES.NETWORK_ERROR,
        retryable: true
      }
    }
    
    // Timeout errors
    if (message.includes('timeout') || message.includes('timed out')) {
      return {
        code: 'TIMEOUT',
        message: ERROR_MESSAGES.TIMEOUT,
        retryable: true
      }
    }
    
    // Rate limit errors
    if (message.includes('rate limit') || message.includes('429') || message.includes('throttl')) {
      return {
        code: 'RATE_LIMIT',
        message: ERROR_MESSAGES.RATE_LIMIT,
        retryable: true
      }
    }
    
    // Server errors
    if (message.includes('500') || message.includes('server error')) {
      return {
        code: 'SERVER_ERROR',
        message: ERROR_MESSAGES.SERVER_ERROR,
        retryable: true
      }
    }
    
    // Session errors
    if (message.includes('session') || message.includes('401') || message.includes('unauthorized')) {
      return {
        code: 'INVALID_SESSION',
        message: ERROR_MESSAGES.INVALID_SESSION,
        retryable: false
      }
    }
    
    // Validation errors
    if (message.includes('validation') || message.includes('invalid')) {
      return {
        code: 'VALIDATION_ERROR',
        message: ERROR_MESSAGES.VALIDATION_ERROR,
        retryable: false
      }
    }
    
    // AWS/Bedrock specific errors
    if (message.includes('bedrock') || message.includes('aws')) {
      if (message.includes('knowledge base')) {
        return {
          code: 'KNOWLEDGE_BASE_UNAVAILABLE',
          message: ERROR_MESSAGES.KNOWLEDGE_BASE_UNAVAILABLE,
          retryable: true
        }
      }
      
      return {
        code: 'AGENT_UNAVAILABLE',
        message: ERROR_MESSAGES.AGENT_UNAVAILABLE,
        retryable: true
      }
    }
  }

  // Handle HTTP response errors
  if (typeof error === 'object' && error !== null) {
    const err = error as any
    
    if (err.status === 429 || err.statusCode === 429) {
      return {
        code: 'RATE_LIMIT',
        message: ERROR_MESSAGES.RATE_LIMIT,
        retryable: true
      }
    }
    
    if (err.status === 500 || err.statusCode === 500) {
      return {
        code: 'SERVER_ERROR',
        message: ERROR_MESSAGES.SERVER_ERROR,
        retryable: true
      }
    }
  }

  // Default unknown error
  return {
    code: 'UNKNOWN_ERROR',
    message: ERROR_MESSAGES.UNKNOWN_ERROR,
    retryable: true
  }
}

/**
 * Type guard for ChatError
 */
function isChatError(value: unknown): value is ChatError {
  if (typeof value !== 'object' || value === null) {
    return false
  }
  
  const obj = value as Record<string, unknown>
  
  return (
    typeof obj.code === 'string' &&
    typeof obj.message === 'string' &&
    typeof obj.retryable === 'boolean'
  )
}

/**
 * Check if two errors should be aggregated
 * Errors are aggregated if they have the same code and occur within the time window
 */
function shouldAggregateErrors(error1: ChatError, error2: ChatError, timeWindow: number): boolean {
  if (error1.code !== error2.code) {
    return false
  }
  
  const timestamp1 = (error1.details?.timestamp as number) || 0
  const timestamp2 = (error2.details?.timestamp as number) || 0
  
  return Math.abs(timestamp2 - timestamp1) <= timeWindow
}

export function useErrorHandler(): ErrorHandler {
  const currentError: Ref<ChatError | null> = ref(null)
  const errorHistory: Ref<ChatError[]> = ref([])

  /**
   * Handle an error by transforming and sanitizing it
   * Implements error aggregation for multiple errors within time window
   */
  const handleError = (error: unknown): void => {
    // Transform the error to a domain error
    const transformedError = transformError(error)
    
    // Add timestamp to error details
    const errorWithTimestamp: ChatError = {
      ...transformedError,
      details: {
        ...transformedError.details,
        timestamp: Date.now()
      }
    }
    
    // Check if we should aggregate with recent errors
    const recentErrors = errorHistory.value.filter(err => {
      const errTimestamp = (err.details?.timestamp as number) || 0
      const currentTimestamp = Date.now()
      return currentTimestamp - errTimestamp <= ERROR_AGGREGATION_WINDOW_MS
    })
    
    // Check if there's a similar recent error
    const similarError = recentErrors.find(err => 
      shouldAggregateErrors(err, errorWithTimestamp, ERROR_AGGREGATION_WINDOW_MS)
    )
    
    if (similarError && recentErrors.length > 0) {
      // Aggregate: update the error message to indicate multiple occurrences
      const count = recentErrors.filter(err => err.code === errorWithTimestamp.code).length + 1
      currentError.value = {
        ...errorWithTimestamp,
        message: `${errorWithTimestamp.message} (${count} occurrences)`,
        details: {
          ...errorWithTimestamp.details,
          count
        }
      }
    } else {
      // New error, not aggregated
      currentError.value = errorWithTimestamp
    }
    
    // Add to error history
    errorHistory.value.push(errorWithTimestamp)
    
    // Limit error history size to prevent memory issues
    if (errorHistory.value.length > 100) {
      errorHistory.value = errorHistory.value.slice(-50)
    }
  }

  /**
   * Clear the current error state
   */
  const clearError = (): void => {
    currentError.value = null
  }

  /**
   * Retry a function with exponential backoff
   * Implements retry logic for retryable errors
   */
  const retry = async <T>(
    fn: () => Promise<T>,
    maxRetries: number = 5
  ): Promise<T> => {
    let lastError: unknown
    
    for (let attempt = 0; attempt < maxRetries; attempt++) {
      try {
        return await fn()
      } catch (error) {
        lastError = error
        
        // Transform error to check if it's retryable
        const transformedError = transformError(error)
        
        if (!transformedError.retryable || attempt === maxRetries - 1) {
          // Not retryable or last attempt, throw the error
          throw error
        }
        
        // Calculate exponential backoff: 1s, 2s, 4s, 8s, 16s (max 30s)
        const backoffMs = Math.min(1000 * Math.pow(2, attempt), 30000)
        
        // Wait before retrying
        await new Promise(resolve => setTimeout(resolve, backoffMs))
      }
    }
    
    // Should never reach here, but throw last error just in case
    throw lastError
  }

  return {
    currentError,
    errorHistory,
    handleError,
    clearError,
    retry
  }
}
