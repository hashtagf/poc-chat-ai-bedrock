package bedrock

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentruntime/types"
	"github.com/aws/smithy-go"

	"github.com/bedrock-chat-poc/backend/domain/entities"
	"github.com/bedrock-chat-poc/backend/domain/services"
)

// Adapter implements the BedrockService interface using AWS SDK v2
type Adapter struct {
	client  *bedrockagentruntime.Client
	agentID string
	aliasID string
	config  AdapterConfig
}

// AdapterConfig holds configuration for the Bedrock adapter
type AdapterConfig struct {
	// MaxRetries is the maximum number of retry attempts for rate limits
	MaxRetries int
	// InitialBackoff is the initial backoff duration for retries
	InitialBackoff time.Duration
	// MaxBackoff is the maximum backoff duration for retries
	MaxBackoff time.Duration
	// RequestTimeout is the timeout for individual requests
	RequestTimeout time.Duration
}

// DefaultConfig returns the default adapter configuration
func DefaultConfig() AdapterConfig {
	return AdapterConfig{
		MaxRetries:     3,
		InitialBackoff: 1 * time.Second,
		MaxBackoff:     30 * time.Second,
		RequestTimeout: 60 * time.Second,
	}
}

// NewAdapter creates a new Bedrock adapter
func NewAdapter(ctx context.Context, agentID, aliasID string, cfg AdapterConfig) (*Adapter, error) {
	if agentID == "" {
		return nil, fmt.Errorf("agentID is required")
	}
	if aliasID == "" {
		return nil, fmt.Errorf("aliasID is required")
	}

	// Load AWS configuration using IAM roles
	awsCfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := bedrockagentruntime.NewFromConfig(awsCfg)

	return &Adapter{
		client:  client,
		agentID: agentID,
		aliasID: aliasID,
		config:  cfg,
	}, nil
}

// InvokeAgent sends a message to the Bedrock agent and returns the complete response
func (a *Adapter) InvokeAgent(ctx context.Context, input services.AgentInput) (*services.AgentResponse, error) {
	// Validate input
	if err := a.validateInput(input); err != nil {
		return nil, &services.DomainError{
			Code:      services.ErrCodeInvalidInput,
			Message:   "Invalid input",
			Retryable: false,
			Cause:     err,
		}
	}

	// Create request with timeout
	reqCtx, cancel := context.WithTimeout(ctx, a.config.RequestTimeout)
	defer cancel()

	// Build the invoke request
	invokeInput := &bedrockagentruntime.InvokeAgentInput{
		AgentId:   aws.String(a.agentID),
		AgentAliasId: aws.String(a.aliasID),
		SessionId: aws.String(input.SessionID),
		InputText: aws.String(input.Message),
	}

	// Execute with retry logic
	var response *bedrockagentruntime.InvokeAgentOutput
	var err error

	for attempt := 0; attempt <= a.config.MaxRetries; attempt++ {
		if attempt > 0 {
			// Calculate exponential backoff
			backoff := a.calculateBackoff(attempt)
			log.Printf("[Bedrock] Retry attempt %d after %v (RequestID: %s)", attempt, backoff, getRequestID(err))
			
			select {
			case <-time.After(backoff):
			case <-reqCtx.Done():
				return nil, a.transformError(reqCtx.Err(), "")
			}
		}

		log.Printf("[Bedrock] InvokeAgent request - SessionID: %s, AgentID: %s", input.SessionID, a.agentID)
		response, err = a.client.InvokeAgent(reqCtx, invokeInput)
		
		if err == nil {
			break
		}

		// Check if error is retryable
		if !a.isRetryable(err) {
			break
		}
	}

	if err != nil {
		requestID := getRequestID(err)
		log.Printf("[Bedrock] InvokeAgent failed - RequestID: %s, Error: %v", requestID, err)
		return nil, a.transformError(err, requestID)
	}

	// Process the streaming response
	return a.processInvokeResponse(ctx, response)
}

