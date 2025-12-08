# Design Document: Bedrock Infrastructure with Terraform

## Overview

This design document outlines the Terraform infrastructure for provisioning Amazon Bedrock Agent and Knowledge Base resources. The infrastructure follows AWS best practices for security, scalability, and maintainability. It supports multiple environments (dev, staging, production) with environment-specific configurations and includes VPC setup for private deployment in production.

The infrastructure is designed to integrate seamlessly with the existing Go backend application, providing the necessary Bedrock resources through Terraform outputs that map directly to application environment variables.

### Requirements Coverage

This design addresses all requirements from the requirements document:

- **Requirement 1** (Bedrock Agent Provisioning): Addressed by Bedrock Agent Module (Section 3.1)
- **Requirement 2** (Knowledge Base with S3): Addressed by Knowledge Base Module (Section 3.2)
- **Requirement 3** (IAM Configuration): Addressed by IAM Module (Section 3.3)
- **Requirement 4** (State Management): Addressed by State Backend Module (Section 3.5)
- **Requirement 5** (Resource Tagging): Addressed by universal tagging in all modules (Section 4.2)
- **Requirement 6** (Terraform Outputs): Addressed by root module outputs (Section 4.3)
- **Requirement 7** (Environment Configuration): Addressed by environment-specific tfvars (Section 4.2)
- **Requirement 8** (Module Structure): Addressed by modular architecture (Section 2.2)
- **Requirement 9** (VPC Configuration): Addressed by VPC Module (Section 3.4)

## Architecture

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     VPC (Production Only)                    │
│  ┌────────────────────────────────────────────────────────┐ │
│  │  Private Subnets (Multi-AZ)                            │ │
│  │  ┌──────────────┐  ┌──────────────┐                   │ │
│  │  │ VPC Endpoint │  │ VPC Endpoint │                   │ │
│  │  │   Bedrock    │  │      S3      │                   │ │
│  │  └──────────────┘  └──────────────┘                   │ │
│  └────────────────────────────────────────────────────────┘ │
│  ┌────────────────────────────────────────────────────────┐ │
│  │  Public Subnets (Multi-AZ)                             │ │
│  │  ┌──────────────┐  ┌──────────────┐                   │ │
│  │  │ NAT Gateway  │  │ NAT Gateway  │                   │ │
│  │  └──────────────┘  └──────────────┘                   │ │
│  └────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                    Bedrock Infrastructure                    │
│  ┌────────────────────────────────────────────────────────┐ │
│  │  Bedrock Agent                                         │ │
│  │  - Foundation Model: Claude/Titan                      │ │
│  │  - Agent Alias: DRAFT                                  │ │
│  │  - IAM Role: Agent Execution Role                      │ │
│  └────────────────────────────────────────────────────────┘ │
│  ┌────────────────────────────────────────────────────────┐ │
│  │  Knowledge Base                                        │ │
│  │  - Embedding Model: Titan Embeddings                   │ │
│  │  - Vector Store: S3                                    │ │
│  │  - Data Source: S3 Bucket                              │ │
│  │  - IAM Role: Knowledge Base Execution Role             │ │
│  └────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                    Storage & Vector Store                    │
│  ┌────────────────────────────────────────────────────────┐ │
│  │  S3 Bucket (Knowledge Base Documents)                  │ │
│  │  - Versioning: Enabled                                 │ │
│  │  - Encryption: AES-256                                 │ │
│  │  - Lifecycle: Optional archival                        │ │
│  └────────────────────────────────────────────────────────┘ │
│  ┌────────────────────────────────────────────────────────┐ │
│  │  S3 Bucket (Vector Storage)                            │ │
│  │  - Versioning: Enabled                                 │ │
│  │  - Encryption: AES-256                                 │ │
│  │  - Purpose: Store vector embeddings                    │ │
│  └────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                    Terraform State Backend                   │
│  ┌────────────────────────────────────────────────────────┐ │
│  │  S3 Bucket (Terraform State)                           │ │
│  │  - Versioning: Enabled                                 │ │
│  │  - Encryption: AES-256                                 │ │
│  │  - Native Locking: Enabled                             │ │
│  │  - Bucket Policy: Restrict access                      │ │
│  └────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

### Module Structure

The Terraform code is organized into reusable modules:

```
terraform/
├── modules/
│   ├── bedrock-agent/          # Bedrock Agent configuration
│   ├── knowledge-base/          # Knowledge Base with S3 vector store
│   ├── iam/                     # IAM roles and policies
│   ├── vpc/                     # VPC with endpoints (production)
│   └── state-backend/           # S3 for remote state
├── environments/
│   ├── dev/
│   │   ├── main.tf
│   │   ├── variables.tf
│   │   ├── outputs.tf
│   │   └── terraform.tfvars
│   ├── staging/
│   │   └── terraform.tfvars
│   └── prod/
│       └── terraform.tfvars
└── backend.tf                   # Remote state configuration
```

## Components and Interfaces

### 1. Bedrock Agent Module

**Purpose**: Provisions a Bedrock Agent with specified foundation model and creates an agent alias.

