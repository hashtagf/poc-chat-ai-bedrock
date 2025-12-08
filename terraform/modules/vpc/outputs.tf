output "vpc_id" {
  description = "The VPC ID"
  value       = aws_vpc.main.id
}

output "private_subnet_ids" {
  description = "List of private subnet IDs"
  value       = aws_subnet.private[*].id
}

output "public_subnet_ids" {
  description = "List of public subnet IDs"
  value       = aws_subnet.public[*].id
}

output "bedrock_vpc_endpoint_id" {
  description = "VPC endpoint ID for Bedrock Agent Runtime"
  value       = aws_vpc_endpoint.bedrock_agent_runtime.id
}

output "s3_vpc_endpoint_id" {
  description = "VPC endpoint ID for S3"
  value       = aws_vpc_endpoint.s3.id
}

output "security_group_id" {
  description = "Security group ID for VPC endpoints"
  value       = aws_security_group.vpc_endpoints.id
}
