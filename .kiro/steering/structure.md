---
inclusion: always
---

# Project Structure & Architecture

This project follows **Hexagonal Architecture** (Ports & Adapters) with Clean Code, TDD, DDD, and SOLID principles.

## Directory Structure

Organize code by architectural layers:

```
src/
├── domain/           # Business logic, entities, value objects (framework-agnostic)
│   ├── entities/     # Domain entities
│   ├── value-objects/
│   └── repositories/ # Repository interfaces (ports)
├── application/      # Use cases, application services
│   └── use-cases/    # Business workflows
├── infrastructure/   # External adapters (AWS Bedrock, databases, APIs)
│   ├── bedrock/      # Amazon Bedrock Agent Core integration
│   ├── knowledge-base/
│   └── repositories/ # Repository implementations
└── interfaces/       # Input adapters (CLI, API, UI)
    └── chat/         # Chat interface implementation

tests/
├── unit/            # Fast, isolated tests for domain logic
├── integration/     # Tests with external dependencies
└── e2e/             # End-to-end scenarios
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
