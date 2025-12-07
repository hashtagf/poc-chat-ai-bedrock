import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount } from '@vue/test-utils'
import * as fc from 'fast-check'
import ErrorDisplay from './ErrorDisplay.vue'
import type { ChatError } from '@/types'

describe('ErrorDisplay', () => {
  beforeEach(() => {
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.restoreAllMocks()
    vi.useRealTimers()
  })

  /**
   * Feature: chat-ui, Property 13: Connection status indication
   * Validates: Requirements 4.4
   * 
   * For any state where the connection to the backend is unavailable,
   * the Chat UI should display a connection status indicator.
   */
  describe('Property 13: Connection status indication', () => {
    it('should display connection status indicator when connection is not connected', () => {
      fc.assert(
        fc.property(
          fc.constantFrom('disconnected' as const, 'connecting' as const),
          (connectionStatus) => {
            // Mount component with non-connected status
            const wrapper = mount(ErrorDisplay, {
              props: {
                error: null,
                connectionStatus
              }
            })

            // The connection status indicator should be visible
            const statusElement = wrapper.find('.connection-status')
            expect(statusElement.exists()).toBe(true)

            // The status text should be present
            const statusText = wrapper.find('.connection-text')
            expect(statusText.exists()).toBe(true)
            expect(statusText.text().length).toBeGreaterThan(0)

            // Verify the text matches the connection status
            if (connectionStatus === 'disconnected') {
              expect(statusText.text()).toBe('Disconnected')
            } else if (connectionStatus === 'connecting') {
              expect(statusText.text()).toBe('Connecting...')
            }

            wrapper.unmount()
          }
        ),
        { numRuns: 100 }
      )
    })

    it('should not display connection status indicator when connected', () => {
      fc.assert(
        fc.property(
          fc.constant('connected' as const),
          (connectionStatus) => {
            // Mount component with connected status
            const wrapper = mount(ErrorDisplay, {
              props: {
                error: null,
                connectionStatus
              }
            })

            // The connection status indicator should NOT be visible
            const statusElement = wrapper.find('.connection-status')
            expect(statusElement.exists()).toBe(false)

            wrapper.unmount()
          }
        ),
        { numRuns: 100 }
      )
    })
  })

  /**
   * Feature: chat-ui, Property 25: Rate limit error handling
   * Validates: Requirements 8.2
   * 
   * For any rate limit error from the Bedrock API, the Chat UI should
   * display a message indicating temporary unavailability.
   */
  describe('Property 25: Rate limit error handling', () => {
    it('should display appropriate message for rate limit errors', () => {
      fc.assert(
        fc.property(
          fc.record({
            code: fc.constant('RATE_LIMIT'),
            message: fc.string({ minLength: 1 }).filter(s => s.trim().length > 0),
            retryable: fc.constant(true),
            details: fc.option(fc.dictionary(fc.string(), fc.anything()), { nil: undefined })
          }),
          (error: ChatError) => {
            // Mount component with rate limit error
            const wrapper = mount(ErrorDisplay, {
              props: {
                error,
                connectionStatus: 'connected'
              }
            })

            // Error should be visible
            const errorContainer = wrapper.find('.error-message-container')
            expect(errorContainer.exists()).toBe(true)

            // Error message should be displayed
            const errorMessage = wrapper.find('.error-message')
            expect(errorMessage.exists()).toBe(true)
            // HTML naturally trims whitespace, so compare trimmed values
            expect(errorMessage.text().trim()).toBe(error.message.trim())

            // Error code should indicate rate limit
            const errorCode = wrapper.find('.error-code')
            expect(errorCode.exists()).toBe(true)
            expect(errorCode.text()).toContain('RATE_LIMIT')

            // Retry button should be available (rate limit errors are retryable)
            const retryButton = wrapper.find('.retry-button')
            expect(retryButton.exists()).toBe(true)

            wrapper.unmount()
          }
        ),
        { numRuns: 100 }
      )
    })
  })

  /**
   * Feature: chat-ui, Property 26: Malformed response handling
   * Validates: Requirements 8.4
   * 
   * For any empty or malformed response from the agent, the Chat UI should
   * handle it gracefully without crashing and display appropriate user messaging.
   */
  describe('Property 26: Malformed response handling', () => {
    it('should handle malformed response errors gracefully', () => {
      fc.assert(
        fc.property(
          fc.record({
            code: fc.constantFrom('MALFORMED_RESPONSE', 'INVALID_INPUT', 'UNKNOWN_ERROR'),
            message: fc.string({ minLength: 1 }).filter(s => s.trim().length > 0),
            retryable: fc.boolean(),
            details: fc.option(fc.dictionary(fc.string(), fc.anything()), { nil: undefined })
          }),
          (error: ChatError) => {
            // Mount component with malformed response error
            const wrapper = mount(ErrorDisplay, {
              props: {
                error,
                connectionStatus: 'connected'
              }
            })

            // Component should not crash and should render
            expect(wrapper.exists()).toBe(true)

            // Error should be visible
            const errorContainer = wrapper.find('.error-message-container')
            expect(errorContainer.exists()).toBe(true)

            // Error message should be displayed (user-friendly)
            const errorMessage = wrapper.find('.error-message')
            expect(errorMessage.exists()).toBe(true)
            // HTML naturally trims whitespace, so compare trimmed values
            expect(errorMessage.text().trim()).toBe(error.message.trim())

            // Dismiss button should always be available
            const dismissButton = wrapper.find('.dismiss-button')
            expect(dismissButton.exists()).toBe(true)

            // Retry button should be available only if error is retryable
            const retryButton = wrapper.find('.retry-button')
            if (error.retryable) {
              expect(retryButton.exists()).toBe(true)
            } else {
              expect(retryButton.exists()).toBe(false)
            }

            wrapper.unmount()
          }
        ),
        { numRuns: 100 }
      )
    })
  })

  // Additional unit tests for component behavior
  describe('Unit Tests', () => {
    it('should emit retry event when retry button is clicked', async () => {
      const error: ChatError = {
        code: 'NETWORK_ERROR',
        message: 'Network error occurred',
        retryable: true
      }

      const wrapper = mount(ErrorDisplay, {
        props: {
          error,
          connectionStatus: 'connected'
        }
      })

      const retryButton = wrapper.find('.retry-button')
      await retryButton.trigger('click')

      expect(wrapper.emitted('retry')).toBeTruthy()
      expect(wrapper.emitted('retry')?.length).toBe(1)

      wrapper.unmount()
    })

    it('should emit dismiss event when dismiss button is clicked', async () => {
      const error: ChatError = {
        code: 'SERVER_ERROR',
        message: 'Server error occurred',
        retryable: true
      }

      const wrapper = mount(ErrorDisplay, {
        props: {
          error,
          connectionStatus: 'connected'
        }
      })

      const dismissButton = wrapper.find('.dismiss-button')
      await dismissButton.trigger('click')

      expect(wrapper.emitted('dismiss')).toBeTruthy()
      expect(wrapper.emitted('dismiss')?.length).toBe(1)

      wrapper.unmount()
    })

    it('should auto-dismiss non-critical errors after 5 seconds', async () => {
      const error: ChatError = {
        code: 'TIMEOUT',
        message: 'Request timed out',
        retryable: true
      }

      const wrapper = mount(ErrorDisplay, {
        props: {
          error,
          connectionStatus: 'connected'
        }
      })

      // Error should be visible initially
      expect(wrapper.find('.error-message-container').exists()).toBe(true)

      // Fast-forward time by 5 seconds
      vi.advanceTimersByTime(5000)
      await wrapper.vm.$nextTick()

      // Dismiss event should be emitted
      expect(wrapper.emitted('dismiss')).toBeTruthy()

      wrapper.unmount()
    })

    it('should not auto-dismiss critical errors', async () => {
      const error: ChatError = {
        code: 'INVALID_SESSION',
        message: 'Session expired',
        retryable: false
      }

      const wrapper = mount(ErrorDisplay, {
        props: {
          error,
          connectionStatus: 'connected'
        }
      })

      // Error should be visible initially
      expect(wrapper.find('.error-message-container').exists()).toBe(true)

      // Fast-forward time by 10 seconds (more than auto-dismiss timeout)
      vi.advanceTimersByTime(10000)
      await wrapper.vm.$nextTick()

      // Dismiss event should NOT be emitted
      expect(wrapper.emitted('dismiss')).toBeFalsy()

      // Error should still be visible
      expect(wrapper.find('.error-message-container').exists()).toBe(true)

      wrapper.unmount()
    })

    it('should display auto-dismiss indicator for non-critical errors', () => {
      const error: ChatError = {
        code: 'TIMEOUT',
        message: 'Request timed out',
        retryable: true
      }

      const wrapper = mount(ErrorDisplay, {
        props: {
          error,
          connectionStatus: 'connected'
        }
      })

      // Auto-dismiss indicator should be visible
      const indicator = wrapper.find('.auto-dismiss-indicator')
      expect(indicator.exists()).toBe(true)

      wrapper.unmount()
    })

    it('should not display auto-dismiss indicator for critical errors', () => {
      const error: ChatError = {
        code: 'AGENT_UNAVAILABLE',
        message: 'AI agent is unavailable',
        retryable: true
      }

      const wrapper = mount(ErrorDisplay, {
        props: {
          error,
          connectionStatus: 'connected'
        }
      })

      // Auto-dismiss indicator should NOT be visible
      const indicator = wrapper.find('.auto-dismiss-indicator')
      expect(indicator.exists()).toBe(false)

      wrapper.unmount()
    })

    it('should clear auto-dismiss timer when error is dismissed manually', async () => {
      const error: ChatError = {
        code: 'TIMEOUT',
        message: 'Request timed out',
        retryable: true
      }

      const wrapper = mount(ErrorDisplay, {
        props: {
          error,
          connectionStatus: 'connected'
        }
      })

      // Manually dismiss the error
      const dismissButton = wrapper.find('.dismiss-button')
      await dismissButton.trigger('click')

      // Fast-forward time
      vi.advanceTimersByTime(10000)
      await wrapper.vm.$nextTick()

      // Dismiss should only be emitted once (from manual dismiss, not auto-dismiss)
      expect(wrapper.emitted('dismiss')?.length).toBe(1)

      wrapper.unmount()
    })
  })
})
