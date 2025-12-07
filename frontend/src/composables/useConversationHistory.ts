import { ref, type Ref } from 'vue'
import type { ConversationHistory, Message } from '@/types'

/**
 * Composable for managing conversation history
 * Handles message storage, retrieval, and ordering
 */
export function useConversationHistory(): ConversationHistory {
  // Reactive array of messages
  const messages: Ref<Message[]> = ref([])

  /**
   * Add a message to the conversation history
   * Messages are automatically ordered by timestamp
   */
  const addMessage = (message: Message): void => {
    // Ensure the message has a timestamp
    if (!message.timestamp) {
      message.timestamp = new Date()
    }

    // Add the message to the array
    messages.value.push(message)

    // Sort messages by timestamp to maintain chronological order
    messages.value.sort((a, b) => a.timestamp.getTime() - b.timestamp.getTime())
  }

  /**
   * Clear all messages from the conversation history
   * Used when resetting a session
   */
  const clearHistory = (): void => {
    messages.value = []
  }

  /**
   * Retrieve a message by its unique identifier
   * Returns undefined if the message is not found
   */
  const getMessageById = (id: string): Message | undefined => {
    return messages.value.find(message => message.id === id)
  }

  return {
    messages,
    addMessage,
    clearHistory,
    getMessageById
  }
}
