<script setup lang="ts">
import { ref, watch, onUnmounted } from 'vue'
import type { ChatError } from '@/types'

// Props
interface ErrorDisplayProps {
  error: ChatError | null
  connectionStatus?: 'connected' | 'disconnected' | 'connecting'
}

const props = withDefaults(defineProps<ErrorDisplayProps>(), {
  connectionStatus: 'connected'
})

// Emits
interface ErrorDisplayEmits {
  (e: 'retry'): void
  (e: 'dismiss'): void
}

const emit = defineEmits<ErrorDisplayEmits>()

// State
const isVisible = ref(false)
const autoDismissTimer = ref<number | null>(null)

// Auto-dismiss timeout for non-critical errors (5 seconds)
const AUTO_DISMISS_TIMEOUT_MS = 5000

// Critical error codes that should not auto-dismiss
const CRITICAL_ERROR_CODES = [
  'INVALID_SESSION',
  'AGENT_UNAVAILABLE',
  'NETWORK_ERROR'
]

/**
 * Check if an error is critical (should not auto-dismiss)
 */
const isCriticalError = (error: ChatError | null): boolean => {
  if (!error) return false
  return CRITICAL_ERROR_CODES.includes(error.code)
}

/**
 * Handle retry button click
 */
const handleRetry = (): void => {
  clearAutoDismissTimer()
  emit('retry')
}

/**
 * Handle dismiss button click
 */
const handleDismiss = (): void => {
  clearAutoDismissTimer()
  isVisible.value = false
  emit('dismiss')
}

/**
 * Handle keyboard events (Escape to dismiss)
 */
const handleKeydown = (event: KeyboardEvent): void => {
  if (event.key === 'Escape') {
    handleDismiss()
  }
}

/**
 * Clear the auto-dismiss timer
 */
const clearAutoDismissTimer = (): void => {
  if (autoDismissTimer.value !== null) {
    window.clearTimeout(autoDismissTimer.value)
    autoDismissTimer.value = null
  }
}

/**
 * Set up auto-dismiss for non-critical errors
 */
const setupAutoDismiss = (error: ChatError | null): void => {
  clearAutoDismissTimer()
  
  if (!error || isCriticalError(error)) {
    return
  }
  
  autoDismissTimer.value = window.setTimeout(() => {
    handleDismiss()
  }, AUTO_DISMISS_TIMEOUT_MS)
}

// Watch for error changes
watch(() => props.error, (newError) => {
  if (newError) {
    isVisible.value = true
    setupAutoDismiss(newError)
    // Add keyboard listener
    window.addEventListener('keydown', handleKeydown)
  } else {
    isVisible.value = false
    clearAutoDismissTimer()
    // Remove keyboard listener
    window.removeEventListener('keydown', handleKeydown)
  }
}, { immediate: true })

// Cleanup on unmount
onUnmounted(() => {
  clearAutoDismissTimer()
  window.removeEventListener('keydown', handleKeydown)
})

/**
 * Get connection status display text
 */
const getConnectionStatusText = (): string => {
  switch (props.connectionStatus) {
    case 'connected':
      return 'Connected'
    case 'disconnected':
      return 'Disconnected'
    case 'connecting':
      return 'Connecting...'
    default:
      return 'Unknown'
  }
}

/**
 * Get connection status icon color
 */
const getConnectionStatusColor = (): string => {
  switch (props.connectionStatus) {
    case 'connected':
      return 'text-green-500'
    case 'disconnected':
      return 'text-red-500'
    case 'connecting':
      return 'text-yellow-500'
    default:
      return 'text-gray-500'
  }
}
</script>

