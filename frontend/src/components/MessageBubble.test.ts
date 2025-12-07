import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import * as fc from 'fast-check'
import MessageBubble from './MessageBubble.vue'
import type { Message, MessageRole, MessageStatus, Citation } from '../types'

// Generators for property-based testing

const messageRoleArb = fc.constantFrom<MessageRole>('user', 'agent')
const messageStatusArb = fc.constantFrom<MessageStatus>('sending', 'sent', 'error')

const citationArb: fc.Arbitrary<Citation> = fc.record({
  sourceId: fc.uuid(),
  sourceName: fc.string({ minLength: 1, maxLength: 50 }),
  excerpt: fc.string({ minLength: 10, maxLength: 200 }),
  confidence: fc.option(fc.float({ min: 0, max: 1 })),
  url: fc.option(fc.webUrl()),
  metadata: fc.option(fc.dictionary(fc.string(), fc.anything()))
})

const messageArb: fc.Arbitrary<Message> = fc.record({
  id: fc.uuid(),
  role: messageRoleArb,
  content: fc.string({ minLength: 1, maxLength: 500 }),
  timestamp: fc.date(),
  citations: fc.option(fc.array(citationArb, { minLength: 1, maxLength: 5 })),
  status: messageStatusArb,
  errorMessage: fc.option(fc.string({ minLength: 1, maxLength: 100 }))
})

