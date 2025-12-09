# Terraform Plan Summary - S3 Vectors Fix

**Date**: December 9, 2025
**Environment**: dev
**Status**: Ready to apply

## Overview

This plan creates a complete Bedrock infrastructure with S3 Vectors storage for the knowledge base. All resources are new (no existing infrastructure to replace).

## Resources to be Created

### Total: 15 resources

#### Bedrock Agent (4 resources)
- `aws_bedrockagent_agent.this` - Main Bedrock agent
- `aws_bedrockagent_agent_alias.draft` - DRAFT alias for the agent
- `terraform_data.agent_preparation` - Agent preparation trigger
- `time_sleep.agent_initialization` - 30s initialization delay

#### IAM Roles & Policies (4 resources)
- `aws_iam_role.agent_role` - IAM role for Bedrock agent
- `aws_iam_role_policy.agent_policy` - Policy for agent (InvokeModel permissions)
- `aws_iam_role.knowledge_base` - IAM role for Knowledge Base
- `aws_iam_role_policy.knowledge_base` - Policy for KB (S3, S3 Vectors, Bedrock permissions)

#### S3 & S3 Vectors (5 resources)
- `aws_s3_bucket.documents` - S3 bucket for source documents
- `aws_s3_bucket_versioning.documents` - Versioning for documents bucket
- `aws_s3_bucket_server_side_encryption_configuration.documents` - AES256 encryption
- `aws_s3vectors_vector_bucket.vectors` - S3 Vectors bucket for embeddings
- `aws_s3vectors_index.main` - Vector index (1536 dimensions, cosine distance)

#### Knowledge Base (2 resources)
- `awscc_bedrock_knowledge_base.main` - Knowledge Base with S3_VECTORS storage
- `awscc_bedrock_data_source.s3` - S3 data source connector

## Key Configuration Details

### S3 Vectors Configuration
- **Vector Bucket**: `bedrock-chat-poc-kb-vectors-dev`
- **Index Name**: `bedrock-chat-poc-kb-index-dev`
- **Dimensions**: 1536 (for Titan Embeddings G1)
- **Distance Metric**: cosine
- **Data Type**: float32
- **Encryption**: AES256

### Knowledge Base Configuration
- **Name**: `bedrock-chat-poc-kb-dev`
- **Storage Type**: S3_VECTORS
- **Embedding Model**: amazon.titan-embed-text-v1
- **Documents Bucket**: `bedrock-chat-poc-kb-docs-dev`

### Bedrock Agent Configuration
- **Name**: `bedrock-chat-poc-agent-dev`
- **Foundation Model**: amazon.nova-2-lite-v1:0
- **Idle Session TTL**: 1800 seconds (30 minutes)
- **Region**: ap-southeast-1

## Resources Replaced vs Updated

**All resources are NEW** - This is a fresh deployment with no existing infrastructure.

- ✅ No resources will be replaced
- ✅ No resources will be updated
- ✅ 15 resources will be created

## Provider Versions

- **AWS Provider**: 6.25.0 (supports S3 Vectors resources)
- **AWSCC Provider**: 1.65.0 (required for Knowledge Base with S3_VECTORS)
- **Time Provider**: 0.13.1

## Cost Implications

### S3 Vectors Storage (Estimated)
- Storage: ~$0.023/GB/month
- Queries: ~$0.0004 per 1000 queries
- **Typical POC cost: $5-10/month**

### Comparison to OpenSearch Serverless
- OpenSearch minimum: ~$700/month (2 OCUs required)
- **Savings: ~$690/month (99% cost reduction)**

## Next Steps

1. ✅ Review this plan summary
2. ⏭️ Run `terraform apply "tfplan"` to create resources
3. ⏭️ Verify resources in AWS Console
4. ⏭️ Upload test documents to S3
5. ⏭️ Run ingestion script
6. ⏭️ Test Knowledge Base queries

## Validation Checklist

After applying, verify:
- [ ] S3 buckets exist with correct configuration
- [ ] Vector bucket has AES256 encryption
- [ ] Vector index exists with 1536 dimensions
- [ ] Knowledge Base exists with S3_VECTORS storage type
- [ ] IAM roles have correct permissions
- [ ] Bedrock agent is created and prepared
- [ ] All resources have proper tags

## Rollback Plan

If issues occur:
1. Run `terraform destroy` to remove all resources
2. Review error messages
3. Fix configuration issues
4. Re-run `terraform plan` and `terraform apply`

## Requirements Validated

This plan addresses:
- ✅ Requirement 1.1: AWS Provider 6.25.0+ for S3 Vectors
- ✅ Requirement 2.1-2.5: Vector bucket configuration
- ✅ Requirement 3.1-3.5: Vector index configuration
- ✅ Requirement 4.1-4.4: Knowledge Base S3_VECTORS storage
- ✅ Requirement 5.1-5.8: IAM permissions for S3 Vectors
- ✅ Requirement 7.1-7.5: Resource dependencies
- ✅ Requirement 9.1-9.2: Backward compatibility (output names)
