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

variable "foundation_model_id" {
  description = "ID of the foundation model for Bedrock Agent"
  type        = string
  default     = "us.amazon.nova-2-lite-v1:0"
}

variable "bedrock_agent_id" {
  description = "ID of the Bedrock Agent"
  type        = string
  default     = "W6R84XTD2X"
}

variable "bedrock_agent_alias_id" {
  description = "Alias ID of the Bedrock Agent"
  type        = string
  default     = "TXENIZDWOS"
}

variable "bedrock_knowledge_base_id" {
  description = "ID of the Bedrock Knowledge Base"
  type        = string
  default     = "AQ5JOUEIGF"
}

variable "tags" {
  description = "Resource tags"
  type        = map(string)
  default     = {}
}
