# S3 Vectors Setup Guide

AWS provider 6.25.0+ now supports S3 Vectors! This module creates everything via Terraform.

## Quick Start (Terraform)

```bash
# Deploy everything with Terraform
cd terraform/environments/dev
terraform init -upgrade  # Upgrade to AWS provider 6.25.0+
terraform apply

# Get Knowledge Base ID from outputs
terraform output knowledge_base_id

# Upload documents and trigger ingestion
aws s3 cp my-doc.pdf s3://$(terraform output -raw documents_bucket_name)/
cd ../../modules/knowledge-base/scripts
./ingest-kb.sh
```

## What Gets Created

- **S3 Documents Bucket**: Stores your source documents (AWS provider)
- **S3 Vectors Bucket**: Stores vector embeddings (AWS provider 6.25.0+)
- **S3 Vectors Index**: Vector index with cosine similarity (AWS provider 6.25.0+)
- **IAM Role**: Bedrock service role with appropriate permissions
- **Knowledge Base**: Bedrock Knowledge Base with Titan embeddings (AWSCC provider)
- **Data Source**: S3 data source linked to documents bucket (AWSCC provider)

## Alternative: CLI Scripts (Legacy)

If you need more control or prefer CLI automation:

## Step 1: Deploy Infrastructure with Terraform

First, deploy S3 buckets and IAM roles:

```bash
cd terraform/environments/dev
terraform apply
```

This creates:
- S3 documents bucket: `bedrock-chat-poc-kb-docs-dev`
- S3 vector bucket: `bedrock-chat-poc-kb-vectors-dev`
- IAM role: `bedrock-chat-poc-kb-role-dev`

## Step 2: Create Knowledge Base via AWS Console

1. Open AWS Console → Amazon Bedrock → Knowledge bases
2. Click **"Create knowledge base"**

### Configuration Details

**Knowledge base details:**
- Name: `bedrock-chat-poc-kb-dev`
- Description: `Knowledge base for chat POC (dev environment)`
- IAM role: Select **"Use an existing service role"**
  - Choose: `bedrock-chat-poc-kb-role-dev`

**Set up data source:**
- Data source name: `bedrock-chat-poc-kb-dev-s3-data-source`
- S3 URI: `s3://bedrock-chat-poc-kb-docs-dev/`

**Select embeddings model:**
- Embeddings model: **Titan Embeddings G1 - Text v1.2**
- Model dimensions: 1536

**Configure vector store:**
- Vector database: Select **"Amazon S3"**
- Quick create: **Yes**
- S3 bucket for vectors: `bedrock-chat-poc-kb-vectors-dev`

**Tags:**
```
Environment = dev
Project = bedrock-chat-poc
ManagedBy = Terraform
```

3. Click **"Create knowledge base"**
4. Wait 2-3 minutes for creation

## Step 3: Get Knowledge Base ID

After creation completes:

```bash
# Get Knowledge Base ID
aws bedrock-agent list-knowledge-bases \
  --region ap-southeast-1 \
  --query "knowledgeBaseSummaries[?name=='bedrock-chat-poc-kb-dev'].knowledgeBaseId | [0]" \
  --output text
```

Save this ID - you'll need it for your application.

## Step 4: Update Environment Variables

Update your `.env` file:

```bash
BEDROCK_KNOWLEDGE_BASE_ID=<your-kb-id-from-step-3>
```

## Step 5: Test Knowledge Base

Upload a test document:

```bash
echo "This is a test document for the knowledge base." > test.txt
aws s3 cp test.txt s3://bedrock-chat-poc-kb-docs-dev/
```

Start ingestion job:

```bash
KB_ID="<your-kb-id>"
DS_ID=$(aws bedrock-agent list-data-sources \
  --knowledge-base-id "$KB_ID" \
  --region ap-southeast-1 \
  --query 'dataSourceSummaries[0].dataSourceId' \
  --output text)

aws bedrock-agent start-ingestion-job \
  --knowledge-base-id "$KB_ID" \
  --data-source-id "$DS_ID" \
  --region ap-southeast-1
```

Check ingestion status:

```bash
aws bedrock-agent list-ingestion-jobs \
  --knowledge-base-id "$KB_ID" \
  --data-source-id "$DS_ID" \
  --region ap-southeast-1 \
  --max-results 1
```

## Cleanup

To delete everything:

```bash
# 1. Delete data source
aws bedrock-agent delete-data-source \
  --knowledge-base-id "$KB_ID" \
  --data-source-id "$DS_ID" \
  --region ap-southeast-1

# 2. Wait for deletion
sleep 10

# 3. Delete knowledge base
aws bedrock-agent delete-knowledge-base \
  --knowledge-base-id "$KB_ID" \
  --region ap-southeast-1

# 4. Destroy Terraform resources
cd terraform/environments/dev
terraform destroy
```

## Cost Estimate

**S3 Vectors:**
- Storage: ~$0.023/GB/month
- Queries: ~$0.0004 per 1000 queries
- **Total for POC**: ~$5-10/month

**vs OpenSearch Serverless:**
- Minimum: ~$700/month (2 OCUs)

**Savings: ~$690/month** 💰

## Why Manual Setup?

- Terraform AWS provider doesn't support S3 Vectors storage configuration
- AWSCC provider's S3 Vectors type is not available via Cloud Control API yet
- AWS Console provides full S3 Vectors support
- This is temporary until Terraform providers add support

## Future Migration

When Terraform adds S3 Vectors support, you can:
1. Import existing Knowledge Base into Terraform state
2. Update module to use native Terraform resources
3. Continue managing with `terraform apply`