**Inputs**:
- `agent_name` (string): Name of the Bedrock Agent
- `foundation_model` (string): Foundation model ID (e.g., "anthropic.claude-v2")
- `agent_instruction` (string): Instructions for the agent behavior
- `agent_role_arn` (string): ARN of the IAM role for agent execution
- `idle_session_ttl` (number): Session timeout in seconds (default: 1800)
- `tags` (map): Resource tags

**Outputs**:
- `agent_id` (string): The Bedrock Agent ID
- `agent_arn` (string): The Bedrock Agent ARN
- `agent_alias_id` (string): The Agent Alias ID for DRAFT version
- `agent_alias_arn` (string): The Agent Alias ARN

**Resources**:
- `aws_bedrockagent_agent`: The Bedrock Agent resource
- `aws_bedrockagent_agent_alias`: Agent alias for versioning
- `terraform_data`: Resource for agent preparation with local-exec provisioner
- `time_sleep`: Delay resource for agent initialization

### 2. Knowledge Base Module

**Purpose**: Provisions a Knowledge Base with S3 data source and S3 vector store.

**Inputs**:
- `knowledge_base_name` (string): Name of the Knowledge Base
- `embedding_model` (string): Embedding model ID (e.g., "amazon.titan-embed-text-v1")
- `kb_role_arn` (string): ARN of the IAM role for Knowledge Base
- `s3_bucket_name` (string): Name of the S3 bucket for documents
- `s3_vector_bucket_name` (string): Name of the S3 bucket for vector storage
- `tags` (map): Resource tags

**Outputs**:
- `knowledge_base_id` (string): The Knowledge Base ID
- `knowledge_base_arn` (string): The Knowledge Base ARN
- `s3_bucket_name` (string): The S3 bucket name for document uploads
- `s3_bucket_arn` (string): The S3 bucket ARN
- `s3_vector_bucket_name` (string): The S3 bucket name for vector storage
- `s3_vector_bucket_arn` (string): The S3 bucket ARN for vector storage
- `data_source_id` (string): The data source ID

**Resources**:
- `aws_s3_bucket`: S3 bucket for knowledge base documents
- `aws_s3_bucket_versioning`: Enable versioning on documents bucket
- `aws_s3_bucket_server_side_encryption_configuration`: Enable encryption on documents bucket
- `aws_s3_bucket_public_access_block`: Block public access to documents bucket
- `aws_s3_bucket`: S3 bucket for vector storage
- `aws_s3_bucket_versioning`: Enable versioning on vector bucket
- `aws_s3_bucket_server_side_encryption_configuration`: Enable encryption on vector bucket
- `aws_s3_bucket_public_access_block`: Block public access to vector bucket
- `aws_bedrockagent_knowledge_base`: The Knowledge Base resource with S3 vector configuration
- `aws_bedrockagent_data_source`: S3 data source configuration

### 3. IAM Module

**Purpose**: Creates IAM roles and policies for Bedrock Agent and Knowledge Base with least privilege permissions.

**Inputs**:
- `project_name` (string): Project name for resource naming
- `environment` (string): Environment name (dev/staging/prod)
- `foundation_model_arn` (string): ARN of the foundation model
- `embedding_model_arn` (string): ARN of the embedding model
- `s3_bucket_arn` (string): ARN of the S3 documents bucket
- `s3_vector_bucket_arn` (string): ARN of the S3 vector bucket
- `tags` (map): Resource tags

**Outputs**:
- `agent_role_arn` (string): ARN of the Bedrock Agent execution role
- `agent_role_name` (string): Name of the Bedrock Agent execution role
- `kb_role_arn` (string): ARN of the Knowledge Base execution role
- `kb_role_name` (string): Name of the Knowledge Base execution role

**Resources**:
- `aws_iam_role`: Agent execution role with trust policy
- `aws_iam_role_policy`: Agent policy for foundation model invocation
- `aws_iam_role`: Knowledge Base execution role with trust policy
- `aws_iam_role_policy`: Knowledge Base policy for S3 (documents and vectors) and embedding model access

**IAM Policies**:

All IAM policies follow the principle of least privilege by:
- Restricting actions to only those required for the specific service
- Limiting resource access to specific ARNs (no wildcards where possible)
- Using condition keys to restrict access by source account and ARN
- Separating roles by function (Agent vs Knowledge Base)

Agent Role Trust Policy:
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "bedrock.amazonaws.com"
      },
      "Action": "sts:AssumeRole",
      "Condition": {
        "StringEquals": {
          "aws:SourceAccount": "${account_id}"
        },
        "ArnLike": {
          "aws:SourceArn": "arn:aws:bedrock:${region}:${account_id}:agent/*"
        }
      }
    }
  ]
}
```

Agent Role Permissions:
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "bedrock:InvokeModel"
      ],
      "Resource": "${foundation_model_arn}"
    }
  ]
}
```

