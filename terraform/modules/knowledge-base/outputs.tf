output "knowledge_base_id" {
  description = "The Knowledge Base ID"
  value       = data.external.kb_details.result.id
}

output "knowledge_base_arn" {
  description = "The Knowledge Base ARN"
  value       = data.external.kb_details.result.arn
}

output "s3_bucket_name" {
  description = "The S3 bucket name for document uploads"
  value       = aws_s3_bucket.kb_documents.id
}

output "s3_bucket_arn" {
  description = "The S3 bucket ARN"
  value       = aws_s3_bucket.kb_documents.arn
}

output "s3_vector_bucket_name" {
  description = "The S3 bucket name for vector storage"
  value       = aws_s3_bucket.kb_vectors.id
}

output "s3_vector_bucket_arn" {
  description = "The S3 bucket ARN for vector storage"
  value       = aws_s3_bucket.kb_vectors.arn
}

output "data_source_id" {
  description = "The data source ID"
  value       = data.external.data_source_details.result.data_source_id
}
