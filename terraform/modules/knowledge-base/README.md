# Knowledge Base Module

Terraform module for Amazon Bedrock Knowledge Base with S3 Vectors storage.

## Prerequisites

- Terraform >= 1.5.0
- **AWS Provider >= 6.25.0** (required for native S3 Vectors support)
- AWS CLI configured with appropriate credentials
- `jq` installed (for ingestion script)
- IAM permissions for Bedrock, S3, and IAM

> **Note:** This module uses the official AWS provider (not AWSCC provider) for all resources. AWS Provider 6.25.0+ includes native support for `aws_bedrockagent_knowledge_base`, `aws_bedrockagent_vector_index`, and `aws_bedrockagent_data_source` resources with S3 Vectors storage.

## Features

- ✅ **Full Terraform support** - No manual console steps required
- ✅ **S3 Vectors storage** - Cost-effective vector storage using native AWS provider resources
- ✅ **Native AWS Provider** - Uses `aws_bedrockagent_knowledge_base` with S3_VECTORS storage type
- ✅ **Vector Index** - Automatic `aws_bedrockagent_vector_index` creation with cosine similarity
- ✅ **Standard S3 Buckets** - Uses `aws_s3_bucket` with encryption, versioning, and public access blocking
- ✅ **Automatic IAM setup** - Proper service roles and policies for S3 and vector operations
- ✅ **Complete Knowledge Base** - KB + Data Source + Index created automatically with proper dependencies
- ✅ **Ingestion script** - Easy document ingestion workflow

## Usage

### Basic Setup

```hcl
module "knowledge_base" {
  source = "../../modules/knowledge-base"

  project_name    = "bedrock-chat-poc"
  environment     = "dev"
  embedding_model = "amazon.titan-embed-text-v1"  # 1536 dimensions
  kb_role_arn     = module.iam.knowledge_base_role_arn
  
  tags = {
    Environment = "dev"
    Project     = "bedrock-chat-poc"
    ManagedBy   = "Terraform"
  }
}
```

**Embedding Model Dimensions:**
- `amazon.titan-embed-text-v1`: 1536 dimensions (default)
- `amazon.titan-embed-text-v2`: 1536 dimensions
- `cohere.embed-english-v3`: 1024 dimensions
- `cohere.embed-multilingual-v3`: 1024 dimensions

> **Important:** The vector index dimension must match your embedding model. This module defaults to 1536 dimensions for Titan models.

### Outputs

```hcl
output "knowledge_base_id" {
  value = module.knowledge_base.knowledge_base_id
}

output "knowledge_base_arn" {
  value = module.knowledge_base.knowledge_base_arn
}

output "documents_bucket_name" {
  value = module.knowledge_base.documents_bucket_name
}

output "vectors_bucket_name" {
  value = module.knowledge_base.vectors_bucket_name
}

output "index_arn" {
  value = module.knowledge_base.index_arn
}
```

## Deployment Workflow

### 1. Deploy Infrastructure

```bash
cd terraform/environments/dev
terraform init -upgrade  # Upgrade to AWS Provider 6.25.0+
terraform apply
```

This creates:
- S3 documents bucket (`aws_s3_bucket`)
- S3 vectors bucket (`aws_s3_bucket` with encryption and versioning)
- Vector index (`aws_bedrockagent_vector_index` with cosine similarity)
- IAM role with S3 and vector operation permissions
- Bedrock Knowledge Base (`aws_bedrockagent_knowledge_base` with S3_VECTORS storage)
- S3 Data Source (`aws_bedrockagent_data_source`)

### 2. Upload Documents

```bash
# Get bucket name from Terraform
BUCKET=$(terraform output -raw documents_bucket_name)

# Upload your documents
aws s3 cp document.pdf s3://$BUCKET/
aws s3 cp documents/ s3://$BUCKET/ --recursive
```

### 3. Trigger Ingestion

