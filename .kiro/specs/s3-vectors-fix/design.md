# Design Document: S3 Vectors Fix for Terraform Knowledge Base Module

## Overview

This design document outlines the fixes needed for the Terraform knowledge base module to properly configure S3 Vectors storage for Amazon Bedrock Knowledge Bases. The current implementation uses the AWSCC (AWS Cloud Control) provider which has limitations and compatibility issues. This design migrates to using the official AWS provider (version 6.25.0+) which now has native support for S3 Vectors.

### Current State Analysis

The existing `terraform/modules/knowledge-base/main.tf` currently uses:
- AWSCC provider for `awscc_bedrock_knowledge_base` resource
- AWS provider for S3 buckets and IAM roles
- Custom `aws_s3vectors_vector_bucket` and `aws_s3vectors_index` resources

### Problems with Current Implementation

1. **Provider Inconsistency**: Mixing AWSCC and AWS providers creates complexity
2. **Limited Support**: AWSCC provider has less community support and documentation
3. **Resource Compatibility**: Some AWS provider features don't work well with AWSCC resources
4. **Update Lag**: AWSCC provider may lag behind AWS provider in feature support

### Proposed Solution

Migrate entirely to the AWS provider (6.25.0+) which now supports:
- `aws_bedrockagent_knowledge_base` with S3 Vectors storage configuration
- Native S3 bucket resources with proper vector storage configuration
- Consistent resource management and state handling

### Requirements Coverage

This design addresses all requirements from the requirements document:

- **Requirement 1** (Correct Provider): Use AWS provider 6.25.0+ for all resources
- **Requirement 2** (Vector Bucket Config): Proper S3 bucket with encryption and versioning
- **Requirement 3** (Vector Index Config): Correct dimensions, distance metric, and data type
- **Requirement 4** (KB Storage Config): S3_VECTORS storage type with proper ARN references
- **Requirement 5** (IAM Permissions): Complete S3 and S3 Vectors permissions
- **Requirement 6** (Documentation): Clear docs on setup and cost savings
- **Requirement 7** (Dependencies): Proper resource ordering with depends_on
- **Requirement 8** (Ingestion): Script for document ingestion workflow
- **Requirement 9** (Compatibility): Maintain variable and output names

## Architecture

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                  Terraform AWS Provider 6.25.0+              │
│                                                               │
│  ┌────────────────────────────────────────────────────────┐ │
│  │  Knowledge Base Module                                  │ │
│  │                                                          │ │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐ │ │
│  │  │ S3 Documents │  │ S3 Vectors   │  │ Vector Index │ │ │
│  │  │ Bucket       │  │ Bucket       │  │              │ │ │
│  │  └──────────────┘  └──────────────┘  └──────────────┘ │ │
│  │         │                  │                  │         │ │
│  │         └──────────────────┴──────────────────┘         │ │
│  │                            │                             │ │
│  │                  ┌─────────▼─────────┐                  │ │
│  │                  │ Knowledge Base    │                  │ │
│  │                  │ (S3_VECTORS)      │                  │ │
│  │                  └───────────────────┘                  │ │
│  └────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

### Resource Dependencies

```
IAM Role → Knowledge Base
S3 Documents Bucket → Data Source → Knowledge Base
S3 Vectors Bucket → Vector Index → Knowledge Base
```


## Components and Interfaces

### 1. Terraform Provider Configuration

**Purpose**: Configure the AWS provider with the correct version for S3 Vectors support.

**Configuration**:
```hcl
terraform {
  required_version = ">= 1.5.0"
  
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 6.25.0"
    }
  }
}
```

**Key Changes**:
- Remove AWSCC provider dependency
- Pin AWS provider to 6.25.0 or higher
- Simplify provider configuration

### 2. S3 Vector Bucket

**Purpose**: Create an S3 bucket specifically for storing vector embeddings.

