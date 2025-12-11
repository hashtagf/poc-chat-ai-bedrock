terraform {
  required_version = ">= 1.5.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 6.25.0"
    }
    awscc = {
      source  = "hashicorp/awscc"
      version = ">= 1.0.0"
    }
  }
}

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = var.tags
  }
}

provider "awscc" {
  region = var.aws_region
}

# Data source to get current AWS account ID and region
data "aws_caller_identity" "current" {}
data "aws_region" "current" {}

# Construct model ARN for IAM policies
locals {
  foundation_model_arn = "arn:aws:bedrock:${data.aws_region.current.id}::foundation-model/${var.foundation_model}"
}

# IAM Module - Creates roles and policies for Bedrock Agent
module "iam" {
  source = "../../modules/iam"

  project_name                = var.project_name
  environment                 = var.environment
  foundation_model_arn        = local.foundation_model_arn
  foundation_model_id         = var.foundation_model
  bedrock_agent_id           = module.bedrock_agent.agent_id
  bedrock_agent_alias_id     = module.bedrock_agent.agent_alias_id
  bedrock_knowledge_base_id  = module.knowledge_base.knowledge_base_id
  tags                       = var.tags
}

# Knowledge Base Module - Creates complete Knowledge Base with S3 Vectors
module "knowledge_base" {
  source = "../../modules/knowledge-base"

  project_name = var.project_name
  environment  = var.environment
  tags         = var.tags
}

# Bedrock Agent Module - Creates Bedrock Agent with agent alias
module "bedrock_agent" {
  source = "../../modules/bedrock-agent"

  agent_name        = var.agent_name
  foundation_model  = var.foundation_model
  agent_instruction = var.agent_instruction
  agent_role_arn    = module.iam.agent_role_arn
  idle_session_ttl  = var.idle_session_ttl
  knowledge_base_id = module.knowledge_base.knowledge_base_id
  tags              = var.tags
}