```bash
# Use the provided script
cd ../../modules/knowledge-base/scripts
./ingest-kb.sh

# Or manually
KB_ID=$(cd ../../../environments/dev && terraform output -raw knowledge_base_id)
DS_ID=$(cd ../../../environments/dev && terraform output -raw data_source_id)

aws bedrock-agent start-ingestion-job \
  --knowledge-base-id "$KB_ID" \
  --data-source-id "$DS_ID" \
  --region ap-southeast-1
```

### 4. Update Application Config

```bash
# Get Knowledge Base ID
cd terraform/environments/dev
terraform output knowledge_base_id

# Add to .env file
echo "BEDROCK_KNOWLEDGE_BASE_ID=$(terraform output -raw knowledge_base_id)" >> ../../../.env
```

## Scripts

### ingest-kb.sh

Triggers ingestion job and monitors progress.

**Usage:**
```bash
cd terraform/modules/knowledge-base/scripts
./ingest-kb.sh

# Or specify environment
ENVIRONMENT=prod ./ingest-kb.sh
```

**Features:**
- Auto-detects KB ID from Terraform outputs
- Real-time progress monitoring
- Shows statistics on completion
- Handles errors gracefully

## Common Workflows

### Initial Setup

```bash
# 1. Deploy infrastructure
cd terraform/environments/dev
terraform init
terraform apply

# 2. Verify in AWS Console
# Bedrock → Knowledge bases → bedrock-chat-poc-kb-dev

# 3. Update .env file
echo "BEDROCK_KNOWLEDGE_BASE_ID=$(terraform output -raw knowledge_base_id)" >> ../../../.env

# 4. Start your application
cd ../../../backend
go run cmd/api/main.go
```

### Adding New Documents

```bash
# 1. Upload documents to S3
BUCKET=$(cd terraform/environments/dev && terraform output -raw documents_bucket_name)
aws s3 cp documents/ s3://$BUCKET/ --recursive

# 2. Trigger ingestion
cd terraform/modules/knowledge-base/scripts
./ingest-kb.sh

# 3. Wait for completion (automatic monitoring)
```

### Switching Environments

```bash
# Deploy to production
cd terraform/environments/prod
terraform apply

# Ingest in production
cd ../../modules/knowledge-base/scripts
ENVIRONMENT=prod ./ingest-kb.sh

# Destroy production
cd ../../../environments/prod
terraform destroy
```

## Troubleshooting

### Provider Version Issues

**Error: "aws_bedrockagent_knowledge_base resource not found"**

```bash
# Solution: Upgrade AWS provider to 6.25.0+
cd terraform/environments/dev
terraform init -upgrade

# Verify provider version
terraform version
# Should show: provider registry.terraform.io/hashicorp/aws v6.25.0 or higher
```

**Error: "Provider version does not meet requirements"**

Update your `terraform` block to require AWS provider 6.25.0+:
```hcl
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 6.25.0"
    }
  }
}
```

### Migration from AWSCC Provider

**Error: "Resource already exists in AWS but not in state"**

If migrating from AWSCC provider to AWS provider:

```bash
# Option 1: Clean slate (recommended for dev)
terraform destroy
terraform init -upgrade
terraform apply

# Option 2: State manipulation (for production)
# Remove old AWSCC resources from state
terraform state rm awscc_bedrock_knowledge_base.main
terraform state rm awscc_bedrock_data_source.s3

# Import as AWS provider resources
terraform import aws_bedrockagent_knowledge_base.main <kb-id>
terraform import aws_bedrockagent_data_source.s3 <kb-id>,<ds-id>
```

See [MIGRATION.md](./MIGRATION.md) for detailed migration instructions.

### Vector Index Configuration

**Error: "InvalidParameterException: dimension must match embedding model"**

Ensure vector index dimensions match your embedding model:
- Titan models: 1536 dimensions (default)
- Cohere models: 1024 dimensions

