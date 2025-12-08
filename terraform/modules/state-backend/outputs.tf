output "state_bucket_name" {
  description = "Name of the S3 bucket used for Terraform state"
  value       = aws_s3_bucket.terraform_state.bucket
}

output "state_bucket_arn" {
  description = "ARN of the S3 bucket used for Terraform state"
  value       = aws_s3_bucket.terraform_state.arn
}

output "state_bucket_region" {
  description = "AWS region where the state bucket is located"
  value       = aws_s3_bucket.terraform_state.region
}