Knowledge Base Role Trust Policy:
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "bedrock.amazonaws.com"
      },
      "Action": "sts:AssumeRole",
      "Condition": {
        "StringEquals": {
          "aws:SourceAccount": "${account_id}"
        },
        "ArnLike": {
          "aws:SourceArn": "arn:aws:bedrock:${region}:${account_id}:knowledge-base/*"
        }
      }
    }
  ]
}
```

Knowledge Base Role Permissions:
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:GetObject",
        "s3:ListBucket"
      ],
      "Resource": [
        "${s3_bucket_arn}",
        "${s3_bucket_arn}/*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "s3:PutObject",
        "s3:GetObject",
        "s3:ListBucket"
      ],
      "Resource": [
        "${s3_vector_bucket_arn}",
        "${s3_vector_bucket_arn}/*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "bedrock:InvokeModel"
      ],
      "Resource": "${embedding_model_arn}"
    }
  ]
}
```

### 4. VPC Module (Production Only)

**Purpose**: Creates VPC infrastructure with private subnets and VPC endpoints for secure Bedrock access.

**Inputs**:
- `vpc_name` (string): Name of the VPC
- `vpc_cidr` (string): CIDR block for VPC (e.g., "10.0.0.0/16")
- `availability_zones` (list): List of AZs (e.g., ["us-east-1a", "us-east-1b"])
- `private_subnet_cidrs` (list): CIDR blocks for private subnets
- `public_subnet_cidrs` (list): CIDR blocks for public subnets
- `enable_nat_gateway` (bool): Enable NAT gateway for internet access (default: true)
- `single_nat_gateway` (bool): Use single NAT gateway for cost savings (default: false for prod)
- `tags` (map): Resource tags

**Outputs**:
- `vpc_id` (string): The VPC ID
- `private_subnet_ids` (list): List of private subnet IDs
- `public_subnet_ids` (list): List of public subnet IDs
- `bedrock_vpc_endpoint_id` (string): VPC endpoint ID for Bedrock
- `s3_vpc_endpoint_id` (string): VPC endpoint ID for S3
- `security_group_id` (string): Security group ID for VPC endpoints

**Resources**:
- `aws_vpc`: VPC with DNS support enabled
- `aws_subnet`: Private subnets across multiple AZs
- `aws_subnet`: Public subnets across multiple AZs
- `aws_internet_gateway`: Internet gateway for public subnets
- `aws_eip`: Elastic IPs for NAT gateways
- `aws_nat_gateway`: NAT gateways in public subnets
- `aws_route_table`: Route tables for private and public subnets
- `aws_route_table_association`: Associate subnets with route tables
- `aws_security_group`: Security group for VPC endpoints
- `aws_vpc_endpoint`: VPC endpoint for Bedrock Agent Runtime
- `aws_vpc_endpoint`: VPC endpoint for S3 (Gateway type)

**VPC Endpoint Configuration**:

Bedrock Agent Runtime Endpoint:
- Service: `com.amazonaws.${region}.bedrock-agent-runtime`
- Type: Interface
- Private DNS: Enabled
- Security Group: Allow HTTPS (443) outbound

S3 Endpoint:
- Service: `com.amazonaws.${region}.s3`
- Type: Gateway
- Route Table: Private subnet route tables

**Application Integration with VPC Endpoints**:

When VPC is enabled, the application must be configured to use VPC endpoints:

1. **Deploy Application in Private Subnets**: The Go backend should run in the same VPC's private subnets
2. **Use Private DNS**: With private DNS enabled, the application uses standard AWS SDK endpoints (e.g., `bedrock-agent-runtime.us-east-1.amazonaws.com`) which automatically resolve to the VPC endpoint
3. **Security Group Configuration**: Application security group must allow outbound HTTPS to VPC endpoint security group
4. **No Code Changes Required**: AWS SDK automatically uses VPC endpoints when private DNS is enabled

**Environment Variable Configuration**:
```bash
# No special endpoint configuration needed when using VPC endpoints with private DNS
AWS_REGION=us-east-1
BEDROCK_AGENT_ID=<from terraform output>
BEDROCK_AGENT_ALIAS_ID=<from terraform output>
# SDK automatically routes through VPC endpoints
```

**Validation**:
- Test connectivity from application to Bedrock using VPC endpoint
- Verify traffic stays within VPC (no internet gateway usage)
- Check VPC Flow Logs to confirm traffic routing through endpoints

### 5. State Backend Module

**Purpose**: Creates S3 bucket for Terraform remote state management with native S3 locking.

**Why Remote State?**

Terraform stores the current state of your infrastructure in a state file. By default, this is stored locally, but for team collaboration and safety, we use remote state:

**S3 for State Storage**:
- **Centralized**: All team members access the same state file
- **Versioned**: S3 versioning allows rollback to previous states
- **Encrypted**: State files may contain sensitive data (IDs, ARNs)
- **Durable**: S3 provides 99.999999999% durability

**Why S3 Native Locking?**

When multiple people or CI/CD pipelines run Terraform simultaneously, they could corrupt the state file. S3 provides native state locking capabilities without requiring DynamoDB:

**State Locking Mechanism**:
1. Before running `terraform apply`, Terraform acquires a lock using S3 object locking
2. The lock is implemented using S3's conditional write operations
3. Other Terraform operations wait until the lock is released
4. After `terraform apply` completes, the lock is released
5. If Terraform crashes, the lock expires automatically after a timeout