Update the dimension in `main.tf`:
```hcl
resource "aws_bedrockagent_vector_index" "main" {
  dimension = 1536  # or 1024 for Cohere
  # ...
}
```

### IAM Permission Errors

**Error: "AccessDeniedException: User is not authorized"**

Verify the Knowledge Base IAM role has all required permissions:

```bash
# Check IAM policy
aws iam get-role-policy \
  --role-name <kb-role-name> \
  --policy-name <policy-name>
```

Required permissions:
- **S3 Documents**: `s3:GetObject`, `s3:ListBucket`
- **S3 Vectors**: `s3:GetObject`, `s3:PutObject`, `s3:DeleteObject`, `s3:ListBucket`
- **Vector Index**: `bedrock:Query`, `bedrock:PutVector`, `bedrock:GetVector`, `bedrock:DeleteVector`
- **Embedding Model**: `bedrock:InvokeModel`

### Knowledge Base Creation Fails

**Error: "InvalidS3BucketException: Cannot access S3 bucket"**

```bash
# Verify bucket exists and is accessible
aws s3 ls s3://$(terraform output -raw vectors_bucket_name)

# Check bucket policy doesn't deny access
aws s3api get-bucket-policy --bucket $(terraform output -raw vectors_bucket_name)
```

**Error: "InvalidParameterException: storage configuration is invalid"**

Verify storage configuration uses correct ARN references:
- `vector_bucket_arn`: Must be S3 bucket ARN (not bucket name)
- `index_arn`: Must be vector index ARN

### Ingestion Issues

**Error: "Ingestion job stuck in IN_PROGRESS"**

```bash
# Check ingestion job status
KB_ID=$(terraform output -raw knowledge_base_id)
DS_ID=$(terraform output -raw data_source_id)

aws bedrock-agent list-ingestion-jobs \
  --knowledge-base-id "$KB_ID" \
  --data-source-id "$DS_ID" \
  --region ap-southeast-1 \
  --max-results 5

# Get detailed job info
aws bedrock-agent get-ingestion-job \
  --knowledge-base-id "$KB_ID" \
  --data-source-id "$DS_ID" \
  --ingestion-job-id "<job-id>" \
  --region ap-southeast-1
```

**Error: "No documents found in S3 bucket"**

```bash
# Verify documents are uploaded
BUCKET=$(terraform output -raw documents_bucket_name)
aws s3 ls s3://$BUCKET/ --recursive

# Check IAM role can access documents
aws s3 cp s3://$BUCKET/test.pdf /tmp/test.pdf --profile <kb-role-profile>
```

### Terraform State Issues

**Error: "Error acquiring state lock"**

```bash
# Wait for other operations to complete, or force unlock (use with caution)
terraform force-unlock <lock-id>
```

**Error: "Resource already exists"**

```bash
# Import existing resource into state
terraform import aws_bedrockagent_knowledge_base.main <kb-id>
```

### Terraform Destroy Fails

**Error: "BucketNotEmpty: The bucket you tried to delete is not empty"**

```bash
# Empty S3 buckets before destroying
DOCS_BUCKET=$(terraform output -raw documents_bucket_name)
VECTORS_BUCKET=$(terraform output -raw vectors_bucket_name)

aws s3 rm s3://$DOCS_BUCKET --recursive
aws s3 rm s3://$VECTORS_BUCKET --recursive

# Then destroy
terraform destroy
```

### Query Performance Issues

**Issue: Queries taking too long**

1. Check vector index configuration (should use cosine similarity)
2. Verify embedding model matches index dimensions
3. Monitor Bedrock API latency in CloudWatch
4. Consider caching frequent queries in your application

**Issue: Poor search results**

1. Verify documents were ingested successfully
2. Check embedding model is appropriate for your content
3. Review document chunking strategy
4. Test with different query phrasings

## Cost Optimization

### S3 Vectors vs OpenSearch Serverless