**Resources**:
```hcl
resource "aws_s3_bucket" "vectors" {
  bucket = "${var.project_name}-kb-vectors-${var.environment}"
  
  tags = merge(var.tags, {
    Name = "${var.project_name}-kb-vectors-${var.environment}"
    Purpose = "Vector Storage"
  })
}

resource "aws_s3_bucket_versioning" "vectors" {
  bucket = aws_s3_bucket.vectors.id
  
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "vectors" {
  bucket = aws_s3_bucket.vectors.id
  
  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

resource "aws_s3_bucket_public_access_block" "vectors" {
  bucket = aws_s3_bucket.vectors.id
  
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}
```

**Key Features**:
- Standard S3 bucket (not a special vector bucket type)
- Versioning enabled for data protection
- AES-256 encryption at rest
- Public access completely blocked
- Consistent tagging with other resources


### 3. Knowledge Base with S3 Vectors Storage

**Purpose**: Create a Bedrock Knowledge Base configured to use S3 Vectors for vector storage.

**Resource**:
```hcl
resource "aws_bedrockagent_knowledge_base" "main" {
  name        = "${var.project_name}-kb-${var.environment}"
  description = "Knowledge base for ${var.project_name} (${var.environment} environment)"
  role_arn    = var.kb_role_arn
  
  knowledge_base_configuration {
    type = "VECTOR"
    
    vector_knowledge_base_configuration {
      embedding_model_arn = "arn:aws:bedrock:${data.aws_region.current.name}::foundation-model/${var.embedding_model}"
    }
  }
  
  storage_configuration {
    type = "S3_VECTORS"
    
    s3_vectors_configuration {
      vector_bucket_arn = aws_s3_bucket.vectors.arn
      index_arn         = aws_bedrockagent_vector_index.main.arn
    }
  }
  
  tags = var.tags
  
  depends_on = [
    aws_bedrockagent_vector_index.main
  ]
}
```

**Key Changes from Current Implementation**:
- Use `aws_bedrockagent_knowledge_base` instead of `awscc_bedrock_knowledge_base`
- Storage configuration uses `s3_vectors_configuration` block
- Explicit dependency on vector index
- Simplified configuration structure

### 4. Vector Index

**Purpose**: Create a vector index for semantic search within the S3 vector bucket.

**Resource**:
```hcl
resource "aws_bedrockagent_vector_index" "main" {
  index_name         = "${var.project_name}-kb-index-${var.environment}"
  vector_bucket_name = aws_s3_bucket.vectors.id
  
  # Titan Embeddings G1 - Text v1.2 uses 1536 dimensions
  dimension       = 1536
  data_type       = "float32"
  distance_metric = "cosine"
  
  tags = var.tags
  
  depends_on = [
    aws_s3_bucket.vectors
  ]
}
```

**Configuration Details**:
- **dimension**: 1536 for Titan Embeddings G1 - Text v1.2
- **data_type**: float32 for standard floating-point vectors
- **distance_metric**: cosine for semantic similarity
- **vector_bucket_name**: References the S3 bucket ID (not ARN)

**Embedding Model Dimensions**:
- Titan Embeddings G1 - Text v1.2: 1536 dimensions
- Titan Embeddings G1 - Text v1: 1536 dimensions
- Cohere Embed English: 1024 dimensions
- Cohere Embed Multilingual: 1024 dimensions


### 5. Data Source Configuration

**Purpose**: Connect the S3 documents bucket as a data source for the Knowledge Base.

**Resource**:
```hcl
resource "aws_bedrockagent_data_source" "s3" {
  knowledge_base_id = aws_bedrockagent_knowledge_base.main.id
  name              = "${var.project_name}-kb-${var.environment}-s3-data-source"
  description       = "S3 data source for ${var.project_name} knowledge base"
  
  data_source_configuration {
    type = "S3"
    
    s3_configuration {
      bucket_arn = aws_s3_bucket.documents.arn
    }
  }
  
  depends_on = [
    aws_bedrockagent_knowledge_base.main
  ]
}
```