**Benefits**:
- **Prevents Concurrent Modifications**: Only one person can modify infrastructure at a time
- **Prevents State Corruption**: Ensures state file integrity
- **Simplified Infrastructure**: No additional DynamoDB table required
- **Cost-Effective**: Only S3 storage costs, no additional service charges
- **Automatic Lock Expiration**: Stale locks expire automatically

**Example Scenario Without Locking**:
```
Time    Developer A              Developer B              State File
10:00   terraform apply          -                        Version 1
10:01   Reading state v1...      terraform apply          Version 1
10:02   Creating resource X...   Reading state v1...      Version 1
10:03   Writing state v2...      Creating resource Y...   Version 2 (has X)
10:04   Done                     Writing state v2...      Version 2 (has Y, missing X!)
```
Result: Resource X is created but not tracked in state → orphaned resource

**With S3 Native Locking**:
```
Time    Developer A              Developer B              Lock Status
10:00   terraform apply          -                        A acquires lock
10:01   Creating resources...    terraform apply          B waits for lock
10:02   Writing state...         Still waiting...         A holds lock
10:03   Done, release lock       Lock acquired!           B acquires lock
10:04   -                        Creating resources...    B holds lock
```
Result: Sequential execution, no conflicts

**Inputs**:
- `state_bucket_name` (string): Name of the S3 bucket for state
- `tags` (map): Resource tags

**Outputs**:
- `state_bucket_name` (string): The state bucket name
- `state_bucket_arn` (string): The state bucket ARN

**Resources**:
- `aws_s3_bucket`: S3 bucket for Terraform state
- `aws_s3_bucket_versioning`: Enable versioning
- `aws_s3_bucket_server_side_encryption_configuration`: Enable encryption
- `aws_s3_bucket_public_access_block`: Block public access

## Data Models

### Environment Configuration (terraform.tfvars)

**Development Environment**:
```hcl
# General
environment     = "dev"
project_name    = "bedrock-chat-poc"
aws_region      = "us-east-1"

# Bedrock Agent
agent_name           = "bedrock-chat-poc-agent-dev"
foundation_model     = "anthropic.claude-v2"
agent_instruction    = "You are a helpful AI assistant for the chat POC application."
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

**Production Environment**:
```hcl
# General
environment     = "prod"
project_name    = "bedrock-chat-poc"
aws_region      = "us-east-1"

# Bedrock Agent
agent_name           = "bedrock-chat-poc-agent-prod"
foundation_model     = "anthropic.claude-v2"
agent_instruction    = "You are a helpful AI assistant for the chat POC application."
idle_session_ttl     = 3600

# Knowledge Base
knowledge_base_name         = "bedrock-chat-poc-kb-prod"
embedding_model             = "amazon.titan-embed-text-v1"
s3_bucket_name              = "bedrock-chat-poc-kb-docs-prod"
s3_vector_bucket_name       = "bedrock-chat-poc-kb-vectors-prod"

# VPC (enabled for prod)
enable_vpc              = true
vpc_cidr                = "10.0.0.0/16"
availability_zones      = ["us-east-1a", "us-east-1b"]
private_subnet_cidrs    = ["10.0.1.0/24", "10.0.2.0/24"]
public_subnet_cidrs     = ["10.0.101.0/24", "10.0.102.0/24"]
enable_nat_gateway      = true
single_nat_gateway      = false

# Tags
tags = {
  Environment = "prod"
  Project     = "bedrock-chat-poc"
  ManagedBy   = "Terraform"
  CreatedAt   = "2025-12-07"
}
```

### Terraform Outputs

The root module outputs match the backend application's environment variables:

```hcl
output "bedrock_agent_id" {
  description = "Bedrock Agent ID for BEDROCK_AGENT_ID env var"
  value       = module.bedrock_agent.agent_id
}

output "bedrock_agent_alias_id" {
  description = "Bedrock Agent Alias ID for BEDROCK_AGENT_ALIAS_ID env var"
  value       = module.bedrock_agent.agent_alias_id
}

output "bedrock_knowledge_base_id" {
  description = "Knowledge Base ID for BEDROCK_KNOWLEDGE_BASE_ID env var"
  value       = module.knowledge_base.knowledge_base_id
}

output "s3_bucket_name" {
  description = "S3 bucket name for uploading knowledge base documents"
  value       = module.knowledge_base.s3_bucket_name
}

output "aws_region" {
  description = "AWS region for AWS_REGION env var"
  value       = var.aws_region
}

