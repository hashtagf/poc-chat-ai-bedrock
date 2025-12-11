---
inclusion: always
---

# Product Context

**Project Type**: POC for chat interface powered by Amazon Bedrock Agent Core with S3 Vectors knowledge base integration.

**Primary Objective**: Validate feasibility of conversational AI using Bedrock Agent Core with context-aware responses from S3 Vectors knowledge base.

**Current Status**: ✅ **FULLY FUNCTIONAL** - Infrastructure deployed, document ingestion working, queries returning results with 99% cost savings over OpenSearch Serverless.

## Requirements Clarification Protocol

Before implementing any feature, you MUST:

1. **Clarify scope**: Ask about expected behavior, inputs, outputs, and success criteria
2. **Identify edge cases**: Confirm error scenarios, validation rules, and boundary conditions
3. **Validate POC alignment**: Ensure the feature supports core objectives (Bedrock capabilities, knowledge base integration, conversational AI)
4. **Understand user impact**: Ask how this solves the user's problem

**Example questions to ask**:
- "Should the chat handle multi-turn conversations with context retention?"
- "What should happen when the knowledge base returns no relevant results?"
- "Do you need conversation history persistence or in-memory only?"

## Implementation Priorities (in order)

1. **Functional correctness**: Working features that meet requirements
2. **AWS best practices**: Follow Bedrock SDK patterns, IAM security, error handling
3. **User experience**: Clear responses, appropriate error messages, reasonable latency
4. **Code maintainability**: Simple, readable code suitable for POC iteration
5. **Cost awareness**: Minimize unnecessary Bedrock API calls

**POC mindset**: Prioritize speed to validation over premature optimization. Build for learning and iteration, not production scale.

## Critical Bedrock Integration Rules

**API Usage**:
- Always implement retry logic with exponential backoff for rate limits
- Log all Bedrock API calls with request IDs for debugging
- Handle streaming responses appropriately (don't buffer entire responses)
- Monitor token usage to stay within quotas

**Knowledge Base Integration**:
- Query knowledge bases before invoking agent when context is needed
- Handle empty or low-confidence results gracefully
- Cache frequent queries to reduce costs (if applicable)
- Validate knowledge base IDs exist before making calls

**Error Handling**:
- Catch and wrap AWS SDK errors with operation context
- Distinguish between retryable (throttling, timeouts) and non-retryable errors (invalid parameters, auth failures)
- Provide user-friendly error messages without exposing internal details
- Log full error details for debugging

**Security**:
- Use IAM roles for credentials—never hardcode access keys
- Validate and sanitize all user inputs before passing to Bedrock
- Implement appropriate timeouts to prevent hanging requests
- Follow principle of least privilege for IAM permissions

## Cost & Performance Considerations

- **Token limits**: Be aware of input/output token limits per model
- **API quotas**: Understand rate limits and implement backoff strategies
- **Knowledge base queries**: Each query incurs costs—optimize query frequency
- **Model selection**: Choose appropriate models for POC (balance cost vs capability)
- **S3 Vectors**: Cost-effective storage (~$5-10/month vs $700/month for OpenSearch)
- **Resource naming**: Keep names short to avoid S3 Vectors 2048-byte metadata limit

## S3 Vectors Specific Considerations

**Naming Conventions for Production**:
- Use short project names (2-4 characters): `kb` instead of `bedrock-chat-poc`
- Abbreviate resource types: `docs`, `vec`, `idx` instead of full words
- Avoid redundant prefixes and region suffixes
- **Critical**: Long resource names cause metadata to exceed 2048-byte limit

**Current Working Configuration**:
- Knowledge Base ID: `AQ5JOUEIGF`
- Agent ID: `W6R84XTD2X` 
- Region: `us-east-1` (broadest model support)
- Storage: S3 Vectors (1536 dimensions, cosine similarity)

## Documentation Requirements

When implementing features, document:

- **Integration patterns**: How you're calling Bedrock APIs, including code examples
- **Configuration**: Required environment variables, IAM permissions, knowledge base setup
- **Limitations**: Known constraints, unsupported scenarios, or workarounds
- **Decisions**: Why you chose specific approaches (e.g., model selection, error handling strategy)

Keep documentation inline with code (comments) or in README files—avoid separate documentation systems for POC.
