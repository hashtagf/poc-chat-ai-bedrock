---
inclusion: always
---

# Project Structure & Architecture

This project follows **Hexagonal Architecture** (Ports & Adapters) with Clean Code, TDD, DDD, and SOLID principles.

## Directory Structure

Current project structure following Hexagonal Architecture:

```
backend/
├── domain/           # Business logic, entities, value objects (framework-agnostic)
│   ├── entities/     # ChatSession, Message, KnowledgeQuery
│   ├── repositories/ # Repository interfaces (ports)
│   └── services/     # Domain services (bedrock_service.go)
├── infrastructure/   # External adapters (AWS Bedrock, MongoDB, etc.)
│   ├── bedrock/      # Bedrock Agent Core integration (adapter.go)
│   └── repositories/ # Repository implementations
├── interfaces/       # Input adapters (WebSocket, HTTP API)
│   └── chat/         # Chat interface implementation
├── cmd/             # Application entry points
│   ├── server/       # Main server application
│   └── wsclient/     # WebSocket client for testing
└── config/          # Configuration management

frontend/
├── src/             # Vue 3 + TypeScript application
│   ├── components/   # Reusable Vue components
│   ├── composables/  # Composition API logic
│   └── views/        # Page components
└── dist/            # Built application

terraform/
├── modules/         # Reusable Terraform modules
│   ├── bedrock-agent/     # Agent and alias configuration
│   ├── knowledge-base/    # S3 Vectors knowledge base
│   └── iam/              # IAM roles and policies
└── environments/    # Environment-specific configurations
    └── dev/         # Development environment (us-east-1)
```

## Architectural Principles

**Dependency Rule**: Dependencies point inward. Domain has no external dependencies.
- Domain → Application → Infrastructure/Interfaces
- Use dependency injection for all external dependencies
- Define ports (interfaces) in domain/application, implement adapters in infrastructure

**Hexagonal Architecture**:
- Domain layer: Pure business logic, no framework dependencies
- Application layer: Orchestrates domain objects, defines use case interfaces
- Infrastructure: AWS SDK, Bedrock clients, external services
- Interfaces: User-facing entry points (chat UI, CLI, API)

**SOLID Principles**:
- Single Responsibility: One class, one reason to change
- Open/Closed: Extend behavior via interfaces, not modification
- Liskov Substitution: Subtypes must be substitutable for base types
- Interface Segregation: Small, focused interfaces over large ones
- Dependency Inversion: Depend on abstractions, not concretions

## Naming Conventions

- Use descriptive, intention-revealing names
- Entities: Nouns representing domain concepts (e.g., `ChatSession`, `KnowledgeQuery`)
- Use Cases: Verb phrases (e.g., `SendMessageToAgent`, `RetrieveFromKnowledgeBase`)
- Repositories: `<Entity>Repository` interface, `<Technology><Entity>Repository` implementation
- Value Objects: Immutable, self-validating (e.g., `AgentId`, `MessageContent`)

## Testing Strategy

**TDD Workflow**: Red → Green → Refactor
- Write failing test first
- Write minimal code to pass
- Refactor while keeping tests green

**Test Organization**:
- Unit tests: Test domain logic in isolation, mock all dependencies
- Integration tests: Test infrastructure adapters with real AWS services (use localstack or test accounts)
- E2E tests: Test complete user workflows

**Test Naming**: `should_<expected_behavior>_when_<condition>`

## Code Organization Rules

- Keep files small and focused (< 200 lines ideal)
- One public class/interface per file
- Co-locate related domain concepts
- Separate tests mirror source structure
- Configuration files in root or dedicated config directory

## Domain-Driven Design

- Use ubiquitous language from AWS Bedrock domain
- Model aggregates around consistency boundaries
- Entities have identity, Value Objects don't
- Domain events for cross-aggregate communication
- Repositories only for aggregate roots