output "vpc_id" {
  description = "VPC ID (production only)"
  value       = var.enable_vpc ? module.vpc[0].vpc_id : null
}
```

## 
Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system—essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*

Since this is infrastructure-as-code, most acceptance criteria are validated through infrastructure testing rather than traditional property-based testing. However, we can define properties for aspects that should hold universally across all resources.

### Property 1: Universal Resource Tagging

*For any* AWS resource created by this Terraform configuration, that resource should have all four required tags: Environment, Project, ManagedBy, and CreatedAt with appropriate values.

**Validates: Requirements 5.1, 5.2, 5.3, 5.4**

### Property 2: IAM Least Privilege Compliance

*For any* IAM policy created by this Terraform configuration, that policy should grant only the minimum permissions required for the service to function, with no wildcard resource ARNs except where AWS services require them.

**Validates: Requirements 3.5**

### Infrastructure Validation Examples

The following acceptance criteria are validated through infrastructure testing (e.g., Terratest, AWS API validation) rather than property-based testing:

**Agent Provisioning** (Requirements 1.1-1.5):
- Verify Bedrock Agent is created with specified foundation model
- Verify IAM role is assigned with correct permissions
- Verify Agent Alias is created for DRAFT version
- Verify Terraform outputs include agent_id and agent_alias_id
- Verify environment-specific tfvars produce different configurations

**Knowledge Base Provisioning** (Requirements 2.1-2.6):
- Verify S3 bucket is created
- Verify S3 documents bucket has versioning and encryption enabled
- Verify S3 vector bucket has versioning and encryption enabled
- Verify Knowledge Base is configured with embedding model
- Verify Knowledge Base is configured with S3 vector store
- Verify S3 data source is configured
- Verify Terraform outputs include knowledge_base_id

**IAM Configuration** (Requirements 3.1-3.5):
- Verify Agent IAM role has foundation model invocation permissions
- Verify Knowledge Base IAM role has S3 documents bucket read permissions
- Verify Knowledge Base IAM role has S3 vector bucket read/write permissions
- Verify Knowledge Base IAM role has embedding model invocation permissions
- Verify all IAM policies follow least privilege principle (validated by Property 2)

**State Management** (Requirements 4.1-4.4):
- Verify Terraform state is stored in S3
- Verify S3 native locking is enabled
- Verify state bucket has encryption enabled
- Verify state bucket has versioning enabled

**Terraform Outputs** (Requirements 6.1-6.5):
- Verify all required outputs are defined: bedrock_agent_id, bedrock_agent_alias_id, bedrock_knowledge_base_id, s3_bucket_name, aws_region

**Environment Configuration** (Requirements 7.1-7.4):
- Verify separate tfvars files exist for dev, staging, prod
- Verify terraform apply accepts -var-file parameter
- Verify tfvars include all required variables
- Verify production configuration has stricter security settings

**Module Structure** (Requirements 8.1-8.4):
- Verify separate modules exist for agent, knowledge-base, iam
- Verify each module has variables.tf and outputs.tf
- Verify module interfaces are well-defined

**VPC Configuration** (Requirements 9.1-9.7):
- Verify private subnets are created across multiple AZs
- Verify public subnets are created
- Verify VPC endpoint for Bedrock Agent Runtime is created
- Verify VPC endpoint for S3 is created
- Verify security groups allow HTTPS egress
- Verify application can use VPC endpoints with private DNS
- Verify VPC is mandatory when environment is prod

## Error Handling

### Terraform Validation Errors

**Invalid Configuration**:
- Use `terraform validate` to catch syntax and configuration errors before apply
- Implement variable validation rules for required formats (e.g., CIDR blocks, model IDs)
- Provide clear error messages for missing required variables

**Resource Creation Failures**:
- Bedrock Agent preparation failures: Implement retry logic with `terraform_data` and `time_sleep` resources
- IAM permission errors: Validate IAM policies before resource creation
- S3 bucket name conflicts: Use unique naming with environment prefix
- Knowledge Base sync failures: Verify S3 vector bucket permissions

**State Management Errors**:
- State locking conflicts: S3 native locking handles this automatically
- State corruption: Enable S3 versioning for state bucket rollback
- Concurrent modifications: Use workspace isolation or separate state files per environment
- Stale locks: S3 native locking expires automatically, no manual intervention needed

### AWS Service Limits

**Quota Exceeded**:
- Bedrock Agent limits: Check service quotas before deployment
- S3 bucket limits: Verify bucket count within account limits
- VPC endpoint limits: Verify endpoint quotas in target region

**Region Availability**:
- Bedrock service availability: Validate region supports Bedrock Agent and Knowledge Base
- Foundation model availability: Check model availability in target region
- S3 vector store support: Verify region supports S3 as Knowledge Base vector store

### Security Errors

**IAM Permission Denied**:
- Insufficient Terraform execution permissions: Ensure deploying user/role has necessary IAM permissions
- Cross-account access issues: Verify trust relationships and resource policies
- KMS key access: Ensure proper KMS key policies for encryption

**VPC Endpoint Access**:
- Security group misconfiguration: Verify security group rules allow required traffic
- Network ACL restrictions: Check network ACLs don't block VPC endpoint traffic
- DNS resolution issues: Ensure private DNS is enabled for interface endpoints

### Rollback Strategy

**Failed Deployment**:
- Use `terraform destroy` to clean up partially created resources
- Leverage S3 state versioning to rollback to previous working state
- Implement `terraform plan` before every apply to preview changes

**Resource Dependencies**:
- Use explicit `depends_on` to ensure proper creation order
- Implement `lifecycle` rules to prevent accidental resource deletion
- Use `prevent_destroy` for critical resources in production

## Testing Strategy

### Infrastructure Testing with Terratest

**Unit Tests**:
- Test individual modules in isolation
- Verify module outputs match expected values
- Test variable validation rules
- Mock AWS API calls for faster testing

**Integration Tests**:
- Deploy full infrastructure to test AWS account
- Verify resources are created with correct configurations
- Test IAM permissions by attempting operations
- Validate VPC connectivity and endpoint functionality
- Test Knowledge Base ingestion with sample documents
- Verify Agent can be invoked through AWS SDK

**Test Structure**:
```go
func TestBedrockAgentModule(t *testing.T) {
    terraformOptions := &terraform.Options{
        TerraformDir: "../modules/bedrock-agent",
        Vars: map[string]interface{}{
            "agent_name": "test-agent",
            "foundation_model": "anthropic.claude-v2",
            // ... other variables
        },
    }
    
    defer terraform.Destroy(t, terraformOptions)
    terraform.InitAndApply(t, terraformOptions)
    
    // Verify outputs
    agentID := terraform.Output(t, terraformOptions, "agent_id")
    assert.NotEmpty(t, agentID)
    
    // Verify resource exists in AWS
    // ... AWS SDK calls to verify
}
```

**Property-Based Tests**:

Test that all resources have required tags:
```go
func TestAllResourcesHaveRequiredTags(t *testing.T) {
    // This test validates Property 1: Universal Resource Tagging
    // **Feature: bedrock-infrastructure, Property 1: Universal Resource Tagging**
    
    terraformOptions := &terraform.Options{
        TerraformDir: "../environments/dev",
    }
    
    defer terraform.Destroy(t, terraformOptions)
    terraform.InitAndApply(t, terraformOptions)
    
    // Get all resources from Terraform state
    resources := terraform.Show(t, terraformOptions)
    
    requiredTags := []string{"Environment", "Project", "ManagedBy", "CreatedAt"}
    
    // For each resource, verify it has all required tags
    for _, resource := range resources {
        if resource.Type != "data" { // Skip data sources
            tags := getResourceTags(resource)
            for _, requiredTag := range requiredTags {
                assert.Contains(t, tags, requiredTag, 
                    "Resource %s missing required tag: %s", 
                    resource.Address, requiredTag)
            }
        }
    }
}
```

Test that all IAM policies follow least privilege:
```go
func TestIAMPoliciesFollowLeastPrivilege(t *testing.T) {
    // This test validates Property 2: IAM Least Privilege Compliance
    // **Feature: bedrock-infrastructure, Property 2: IAM Least Privilege Compliance**
    
    terraformOptions := &terraform.Options{
        TerraformDir: "../environments/dev",
    }
    
    defer terraform.Destroy(t, terraformOptions)
    terraform.InitAndApply(t, terraformOptions)
    
    // Get IAM role policies from Terraform state
    agentRoleName := terraform.Output(t, terraformOptions, "agent_role_name")
    kbRoleName := terraform.Output(t, terraformOptions, "kb_role_name")
    
    // Verify Agent role policy
    agentPolicy := getIAMRolePolicy(t, agentRoleName)
    assert.NotContains(t, agentPolicy, "\"Resource\": \"*\"", 
        "Agent IAM policy should not use wildcard resources")
    assert.Contains(t, agentPolicy, "bedrock:InvokeModel",
        "Agent IAM policy should only allow InvokeModel action")
    
    // Verify Knowledge Base role policy
    kbPolicy := getIAMRolePolicy(t, kbRoleName)
    assert.NotContains(t, kbPolicy, "\"Action\": \"*\"",
        "KB IAM policy should not use wildcard actions")
    
    // Verify specific permissions are scoped
    verifyPolicyHasSpecificResourceARN(t, kbPolicy, "s3:GetObject")
    verifyPolicyHasSpecificResourceARN(t, kbPolicy, "s3:PutObject")
    verifyPolicyHasSpecificResourceARN(t, kbPolicy, "bedrock:InvokeModel")
}
```

**Test Execution**:
- Run unit tests on every commit
- Run integration tests on pull requests
- Run full infrastructure tests in isolated AWS account
- Clean up test resources after test completion
- Use test fixtures for consistent test data

**Test Coverage**:
- Module interface contracts (inputs/outputs)
- Resource creation and configuration
- IAM permission validation
- VPC connectivity
- State management
- Multi-environment deployment
- Resource tagging compliance

### Manual Validation

**Post-Deployment Checks**:
1. Verify Bedrock Agent can be invoked through AWS Console
2. Upload test document to S3 and trigger Knowledge Base sync
3. Query Knowledge Base through Agent to verify RAG functionality
4. Test VPC endpoint connectivity from private subnet
5. Verify application can connect using Terraform outputs

**Security Validation**:
1. Review IAM policies for least privilege compliance
2. Verify encryption is enabled on all data at rest
3. Check VPC security groups and network ACLs
4. Validate no public access to S3 buckets
5. Review CloudTrail logs for API activity

## Design Decisions and Rationale

### Key Architectural Decisions

**1. Modular Terraform Structure**
- **Decision**: Organize infrastructure into reusable modules (bedrock-agent, knowledge-base, iam, vpc, state-backend)
- **Rationale**: Enables code reuse across environments, simplifies testing, and follows Terraform best practices. Each module has a single responsibility and clear interface.
- **Trade-offs**: Slightly more complex initial setup, but significantly easier to maintain and extend.

**2. Remote State with S3 Native Locking**
- **Decision**: Use S3 for state storage with native locking instead of DynamoDB
- **Rationale**: Simplifies infrastructure by eliminating DynamoDB table, reduces costs, and S3 native locking is sufficient for POC collaboration needs. Automatic lock expiration prevents stale lock issues.
- **Trade-offs**: Slightly newer feature (requires Terraform >= 1.5), but simpler and more cost-effective than S3 + DynamoDB approach.

**3. S3 for Vector Store**
- **Decision**: Use S3 as the vector store instead of OpenSearch Serverless or other vector databases
- **Rationale**: Simplest integration with Bedrock Knowledge Base, lowest cost for POC, no additional infrastructure to manage, and sufficient for development/testing workloads.
- **Trade-offs**: May have higher latency than dedicated vector databases, but acceptable for POC validation. Can migrate to OpenSearch Serverless or other vector stores for production if needed.

**4. VPC Optional for Development, Mandatory for Production**
- **Decision**: Make VPC deployment optional via `enable_vpc` variable, but enforce for production
- **Rationale**: Reduces cost and complexity for development/testing environments while ensuring production security requirements are met.
- **Trade-offs**: Different network architectures between environments, but acceptable for POC with clear production path.

**5. Environment-Specific tfvars Files**
- **Decision**: Separate tfvars files for each environment (dev, staging, prod) rather than workspaces
- **Rationale**: Explicit configuration per environment, easier to review changes, better for CI/CD pipelines, and clearer separation of concerns.
- **Trade-offs**: More files to maintain, but better visibility and control.

**6. Agent Preparation with local-exec Provisioner**
- **Decision**: Use Terraform `local-exec` provisioner to call AWS CLI for agent preparation
- **Rationale**: Bedrock Agents require a preparation step after creation that isn't natively supported by Terraform. This workaround ensures agents are ready for use.
- **Trade-offs**: Requires AWS CLI installed on deployment machine, but necessary until Terraform provider adds native support.

**7. Interface VPC Endpoints for Bedrock, Gateway for S3**
- **Decision**: Use interface endpoints for Bedrock Agent Runtime and gateway endpoints for S3
- **Rationale**: Interface endpoints required for Bedrock (no gateway option), gateway endpoints for S3 are free and sufficient for our use case.
- **Trade-offs**: Interface endpoints have hourly costs, but necessary for private Bedrock access.

**8. Terraform Outputs Match Application Environment Variables**
- **Decision**: Name Terraform outputs to directly correspond to backend application environment variables
- **Rationale**: Simplifies deployment workflow, reduces configuration errors, and makes integration obvious.
- **Trade-offs**: Couples infrastructure naming to application expectations, but acceptable for integrated POC.

## Implementation Notes

### Terraform Version Requirements

- Terraform >= 1.5.0
- AWS Provider >= 5.0.0
- Required provider features:
  - `aws_bedrockagent_agent` resource
  - `aws_bedrockagent_knowledge_base` resource with S3 vector store support
  - `aws_s3_bucket` resource

### Agent Preparation Workaround

Bedrock Agents require a preparation step after creation. Terraform doesn't natively support this, so we use a workaround:

```hcl
resource "terraform_data" "agent_prepare" {
  triggers_replace = [
    aws_bedrockagent_agent.this.id,
    aws_bedrockagent_agent.this.agent_version
  ]
  
  provisioner "local-exec" {
    command = <<-EOT
      aws bedrock-agent prepare-agent \
        --agent-id ${aws_bedrockagent_agent.this.id} \
        --region ${var.aws_region}
    EOT
  }
}

