# Remote State Configuration
# 
# This configuration uses S3 for remote state storage with native locking.
# S3 native locking is enabled automatically when using the S3 backend (Terraform >= 1.5.0).
# No DynamoDB table is required.
#
# Before using this backend configuration:
# 1. Deploy the state-backend module to create the S3 bucket
# 2. Update the bucket name below to match your state bucket
# 3. Run `terraform init` to migrate state to S3

terraform {
  backend "s3" {
    # Update this bucket name to match your state backend bucket
    # Example: "bedrock-chat-poc-terraform-state-staging"
    bucket = "bedrock-chat-poc-terraform-state"

    # State file path within the bucket
    key = "environments/staging/terraform.tfstate"

    # AWS region where the state bucket is located
    region = "us-east-1"

    # Enable encryption at rest for state file
    encrypt = true

    # S3 native locking is enabled automatically
    # No additional configuration needed
  }
}
