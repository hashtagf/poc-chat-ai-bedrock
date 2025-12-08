# Knowledge Base Module

This Terraform module provisions an Amazon Bedrock Knowledge Base with S3 data source and S3 Vectors for vector storage.

## Overview

The module creates:
- S3 bucket for knowledge base documents (data source)
- S3 bucket for vector storage with S3 Vectors
- S3 Vector Index for embeddings
- Bedrock Knowledge Base configured with embedding model
- S3 data source connection

## Vector Store Implementation

**Current Implementation**: This module uses **S3 Vectors** as the vector store.

**Benefits**:
- **Cost-effective**: ~$10/month for small document sets vs ~$700/month for OpenSearch Serverless
- **Simplified architecture**: No need for OpenSearch Serverless collections, policies, or network configuration
- **Fully managed**: AWS handles vector indexing and querying
- **Fast queries**: Sub-second cold query latency, ~100ms warm query latency

**Supported Embedding Models**:
- Amazon Titan Embed Text v1 (1536 dimensions)
- Amazon Titan Embed Text v2 (1024 dimensions)
- Cohere Embed models (1024 dimensions)

The vector dimensions are automatically configured based on the embedding model selected

## Usage

```hcl
module "knowledge_base" {
  source = "./modules/knowledge-base"

  knowledge_base_name     = "my-knowledge-base"
  embedding_model         = "amazon.titan-embed-text-v1"
  kb_role_arn            = module.iam.kb_role_arn
  s3_bucket_name         = "my-kb-documents-bucket"
  s3_vector_bucket_name  = "my-kb-vectors-bucket"

  tags = {
    Environment = "dev"
    Project     = "my-project"
  }
}
```

## Inputs

| Name | Description | Type | Required |
|------|-------------|------|----------|
| knowledge_base_name | Name of the Knowledge Base | string | yes |
| embedding_model | Embedding model ID (e.g., amazon.titan-embed-text-v1) | string | yes |
| kb_role_arn | ARN of the IAM role for Knowledge Base | string | yes |
| s3_bucket_name | Name of the S3 bucket for documents | string | yes |
| s3_vector_bucket_name | Name of the S3 bucket for vector storage | string | yes |
| tags | Resource tags | map(string) | no |

## Outputs

| Name | Description |
|------|-------------|
| knowledge_base_id | The Knowledge Base ID |
| knowledge_base_arn | The Knowledge Base ARN |
| s3_bucket_name | The S3 bucket name for document uploads |
| s3_bucket_arn | The S3 bucket ARN |
| s3_vector_bucket_name | The S3 bucket name for vector storage |
| s3_vector_bucket_arn | The S3 bucket ARN for vector storage |
| data_source_id | The data source ID |

## S3 Bucket Configuration

Both S3 buckets are configured with:
- Versioning enabled
- AES-256 encryption at rest
- Public access blocked
- Appropriate tags

## OpenSearch Serverless Configuration

The OpenSearch Serverless collection is configured with:
- Type: VECTORSEARCH
- Access policy: Grants KB role permissions for collection and index operations
- Network policy: Public access (can be restricted to VPC in production)
- Encryption policy: AWS-owned key

## Data Source Configuration

The S3 data source is automatically configured to:
- Connect to the documents bucket
- Sync documents for ingestion
- Process documents through the embedding model

## Document Upload and Sync

After deployment:

1. Upload documents to the S3 bucket:
```bash
aws s3 cp documents/ s3://$(terraform output -raw s3_bucket_name)/ --recursive
```

2. Start an ingestion job to sync documents:
```bash
aws bedrock-agent start-ingestion-job \
  --knowledge-base-id $(terraform output -raw knowledge_base_id) \
  --data-source-id $(terraform output -raw data_source_id)
```

3. Check ingestion job status:
```bash
aws bedrock-agent list-ingestion-jobs \
  --knowledge-base-id $(terraform output -raw knowledge_base_id) \
  --data-source-id $(terraform output -raw data_source_id)
```

## Requirements

- Terraform >= 1.5.0
- AWS Provider >= 5.0.0
- Time Provider >= 0.9.0

## Dependencies

This module requires the IAM module to be deployed first to create the Knowledge Base execution role.
