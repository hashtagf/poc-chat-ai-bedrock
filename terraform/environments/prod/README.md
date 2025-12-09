# Production Environment

This directory contains the Terraform configuration for the **production** environment of the Bedrock Chat POC infrastructure.

## Key Differences from Dev/Staging

The production environment includes additional security and high-availability features:

1. **VPC Deployment**: Enabled by default with private subnets and VPC endpoints
2. **Multi-AZ Configuration**: Resources deployed across multiple availability zones
3. **Multiple NAT Gateways**: One NAT gateway per AZ for high availability (no single point of failure)
4. **Longer Session Timeout**: 3600 seconds (1 hour) vs 1800 seconds in dev
5. **Stricter Security**: VPC endpoints ensure traffic stays within AWS network

## VPC Architecture

The production environment includes:

- **VPC**: 10.0.0.0/16 CIDR block
- **Private Subnets**: 10.0.1.0/24, 10.0.2.0/24 (across ap-southeast-1a, ap-southeast-1b)
- **Public Subnets**: 10.0.101.0/24, 10.0.102.0/24 (for NAT gateways)
- **VPC Endpoints**:
  - Bedrock Agent Runtime (interface endpoint with private DNS)
  - S3 (gateway endpoint)
- **NAT Gateways**: One per availability zone for high availability

## Prerequisites

1. **AWS Credentials**: Configure AWS credentials with appropriate permissions
2. **State Backend**: Deploy the state-backend module first (see `terraform/modules/state-backend/README.md`)
3. **Backend Configuration**: Update `backend.tf` with your state bucket name
4. **Terraform Version**: >= 1.5.0

## Deployment Steps

### 1. Initialize Terraform

```bash
cd terraform/environments/prod
terraform init
```

This will:
- Download required providers
- Initialize modules
- Configure remote state backend

### 2. Review Configuration

Review the `terraform.tfvars` file and adjust values as needed:

```bash
cat terraform.tfvars
```

Key variables to review:
- `aws_region`: AWS region for deployment
- `foundation_model`: Bedrock foundation model ID
- `embedding_model`: Embedding model for Knowledge Base
- `vpc_cidr`: VPC CIDR block (ensure no conflicts with existing VPCs)
- `availability_zones`: AZs for multi-AZ deployment

### 3. Plan Deployment

```bash
terraform plan
```

Review the plan output carefully. Expected resources:
- VPC with subnets, route tables, NAT gateways, VPC endpoints
- IAM roles and policies for Bedrock Agent and Knowledge Base
- S3 buckets for documents and vectors
- Bedrock Agent with agent alias
- Bedrock Knowledge Base with S3 data source

### 4. Apply Configuration

```bash
terraform apply
```

Type `yes` when prompted to confirm.

Deployment typically takes 5-10 minutes due to:
- VPC endpoint creation
- NAT gateway provisioning
- Bedrock Agent preparation

### 5. Retrieve Outputs

After successful deployment:

```bash
terraform output
```

Copy these values to your application's environment variables:

```bash
export BEDROCK_AGENT_ID=$(terraform output -raw bedrock_agent_id)
export BEDROCK_AGENT_ALIAS_ID=$(terraform output -raw bedrock_agent_alias_id)
export BEDROCK_KNOWLEDGE_BASE_ID=$(terraform output -raw bedrock_knowledge_base_id)
export AWS_REGION=$(terraform output -raw aws_region)
```

## Application Deployment in VPC

When deploying your application in the production VPC:

1. **Deploy in Private Subnets**: Use the `private_subnet_ids` output
2. **Security Group Configuration**: Allow outbound HTTPS to VPC endpoint security group
3. **No Code Changes**: AWS SDK automatically uses VPC endpoints with private DNS enabled
4. **Verify Connectivity**: Test Bedrock API calls from within the VPC

Example security group rule for application:

```hcl
resource "aws_security_group_rule" "app_to_vpc_endpoints" {
  type                     = "egress"
  from_port                = 443
  to_port                  = 443
  protocol                 = "tcp"
  source_security_group_id = module.vpc.security_group_id
  security_group_id        = aws_security_group.app.id
}
```

## Knowledge Base Document Upload

Upload documents to the S3 bucket:

```bash
# Get bucket name
BUCKET_NAME=$(terraform output -raw s3_bucket_name)

# Upload documents
aws s3 cp ./documents/ s3://${BUCKET_NAME}/ --recursive

# Sync Knowledge Base (triggers ingestion)
aws bedrock-agent start-ingestion-job \
  --knowledge-base-id $(terraform output -raw bedrock_knowledge_base_id) \
  --data-source-id $(terraform output -raw data_source_id)
```

## Cost Considerations

Production environment costs include:

- **NAT Gateways**: ~$0.045/hour per gateway × 2 = ~$65/month
- **VPC Endpoints**: ~$0.01/hour per endpoint × 2 = ~$15/month
- **S3 Storage**: Variable based on document size
- **Bedrock API Calls**: Pay per request and token usage

To reduce costs in non-production testing:
- Set `single_nat_gateway = true` (not recommended for production)
- Use dev environment without VPC

## Updating Infrastructure

To update the infrastructure:

```bash
# Pull latest changes
git pull

# Review changes
terraform plan

# Apply updates
terraform apply
```

## Destroying Infrastructure

⚠️ **WARNING**: This will delete all resources including S3 buckets and data.

```bash
terraform destroy
```

Type `yes` when prompted.

Note: S3 buckets must be empty before destruction. If you have versioned objects, you may need to manually empty buckets first.

## Troubleshooting

### State Locking Issues

If you encounter state locking errors:

```bash
# S3 native locking expires automatically after timeout
# Wait a few minutes and retry
terraform apply
```

### VPC Endpoint Connection Issues

If application cannot reach Bedrock:

1. Verify private DNS is enabled on VPC endpoint
2. Check security group rules allow HTTPS outbound
3. Verify application is in private subnet
4. Check VPC Flow Logs for blocked traffic

### Agent Preparation Failures

If agent preparation fails:

```bash
# Check agent status
aws bedrock-agent get-agent --agent-id <agent-id>

# Manually prepare agent
aws bedrock-agent prepare-agent --agent-id <agent-id>
```

## Security Best Practices

1. **IAM Permissions**: Use least privilege IAM roles
2. **S3 Bucket Policies**: Restrict access to specific IAM roles
3. **VPC Endpoints**: Keep traffic within AWS network
4. **Encryption**: All data encrypted at rest (S3, state bucket)
5. **Network Isolation**: Deploy application in private subnets
6. **Monitoring**: Enable VPC Flow Logs and CloudTrail

## Support

For issues or questions:
1. Check `terraform/README.md` for general guidance
2. Review module READMEs in `terraform/modules/`
3. Consult AWS Bedrock documentation
4. Check CloudWatch Logs for runtime errors
