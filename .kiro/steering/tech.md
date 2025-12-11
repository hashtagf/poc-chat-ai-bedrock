---
inclusion: always
---

# Technology Stack

## Stack Overview

- **Backend**: Go 1.23+ (idiomatic patterns, AWS SDK v2)
- **Frontend**: Vue 3 Composition API + TypeScript + Tailwind CSS
- **AI/ML**: Amazon Bedrock Agent Core with Knowledge Base (S3 Vectors)
- **Data**: MongoDB 7.0 for session storage
- **Infrastructure**: Docker, Terraform, AWS (us-east-1)
- **WebSocket**: Real-time chat communication

## Go Code Standards

**Error Handling**: Never ignore errors. Always handle explicitly with proper context.

**Patterns**:
- Use `context.Context` for cancellation/timeouts in all I/O operations
- Dependency injection via interfaces, not concrete types
- Table-driven tests for comprehensive coverage
- Keep packages focused with single responsibility

**Formatting**: Run `gofmt` before committing. Follow Effective Go conventions.

**Example**:
```go
// Good: explicit error handling with context
func (s *Service) Process(ctx context.Context, id string) error {
    result, err := s.repo.Find(ctx, id)
    if err != nil {
        return fmt.Errorf("failed to find %s: %w", id, err)
    }
    return s.handle(ctx, result)
}
```

## Vue 3 Standards

- Use `<script setup>` with Composition API
- Extract reusable logic into composables
- Tailwind utilities only—avoid custom CSS
- TypeScript for type safety when available

## AWS Bedrock Integration

**Critical Rules**:
- AWS SDK for Go v2 only
- IAM roles for credentials—never hardcode keys
- Implement exponential backoff for rate limits
- Log all API calls with request IDs for debugging
- Cache knowledge base queries to reduce costs

**Error Handling**: Wrap Bedrock errors with context about the operation and input parameters.

## Docker Standards

- Multi-stage builds to minimize image size
- Pin exact versions—never use `latest`
- Include health checks in docker-compose.yml
- Use environment variables for all configuration
- Maintain `.dockerignore` to exclude build artifacts

## Terraform Standards

- Remote state with S3 + DynamoDB locking
- Separate tfvars per environment (dev/staging/prod)
- Tag all resources: `Environment`, `Project`, `ManagedBy`
- Run `terraform fmt` and `terraform validate` before commits
- Use modules for reusable infrastructure patterns

## WebSocket Integration

- Real-time bidirectional communication for chat
- Gorilla WebSocket library for Go backend
- Session management with MongoDB persistence
- Graceful connection handling and reconnection
- Message streaming for Bedrock Agent responses

## Current Infrastructure (Deployed)

**AWS Resources (us-east-1)**:
- Knowledge Base: `AQ5JOUEIGF` (S3 Vectors, ACTIVE)
- Bedrock Agent: `W6R84XTD2X` (PREPARED with alias)
- S3 Buckets: `kb-docs-dev-*`, `kb-vec-dev`
- Vector Index: `kb-idx-dev` (1536 dimensions, cosine)

**Local Development**:
- MongoDB 7.0 container for session storage
- Docker Compose with health checks
- Environment-based configuration

## Common Commands

```bash
# Backend development
cd backend
go build -o bin/server ./cmd/server
go test ./... -v -cover

# Frontend development  
cd frontend
npm run dev
npm run build
npm run test

# Infrastructure
cd terraform/environments/dev
terraform plan
terraform apply

# Local development (full stack)
docker-compose up --build

# Format all code
gofmt -w ./backend && terraform fmt -recursive ./terraform

# Test API endpoints
./backend/test_api.sh
```

## Environment Configuration

**Required Environment Variables**:
```bash
# AWS Configuration
AWS_REGION=us-east-1
BEDROCK_AGENT_ID=W6R84XTD2X
BEDROCK_AGENT_ALIAS_ID=TXENIZDWOS
BEDROCK_KNOWLEDGE_BASE_ID=AQ5JOUEIGF

# MongoDB
MONGO_URI=mongodb://admin:password@localhost:27017
MONGO_DATABASE=chatdb

# Server
PORT=8080
WS_TIMEOUT=30s
SESSION_TIMEOUT=30m
```
