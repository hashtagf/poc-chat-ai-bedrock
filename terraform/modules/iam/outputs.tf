output "agent_role_arn" {
  description = "ARN of the Bedrock Agent execution role"
  value       = aws_iam_role.agent_role.arn
}

output "agent_role_name" {
  description = "Name of the Bedrock Agent execution role"
  value       = aws_iam_role.agent_role.name
}

output "kb_role_arn" {
  description = "ARN of the Knowledge Base execution role"
  value       = aws_iam_role.kb_role.arn
}

output "kb_role_name" {
  description = "Name of the Knowledge Base execution role"
  value       = aws_iam_role.kb_role.name
}
