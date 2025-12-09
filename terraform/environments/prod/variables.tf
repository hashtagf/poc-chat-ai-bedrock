# General Configuration
variable "environment" {
  description = "Environment name (dev, staging, prod)"
  type        = string
  default     = "prod"

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
  default     = "ap-southeast-1"

  validation {
    condition     = can(regex("^[a-z]{2}-[a-z]+-[0-9]{1}$", var.aws_region))
    error_message = "AWS region must match the pattern: ap-southeast-1, eu-west-1, etc."
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
  default     = 3600
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

# VPC Configuration (enabled for production)
variable "enable_vpc" {
  description = "Enable VPC deployment"
  type        = bool
  default     = true
}

variable "vpc_cidr" {
  description = "CIDR block for VPC (e.g., '10.0.0.0/16')"
  type        = string
  default     = "10.0.0.0/16"

  validation {
    condition     = can(cidrhost(var.vpc_cidr, 0))
    error_message = "VPC CIDR must be a valid IPv4 CIDR block."
  }
}

variable "availability_zones" {
  description = "List of availability zones (e.g., ['ap-southeast-1a', 'ap-southeast-1b'])"
  type        = list(string)
  default     = ["ap-southeast-1a", "ap-southeast-1b"]

  validation {
    condition     = length(var.availability_zones) >= 2
    error_message = "At least 2 availability zones must be specified for high availability."
  }
}

variable "private_subnet_cidrs" {
  description = "CIDR blocks for private subnets"
  type        = list(string)
  default     = ["10.0.1.0/24", "10.0.2.0/24"]

  validation {
    condition     = length(var.private_subnet_cidrs) >= 2
    error_message = "At least 2 private subnet CIDRs must be specified."
  }

  validation {
    condition     = alltrue([for cidr in var.private_subnet_cidrs : can(cidrhost(cidr, 0))])
    error_message = "All private subnet CIDRs must be valid IPv4 CIDR blocks."
  }
}

variable "public_subnet_cidrs" {
  description = "CIDR blocks for public subnets"
  type        = list(string)
  default     = ["10.0.101.0/24", "10.0.102.0/24"]

  validation {
    condition     = length(var.public_subnet_cidrs) >= 2
    error_message = "At least 2 public subnet CIDRs must be specified."
  }

  validation {
    condition     = alltrue([for cidr in var.public_subnet_cidrs : can(cidrhost(cidr, 0))])
    error_message = "All public subnet CIDRs must be valid IPv4 CIDR blocks."
  }
}

variable "enable_nat_gateway" {
  description = "Enable NAT gateway for internet access from private subnets"
  type        = bool
  default     = true
}

variable "single_nat_gateway" {
  description = "Use single NAT gateway for cost savings (default: false for high availability in production)"
  type        = bool
  default     = false
}

# Tags
variable "tags" {
  description = "Resource tags"
  type        = map(string)
  default     = {}
}
