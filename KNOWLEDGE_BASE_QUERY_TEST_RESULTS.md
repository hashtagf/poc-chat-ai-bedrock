# Knowledge Base Query Test Results

## Task 14: Test Knowledge Base Query Functionality

### Date: December 9, 2025
### Knowledge Base ID: BKGKACVNCY
### Region: ap-southeast-1

## Summary

Testing of the Knowledge Base query functionality revealed that **embedding models in the ap-southeast-1 region have limited availability**:

### Available Embedding Models in ap-southeast-1:
- ✅ cohere.embed-v4:0 (NOT supported for Knowledge Bases)
- ✅ cohere.embed-english-v3 (Requires AWS Marketplace subscription)
- ✅ cohere.embed-multilingual-v3 (Requires AWS Marketplace subscription)
- ❌ amazon.titan-embed-text-v1 (NOT available in this region)
- ❌ amazon.titan-embed-text-v2 (NOT available in this region)

## Issue Encountered

When attempting to use Cohere embedding models, the following error occurred:

```
An error occurred (ValidationException) when calling the StartIngestionJob operation: 
Knowledge base role is not able to call specified bedrock embedding model: 
Model access is denied due to IAM user or service role is not authorized to perform 
the required AWS Marketplace actions (aws-marketplace:ViewSubscriptions, 
aws-marketplace:Subscribe) to enable access to this model.
```

## Root Cause

1. **Titan embedding models are not available** in the ap-southeast-1 region
2. **Cohere embedding models require AWS Marketplace subscription** which is not currently enabled
3. The Knowledge Base cannot ingest documents without a valid embedding model

## Required Actions

To complete this task, one of the following actions is required:

### Option 1: Enable Cohere Model Access (Recommended)
1. Navigate to AWS Bedrock Console → Model access
2. Request access to Cohere embedding models
3. Accept AWS Marketplace terms
4. Wait for access to be granted (usually immediate)
5. Re-run ingestion job
6. Test query functionality

### Option 2: Use a Different Region
Deploy the Knowledge Base in a region where Titan embedding models are available:
- us-east-1
- us-west-2
- eu-west-1
- etc.

### Option 3: Use Cohere with Manual Subscription
Subscribe to Cohere models through AWS Marketplace manually and then proceed with testing.

## Testing Plan (Once Model Access is Enabled)

Once model access is resolved, the following tests should be performed:

### Test 1: Basic Query
```bash
aws bedrock-agent-runtime retrieve \
  --knowledge-base-id BKGKACVNCY \
  --retrieval-query '{"text":"What is Amazon Bedrock?"}' \
  --region ap-southeast-1
```

**Expected**: Returns relevant results with citations

### Test 2: Multiple Queries
Test with various query types:
- Factual questions
- Conceptual questions
- Questions about specific features
- Edge cases (empty query, very long query)

### Test 3: Latency Verification
Measure query response time:
```bash
time aws bedrock-agent-runtime retrieve \
  --knowledge-base-id BKGKACVNCY \
  --retrieval-query '{"text":"test query"}' \
  --region ap-southeast-1
```

**Expected**: Sub-second response time

### Test 4: Citation Verification
Verify that results include:
- Source document references
- Confidence scores
- Relevant text excerpts

## Current Status

- ✅ Knowledge Base created successfully
- ✅ S3 Vectors storage configured
- ✅ IAM permissions configured
- ✅ Test document uploaded
- ❌ Ingestion blocked (model access required)
- ❌ Query testing blocked (no vectors to query)

## Recommendations

1. **Immediate**: Enable Cohere model access in AWS Bedrock Console
2. **Short-term**: Complete ingestion and query testing
3. **Long-term**: Consider deploying to a region with Titan model support for cost optimization

## Configuration Details

### Current Knowledge Base Configuration:
- **Embedding Model**: amazon.titan-embed-text-v1 (configured but not available)
- **Vector Dimensions**: 1536
- **Distance Metric**: cosine
- **Storage Type**: S3_VECTORS
- **Documents Bucket**: bedrock-chat-poc-kb-docs-dev
- **Vectors Bucket**: bedrock-chat-poc-kb-vectors-dev

### Required Configuration Change:
- **Embedding Model**: cohere.embed-english-v3 (requires access)
- **Vector Dimensions**: 1024
- **Distance Metric**: cosine (unchanged)

## Next Steps

1. User to enable Cohere model access in AWS Console
2. Update Terraform configuration to use cohere.embed-english-v3
3. Apply Terraform changes
4. Run ingestion job
5. Test query functionality
6. Complete task 14 validation

## References

- [AWS Bedrock Model Access Documentation](https://docs.aws.amazon.com/bedrock/latest/userguide/model-access.html)
- [Cohere Embedding Models](https://docs.cohere.com/docs/embeddings)
- [S3 Vectors Documentation](https://docs.aws.amazon.com/bedrock/latest/userguide/knowledge-base-setup.html)
