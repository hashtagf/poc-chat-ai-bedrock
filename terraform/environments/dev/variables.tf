# General Configuration
variable "environment" {
  description = "Environment name (dev, staging, prod)"
  type        = string
  default     = "dev"

  validation {
    condition     = contains(["dev", "staging", "prod"], var.environment)
    error_message = "Environment must be dev, staging, or prod."
  }
}

variable "project_name" {
  description = "Project name for resource naming"
  type        = string
  default     = "bedrock-chat-poc"
}

variable "aws_region" {
  description = "AWS region for resource deployment"
  type        = string
  default     = "us-east-1"

  validation {
    condition     = can(regex("^[a-z]{2}-[a-z]+-[0-9]{1}$", var.aws_region))
    error_message = "AWS region must match the pattern: us-east-1, eu-west-1, etc."
  }
}

# Bedrock Agent Configuration
variable "agent_name" {
  description = "Name of the Bedrock Agent"
  type        = string
}

variable "foundation_model" {
  description = "Foundation model ID (e.g., anthropic.claude-v2)"
  type        = string

  validation {
    condition     = can(regex("^[a-z]+\\.", var.foundation_model))
    error_message = "Foundation model ID must start with provider name (e.g., anthropic., amazon.)."
  }
}

variable "agent_instruction" {
  description = "Instructions for the agent behavior"
  type        = string
  default     = "You are a helpful AI assistant for the chat POC application."
}

variable "idle_session_ttl" {
  description = "Session timeout in seconds"
  type        = number
  default     = 1800
}

# Knowledge Base Configuration
variable "knowledge_base_name" {
  description = "Name of the Knowledge Base"
  type        = string
}

variable "embedding_model" {
  description = "Embedding model ID (e.g., amazon.titan-embed-text-v1)"
  type        = string

  validation {
    condition     = can(regex("^[a-z]+\\.", var.embedding_model))
    error_message = "Embedding model ID must start with provider name (e.g., amazon., cohere.)."
  }
}

variable "s3_bucket_name" {
  description = "Name of the S3 bucket for knowledge base documents"
  type        = string
}

variable "s3_vector_bucket_name" {
  description = "Name of the S3 bucket for vector storage"
  type        = string
}

# VPC Configuration (disabled for dev)
variable "enable_vpc" {
  description = "Enable VPC deployment"
  type        = bool
  default     = false
}

# Tags
variable "tags" {
  description = "Resource tags"
  type        = map(string)
  default     = {}
}
