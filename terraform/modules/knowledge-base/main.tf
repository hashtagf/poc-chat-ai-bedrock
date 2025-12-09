# AWS Provider 6.25.0+ is required for S3 Vectors resources (aws_s3vectors_vector_bucket, aws_s3vectors_index).
# AWSCC provider is still required for aws_bedrockagent_knowledge_base with S3_VECTORS storage type
# as the AWS provider doesn't yet support S3 Vectors configuration in knowledge bases.
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 6.25.0"
    }
    awscc = {
      source  = "hashicorp/awscc"
      version = ">= 1.0.0"
    }
    time = {
      source  = "hashicorp/time"
      version = "~> 0.9"
    }
  }
}

# S3 bucket for documents
resource "aws_s3_bucket" "documents" {
  bucket = "${var.project_name}-docs-${var.environment}-${substr(md5("${var.project_name}-${var.environment}"), 0, 8)}"

  tags = merge(var.tags, {
    Name = "${var.project_name}-docs-${var.environment}"
  })
}

resource "aws_s3_bucket_versioning" "documents" {
  bucket = aws_s3_bucket.documents.id

  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "documents" {
  bucket = aws_s3_bucket.documents.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

# S3 Vectors bucket using AWS provider (6.25.0+)
resource "aws_s3vectors_vector_bucket" "vectors" {
  vector_bucket_name = "${var.project_name}-vec-${var.environment}"

  encryption_configuration {
    sse_type = "AES256"
  }

  tags = var.tags
}

# S3 Vectors index
resource "aws_s3vectors_index" "main" {
  index_name         = "${var.project_name}-idx-${var.environment}"
  vector_bucket_name = aws_s3vectors_vector_bucket.vectors.vector_bucket_name

  # Titan Embeddings G1 - Text v1.2 uses 1536 dimensions
  dimension       = 1536
  data_type       = "float32"
  distance_metric = "cosine"

  tags = var.tags
}

# IAM role for Knowledge Base
resource "aws_iam_role" "knowledge_base" {
  name = "${var.project_name}-role-${var.environment}"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "bedrock.amazonaws.com"
        }
        Action = "sts:AssumeRole"
        Condition = {
          StringEquals = {
            "aws:SourceAccount" = data.aws_caller_identity.current.account_id
          }
          ArnLike = {
            "aws:SourceArn" = "arn:aws:bedrock:${data.aws_region.current.id}:${data.aws_caller_identity.current.account_id}:knowledge-base/*"
          }
        }
      }
    ]
  })

  tags = var.tags
}

# IAM policy for Knowledge Base
resource "aws_iam_role_policy" "knowledge_base" {
  name = "${var.project_name}-pol-${var.environment}"
  role = aws_iam_role.knowledge_base.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "BedrockInvokeModel"
        Effect = "Allow"
        Action = [
          "bedrock:InvokeModel"
        ]
        Resource = "arn:aws:bedrock:${data.aws_region.current.id}::foundation-model/amazon.titan-embed-text-v1"
      },
      {
        Sid    = "S3DocumentsAccess"
        Effect = "Allow"
        Action = [
          "s3:GetObject",
          "s3:ListBucket"
        ]
        Resource = [
          aws_s3_bucket.documents.arn,
          "${aws_s3_bucket.documents.arn}/*"
        ]
      },
      {
        Sid    = "S3VectorsAccess"
        Effect = "Allow"
        Action = [
          "s3:GetObject",
          "s3:PutObject",
          "s3:DeleteObject",
          "s3:ListBucket"
        ]
        Resource = [
          aws_s3vectors_vector_bucket.vectors.vector_bucket_arn,
          "${aws_s3vectors_vector_bucket.vectors.vector_bucket_arn}/*"
        ]
      },
      {
        Sid    = "S3VectorsIndexAccess"
        Effect = "Allow"
        Action = [
          "s3vectors:Query",
          "s3vectors:QueryVectors",
          "s3vectors:GetVectors",
          "s3vectors:PutVector",
          "s3vectors:PutVectors",
          "s3vectors:DeleteVector",
          "s3vectors:GetVector"
        ]
        Resource = aws_s3vectors_index.main.index_arn
      }
    ]
  })
}

# Wait for IAM role to propagate
resource "time_sleep" "iam_propagation" {
  create_duration = "30s"

  triggers = {
    policy_checksum = sha256(aws_iam_role_policy.knowledge_base.policy)
  }

  depends_on = [
    aws_iam_role_policy.knowledge_base
  ]
}

# Knowledge Base using AWSCC provider (AWS provider doesn't support S3 Vectors storage yet)
resource "awscc_bedrock_knowledge_base" "main" {
  name        = "${var.project_name}-${var.environment}"
  description = "Knowledge base for ${var.project_name} (${var.environment} environment)"
  role_arn    = aws_iam_role.knowledge_base.arn

  knowledge_base_configuration = {
    type = "VECTOR"
    vector_knowledge_base_configuration = {
      embedding_model_arn = "arn:aws:bedrock:${data.aws_region.current.id}::foundation-model/amazon.titan-embed-text-v1"
    }
  }

  storage_configuration = {
    type = "S3_VECTORS"
    s3_vectors_configuration = {
      vector_bucket_arn = aws_s3vectors_vector_bucket.vectors.vector_bucket_arn
      index_arn         = aws_s3vectors_index.main.index_arn
    }
  }

  tags = var.tags

  depends_on = [
    time_sleep.iam_propagation
  ]
}

# Data Source for Knowledge Base
resource "awscc_bedrock_data_source" "s3" {
  knowledge_base_id = awscc_bedrock_knowledge_base.main.knowledge_base_id
  name              = "${var.project_name}-${var.environment}-s3"
  description       = "S3 data source for ${var.project_name} knowledge base"

  data_source_configuration = {
    type = "S3"
    s3_configuration = {
      bucket_arn = aws_s3_bucket.documents.arn
    }
  }
}

# Data sources
data "aws_caller_identity" "current" {}
data "aws_region" "current" {}