**Key Changes**:
- Use `aws_bedrockagent_data_source` instead of `awscc_bedrock_data_source`
- Simplified configuration structure
- Explicit dependency on Knowledge Base

### 6. IAM Permissions for S3 Vectors

**Purpose**: Grant the Knowledge Base IAM role permissions to access S3 buckets and perform vector operations.

**Updated IAM Policy**:
```hcl
resource "aws_iam_role_policy" "knowledge_base" {
  name = "${var.project_name}-kb-policy-${var.environment}"
  role = aws_iam_role.knowledge_base.id
  
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "BedrockInvokeModel"
        Effect = "Allow"
        Action = [
          "bedrock:InvokeModel"
        ]
        Resource = "arn:aws:bedrock:${data.aws_region.current.name}::foundation-model/${var.embedding_model}"
      },
      {
        Sid    = "S3DocumentsAccess"
        Effect = "Allow"
        Action = [
          "s3:GetObject",
          "s3:ListBucket"
        ]
        Resource = [
          aws_s3_bucket.documents.arn,
          "${aws_s3_bucket.documents.arn}/*"
        ]
      },
      {
        Sid    = "S3VectorsAccess"
        Effect = "Allow"
        Action = [
          "s3:GetObject",
          "s3:PutObject",
          "s3:DeleteObject",
          "s3:ListBucket"
        ]
        Resource = [
          aws_s3_bucket.vectors.arn,
          "${aws_s3_bucket.vectors.arn}/*"
        ]
      },
      {
        Sid    = "S3VectorsIndexAccess"
        Effect = "Allow"
        Action = [
          "bedrock:Query",
          "bedrock:PutVector",
          "bedrock:DeleteVector",
          "bedrock:GetVector"
        ]
        Resource = aws_bedrockagent_vector_index.main.arn
      }
    ]
  })
}
```

**Key Permissions**:
- **S3 Documents**: Read-only access (GetObject, ListBucket)
- **S3 Vectors**: Full access (GetObject, PutObject, DeleteObject, ListBucket)
- **Vector Index**: Query and manage vectors (Query, PutVector, GetVector, DeleteVector)
- **Embedding Model**: Invoke model for generating embeddings

**Note**: The S3 Vectors API actions are prefixed with `bedrock:` not `s3vectors:` when used through Bedrock Knowledge Bases.


## Data Models

### Module Variables

```hcl
variable "project_name" {
  description = "Project name used for resource naming"
  type        = string
}

variable "environment" {
  description = "Environment name (dev, staging, prod)"
  type        = string
  
  validation {
    condition     = contains(["dev", "staging", "prod"], var.environment)
    error_message = "Environment must be dev, staging, or prod"
  }
}

variable "embedding_model" {
  description = "Embedding model ID (e.g., amazon.titan-embed-text-v1)"
  type        = string
  default     = "amazon.titan-embed-text-v1"
}

variable "kb_role_arn" {
  description = "ARN of the IAM role for Knowledge Base"
  type        = string
}

variable "tags" {
  description = "Tags to apply to all resources"
  type        = map(string)
  default     = {}
}
```

### Module Outputs

```hcl
output "knowledge_base_id" {
  description = "The ID of the Knowledge Base"
  value       = aws_bedrockagent_knowledge_base.main.id
}

output "knowledge_base_arn" {
  description = "The ARN of the Knowledge Base"
  value       = aws_bedrockagent_knowledge_base.main.arn
}

output "data_source_id" {
  description = "The ID of the S3 Data Source"
  value       = aws_bedrockagent_data_source.s3.id
}

output "documents_bucket_name" {
  description = "Name of the S3 bucket for documents"
  value       = aws_s3_bucket.documents.id
}

output "documents_bucket_arn" {
  description = "ARN of the S3 bucket for documents"
  value       = aws_s3_bucket.documents.arn
}

output "vectors_bucket_name" {
  description = "Name of the S3 Vectors bucket"
  value       = aws_s3_bucket.vectors.id
}

output "vectors_bucket_arn" {
  description = "ARN of the S3 Vectors bucket"
  value       = aws_s3_bucket.vectors.arn
}

output "index_name" {
  description = "Name of the vector index"
  value       = aws_bedrockagent_vector_index.main.index_name
}

output "index_arn" {
  description = "ARN of the vector index"
  value       = aws_bedrockagent_vector_index.main.arn
}

output "role_arn" {
  description = "ARN of the IAM role for Knowledge Base"
  value       = var.kb_role_arn
}
```

