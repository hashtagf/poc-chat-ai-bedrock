# Docker Setup Guide

This document provides instructions for running the Chat UI application using Docker.

## Prerequisites

- Docker Engine 20.10+
- Docker Compose 2.0+
- AWS credentials with Bedrock access
- Configured Bedrock Agent and Knowledge Base

## Quick Start

1. **Copy environment configuration:**
   ```bash
   cp .env.example .env
   ```

2. **Edit `.env` file with your AWS and Bedrock configuration:**
   - Set your AWS credentials (or use IAM roles in production)
   - Configure Bedrock Agent ID and Knowledge Base ID
   - Adjust other settings as needed

3. **Build and start all services:**
   ```bash
   docker-compose up --build
   ```

4. **Access the application:**
   - Frontend: http://localhost:5173
   - Backend API: http://localhost:8080
   - MongoDB: localhost:27017

## Services

### Backend (Go)
- **Port:** 8080
- **Health Check:** http://localhost:8080/health
- **Dependencies:** MongoDB, AWS Bedrock

### Frontend (Vue 3 + Nginx)
- **Port:** 5173 (mapped to 80 in container)
- **Health Check:** http://localhost:5173/
- **Dependencies:** Backend

### MongoDB
- **Port:** 27017
- **Database:** chatdb
- **Health Check:** Internal mongosh ping

## Environment Variables

### Required AWS Configuration

```bash
AWS_REGION=ap-southeast-1                    # AWS region for Bedrock
AWS_ACCESS_KEY_ID=<your-key>            # AWS access key (dev only)
AWS_SECRET_ACCESS_KEY=<your-secret>     # AWS secret key (dev only)
```

**Production Note:** Use IAM roles instead of hardcoded credentials.

### Required Bedrock Configuration

```bash
BEDROCK_AGENT_ID=<agent-id>                    # Your Bedrock Agent ID
BEDROCK_AGENT_ALIAS_ID=TSTALIASID              # Agent alias (default: test)
BEDROCK_KNOWLEDGE_BASE_ID=<kb-id>              # Knowledge Base ID
BEDROCK_MODEL_ID=anthropic.claude-v2           # Model to use
```

### Optional Configuration

```bash
# MongoDB
MONGO_ROOT_USERNAME=admin
MONGO_ROOT_PASSWORD=password
MONGO_DATABASE=chatdb

# WebSocket
WS_TIMEOUT=30s
WS_BUFFER_SIZE=8192

# Session
SESSION_TIMEOUT=30m

# Frontend
VITE_API_URL=http://localhost:8080
VITE_WS_URL=ws://localhost:8080
```

## Docker Commands

### Start services
```bash
docker-compose up
```

### Start in detached mode
```bash
docker-compose up -d
```

### Rebuild and start
```bash
docker-compose up --build
```

### Stop services
```bash
docker-compose down
```

### Stop and remove volumes
```bash
docker-compose down -v
```

### View logs
```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f backend
docker-compose logs -f frontend
docker-compose logs -f mongodb
```

### Check service health
```bash
docker-compose ps
```

## Health Checks

All services include health checks:

- **Backend:** Checks `/health` endpoint every 30s
- **Frontend:** Checks nginx root every 30s
- **MongoDB:** Checks database ping every 10s

Services will show as "healthy" once they pass health checks.

## Troubleshooting

### Backend fails to start

1. **Check AWS credentials:**
   ```bash
   docker-compose logs backend | grep -i aws
   ```

2. **Verify Bedrock configuration:**
   - Ensure Agent ID and Knowledge Base ID are correct
   - Verify IAM permissions for Bedrock access

3. **Check MongoDB connection:**
   ```bash
   docker-compose logs mongodb
   ```

### Frontend cannot connect to backend

1. **Verify backend is healthy:**
   ```bash
   docker-compose ps backend
   ```

2. **Check environment variables:**
   ```bash
   docker-compose config | grep VITE
   ```

3. **Test backend directly:**
   ```bash
   curl http://localhost:8080/health
   ```

### MongoDB connection issues

1. **Check MongoDB logs:**
   ```bash
   docker-compose logs mongodb
   ```

2. **Verify credentials in `.env` file**

3. **Test MongoDB connection:**
   ```bash
   docker-compose exec mongodb mongosh -u admin -p password
   ```

## Development Workflow

### Rebuild specific service
```bash
docker-compose up --build backend
docker-compose up --build frontend
```

### Run backend tests
```bash
docker-compose exec backend go test ./...
```

### Access MongoDB shell
```bash
docker-compose exec mongodb mongosh -u admin -p password chatdb
```

### Clean rebuild
```bash
docker-compose down -v
docker-compose build --no-cache
docker-compose up
```

## Production Considerations

1. **Use IAM Roles:** Remove AWS credentials from environment variables and use IAM roles for ECS/EKS

2. **Secure MongoDB:** 
   - Use strong passwords
   - Enable authentication
   - Use MongoDB Atlas or managed service

3. **HTTPS:** Configure nginx with SSL certificates

4. **Resource Limits:** Add resource constraints in docker-compose.yml:
   ```yaml
   deploy:
     resources:
       limits:
         cpus: '1'
         memory: 1G
   ```

5. **Logging:** Configure centralized logging (CloudWatch, ELK, etc.)

6. **Monitoring:** Add health check endpoints and monitoring

## Architecture

```
┌─────────────┐
│   Browser   │
└──────┬──────┘
       │ HTTP/WS
       ↓
┌─────────────┐
│  Frontend   │ (nginx:80 → :5173)
│  (Vue 3)    │
└──────┬──────┘
       │ HTTP/WS
       ↓
┌─────────────┐
│   Backend   │ (:8080)
│    (Go)     │
└──┬────┬─────┘
   │    │
   │    └─────→ AWS Bedrock Agent Core
   │
   ↓
┌─────────────┐
│   MongoDB   │ (:27017)
└─────────────┘
```

## Multi-Stage Builds

Both Dockerfiles use multi-stage builds for optimization:

### Backend
- **Stage 1 (builder):** Compiles Go binary
- **Stage 2 (runtime):** Minimal Alpine image with binary only

### Frontend
- **Stage 1 (builder):** Builds Vue app with npm
- **Stage 2 (runtime):** Nginx serves static files

This approach minimizes image size and improves security.
