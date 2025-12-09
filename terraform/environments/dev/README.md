# Development Environment

This directory contains the Terraform configuration for the development environment of the Bedrock Chat POC infrastructure.

## Overview

The development environment includes:
- Bedrock Agent with Claude foundation model
- Knowledge Base with S3 vector store
- IAM roles and policies for Bedrock services
- S3 buckets for documents and vector storage
- **VPC is disabled** for cost optimization

## Prerequisites

- Terraform >= 1.5.0
- AWS CLI configured with appropriate credentials
- AWS account with Bedrock service enabled
- Permissions to create IAM roles, S3 buckets, and Bedrock resources

## Configuration

The environment is configured through `terraform.tfvars`:

```hcl
environment  = "dev"
project_name = "bedrock-chat-poc"
aws_region   = "ap-southeast-1"

agent_name        = "bedrock-chat-poc-agent-dev"
foundation_model  = "anthropic.claude-v2"
knowledge_base_name   = "bedrock-chat-poc-kb-dev"
embedding_model       = "amazon.titan-embed-text-v1"

enable_vpc = false  # VPC disabled for dev
```

## Deployment Steps

### 1. Bootstrap State Backend (First Time Only)

Before deploying the dev environment, you need to create the S3 bucket for Terraform state:

```bash
cd ../../modules/state-backend
terraform init
terraform apply -var="state_bucket_name=bedrock-chat-poc-terraform-state"
```

### 2. Update Backend Configuration

Edit `backend.tf` and update the bucket name to match your state bucket:

```hcl
terraform {
  backend "s3" {
    bucket = "bedrock-chat-poc-terraform-state"  # Update this
    key    = "environments/dev/terraform.tfstate"
    region = "ap-southeast-1"
    encrypt = true
  }
}
```

### 3. Initialize Terraform

```bash
terraform init
```

This will:
- Download required providers
- Initialize modules
- Configure the S3 backend

### 4. Review the Plan

```bash
terraform plan -var-file=terraform.tfvars
```

Review the resources that will be created.

### 5. Apply the Configuration

```bash
terraform apply -var-file=terraform.tfvars
```

Type `yes` when prompted to confirm.

## Outputs

After successful deployment, Terraform will output the following values needed for the backend application:

```bash
terraform output
```

Example output:
```
bedrock_agent_id         = "ABCDEFGHIJ"
bedrock_agent_alias_id   = "TSTALIASID"
bedrock_knowledge_base_id = "KBID123456"
s3_bucket_name           = "bedrock-chat-poc-kb-docs-dev"
aws_region               = "ap-southeast-1"
```

### Configure Backend Application

Update your backend `.env` file with these values:

```bash
# Get outputs in JSON format
terraform output -json > outputs.json

# Or manually copy values
export BEDROCK_AGENT_ID=$(terraform output -raw bedrock_agent_id)
export BEDROCK_AGENT_ALIAS_ID=$(terraform output -raw bedrock_agent_alias_id)
export BEDROCK_KNOWLEDGE_BASE_ID=$(terraform output -raw bedrock_knowledge_base_id)
export AWS_REGION=$(terraform output -raw aws_region)
```

## Knowledge Base Setup

### Upload Documents

Upload documents to the S3 bucket for the Knowledge Base to ingest:

```bash
# Get the bucket name
BUCKET_NAME=$(terraform output -raw s3_bucket_name)

# Upload documents
aws s3 cp ./documents/ s3://$BUCKET_NAME/ --recursive
```

### Sync Knowledge Base

After uploading documents, trigger a sync job:

```bash
# Get the Knowledge Base ID and Data Source ID
KB_ID=$(terraform output -raw bedrock_knowledge_base_id)
DS_ID=$(terraform output -raw data_source_id)

# Start ingestion job
aws bedrock-agent start-ingestion-job \
  --knowledge-base-id $KB_ID \
  --data-source-id $DS_ID \
  --region ap-southeast-1
```

Monitor the ingestion job status:

```bash
aws bedrock-agent list-ingestion-jobs \
  --knowledge-base-id $KB_ID \
  --data-source-id $DS_ID \
  --region ap-southeast-1
```

## Maintenance

### Update Configuration

1. Modify `terraform.tfvars` with new values
2. Run `terraform plan` to preview changes
3. Run `terraform apply` to apply changes

### Destroy Resources

To tear down the development environment:

```bash
terraform destroy -var-file=terraform.tfvars
```

**Warning**: This will delete all resources including S3 buckets and their contents.

## Cost Optimization

The development environment is optimized for cost:
- VPC is disabled (no VPC endpoint charges)
- Single-region deployment
- S3 vector store (no OpenSearch Serverless charges)
- On-demand Bedrock pricing

Estimated monthly cost: $10-50 depending on usage.

## Troubleshooting

### State Locking Issues

If you encounter state locking errors, S3 native locking will automatically expire stale locks. Wait a few minutes and retry.

### Agent Preparation Failures

The agent requires a preparation step after creation. If this fails:

```bash
# Manually prepare the agent
aws bedrock-agent prepare-agent \
  --agent-id $(terraform output -raw bedrock_agent_id) \
  --region ap-southeast-1
```

### Permission Errors

Ensure your AWS credentials have permissions to:
- Create IAM roles and policies
- Create S3 buckets
- Create Bedrock agents and knowledge bases
- Invoke Bedrock models

### Module Not Found

If Terraform can't find modules, run:

```bash
terraform init -upgrade
```

## Next Steps

After deploying the infrastructure:

1. Configure the backend application with Terraform outputs
2. Upload documents to the Knowledge Base S3 bucket
3. Trigger Knowledge Base sync
4. Test the agent through the backend API
5. Monitor CloudWatch logs for errors

## Related Documentation

- [Main Terraform README](../../README.md)
- [State Backend Module](../../modules/state-backend/README.md)
- [Knowledge Base Module](../../modules/knowledge-base/README.md)
- [Backend Configuration](../../../backend/docs/CONFIGURATION.md)
