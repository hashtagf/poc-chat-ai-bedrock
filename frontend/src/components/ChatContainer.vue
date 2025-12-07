<script setup lang="ts">
import { ref, computed, watch, onMounted, nextTick } from 'vue'
import MessageList from './MessageList.vue'
import MessageInput from './MessageInput.vue'
import ErrorDisplay from './ErrorDisplay.vue'
import { useChatService } from '@/composables/useChatService'
import { useConversationHistory } from '@/composables/useConversationHistory'
import { useSessionManager } from '@/composables/useSessionManager'
import { useErrorHandler } from '@/composables/useErrorHandler'
import type { Message, ChatError } from '@/types'

// Refs for child components
const messageInputRef = ref<InstanceType<typeof MessageInput> | null>(null)
const messageListRef = ref<InstanceType<typeof MessageList> | null>(null)

// Initialize composables
const sessionManager = useSessionManager()
const conversationHistory = useConversationHistory()
const errorHandler = useErrorHandler()

// Initialize chat service with session ID and callbacks
const chatService = useChatService({
  sessionId: sessionManager.currentSessionId.value,
  onMessageComplete: handleMessageComplete,
  onError: handleChatError
})

// Connection status tracking
const connectionStatus = ref<'connected' | 'disconnected' | 'connecting'>('connecting')

// Loading states
const isInitializing = ref(true)
const isSendingMessage = ref(false)

// Computed properties
const isInputDisabled = computed(() => {
  return isSendingMessage.value || chatService.isStreaming.value || connectionStatus.value !== 'connected'
})

const currentError = computed(() => {
  return chatService.error.value || errorHandler.currentError.value
})

/**
 * Generate a UUID v4 for message IDs
 */
const generateMessageId = (): string => {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, (c) => {
    const r = (Math.random() * 16) | 0
    const v = c === 'x' ? r : (r & 0x3) | 0x8
    return v.toString(16)
  })
}

/**
 * Handle message submission from MessageInput
 */
const handleMessageSubmit = async (content: string): Promise<void> => {
  try {
    isSendingMessage.value = true
    
    // Create user message
    const userMessage: Message = {
      id: generateMessageId(),
      role: 'user',
      content,
      timestamp: new Date(),
      status: 'sending'
    }
    
    // Add user message to history
    conversationHistory.addMessage(userMessage)
    
    // Update session metadata
    sessionManager.sessionMetadata.value.messageCount++
    
    // Send message through chat service
    await chatService.sendMessage(content)
    
    // Update user message status to sent
    userMessage.status = 'sent'
    
  } catch (error) {
    // Handle error
    errorHandler.handleError(error)
    
    // Update the last message status to error if it exists
    const messages = conversationHistory.messages.value
    if (messages.length > 0) {
      const lastMessage = messages[messages.length - 1]
      if (lastMessage.role === 'user') {
        lastMessage.status = 'error'
        lastMessage.errorMessage = 'Failed to send message'
      }
    }
  } finally {
    isSendingMessage.value = false
  }
}

/**
 * Handle completion of streaming message from chat service
 */
function handleMessageComplete(content: string): void {
  // Create agent message
  const agentMessage: Message = {
    id: generateMessageId(),
    role: 'agent',
    content,
    timestamp: new Date(),
    status: 'sent'
    // TODO: Add citations when backend provides them
  }
  
  // Add agent message to history
  conversationHistory.addMessage(agentMessage)
  
  // Update session metadata
  sessionManager.sessionMetadata.value.messageCount++
}

/**
 * Handle errors from chat service
 */
function handleChatError(error: ChatError): void {
  errorHandler.handleError(error)
}

/**
 * Handle retry action from ErrorDisplay
 */
const handleRetry = (): void => {
  // Clear the error
  chatService.clearError()
  errorHandler.clearError()
  
  // Retry the last failed message if it exists
  const messages = conversationHistory.messages.value
  const lastMessage = messages[messages.length - 1]
  
  if (lastMessage && lastMessage.role === 'user' && lastMessage.status === 'error') {
    handleMessageSubmit(lastMessage.content)
  }
}

/**
 * Handle error dismissal from ErrorDisplay
 */
const handleErrorDismiss = (): void => {
  chatService.clearError()
  errorHandler.clearError()
}

/**
 * Create a new session
 */
const handleNewSession = async (): Promise<void> => {
  try {
    // Clear current conversation history
    conversationHistory.clearHistory()
    
    // Create new session
    await sessionManager.createNewSession()
    
    // Clear any errors
    chatService.clearError()
    errorHandler.clearError()
    
    // Focus input field
    await nextTick()
    if (messageInputRef.value) {
      messageInputRef.value.focus()
    }
    
    // Scroll to bottom
    if (messageListRef.value) {
      messageListRef.value.scrollToBottom(false)
    }
  } catch (error) {
    errorHandler.handleError(error)
  }
}

/**
 * Format date for display
 */
const formatSessionDate = (date: Date): string => {
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  const minutes = Math.floor(diff / 60000)
  const hours = Math.floor(diff / 3600000)
  const days = Math.floor(diff / 86400000)
  
  if (minutes < 1) return 'Just now'
  if (minutes < 60) return `${minutes}m ago`
  if (hours < 24) return `${hours}h ago`
  if (days < 7) return `${days}d ago`
  
  return date.toLocaleDateString()
}

