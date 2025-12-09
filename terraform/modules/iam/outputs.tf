output "agent_role_arn" {
  description = "ARN of the Bedrock Agent execution role"
  value       = aws_iam_role.agent_role.arn
}

output "agent_role_name" {
  description = "Name of the Bedrock Agent execution role"
  value       = aws_iam_role.agent_role.name
}
