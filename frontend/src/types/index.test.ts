import { describe, it, expect } from 'vitest'
import {
  isMessage,
  isCitation,
  isSession,
  isChatError,
  isMessageRole,
  isMessageStatus,
  type Message,
  type Citation,
  type Session,
  type ChatError,
} from './index'

describe('Type Guards', () => {
  describe('isMessage', () => {
    it('should return true for valid message', () => {
      const validMessage: Message = {
        id: '123',
        role: 'user',
        content: 'Hello',
        timestamp: new Date(),
        status: 'sent',
      }
      expect(isMessage(validMessage)).toBe(true)
    })

    it('should return true for message with optional fields', () => {
      const messageWithOptionals: Message = {
        id: '123',
        role: 'agent',
        content: 'Response',
        timestamp: new Date(),
        status: 'sent',
        citations: [],
        errorMessage: 'Some error',
      }
      expect(isMessage(messageWithOptionals)).toBe(true)
    })

    it('should return false for invalid message', () => {
      expect(isMessage(null)).toBe(false)
      expect(isMessage(undefined)).toBe(false)
      expect(isMessage({})).toBe(false)
      expect(isMessage({ id: '123' })).toBe(false)
      expect(isMessage({ id: 123, role: 'user', content: 'test', timestamp: new Date(), status: 'sent' })).toBe(false)
    })

    it('should return false for message with invalid role', () => {
      const invalidRole = {
        id: '123',
        role: 'invalid',
        content: 'Hello',
        timestamp: new Date(),
        status: 'sent',
      }
      expect(isMessage(invalidRole)).toBe(false)
    })

    it('should return false for message with invalid status', () => {
      const invalidStatus = {
        id: '123',
        role: 'user',
        content: 'Hello',
        timestamp: new Date(),
        status: 'invalid',
      }
      expect(isMessage(invalidStatus)).toBe(false)
    })
  })

  describe('isCitation', () => {
    it('should return true for valid citation', () => {
      const validCitation: Citation = {
        sourceId: 'src-123',
        sourceName: 'Document',
        excerpt: 'Some text',
      }
      expect(isCitation(validCitation)).toBe(true)
    })

    it('should return true for citation with optional fields', () => {
      const citationWithOptionals: Citation = {
        sourceId: 'src-123',
        sourceName: 'Document',
        excerpt: 'Some text',
        confidence: 0.95,
        url: 'https://example.com',
        metadata: { key: 'value' },
      }
      expect(isCitation(citationWithOptionals)).toBe(true)
    })

    it('should return false for invalid citation', () => {
      expect(isCitation(null)).toBe(false)
      expect(isCitation({})).toBe(false)
      expect(isCitation({ sourceId: '123' })).toBe(false)
    })
  })

  describe('isSession', () => {
    it('should return true for valid session', () => {
      const validSession: Session = {
        id: 'session-123',
        createdAt: new Date(),
        messageCount: 0,
      }
      expect(isSession(validSession)).toBe(true)
    })

    it('should return true for session with optional fields', () => {
      const sessionWithOptionals: Session = {
        id: 'session-123',
        createdAt: new Date(),
        lastMessageAt: new Date(),
        messageCount: 5,
      }
      expect(isSession(sessionWithOptionals)).toBe(true)
    })

    it('should return false for invalid session', () => {
      expect(isSession(null)).toBe(false)
      expect(isSession({})).toBe(false)
      expect(isSession({ id: 'session-123' })).toBe(false)
    })
  })

  describe('isChatError', () => {
    it('should return true for valid chat error', () => {
      const validError: ChatError = {
        code: 'ERR_001',
        message: 'Something went wrong',
        retryable: true,
      }
      expect(isChatError(validError)).toBe(true)
    })

    it('should return true for error with optional fields', () => {
      const errorWithOptionals: ChatError = {
        code: 'ERR_001',
        message: 'Something went wrong',
        retryable: false,
        details: { reason: 'timeout' },
      }
      expect(isChatError(errorWithOptionals)).toBe(true)
    })

    it('should return false for invalid error', () => {
      expect(isChatError(null)).toBe(false)
      expect(isChatError({})).toBe(false)
      expect(isChatError({ code: 'ERR_001' })).toBe(false)
    })
  })

  describe('isMessageRole', () => {
    it('should return true for valid roles', () => {
      expect(isMessageRole('user')).toBe(true)
      expect(isMessageRole('agent')).toBe(true)
    })

    it('should return false for invalid roles', () => {
      expect(isMessageRole('invalid')).toBe(false)
      expect(isMessageRole(null)).toBe(false)
      expect(isMessageRole(undefined)).toBe(false)
      expect(isMessageRole(123)).toBe(false)
    })
  })

  describe('isMessageStatus', () => {
    it('should return true for valid statuses', () => {
      expect(isMessageStatus('sending')).toBe(true)
      expect(isMessageStatus('sent')).toBe(true)
      expect(isMessageStatus('error')).toBe(true)
    })

    it('should return false for invalid statuses', () => {
      expect(isMessageStatus('invalid')).toBe(false)
      expect(isMessageStatus(null)).toBe(false)
      expect(isMessageStatus(undefined)).toBe(false)
      expect(isMessageStatus(123)).toBe(false)
    })
  })
})
