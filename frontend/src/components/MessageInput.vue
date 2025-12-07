<script setup lang="ts">
import { ref, computed } from "vue";

// Props
interface MessageInputProps {
  disabled?: boolean;
  placeholder?: string;
}

const props = withDefaults(defineProps<MessageInputProps>(), {
  disabled: false,
  placeholder: "Type your message...",
});

// Emits
interface MessageInputEmits {
  (e: "submit", content: string): void;
}

const emit = defineEmits<MessageInputEmits>();

// State
const inputContent = ref("");
const inputRef = ref<HTMLTextAreaElement | null>(null);

// Constants
const MAX_LENGTH = 2000;

// Computed
const isInputEmpty = computed(() => {
  return inputContent.value.trim().length === 0;
});

const isInputTooLong = computed(() => {
  return inputContent.value.length > MAX_LENGTH;
});

const canSubmit = computed(() => {
  return !props.disabled && !isInputEmpty.value && !isInputTooLong.value;
});

const characterCount = computed(() => {
  return inputContent.value.length;
});

const showCharacterCount = computed(() => {
  return characterCount.value > MAX_LENGTH * 0.8; // Show when 80% of limit reached
});

/**
 * Handle form submission
 */
const handleSubmit = (): void => {
  if (!canSubmit.value) {
    return;
  }

  const content = inputContent.value.trim();

  // Emit the submit event
  emit("submit", content);

  // Clear the input field
  inputContent.value = "";

  // Restore focus to input field
  if (inputRef.value) {
    inputRef.value.focus();
  }
};

/**
 * Handle Enter key press
 * Submit on Enter, allow Shift+Enter for new line
 */
const handleKeydown = (event: KeyboardEvent): void => {
  if (event.key === "Enter" && !event.shiftKey) {
    event.preventDefault();
    handleSubmit();
  }
};

/**
 * Focus the input field (exposed for parent components)
 */
const focus = (): void => {
  if (inputRef.value) {
    inputRef.value.focus();
  }
};

// Expose focus method
defineExpose({
  focus,
});
</script>

<template>
  <div class="message-input-container">
    <form @submit.prevent="handleSubmit" class="message-input-form">
      <div class="input-wrapper">
        <textarea
          ref="inputRef"
          v-model="inputContent"
          :placeholder="placeholder"
          :disabled="disabled"
          :aria-label="placeholder"
          :maxlength="MAX_LENGTH + 100"
          class="message-textarea"
          :class="{
            disabled: disabled,
            error: isInputTooLong,
          }"
          rows="1"
          @keydown="handleKeydown"
        />

        <!-- Character count indicator -->
        <div
          v-if="showCharacterCount"
          class="character-count"
          :class="{ error: isInputTooLong }"
          aria-live="polite"
        >
          {{ characterCount }} / {{ MAX_LENGTH }}
        </div>
      </div>

      <button
        type="submit"
        :disabled="!canSubmit"
        :aria-label="disabled ? 'Sending message...' : 'Send message'"
        class="send-button"
        :class="{
          disabled: !canSubmit,
          loading: disabled,
        }"
      >
        <span v-if="!disabled" class="send-icon">
          <!-- Send icon (paper plane) -->
          <svg
            xmlns="http://www.w3.org/2000/svg"
            viewBox="0 0 24 24"
            fill="currentColor"
            class="w-5 h-5"
            aria-hidden="true"
          >
            <path
              d="M3.478 2.405a.75.75 0 00-.926.94l2.432 7.905H13.5a.75.75 0 010 1.5H4.984l-2.432 7.905a.75.75 0 00.926.94 60.519 60.519 0 0018.445-8.986.75.75 0 000-1.218A60.517 60.517 0 003.478 2.405z"
            />
          </svg>
        </span>

        <!-- Loading indicator -->
        <span v-else class="loading-spinner" aria-hidden="true">
          <svg
            class="animate-spin h-5 w-5"
            xmlns="http://www.w3.org/2000/svg"
            fill="none"
            viewBox="0 0 24 24"
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
        </span>
      </button>
    </form>

    <!-- Validation error message -->
    <div
      v-if="isInputTooLong"
      class="error-message"
      role="alert"
      aria-live="assertive"
    >
      Message exceeds maximum length of {{ MAX_LENGTH }} characters
    </div>
  </div>
</template>

<style scoped>
.message-input-container {
  @apply w-full;
}

.message-input-form {
  @apply flex items-end gap-2 p-4 bg-white border-t border-gray-200;
}

.input-wrapper {
  @apply flex-1 relative;
}

.message-textarea {
  @apply w-full px-4 py-3 pr-20 border border-gray-300 rounded-lg resize-none focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all;
  min-height: 44px;
  max-height: 200px;
}

.message-textarea.disabled {
  @apply bg-gray-100 cursor-not-allowed opacity-60;
}

.message-textarea.error {
  @apply border-red-500 focus:ring-red-500;
}

.character-count {
  @apply absolute bottom-2 right-2 text-xs text-gray-500 pointer-events-none;
}

.character-count.error {
  @apply text-red-500 font-semibold;
}

.send-button {
  @apply flex items-center justify-center w-12 h-12 bg-blue-600 text-white rounded-lg hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 transition-all;
}

.send-button.disabled {
  @apply bg-gray-300 cursor-not-allowed hover:bg-gray-300;
}

.send-button.loading {
  @apply bg-blue-600 cursor-wait;
}

.send-icon,
.loading-spinner {
  @apply flex items-center justify-center;
}

.error-message {
  @apply px-4 py-2 text-sm text-red-600 bg-red-50 border-t border-red-200;
}

/* Reduced motion support */
@media (prefers-reduced-motion: reduce) {
  .message-textarea,
  .send-button {
    @apply transition-none;
  }

  .animate-spin {
    animation: none;
  }
}
</style>
