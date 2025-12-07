---
inclusion: always
---

---
inclusion: always
---

# Technology Stack

## Stack Overview

- **Backend**: Go (idiomatic patterns, standard library)
- **Frontend**: Vue 3 Composition API + Tailwind CSS
- **AI/ML**: Amazon Bedrock Agent Core
- **Data**: MongoDB with vector support
- **Infrastructure**: Docker, Terraform, Kafka
- **Monitoring**: LGTM Stack

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

## Kafka Integration

- Consumer groups for horizontal scaling
- Idempotent message processing (handle duplicates)
- Dead letter queues for failed messages
- Monitor consumer lag metrics
- Graceful shutdown with context cancellation

## Common Commands

```bash
# Build
go build -o bin/app ./cmd/app

# Test with coverage
go test ./... -v -cover

# Local development
docker-compose up --build

# Format all code
gofmt -w . && terraform fmt -recursive

# Deploy to environment
terraform apply -var-file=environments/prod.tfvars
```
