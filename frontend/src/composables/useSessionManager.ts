import { ref, type Ref } from 'vue'
import type { SessionManager, SessionMetadata } from '@/types'

/**
 * Composable for managing chat sessions
 * Handles session creation, switching, and metadata tracking
 */
export function useSessionManager(): SessionManager {
  // Generate a UUID v4
  const generateUUID = (): string => {
    return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, (c) => {
      const r = (Math.random() * 16) | 0
      const v = c === 'x' ? r : (r & 0x3) | 0x8
      return v.toString(16)
    })
  }

  // Initialize with a new session
  const initialSessionId = generateUUID()
  const currentSessionId: Ref<string> = ref(initialSessionId)
  
  const sessionMetadata: Ref<SessionMetadata> = ref({
    id: initialSessionId,
    createdAt: new Date(),
    messageCount: 0
  })

  // Store sessions in memory (maps session ID to metadata)
  const sessions = new Map<string, SessionMetadata>()
  sessions.set(initialSessionId, sessionMetadata.value)

  /**
   * Create a new session with a unique identifier
   * Clears current session state and generates new session metadata
   */
  const createNewSession = async (): Promise<string> => {
    const newSessionId = generateUUID()
    const newMetadata: SessionMetadata = {
      id: newSessionId,
      createdAt: new Date(),
      messageCount: 0
    }

    // Store the new session
    sessions.set(newSessionId, newMetadata)

    // Update current session
    currentSessionId.value = newSessionId
    sessionMetadata.value = newMetadata

    return newSessionId
  }

  /**
   * Load an existing session by ID
   * Switches to the specified session and restores its metadata
   */
  const loadSession = async (sessionId: string): Promise<void> => {
    const existingSession = sessions.get(sessionId)
    
    if (!existingSession) {
      throw new Error(`Session ${sessionId} not found`)
    }

    currentSessionId.value = sessionId
    sessionMetadata.value = existingSession
  }

  return {
    currentSessionId,
    createNewSession,
    loadSession,
    sessionMetadata
  }
}