// Watch for connection status changes
watch(() => chatService.error.value, (error) => {
  if (error?.code === 'NETWORK_ERROR' || error?.code === 'CONNECTION_LOST') {
    connectionStatus.value = 'disconnected'
  } else if (error?.code === 'STREAMING_IN_PROGRESS') {
    connectionStatus.value = 'connected'
  }
})

// Initialize component
onMounted(async () => {
  // Simulate initialization delay
  await new Promise(resolve => setTimeout(resolve, 500))
  
  isInitializing.value = false
  connectionStatus.value = 'connected'
  
  // Focus input field
  await nextTick()
  if (messageInputRef.value) {
    messageInputRef.value.focus()
  }
})
</script>

<template>
  <div class="chat-container" role="main" aria-label="Chat interface">
    <!-- Loading state -->
    <div v-if="isInitializing" class="loading-container">
      <div class="loading-spinner">
        <svg
          class="animate-spin h-12 w-12 text-blue-600"
          xmlns="http://www.w3.org/2000/svg"
          fill="none"
          viewBox="0 0 24 24"
          aria-hidden="true"
        >
          <circle
            class="opacity-25"
            cx="12"
            cy="12"
            r="10"
            stroke="currentColor"
            stroke-width="4"
          ></circle>
          <path
            class="opacity-75"
            fill="currentColor"
            d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
          ></path>
        </svg>
      </div>
      <p class="loading-text">Initializing chat...</p>
    </div>

    <!-- Main chat interface -->
    <div v-else class="chat-layout">
      <!-- Header with session controls -->
      <header class="chat-header" role="banner">
        <div class="header-content">
          <div class="header-title">
            <h1 class="title-text">Bedrock Agent Chat</h1>
            <div class="session-info" aria-live="polite">
              <span class="session-label">Session:</span>
              <span class="session-time">{{ formatSessionDate(sessionManager.sessionMetadata.value.createdAt) }}</span>
              <span class="session-messages" aria-label="Message count">
                {{ sessionManager.sessionMetadata.value.messageCount }} messages
              </span>
            </div>
          </div>
          
          <button
            @click="handleNewSession"
            class="new-session-button"
            aria-label="Start new session"
          >
            <svg
              xmlns="http://www.w3.org/2000/svg"
              viewBox="0 0 24 24"
              fill="currentColor"
              class="w-5 h-5"
              aria-hidden="true"
            >
              <path
                fill-rule="evenodd"
                d="M12 3.75a.75.75 0 01.75.75v6.75h6.75a.75.75 0 010 1.5h-6.75v6.75a.75.75 0 01-1.5 0v-6.75H4.5a.75.75 0 010-1.5h6.75V4.5a.75.75 0 01.75-.75z"
                clip-rule="evenodd"
              />
            </svg>
            <span>New Session</span>
          </button>
        </div>
      </header>

      <!-- Error display (fixed position) -->
      <ErrorDisplay
        :error="currentError"
        :connection-status="connectionStatus"
        @retry="handleRetry"
        @dismiss="handleErrorDismiss"
      />

      <!-- Message list -->
      <MessageList
        ref="messageListRef"
        :messages="conversationHistory.messages.value"
        :is-streaming="chatService.isStreaming.value"
        :streaming-content="chatService.streamingMessage.value"
      />

      <!-- ARIA live region for screen readers -->
      <div
        class="sr-only"
        role="status"
        aria-live="polite"
        aria-atomic="true"
      >
        <span v-if="chatService.isStreaming.value">
          Agent is responding
        </span>
        <span v-else-if="isSendingMessage">
          Sending message
        </span>
      </div>

      <!-- Message input -->
      <MessageInput
        ref="messageInputRef"
        :disabled="isInputDisabled"
        placeholder="Type your message..."
        @submit="handleMessageSubmit"
      />
    </div>
  </div>
</template>

<style scoped>
.chat-container {
  @apply h-screen w-full bg-gray-50 flex flex-col;
}

.loading-container {
  @apply flex-1 flex flex-col items-center justify-center;
}

.loading-spinner {
  @apply mb-4;
}

.loading-text {
  @apply text-lg text-gray-600;
}

.chat-layout {
  @apply h-full flex flex-col;
}

/* Header */
.chat-header {
  @apply bg-white border-b border-gray-200 shadow-sm;
}

.header-content {
  @apply flex items-center justify-between px-6 py-4;
}

.header-title {
  @apply flex-1;
}

.title-text {
  @apply text-2xl font-bold text-gray-900 mb-1;
}

.session-info {
  @apply flex items-center gap-3 text-sm text-gray-600;
}

.session-label {
  @apply font-medium;
}

.session-time {
  @apply text-gray-500;
}

.session-messages {
  @apply px-2 py-1 bg-blue-100 text-blue-700 rounded-full text-xs font-medium;
}

.new-session-button {
  @apply flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 transition-colors;
}

/* Screen reader only content */
.sr-only {
  @apply absolute w-px h-px p-0 -m-px overflow-hidden whitespace-nowrap border-0;
  clip: rect(0, 0, 0, 0);
}

/* Reduced motion support */
@media (prefers-reduced-motion: reduce) {
  .animate-spin {
    animation: none;
  }
  
  .new-session-button {
    @apply transition-none;
  }
}

/* Responsive design */
@media (max-width: 640px) {
  .header-content {
    @apply flex-col items-start gap-3;
  }
  
  .new-session-button {
    @apply w-full justify-center;
  }
  
  .session-info {
    @apply flex-wrap;
  }
}
</style>
