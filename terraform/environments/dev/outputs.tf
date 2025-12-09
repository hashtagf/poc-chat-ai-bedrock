# Outputs that map to application environment variables

output "bedrock_agent_id" {
  description = "Bedrock Agent ID for BEDROCK_AGENT_ID env var"
  value       = module.bedrock_agent.agent_id
}

output "bedrock_agent_alias_id" {
  description = "Bedrock Agent Alias ID for BEDROCK_AGENT_ALIAS_ID env var"
  value       = module.bedrock_agent.agent_alias_id
}

output "knowledge_base_id" {
  description = "Knowledge Base ID for BEDROCK_KNOWLEDGE_BASE_ID env var"
  value       = module.knowledge_base.knowledge_base_id
}

output "documents_bucket_name" {
  description = "S3 bucket name for uploading knowledge base documents"
  value       = module.knowledge_base.documents_bucket_name
}

output "aws_region" {
  description = "AWS region for AWS_REGION env var"
  value       = var.aws_region
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

output "vectors_bucket_name" {
  description = "S3 bucket name for vector storage"
  value       = module.knowledge_base.vectors_bucket_name
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
  value       = module.knowledge_base.role_arn
}