resource "time_sleep" "agent_prepare_wait" {
  depends_on = [terraform_data.agent_prepare]
  create_duration = "30s"
}
```

### S3 Vector Store Overview

**What is S3 Vector Store?**

Amazon Bedrock Knowledge Base supports using S3 as a vector store, which is the simplest and most cost-effective option for POC and development environments. Instead of requiring a dedicated vector database like OpenSearch Serverless, Bedrock can store and retrieve vector embeddings directly from S3.

**Key Characteristics**:
- **No Additional Infrastructure**: Uses standard S3 buckets, no specialized services required
- **Cost-Effective**: Only pay for S3 storage and API requests, no compute charges
- **Simple Setup**: Minimal configuration compared to vector databases
- **Suitable for POC**: Perfect for development, testing, and proof-of-concept workloads
- **Easy Migration Path**: Can migrate to OpenSearch Serverless or other vector stores later if needed

**How S3 Vector Store Works with Bedrock Knowledge Base**:

1. **Document Ingestion**: When you upload documents to the source S3 bucket, Bedrock processes them
2. **Embedding Generation**: Bedrock uses the specified embedding model to generate vector embeddings
3. **Vector Storage**: Embeddings are stored in the designated S3 vector bucket
4. **Semantic Search**: When users query, Bedrock retrieves relevant vectors from S3
5. **Context Retrieval**: Matching document chunks are returned to augment agent responses

**Configuration Requirements**:
- **Two S3 Buckets**: One for source documents, one for vector storage
- **IAM Permissions**: Knowledge Base role needs read access to documents bucket and read/write access to vector bucket
- **Encryption**: Both buckets should have encryption enabled
- **Versioning**: Enable versioning for data protection and rollback capability

**Cost Considerations**:
- **S3 Storage**: Pay only for data stored (documents + vectors)
- **S3 API Requests**: Charges for PUT/GET operations during ingestion and queries
- **No Compute Costs**: Unlike OpenSearch Serverless, no OCU charges
- **Typical POC Cost**: Very low, usually under $10/month for small document sets

**When to Use S3 vs OpenSearch Serverless**:

Use S3 Vector Store when:
- Building POC or development environments
- Working with small to medium document sets (< 10,000 documents)
- Cost optimization is a priority
- Simple setup is preferred
- Query latency requirements are moderate (< 1 second acceptable)

Consider OpenSearch Serverless when:
- Production workloads with high query volume
- Large document sets (> 10,000 documents)
- Sub-second query latency required
- Advanced vector search features needed
- High availability and redundancy critical

### Cost Optimization

**Development Environment**:
- Use single NAT gateway instead of multi-AZ
- Disable VPC endpoints (use public internet)
- Use S3 vector store for lowest cost
- Implement S3 lifecycle policies for old documents and vectors

**Production Environment**:
- Multi-AZ NAT gateways for high availability
- VPC endpoints to reduce data transfer costs
- Consider migrating to OpenSearch Serverless for better performance
- Monitor Bedrock API usage and optimize queries
- Enable S3 Intelligent-Tiering for cost optimization

### Backend Configuration

The backend configuration uses S3 with native locking. Here's the configuration format:

```hcl
terraform {
  backend "s3" {
    bucket = "your-terraform-state-bucket"
    key    = "bedrock-infrastructure/terraform.tfstate"
    region = "us-east-1"
    encrypt = true
    
    # S3 native locking is enabled automatically when using S3 backend
    # No DynamoDB table required
  }
}
```

**S3 Native Locking**:
- Terraform automatically uses S3's conditional write operations for locking
- No additional configuration needed beyond the S3 backend
- Locks are stored as metadata on the state object
- Automatic lock expiration prevents stale locks

### Deployment Workflow

1. **Bootstrap State Backend** (one-time):
   ```bash
   cd terraform/modules/state-backend
   terraform init
   terraform apply
   ```

2. **Configure Backend** (one-time):
   Update `backend.tf` with state bucket name

3. **Deploy Environment**:
   ```bash
   cd terraform/environments/dev
   terraform init
   terraform plan -var-file=terraform.tfvars
   terraform apply -var-file=terraform.tfvars
   ```

4. **Update Application Configuration**:
   ```bash
   terraform output -json > outputs.json
   # Update .env file with output values
   ```

5. **Upload Knowledge Base Documents**:
   ```bash
   aws s3 cp documents/ s3://$(terraform output -raw s3_bucket_name)/ --recursive
   ```

6. **Sync Knowledge Base**:
   ```bash
   aws bedrock-agent start-ingestion-job \
     --knowledge-base-id $(terraform output -raw bedrock_knowledge_base_id) \
     --data-source-id <data-source-id>
   ```

### Maintenance and Updates

**Updating Foundation Models**:
- Update `foundation_model` variable in tfvars
- Run `terraform plan` to preview changes
- Apply changes during maintenance window
- Test agent functionality after update

**Scaling Vector Storage**:
- S3 scales automatically with no configuration needed
- Monitor S3 storage metrics and request rates in CloudWatch
- Consider migrating to OpenSearch Serverless if query latency becomes an issue
- Implement S3 lifecycle policies to archive old vectors if needed

**Rotating IAM Credentials**:
- Use IAM roles instead of access keys
- Rotate service role policies as needed
- Audit IAM permissions regularly

**Disaster Recovery**:
- S3 bucket versioning enables document recovery
- Terraform state versioning enables infrastructure rollback
- Regular backups of Knowledge Base data
- Document recovery procedures

## Future Enhancements

**Monitoring and Observability**:
- CloudWatch dashboards for Bedrock metrics
- Alarms for API throttling and errors
- VPC Flow Logs for network troubleshooting
- Cost allocation tags for billing analysis

**Advanced Features**:
- Multi-region deployment for disaster recovery
- Blue-green deployment strategy for zero-downtime updates
- Automated Knowledge Base sync on S3 events
- Custom Lambda functions for action groups
- Guardrails for content filtering

**Security Enhancements**:
- AWS WAF for API protection
- Secrets Manager for sensitive configuration
- KMS customer-managed keys for encryption
- VPC endpoint policies for fine-grained access control
- AWS Config rules for compliance monitoring

**CI/CD Integration**:
- Automated Terraform validation in CI pipeline
- Terratest execution on pull requests
- Automated deployment to dev environment
- Manual approval for production deployment
- Drift detection and remediation
