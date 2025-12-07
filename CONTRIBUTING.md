# Contributing to Bedrock Chat UI

Thank you for your interest in contributing to the Bedrock Chat UI project! This document provides guidelines and instructions for contributing.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Process](#development-process)
- [Coding Standards](#coding-standards)
- [Testing Requirements](#testing-requirements)
- [Commit Guidelines](#commit-guidelines)
- [Pull Request Process](#pull-request-process)
- [Issue Reporting](#issue-reporting)

## Code of Conduct

### Our Pledge

We are committed to providing a welcoming and inclusive environment for all contributors, regardless of experience level, background, or identity.

### Expected Behavior

- Be respectful and considerate in all interactions
- Provide constructive feedback
- Focus on what is best for the project and community
- Show empathy towards other community members
- Accept constructive criticism gracefully

### Unacceptable Behavior

- Harassment, discrimination, or offensive comments
- Personal attacks or trolling
- Publishing others' private information
- Any conduct that would be inappropriate in a professional setting

## Getting Started

### Prerequisites

Before contributing, ensure you have:

- Go 1.21+ installed
- Node.js 18+ installed
- Git configured with your name and email
- AWS account with Bedrock access (for testing Bedrock features)
- Familiarity with the project architecture (see README.md)

### Setting Up Development Environment

1. **Fork the repository**
   ```bash
   # Click "Fork" on GitHub, then clone your fork
   git clone https://github.com/YOUR_USERNAME/bedrock-chat-poc.git
   cd bedrock-chat-poc
   ```

2. **Add upstream remote**
   ```bash
   git remote add upstream https://github.com/ORIGINAL_OWNER/bedrock-chat-poc.git
   ```

3. **Install dependencies**
   ```bash
   # Backend
   cd backend
   go mod download
   
   # Frontend
   cd ../frontend
   npm install
   ```

4. **Configure environment**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

5. **Verify setup**
   ```bash
   # Backend tests
   cd backend
   go test ./... -v
   
   # Frontend tests
   cd ../frontend
   npm test
   ```

## Development Process

### Branching Strategy

We use a feature branch workflow:

- `main` - Production-ready code
- `develop` - Integration branch for features
- `feature/*` - New features
- `bugfix/*` - Bug fixes
- `hotfix/*` - Urgent production fixes

### Creating a Feature Branch

```bash
# Update your local main branch
git checkout main
git pull upstream main

# Create a feature branch
git checkout -b feature/your-feature-name

# Or for bug fixes
git checkout -b bugfix/issue-description
```

### Making Changes

1. **Write code** following our coding standards
2. **Write tests** for new functionality
3. **Run tests** to ensure nothing breaks
4. **Update documentation** if needed
5. **Commit changes** with clear messages

### Keeping Your Branch Updated

```bash
# Fetch latest changes from upstream
git fetch upstream

# Rebase your branch on upstream/main
git rebase upstream/main

# Resolve any conflicts
# Then continue the rebase
git rebase --continue

# Force push to your fork (if already pushed)
git push origin feature/your-feature-name --force
```

## Coding Standards

### Backend (Go)

**Follow Go best practices:**

- Use `gofmt` for formatting
- Follow [Effective Go](https://go.dev/doc/effective_go) guidelines
- Use meaningful variable and function names
- Keep functions small and focused (< 50 lines ideal)
- Document all exported functions and types
- Handle errors explicitly, never ignore them
- Use context for cancellation and timeouts

**Example:**

```go
// Good: Clear, documented, error handling
// CreateSession creates a new conversation session with a unique identifier.
// Returns an error if session creation fails.
func (r *Repository) CreateSession(ctx context.Context, session *entities.Session) error {
    if session == nil {
        return fmt.Errorf("session cannot be nil")
    }
    
    if err := r.validate(session); err != nil {
        return fmt.Errorf("invalid session: %w", err)
    }
    
    r.mu.Lock()
    defer r.mu.Unlock()
    
    r.sessions[session.ID] = session
    return nil
}

// Bad: No documentation, poor error handling
func (r *Repository) CreateSession(ctx context.Context, s *entities.Session) error {
    r.sessions[s.ID] = s
    return nil
}
```

**Architecture:**

- Follow hexagonal architecture principles
- Domain layer has no external dependencies
- Use dependency injection via interfaces
- Keep business logic in domain layer
- Infrastructure adapters implement domain interfaces

### Frontend (TypeScript/Vue)

**Follow Vue and TypeScript best practices:**

- Use `<script setup>` with Composition API
- Extract reusable logic into composables
- Use TypeScript strict mode
- Use Tailwind utilities, avoid custom CSS
- Keep components small and focused
- Use meaningful prop and event names

**Example:**

```typescript
// Good: Type-safe, clear, composable
<script setup lang="ts">
import { ref, computed } from 'vue'
import type { Message } from '@/types'

interface Props {
  messages: Message[]
  isStreaming: boolean
}

interface Emits {
  (e: 'scroll-to-bottom'): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const hasMessages = computed(() => props.messages.length > 0)

const scrollToBottom = () => {
  emit('scroll-to-bottom')
}
</script>

// Bad: No types, unclear props
<script setup>
const props = defineProps(['data', 'loading'])
const emit = defineEmits(['update'])
</script>
```

**Component Structure:**

- Components for presentation
- Composables for business logic
- Types for data models
- Keep components under 200 lines
- One component per file

### Code Formatting

**Backend:**
```bash
# Format all Go files
gofmt -w .

# Run linter
go vet ./...

# Check for common issues
golint ./...
```

**Frontend:**
```bash
# Format all files
npm run format

# Run linter
npm run lint

# Fix auto-fixable issues
npm run lint -- --fix
```

## Testing Requirements

### Test Coverage

All contributions must include appropriate tests:

- **Backend**: Minimum 80% coverage for new code
- **Frontend**: Minimum 80% coverage for new code
- **Property tests**: For universal properties
- **Integration tests**: For critical user flows

### Backend Testing

**Unit Tests:**

```go
func TestCreateSession(t *testing.T) {
    tests := []struct {
        name    string
        session *entities.Session
        wantErr bool
    }{
        {
            name: "valid session",
            session: &entities.Session{
                ID:        "test-id",
                CreatedAt: time.Now(),
            },
            wantErr: false,
        },
        {
            name:    "nil session",
            session: nil,
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            repo := NewMemorySessionRepository()
            err := repo.CreateSession(context.Background(), tt.session)
            
            if (err != nil) != tt.wantErr {
                t.Errorf("CreateSession() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

**Run tests:**
```bash
cd backend
go test ./... -v
go test ./... -cover
```

### Frontend Testing

**Unit Tests:**

```typescript
import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import MessageInput from './MessageInput.vue'

describe('MessageInput', () => {
  it('should disable submit button when input is empty', () => {
    const wrapper = mount(MessageInput)
    const button = wrapper.find('button[type="submit"]')
    
    expect(button.attributes('disabled')).toBeDefined()
  })
  
  it('should emit submit event with content', async () => {
    const wrapper = mount(MessageInput)
    const input = wrapper.find('input')
    const button = wrapper.find('button[type="submit"]')
    
    await input.setValue('Hello')
    await button.trigger('click')
    
    expect(wrapper.emitted('submit')).toBeTruthy()
    expect(wrapper.emitted('submit')?.[0]).toEqual(['Hello'])
  })
})
```

**Property Tests:**

```typescript
import { test } from 'vitest'
import fc from 'fast-check'

// Feature: chat-ui, Property 2: Input validation prevents invalid submission
test('Property 2: whitespace-only messages are rejected', () => {
  fc.assert(
    fc.property(
      fc.stringOf(fc.constantFrom(' ', '\t', '\n', '\r')),
      (whitespace) => {
        const result = validateMessage(whitespace)
        return result === false
      }
    ),
    { numRuns: 100 }
  )
})
```

**Run tests:**
```bash
cd frontend
npm test
npm run test:coverage
```

### Test Requirements Checklist

Before submitting a PR, ensure:

- [ ] All new functions have unit tests
- [ ] All new components have component tests
- [ ] Property tests for universal properties
- [ ] Integration tests for new features
- [ ] All tests pass locally
- [ ] Test coverage meets minimum requirements
- [ ] No tests are skipped or disabled

## Commit Guidelines

### Commit Message Format

We follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

**Examples:**

```
feat(chat): add citation display component

Implement CitationDisplay component to show knowledge base citations
with expandable details including source name, excerpt, and confidence score.

Closes #123
```

```
fix(websocket): handle connection timeout correctly

Fix issue where WebSocket connections would hang indefinitely on timeout.
Now properly closes connection and triggers reconnection logic.

Fixes #456
```

```
docs(readme): update installation instructions

Add detailed steps for Docker setup and troubleshooting common issues.
```

### Commit Best Practices

- Write clear, descriptive commit messages
- Keep commits focused on a single change
- Reference issue numbers when applicable
- Use present tense ("add feature" not "added feature")
- Keep subject line under 72 characters
- Provide context in the body if needed

## Pull Request Process

### Before Submitting

1. **Update your branch**
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

2. **Run all tests**
   ```bash
   # Backend
   cd backend && go test ./... -v
   
   # Frontend
   cd frontend && npm test
   ```

3. **Run linters**
   ```bash
   # Backend
   cd backend && gofmt -w . && go vet ./...
   
   # Frontend
   cd frontend && npm run lint
   ```

4. **Update documentation** if needed

5. **Ensure clean commit history**
   ```bash
   # Squash commits if needed
   git rebase -i upstream/main
   ```

### Creating a Pull Request

1. **Push to your fork**
   ```bash
   git push origin feature/your-feature-name
   ```

2. **Open PR on GitHub**
   - Go to the original repository
   - Click "New Pull Request"
   - Select your fork and branch
   - Fill out the PR template

3. **PR Title Format**
   ```
   feat(scope): brief description
   ```

4. **PR Description Template**
   ```markdown
   ## Description
   Brief description of changes
   
   ## Type of Change
   - [ ] Bug fix
   - [ ] New feature
   - [ ] Breaking change
   - [ ] Documentation update
   
   ## Testing
   - [ ] Unit tests added/updated
   - [ ] Integration tests added/updated
   - [ ] All tests passing
   
   ## Checklist
   - [ ] Code follows project style guidelines
   - [ ] Self-review completed
   - [ ] Documentation updated
   - [ ] No new warnings generated
   
   ## Related Issues
   Closes #123
   ```

### PR Review Process

1. **Automated checks** must pass:
   - All tests pass
   - Linters pass
   - Build succeeds

2. **Code review** by maintainers:
   - At least one approval required
   - Address all review comments
   - Make requested changes

3. **Update PR** if needed:
   ```bash
   # Make changes
   git add .
   git commit -m "fix: address review comments"
   git push origin feature/your-feature-name
   ```

4. **Merge** once approved:
   - Maintainers will merge your PR
   - Delete your feature branch after merge

## Issue Reporting

### Before Creating an Issue

1. **Search existing issues** to avoid duplicates
2. **Check documentation** for solutions
3. **Try latest version** to see if issue is fixed
4. **Gather information** about the issue

### Creating a Bug Report

Use the bug report template:

```markdown
## Bug Description
Clear description of the bug

## Steps to Reproduce
1. Go to '...'
2. Click on '...'
3. See error

## Expected Behavior
What should happen

## Actual Behavior
What actually happens

## Environment
- OS: [e.g., macOS 13.0]
- Go version: [e.g., 1.21.0]
- Node version: [e.g., 18.0.0]
- Browser: [e.g., Chrome 120]

## Additional Context
Screenshots, logs, etc.
```

### Creating a Feature Request

Use the feature request template:

```markdown
## Feature Description
Clear description of the feature

## Use Case
Why is this feature needed?

## Proposed Solution
How should it work?

## Alternatives Considered
Other approaches you've thought about

## Additional Context
Mockups, examples, etc.
```

## Questions?

If you have questions about contributing:

1. Check the [README](README.md) and other documentation
2. Search existing issues and discussions
3. Ask in GitHub Discussions
4. Contact maintainers

## Recognition

Contributors will be recognized in:
- GitHub contributors list
- Release notes for significant contributions
- Project documentation

Thank you for contributing to Bedrock Chat UI! 🎉
