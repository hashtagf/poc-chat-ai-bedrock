variable "agent_name" {
  description = "Name of the Bedrock Agent"
  type        = string

  validation {
    condition     = length(var.agent_name) > 0 && length(var.agent_name) <= 100
    error_message = "Agent name must be between 1 and 100 characters"
  }
}

variable "foundation_model" {
  description = "Foundation model ID (e.g., 'anthropic.claude-v2', 'amazon.titan-text-express-v1')"
  type        = string

  validation {
    condition     = can(regex("^[a-z0-9-]+\\.[a-z0-9-]+", var.foundation_model))
    error_message = "Foundation model ID must be in format 'provider.model-name'"
  }
}

variable "agent_instruction" {
  description = "Instructions for the agent behavior"
  type        = string

  validation {
    condition     = length(var.agent_instruction) > 0
    error_message = "Agent instruction cannot be empty"
  }
}

variable "agent_role_arn" {
  description = "ARN of the IAM role for agent execution"
  type        = string

  validation {
    condition     = can(regex("^arn:aws:iam::[0-9]{12}:role/", var.agent_role_arn))
    error_message = "Agent role ARN must be a valid IAM role ARN"
  }
}

variable "idle_session_ttl" {
  description = "Session timeout in seconds"
  type        = number
  default     = 1800

  validation {
    condition     = var.idle_session_ttl >= 60 && var.idle_session_ttl <= 3600
    error_message = "Idle session TTL must be between 60 and 3600 seconds"
  }
}

variable "tags" {
  description = "Resource tags"
  type        = map(string)
  default     = {}
}