describe('MessageBubble', () => {
  describe('Property 16: Semantic HTML structure', () => {
    // Feature: chat-ui, Property 16: Semantic HTML structure
    // Validates: Requirements 5.3
    it('should use semantic HTML elements for all rendered messages', () => {
      fc.assert(
        fc.property(messageArb, (message) => {
          const wrapper = mount(MessageBubble, {
            props: { message }
          })

          // The root element should be an <article> tag (semantic element for message)
          expect(wrapper.element.tagName).toBe('ARTICLE')
          
          // Should have proper ARIA label
          expect(wrapper.attributes('aria-label')).toBeDefined()
          expect(wrapper.attributes('aria-label')).toContain(message.role)
          
          // Message text should be in a <p> tag (semantic element for paragraph)
          const messageText = wrapper.find('.message-text')
          expect(messageText.exists()).toBe(true)
          expect(messageText.element.tagName).toBe('P')
          
          // Timestamp should use <time> element (semantic element for time)
          const timestamp = wrapper.find('time')
          expect(timestamp.exists()).toBe(true)
          expect(timestamp.element.tagName).toBe('TIME')
          expect(timestamp.attributes('datetime')).toBeDefined()
          
          // If there are citations, they should have proper role
          if (message.citations && message.citations.length > 0) {
            const citationIndicator = wrapper.find('.citation-indicator')
            expect(citationIndicator.exists()).toBe(true)
            expect(citationIndicator.attributes('role')).toBe('status')
            expect(citationIndicator.attributes('aria-label')).toBeDefined()
          }
          
          // Status indicator should have role="status"
          if (message.role === 'user') {
            const statusIndicator = wrapper.find('.message-status')
            expect(statusIndicator.exists()).toBe(true)
            expect(statusIndicator.attributes('role')).toBe('status')
            expect(statusIndicator.attributes('aria-label')).toBeDefined()
          }
          
          // Error messages should have role="alert"
          if (message.status === 'error' && message.errorMessage) {
            const errorMsg = wrapper.find('.error-message')
            expect(errorMsg.exists()).toBe(true)
            expect(errorMsg.attributes('role')).toBe('alert')
          }
        }),
        { numRuns: 100 }
      )
    })
  })

  describe('Property 14: Citation visual distinction', () => {
    // Feature: chat-ui, Property 14: Citation visual distinction
    // Validates: Requirements 4.5, 9.1
    it('should visually distinguish messages with citations from those without', () => {
      fc.assert(
        fc.property(
          messageArb,
          fc.boolean(),
          (baseMessage, hasCitations) => {
            // Create message with or without citations based on the boolean
            const message: Message = {
              ...baseMessage,
              citations: hasCitations 
                ? [
                    {
                      sourceId: 'test-source-1',
                      sourceName: 'Test Source',
                      excerpt: 'Test excerpt from knowledge base',
                      confidence: 0.95
                    }
                  ]
                : undefined
            }

            const wrapper = mount(MessageBubble, {
              props: { message }
            })

            const citationIndicator = wrapper.find('.citation-indicator')

            if (hasCitations) {
              // Citation indicator should be present and visible
              expect(citationIndicator.exists()).toBe(true)
              
              // Should have citation icon
              const citationIcon = citationIndicator.find('.citation-icon')
              expect(citationIcon.exists()).toBe(true)
              
              // Should display citation count
              const citationCount = citationIndicator.find('.citation-count')
              expect(citationCount.exists()).toBe(true)
              expect(citationCount.text()).toBe('1')
              
              // Should have proper ARIA label
              expect(citationIndicator.attributes('aria-label')).toContain('citation')
            } else {
              // Citation indicator should not be present
              expect(citationIndicator.exists()).toBe(false)
            }
          }
        ),
        { numRuns: 100 }
      )
    })

    it('should display correct citation count for messages with multiple citations', () => {
      fc.assert(
        fc.property(
          messageArb,
          fc.array(citationArb, { minLength: 1, maxLength: 10 }),
          (baseMessage, citations) => {
            const message: Message = {
              ...baseMessage,
              citations
            }

            const wrapper = mount(MessageBubble, {
              props: { message }
            })

            const citationIndicator = wrapper.find('.citation-indicator')
            expect(citationIndicator.exists()).toBe(true)

            const citationCount = citationIndicator.find('.citation-count')
            expect(citationCount.exists()).toBe(true)
            expect(citationCount.text()).toBe(citations.length.toString())

            // ARIA label should reflect the count
            const ariaLabel = citationIndicator.attributes('aria-label')
            expect(ariaLabel).toContain(citations.length.toString())
            expect(ariaLabel).toContain('citation')
          }
        ),
        { numRuns: 100 }
      )
    })
  })

  describe('Additional property tests', () => {
    it('should display timestamp for all messages', () => {
      fc.assert(
        fc.property(messageArb, (message) => {
          const wrapper = mount(MessageBubble, {
            props: { message }
          })

          const timestamp = wrapper.find('.message-timestamp')
          expect(timestamp.exists()).toBe(true)
          expect(timestamp.text()).toBeTruthy()
          
          // Should have datetime attribute
          const timeElement = wrapper.find('time')
          expect(timeElement.attributes('datetime')).toBe(message.timestamp.toISOString())
        }),
        { numRuns: 100 }
      )
    })

    it('should apply correct role-based styling', () => {
      fc.assert(
        fc.property(messageArb, (message) => {
          const wrapper = mount(MessageBubble, {
            props: { message }
          })

          if (message.role === 'user') {
            expect(wrapper.classes()).toContain('message-user')
            const content = wrapper.find('.message-content')
            expect(content.classes()).toContain('user-content')
          } else {
            expect(wrapper.classes()).toContain('message-agent')
            const content = wrapper.find('.message-content')
            expect(content.classes()).toContain('agent-content')
          }
        }),
        { numRuns: 100 }
      )
    })

    it('should display status indicator for user messages', () => {
      fc.assert(
        fc.property(
          messageArb.filter(m => m.role === 'user'),
          (message) => {
            const wrapper = mount(MessageBubble, {
              props: { message }
            })

            const statusIndicator = wrapper.find('.message-status')
            expect(statusIndicator.exists()).toBe(true)
            expect(statusIndicator.classes()).toContain(`status-${message.status}`)
            
            // Should have status icon
            const statusIcon = statusIndicator.find('.status-icon')
            expect(statusIcon.exists()).toBe(true)
          }
        ),
        { numRuns: 100 }
      )
    })

    it('should display error message when status is error and errorMessage is provided', () => {
      fc.assert(
        fc.property(
          messageArb.filter(m => m.status === 'error' && m.errorMessage && m.errorMessage.trim().length > 0),
          (message) => {
            const wrapper = mount(MessageBubble, {
              props: { message }
            })

            const errorMsg = wrapper.find('.error-message')
            expect(errorMsg.exists()).toBe(true)
            // Vue trims whitespace in text content, so we compare trimmed values
            expect(errorMsg.text().trim()).toBe(message.errorMessage!.trim())
            expect(errorMsg.attributes('role')).toBe('alert')
          }
        ),
        { numRuns: 100 }
      )
    })
  })
})
