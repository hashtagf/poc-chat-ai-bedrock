# Get current AWS account ID and region
data "aws_caller_identity" "current" {}
data "aws_region" "current" {}

locals {
  account_id = data.aws_caller_identity.current.account_id
  region     = data.aws_region.current.name
}

# Bedrock Agent IAM Role
resource "aws_iam_role" "agent_role" {
  name = "${var.project_name}-agent-role-${var.environment}"

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
            "aws:SourceAccount" = local.account_id
          }
          ArnLike = {
            "aws:SourceArn" = "arn:aws:bedrock:${local.region}:${local.account_id}:agent/*"
          }
        }
      }
    ]
  })

  tags = merge(
    var.tags,
    {
      Name = "${var.project_name}-agent-role-${var.environment}"
    }
  )
}

# Bedrock Agent IAM Policy
resource "aws_iam_role_policy" "agent_policy" {
  name = "${var.project_name}-agent-policy-${var.environment}"
  role = aws_iam_role.agent_role.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "bedrock:InvokeModel"
        ]
        Resource = var.foundation_model_arn
      }
    ]
  })
}

# Knowledge Base IAM Role
resource "aws_iam_role" "kb_role" {
  name = "${var.project_name}-kb-role-${var.environment}"

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
            "aws:SourceAccount" = local.account_id
          }
          ArnLike = {
            "aws:SourceArn" = "arn:aws:bedrock:${local.region}:${local.account_id}:knowledge-base/*"
          }
        }
      }
    ]
  })

  tags = merge(
    var.tags,
    {
      Name = "${var.project_name}-kb-role-${var.environment}"
    }
  )
}

# Knowledge Base IAM Policy
resource "aws_iam_role_policy" "kb_policy" {
  name = "${var.project_name}-kb-policy-${var.environment}"
  role = aws_iam_role.kb_role.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "s3:GetObject",
          "s3:ListBucket"
        ]
        Resource = [
          var.s3_bucket_arn,
          "${var.s3_bucket_arn}/*"
        ]
      },
      {
        Effect = "Allow"
        Action = [
          "s3:PutObject",
          "s3:GetObject",
          "s3:ListBucket",
          "s3:DeleteObject"
        ]
        Resource = [
          var.s3_vector_bucket_arn,
          "${var.s3_vector_bucket_arn}/*"
        ]
      },
      {
        Effect = "Allow"
        Action = [
          "bedrock:InvokeModel"
        ]
        Resource = var.embedding_model_arn
      },
      {
        Effect = "Allow"
        Action = [
          "s3:CreateVectorIndex",
          "s3:DeleteVectorIndex",
          "s3:DescribeVectorIndex",
          "s3:UpdateVectorIndex",
          "s3:PutVectorIndexEntry",
          "s3:GetVectorIndexEntry",
          "s3:DeleteVectorIndexEntry",
          "s3:QueryVectorIndex"
        ]
        Resource = [
          var.s3_vector_bucket_arn,
          "${var.s3_vector_bucket_arn}/*"
        ]
      }
    ]
  })
}