**Backward Compatibility**: All output names remain the same as the current implementation to ensure existing environment configurations don't break.


## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system—essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*

Since this is infrastructure-as-code, most acceptance criteria are validated through Terraform configuration validation and infrastructure testing rather than traditional property-based testing. However, we can define one key property:

### Property 1: Universal Resource Tagging

*For any* AWS resource created by this Terraform module, that resource should have all required tags from the `var.tags` input variable applied to it.

**Validates: Requirements 2.5**

**Rationale**: Consistent tagging across all resources is essential for cost tracking, resource management, and compliance. This property ensures that when tags are passed to the module, they are applied to every resource that supports tagging.

**Resources that must be tagged**:
- aws_s3_bucket.documents
- aws_s3_bucket.vectors
- aws_bedrockagent_knowledge_base.main
- aws_bedrockagent_vector_index.main
- aws_bedrockagent_data_source.s3

### Infrastructure Validation Examples

The following acceptance criteria are validated through Terraform configuration validation and infrastructure testing:

**Provider Configuration** (Requirements 1.1-1.4):
- Verify AWS provider version is 6.25.0 or higher
- Verify no AWSCC provider is used
- Verify aws_bedrockagent_knowledge_base resource is used
- Verify provider version is pinned

**Vector Bucket Configuration** (Requirements 2.1-2.4):
- Verify aws_s3_bucket resource exists with correct naming
- Verify encryption configuration block exists with AES256
- Verify versioning configuration block exists with status "Enabled"
- Verify public access block exists with all options set to true

**Vector Index Configuration** (Requirements 3.1-3.5):
- Verify dimension = 1536 for Titan Embeddings
- Verify distance_metric = "cosine"
- Verify data_type = "float32"
- Verify vector_bucket_name references the S3 bucket
- Verify index_arn output exists

**Knowledge Base Storage Configuration** (Requirements 4.1-4.4):
- Verify storage_configuration.type = "S3_VECTORS"
- Verify s3_vectors_configuration.vector_bucket_arn is set
- Verify s3_vectors_configuration.index_arn is set
- Verify terraform validate passes

**IAM Permissions** (Requirements 5.1-5.8):
- Verify s3:GetObject permission on vector bucket
- Verify s3:PutObject permission on vector bucket
- Verify s3:DeleteObject permission on vector bucket
- Verify s3:ListBucket permission on vector bucket
- Verify bedrock:Query permission on index ARN
- Verify bedrock:PutVector permission on index ARN
- Verify bedrock:GetVector permission on index ARN
- Verify bedrock:DeleteVector permission on index ARN

**Resource Dependencies** (Requirements 7.1-7.5):
- Verify vector index depends_on vector bucket
- Verify Knowledge Base depends_on vector index
- Verify Knowledge Base uses kb_role_arn variable
- Verify explicit depends_on blocks exist where needed
- Verify terraform destroy completes without errors

**Ingestion Script** (Requirements 8.1-8.5):
- Verify ingest-kb.sh script exists
- Verify script calls aws bedrock-agent start-ingestion-job
- Verify script monitors job status with polling
- Verify script displays completion statistics
- Verify script displays error messages on failure

**Backward Compatibility** (Requirements 9.1-9.2):
- Verify all variable names match existing module
- Verify all output names match existing module


## Error Handling

### Terraform Validation Errors

**Provider Version Mismatch**:
- Error: "Provider version does not meet requirements"
- Solution: Upgrade AWS provider to 6.25.0 or higher
- Command: `terraform init -upgrade`

