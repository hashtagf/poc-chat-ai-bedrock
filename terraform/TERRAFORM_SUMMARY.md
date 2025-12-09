# Terraform Configuration Summary

## Overview

Complete Terraform infrastructure for Bedrock Chat POC with Knowledge Base integration.

## Structure

```
terraform/
├── environments/
│   ├── dev/          # Development environment
│   ├── staging/      # Staging environment (template)
│   └── prod/         # Production environment (template)
├── modules/
│   ├── bedrock-agent/      # Bedrock Agent resources
│   ├── iam/                # IAM roles (Agent only)
│   ├── knowledge-base/     # Complete KB with S3 Vectors
│   ├── state-backend/      # Remote state setup
│   └── vpc/                # VPC resources (optional)
└── backend.tf              # Remote state configuration
```

## Modules

### 1. knowledge-base

**Purpose**: Complete Knowledge Base with S3 Vectors storage

**Resources Created**:
- S3 Documents Bucket (AWS provider)
- S3 Vectors Bucket (AWS provider 6.25.0+)
- S3 Vectors Index (AWS provider 6.25.0+)
- IAM Role for Knowledge Base
- Bedrock Knowledge Base (AWSCC provider)
- S3 Data Source (AWSCC provider)

**Inputs**:
- `project_name`: Project name for resource naming
- `environment`: Environment (dev/staging/prod)
- `tags`: Resource tags

**Outputs**:
- `knowledge_base_id`: KB ID for application
- `knowledge_base_arn`: KB ARN
- `documents_bucket_name`: S3 bucket for documents
- `vectors_bucket_name`: S3 bucket for vectors
- `index_name`: Vector index name
- `index_arn`: Vector index ARN
- `data_source_id`: Data source ID
- `role_arn`: IAM role ARN

### 2. iam

**Purpose**: IAM roles for Bedrock Agent

**Resources Created**:
- Agent IAM Role
- Agent IAM Policy

**Inputs**:
- `project_name`: Project name
- `environment`: Environment
- `foundation_model_arn`: Model ARN for permissions
- `tags`: Resource tags

**Outputs**:
- `agent_role_arn`: Agent role ARN
- `agent_role_name`: Agent role name

### 3. bedrock-agent

**Purpose**: Bedrock Agent with alias

**Resources Created**:
- Bedrock Agent
- Agent Alias

**Inputs**:
- `agent_name`: Agent name
- `foundation_model`: Model ID
- `agent_instruction`: Agent instructions
- `agent_role_arn`: IAM role ARN
- `idle_session_ttl`: Session timeout
- `tags`: Resource tags

**Outputs**:
- `agent_id`: Agent ID
- `agent_arn`: Agent ARN
- `agent_alias_id`: Alias ID

## Provider Requirements

- **AWS Provider**: >= 6.25.0 (for S3 Vectors support)
- **AWSCC Provider**: >= 1.0.0 (for Knowledge Base)
- **Terraform**: >= 1.5.0

## Deployment

### Initial Setup

```bash
# 1. Navigate to environment
cd terraform/environments/dev

# 2. Initialize Terraform
terraform init -upgrade

# 3. Review plan
terraform plan

# 4. Apply configuration
terraform apply

# 5. Get outputs
terraform output
```

### Update .env File

```bash
# Get Knowledge Base ID
KB_ID=$(terraform output -raw knowledge_base_id)
echo "BEDROCK_KNOWLEDGE_BASE_ID=$KB_ID" >> ../../../.env

# Get Agent IDs
AGENT_ID=$(terraform output -raw bedrock_agent_id)
ALIAS_ID=$(terraform output -raw bedrock_agent_alias_id)
echo "BEDROCK_AGENT_ID=$AGENT_ID" >> ../../../.env
echo "BEDROCK_AGENT_ALIAS_ID=$ALIAS_ID" >> ../../../.env
```

### Upload Documents

```bash
# Get bucket name
BUCKET=$(terraform output -raw documents_bucket_name)

# Upload documents
aws s3 cp my-document.pdf s3://$BUCKET/
aws s3 cp documents/ s3://$BUCKET/ --recursive
```

### Trigger Ingestion

```bash
cd ../../modules/knowledge-base/scripts
./ingest-kb.sh
```

## Configuration Files

### terraform.tfvars

```hcl
# General
environment  = "dev"
project_name = "bedrock-chat-poc"
aws_region   = "ap-southeast-1"

# Bedrock Agent
agent_name        = "bedrock-chat-poc-agent-dev"
foundation_model  = "amazon.nova-2-lite-v1:0"
agent_instruction = "You are a helpful AI assistant."
idle_session_ttl  = 1800

# VPC (optional)
enable_vpc = false

# Tags
tags = {
  Environment = "dev"
  Project     = "bedrock-chat-poc"
  ManagedBy   = "Terraform"
}
```

## Validation

```bash
# Format code
terraform fmt -recursive

# Validate configuration
terraform validate

# Check for issues
terraform plan
```

## Cost Estimation

### S3 Vectors (Knowledge Base)
- Storage: ~$0.023/GB/month
- Queries: ~$0.0004 per 1000 queries
- **Typical POC**: $5-10/month

### Bedrock Agent
- Model invocations: Pay per token
- **Typical POC**: $10-50/month

### Total Estimated Cost
- **POC Environment**: $15-60/month
- **vs OpenSearch Serverless**: $700+/month
- **Savings**: ~$640-685/month (95%+)

## Troubleshooting

### Provider Version Issues

```bash
# Upgrade providers
terraform init -upgrade

# Check versions
terraform version
terraform providers
```

### Validation Errors

```bash
# Check configuration
terraform validate

# Review plan
terraform plan

# Check module sources
terraform get -update
```

### State Issues

```bash
# Refresh state
terraform refresh

# List resources
terraform state list

# Show specific resource
terraform state show module.knowledge_base.aws_s3_bucket.documents
```

## Migration Notes

### From Manual Setup

If you previously created resources manually:

1. **Import existing resources**:
```bash
terraform import module.knowledge_base.aws_s3_bucket.documents bucket-name
```

2. **Update state**:
```bash
terraform plan  # Verify no changes
```

3. **Continue with Terraform**:
```bash
terraform apply
```

## Best Practices

1. **Always use remote state** for team collaboration
2. **Separate environments** (dev/staging/prod)
3. **Tag all resources** for cost tracking
4. **Version control** all .tf files
5. **Review plans** before applying
6. **Use workspaces** for environment isolation (optional)

## References

- [AWS Provider Documentation](https://registry.terraform.io/providers/hashicorp/aws/latest/docs)
- [AWSCC Provider Documentation](https://registry.terraform.io/providers/hashicorp/awscc/latest/docs)
- [Bedrock Knowledge Base](https://docs.aws.amazon.com/bedrock/latest/userguide/knowledge-base.html)
- [S3 Vectors](https://docs.aws.amazon.com/AmazonS3/latest/userguide/s3-vectors.html)
