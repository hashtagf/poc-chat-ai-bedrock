// Type definitions for the Chat UI
// This file will contain all TypeScript interfaces and types

export type MessageRole = 'user' | 'agent'
export type MessageStatus = 'sending' | 'sent' | 'error'

export interface Message {
  id: string
  role: MessageRole
  content: string
  timestamp: Date
  citations?: Citation[]
  status: MessageStatus
  errorMessage?: string
}

export interface Citation {
  sourceId: string
  sourceName: string
  excerpt: string
  confidence?: number
  url?: string
  metadata?: Record<string, any>
}

export interface Session {
  id: string
  createdAt: Date
  lastMessageAt?: Date
  messageCount: number
}

export interface ChatError {
  code: string
  message: string
  retryable: boolean
  details?: Record<string, any>
}

export interface SessionMetadata {
  id: string
  createdAt: Date
  messageCount: number
}

// Composable Interfaces

export interface ChatService {
  sendMessage(content: string): Promise<void>
  streamingMessage: import('vue').Ref<string>
  isStreaming: import('vue').Ref<boolean>
  error: import('vue').Ref<ChatError | null>
  clearError(): void
}

export interface ConversationHistory {
  messages: import('vue').Ref<Message[]>
  addMessage(message: Message): void
  clearHistory(): void
  getMessageById(id: string): Message | undefined
}

export interface SessionManager {
  currentSessionId: import('vue').Ref<string>
  createNewSession(): Promise<string>
  loadSession(sessionId: string): Promise<void>
  sessionMetadata: import('vue').Ref<SessionMetadata>
}

// Type Guards

export function isMessage(value: unknown): value is Message {
  if (typeof value !== 'object' || value === null) {
    return false
  }
  
  const obj = value as Record<string, unknown>
  
  return (
    typeof obj.id === 'string' &&
    (obj.role === 'user' || obj.role === 'agent') &&
    typeof obj.content === 'string' &&
    obj.timestamp instanceof Date &&
    (obj.status === 'sending' || obj.status === 'sent' || obj.status === 'error') &&
    (obj.citations === undefined || Array.isArray(obj.citations)) &&
    (obj.errorMessage === undefined || typeof obj.errorMessage === 'string')
  )
}

export function isCitation(value: unknown): value is Citation {
  if (typeof value !== 'object' || value === null) {
    return false
  }
  
  const obj = value as Record<string, unknown>
  
  return (
    typeof obj.sourceId === 'string' &&
    typeof obj.sourceName === 'string' &&
    typeof obj.excerpt === 'string' &&
    (obj.confidence === undefined || typeof obj.confidence === 'number') &&
    (obj.url === undefined || typeof obj.url === 'string') &&
    (obj.metadata === undefined || typeof obj.metadata === 'object')
  )
}

export function isSession(value: unknown): value is Session {
  if (typeof value !== 'object' || value === null) {
    return false
  }
  
  const obj = value as Record<string, unknown>
  
  return (
    typeof obj.id === 'string' &&
    obj.createdAt instanceof Date &&
    (obj.lastMessageAt === undefined || obj.lastMessageAt instanceof Date) &&
    typeof obj.messageCount === 'number'
  )
}

export function isChatError(value: unknown): value is ChatError {
  if (typeof value !== 'object' || value === null) {
    return false
  }
  
  const obj = value as Record<string, unknown>
  
  return (
    typeof obj.code === 'string' &&
    typeof obj.message === 'string' &&
    typeof obj.retryable === 'boolean' &&
    (obj.details === undefined || typeof obj.details === 'object')
  )
}

export function isMessageRole(value: unknown): value is MessageRole {
  return value === 'user' || value === 'agent'
}

export function isMessageStatus(value: unknown): value is MessageStatus {
  return value === 'sending' || value === 'sent' || value === 'error'
}
