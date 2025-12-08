# Terraform Infrastructure for Bedrock Chat POC

This directory contains Terraform configurations for provisioning Amazon Bedrock infrastructure including Bedrock Agent, Knowledge Base, IAM roles, and optional VPC setup.

## Directory Structure

```
terraform/
├── modules/                    # Reusable Terraform modules
│   ├── state-backend/         # S3 bucket for remote state storage
│   ├── iam/                   # IAM roles and policies (to be created)
│   ├── bedrock-agent/         # Bedrock Agent configuration (to be created)
│   ├── knowledge-base/        # Knowledge Base with S3 vector store (to be created)
│   └── vpc/                   # VPC with endpoints for production (to be created)
├── environments/              # Environment-specific configurations
│   ├── dev/                   # Development environment
│   ├── staging/               # Staging environment
│   └── prod/                  # Production environment
└── backend.tf                 # Remote state backend configuration template
```

## Prerequisites

### Required Software Versions

- **Terraform**: >= 1.5.0 (recommended: latest 1.x version)
  - Download: https://www.terraform.io/downloads.html
  - Verify: `terraform version`
  
- **AWS CLI**: >= 2.13.0 (recommended: latest 2.x version)
  - Download: https://aws.amazon.com/cli/
  - Verify: `aws --version`

### AWS Account Requirements

- AWS account with permissions to create:
  - S3 buckets
  - IAM roles and policies
  - Bedrock Agent and Knowledge Base resources
  - VPC resources (for production)
  
- AWS credentials configured via one of:
  - AWS CLI: `aws configure`
  - Environment variables: `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`
  - IAM role (recommended for EC2/ECS deployments)

### Bedrock Service Access

- Bedrock service must be enabled in your AWS account
- Access to required foundation models (e.g., Claude, Titan)
- Access to embedding models (e.g., Titan Embeddings)
- Verify model access in AWS Console: Bedrock → Model access

## Getting Started

### Step 1: Bootstrap State Backend

Before deploying any infrastructure, you need to create the S3 bucket for storing Terraform state.

1. Navigate to the state-backend module:
   ```bash
   cd modules/state-backend
   ```

2. Create a `terraform.tfvars` file:
   ```hcl
   state_bucket_name = "bedrock-chat-poc-terraform-state-dev"
   tags = {
     Environment = "dev"
     Project     = "bedrock-chat-poc"
     ManagedBy   = "Terraform"
     CreatedAt   = "2025-12-07"
   }
   ```

3. Initialize and apply:
   ```bash
   terraform init
   terraform plan
   terraform apply
   ```

4. Note the bucket name from the output - you'll need it for the backend configuration.

### Step 2: Configure Remote State Backend

After creating the state bucket, configure each environment to use it:

1. Copy the backend configuration to your environment directory:
   ```bash
   cp backend.tf environments/dev/backend.tf
   ```

2. Update the backend configuration with your state bucket name:
   ```hcl
   terraform {
     backend "s3" {
       bucket  = "bedrock-chat-poc-terraform-state-dev"
       key     = "dev/terraform.tfstate"
       region  = "us-east-1"
       encrypt = true
     }
   }
   ```

3. Initialize Terraform with the backend:
   ```bash
   cd environments/dev
   terraform init
   ```

### Step 3: Deploy Infrastructure

Once the backend is configured, you can deploy the infrastructure:

1. Navigate to your environment directory:
   ```bash
   cd environments/dev
   ```

