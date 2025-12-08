output "agent_id" {
  description = "The Bedrock Agent ID"
  value       = aws_bedrockagent_agent.this.id
}

output "agent_arn" {
  description = "The Bedrock Agent ARN"
  value       = aws_bedrockagent_agent.this.agent_arn
}

output "agent_alias_id" {
  description = "The Agent Alias ID for DRAFT version"
  value       = aws_bedrockagent_agent_alias.draft.agent_alias_id
}

output "agent_alias_arn" {
  description = "The Agent Alias ARN"
  value       = aws_bedrockagent_agent_alias.draft.agent_alias_arn
}
