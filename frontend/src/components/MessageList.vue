<script setup lang="ts">
import { ref, computed, watch, nextTick, onMounted, onUnmounted } from 'vue'
import MessageBubble from './MessageBubble.vue'
import type { Message } from '../types'

// Props
interface MessageListProps {
  messages: Message[]
  isStreaming?: boolean
  streamingContent?: string
}

const props = withDefaults(defineProps<MessageListProps>(), {
  isStreaming: false,
  streamingContent: ''
})

// Refs
const containerRef = ref<HTMLElement | null>(null)
const isUserScrolling = ref(false)
const scrollTimeout = ref<number | null>(null)

// Computed
const hasMessages = computed(() => props.messages.length > 0)
const shouldUseVirtualScrolling = computed(() => props.messages.length > 100)

// Streaming message for display
const streamingMessage = computed<Message | null>(() => {
  if (!props.isStreaming || !props.streamingContent) {
    return null
  }
  
  return {
    id: 'streaming-temp',
    role: 'agent',
    content: props.streamingContent,
    timestamp: new Date(),
    status: 'sending'
  }
})

/**
 * Check if user is at or near the bottom of the scroll container
 */
const isNearBottom = (): boolean => {
  if (!containerRef.value) return true
  
  const { scrollTop, scrollHeight, clientHeight } = containerRef.value
  const threshold = 100 // pixels from bottom
  
  return scrollHeight - scrollTop - clientHeight < threshold
}

/**
 * Scroll to the bottom of the message list
 */
const scrollToBottom = (smooth = true): void => {
  if (!containerRef.value) return
  
  const behavior = smooth ? 'smooth' : 'auto'
  containerRef.value.scrollTo({
    top: containerRef.value.scrollHeight,
    behavior
  })
}

/**
 * Handle scroll events to detect user scrolling
 */
const handleScroll = (): void => {
  if (!containerRef.value) return
  
  // Clear existing timeout
  if (scrollTimeout.value !== null) {
    window.clearTimeout(scrollTimeout.value)
  }
  
  // Check if user is near bottom
  const nearBottom = isNearBottom()
  
  // If user scrolls to bottom, resume auto-scroll
  if (nearBottom) {
    isUserScrolling.value = false
  } else {
    // User is scrolling away from bottom
    isUserScrolling.value = true
    
    // Reset user scrolling flag after 2 seconds of no scroll activity
    scrollTimeout.value = window.setTimeout(() => {
      if (isNearBottom()) {
        isUserScrolling.value = false
      }
    }, 2000)
  }
}

/**
 * Auto-scroll when new messages arrive
 */
watch(
  () => props.messages.length,
  async (newLength, oldLength) => {
    // Only auto-scroll if:
    // 1. A new message was added (not initial load or message removal)
    // 2. User is not actively scrolling up
    if (newLength > oldLength && !isUserScrolling.value) {
      await nextTick()
      scrollToBottom(true)
    }
  }
)

/**
 * Auto-scroll when streaming content updates
 */
watch(
  () => props.streamingContent,
  async () => {
    // Only auto-scroll during streaming if user is not scrolling up
    if (props.isStreaming && !isUserScrolling.value) {
      await nextTick()
      scrollToBottom(true)
    }
  }
)

/**
 * Scroll to bottom on mount
 */
onMounted(() => {
  nextTick(() => {
    scrollToBottom(false)
  })
})

/**
 * Cleanup on unmount
 */
onUnmounted(() => {
  if (scrollTimeout.value !== null) {
    window.clearTimeout(scrollTimeout.value)
  }
})

// Expose methods for parent components
defineExpose({
  scrollToBottom
})
</script>

