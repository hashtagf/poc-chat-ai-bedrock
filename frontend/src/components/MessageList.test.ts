import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import * as fc from 'fast-check'
import MessageList from './MessageList.vue'
import type { Message, MessageRole, MessageStatus } from '../types'

// Generators for property-based testing

const messageRoleArb = fc.constantFrom<MessageRole>('user', 'agent')
const messageStatusArb = fc.constantFrom<MessageStatus>('sending', 'sent', 'error')

const messageArb: fc.Arbitrary<Message> = fc.record({
  id: fc.uuid(),
  role: messageRoleArb,
  content: fc.string({ minLength: 1, maxLength: 500 }),
  timestamp: fc.date(),
  citations: fc.constant(undefined), // Simplified for these tests
  status: messageStatusArb,
  errorMessage: fc.option(fc.string({ minLength: 1, maxLength: 100 }))
})

// Helper to create a sequence of messages with increasing timestamps
const messageSequenceArb = (minLength: number, maxLength: number): fc.Arbitrary<Message[]> => {
  return fc.array(messageArb, { minLength, maxLength }).map(messages => {
    // Sort by timestamp to ensure chronological order
    return messages
      .map((msg, index) => ({
        ...msg,
        timestamp: new Date(Date.now() + index * 1000) // Each message 1 second apart
      }))
      .sort((a, b) => a.timestamp.getTime() - b.timestamp.getTime())
  })
}

