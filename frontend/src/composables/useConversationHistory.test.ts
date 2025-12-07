import { describe, it, expect } from 'vitest'
import * as fc from 'fast-check'
import { useConversationHistory } from './useConversationHistory'
import type { Message, MessageRole, MessageStatus } from '@/types'

describe('useConversationHistory', () => {
  describe('Property 9: Chronological message ordering', () => {
    // Feature: chat-ui, Property 9: Chronological message ordering
    // Validates: Requirements 3.1
    it('should maintain chronological order for any set of messages', () => {
      fc.assert(
        fc.property(
          // Generate an array of messages with random timestamps
          fc.array(
            fc.record({
              id: fc.uuid(),
              role: fc.constantFrom<MessageRole>('user', 'agent'),
              content: fc.string({ minLength: 1, maxLength: 200 }),
              timestamp: fc.date({ min: new Date('2020-01-01'), max: new Date('2030-12-31') }),
              status: fc.constantFrom<MessageStatus>('sending', 'sent', 'error')
            }),
            { minLength: 0, maxLength: 50 }
          ),
          (messageData) => {
            // Create a new conversation history instance
            const history = useConversationHistory()

            // Add all messages
            messageData.forEach(data => {
              history.addMessage(data as Message)
            })

            // Verify messages are in chronological order
            const messages = history.messages.value
            for (let i = 1; i < messages.length; i++) {
              const prevTimestamp = messages[i - 1].timestamp.getTime()
              const currTimestamp = messages[i].timestamp.getTime()
              
              // Each message should have a timestamp >= the previous message
              expect(currTimestamp).toBeGreaterThanOrEqual(prevTimestamp)
            }
          }
        ),
        { numRuns: 100 }
      )
    })
  })

  describe('Property 11: Timestamp display', () => {
    // Feature: chat-ui, Property 11: Timestamp display
    // Validates: Requirements 3.4
    it('should include timestamp for any message', () => {
      fc.assert(
        fc.property(
          // Generate random messages
          fc.array(
            fc.record({
              id: fc.uuid(),
              role: fc.constantFrom<MessageRole>('user', 'agent'),
              content: fc.string({ minLength: 1, maxLength: 200 }),
              timestamp: fc.date({ min: new Date('2020-01-01'), max: new Date('2030-12-31') }),
              status: fc.constantFrom<MessageStatus>('sending', 'sent', 'error')
            }),
            { minLength: 1, maxLength: 30 }
          ),
          (messageData) => {
            // Create a new conversation history instance
            const history = useConversationHistory()

            // Add all messages
            messageData.forEach(data => {
              history.addMessage(data as Message)
            })

            // Verify every message has a timestamp
            const messages = history.messages.value
            messages.forEach(message => {
              expect(message.timestamp).toBeDefined()
              expect(message.timestamp).toBeInstanceOf(Date)
              expect(message.timestamp.getTime()).not.toBeNaN()
            })
          }
        ),
        { numRuns: 100 }
      )
    })
  })

  describe('Basic functionality tests', () => {
    it('should start with empty message array', () => {
      const history = useConversationHistory()
      expect(history.messages.value).toEqual([])
    })

    it('should add a message to the history', () => {
      const history = useConversationHistory()
      const message: Message = {
        id: '123',
        role: 'user',
        content: 'Hello',
        timestamp: new Date(),
        status: 'sent'
      }

      history.addMessage(message)
      expect(history.messages.value).toHaveLength(1)
      expect(history.messages.value[0]).toEqual(message)
    })

    it('should clear all messages', () => {
      const history = useConversationHistory()
      const message: Message = {
        id: '123',
        role: 'user',
        content: 'Hello',
        timestamp: new Date(),
        status: 'sent'
      }

      history.addMessage(message)
      expect(history.messages.value).toHaveLength(1)

      history.clearHistory()
      expect(history.messages.value).toEqual([])
    })

    it('should retrieve message by id', () => {
      const history = useConversationHistory()
      const message: Message = {
        id: '123',
        role: 'user',
        content: 'Hello',
        timestamp: new Date(),
        status: 'sent'
      }

      history.addMessage(message)
      const retrieved = history.getMessageById('123')
      expect(retrieved).toEqual(message)
    })

    it('should return undefined for non-existent message id', () => {
      const history = useConversationHistory()
      const retrieved = history.getMessageById('non-existent')
      expect(retrieved).toBeUndefined()
    })

    it('should maintain chronological order when adding messages out of order', () => {
      const history = useConversationHistory()
      
      const message1: Message = {
        id: '1',
        role: 'user',
        content: 'Third',
        timestamp: new Date('2024-01-03'),
        status: 'sent'
      }
      
      const message2: Message = {
        id: '2',
        role: 'agent',
        content: 'First',
        timestamp: new Date('2024-01-01'),
        status: 'sent'
      }
      
      const message3: Message = {
        id: '3',
        role: 'user',
        content: 'Second',
        timestamp: new Date('2024-01-02'),
        status: 'sent'
      }

      // Add messages out of chronological order
      history.addMessage(message1)
      history.addMessage(message2)
      history.addMessage(message3)

      // Verify they are stored in chronological order
      const messages = history.messages.value
      expect(messages[0].id).toBe('2') // First timestamp
      expect(messages[1].id).toBe('3') // Second timestamp
      expect(messages[2].id).toBe('1') // Third timestamp
    })
  })
})