2. Review the `terraform.tfvars` file and customize as needed:
   ```hcl
   # General
   environment     = "dev"
   project_name    = "bedrock-chat-poc"
   aws_region      = "us-east-1"

   # Bedrock Agent
   agent_name           = "bedrock-chat-poc-agent-dev"
   foundation_model     = "anthropic.claude-v2"
   agent_instruction    = "You are a helpful AI assistant."
   idle_session_ttl     = 1800

   # Knowledge Base
   knowledge_base_name         = "bedrock-chat-poc-kb-dev"
   embedding_model             = "amazon.titan-embed-text-v1"
   s3_bucket_name              = "bedrock-chat-poc-kb-docs-dev"
   s3_vector_bucket_name       = "bedrock-chat-poc-kb-vectors-dev"

   # VPC (disabled for dev)
   enable_vpc = false

   # Tags
   tags = {
     Environment = "dev"
     Project     = "bedrock-chat-poc"
     ManagedBy   = "Terraform"
     CreatedAt   = "2025-12-07"
   }
   ```

3. Initialize Terraform (if not already done):
   ```bash
   terraform init
   ```

4. Review the planned changes:
   ```bash
   terraform plan
   ```

5. Apply the configuration:
   ```bash
   terraform apply
   ```

6. Confirm the changes by typing `yes` when prompted

7. Save the outputs for application configuration:
   ```bash
   terraform output > terraform-outputs.txt
   ```

## State Management

### Remote State with S3

This project uses S3 for remote state storage with the following benefits:

- **Centralized State**: All team members access the same state file
- **Versioning**: S3 versioning enables rollback to previous states
- **Encryption**: State files are encrypted at rest and in transit
- **Native Locking**: S3 provides built-in state locking without DynamoDB

### S3 Native Locking

S3 backend uses conditional write operations for state locking:

- Prevents concurrent modifications by multiple users or CI/CD pipelines
- Automatic lock acquisition before operations
- Automatic lock release after completion
- Stale locks expire automatically (no manual intervention needed)

### Working with State

```bash
# View current state
terraform show

# List resources in state
terraform state list

# View specific resource
terraform state show <resource_address>

# Refresh state from actual infrastructure
terraform refresh
```

## Environment Management

### Development Environment

- Located in `environments/dev/`
- Uses minimal resources for cost savings
- VPC disabled (uses public endpoints)
- Suitable for testing and development

### Staging Environment

- Located in `environments/staging/`
- Mirrors production configuration
- VPC disabled for simplicity
- Used for pre-production validation

### Production Environment

- Located in `environments/prod/`
- Full security configuration
- VPC enabled with private subnets and endpoints
- Multi-AZ deployment for high availability

## Updating Application Configuration from Terraform Outputs

After deploying the infrastructure, you need to configure your backend application with the Terraform outputs.

### Step 1: Extract Terraform Outputs

```bash
cd environments/dev
terraform output
```

You'll see outputs like:
```
bedrock_agent_id = "ABCDEFGHIJ"
bedrock_agent_alias_id = "TSTALIASID"
bedrock_knowledge_base_id = "KB123456789"
s3_bucket_name = "bedrock-chat-poc-kb-docs-dev"
aws_region = "us-east-1"
```

### Step 2: Update Backend Environment Variables

Update your backend `.env` file or environment configuration:

```bash
# Navigate to backend directory
cd ../../backend

# Update .env file
cat > .env << EOF
AWS_REGION=us-east-1
BEDROCK_AGENT_ID=ABCDEFGHIJ
BEDROCK_AGENT_ALIAS_ID=TSTALIASID
BEDROCK_KNOWLEDGE_BASE_ID=KB123456789
S3_BUCKET_NAME=bedrock-chat-poc-kb-docs-dev
EOF
```

### Step 3: Verify Configuration

```bash
# Source the environment variables
source .env

# Verify they're set
echo $BEDROCK_AGENT_ID
echo $BEDROCK_AGENT_ALIAS_ID
```

### Step 4: Restart Application

```bash
# If running locally
make run

# If using Docker
docker-compose down
docker-compose up --build
```

### Automated Configuration Script

You can automate this process with a script:

