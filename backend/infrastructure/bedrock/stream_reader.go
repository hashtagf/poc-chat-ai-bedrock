package bedrock

import (
	"context"
	"errors"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentruntime/types"

	"github.com/bedrock-chat-poc/backend/domain/entities"
	"github.com/bedrock-chat-poc/backend/domain/services"
)

// streamReader implements the StreamReader interface for Bedrock event streams
type streamReader struct {
	ctx       context.Context
	stream    *bedrockagentruntime.InvokeAgentEventStream
	buffer    []string
	citations []entities.Citation
	done      bool
	requestID string
	eventChan <-chan types.ResponseStream
}

// newStreamReader creates a new stream reader
func newStreamReader(ctx context.Context, stream *bedrockagentruntime.InvokeAgentEventStream, requestID string) services.StreamReader {
	return &streamReader{
		ctx:       ctx,
		stream:    stream,
		buffer:    make([]string, 0),
		citations: make([]entities.Citation, 0),
		done:      false,
		requestID: requestID,
		eventChan: stream.Events(),
	}
}

// Read returns the next chunk of content, a done flag, and any error
func (sr *streamReader) Read() (chunk string, done bool, err error) {
	// Check if already done
	if sr.done {
		return "", true, nil
	}

	// Check context cancellation
	select {
	case <-sr.ctx.Done():
		sr.done = true
		return "", true, sr.ctx.Err()
	default:
	}

	// If we have buffered content, return it
	if len(sr.buffer) > 0 {
		chunk = sr.buffer[0]
		sr.buffer = sr.buffer[1:]
		return chunk, false, nil
	}

	// Read next event from stream
	event, ok := <-sr.eventChan
	if !ok {
		// Channel closed, check for errors
		if err := sr.stream.Err(); err != nil {
			sr.done = true
			log.Printf("[Bedrock] Stream error - RequestID: %s, Error: %v", sr.requestID, err)
			return "", true, sr.transformStreamError(err)
		}
		
		sr.done = true
		log.Printf("[Bedrock] Stream completed - RequestID: %s", sr.requestID)
		return "", true, nil
	}

	// Process event
	switch e := event.(type) {
	case *types.ResponseStreamMemberChunk:
		// Extract text content
		if e.Value.Bytes != nil {
			content := string(e.Value.Bytes)
			log.Printf("[Bedrock] Stream chunk received - Length: %d, RequestID: %s", len(content), sr.requestID)
			
			// Store citations for later retrieval
			if e.Value.Attribution != nil && e.Value.Attribution.Citations != nil {
				for _, citation := range e.Value.Attribution.Citations {
					sr.citations = append(sr.citations, convertCitation(citation))
				}
			}

			return content, false, nil
		}

	case *types.ResponseStreamMemberTrace:
		// Log trace information for debugging
		log.Printf("[Bedrock] Trace event received - RequestID: %s", sr.requestID)
		// Continue to next event
		return sr.Read()

	default:
		log.Printf("[Bedrock] Unknown event type: %T - RequestID: %s", e, sr.requestID)
		// Continue to next event
		return sr.Read()
	}

	// No content in this event, read next
	return sr.Read()
}

// ReadCitation returns the next citation if available
func (sr *streamReader) ReadCitation() (*entities.Citation, error) {
	if len(sr.citations) == 0 {
		return nil, nil
	}

	citation := sr.citations[0]
	sr.citations = sr.citations[1:]
	return &citation, nil
}

// Close closes the stream reader
func (sr *streamReader) Close() error {
	sr.done = true
	log.Printf("[Bedrock] Stream reader closed - RequestID: %s", sr.requestID)
	return nil
}

// transformStreamError transforms streaming errors to domain errors
func (sr *streamReader) transformStreamError(err error) error {
	if err == nil {
		return nil
	}

	// Check for context errors
	if errors.Is(err, context.DeadlineExceeded) {
		return &services.DomainError{
			Code:      services.ErrCodeTimeout,
			Message:   "Stream timed out",
			Retryable: true,
			Cause:     err,
		}
	}

	if errors.Is(err, context.Canceled) {
		return &services.DomainError{
			Code:      services.ErrCodeNetworkError,
			Message:   "Stream canceled",
			Retryable: false,
			Cause:     err,
		}
	}

	// Generic stream error
	return &services.DomainError{
		Code:      services.ErrCodeMalformedStream,
		Message:   "Error reading from stream",
		Retryable: false,
		Cause:     err,
	}
}

// convertCitation converts a Bedrock citation to domain citation
func convertCitation(citation types.Citation) entities.Citation {
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