// InvokeAgentStream sends a message to the Bedrock agent and returns a streaming response
func (a *Adapter) InvokeAgentStream(ctx context.Context, input services.AgentInput) (services.StreamReader, error) {
	// Validate input
	if err := a.validateInput(input); err != nil {
		return nil, &services.DomainError{
			Code:      services.ErrCodeInvalidInput,
			Message:   "Invalid input",
			Retryable: false,
			Cause:     err,
		}
	}

	// Build the invoke request
	invokeInput := &bedrockagentruntime.InvokeAgentInput{
		AgentId:      aws.String(a.agentID),
		AgentAliasId: aws.String(a.aliasID),
		SessionId:    aws.String(input.SessionID),
		InputText:    aws.String(input.Message),
	}

	// Execute with retry logic
	var response *bedrockagentruntime.InvokeAgentOutput
	var err error

	for attempt := 0; attempt <= a.config.MaxRetries; attempt++ {
		if attempt > 0 {
			// Calculate exponential backoff
			backoff := a.calculateBackoff(attempt)
			log.Printf("[Bedrock] Stream retry attempt %d after %v", attempt, backoff)
			
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return nil, a.transformError(ctx.Err(), "")
			}
		}

		log.Printf("[Bedrock] InvokeAgentStream request - SessionID: %s, AgentID: %s", input.SessionID, a.agentID)
		response, err = a.client.InvokeAgent(ctx, invokeInput)
		
		if err == nil {
			break
		}

		// Check if error is retryable
		if !a.isRetryable(err) {
			break
		}
	}

	if err != nil {
		requestID := getRequestID(err)
		log.Printf("[Bedrock] InvokeAgentStream failed - RequestID: %s, Error: %v", requestID, err)
		return nil, a.transformError(err, requestID)
	}

	// Return stream reader
	stream := response.GetStream()
	if stream == nil {
		return nil, &services.DomainError{
			Code:      services.ErrCodeServiceError,
			Message:   "No event stream in response",
			Retryable: false,
		}
	}
	return newStreamReader(ctx, stream, getRequestID(err)), nil
}

// validateInput validates the agent input
func (a *Adapter) validateInput(input services.AgentInput) error {
	if input.SessionID == "" {
		return errors.New("session ID is required")
	}
	if input.Message == "" {
		return errors.New("message is required")
	}
	if len(input.Message) > 25000 {
		return errors.New("message exceeds maximum length of 25000 characters")
	}
	return nil
}

// processInvokeResponse processes the complete invoke response
func (a *Adapter) processInvokeResponse(ctx context.Context, output *bedrockagentruntime.InvokeAgentOutput) (*services.AgentResponse, error) {
	response := &services.AgentResponse{
		Content:   "",
		Citations: []entities.Citation{},
		Metadata:  make(map[string]interface{}),
	}

	// Process event stream
	stream := output.GetStream()
	if stream == nil {
		return response, nil
	}

	for event := range stream.Events() {
		switch e := event.(type) {
		case *types.ResponseStreamMemberChunk:
			// Extract text content
			if e.Value.Bytes != nil {
				response.Content += string(e.Value.Bytes)
			}

			// Extract citations if available
			if e.Value.Attribution != nil && e.Value.Attribution.Citations != nil {
				for _, citation := range e.Value.Attribution.Citations {
					response.Citations = append(response.Citations, a.convertCitation(citation))
				}
			}

		case *types.ResponseStreamMemberTrace:
			// Log trace information for debugging
			log.Printf("[Bedrock] Trace event received")

		default:
			log.Printf("[Bedrock] Unknown event type: %T", e)
		}
	}

	// Check for stream errors
	if err := stream.Err(); err != nil {
		return nil, a.transformError(err, "")
	}

	log.Printf("[Bedrock] InvokeAgent completed - Content length: %d, Citations: %d", len(response.Content), len(response.Citations))
	return response, nil
}

