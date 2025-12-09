terraform {
  required_version = ">= 1.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 6.25.0"
    }
    time = {
      source  = "hashicorp/time"
      version = "~> 0.9"
    }
  }
}

# Get current AWS region
data "aws_region" "current" {}

# Bedrock Agent resource
resource "aws_bedrockagent_agent" "this" {
  agent_name                  = var.agent_name
  agent_resource_role_arn     = var.agent_role_arn
  foundation_model            = var.foundation_model
  instruction                 = var.agent_instruction
  idle_session_ttl_in_seconds = var.idle_session_ttl

  tags = var.tags
}

# Wait for agent to be available before preparation
resource "time_sleep" "agent_creation_wait" {
  create_duration = "10s"
  
  depends_on = [aws_bedrockagent_agent.this]
}

# Prepare the agent (required step after creation)
resource "terraform_data" "agent_preparation" {
  triggers_replace = {
    agent_id = aws_bedrockagent_agent.this.id
  }

  provisioner "local-exec" {
    command = "aws bedrock-agent prepare-agent --agent-id ${aws_bedrockagent_agent.this.id} --region ${data.aws_region.current.name} || true"
  }

  depends_on = [time_sleep.agent_creation_wait]
}

# Wait for agent to be fully initialized after preparation
resource "time_sleep" "agent_initialization" {
  create_duration = "30s"

  depends_on = [terraform_data.agent_preparation]
}

# Create agent alias for DRAFT version
resource "aws_bedrockagent_agent_alias" "draft" {
  agent_id         = aws_bedrockagent_agent.this.id
  agent_alias_name = "DRAFT"
  description      = "DRAFT version alias for ${var.agent_name}"

  tags = var.tags

  depends_on = [time_sleep.agent_initialization]
}