```bash
#!/bin/bash
# update-config.sh

cd terraform/environments/dev

# Extract outputs
AGENT_ID=$(terraform output -raw bedrock_agent_id)
AGENT_ALIAS_ID=$(terraform output -raw bedrock_agent_alias_id)
KB_ID=$(terraform output -raw bedrock_knowledge_base_id)
S3_BUCKET=$(terraform output -raw s3_bucket_name)
REGION=$(terraform output -raw aws_region)

# Update backend .env
cd ../../../backend
cat > .env << EOF
AWS_REGION=$REGION
BEDROCK_AGENT_ID=$AGENT_ID
BEDROCK_AGENT_ALIAS_ID=$AGENT_ALIAS_ID
BEDROCK_KNOWLEDGE_BASE_ID=$KB_ID
S3_BUCKET_NAME=$S3_BUCKET
EOF

echo "✓ Backend configuration updated successfully"
```

## Knowledge Base Document Management

### Uploading Documents to Knowledge Base

After deploying the infrastructure, you can upload documents to the Knowledge Base:

#### Step 1: Prepare Your Documents

Supported formats:
- PDF (.pdf)
- Text files (.txt)
- Markdown (.md)
- HTML (.html)
- Microsoft Word (.doc, .docx)

Best practices:
- Use clear, descriptive filenames
- Keep documents focused on specific topics
- Include metadata in document headers when possible
- Limit individual file size to 50MB

#### Step 2: Upload to S3

```bash
# Get the S3 bucket name from Terraform outputs
cd terraform/environments/dev
S3_BUCKET=$(terraform output -raw s3_bucket_name)

# Upload a single document
aws s3 cp /path/to/document.pdf s3://$S3_BUCKET/

# Upload multiple documents
aws s3 cp /path/to/documents/ s3://$S3_BUCKET/ --recursive

# Upload with metadata
aws s3 cp document.pdf s3://$S3_BUCKET/ \
  --metadata "category=technical,version=1.0"
```

#### Step 3: Sync Knowledge Base

After uploading documents, trigger a sync to update the vector embeddings:

```bash
# Get Knowledge Base ID
KB_ID=$(terraform output -raw bedrock_knowledge_base_id)

# Get Data Source ID (from Terraform state or AWS Console)
DATA_SOURCE_ID=$(aws bedrock-agent list-data-sources \
  --knowledge-base-id $KB_ID \
  --query 'dataSourceSummaries[0].dataSourceId' \
  --output text)

# Start ingestion job
aws bedrock-agent start-ingestion-job \
  --knowledge-base-id $KB_ID \
  --data-source-id $DATA_SOURCE_ID

# Check ingestion job status
aws bedrock-agent list-ingestion-jobs \
  --knowledge-base-id $KB_ID \
  --data-source-id $DATA_SOURCE_ID
```

#### Step 4: Verify Ingestion

```bash
# Check ingestion job status
aws bedrock-agent get-ingestion-job \
  --knowledge-base-id $KB_ID \
  --data-source-id $DATA_SOURCE_ID \
  --ingestion-job-id <job-id>

# Expected status progression:
# STARTING → IN_PROGRESS → COMPLETE
```

### Automated Document Upload Script

```bash
#!/bin/bash
# upload-kb-docs.sh

set -e

# Configuration
DOCS_DIR="$1"
ENVIRONMENT="${2:-dev}"

if [ -z "$DOCS_DIR" ]; then
  echo "Usage: ./upload-kb-docs.sh <docs-directory> [environment]"
  exit 1
fi

# Get Terraform outputs
cd terraform/environments/$ENVIRONMENT
S3_BUCKET=$(terraform output -raw s3_bucket_name)
KB_ID=$(terraform output -raw bedrock_knowledge_base_id)

echo "📤 Uploading documents to $S3_BUCKET..."
aws s3 sync "$DOCS_DIR" "s3://$S3_BUCKET/" --delete

echo "🔄 Getting data source ID..."
DATA_SOURCE_ID=$(aws bedrock-agent list-data-sources \
  --knowledge-base-id $KB_ID \
  --query 'dataSourceSummaries[0].dataSourceId' \
  --output text)

echo "🚀 Starting ingestion job..."
JOB_ID=$(aws bedrock-agent start-ingestion-job \
  --knowledge-base-id $KB_ID \
  --data-source-id $DATA_SOURCE_ID \
  --query 'ingestionJob.ingestionJobId' \
  --output text)

echo "✓ Ingestion job started: $JOB_ID"
echo "Monitor status with:"
echo "  aws bedrock-agent get-ingestion-job \\"
echo "    --knowledge-base-id $KB_ID \\"
echo "    --data-source-id $DATA_SOURCE_ID \\"
echo "    --ingestion-job-id $JOB_ID"
```