// convertCitation converts a Bedrock citation to domain citation
func (a *Adapter) convertCitation(citation types.Citation) entities.Citation {
	domainCitation := entities.Citation{
		Metadata: make(map[string]interface{}),
	}

	if citation.GeneratedResponsePart != nil && citation.GeneratedResponsePart.TextResponsePart != nil {
		domainCitation.Excerpt = aws.ToString(citation.GeneratedResponsePart.TextResponsePart.Text)
	}

	if len(citation.RetrievedReferences) > 0 {
		ref := citation.RetrievedReferences[0]
		
		if ref.Content != nil && ref.Content.Text != nil {
			domainCitation.SourceName = aws.ToString(ref.Content.Text)
		}

		if ref.Location != nil && ref.Location.S3Location != nil {
			domainCitation.SourceID = aws.ToString(ref.Location.S3Location.Uri)
			domainCitation.URL = aws.ToString(ref.Location.S3Location.Uri)
		}

		if ref.Metadata != nil {
			for k, v := range ref.Metadata {
				domainCitation.Metadata[k] = v
			}
		}
	}

	return domainCitation
}

// calculateBackoff calculates exponential backoff duration
func (a *Adapter) calculateBackoff(attempt int) time.Duration {
	backoff := float64(a.config.InitialBackoff) * math.Pow(2, float64(attempt-1))
	if backoff > float64(a.config.MaxBackoff) {
		backoff = float64(a.config.MaxBackoff)
	}
	return time.Duration(backoff)
}

// isRetryable determines if an error is retryable
func (a *Adapter) isRetryable(err error) bool {
	if err == nil {
		return false
	}

	// Check for context errors (timeout, cancellation)
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return false
	}

	// Check for AWS SDK errors
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		code := apiErr.ErrorCode()
		
		// Retryable error codes
		switch code {
		case "ThrottlingException", "TooManyRequestsException", "ServiceUnavailableException":
			return true
		}
	}

	return false
}

// transformError transforms AWS SDK errors to domain errors
func (a *Adapter) transformError(err error, requestID string) error {
	if err == nil {
		return nil
	}

	// Context errors
	if errors.Is(err, context.DeadlineExceeded) {
		return &services.DomainError{
			Code:      services.ErrCodeTimeout,
			Message:   "Request timed out",
			Retryable: true,
			Cause:     err,
		}
	}

	if errors.Is(err, context.Canceled) {
		return &services.DomainError{
			Code:      services.ErrCodeNetworkError,
			Message:   "Request canceled",
			Retryable: false,
			Cause:     err,
		}
	}

	// AWS SDK errors
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		code := apiErr.ErrorCode()
		message := apiErr.ErrorMessage()

		log.Printf("[Bedrock] AWS API Error - Code: %s, Message: %s, RequestID: %s", code, message, requestID)

		switch code {
		case "ThrottlingException", "TooManyRequestsException":
			return &services.DomainError{
				Code:      services.ErrCodeRateLimit,
				Message:   "Rate limit exceeded. Please try again later.",
				Retryable: true,
				Cause:     err,
			}

		case "ValidationException", "InvalidParameterException":
			return &services.DomainError{
				Code:      services.ErrCodeInvalidInput,
				Message:   "Invalid input parameters",
				Retryable: false,
				Cause:     err,
			}

		case "AccessDeniedException", "UnauthorizedException":
			return &services.DomainError{
				Code:      services.ErrCodeUnauthorized,
				Message:   "Unauthorized access to Bedrock service",
				Retryable: false,
				Cause:     err,
			}

		case "ServiceUnavailableException", "InternalServerException":
			return &services.DomainError{
				Code:      services.ErrCodeServiceError,
				Message:   "Service temporarily unavailable",
				Retryable: true,
				Cause:     err,
			}

		default:
			return &services.DomainError{
				Code:      services.ErrCodeServiceError,
				Message:   fmt.Sprintf("Bedrock service error: %s", message),
				Retryable: false,
				Cause:     err,
			}
		}
	}

	// Generic error
	return &services.DomainError{
		Code:      services.ErrCodeServiceError,
		Message:   "An unexpected error occurred",
		Retryable: false,
		Cause:     err,
	}
}

// getRequestID extracts the request ID from an error
func getRequestID(err error) string {
	if err == nil {
		return ""
	}

	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		// Try to extract request ID from error metadata
		// This is AWS SDK specific and may vary
		return "unknown"
	}

	return ""
}
