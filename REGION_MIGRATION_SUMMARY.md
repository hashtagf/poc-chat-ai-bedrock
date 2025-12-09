# Region Migration Summary - ap-southeast-1 to us-east-1

**Date**: December 10, 2025  
**Migration Reason**: Titan embedding model availability  
**Status**: ✅ Infrastructure Complete | ⚠️ Ingestion Blocked by S3 Vectors Limitation

## Migration Overview

Successfully migrated the Bedrock Knowledge Base infrastructure from ap-southeast-1 to us-east-1 to enable use of Amazon Titan embedding models.

## Changes Made

### 1. Region Configuration
- **Old Region**: ap-southeast-1 (Singapore)
- **New Region**: us-east-1 (N. Virginia)
- **Reason**: Titan embedding models not available in ap-southeast-1

### 2. Updated Files
- `terraform/environments/dev/terraform.tfvars`: Changed aws_region to us-east-1
- `.env`: Updated AWS_REGION to us-east-1
- `terraform/environments/dev/backend.tf`: Updated backend region to ap-southeast-1 (state bucket location)

### 3. Infrastructure Fixes
- Added region suffix to S3 bucket names to avoid conflicts
- Added 10-second delay before agent preparation
- Added `data.aws_region.current` to bedrock-agent module
- Added `s3vectors:PutVectors` permission to IAM policy
- Fixed agent preparation command to include region parameter

### 4. New Resource IDs (us-east-1)
```
Knowledge Base ID: S2MAASTZYB
Data Source ID: AK37QCGLKJ
Agent ID: S59QKPDIVN
Agent Alias ID: ZW2BJU1XDI
Documents Bucket: bedrock-chat-poc-kb-docs-dev-us-east-1
Vectors Bucket: bedrock-chat-poc-kb-vectors-dev
```

## Infrastructure Status

### ✅ Successfully Deployed

| Resource | Status | Details |
|----------|--------|---------|
| S3 Documents Bucket | ✅ Created | bedrock-chat-poc-kb-docs-dev-us-east-1 |
| S3 Vectors Bucket | ✅ Created | bedrock-chat-poc-kb-vectors-dev |
| Vector Index | ✅ Created | 1536 dimensions, cosine metric |
| Knowledge Base | ✅ ACTIVE | S3_VECTORS storage type |
| Data Source | ✅ Created | Connected to documents bucket |
| Bedrock Agent | ✅ PREPARED | With alias |
| IAM Roles & Policies | ✅ Created | All permissions configured |

### Verification Commands

```bash
# Knowledge Base status
aws bedrock-agent get-knowledge-base \
  --knowledge-base-id S2MAASTZYB \
  --region us-east-1

# Embedding model verification
aws bedrock list-foundation-models \
  --region us-east-1 \
  --output json | jq -r '.modelSummaries[] | select(.modelId | contains("titan-embed"))'

# S3 Vectors bucket
aws s3vectors list-vector-buckets --region us-east-1

# Vector index
aws s3vectors list-indexes \
  --vector-bucket-name bedrock-chat-poc-kb-vectors-dev \
  --region us-east-1
```

## Critical Issue Discovered: S3 Vectors Metadata Limitation

### Problem

Document ingestion consistently fails with the following error:

```
Invalid record for key '<uuid>': Filterable metadata must have at most 2048 bytes
```

### Root Cause

S3 Vectors has a **2048-byte limit on filterable metadata** per vector. When Bedrock ingests documents, it automatically adds metadata including:
- Full S3 URI (e.g., `s3://bedrock-chat-poc-kb-docs-dev-us-east-1/path/to/document.txt`)
- Document metadata
- Chunk information
- Other system metadata

With long bucket names and file paths, this metadata easily exceeds the 2048-byte limit.

### Evidence

Multiple ingestion attempts with different document sizes all failed:
- `test-document.txt` (1,491 bytes) - FAILED
- `test-document-simple.txt` (800 bytes) - FAILED  
- `simple.txt` (200 bytes) - FAILED
- `docs/test.txt` (30 bytes) - FAILED

All failures show the same metadata size error, indicating the issue is with Bedrock's automatic metadata generation, not document content.