### Document Update Workflow

When updating existing documents:

1. **Upload new version**: Upload the updated document to S3 (overwrites existing)
2. **Trigger sync**: Start a new ingestion job
3. **Wait for completion**: Monitor job status until COMPLETE
4. **Test queries**: Verify the Knowledge Base returns updated information

### Managing Document Versions

S3 versioning is enabled on the documents bucket:

```bash
# List all versions of a document
aws s3api list-object-versions \
  --bucket $S3_BUCKET \
  --prefix document.pdf

# Restore a previous version
aws s3api copy-object \
  --bucket $S3_BUCKET \
  --copy-source $S3_BUCKET/document.pdf?versionId=<version-id> \
  --key document.pdf
```

## Common Commands

```bash
# Format all Terraform files
terraform fmt -recursive

# Validate configuration
terraform validate

# Plan changes
terraform plan

# Plan with specific var file
terraform plan -var-file=terraform.tfvars

# Apply changes
terraform apply

# Apply without confirmation (use in CI/CD)
terraform apply -auto-approve

# Destroy infrastructure (use with caution!)
terraform destroy

# View outputs
terraform output

# View specific output
terraform output bedrock_agent_id

# View output in raw format (no quotes)
terraform output -raw bedrock_agent_id

# Update outputs without applying changes
terraform refresh

# Show current state
terraform show

# List all resources
terraform state list

# Import existing resource
terraform import <resource_type>.<name> <resource_id>
```

## Troubleshooting

### State Locking Issues

**Problem**: `Error acquiring the state lock`

**Causes**:
- Another Terraform operation is running
- Previous operation crashed without releasing lock
- Network interruption during operation

**Solutions**:

1. **Wait for automatic expiration**: S3 native locks expire automatically after a timeout
   ```bash
   # Wait 5-10 minutes and try again
   terraform plan
   ```

2. **Verify no other operations are running**:
   ```bash
   # Check for running Terraform processes
   ps aux | grep terraform
   
   # Check CI/CD pipelines
   # Verify no other team members are running Terraform
   ```

3. **Force unlock (last resort)**:
   ```bash
   # Only use if you're certain no other operations are running
   terraform force-unlock <lock-id>
   ```

**Prevention**:
- Coordinate with team members before running Terraform
- Use workspaces or separate state files for parallel work
- Implement proper CI/CD locking mechanisms

### Backend Initialization Errors

**Problem**: `Error configuring the backend "s3"`

**Common causes and solutions**:

1. **State bucket doesn't exist**:
   ```bash
   # Verify bucket exists
   aws s3 ls s3://your-state-bucket-name
   
   # If not, create it using the state-backend module
   cd modules/state-backend
   terraform init
   terraform apply
   ```

2. **Incorrect AWS credentials**:
   ```bash
   # Verify credentials are configured
   aws sts get-caller-identity
   
   # If not configured, run:
   aws configure
   ```

3. **Wrong region**:
   ```bash
   # Verify bucket region matches backend.tf
   aws s3api get-bucket-location --bucket your-state-bucket-name
   
   # Update backend.tf with correct region
   ```

