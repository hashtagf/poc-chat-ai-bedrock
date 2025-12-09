# Troubleshooting Guide

Comprehensive troubleshooting guide for the Bedrock Chat UI application.

## Table of Contents

- [Quick Diagnostics](#quick-diagnostics)
- [Backend Issues](#backend-issues)
- [Frontend Issues](#frontend-issues)
- [WebSocket Issues](#websocket-issues)
- [Bedrock Integration Issues](#bedrock-integration-issues)
- [Docker Issues](#docker-issues)
- [Test Failures](#test-failures)
- [Performance Issues](#performance-issues)
- [Common Error Messages](#common-error-messages)
- [Debug Mode](#debug-mode)
- [Getting Help](#getting-help)

## Quick Diagnostics

Run these commands to quickly diagnose common issues:

```bash
# Check if backend is running
curl http://localhost:8080/health

# Check if frontend is accessible
curl http://localhost:5173

# Check Docker containers
docker-compose ps

# Check logs
docker-compose logs backend
docker-compose logs frontend

# Check Go version
go version  # Should be 1.21+

# Check Node version
node --version  # Should be 18+

# Check ports
lsof -i :8080  # Backend port
lsof -i :5173  # Frontend port
```

## Backend Issues

### Issue: Backend Won't Start

**Symptoms:**
- Server crashes immediately
- "Address already in use" error
- "Cannot connect to database" error

**Solutions:**

1. **Check if port is in use:**
   ```bash
   lsof -i :8080
   
   # Kill the process
   kill -9 <PID>
   
   # Or use a different port
   export SERVER_PORT=8081
   go run cmd/server/main.go
   ```

2. **Verify Go version:**
   ```bash
   go version
   # Should be 1.21 or higher
   
   # Update if needed
   # Download from https://go.dev/dl/
   ```

3. **Check dependencies:**
   ```bash
   cd backend
   go mod download
   go mod tidy
   go mod verify
   ```

4. **Check environment variables:**
   ```bash
   # Verify .env file exists
   ls -la .env
   
   # Check required variables
   grep -E "AWS_REGION|ENVIRONMENT" .env
   
   # Load environment
   export $(cat .env | grep -v '^#' | xargs)
   ```

5. **Run in debug mode:**
   ```bash
   export LOG_LEVEL=debug
   go run cmd/server/main.go
   ```

### Issue: Backend Crashes During Runtime

**Symptoms:**
- Server stops responding
- Panic messages in logs
- Memory errors

**Solutions:**

1. **Check logs for panic:**
   ```bash
   # Look for stack traces
   grep -A 20 "panic:" logs/server.log
   ```

2. **Check memory usage:**
   ```bash
   # Monitor memory
   top -p $(pgrep server)
   
   # Check for memory leaks
   go test ./... -memprofile=mem.prof
   go tool pprof mem.prof
   ```

3. **Check for race conditions:**
   ```bash
   go run -race cmd/server/main.go
   ```

4. **Verify database connections:**
   ```bash
   # Check MongoDB connection
   mongosh --eval "db.adminCommand('ping')"
   ```

### Issue: Slow Response Times

**Symptoms:**
- Requests take >5 seconds
- Timeouts
- High CPU usage

**Solutions:**

1. **Check Bedrock latency:**
   ```bash
   # Enable request timing logs
   export LOG_LEVEL=debug
   
   # Look for slow Bedrock calls
   grep "Bedrock request took" logs/server.log
   ```

2. **Profile the application:**
   ```bash
   # CPU profiling
   go test ./... -cpuprofile=cpu.prof
   go tool pprof cpu.prof
   
   # Look for hot spots
   (pprof) top10
   (pprof) list <function_name>
   ```

3. **Check database queries:**
   ```bash
   # Enable query logging
   export MONGO_LOG_LEVEL=debug
   ```

4. **Increase timeouts:**
   ```bash
   # In .env
   BEDROCK_REQUEST_TIMEOUT=120s
   WS_STREAM_TIMEOUT=10m
   ```

## Frontend Issues

### Issue: Frontend Won't Start

**Symptoms:**
- "Port already in use" error
- Build errors
- Dependency errors

**Solutions:**

1. **Check if port is in use:**
   ```bash
   lsof -i :5173
   
   # Vite will auto-select next port
   # Or specify port
   npm run dev -- --port 3000
   ```

2. **Verify Node version:**
   ```bash
   node --version
   # Should be 18 or higher
   
   # Update if needed
   nvm install 18
   nvm use 18
   ```

3. **Clear and reinstall dependencies:**
   ```bash
   cd frontend
   rm -rf node_modules package-lock.json
   npm install
   
   # Or use clean install
   npm ci
   ```

4. **Clear Vite cache:**
   ```bash
   rm -rf node_modules/.vite
   npm run dev
   ```

5. **Check for TypeScript errors:**
   ```bash
   npm run type-check
   ```

### Issue: Frontend Build Fails

**Symptoms:**
- Build errors during `npm run build`
- TypeScript errors
- Missing dependencies

**Solutions:**

1. **Check TypeScript errors:**
   ```bash
   npm run type-check
   
   # Fix errors in reported files
   ```

2. **Check for missing dependencies:**
   ```bash
   npm install
   npm audit fix
   ```

3. **Clear build cache:**
   ```bash
   rm -rf dist
   rm -rf node_modules/.vite
   npm run build
   ```

4. **Check environment variables:**
   ```bash
   # Verify .env file
   cat .env
   
   # Should have VITE_ prefixed variables
   VITE_API_URL=http://localhost:8080
   VITE_WS_URL=ws://localhost:8080
   ```

### Issue: Frontend Shows Blank Page

**Symptoms:**
- White screen
- No errors in console
- Network requests fail

**Solutions:**

1. **Check browser console:**
   ```
   Open DevTools (F12)
   Check Console tab for errors
   Check Network tab for failed requests
   ```

2. **Verify backend is running:**
   ```bash
   curl http://localhost:8080/health
   ```

3. **Check CORS errors:**
   ```
   Look for CORS errors in browser console
   Backend should allow frontend origin
   ```

4. **Clear browser cache:**
   ```
   Hard refresh: Ctrl+Shift+R (Windows/Linux) or Cmd+Shift+R (Mac)
   Or clear cache in browser settings
   ```

5. **Check API URL configuration:**
   ```bash
   # In frontend/.env
   VITE_API_URL=http://localhost:8080
   VITE_WS_URL=ws://localhost:8080
   
   # Restart frontend after changing
   ```

## WebSocket Issues

### Issue: WebSocket Connection Fails

**Symptoms:**
- "WebSocket connection failed" error
- Connection immediately closes
- No messages received

**Solutions:**

1. **Verify backend is running:**
   ```bash
   curl http://localhost:8080/health
   ```

2. **Check WebSocket URL:**
   ```javascript
   // Should be ws:// not http://
   const ws = new WebSocket('ws://localhost:8080/api/chat/stream');
   ```

3. **Check browser console:**
   ```
   Look for WebSocket errors
   Check if connection is being blocked
   ```

4. **Test with WebSocket client:**
   ```bash
   cd backend
   go build -o bin/wsclient cmd/wsclient/main.go
   ./bin/wsclient -session <session-id> -message "test"
   ```

5. **Check firewall:**
   ```bash
   # Allow port 8080
   sudo ufw allow 8080
   ```

### Issue: WebSocket Disconnects Frequently

**Symptoms:**
- Connection drops after a few seconds
- "Connection closed" errors
- Reconnection loops

**Solutions:**

1. **Increase timeouts:**
   ```bash
   # In .env
   WS_TIMEOUT=60s
   WS_STREAM_TIMEOUT=10m
   ```

2. **Check network stability:**
   ```bash
   # Ping backend
   ping localhost
   
   # Check for packet loss
   ```

3. **Implement reconnection logic:**
   ```javascript
   let reconnectAttempts = 0;
   const maxAttempts = 5;
   
   function connect() {
     const ws = new WebSocket('ws://localhost:8080/api/chat/stream');
     
     ws.onclose = () => {
       if (reconnectAttempts < maxAttempts) {
         const backoff = Math.min(1000 * Math.pow(2, reconnectAttempts), 30000);
         setTimeout(() => {
           reconnectAttempts++;
           connect();
         }, backoff);
       }
     };
     
     return ws;
   }
   ```

4. **Check for proxy issues:**
   ```bash
   # Disable proxy if using one
   unset HTTP_PROXY
   unset HTTPS_PROXY
   ```

### Issue: Messages Not Received

**Symptoms:**
- WebSocket connected but no messages
- Sent messages don't get responses
- Stream hangs

**Solutions:**

1. **Check message format:**
   ```javascript
   // Correct format
   ws.send(JSON.stringify({
     session_id: 'valid-uuid',
     content: 'message text'
   }));
   ```

2. **Verify session ID:**
   ```bash
   # Create new session
   curl -X POST http://localhost:8080/api/sessions
   
   # Use returned session ID
   ```

3. **Check backend logs:**
   ```bash
   docker-compose logs backend | grep -i error
   ```

4. **Test with simple message:**
   ```javascript
   ws.send(JSON.stringify({
     session_id: sessionId,
     content: 'hello'
   }));
   ```

## Bedrock Integration Issues

### Issue: Rate Limit Exceeded

**Symptoms:**
- "RATE_LIMIT_EXCEEDED" error
- 429 status code
- Requests fail after several attempts

**Solutions:**

1. **Wait and retry:**
   ```
   Wait 30-60 seconds before retrying
   Bedrock has per-minute and per-hour limits
   ```

2. **Increase retry configuration:**
   ```bash
   # In .env
   BEDROCK_MAX_RETRIES=5
   BEDROCK_INITIAL_BACKOFF=2s
   BEDROCK_MAX_BACKOFF=60s
   ```

3. **Check service quotas:**
   ```bash
   aws service-quotas list-service-quotas \
     --service-code bedrock \
     --region ap-southeast-1
   ```

4. **Request quota increase:**
   ```
   Go to AWS Service Quotas console
   Request increase for Bedrock quotas
   ```

### Issue: Invalid Agent ID

**Symptoms:**
- "Agent not found" error
- "Invalid agent ID" error
- 404 status code

**Solutions:**

1. **Verify agent ID in AWS Console:**
   ```
   Go to AWS Bedrock console
   Navigate to Agents
   Copy the correct Agent ID
   ```

2. **Update configuration:**
   ```bash
   # In .env
   BEDROCK_AGENT_ID=ABCDEFGHIJ
   BEDROCK_AGENT_ALIAS_ID=TSTALIASID
   ```

3. **Check agent status:**
   ```bash
   aws bedrock-agent get-agent \
     --agent-id ABCDEFGHIJ \
     --region ap-southeast-1
   ```

4. **Verify IAM permissions:**
   ```bash
   aws sts get-caller-identity
   
   # Check if role has bedrock:GetAgent permission
   ```

### Issue: Insufficient IAM Permissions

**Symptoms:**
- "AccessDeniedException" error
- "User is not authorized" error
- 403 status code

**Solutions:**

1. **Check current IAM identity:**
   ```bash
   aws sts get-caller-identity
   ```

2. **Verify required permissions:**
   ```json
   {
     "Version": "2012-10-17",
     "Statement": [
       {
         "Effect": "Allow",
         "Action": [
           "bedrock:InvokeAgent",
           "bedrock:InvokeAgentStream",
           "bedrock:GetAgent"
         ],
         "Resource": "*"
       }
     ]
   }
   ```

3. **Attach policy to IAM role/user:**
   ```bash
   aws iam attach-user-policy \
     --user-name your-user \
     --policy-arn arn:aws:iam::aws:policy/AmazonBedrockFullAccess
   ```

4. **Use IAM role (recommended):**
   ```bash
   # For EC2
   aws ec2 associate-iam-instance-profile \
     --instance-id i-1234567890abcdef0 \
     --iam-instance-profile Name=BedrockRole
   ```

### Issue: Timeout Errors

**Symptoms:**
- "Request timed out" error
- "Stream timeout" error
- Long wait times

**Solutions:**

1. **Increase timeouts:**
   ```bash
   # In .env
   BEDROCK_REQUEST_TIMEOUT=120s
   WS_STREAM_TIMEOUT=10m
   WS_CHUNK_TIMEOUT=60s
   ```

2. **Check network latency:**
   ```bash
   # Ping Bedrock endpoint
   ping bedrock-agent-runtime.ap-southeast-1.amazonaws.com
   ```

3. **Use closer AWS region:**
   ```bash
   # In .env
   AWS_REGION=ap-southeast-1  # Choose closest region
   ```

4. **Check Bedrock service status:**
   ```
   Visit AWS Service Health Dashboard
   Check for Bedrock outages
   ```

## Docker Issues

### Issue: Docker Build Fails

**Symptoms:**
- Build errors during `docker-compose build`
- "No space left on device" error
- Dependency download failures

**Solutions:**

1. **Clean Docker cache:**
   ```bash
   docker system prune -a
   docker volume prune
   ```

2. **Build with no cache:**
   ```bash
   docker-compose build --no-cache
   ```

3. **Check disk space:**
   ```bash
   df -h
   
   # Clean up if needed
   docker system df
   docker system prune -a --volumes
   ```

4. **Check Docker daemon:**
   ```bash
   docker info
   systemctl status docker
   ```

### Issue: Container Crashes

**Symptoms:**
- Container exits immediately
- "Exited (1)" status
- Container restarts continuously

**Solutions:**

1. **Check container logs:**
   ```bash
   docker-compose logs backend
   docker-compose logs frontend
   
   # Follow logs
   docker-compose logs -f backend
   ```

2. **Check container status:**
   ```bash
   docker-compose ps
   docker inspect <container-id>
   ```

3. **Run container interactively:**
   ```bash
   docker-compose run backend sh
   
   # Debug inside container
   ```

4. **Check resource limits:**
   ```bash
   # In docker-compose.yml
   services:
     backend:
       deploy:
         resources:
           limits:
             memory: 1G
             cpus: '1'
   ```

### Issue: Containers Can't Communicate

**Symptoms:**
- Frontend can't reach backend
- "Connection refused" errors
- Network errors

**Solutions:**

1. **Check network:**
   ```bash
   docker network ls
   docker network inspect <network-name>
   ```

2. **Verify service names:**
   ```yaml
   # In docker-compose.yml
   services:
     backend:
       # Frontend should use 'backend' as hostname
     frontend:
       environment:
         - VITE_API_URL=http://backend:8080
   ```

3. **Recreate network:**
   ```bash
   docker-compose down
   docker network prune
   docker-compose up
   ```

4. **Check port mappings:**
   ```bash
   docker-compose ps
   # Verify ports are correctly mapped
   ```

## Test Failures

### Issue: Backend Tests Fail

**Symptoms:**
- Test failures during `go test`
- Race condition errors
- Timeout errors

**Solutions:**

1. **Run with verbose output:**
   ```bash
   go test ./... -v
   ```

2. **Run specific test:**
   ```bash
   go test ./path/to/package -run TestName -v
   ```

3. **Check for race conditions:**
   ```bash
   go test ./... -race
   ```

4. **Increase test timeout:**
   ```bash
   go test ./... -timeout 30s
   ```

5. **Clean test cache:**
   ```bash
   go clean -testcache
   go test ./... -v
   ```

### Issue: Frontend Tests Fail

**Symptoms:**
- Test failures during `npm test`
- Timeout errors
- DOM errors

**Solutions:**

1. **Run with verbose output:**
   ```bash
   npm test -- --reporter=verbose
   ```

2. **Run specific test:**
   ```bash
   npm test -- MessageInput.test.ts
   ```

3. **Clear test cache:**
   ```bash
   npm test -- --clearCache
   ```

4. **Check for async issues:**
   ```typescript
   // Use await for async operations
   await wrapper.vm.$nextTick()
   ```

5. **Increase timeout:**
   ```typescript
   test('my test', async () => {
     // ...
   }, 10000) // 10 second timeout
   ```

### Issue: Property Tests Fail

**Symptoms:**
- Random test failures
- Counterexample found
- Shrinking errors

**Solutions:**

1. **Review counterexample:**
   ```
   Property test output shows failing input
   Verify if it's a valid edge case
   ```

2. **Fix the code or property:**
   ```typescript
   // Either fix the implementation
   // Or adjust the property/generator
   ```

3. **Increase iterations:**
   ```typescript
   fc.assert(
     fc.property(arb, (value) => {
       // test
     }),
     { numRuns: 1000 } // More iterations
   )
   ```

4. **Replay specific seed:**
   ```typescript
   fc.assert(
     fc.property(arb, (value) => {
       // test
     }),
     { seed: 123456 } // Replay failing seed
   )
   ```

## Performance Issues

### Issue: High Memory Usage

**Symptoms:**
- Application uses >1GB memory
- Out of memory errors
- Slow performance

**Solutions:**

1. **Profile memory usage:**
   ```bash
   # Backend
   go test ./... -memprofile=mem.prof
   go tool pprof mem.prof
   
   # Frontend
   # Use browser DevTools Memory profiler
   ```

2. **Check for memory leaks:**
   ```bash
   # Look for growing memory over time
   watch -n 1 'ps aux | grep server'
   ```

3. **Limit message history:**
   ```typescript
   // In frontend
   const MAX_MESSAGES = 500
   if (messages.length > MAX_MESSAGES) {
     messages.splice(0, messages.length - MAX_MESSAGES)
   }
   ```

4. **Clear old sessions:**
   ```go
   // In backend
   // Implement session cleanup
   // Remove sessions older than 30 minutes
   ```

### Issue: High CPU Usage

**Symptoms:**
- CPU usage >80%
- Slow response times
- System lag

**Solutions:**

1. **Profile CPU usage:**
   ```bash
   go test ./... -cpuprofile=cpu.prof
   go tool pprof cpu.prof
   (pprof) top10
   ```

2. **Check for infinite loops:**
   ```bash
   # Look for hot spots in profiler
   (pprof) list <function_name>
   ```

3. **Optimize hot paths:**
   ```
   Reduce allocations
   Use buffering
   Cache results
   ```

4. **Limit concurrent requests:**
   ```go
   // Add rate limiting
   limiter := rate.NewLimiter(10, 100)
   ```

## Common Error Messages

### "Address already in use"

**Cause:** Port is already occupied

**Solution:**
```bash
lsof -i :8080
kill -9 <PID>
```

### "Cannot find module"

**Cause:** Missing dependency

**Solution:**
```bash
npm install
# or
go mod download
```

### "CORS policy blocked"

**Cause:** CORS not configured

**Solution:**
```go
// In backend handler
w.Header().Set("Access-Control-Allow-Origin", "*")
```

### "Session not found"

**Cause:** Invalid or expired session

**Solution:**
```bash
# Create new session
curl -X POST http://localhost:8080/api/sessions
```

### "WebSocket connection failed"

**Cause:** Backend not running or wrong URL

**Solution:**
```bash
# Check backend
curl http://localhost:8080/health

# Use correct WebSocket URL
ws://localhost:8080/api/chat/stream
```

## Debug Mode

### Enable Debug Logging

**Backend:**
```bash
export LOG_LEVEL=debug
go run cmd/server/main.go
```

**Frontend:**
```javascript
// In browser console
localStorage.setItem('debug', '*')
```

### View Detailed Logs

**Backend:**
```bash
# Tail logs
tail -f logs/server.log

# Search logs
grep -i error logs/server.log
```

**Frontend:**
```
Open browser DevTools (F12)
Check Console tab
Enable verbose logging
```

### Configuration Endpoint

**Development only:**
```bash
curl http://localhost:8080/api/config
```

## Getting Help

If you're still experiencing issues:

1. **Check documentation:**
   - [README](README.md)
   - [Configuration Guide](backend/docs/CONFIGURATION.md)
   - [API Documentation](backend/docs/API.md)

2. **Search existing issues:**
   - GitHub Issues
   - Stack Overflow

3. **Gather information:**
   - Error messages
   - Logs
   - Steps to reproduce
   - Environment details

4. **Create a GitHub issue:**
   - Use issue template
   - Provide detailed information
   - Include logs and screenshots

5. **Contact support:**
   - Email: support@example.com
   - Slack: #bedrock-chat-ui

## Useful Commands

```bash
# Health checks
curl http://localhost:8080/health
curl http://localhost:5173

# View logs
docker-compose logs -f
tail -f logs/server.log

# Restart services
docker-compose restart
docker-compose restart backend

# Clean everything
docker-compose down -v
rm -rf node_modules
go clean -cache

# Run tests
go test ./... -v
npm test

# Check versions
go version
node --version
docker --version

# Check ports
lsof -i :8080
lsof -i :5173

# Check processes
ps aux | grep server
ps aux | grep node
```