### Impact

- ❌ Cannot ingest documents into Knowledge Base
- ❌ Cannot create vector embeddings
- ❌ Cannot test query functionality
- ❌ Cannot complete end-to-end RAG workflow

### Potential Solutions

#### Option 1: Use Shorter Bucket Names (Recommended to Try)
- Shorten project name in terraform.tfvars
- Use abbreviated environment names
- Remove region suffix from bucket name
- **Estimated savings**: ~50-100 bytes in metadata

#### Option 2: Switch to OpenSearch Serverless
- Change storage_configuration type to "OPENSEARCH_SERVERLESS"
- No metadata size limitations
- **Cost**: ~$700/month vs $5-10/month for S3 Vectors
- **Tradeoff**: 70x cost increase

#### Option 3: Use Pinecone or Other Vector DB
- Change storage type to supported third-party vector database
- Requires additional infrastructure setup
- Additional costs for vector database service

#### Option 4: Wait for AWS Fix
- This appears to be a limitation/bug in S3 Vectors integration
- May require AWS to increase metadata limit or optimize metadata generation
- No timeline available

### Recommendation

**Immediate**: Try Option 1 (shorter bucket names) as it's the quickest test.

**If Option 1 fails**: This is a fundamental limitation of S3 Vectors with Bedrock Knowledge Bases that makes it **unusable for production** despite the cost savings. Would need to switch to OpenSearch Serverless or another vector database.

## Task 16 Checkpoint Status

### Infrastructure Deployment: ✅ COMPLETE

All Terraform resources successfully deployed in us-east-1:
- S3 Vectors storage configured correctly
- Knowledge Base ACTIVE with Titan embedding model
- All IAM permissions properly configured
- Application configuration updated

### Functional Testing: ❌ BLOCKED

- **Ingestion**: Blocked by S3 Vectors metadata limitation
- **Queries**: Cannot test (no vectors to query)
- **End-to-end RAG**: Cannot test (ingestion blocked)

### S3 Vectors Migration: ✅ TECHNICALLY COMPLETE

The migration from AWSCC provider to AWS provider with S3 Vectors storage is complete. However, we've discovered a critical limitation that prevents actual use of S3 Vectors with Bedrock Knowledge Bases.

## Lessons Learned

1. **S3 Vectors Metadata Limitation**: The 2048-byte metadata limit is a significant constraint that wasn't documented in AWS documentation or the original requirements.

2. **Regional Model Availability**: Always verify model availability in target region before deployment.

3. **Bucket Naming**: Shorter bucket names are preferable to avoid metadata size issues.

4. **Testing Early**: This limitation would have been discovered earlier with actual ingestion testing during development.

## Next Steps - User Decision Required

Given the S3 Vectors metadata limitation, you need to choose:

1. **Try shorter bucket names** (quick test)
   - Update project_name to something short (e.g., "kb-poc")
   - Redeploy and test ingestion

2. **Switch to OpenSearch Serverless** (proven solution)
   - Update storage configuration
   - Accept ~$690/month additional cost
   - Complete functional testing

3. **Investigate alternative vector databases** (Pinecone, etc.)
   - Research options
   - Additional setup required
   - Variable costs

4. **Document and defer** (wait for AWS fix)
   - Infrastructure is ready
   - Document known limitation
   - Monitor AWS for updates

## Files Updated

- `terraform/environments/dev/terraform.tfvars`
- `terraform/environments/dev/backend.tf`
- `terraform/modules/knowledge-base/main.tf`
- `terraform/modules/bedrock-agent/main.tf`
- `.env`

## Related Documentation

- Initial validation: `CHECKPOINT_VALIDATION.md`
- Deployment details: `terraform/environments/dev/DEPLOYMENT_VALIDATION.md`
- Ingestion attempts: `INGESTION_TEST_RESULTS.md`
- Task list: `.kiro/specs/s3-vectors-fix/tasks.md`

---

**Migration Date**: December 10, 2025  
**Migrated By**: Kiro AI Agent  
**Infrastructure Status**: ✅ DEPLOYED  
**Functional Status**: ⚠️ BLOCKED BY S3 VECTORS LIMITATION
