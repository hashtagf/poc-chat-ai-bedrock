# State Backend Module

This module creates an S3 bucket for storing Terraform remote state with versioning, encryption, and native locking capabilities.

## Purpose

The state backend module must be deployed **first** before any other infrastructure, as it creates the S3 bucket that will store the Terraform state for all other modules and environments.

## Features

- **S3 Bucket**: Stores Terraform state files
- **Versioning**: Enabled for state rollback capability
- **Encryption**: AES-256 server-side encryption at rest
- **Native Locking**: S3 conditional writes provide automatic state locking
- **Public Access Block**: Prevents accidental public exposure
- **SSL Enforcement**: Requires HTTPS for all bucket operations
- **Lifecycle Protection**: Prevents accidental deletion

## Usage

### Step 1: Create terraform.tfvars

Copy the example file and customize:

```bash
cp terraform.tfvars.example terraform.tfvars
```

Edit `terraform.tfvars`:

```hcl
state_bucket_name = "your-unique-bucket-name-dev"

tags = {
  Environment = "dev"
  Project     = "bedrock-chat-poc"
  ManagedBy   = "Terraform"
  CreatedAt   = "2025-12-07"
}
```

**Important**: The bucket name must be globally unique across all AWS accounts.

### Step 2: Initialize Terraform

```bash
terraform init
```

### Step 3: Review the Plan

```bash
terraform plan
```

### Step 4: Apply the Configuration

```bash
terraform apply
```

### Step 5: Note the Outputs

After successful deployment, note the bucket name:

```bash
terraform output state_bucket_name
```

You'll need this bucket name to configure the backend for other environments.

## Inputs

| Name | Description | Type | Required | Default |
|------|-------------|------|----------|---------|
| state_bucket_name | Name of the S3 bucket for Terraform state | string | Yes | - |
| tags | Tags to apply to all resources | map(string) | No | {} |

## Outputs

| Name | Description |
|------|-------------|
| state_bucket_name | Name of the S3 bucket used for Terraform state |
| state_bucket_arn | ARN of the S3 bucket used for Terraform state |
| state_bucket_region | AWS region where the state bucket is located |

## S3 Native Locking

This module uses S3's native state locking capabilities:

- **No DynamoDB Required**: S3 backend uses conditional write operations for locking
- **Automatic Lock Management**: Terraform handles lock acquisition and release
- **Stale Lock Expiration**: Locks expire automatically if Terraform crashes
- **Cost Effective**: Only S3 storage costs, no additional service charges

### How It Works

1. Before `terraform apply`, Terraform acquires a lock using S3 conditional writes
2. Other Terraform operations wait until the lock is released
3. After completion, the lock is automatically released
4. If Terraform crashes, the lock expires after a timeout

## Security Features

### Encryption

- **At Rest**: AES-256 server-side encryption
- **In Transit**: SSL/TLS enforced via bucket policy

### Access Control

- **Public Access Block**: All public access blocked
- **Bucket Policy**: Denies non-HTTPS requests
- **IAM**: Access controlled through IAM policies

### Lifecycle Protection

The bucket has `prevent_destroy` lifecycle rule to prevent accidental deletion. To delete:

1. Remove the `prevent_destroy` rule from `main.tf`
2. Run `terraform apply` to update the configuration
3. Run `terraform destroy`

## Best Practices

1. **One Bucket Per Environment**: Create separate state buckets for dev, staging, and prod
2. **Unique Names**: Use descriptive, unique bucket names (e.g., `project-terraform-state-env`)
3. **Version Control**: Commit the module code but not `terraform.tfvars`
4. **Access Control**: Limit IAM permissions to only those who need state access
5. **Monitoring**: Enable CloudTrail logging for state bucket access audit

## Troubleshooting

### Bucket Name Already Exists

S3 bucket names are globally unique. If you get an error that the bucket already exists:

1. Choose a different, more unique bucket name
2. Add a random suffix or your AWS account ID to the name

### Permission Denied

Ensure your AWS credentials have the following permissions:

- `s3:CreateBucket`
- `s3:PutBucketVersioning`
- `s3:PutEncryptionConfiguration`
- `s3:PutBucketPublicAccessBlock`
- `s3:PutBucketPolicy`

### Cannot Destroy Bucket

The bucket has lifecycle protection. See "Lifecycle Protection" section above.

## Example: Multi-Environment Setup

Create separate state buckets for each environment:

**Development**:
```hcl
state_bucket_name = "bedrock-chat-poc-terraform-state-dev"
tags = { Environment = "dev" }
```

**Staging**:
```hcl
state_bucket_name = "bedrock-chat-poc-terraform-state-staging"
tags = { Environment = "staging" }
```

**Production**:
```hcl
state_bucket_name = "bedrock-chat-poc-terraform-state-prod"
tags = { Environment = "prod" }
```

## Next Steps

After creating the state bucket:

1. Note the bucket name from the output
2. Configure the backend in your environment directories
3. Initialize Terraform in each environment with the backend configuration
4. Deploy the remaining infrastructure modules