<template>
  <div class="error-display-container">
    <!-- Connection Status Indicator -->
    <div 
      v-if="connectionStatus !== 'connected'"
      class="connection-status"
      role="status"
      aria-live="polite"
    >
      <div class="connection-status-content">
        <span 
          class="connection-indicator"
          :class="getConnectionStatusColor()"
          aria-hidden="true"
        >
          ●
        </span>
        <span class="connection-text">
          {{ getConnectionStatusText() }}
        </span>
      </div>
    </div>

    <!-- Error Message Display -->
    <transition
      name="error-fade"
      @before-enter="() => {}"
      @enter="() => {}"
      @leave="() => {}"
    >
      <div
        v-if="error && isVisible"
        class="error-message-container"
        role="alert"
        aria-live="assertive"
      >
        <div class="error-content">
          <!-- Error Icon -->
          <div class="error-icon" aria-hidden="true">
            <svg
              xmlns="http://www.w3.org/2000/svg"
              viewBox="0 0 24 24"
              fill="currentColor"
              class="w-5 h-5"
            >
              <path
                fill-rule="evenodd"
                d="M9.401 3.003c1.155-2 4.043-2 5.197 0l7.355 12.748c1.154 2-.29 4.5-2.599 4.5H4.645c-2.309 0-3.752-2.5-2.598-4.5L9.4 3.003zM12 8.25a.75.75 0 01.75.75v3.75a.75.75 0 01-1.5 0V9a.75.75 0 01.75-.75zm0 8.25a.75.75 0 100-1.5.75.75 0 000 1.5z"
                clip-rule="evenodd"
              />
            </svg>
          </div>

          <!-- Error Message -->
          <div class="error-text">
            <p class="error-message">{{ error.message }}</p>
            
            <!-- Error Code (for debugging, hidden from screen readers) -->
            <p class="error-code" aria-hidden="true">
              Error Code: {{ error.code }}
            </p>
          </div>

          <!-- Action Buttons -->
          <div class="error-actions">
            <!-- Retry Button (only for retryable errors) -->
            <button
              v-if="error.retryable"
              @click="handleRetry"
              class="retry-button"
              aria-label="Retry action"
            >
              Retry
            </button>

            <!-- Dismiss Button -->
            <button
              @click="handleDismiss"
              class="dismiss-button"
              aria-label="Dismiss error message"
            >
              <svg
                xmlns="http://www.w3.org/2000/svg"
                viewBox="0 0 24 24"
                fill="currentColor"
                class="w-5 h-5"
              >
                <path
                  fill-rule="evenodd"
                  d="M5.47 5.47a.75.75 0 011.06 0L12 10.94l5.47-5.47a.75.75 0 111.06 1.06L13.06 12l5.47 5.47a.75.75 0 11-1.06 1.06L12 13.06l-5.47 5.47a.75.75 0 01-1.06-1.06L10.94 12 5.47 6.53a.75.75 0 010-1.06z"
                  clip-rule="evenodd"
                />
              </svg>
            </button>
          </div>
        </div>

        <!-- Auto-dismiss indicator for non-critical errors -->
        <div
          v-if="!isCriticalError(error)"
          class="auto-dismiss-indicator"
          aria-hidden="true"
        >
          <div class="auto-dismiss-bar"></div>
        </div>
      </div>
    </transition>
  </div>
</template>

<style scoped>
.error-display-container {
  @apply fixed top-4 right-4 z-50 max-w-md;
}

.connection-status {
  @apply mb-2 bg-white border border-gray-300 rounded-lg shadow-md;
}

.connection-status-content {
  @apply flex items-center gap-2 px-4 py-2;
}

.connection-indicator {
  @apply text-xl leading-none;
}

.connection-text {
  @apply text-sm font-medium text-gray-700;
}

.error-message-container {
  @apply bg-red-50 border border-red-300 rounded-lg shadow-lg overflow-hidden;
}

.error-content {
  @apply flex items-start gap-3 p-4;
}

.error-icon {
  @apply flex-shrink-0 text-red-600;
}

.error-text {
  @apply flex-1 min-w-0;
}

.error-message {
  @apply text-sm font-medium text-red-800 mb-1;
}

.error-code {
  @apply text-xs text-red-600 opacity-75;
}

.error-actions {
  @apply flex items-center gap-2 flex-shrink-0;
}

.retry-button {
  @apply px-3 py-1 text-sm font-medium text-red-700 bg-red-100 rounded hover:bg-red-200 focus:outline-none focus:ring-2 focus:ring-red-500 focus:ring-offset-2 transition-colors;
}

.dismiss-button {
  @apply p-1 text-red-600 hover:text-red-800 focus:outline-none focus:ring-2 focus:ring-red-500 rounded transition-colors;
}

.auto-dismiss-indicator {
  @apply h-1 bg-red-200 overflow-hidden;
}

.auto-dismiss-bar {
  @apply h-full bg-red-500;
  animation: auto-dismiss 5s linear forwards;
}

@keyframes auto-dismiss {
  from {
    width: 100%;
  }
  to {
    width: 0%;
  }
}

/* Transition animations */
.error-fade-enter-active,
.error-fade-leave-active {
  @apply transition-all duration-300;
}

.error-fade-enter-from {
  @apply opacity-0 translate-x-4;
}

.error-fade-leave-to {
  @apply opacity-0 translate-x-4;
}

/* Reduced motion support */
@media (prefers-reduced-motion: reduce) {
  .error-fade-enter-active,
  .error-fade-leave-active {
    @apply transition-none;
  }

  .error-fade-enter-from,
  .error-fade-leave-to {
    @apply translate-x-0;
  }

  .auto-dismiss-bar {
    animation: none;
  }
}
</style>
