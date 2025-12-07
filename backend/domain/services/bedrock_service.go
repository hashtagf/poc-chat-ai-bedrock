package services

import (
	"context"

	"github.com/bedrock-chat-poc/backend/domain/entities"
)

// BedrockService defines the interface for interacting with Amazon Bedrock Agent Core
// This is a port in hexagonal architecture - the domain defines what it needs
type BedrockService interface {
	// InvokeAgent sends a message to the Bedrock agent and returns the complete response
	InvokeAgent(ctx context.Context, input AgentInput) (*AgentResponse, error)

	// InvokeAgentStream sends a message to the Bedrock agent and returns a streaming response
	InvokeAgentStream(ctx context.Context, input AgentInput) (StreamReader, error)
}

// AgentInput represents the input to the Bedrock agent
type AgentInput struct {
	SessionID        string
	Message          string
	KnowledgeBaseIDs []string
}

// AgentResponse represents the complete response from the Bedrock agent
type AgentResponse struct {
	Content   string
	Citations []entities.Citation
	Metadata  map[string]interface{}
	RequestID string
}

// StreamReader provides an interface for reading streaming responses
type StreamReader interface {
	// Read returns the next chunk of content, a done flag, and any error
	Read() (chunk string, done bool, err error)

	// ReadCitation returns the next citation if available
	ReadCitation() (*entities.Citation, error)

	// Close closes the stream reader
	Close() error
}

// DomainError represents errors that occur in the domain layer
type DomainError struct {
	Code      string
	Message   string
	Retryable bool
	Cause     error
}

func (e *DomainError) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}

func (e *DomainError) Unwrap() error {
	return e.Cause
}

// Common error codes
const (
	ErrCodeRateLimit       = "RATE_LIMIT_EXCEEDED"
	ErrCodeInvalidInput    = "INVALID_INPUT"
	ErrCodeServiceError    = "SERVICE_ERROR"
	ErrCodeNetworkError    = "NETWORK_ERROR"
	ErrCodeTimeout         = "TIMEOUT"
	ErrCodeUnauthorized    = "UNAUTHORIZED"
	ErrCodeMalformedStream = "MALFORMED_STREAM"
)
