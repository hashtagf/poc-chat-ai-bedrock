<script setup lang="ts">
import { ref, computed } from 'vue'
import type { Citation } from '../types'

// Props
interface CitationDisplayProps {
  citations?: Citation[]
  expanded?: boolean
}

const props = withDefaults(defineProps<CitationDisplayProps>(), {
  citations: () => [],
  expanded: false
})

// State
const expandedCitations = ref<Set<string>>(new Set())

// Computed
const hasCitations = computed(() => {
  return props.citations && props.citations.length > 0
})

const citationCount = computed(() => {
  return props.citations?.length || 0
})

// Methods
const toggleCitation = (sourceId: string) => {
  if (expandedCitations.value.has(sourceId)) {
    expandedCitations.value.delete(sourceId)
  } else {
    expandedCitations.value.add(sourceId)
  }
}

const isCitationExpanded = (sourceId: string) => {
  return props.expanded || expandedCitations.value.has(sourceId)
}

const formatConfidence = (confidence?: number) => {
  if (confidence === undefined) return null
  return `${Math.round(confidence * 100)}%`
}

/**
 * Handle keyboard events for citation interaction
 */
const handleCitationKeydown = (event: KeyboardEvent, sourceId: string) => {
  if (event.key === 'Enter' || event.key === ' ') {
    event.preventDefault()
    toggleCitation(sourceId)
  } else if (event.key === 'Escape' && isCitationExpanded(sourceId)) {
    event.preventDefault()
    toggleCitation(sourceId)
  }
}
</script>

<template>
  <div 
    class="citation-display"
    role="region"
    :aria-label="hasCitations ? `${citationCount} citation${citationCount > 1 ? 's' : ''}` : 'No citations'"
  >
    <!-- No citations indicator -->
    <div 
      v-if="!hasCitations"
      class="no-citations"
      role="status"
    >
      <svg 
        xmlns="http://www.w3.org/2000/svg" 
        viewBox="0 0 20 20" 
        fill="currentColor" 
        class="no-citations-icon"
        aria-hidden="true"
      >
        <path 
          fill-rule="evenodd" 
          d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a.75.75 0 000 1.5h.253a.25.25 0 01.244.304l-.459 2.066A1.75 1.75 0 0010.747 15H11a.75.75 0 000-1.5h-.253a.25.25 0 01-.244-.304l.459-2.066A1.75 1.75 0 009.253 9H9z" 
          clip-rule="evenodd" 
        />
      </svg>
      <span class="no-citations-text">Response based on general knowledge</span>
    </div>

    <!-- Citations list -->
    <div 
      v-else
      class="citations-list"
    >
      <div 
        v-for="(citation, index) in citations" 
        :key="citation.sourceId"
        class="citation-item"
        :class="{ 
          'citation-expanded': isCitationExpanded(citation.sourceId),
          'citation-multiple': citationCount > 1
        }"
      >
        <!-- Citation header (clickable) -->
        <button
          class="citation-header"
          :aria-expanded="isCitationExpanded(citation.sourceId)"
          :aria-controls="`citation-details-${citation.sourceId}`"
          @click="toggleCitation(citation.sourceId)"
          @keydown="handleCitationKeydown($event, citation.sourceId)"
        >
          <div class="citation-header-content">
            <!-- Citation number badge -->
            <span 
              class="citation-badge"
              :aria-label="`Citation ${index + 1}`"
            >
              {{ index + 1 }}
            </span>
            
            <!-- Source name -->
            <span class="citation-source-name">
              {{ citation.sourceName }}
            </span>
            
            <!-- Confidence score (if available) -->
            <span 
              v-if="citation.confidence !== undefined"
              class="citation-confidence"
              :aria-label="`Confidence: ${formatConfidence(citation.confidence)}`"
            >
              {{ formatConfidence(citation.confidence) }}
            </span>
          </div>
          
          <!-- Expand/collapse icon -->
          <svg 
            xmlns="http://www.w3.org/2000/svg" 
            viewBox="0 0 20 20" 
            fill="currentColor" 
            class="citation-toggle-icon"
            :class="{ 'citation-toggle-expanded': isCitationExpanded(citation.sourceId) }"
            aria-hidden="true"
          >
            <path 
              fill-rule="evenodd" 
              d="M5.23 7.21a.75.75 0 011.06.02L10 11.168l3.71-3.938a.75.75 0 111.08 1.04l-4.25 4.5a.75.75 0 01-1.08 0l-4.25-4.5a.75.75 0 01.02-1.06z" 
              clip-rule="evenodd" 
            />
          </svg>
        </button>

        <!-- Citation details (expandable) -->
        <div
          v-if="isCitationExpanded(citation.sourceId)"
          :id="`citation-details-${citation.sourceId}`"
          class="citation-details"
        >
          <!-- Excerpt -->
          <div class="citation-excerpt">
            <p class="citation-excerpt-label">Excerpt:</p>
            <p class="citation-excerpt-text">{{ citation.excerpt }}</p>
          </div>

          <!-- URL (if available) -->
          <div 
            v-if="citation.url"
            class="citation-url"
          >
            <a 
              :href="citation.url"
              target="_blank"
              rel="noopener noreferrer"
              class="citation-link"
            >
              View source
              <svg 
                xmlns="http://www.w3.org/2000/svg" 
                viewBox="0 0 20 20" 
                fill="currentColor" 
                class="citation-link-icon"
                aria-hidden="true"
              >
                <path 
                  fill-rule="evenodd" 
                  d="M4.25 5.5a.75.75 0 00-.75.75v8.5c0 .414.336.75.75.75h8.5a.75.75 0 00.75-.75v-4a.75.75 0 011.5 0v4A2.25 2.25 0 0112.75 17h-8.5A2.25 2.25 0 012 14.75v-8.5A2.25 2.25 0 014.25 4h5a.75.75 0 010 1.5h-5z" 
                  clip-rule="evenodd" 
                />
                <path 
                  fill-rule="evenodd" 
                  d="M6.194 12.753a.75.75 0 001.06.053L16.5 4.44v2.81a.75.75 0 001.5 0v-4.5a.75.75 0 00-.75-.75h-4.5a.75.75 0 000 1.5h2.553l-9.056 8.194a.75.75 0 00-.053 1.06z" 
                  clip-rule="evenodd" 
                />
              </svg>
            </a>
          </div>

          <!-- Metadata (if available) -->
          <div 
            v-if="citation.metadata && Object.keys(citation.metadata).length > 0"
            class="citation-metadata"
          >
            <p class="citation-metadata-label">Additional information:</p>
            <dl class="citation-metadata-list">
              <template v-for="(value, key) in citation.metadata" :key="key">
                <dt class="citation-metadata-key">{{ key }}:</dt>
                <dd class="citation-metadata-value">{{ value }}</dd>
              </template>
            </dl>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.citation-display {
  @apply mt-3;
}

