# Documentation Index

Welcome to the Bedrock Chat UI documentation! This index helps you find the right documentation for your needs.

## Quick Start

New to the project? Start here:

1. **[README](../README.md)** - Project overview, quick start, and basic usage
2. **[DOCKER.md](../DOCKER.md)** - Docker setup and deployment
3. **[Configuration Guide](../backend/docs/CONFIGURATION.md)** - Environment configuration

## For Users

### Getting Started
- **[README](../README.md)** - Installation and setup instructions
- **[Quick Start Guide](../README.md#quick-start)** - Get up and running in minutes
- **[Configuration Guide](../backend/docs/CONFIGURATION.md)** - Configure the application

### Using the Application
- **[API Documentation](../backend/docs/API.md)** - Complete API reference
- **[Component Documentation](../README.md#component-documentation)** - Frontend components and props
- **[Troubleshooting Guide](../TROUBLESHOOTING.md)** - Common issues and solutions

### Deployment
- **[Docker Setup](../DOCKER.md)** - Containerized deployment
- **[Deployment Guide](../README.md#deployment)** - Production deployment options
- **[Configuration Guide](../backend/docs/CONFIGURATION.md)** - Production configuration

## For Developers

### Development Setup
- **[Contributing Guide](../CONTRIBUTING.md)** - How to contribute
- **[Development Guide](../README.md#development)** - Development workflow
- **[Coding Standards](../CONTRIBUTING.md#coding-standards)** - Code style guidelines

### Architecture
- **[Design Document](../.kiro/specs/chat-ui/design.md)** - System design and architecture
- **[Requirements](../.kiro/specs/chat-ui/requirements.md)** - Feature requirements
- **[Project Structure](../README.md#project-structure)** - Directory organization

### Backend Development
- **[Backend README](../backend/README.md)** - Backend-specific documentation
- **[Bedrock Adapter](../backend/infrastructure/bedrock/README.md)** - AWS Bedrock integration
- **[API Documentation](../backend/docs/API.md)** - API endpoints and protocols
- **[Configuration Guide](../backend/docs/CONFIGURATION.md)** - Backend configuration

### Frontend Development
- **[Frontend README](../frontend/README.md)** - Frontend-specific documentation
- **[Component Documentation](../README.md#component-documentation)** - Vue components
- **[Type Definitions](../README.md#type-definitions)** - TypeScript types

### Testing
- **[Testing Guide](../README.md#testing)** - Running tests
- **[Test Requirements](../CONTRIBUTING.md#testing-requirements)** - Writing tests
- **[Design Document - Testing Strategy](../.kiro/specs/chat-ui/design.md#testing-strategy)** - Test approach

## Reference Documentation

### API Reference
- **[API Documentation](../backend/docs/API.md)** - Complete API reference
  - REST endpoints
  - WebSocket protocol
  - Error codes
  - Examples

### Configuration Reference
- **[Configuration Guide](../backend/docs/CONFIGURATION.md)** - All configuration options
  - Environment variables
  - AWS credentials
  - Bedrock configuration
  - WebSocket settings
  - Logging configuration

### Component Reference
- **[Component Documentation](../README.md#component-documentation)** - All components
  - Props and events
  - Usage examples
  - Type definitions

## Troubleshooting

### Common Issues
- **[Troubleshooting Guide](../TROUBLESHOOTING.md)** - Comprehensive troubleshooting
  - Backend issues
  - Frontend issues
  - WebSocket issues
  - Bedrock integration issues
  - Docker issues
  - Test failures

### Quick Diagnostics
- **[Quick Diagnostics](../TROUBLESHOOTING.md#quick-diagnostics)** - Fast issue detection
- **[Common Error Messages](../TROUBLESHOOTING.md#common-error-messages)** - Error explanations
- **[Debug Mode](../TROUBLESHOOTING.md#debug-mode)** - Enable detailed logging

## Specifications

### Feature Specifications
- **[Requirements Document](../.kiro/specs/chat-ui/requirements.md)** - Feature requirements
- **[Design Document](../.kiro/specs/chat-ui/design.md)** - System design
- **[Implementation Tasks](../.kiro/specs/chat-ui/tasks.md)** - Development tasks

### Development Guidelines
- **[Product Context](../.kiro/steering/product.md)** - Product goals and priorities
- **[Architecture Guidelines](../.kiro/steering/structure.md)** - Hexagonal architecture
- **[Technology Standards](../.kiro/steering/tech.md)** - Tech stack standards

## By Topic

### AWS Bedrock
- [Bedrock Adapter](../backend/infrastructure/bedrock/README.md)
- [Configuration Guide - Bedrock Section](../backend/docs/CONFIGURATION.md#bedrock-configuration)
- [Troubleshooting - Bedrock Issues](../TROUBLESHOOTING.md#bedrock-integration-issues)

### WebSocket
- [API Documentation - WebSocket Protocol](../backend/docs/API.md#websocket-protocol)
- [Configuration Guide - WebSocket Section](../backend/docs/CONFIGURATION.md#websocket-configuration)
- [Troubleshooting - WebSocket Issues](../TROUBLESHOOTING.md#websocket-issues)

### Docker
- [Docker Setup Guide](../DOCKER.md)
- [Troubleshooting - Docker Issues](../TROUBLESHOOTING.md#docker-issues)
- [Deployment Guide](../README.md#deployment)

### Testing
- [Testing Guide](../README.md#testing)
- [Test Requirements](../CONTRIBUTING.md#testing-requirements)
- [Troubleshooting - Test Failures](../TROUBLESHOOTING.md#test-failures)

### Configuration
- [Configuration Guide](../backend/docs/CONFIGURATION.md)
- [Environment Variables](../README.md#configuration)
- [.env.example](../.env.example)

## By Role

### I'm a User
1. [README](../README.md) - Start here
2. [Quick Start](../README.md#quick-start) - Get running
3. [Troubleshooting](../TROUBLESHOOTING.md) - Fix issues

### I'm a Developer
1. [Contributing Guide](../CONTRIBUTING.md) - Start here
2. [Development Guide](../README.md#development) - Development workflow
3. [Design Document](../.kiro/specs/chat-ui/design.md) - Architecture

### I'm a DevOps Engineer
1. [Docker Setup](../DOCKER.md) - Start here
2. [Deployment Guide](../README.md#deployment) - Deploy to production
3. [Configuration Guide](../backend/docs/CONFIGURATION.md) - Configure services

### I'm a QA Engineer
1. [Testing Guide](../README.md#testing) - Start here
2. [API Documentation](../backend/docs/API.md) - Test endpoints
3. [Troubleshooting](../TROUBLESHOOTING.md) - Debug issues

## External Resources

### AWS Documentation
- [Amazon Bedrock Documentation](https://docs.aws.amazon.com/bedrock/)
- [Bedrock Agent Core](https://docs.aws.amazon.com/bedrock/latest/userguide/agents.html)
- [AWS SDK for Go v2](https://aws.github.io/aws-sdk-go-v2/)

### Framework Documentation
- [Vue 3 Documentation](https://vuejs.org/)
- [Go Documentation](https://go.dev/doc/)
- [Vite Documentation](https://vitejs.dev/)
- [Tailwind CSS](https://tailwindcss.com/)

### Testing Libraries
- [Vitest Documentation](https://vitest.dev/)
- [fast-check Documentation](https://fast-check.dev/)
- [Vue Test Utils](https://test-utils.vuejs.org/)

## Document Status

| Document | Status | Last Updated |
|----------|--------|--------------|
| README.md | ✅ Complete | 2024-01-01 |
| DOCKER.md | ✅ Complete | 2024-01-01 |
| CONTRIBUTING.md | ✅ Complete | 2024-01-01 |
| TROUBLESHOOTING.md | ✅ Complete | 2024-01-01 |
| backend/docs/API.md | ✅ Complete | 2024-01-01 |
| backend/docs/CONFIGURATION.md | ✅ Complete | 2024-01-01 |
| backend/infrastructure/bedrock/README.md | ✅ Complete | 2024-01-01 |
| backend/README.md | ✅ Complete | 2024-01-01 |
| frontend/README.md | ✅ Complete | 2024-01-01 |

## Need Help?

Can't find what you're looking for?

1. **Search the documentation** - Use your browser's search (Ctrl+F / Cmd+F)
2. **Check the index** - This page lists all documentation
3. **Review troubleshooting** - [Troubleshooting Guide](../TROUBLESHOOTING.md)
4. **Ask for help** - Create a GitHub issue or contact support

## Contributing to Documentation

Found an error or want to improve the documentation?

1. Read the [Contributing Guide](../CONTRIBUTING.md)
2. Make your changes
3. Submit a pull request

Documentation improvements are always welcome!

---

**Last Updated:** 2024-01-01  
**Version:** 1.0.0