**Resource Not Found**:
- Error: "aws_bedrockagent_knowledge_base resource not found"
- Solution: Ensure AWS provider is 6.25.0+
- Verify: `terraform version` shows correct provider version

**Invalid Storage Configuration**:
- Error: "storage_configuration.type must be one of: OPENSEARCH_SERVERLESS, PINECONE, RDS, S3_VECTORS"
- Solution: Verify type is exactly "S3_VECTORS" (case-sensitive)
- Check: Storage configuration block syntax

### Resource Creation Failures

**Vector Bucket Creation Fails**:
- Error: "BucketAlreadyExists" or "BucketAlreadyOwnedByYou"
- Solution: Choose a globally unique bucket name
- Fix: Update project_name or environment variables

**Vector Index Creation Fails**:
- Error: "InvalidParameterException: dimension must match embedding model"
- Solution: Verify dimension matches your embedding model
- Titan Embeddings G1: 1536 dimensions
- Cohere Embed: 1024 dimensions

**Knowledge Base Creation Fails**:
- Error: "AccessDeniedException: IAM role lacks permissions"
- Solution: Verify IAM role has all required permissions
- Check: IAM policy includes S3, Bedrock, and vector operations

**Data Source Sync Fails**:
- Error: "InvalidS3BucketException: Cannot access S3 bucket"
- Solution: Verify IAM role has s3:GetObject and s3:ListBucket permissions
- Check: Bucket policy doesn't deny access

### State Management Errors

**State Lock Conflicts**:
- Error: "Error acquiring state lock"
- Solution: Wait for other operations to complete
- Emergency: Use `terraform force-unlock` (with caution)

**Resource Already Exists**:
- Error: "Resource already exists in AWS but not in state"
- Solution: Import existing resource into state
- Command: `terraform import aws_bedrockagent_knowledge_base.main <kb-id>`

### Migration Errors

**Replacing Existing Resources**:
- Warning: "Plan shows resources will be replaced"
- Impact: Knowledge Base will be recreated (data loss possible)
- Solution: Export data before migration, re-import after
- Prevention: Use `terraform plan` before `terraform apply`

**AWSCC to AWS Provider Migration**:
- Issue: Resources managed by AWSCC provider need to be migrated
- Solution: Remove AWSCC resources from state, import as AWS resources
- Steps:
  1. `terraform state rm awscc_bedrock_knowledge_base.main`
  2. Update configuration to use aws_bedrockagent_knowledge_base
  3. `terraform import aws_bedrockagent_knowledge_base.main <kb-id>`


## Testing Strategy

### Terraform Configuration Validation

**Syntax and Configuration Tests**:
```bash
# Validate Terraform syntax
terraform validate

# Format check
terraform fmt -check -recursive

# Plan without applying
terraform plan -out=tfplan

# Show plan in JSON for automated testing
terraform show -json tfplan | jq
```

**Validation Checks**:
- All resources use AWS provider (not AWSCC)
- Provider version is 6.25.0 or higher
- All required variables are defined
- All outputs are defined
- Resource dependencies are correct

### Infrastructure Testing

**Manual Testing Workflow**:
```bash
# 1. Initialize Terraform
cd terraform/modules/knowledge-base
terraform init

# 2. Validate configuration
terraform validate

# 3. Plan deployment
terraform plan -var="project_name=test" -var="environment=dev" -var="kb_role_arn=arn:aws:iam::123456789012:role/test"

# 4. Apply to test account
terraform apply -auto-approve

# 5. Verify resources in AWS Console
# - Check S3 buckets exist
# - Check vector index exists
# - Check Knowledge Base exists with S3_VECTORS storage

# 6. Test ingestion
aws s3 cp test-document.pdf s3://$(terraform output -raw documents_bucket_name)/
./scripts/ingest-kb.sh

# 7. Test query
aws bedrock-agent-runtime retrieve \
  --knowledge-base-id $(terraform output -raw knowledge_base_id) \
  --retrieval-query text="test query" \
  --region ap-southeast-1

# 8. Cleanup
terraform destroy -auto-approve
```

