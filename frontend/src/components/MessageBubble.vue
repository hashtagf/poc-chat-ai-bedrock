<script setup lang="ts">
import { computed } from 'vue'
import type { Message } from '../types'

// Props
interface MessageBubbleProps {
  message: Message
}

const props = defineProps<MessageBubbleProps>()

// Computed
const isUserMessage = computed(() => props.message.role === 'user')
const isAgentMessage = computed(() => props.message.role === 'agent')

const formattedTimestamp = computed(() => {
  const date = props.message.timestamp
  const now = new Date()
  const diffMs = now.getTime() - date.getTime()
  const diffMins = Math.floor(diffMs / 60000)
  
  // Less than 1 minute ago
  if (diffMins < 1) {
    return 'Just now'
  }
  
  // Less than 60 minutes ago
  if (diffMins < 60) {
    return `${diffMins}m ago`
  }
  
  // Less than 24 hours ago
  const diffHours = Math.floor(diffMins / 60)
  if (diffHours < 24) {
    return `${diffHours}h ago`
  }
  
  // Format as time if today, otherwise include date
  const isToday = date.toDateString() === now.toDateString()
  if (isToday) {
    return date.toLocaleTimeString('en-US', { 
      hour: 'numeric', 
      minute: '2-digit',
      hour12: true 
    })
  }
  
  // Format with date
  return date.toLocaleString('en-US', {
    month: 'short',
    day: 'numeric',
    hour: 'numeric',
    minute: '2-digit',
    hour12: true
  })
})

const statusIcon = computed(() => {
  switch (props.message.status) {
    case 'sending':
      return 'sending'
    case 'sent':
      return 'sent'
    case 'error':
      return 'error'
    default:
      return 'sent'
  }
})

const statusLabel = computed(() => {
  switch (props.message.status) {
    case 'sending':
      return 'Sending...'
    case 'sent':
      return 'Sent'
    case 'error':
      return 'Failed to send'
    default:
      return 'Sent'
  }
})

const hasCitations = computed(() => {
  return props.message.citations && props.message.citations.length > 0
})

const citationCount = computed(() => {
  return props.message.citations?.length || 0
})
</script>

<template>
  <article
    class="message-bubble"
    :class="{
      'message-user': isUserMessage,
      'message-agent': isAgentMessage,
      'message-error': message.status === 'error'
    }"
    :aria-label="`Message from ${message.role} at ${formattedTimestamp}`"
  >
    <div class="message-container">
      <!-- Message content -->
      <div class="message-content-wrapper">
        <div 
          class="message-content"
          :class="{
            'user-content': isUserMessage,
            'agent-content': isAgentMessage
          }"
        >
          <p class="message-text">{{ message.content }}</p>
          
          <!-- Citation indicator -->
          <div 
            v-if="hasCitations" 
            class="citation-indicator"
            :aria-label="`${citationCount} citation${citationCount > 1 ? 's' : ''} available`"
            role="status"
          >
            <svg 
              xmlns="http://www.w3.org/2000/svg" 
              viewBox="0 0 20 20" 
              fill="currentColor" 
              class="citation-icon"
              aria-hidden="true"
            >
              <path d="M7 3.5A1.5 1.5 0 018.5 2h3.879a1.5 1.5 0 011.06.44l3.122 3.12A1.5 1.5 0 0117 6.622V12.5a1.5 1.5 0 01-1.5 1.5h-1v-3.379a3 3 0 00-.879-2.121L10.5 5.379A3 3 0 008.379 4.5H7v-1z" />
              <path d="M4.5 6A1.5 1.5 0 003 7.5v9A1.5 1.5 0 004.5 18h7a1.5 1.5 0 001.5-1.5v-5.879a1.5 1.5 0 00-.44-1.06L9.44 6.439A1.5 1.5 0 008.378 6H4.5z" />
            </svg>
            <span class="citation-count">{{ citationCount }}</span>
          </div>
        </div>
        
        <!-- Metadata row -->
        <div class="message-metadata">
          <!-- Timestamp -->
          <time 
            :datetime="message.timestamp.toISOString()"
            class="message-timestamp"
          >
            {{ formattedTimestamp }}
          </time>
          
          <!-- Status indicator (for user messages) -->
          <div 
            v-if="isUserMessage"
            class="message-status"
            :class="`status-${message.status}`"
            :aria-label="statusLabel"
            role="status"
          >
            <!-- Sending icon -->
            <svg 
              v-if="statusIcon === 'sending'"
              class="status-icon animate-spin"
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
              />
              <path 
                class="opacity-75" 
                fill="currentColor" 
                d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
              />
            </svg>
            
            <!-- Sent icon (checkmark) -->
            <svg 
              v-else-if="statusIcon === 'sent'"
              class="status-icon"
              xmlns="http://www.w3.org/2000/svg"
              viewBox="0 0 20 20"
              fill="currentColor"
              aria-hidden="true"
            >
              <path 
                fill-rule="evenodd" 
                d="M16.704 4.153a.75.75 0 01.143 1.052l-8 10.5a.75.75 0 01-1.127.075l-4.5-4.5a.75.75 0 011.06-1.06l3.894 3.893 7.48-9.817a.75.75 0 011.05-.143z" 
                clip-rule="evenodd" 
              />
            </svg>
            
            <!-- Error icon -->
            <svg 
              v-else-if="statusIcon === 'error'"
              class="status-icon"
              xmlns="http://www.w3.org/2000/svg"
              viewBox="0 0 20 20"
              fill="currentColor"
              aria-hidden="true"
            >
              <path 
                fill-rule="evenodd" 
                d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-8-5a.75.75 0 01.75.75v4.5a.75.75 0 01-1.5 0v-4.5A.75.75 0 0110 5zm0 10a1 1 0 100-2 1 1 0 000 2z" 
                clip-rule="evenodd" 
              />
            </svg>
          </div>
        </div>
        
        <!-- Error message -->
        <div 
          v-if="message.status === 'error' && message.errorMessage"
          class="error-message"
          role="alert"
        >
          {{ message.errorMessage }}
        </div>
      </div>
    </div>
  </article>
