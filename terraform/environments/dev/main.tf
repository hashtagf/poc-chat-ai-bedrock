terraform {
  required_version = ">= 1.5.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 5.0.0"
    }
  }
}

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = var.tags
  }
}

# Data source to get current AWS account ID and region
data "aws_caller_identity" "current" {}
data "aws_region" "current" {}

# Construct model ARNs for IAM policies
locals {
  foundation_model_arn = "arn:aws:bedrock:${data.aws_region.current.name}::foundation-model/${var.foundation_model}"
  embedding_model_arn  = "arn:aws:bedrock:${data.aws_region.current.name}::foundation-model/${var.embedding_model}"
}

# IAM Module - Creates roles and policies for Bedrock Agent and Knowledge Base
module "iam" {
  source = "../../modules/iam"

  project_name         = var.project_name
  environment          = var.environment
  foundation_model_arn = local.foundation_model_arn
  embedding_model_arn  = local.embedding_model_arn
  s3_bucket_arn        = module.knowledge_base.s3_bucket_arn
  s3_vector_bucket_arn = module.knowledge_base.s3_vector_bucket_arn
  tags                 = var.tags
}

# Knowledge Base Module - Creates S3 buckets and Knowledge Base with S3 vector store
module "knowledge_base" {
  source = "../../modules/knowledge-base"

  knowledge_base_name   = var.knowledge_base_name
  embedding_model       = var.embedding_model
  kb_role_arn           = module.iam.kb_role_arn
  s3_bucket_name        = var.s3_bucket_name
  s3_vector_bucket_name = var.s3_vector_bucket_name
  tags                  = var.tags
}

# Bedrock Agent Module - Creates Bedrock Agent with agent alias
module "bedrock_agent" {
  source = "../../modules/bedrock-agent"

  agent_name        = var.agent_name
  foundation_model  = var.foundation_model
  agent_instruction = var.agent_instruction
  agent_role_arn    = module.iam.agent_role_arn
  idle_session_ttl  = var.idle_session_ttl
  tags              = var.tags
}
