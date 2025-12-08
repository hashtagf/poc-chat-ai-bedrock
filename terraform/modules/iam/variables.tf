variable "project_name" {
  description = "Project name for resource naming"
  type        = string
}

variable "environment" {
  description = "Environment name (dev/staging/prod)"
  type        = string
  validation {
    condition     = contains(["dev", "staging", "prod"], var.environment)
    error_message = "Environment must be dev, staging, or prod"
  }
}

variable "foundation_model_arn" {
  description = "ARN of the foundation model for Bedrock Agent"
  type        = string
}

variable "embedding_model_arn" {
  description = "ARN of the embedding model for Knowledge Base"
  type        = string
}

variable "s3_bucket_arn" {
  description = "ARN of the S3 bucket for knowledge base documents"
  type        = string
}

variable "s3_vector_bucket_arn" {
  description = "ARN of the S3 bucket for vector storage"
  type        = string
}

variable "tags" {
  description = "Resource tags"
  type        = map(string)
  default     = {}
}
