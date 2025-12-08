variable "knowledge_base_name" {
  description = "Name of the Knowledge Base"
  type        = string
}

variable "embedding_model" {
  description = "Embedding model ID (e.g., amazon.titan-embed-text-v1)"
  type        = string
}

variable "kb_role_arn" {
  description = "ARN of the IAM role for Knowledge Base"
  type        = string
}

variable "s3_bucket_name" {
  description = "Name of the S3 bucket for knowledge base documents"
  type        = string
}

variable "s3_vector_bucket_name" {
  description = "Name of the S3 bucket for vector storage"
  type        = string
}

variable "vector_dimensions" {
  description = "Number of dimensions for vector embeddings (e.g., 1536 for Titan, 1024 for Cohere)"
  type        = number
  default     = 1536
}

variable "tags" {
  description = "Resource tags"
  type        = map(string)
  default     = {}
}