4. **Insufficient permissions**:
   ```bash
   # Test S3 access
   aws s3 ls s3://your-state-bucket-name
   
   # Required permissions:
   # - s3:ListBucket
   # - s3:GetObject
   # - s3:PutObject
   # - s3:DeleteObject
   ```

### Permission Errors

**Problem**: `AccessDenied` or `UnauthorizedOperation`

**Solutions**:

1. **Verify IAM permissions**:
   ```bash
   # Check current identity
   aws sts get-caller-identity
   
   # Test specific permissions
   aws iam simulate-principal-policy \
     --policy-source-arn <your-role-arn> \
     --action-names bedrock:CreateAgent s3:CreateBucket
   ```

2. **Required IAM permissions for deployment**:
   - S3: CreateBucket, PutBucketVersioning, PutBucketEncryption
   - IAM: CreateRole, PutRolePolicy, AttachRolePolicy
   - Bedrock: CreateAgent, CreateKnowledgeBase, CreateDataSource
   - VPC: CreateVpc, CreateSubnet, CreateVpcEndpoint (for production)

3. **Check service-linked roles**:
   ```bash
   # Bedrock may require service-linked roles
   aws iam get-role --role-name AWSServiceRoleForAmazonBedrock
   
   # If missing, create it:
   aws iam create-service-linked-role --aws-service-name bedrock.amazonaws.com
   ```

### Bedrock Agent Creation Failures

**Problem**: `Error creating Bedrock Agent`

**Common causes**:

1. **Model not available in region**:
   ```bash
   # List available models
   aws bedrock list-foundation-models --region us-east-1
   
   # Verify your model is in the list
   ```

2. **Model access not granted**:
   - Go to AWS Console → Bedrock → Model access
   - Request access to required models (Claude, Titan)
   - Wait for approval (usually instant for most models)

3. **Invalid agent instruction**:
   - Ensure instruction is not empty
   - Keep instruction under 4000 characters
   - Avoid special characters that might cause parsing issues

4. **Agent preparation timeout**:
   ```bash
   # The agent needs time to prepare after creation
   # If you see timeout errors, increase the time_sleep duration
   # in modules/bedrock-agent/main.tf
   ```

### Knowledge Base Ingestion Failures

**Problem**: Ingestion job fails or gets stuck

**Solutions**:

1. **Check document format**:
   ```bash
   # Verify file is in supported format
   file document.pdf
   
   # Supported: PDF, TXT, MD, HTML, DOC, DOCX
   ```

2. **Check document size**:
   ```bash
   # Files should be under 50MB
   ls -lh document.pdf
   
   # If too large, split into smaller files
   ```

3. **Verify S3 permissions**:
   ```bash
   # Knowledge Base role needs read access
   aws s3 ls s3://$S3_BUCKET/
   
   # Check IAM role permissions in AWS Console
   ```

4. **Check ingestion job logs**:
   ```bash
   # Get detailed error information
   aws bedrock-agent get-ingestion-job \
     --knowledge-base-id $KB_ID \
     --data-source-id $DATA_SOURCE_ID \
     --ingestion-job-id $JOB_ID \
     --query 'ingestionJob.failureReasons'
   ```

### VPC Endpoint Connection Issues

**Problem**: Application can't reach Bedrock through VPC endpoint

**Solutions**:

1. **Verify private DNS is enabled**:
   ```bash
   # Check VPC endpoint configuration
   aws ec2 describe-vpc-endpoints \
     --filters "Name=service-name,Values=com.amazonaws.us-east-1.bedrock-agent-runtime"
   ```

2. **Check security group rules**:
   ```bash
   # Ensure security group allows HTTPS (443) outbound
   aws ec2 describe-security-groups --group-ids <sg-id>
   ```

3. **Test connectivity from application**:
   ```bash
   # From application instance
   curl -v https://bedrock-agent-runtime.us-east-1.amazonaws.com
   
   # Should resolve to private IP (10.x.x.x)
   ```