</template>

<style scoped>
.message-bubble {
  @apply w-full mb-4;
}

.message-container {
  @apply flex;
}

/* User message alignment */
.message-user .message-container {
  @apply justify-end;
}

/* Agent message alignment */
.message-agent .message-container {
  @apply justify-start;
}

.message-content-wrapper {
  @apply max-w-[75%] flex flex-col;
}

.message-content {
  @apply rounded-lg px-4 py-3 shadow-sm;
}

/* User message styling */
.user-content {
  @apply bg-blue-600 text-white;
}

/* Agent message styling */
.agent-content {
  @apply bg-white text-gray-900 border border-gray-200;
}

/* Error state */
.message-error .user-content {
  @apply bg-red-600;
}

.message-error .agent-content {
  @apply border-red-300 bg-red-50;
}

.message-text {
  @apply m-0 whitespace-pre-wrap break-words;
}

/* Citation indicator */
.citation-indicator {
  @apply flex items-center gap-1 mt-2 pt-2 border-t border-current opacity-75;
}

.user-content .citation-indicator {
  @apply border-white/30;
}

.agent-content .citation-indicator {
  @apply border-gray-300;
}

.citation-icon {
  @apply w-4 h-4;
}

.citation-count {
  @apply text-sm font-medium;
}

/* Metadata */
.message-metadata {
  @apply flex items-center gap-2 mt-1 px-1;
}

.message-user .message-metadata {
  @apply justify-end;
}

.message-agent .message-metadata {
  @apply justify-start;
}

.message-timestamp {
  @apply text-xs text-gray-500;
}

/* Status indicator */
.message-status {
  @apply flex items-center;
}

.status-icon {
  @apply w-4 h-4;
}

.status-sending .status-icon {
  @apply text-gray-400;
}

.status-sent .status-icon {
  @apply text-green-600;
}

.status-error .status-icon {
  @apply text-red-600;
}

/* Error message */
.error-message {
  @apply mt-2 px-3 py-2 text-sm text-red-700 bg-red-100 border border-red-300 rounded;
}

/* Reduced motion support */
@media (prefers-reduced-motion: reduce) {
  .animate-spin {
    animation: none;
  }
}
</style>