<template>
  <div 
    ref="containerRef"
    class="message-list-container"
    :class="{ 'virtual-scroll': shouldUseVirtualScrolling }"
    role="log"
    aria-live="polite"
    aria-label="Conversation messages"
    @scroll="handleScroll"
  >
    <div class="message-list-content">
      <!-- Empty state -->
      <div v-if="!hasMessages && !isStreaming" class="empty-state">
        <div class="empty-state-icon">
          <svg 
            xmlns="http://www.w3.org/2000/svg" 
            fill="none" 
            viewBox="0 0 24 24" 
            stroke-width="1.5" 
            stroke="currentColor" 
            class="w-16 h-16"
            aria-hidden="true"
          >
            <path 
              stroke-linecap="round" 
              stroke-linejoin="round" 
              d="M8.625 12a.375.375 0 11-.75 0 .375.375 0 01.75 0zm0 0H8.25m4.125 0a.375.375 0 11-.75 0 .375.375 0 01.75 0zm0 0H12m4.125 0a.375.375 0 11-.75 0 .375.375 0 01.75 0zm0 0h-.375M21 12c0 4.556-4.03 8.25-9 8.25a9.764 9.764 0 01-2.555-.337A5.972 5.972 0 015.41 20.97a5.969 5.969 0 01-.474-.065 4.48 4.48 0 00.978-2.025c.09-.457-.133-.901-.467-1.226C3.93 16.178 3 14.189 3 12c0-4.556 4.03-8.25 9-8.25s9 3.694 9 8.25z" 
            />
          </svg>
        </div>
        <p class="empty-state-text">Start a conversation</p>
        <p class="empty-state-subtext">Send a message to begin chatting with the AI agent</p>
      </div>

      <!-- Message list -->
      <div v-else class="messages">
        <MessageBubble
          v-for="message in messages"
          :key="message.id"
          :message="message"
        />

        <!-- Streaming message -->
        <MessageBubble
          v-if="streamingMessage"
          :key="streamingMessage.id"
          :message="streamingMessage"
        />

        <!-- Typing indicator (shown during streaming with no content yet) -->
        <div 
          v-if="isStreaming && !streamingContent" 
          class="typing-indicator"
          role="status"
          aria-label="Agent is typing"
        >
          <div class="typing-dot"></div>
          <div class="typing-dot"></div>
          <div class="typing-dot"></div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.message-list-container {
  @apply flex-1 overflow-y-auto overflow-x-hidden bg-gray-50;
  scroll-behavior: smooth;
}

/* Disable smooth scrolling for reduced motion */
@media (prefers-reduced-motion: reduce) {
  .message-list-container {
    scroll-behavior: auto;
  }
}

.message-list-content {
  @apply min-h-full flex flex-col;
}

/* Empty state */
.empty-state {
  @apply flex-1 flex flex-col items-center justify-center p-8 text-center;
}

.empty-state-icon {
  @apply text-gray-300 mb-4;
}

.empty-state-text {
  @apply text-xl font-semibold text-gray-700 mb-2;
}

.empty-state-subtext {
  @apply text-sm text-gray-500 max-w-md;
}

/* Messages container */
.messages {
  @apply flex-1 p-4 space-y-0;
}

/* Typing indicator */
.typing-indicator {
  @apply flex items-center gap-1 p-4;
}

.typing-dot {
  @apply w-2 h-2 bg-gray-400 rounded-full;
  animation: typing 1.4s infinite;
}

.typing-dot:nth-child(2) {
  animation-delay: 0.2s;
}

.typing-dot:nth-child(3) {
  animation-delay: 0.4s;
}

@keyframes typing {
  0%, 60%, 100% {
    transform: translateY(0);
    opacity: 0.7;
  }
  30% {
    transform: translateY(-10px);
    opacity: 1;
  }
}

/* Disable animations for reduced motion */
@media (prefers-reduced-motion: reduce) {
  .typing-dot {
    animation: none;
  }
}

/* Virtual scrolling optimization (future enhancement) */
.virtual-scroll {
  /* Placeholder for virtual scrolling implementation */
  /* Would use a library like vue-virtual-scroller for production */
}
</style>
