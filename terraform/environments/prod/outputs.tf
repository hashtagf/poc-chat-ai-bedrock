# Outputs that map to application environment variables

output "bedrock_agent_id" {
  description = "Bedrock Agent ID for BEDROCK_AGENT_ID env var"
  value       = module.bedrock_agent.agent_id
}

output "bedrock_agent_alias_id" {
  description = "Bedrock Agent Alias ID for BEDROCK_AGENT_ALIAS_ID env var"
  value       = module.bedrock_agent.agent_alias_id
}

output "bedrock_knowledge_base_id" {
  description = "Knowledge Base ID for BEDROCK_KNOWLEDGE_BASE_ID env var"
  value       = module.knowledge_base.knowledge_base_id
}

output "s3_bucket_name" {
  description = "S3 bucket name for uploading knowledge base documents"
  value       = module.knowledge_base.s3_bucket_name
}

output "aws_region" {
  description = "AWS region for AWS_REGION env var"
  value       = var.aws_region
}

# VPC Outputs (production only)
output "vpc_id" {
  description = "VPC ID (production only)"
  value       = var.enable_vpc ? module.vpc[0].vpc_id : null
}

output "private_subnet_ids" {
  description = "List of private subnet IDs (production only)"
  value       = var.enable_vpc ? module.vpc[0].private_subnet_ids : null
}

output "public_subnet_ids" {
  description = "List of public subnet IDs (production only)"
  value       = var.enable_vpc ? module.vpc[0].public_subnet_ids : null
}

output "bedrock_vpc_endpoint_id" {
  description = "VPC endpoint ID for Bedrock Agent Runtime (production only)"
  value       = var.enable_vpc ? module.vpc[0].bedrock_vpc_endpoint_id : null
}

output "s3_vpc_endpoint_id" {
  description = "VPC endpoint ID for S3 (production only)"
  value       = var.enable_vpc ? module.vpc[0].s3_vpc_endpoint_id : null
}

output "vpc_endpoint_security_group_id" {
  description = "Security group ID for VPC endpoints (production only)"
  value       = var.enable_vpc ? module.vpc[0].security_group_id : null
}

# Additional outputs for reference
output "agent_arn" {
  description = "Bedrock Agent ARN"
  value       = module.bedrock_agent.agent_arn
}

output "knowledge_base_arn" {
  description = "Knowledge Base ARN"
  value       = module.knowledge_base.knowledge_base_arn
}

output "s3_vector_bucket_name" {
  description = "S3 bucket name for vector storage"
  value       = module.knowledge_base.s3_vector_bucket_name
}

output "data_source_id" {
  description = "Knowledge Base data source ID"
  value       = module.knowledge_base.data_source_id
}

output "agent_role_arn" {
  description = "IAM role ARN for Bedrock Agent"
  value       = module.iam.agent_role_arn
}

output "kb_role_arn" {
  description = "IAM role ARN for Knowledge Base"
  value       = module.iam.kb_role_arn
}
