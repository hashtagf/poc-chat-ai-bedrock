import { describe, it, expect } from 'vitest'
import * as fc from 'fast-check'
import { useSessionManager } from './useSessionManager'

describe('useSessionManager', () => {
  describe('Property 20: Session reset clears history', () => {
    // Feature: chat-ui, Property 20: Session reset clears history
    // Validates: Requirements 7.1, 7.2
    it('should generate a unique session ID and reset metadata when creating a new session', async () => {
      await fc.assert(
        fc.asyncProperty(
          fc.integer({ min: 0, max: 100 }), // Simulate message count
          async (messageCount) => {
            const manager = useSessionManager()
            
            // Capture initial session
            const initialSessionId = manager.currentSessionId.value
            const initialCreatedAt = manager.sessionMetadata.value.createdAt
            
            // Simulate some activity by updating message count
            manager.sessionMetadata.value.messageCount = messageCount
            
            // Create new session
            const newSessionId = await manager.createNewSession()
            
            // Verify new session has different ID
            expect(newSessionId).not.toBe(initialSessionId)
            expect(manager.currentSessionId.value).toBe(newSessionId)
            
            // Verify session metadata is reset
            expect(manager.sessionMetadata.value.id).toBe(newSessionId)
            expect(manager.sessionMetadata.value.messageCount).toBe(0)
            expect(manager.sessionMetadata.value.createdAt).not.toBe(initialCreatedAt)
            
            // Verify new session ID is a valid UUID v4 format
            const uuidRegex = /^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i
            expect(newSessionId).toMatch(uuidRegex)
          }
        ),
        { numRuns: 100 }
      )
    })
  })

  describe('Property 21: Session isolation', () => {
    // Feature: chat-ui, Property 21: Session isolation
    // Validates: Requirements 7.3
    it('should maintain separate metadata for different sessions', async () => {
      await fc.assert(
        fc.asyncProperty(
          fc.array(fc.integer({ min: 0, max: 50 }), { minLength: 2, maxLength: 5 }),
          async (messageCounts) => {
            const manager = useSessionManager()
            const sessionIds: string[] = []
            
            // Create multiple sessions with different message counts
            for (let i = 0; i < messageCounts.length; i++) {
              const sessionId = await manager.createNewSession()
              sessionIds.push(sessionId)
              manager.sessionMetadata.value.messageCount = messageCounts[i]
            }
            
            // Verify each session maintains its own metadata
            for (let i = 0; i < sessionIds.length; i++) {
              await manager.loadSession(sessionIds[i])
              
              // Each session should have its own unique ID
              expect(manager.currentSessionId.value).toBe(sessionIds[i])
              expect(manager.sessionMetadata.value.id).toBe(sessionIds[i])
              
              // Message count should match what was set for this session
              expect(manager.sessionMetadata.value.messageCount).toBe(messageCounts[i])
            }
            
            // Verify all session IDs are unique
            const uniqueIds = new Set(sessionIds)
            expect(uniqueIds.size).toBe(sessionIds.length)
          }
        ),
        { numRuns: 100 }
      )
    })
  })

  describe('Property 22: Session metadata display', () => {
    // Feature: chat-ui, Property 22: Session metadata display
    // Validates: Requirements 7.4
    it('should always expose session ID and creation timestamp in metadata', async () => {
      await fc.assert(
        fc.asyncProperty(
          fc.integer({ min: 1, max: 10 }),
          async (numSessions) => {
            const manager = useSessionManager()
            
            for (let i = 0; i < numSessions; i++) {
              const sessionId = await manager.createNewSession()
              
              // Verify metadata contains required fields
              expect(manager.sessionMetadata.value.id).toBe(sessionId)
              expect(manager.sessionMetadata.value.id).toBe(manager.currentSessionId.value)
              expect(manager.sessionMetadata.value.createdAt).toBeInstanceOf(Date)
              expect(manager.sessionMetadata.value.createdAt.getTime()).toBeLessThanOrEqual(Date.now())
              
              // Verify messageCount is initialized
              expect(typeof manager.sessionMetadata.value.messageCount).toBe('number')
              expect(manager.sessionMetadata.value.messageCount).toBeGreaterThanOrEqual(0)
            }
          }
        ),
        { numRuns: 100 }
      )
    })
  })

  describe('Property 23: New session input focus', () => {
    // Feature: chat-ui, Property 23: New session input focus
    // Validates: Requirements 7.5
    it('should return session ID immediately for focus management', async () => {
      await fc.assert(
        fc.asyncProperty(
          fc.integer({ min: 1, max: 20 }),
          async (numCreations) => {
            const manager = useSessionManager()
            
            for (let i = 0; i < numCreations; i++) {
              const beforeCreation = Date.now()
              const sessionId = await manager.createNewSession()
              const afterCreation = Date.now()
              
              // Verify session ID is returned immediately (within reasonable time)
              const creationTime = afterCreation - beforeCreation
              expect(creationTime).toBeLessThan(100) // Should be nearly instant
              
              // Verify the returned ID matches current session
              expect(sessionId).toBe(manager.currentSessionId.value)
              
              // Verify session is immediately usable
              expect(manager.sessionMetadata.value.id).toBe(sessionId)
            }
          }
        ),
        { numRuns: 100 }
      )
    })
  })
})
