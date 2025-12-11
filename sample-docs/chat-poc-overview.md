# Chat POC Application Overview

## Project Description
This is a Proof of Concept (POC) for a chat interface powered by Amazon Bedrock Agent Core with knowledge base integration. The application demonstrates the feasibility of conversational AI using Bedrock Agent Core with context-aware responses.

## Architecture
The application follows Hexagonal Architecture (Ports & Adapters) with Clean Code, TDD, DDD, and SOLID principles.

### Technology Stack
- **Backend**: Go (idiomatic patterns, standard library)
- **Frontend**: Vue 3 Composition API + Tailwind CSS
- **AI/ML**: Amazon Bedrock Agent Core
- **Data**: MongoDB with vector support
- **Infrastructure**: Docker, Terraform, Kafka
- **Monitoring**: LGTM Stack

## Key Features
1. **Real-time Chat Interface**: WebSocket-based chat with streaming responses
2. **Knowledge Base Integration**: Context-aware responses from integrated knowledge bases
3. **Session Management**: Persistent chat sessions with message history
4. **Error Handling**: Comprehensive error handling with retry logic
5. **Scalable Architecture**: Microservices architecture with Docker containers

## API Endpoints
- `POST /api/sessions` - Create new chat session
- `GET /api/sessions/{id}` - Get session details
- `WebSocket /api/chat` - Real-time chat communication

## Configuration
The application requires the following environment variables:
- `BEDROCK_AGENT_ID` - Amazon Bedrock Agent ID
- `BEDROCK_AGENT_ALIAS_ID` - Agent Alias ID
- `BEDROCK_KNOWLEDGE_BASE_ID` - Knowledge Base ID
- `AWS_REGION` - AWS Region (us-east-1)

## Deployment
The application can be deployed using Docker Compose or Terraform for AWS infrastructure.