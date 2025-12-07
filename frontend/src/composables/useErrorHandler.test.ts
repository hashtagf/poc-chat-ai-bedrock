import { describe, it, expect, beforeEach } from 'vitest'
import * as fc from 'fast-check'
import { useErrorHandler } from './useErrorHandler'
import type { ChatError } from '@/types'

describe('useErrorHandler', () => {
  describe('Property 12: Error message sanitization', () => {
    // Feature: chat-ui, Property 12: Error message sanitization
    // Validates: Requirements 4.2
    it('should sanitize error messages to remove internal details for any error', () => {
      fc.assert(
        fc.property(
          // Generate various error types with internal details
          fc.oneof(
            // Error with stack trace
            fc.record({
              message: fc.string({ minLength: 5, maxLength: 50 }),
              stack: fc.string()
            }).map(obj => {
              const err = new Error(obj.message)
              err.stack = `Error: ${obj.message}\n    at Object.<anonymous> (/path/to/file.js:10:15)\n    at Module._compile (internal/modules/cjs/loader.js:1063:30)`
              return err
            }),
            // Error with file paths
            fc.string({ minLength: 5, maxLength: 50 }).map(msg => 
              new Error(`${msg} at /usr/local/app/src/components/Chat.vue:42`)
            ),
            // AWS SDK style error
            fc.string({ minLength: 5, maxLength: 50 }).map(msg =>
              new Error(`BedrockRuntimeException: ${msg} (Service: AmazonBedrockRuntime; Status Code: 500)`)
            ),
            // String error with paths
            fc.string({ minLength: 5, maxLength: 50 }).map(msg =>
              `Error: ${msg}\n    at processTicksAndRejections (internal/process/task_queues.js:95:5)`
            )
          ),
          (error) => {
            const handler = useErrorHandler()
            
            // Handle the error
            handler.handleError(error)
            
            // Get the current error
            const currentError = handler.currentError.value
            
            // Verify error was handled
            expect(currentError).not.toBeNull()
            
            if (currentError) {
              // Verify no stack traces in the message
              expect(currentError.message).not.toMatch(/at\s+.*\(.*:\d+:\d+\)/)
              
              // Verify no file paths in the message
              expect(currentError.message).not.toMatch(/\/[^\s]+\/[^\s]+/)
              
              // Verify no AWS SDK internal details
              expect(currentError.message).not.toMatch(/Service:/)
              expect(currentError.message).not.toMatch(/Status Code:/)
              
              // Verify message is user-friendly (not empty and reasonable length)
              expect(currentError.message.length).toBeGreaterThan(0)
              expect(currentError.message.length).toBeLessThan(200)
              
              // Verify it's a proper ChatError with required fields
              expect(currentError.code).toBeDefined()
              expect(typeof currentError.code).toBe('string')
              expect(typeof currentError.retryable).toBe('boolean')
            }
          }
        ),
        { numRuns: 100 }
      )
    })
  })

  describe('Property 19: Infrastructure error transformation', () => {
    // Feature: chat-ui, Property 19: Infrastructure error transformation
    // Validates: Requirements 6.3
    it('should transform infrastructure errors to domain errors for any infrastructure error', () => {
      fc.assert(
        fc.property(
          // Generate various infrastructure error types
          fc.oneof(
            // Network errors
            fc.constantFrom(
              new Error('Network request failed'),
              new Error('fetch failed: connection timeout'),
              new Error('ECONNREFUSED: Connection refused')
            ),
            // AWS/Bedrock errors
            fc.constantFrom(
              new Error('BedrockRuntimeException: Agent unavailable'),
              new Error('AWS.BedrockRuntime.ThrottlingException'),
              new Error('Bedrock knowledge base query failed')
            ),
            // HTTP errors with status codes
            fc.record({
              message: fc.string({ minLength: 5, maxLength: 50 }),
              status: fc.constantFrom(429, 500, 503, 401, 404)
            }).map(obj => {
              const err = new Error(obj.message) as any
              err.status = obj.status
              return err
            }),
            // Timeout errors
            fc.constantFrom(
              new Error('Request timeout exceeded'),
              new Error('Operation timed out after 30s')
            )
          ),
          (infrastructureError) => {
            const handler = useErrorHandler()
            
            // Handle the infrastructure error
            handler.handleError(infrastructureError)
            
            // Get the current error
            const currentError = handler.currentError.value
            
            // Verify error was transformed
            expect(currentError).not.toBeNull()
            
            if (currentError) {
              // Verify it's a domain ChatError, not the raw infrastructure error
              expect(currentError).toHaveProperty('code')
              expect(currentError).toHaveProperty('message')
              expect(currentError).toHaveProperty('retryable')
              
              // Verify the error code is a domain error code, not an infrastructure code
              const domainErrorCodes = [
                'NETWORK_ERROR',
                'CONNECTION_LOST',
                'RATE_LIMIT',
                'SERVER_ERROR',
                'TIMEOUT',
                'INVALID_SESSION',
                'VALIDATION_ERROR',
                'STREAM_INTERRUPTED',
                'KNOWLEDGE_BASE_UNAVAILABLE',
                'AGENT_UNAVAILABLE',
                'INVALID_INPUT',
                'MALFORMED_RESPONSE',
                'UNKNOWN_ERROR'
              ]
              
              expect(domainErrorCodes).toContain(currentError.code)
              
              // Verify the message is user-friendly, not technical infrastructure details
              expect(currentError.message).not.toMatch(/BedrockRuntimeException/)
              expect(currentError.message).not.toMatch(/AWS\./)
              expect(currentError.message).not.toMatch(/ECONNREFUSED/)
              expect(currentError.message).not.toMatch(/ThrottlingException/)
              
              // Verify retryable is set appropriately
              expect(typeof currentError.retryable).toBe('boolean')
            }
          }
        ),
        { numRuns: 100 }
      )
    })
  })

  describe('Property 27: Error aggregation', () => {
    // Feature: chat-ui, Property 27: Error aggregation
    // Validates: Requirements 8.5
    it('should aggregate multiple similar errors within time window', () => {
      fc.assert(
        fc.property(
          // Generate a sequence of similar errors
          fc.record({
            errorType: fc.constantFrom('NETWORK_ERROR', 'RATE_LIMIT', 'SERVER_ERROR', 'TIMEOUT'),
            count: fc.integer({ min: 2, max: 10 })
          }),
          (testData) => {
            const handler = useErrorHandler()
            
            // Create errors of the same type
            const errors: Error[] = []
            for (let i = 0; i < testData.count; i++) {
              switch (testData.errorType) {
                case 'NETWORK_ERROR':
                  errors.push(new Error('Network request failed'))
                  break
                case 'RATE_LIMIT':
                  errors.push(new Error('Rate limit exceeded: 429'))
                  break
                case 'SERVER_ERROR':
                  errors.push(new Error('Internal server error: 500'))
                  break
                case 'TIMEOUT':
                  errors.push(new Error('Request timeout exceeded'))
                  break
              }
            }
            
            // Handle all errors in quick succession (within aggregation window)
            // Since they're handled synchronously, they're all within the time window
            for (const error of errors) {
              handler.handleError(error)
            }
            
            // Get the current error
            const currentError = handler.currentError.value
            
            // Verify error was aggregated
            expect(currentError).not.toBeNull()
            
            if (currentError) {
              // For multiple errors (count >= 2), the system should aggregate them
              // This means the final error should show aggregation
              const hasCountInMessage = currentError.message.includes('occurrences') || 
                                       currentError.message.match(/\(\d+\s+occurrences?\)/)
              const hasCountInDetails = currentError.details?.count !== undefined && 
                                       currentError.details.count >= 2
              
              // When we have 2+ errors of the same type, aggregation should occur
              // The final error should indicate this via count in details or message
              expect(hasCountInMessage || hasCountInDetails).toBe(true)
              
              // If count is in details, verify it's at least 2 (since we have multiple errors)
              if (hasCountInDetails) {
                expect(currentError.details?.count).toBeGreaterThanOrEqual(2)
                expect(currentError.details?.count).toBeLessThanOrEqual(testData.count)
              }
              
              // Verify error history contains all errors
              expect(handler.errorHistory.value.length).toBe(testData.count)
            }
          }
        ),
        { numRuns: 100 }
      )
    })
  })

  describe('Basic functionality tests', () => {
    let handler: ReturnType<typeof useErrorHandler>

    beforeEach(() => {
      handler = useErrorHandler()
    })

    it('should start with no current error', () => {
      expect(handler.currentError.value).toBeNull()
    })

    it('should start with empty error history', () => {
      expect(handler.errorHistory.value).toEqual([])
    })

    it('should handle a simple error', () => {
      const error = new Error('Test error')
      handler.handleError(error)
      
      expect(handler.currentError.value).not.toBeNull()
      expect(handler.errorHistory.value.length).toBe(1)
    })

    it('should clear current error', () => {
      const error = new Error('Test error')
      handler.handleError(error)
      
      expect(handler.currentError.value).not.toBeNull()
      
      handler.clearError()
      expect(handler.currentError.value).toBeNull()
    })

    it('should transform network errors correctly', () => {
      const error = new Error('Network request failed')
      handler.handleError(error)
      
      const currentError = handler.currentError.value
      expect(currentError?.code).toBe('NETWORK_ERROR')
      expect(currentError?.retryable).toBe(true)
    })

    it('should transform rate limit errors correctly', () => {
      const error = new Error('Rate limit exceeded: 429')
      handler.handleError(error)
      
      const currentError = handler.currentError.value
      expect(currentError?.code).toBe('RATE_LIMIT')
      expect(currentError?.retryable).toBe(true)
    })

    it('should transform timeout errors correctly', () => {
      const error = new Error('Request timeout exceeded')
      handler.handleError(error)
      
      const currentError = handler.currentError.value
      expect(currentError?.code).toBe('TIMEOUT')
      expect(currentError?.retryable).toBe(true)
    })

    it('should handle validation errors as non-retryable', () => {
      const error = new Error('Validation failed: invalid input')
      handler.handleError(error)
      
      const currentError = handler.currentError.value
      expect(currentError?.code).toBe('VALIDATION_ERROR')
      expect(currentError?.retryable).toBe(false)
    })

    it('should retry a function with exponential backoff', async () => {
      let attempts = 0
      const fn = async () => {
        attempts++
        if (attempts < 3) {
          throw new Error('Temporary failure')
        }
        return 'success'
      }

      const result = await handler.retry(fn, 5)
      
      expect(result).toBe('success')
      expect(attempts).toBe(3)
    })

    it('should stop retrying after max attempts', async () => {
      const fn = async () => {
        throw new Error('Persistent failure')
      }

      await expect(handler.retry(fn, 3)).rejects.toThrow('Persistent failure')
    })

    it('should not retry non-retryable errors', async () => {
      let attempts = 0
      const fn = async () => {
        attempts++
        throw new Error('Validation failed: invalid input')
      }

      await expect(handler.retry(fn, 5)).rejects.toThrow()
      expect(attempts).toBe(1) // Should only try once
    })
  })
})
