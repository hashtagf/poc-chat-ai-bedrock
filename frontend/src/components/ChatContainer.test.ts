import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { nextTick } from 'vue'
import ChatContainer from './ChatContainer.vue'
import MessageList from './MessageList.vue'
import MessageInput from './MessageInput.vue'
import ErrorDisplay from './ErrorDisplay.vue'

// Mock the composables
vi.mock('@/composables/useChatService', () => ({
  useChatService: vi.fn(() => ({
    sendMessage: vi.fn(),
    streamingMessage: { value: '' },
    isStreaming: { value: false },
    error: { value: null },
    clearError: vi.fn()
  }))
}))

vi.mock('@/composables/useConversationHistory', () => ({
  useConversationHistory: vi.fn(() => ({
    messages: { value: [] },
    addMessage: vi.fn(),
    clearHistory: vi.fn(),
    getMessageById: vi.fn()
  }))
}))

vi.mock('@/composables/useSessionManager', () => ({
  useSessionManager: vi.fn(() => ({
    currentSessionId: { value: 'test-session-id' },
    createNewSession: vi.fn().mockResolvedValue('new-session-id'),
    loadSession: vi.fn(),
    sessionMetadata: {
      value: {
        id: 'test-session-id',
        createdAt: new Date(),
        messageCount: 0
      }
    }
  }))
}))

vi.mock('@/composables/useErrorHandler', () => ({
  useErrorHandler: vi.fn(() => ({
    currentError: { value: null },
    errorHistory: { value: [] },
    handleError: vi.fn(),
    clearError: vi.fn(),
    retry: vi.fn()
  }))
}))

describe('ChatContainer', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should render the chat container', async () => {
    const wrapper = mount(ChatContainer)
    
    // Wait for initialization
    await new Promise(resolve => setTimeout(resolve, 600))
    await nextTick()
    
    expect(wrapper.find('.chat-container').exists()).toBe(true)
  })

  it('should display loading state initially', () => {
    const wrapper = mount(ChatContainer)
    
    expect(wrapper.find('.loading-container').exists()).toBe(true)
    expect(wrapper.text()).toContain('Initializing chat')
  })

  it('should render all sub-components after initialization', async () => {
    const wrapper = mount(ChatContainer)
    
    // Wait for initialization
    await new Promise(resolve => setTimeout(resolve, 600))
    await nextTick()
    
    expect(wrapper.findComponent(MessageList).exists()).toBe(true)
    expect(wrapper.findComponent(MessageInput).exists()).toBe(true)
    expect(wrapper.findComponent(ErrorDisplay).exists()).toBe(true)
  })

  it('should display session information', async () => {
    const wrapper = mount(ChatContainer)
    
    // Wait for initialization
    await new Promise(resolve => setTimeout(resolve, 600))
    await nextTick()
    
    expect(wrapper.find('.session-info').exists()).toBe(true)
    expect(wrapper.text()).toContain('Session:')
    expect(wrapper.text()).toContain('messages')
  })

  it('should have a new session button', async () => {
    const wrapper = mount(ChatContainer)
    
    // Wait for initialization
    await new Promise(resolve => setTimeout(resolve, 600))
    await nextTick()
    
    const newSessionButton = wrapper.find('.new-session-button')
    expect(newSessionButton.exists()).toBe(true)
    expect(newSessionButton.text()).toContain('New Session')
  })

  it('should have ARIA live region for screen readers', async () => {
    const wrapper = mount(ChatContainer)
    
    // Wait for initialization
    await new Promise(resolve => setTimeout(resolve, 600))
    await nextTick()
    
    const liveRegion = wrapper.find('[role="status"][aria-live="polite"]')
    expect(liveRegion.exists()).toBe(true)
  })

  it('should have proper semantic HTML structure', async () => {
    const wrapper = mount(ChatContainer)
    
    // Wait for initialization
    await new Promise(resolve => setTimeout(resolve, 600))
    await nextTick()
    
    expect(wrapper.find('[role="main"]').exists()).toBe(true)
    expect(wrapper.find('[role="banner"]').exists()).toBe(true)
  })

  it('should disable input when streaming', async () => {
    const { useChatService } = await import('@/composables/useChatService')
    const mockChatService = useChatService as any
    
    mockChatService.mockReturnValue({
      sendMessage: vi.fn(),
      streamingMessage: { value: 'streaming content' },
      isStreaming: { value: true },
      error: { value: null },
      clearError: vi.fn()
    })
    
    const wrapper = mount(ChatContainer)
    
    // Wait for initialization
    await new Promise(resolve => setTimeout(resolve, 600))
    await nextTick()
    
    const messageInput = wrapper.findComponent(MessageInput)
    expect(messageInput.props('disabled')).toBe(true)
  })

  it('should format session date correctly', async () => {
    const wrapper = mount(ChatContainer)
    
    // Wait for initialization
    await new Promise(resolve => setTimeout(resolve, 600))
    await nextTick()
    
    const sessionTime = wrapper.find('.session-time')
    expect(sessionTime.exists()).toBe(true)
    // Should show "Just now" or similar for a new session
    expect(sessionTime.text()).toBeTruthy()
  })
})
