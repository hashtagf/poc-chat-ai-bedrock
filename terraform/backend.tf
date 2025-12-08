# Terraform Backend Configuration
# 
# This file configures remote state storage in S3 with native locking.
# S3 provides built-in state locking without requiring DynamoDB.
#
# IMPORTANT: This backend configuration should be customized per environment.
# Copy this file to each environment directory (dev/staging/prod) and update
# the bucket name and key path accordingly.
#
# Example usage:
#   terraform init -backend-config="bucket=my-terraform-state-bucket"
#
# Note: The state backend bucket must be created manually before running
# terraform init, or use the state-backend module to create it first.

terraform {
  backend "s3" {
    # Bucket name - should be unique and environment-specific
    # Example: "bedrock-chat-poc-terraform-state-dev"
    bucket = "bedrock-chat-poc-terraform-state"

    # Path within the bucket where state will be stored
    # Use environment-specific paths to separate state files
    key = "terraform.tfstate"

    # AWS region where the state bucket is located
    region = "ap-southeast-1"

    # Enable encryption at rest for state file
    encrypt = true

    # S3 native locking is automatic - no DynamoDB table needed
    # Terraform uses S3's conditional write operations for locking
  }
}