describe('MessageList', () => {
  // Mock scrollTo and scroll properties
  let mockScrollTo: ReturnType<typeof vi.fn>
  let mockScrollHeight: number
  let mockScrollTop: number
  let mockClientHeight: number

  beforeEach(() => {
    mockScrollTo = vi.fn()
    mockScrollHeight = 1000
    mockScrollTop = 0
    mockClientHeight = 500

    // Mock HTMLElement.prototype.scrollTo
    Object.defineProperty(HTMLElement.prototype, 'scrollTo', {
      configurable: true,
      value: mockScrollTo
    })

    // Mock scroll properties
    Object.defineProperty(HTMLElement.prototype, 'scrollHeight', {
      configurable: true,
      get() {
        return mockScrollHeight
      }
    })

    Object.defineProperty(HTMLElement.prototype, 'scrollTop', {
      configurable: true,
      get() {
        return mockScrollTop
      },
      set(value: number) {
        mockScrollTop = value
      }
    })

    Object.defineProperty(HTMLElement.prototype, 'clientHeight', {
      configurable: true,
      get() {
        return mockClientHeight
      }
    })
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  describe('Property 10: Auto-scroll on new message', () => {
    // Feature: chat-ui, Property 10: Auto-scroll on new message
    // Validates: Requirements 3.2
    it('should automatically scroll to bottom when a new message is added and user is at or near bottom', async () => {
      await fc.assert(
        fc.asyncProperty(
          messageSequenceArb(1, 5),
          messageArb,
          async (initialMessages, newMessage) => {
            // Ensure new message has a later timestamp
            const lastTimestamp = initialMessages[initialMessages.length - 1]?.timestamp.getTime() || Date.now()
            const messageToAdd: Message = {
              ...newMessage,
              timestamp: new Date(lastTimestamp + 1000)
            }

            // Mount with initial messages
            const wrapper = mount(MessageList, {
              props: {
                messages: initialMessages
              }
            })

            await flushPromises()

            // Reset mock to track new calls
            mockScrollTo.mockClear()

            // Simulate user being at bottom (scrollTop + clientHeight >= scrollHeight - threshold)
            mockScrollTop = mockScrollHeight - mockClientHeight // Exactly at bottom
            
            // Add new message by updating props
            await wrapper.setProps({
              messages: [...initialMessages, messageToAdd]
            })

            await flushPromises()

            // Should have called scrollTo to scroll to bottom
            expect(mockScrollTo).toHaveBeenCalled()
            
            // Verify it scrolled to the bottom (scrollTop should equal scrollHeight)
            const lastCall = mockScrollTo.mock.calls[mockScrollTo.mock.calls.length - 1]
            expect(lastCall).toBeDefined()
            expect(lastCall[0]).toHaveProperty('top')
            expect(lastCall[0].top).toBe(mockScrollHeight)
          }
        ),
        { numRuns: 100 }
      )
    })

    it('should NOT auto-scroll when user has scrolled up away from bottom', async () => {
      await fc.assert(
        fc.asyncProperty(
          messageSequenceArb(3, 10),
          messageArb,
          async (initialMessages, newMessage) => {
            // Ensure new message has a later timestamp
            const lastTimestamp = initialMessages[initialMessages.length - 1]?.timestamp.getTime() || Date.now()
            const messageToAdd: Message = {
              ...newMessage,
              timestamp: new Date(lastTimestamp + 1000)
            }

            // Mount with initial messages
            const wrapper = mount(MessageList, {
              props: {
                messages: initialMessages
              }
            })

            await flushPromises()

            // Simulate user scrolling up (not near bottom)
            // User is more than 100px from bottom (threshold)
            mockScrollTop = 0 // At the top
            mockScrollHeight = 2000
            mockClientHeight = 500
            
            // Trigger scroll event to set isUserScrolling flag
            const container = wrapper.find('.message-list-container')
            await container.trigger('scroll')
            await flushPromises()

            // Reset mock to track new calls after scroll
            mockScrollTo.mockClear()

            // Add new message
            await wrapper.setProps({
              messages: [...initialMessages, messageToAdd]
            })

            await flushPromises()

            // Should NOT have called scrollTo because user is scrolled up
            expect(mockScrollTo).not.toHaveBeenCalled()
          }
        ),
        { numRuns: 100 }
      )
    })

    it('should resume auto-scroll when user scrolls back to bottom', async () => {
      await fc.assert(
        fc.asyncProperty(
          messageSequenceArb(2, 5),
          messageArb,
          async (initialMessages, newMessage) => {
            const lastTimestamp = initialMessages[initialMessages.length - 1]?.timestamp.getTime() || Date.now()
            const messageToAdd: Message = {
              ...newMessage,
              timestamp: new Date(lastTimestamp + 1000)
            }

            const wrapper = mount(MessageList, {
              props: {
                messages: initialMessages
              }
            })

            await flushPromises()

            // User scrolls up
            mockScrollTop = 0
            mockScrollHeight = 2000
            mockClientHeight = 500
            
            const container = wrapper.find('.message-list-container')
            await container.trigger('scroll')
            await flushPromises()

            mockScrollTo.mockClear()

            // User scrolls back to bottom
            mockScrollTop = mockScrollHeight - mockClientHeight - 50 // Within threshold (100px)
            await container.trigger('scroll')
            await flushPromises()

            // Add new message
            await wrapper.setProps({
              messages: [...initialMessages, messageToAdd]
            })

            await flushPromises()

            // Should auto-scroll again since user is back at bottom
            expect(mockScrollTo).toHaveBeenCalled()
          }
        ),
        { numRuns: 100 }
      )
    })
  })

  describe('Additional tests', () => {
    it('should render all messages in the provided array', async () => {
      await fc.assert(
        fc.asyncProperty(
          messageSequenceArb(1, 20),
          async (messages) => {
            const wrapper = mount(MessageList, {
              props: { messages }
            })

            await flushPromises()

            // Should render a MessageBubble for each message
            const bubbles = wrapper.findAllComponents({ name: 'MessageBubble' })
            expect(bubbles.length).toBe(messages.length)
          }
        ),
        { numRuns: 100 }
      )
    })

    it('should display streaming message when isStreaming is true', async () => {
      await fc.assert(
        fc.asyncProperty(
          messageSequenceArb(1, 5),
          fc.string({ minLength: 1, maxLength: 200 }),
          async (messages, streamingContent) => {
            const wrapper = mount(MessageList, {
              props: {
                messages,
                isStreaming: true,
                streamingContent
              }
            })

            await flushPromises()

            // Should render regular messages + streaming message
            const bubbles = wrapper.findAllComponents({ name: 'MessageBubble' })
            expect(bubbles.length).toBe(messages.length + 1)

            // Last bubble should be the streaming message
            const lastBubble = bubbles[bubbles.length - 1]
            expect(lastBubble.props('message').content).toBe(streamingContent)
            expect(lastBubble.props('message').role).toBe('agent')
            expect(lastBubble.props('message').status).toBe('sending')
          }
        ),
        { numRuns: 100 }
      )
    })

    it('should display typing indicator when streaming with no content', async () => {
      await fc.assert(
        fc.asyncProperty(
          messageSequenceArb(0, 5),
          async (messages) => {
            const wrapper = mount(MessageList, {
              props: {
                messages,
                isStreaming: true,
                streamingContent: ''
              }
            })

            await flushPromises()

            const typingIndicator = wrapper.find('.typing-indicator')
            expect(typingIndicator.exists()).toBe(true)
            expect(typingIndicator.attributes('role')).toBe('status')
            expect(typingIndicator.attributes('aria-label')).toContain('typing')
          }
        ),
        { numRuns: 100 }
      )
    })

    it('should display empty state when no messages and not streaming', async () => {
      const wrapper = mount(MessageList, {
        props: {
          messages: [],
          isStreaming: false
        }
      })

      await flushPromises()

      const emptyState = wrapper.find('.empty-state')
      expect(emptyState.exists()).toBe(true)
      expect(emptyState.text()).toContain('Start a conversation')
    })

    it('should have proper ARIA attributes for accessibility', async () => {
      await fc.assert(
        fc.asyncProperty(
          messageSequenceArb(0, 5),
          async (messages) => {
            const wrapper = mount(MessageList, {
              props: { messages }
            })

            await flushPromises()

            const container = wrapper.find('.message-list-container')
            expect(container.attributes('role')).toBe('log')
            expect(container.attributes('aria-live')).toBe('polite')
            expect(container.attributes('aria-label')).toBeDefined()
          }
        ),
        { numRuns: 100 }
      )
    })
  })
})