**S3 Vectors Storage (This Module):**
- **Storage**: ~$0.023/GB/month (S3 Standard)
- **Vector Operations**: Included with Bedrock pricing
- **Queries**: ~$0.0004 per 1000 queries
- **Minimum Cost**: None - pay only for what you use
- **Typical POC Cost**: $5-10/month
- **Best For**: POCs, development, cost-sensitive production workloads

**OpenSearch Serverless:**
- **Minimum**: ~$700/month (2 OCUs required, even with no data)
- **Storage**: Additional costs for data storage
- **Scaling**: Automatic but expensive
- **Best For**: High-throughput production workloads requiring sub-millisecond latency

**Cost Savings: ~$690/month (99% reduction)** 💰

### Why S3 Vectors?

1. **No Minimum Cost**: Pay only for storage and queries used
2. **Seamless Integration**: Native Bedrock Knowledge Base support
3. **Automatic Scaling**: S3 handles all scaling automatically
4. **Simple Management**: No cluster management or capacity planning
5. **Fast Enough**: Sub-second query latency for most use cases
6. **Production Ready**: Fully supported by AWS for production workloads

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│         aws_bedrockagent_knowledge_base (S3_VECTORS)            │
│  ┌───────────────────────────────────────────────────────────┐ │
│  │  Titan Embeddings G1 - Text v1.2 (1536 dimensions)       │ │
│  │  Storage Type: S3_VECTORS                                 │ │
│  └───────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
                              │
                              │
        ┌─────────────────────┼─────────────────────┐
        │                     │                     │
        ▼                     ▼                     ▼
┌──────────────────┐  ┌──────────────────┐  ┌─────────────────────┐
│ aws_s3_bucket    │  │ aws_s3_bucket    │  │ aws_bedrockagent_   │
│ (documents)      │  │ (vectors)        │  │ vector_index        │
│                  │  │                  │  │                     │
│ - Source docs    │  │ - Embeddings     │  │ - Dimension: 1536   │
│ - Read-only      │  │ - Encrypted      │  │ - Metric: cosine    │
│                  │  │ - Versioned      │  │ - Type: float32     │
└──────────────────┘  └──────────────────┘  └─────────────────────┘
        │                     │                     │
        └─────────────────────┴─────────────────────┘
                              │
                              ▼
                    ┌──────────────────┐
                    │ aws_bedrockagent_│
                    │ data_source      │
                    │ (S3)             │
                    └──────────────────┘

Resource Dependencies:
1. S3 Vectors Bucket → Vector Index → Knowledge Base
2. S3 Documents Bucket → Data Source → Knowledge Base
3. IAM Role → Knowledge Base
```

## Module Structure

```
knowledge-base/
├── main.tf              # Main resources (KB, buckets, IAM)
├── variables.tf         # Input variables
├── outputs.tf           # Output values
├── README.md           # This file
├── S3_VECTORS_SETUP.md # Legacy manual setup guide
└── scripts/
    └── ingest-kb.sh    # Ingestion helper script
```

## Migration from AWSCC Provider

If you're upgrading from a previous version that used the AWSCC provider, see [MIGRATION.md](./MIGRATION.md) for detailed migration instructions. The module now uses the official AWS provider exclusively for better stability and support.

**Key Changes:**
- `awscc_bedrock_knowledge_base` → `aws_bedrockagent_knowledge_base`
- `awscc_bedrock_data_source` → `aws_bedrockagent_data_source`
- Custom S3 Vectors resources → `aws_bedrockagent_vector_index`
- AWSCC provider removed entirely

## Legacy Documentation

See [S3_VECTORS_SETUP.md](./S3_VECTORS_SETUP.md) for the legacy manual setup approach (deprecated).

## Support

For issues or questions:
1. Check [TROUBLESHOOTING.md](../../../TROUBLESHOOTING.md)
2. Review AWS Bedrock documentation
3. Verify IAM permissions and quotas
