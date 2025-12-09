import { ref, type Ref } from 'vue'
import type { SessionManager, SessionMetadata } from '@/types'

/**
 * Composable for managing chat sessions
 * Handles session creation, switching, and metadata tracking
 */
export function useSessionManager(): SessionManager {
  // API configuration
  const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080'

  // Initialize with empty session (will be created on mount)
  const currentSessionId: Ref<string> = ref('')
  
  const sessionMetadata: Ref<SessionMetadata> = ref({
    id: '',
    createdAt: new Date(),
    messageCount: 0
  })

  // Store sessions in memory (maps session ID to metadata)
  const sessions = new Map<string, SessionMetadata>()

  /**
   * Create a new session with a unique identifier
   * Calls backend API to create session in database
   */
  const createNewSession = async (): Promise<string> => {
    try {
      // Call backend API to create session
      const response = await fetch(`${API_URL}/api/sessions`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        }
      })

      if (!response.ok) {
        throw new Error(`Failed to create session: ${response.statusText}`)
      }

      const data = await response.json()
      const newSessionId = data.id
      const newMetadata: SessionMetadata = {
        id: newSessionId,
        createdAt: new Date(data.createdAt),
        messageCount: data.messageCount || 0
      }

      // Store the new session
      sessions.set(newSessionId, newMetadata)

      // Update current session
      currentSessionId.value = newSessionId
      sessionMetadata.value = newMetadata

      return newSessionId
    } catch (error) {
      console.error('Failed to create session:', error)
      throw error
    }
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
