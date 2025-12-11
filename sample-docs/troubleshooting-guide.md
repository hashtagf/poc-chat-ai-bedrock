# Troubleshooting Guide

## Common Issues and Solutions

### Connection Issues

#### WebSocket Connection Failed
**Problem**: Cannot establish WebSocket connection to chat endpoint.

**Solutions**:
1. Check if backend server is running on correct port
2. Verify WebSocket URL format: `ws://localhost:8080/api/chat`
3. Check firewall settings
4. Ensure CORS is properly configured

#### Session Creation Failed
**Problem**: POST request to `/api/sessions` returns error.

**Solutions**:
1. Verify Content-Type header is `application/json`
2. Check request body format
3. Ensure backend service is healthy
4. Check server logs for detailed error information

### Bedrock Integration Issues

#### Agent Not Responding
**Problem**: Messages sent but no response from Bedrock Agent.

**Solutions**:
1. Verify `BEDROCK_AGENT_ID` environment variable
2. Check `BEDROCK_AGENT_ALIAS_ID` configuration
3. Ensure AWS credentials are properly configured
4. Verify agent is in "PREPARED" status
5. Check IAM permissions for Bedrock access

#### Knowledge Base Not Working
**Problem**: Agent responds but doesn't use knowledge base information.

**Solutions**:
1. Verify `BEDROCK_KNOWLEDGE_BASE_ID` is set
2. Check knowledge base association with agent
3. Ensure knowledge base has indexed documents
4. Verify data source sync is completed
5. Check knowledge base permissions

### Performance Issues

#### Slow Response Times
**Problem**: Agent takes too long to respond.

**Solutions**:
1. Check network connectivity to AWS
2. Verify AWS region configuration
3. Monitor Bedrock API quotas and limits
4. Check for rate limiting
5. Review timeout configurations

#### High Memory Usage
**Problem**: Application consuming excessive memory.

**Solutions**:
1. Check for memory leaks in session management
2. Review WebSocket connection handling
3. Monitor goroutine count
4. Check streaming response processing
5. Verify proper resource cleanup

### Configuration Issues

#### Environment Variables Not Loaded
**Problem**: Application not reading environment variables.

**Solutions**:
1. Verify `.env` file exists and is readable
2. Check environment variable names (case-sensitive)
3. Ensure proper shell environment setup
4. Use `source .env` before running application
5. Check for special characters in values

#### AWS Credentials Issues
**Problem**: Authentication errors with AWS services.

**Solutions**:
1. Verify AWS credentials are configured
2. Check IAM user permissions
3. Ensure correct AWS region
4. Verify credential file format
5. Test with AWS CLI: `aws sts get-caller-identity`

## Debugging Tips

### Enable Debug Logging
Set environment variable:
```bash
LOG_LEVEL=debug
```

### Check Service Health
```bash
curl http://localhost:8080/health
```

### Monitor WebSocket Connection
Use browser developer tools to monitor WebSocket messages and connection status.

### AWS CLI Testing
Test Bedrock access:
```bash
aws bedrock-agent get-agent --agent-id YOUR_AGENT_ID --region us-east-1
```