4. **Check route tables**:
   ```bash
   # Verify private subnets route to VPC endpoints
   aws ec2 describe-route-tables --filters "Name=vpc-id,Values=<vpc-id>"
   ```

### Terraform State Drift

**Problem**: Actual infrastructure doesn't match Terraform state

**Solutions**:

1. **Detect drift**:
   ```bash
   # Refresh state and show differences
   terraform plan -refresh-only
   ```

2. **Import manually created resources**:
   ```bash
   # Import existing resource into state
   terraform import aws_s3_bucket.example my-bucket-name
   ```

3. **Reconcile differences**:
   ```bash
   # Option 1: Update Terraform to match reality
   terraform apply -refresh-only
   
   # Option 2: Update infrastructure to match Terraform
   terraform apply
   ```

### Resource Already Exists Errors

**Problem**: `Resource already exists` when applying

**Solutions**:

1. **Import existing resource**:
   ```bash
   # Find the resource ID in AWS Console
   terraform import <resource_type>.<name> <resource_id>
   
   # Example:
   terraform import aws_s3_bucket.kb_docs bedrock-chat-poc-kb-docs-dev
   ```

2. **Use different resource names**:
   - Update `terraform.tfvars` with unique names
   - Ensure environment prefix is included

3. **Clean up orphaned resources**:
   ```bash
   # List resources not in Terraform state
   # Manually delete from AWS Console or CLI
   aws s3 rb s3://old-bucket-name --force
   ```

### Common Error Messages

| Error | Cause | Solution |
|-------|-------|----------|
| `NoSuchBucket` | State bucket doesn't exist | Create state bucket first |
| `AccessDenied` | Insufficient IAM permissions | Add required IAM policies |
| `ResourceNotFoundException` | Bedrock resource not found | Verify resource ID and region |
| `ValidationException` | Invalid parameter value | Check input validation rules |
| `ThrottlingException` | Too many API requests | Implement retry with backoff |
| `ConflictException` | Resource name already in use | Use unique resource names |
| `ServiceQuotaExceededException` | AWS service limit reached | Request quota increase |

### Getting Help

If you're still stuck:

1. **Enable Terraform debug logging**:
   ```bash
   export TF_LOG=DEBUG
   terraform apply
   ```

2. **Check AWS CloudTrail logs**:
   - Go to AWS Console → CloudTrail
   - Search for failed API calls
   - Review error messages and request parameters

3. **Review Terraform state**:
   ```bash
   terraform show
   terraform state list
   ```

4. **Consult documentation**:
   - Design document: `.kiro/specs/bedrock-infrastructure/design.md`
   - Requirements: `.kiro/specs/bedrock-infrastructure/requirements.md`
   - AWS Bedrock docs: https://docs.aws.amazon.com/bedrock/
   - Terraform AWS provider: https://registry.terraform.io/providers/hashicorp/aws/latest/docs

## Security Best Practices

- Never commit `terraform.tfvars` files with sensitive data
- Use IAM roles instead of access keys when possible
- Enable MFA for production deployments
- Review all changes with `terraform plan` before applying
- Use separate AWS accounts for dev/staging/prod when possible
- Regularly rotate AWS credentials
- Enable CloudTrail logging for audit trails

## Cost Optimization

- Destroy dev/staging environments when not in use
- Use single NAT gateway in non-production environments
- Monitor Bedrock API usage and costs
- Set up AWS Budgets alerts
- Review and clean up unused resources regularly

## Next Steps

1. Complete the remaining modules (IAM, Bedrock Agent, Knowledge Base, VPC)
2. Create environment-specific tfvars files
3. Deploy to development environment first
4. Test and validate infrastructure
5. Deploy to staging and production

## Support

For issues or questions:
- Review the design document: `.kiro/specs/bedrock-infrastructure/design.md`
- Check requirements: `.kiro/specs/bedrock-infrastructure/requirements.md`
- Consult AWS Bedrock documentation
- Review Terraform AWS provider documentation