/* No citations indicator */
.no-citations {
  @apply flex items-center gap-2 px-3 py-2 text-sm text-gray-600 bg-gray-50 border border-gray-200 rounded-lg;
}

.no-citations-icon {
  @apply w-5 h-5 flex-shrink-0;
}

.no-citations-text {
  @apply font-medium;
}

/* Citations list */
.citations-list {
  @apply space-y-2;
}

/* Citation item */
.citation-item {
  @apply border border-gray-200 rounded-lg overflow-hidden bg-white transition-all;
}

.citation-item:hover {
  @apply border-gray-300 shadow-sm;
}

.citation-expanded {
  @apply border-blue-300 shadow-md;
}

/* Multiple citations get distinct colors */
.citation-multiple:nth-child(1) {
  @apply border-l-4 border-l-blue-500;
}

.citation-multiple:nth-child(2) {
  @apply border-l-4 border-l-green-500;
}

.citation-multiple:nth-child(3) {
  @apply border-l-4 border-l-purple-500;
}

.citation-multiple:nth-child(4) {
  @apply border-l-4 border-l-orange-500;
}

.citation-multiple:nth-child(n+5) {
  @apply border-l-4 border-l-gray-500;
}

/* Citation header */
.citation-header {
  @apply w-full flex items-center justify-between px-3 py-2 text-left cursor-pointer hover:bg-gray-50 transition-colors;
}

.citation-header:focus {
  @apply outline-none ring-2 ring-blue-500 ring-inset;
}

.citation-header-content {
  @apply flex items-center gap-2 flex-1 min-w-0;
}

/* Citation badge */
.citation-badge {
  @apply flex-shrink-0 w-6 h-6 flex items-center justify-center text-xs font-bold text-white bg-blue-600 rounded-full;
}

.citation-multiple:nth-child(1) .citation-badge {
  @apply bg-blue-600;
}

.citation-multiple:nth-child(2) .citation-badge {
  @apply bg-green-600;
}

.citation-multiple:nth-child(3) .citation-badge {
  @apply bg-purple-600;
}

.citation-multiple:nth-child(4) .citation-badge {
  @apply bg-orange-600;
}

.citation-multiple:nth-child(n+5) .citation-badge {
  @apply bg-gray-600;
}

/* Source name */
.citation-source-name {
  @apply flex-1 font-medium text-gray-900 truncate;
}

/* Confidence score */
.citation-confidence {
  @apply flex-shrink-0 px-2 py-1 text-xs font-semibold text-green-700 bg-green-100 rounded-full;
}

/* Toggle icon */
.citation-toggle-icon {
  @apply w-5 h-5 text-gray-400 transition-transform flex-shrink-0;
}

.citation-toggle-expanded {
  @apply rotate-180;
}

/* Citation details */
.citation-details {
  @apply px-3 pb-3 space-y-3 border-t border-gray-100;
}

/* Excerpt */
.citation-excerpt {
  @apply pt-3;
}

.citation-excerpt-label {
  @apply text-xs font-semibold text-gray-500 uppercase tracking-wide mb-1;
}

.citation-excerpt-text {
  @apply text-sm text-gray-700 italic;
}

/* URL */
.citation-url {
  @apply pt-2;
}

.citation-link {
  @apply inline-flex items-center gap-1 text-sm font-medium text-blue-600 hover:text-blue-800 hover:underline;
}

.citation-link:focus {
  @apply outline-none ring-2 ring-blue-500 ring-offset-2 rounded;
}

.citation-link-icon {
  @apply w-4 h-4;
}

/* Metadata */
.citation-metadata {
  @apply pt-2;
}

.citation-metadata-label {
  @apply text-xs font-semibold text-gray-500 uppercase tracking-wide mb-2;
}

.citation-metadata-list {
  @apply grid grid-cols-[auto_1fr] gap-x-3 gap-y-1 text-sm;
}

.citation-metadata-key {
  @apply font-medium text-gray-600;
}

.citation-metadata-value {
  @apply text-gray-900;
}

/* Reduced motion support */
@media (prefers-reduced-motion: reduce) {
  .citation-item,
  .citation-header,
  .citation-toggle-icon {
    transition: none;
  }
}
</style>