**Automated Testing with Terratest** (Optional):
```go
func TestKnowledgeBaseModule(t *testing.T) {
    // This test validates the Knowledge Base module configuration
    // **Feature: s3-vectors-fix, Property 1: Universal Resource Tagging**
    
    terraformOptions := &terraform.Options{
        TerraformDir: "../modules/knowledge-base",
        Vars: map[string]interface{}{
            "project_name": "test-kb",
            "environment":  "dev",
            "kb_role_arn":  "arn:aws:iam::123456789012:role/test",
            "tags": map[string]string{
                "Environment": "test",
                "Project":     "test-kb",
            },
        },
    }
    
    defer terraform.Destroy(t, terraformOptions)
    terraform.InitAndApply(t, terraformOptions)
    
    // Verify outputs
    kbID := terraform.Output(t, terraformOptions, "knowledge_base_id")
    assert.NotEmpty(t, kbID)
    
    vectorsBucket := terraform.Output(t, terraformOptions, "vectors_bucket_name")
    assert.Contains(t, vectorsBucket, "test-kb")
    assert.Contains(t, vectorsBucket, "vectors")
    
    // Verify resources in AWS
    // ... AWS SDK calls to verify resource configuration
}
```

### Integration Testing

**End-to-End RAG Workflow Test**:
1. Deploy infrastructure with Terraform
2. Upload test documents to S3 documents bucket
3. Trigger ingestion job
4. Wait for ingestion to complete
5. Query Knowledge Base with test questions
6. Verify responses include citations
7. Verify vector embeddings are stored in S3 vectors bucket
8. Clean up resources

**Cost Validation**:
- Monitor AWS costs for S3 Vectors vs OpenSearch Serverless
- Expected: ~$5-10/month for S3 Vectors
- Compare: ~$700/month for OpenSearch Serverless minimum
- Savings: ~$690/month (99% cost reduction)


## Migration Strategy

### Pre-Migration Checklist

1. **Backup Current State**:
   ```bash
   cd terraform/environments/dev
   terraform state pull > backup-state.json
   ```

2. **Document Current Resources**:
   ```bash
   terraform show > current-resources.txt
   ```

3. **Export Knowledge Base Data** (if needed):
   - Note: Vector embeddings cannot be exported directly
   - Keep source documents in S3 for re-ingestion
   - Document any custom metadata or configurations

### Migration Steps

**Option 1: Clean Slate (Recommended for Dev)**:
```bash
# 1. Destroy existing resources
cd terraform/environments/dev
terraform destroy

# 2. Update module code with new AWS provider resources
# (Apply the changes from this design document)

# 3. Re-deploy with new configuration
terraform init -upgrade
terraform apply

# 4. Re-upload documents and trigger ingestion
aws s3 sync ./documents/ s3://$(terraform output -raw documents_bucket_name)/
cd ../../modules/knowledge-base/scripts
./ingest-kb.sh
```

**Option 2: In-Place Migration (For Production)**:
```bash
# 1. Remove AWSCC resources from state (don't destroy in AWS)
terraform state rm awscc_bedrock_knowledge_base.main
terraform state rm awscc_bedrock_data_source.s3

# 2. Update module code with new AWS provider resources

# 3. Import existing resources into new state
terraform import aws_bedrockagent_knowledge_base.main <knowledge-base-id>
terraform import aws_bedrockagent_data_source.s3 <data-source-id>

# 4. Verify plan shows no changes
terraform plan

# 5. If plan shows changes, adjust configuration to match existing resources
```

### Post-Migration Validation

1. **Verify Resources**:
   ```bash
   # Check Knowledge Base
   aws bedrock-agent get-knowledge-base \
     --knowledge-base-id $(terraform output -raw knowledge_base_id)
   
   # Check Data Source
   aws bedrock-agent get-data-source \
     --knowledge-base-id $(terraform output -raw knowledge_base_id) \
     --data-source-id $(terraform output -raw data_source_id)
   
   # Check S3 buckets
   aws s3 ls $(terraform output -raw vectors_bucket_name)
   ```

2. **Test Query**:
   ```bash
   aws bedrock-agent-runtime retrieve \
     --knowledge-base-id $(terraform output -raw knowledge_base_id) \
     --retrieval-query text="test query" \
     --region ap-southeast-1
   ```

3. **Verify Application Integration**:
   ```bash
   # Update .env file with new outputs
   echo "BEDROCK_KNOWLEDGE_BASE_ID=$(terraform output -raw knowledge_base_id)" >> ../../../.env
   
   # Test backend application
   cd ../../../backend
   go run cmd/server/main.go
   ```

### Rollback Plan

If migration fails:

1. **Restore State**:
   ```bash
   terraform state push backup-state.json
   ```

2. **Revert Code Changes**:
   ```bash
   git checkout HEAD -- terraform/modules/knowledge-base/
   ```

3. **Re-initialize**:
   ```bash
   terraform init
   terraform plan  # Verify state matches AWS
   ```


## Documentation Updates

### Module README Updates

The `terraform/modules/knowledge-base/README.md` should be updated to include:

1. **Provider Requirements**:
   - AWS Provider 6.25.0+ required for S3 Vectors support
   - Remove references to AWSCC provider

2. **Cost Comparison**:
   ```markdown
   ## Cost Optimization
   
   **S3 Vectors Storage:**
   - Storage: ~$0.023/GB/month
   - Queries: ~$0.0004 per 1000 queries
   - **Typical POC cost: $5-10/month**
   
   **vs OpenSearch Serverless:**
   - Minimum: ~$700/month (2 OCUs required)
   - **Savings: ~$690/month** 💰
   ```

3. **Configuration Examples**:
   - Show complete module usage with all variables
   - Include example outputs
   - Document embedding model dimensions

4. **Troubleshooting Section**:
   - Common errors and solutions
   - Migration guide from AWSCC provider
   - IAM permission issues

### Environment README Updates

Update `terraform/environments/dev/README.md` and similar files:

1. **Deployment Instructions**:
   - Add provider upgrade step: `terraform init -upgrade`
   - Document required AWS provider version

2. **Migration Notes**:
   - Link to migration strategy in design document
   - Warn about resource replacement

### Root Terraform README Updates

Update `terraform/README.md`:

1. **Prerequisites**:
   - Terraform >= 1.5.0
   - AWS Provider >= 6.25.0
   - AWS CLI configured

2. **Quick Start**:
   ```markdown
   ## Quick Start
   
   1. Initialize Terraform with provider upgrade:
      ```bash
      cd terraform/environments/dev
      terraform init -upgrade
      ```
   
   2. Review the plan:
      ```bash
      terraform plan
      ```
   
   3. Apply the configuration:
      ```bash
      terraform apply
      ```
   
   4. Upload documents and trigger ingestion:
      ```bash
      aws s3 cp documents/ s3://$(terraform output -raw documents_bucket_name)/ --recursive
      cd ../../modules/knowledge-base/scripts
      ./ingest-kb.sh
      ```
   ```

3. **S3 Vectors Benefits**:
   - Cost-effective vector storage
   - Sub-second query latency
   - Seamless Bedrock integration
   - No infrastructure management

## Summary

This design provides a complete migration path from the AWSCC provider to the AWS provider for S3 Vectors support in the Terraform knowledge base module. Key improvements include:

1. **Simplified Configuration**: Single provider (AWS) instead of mixed providers
2. **Better Support**: Official AWS provider with better documentation and community support
3. **Cost Savings**: S3 Vectors provides 99% cost reduction vs OpenSearch Serverless
4. **Backward Compatibility**: All variable and output names remain the same
5. **Clear Migration Path**: Both clean slate and in-place migration options
6. **Comprehensive Testing**: Validation, integration, and end-to-end testing strategies

The implementation will maintain all existing functionality while improving reliability and reducing complexity.
