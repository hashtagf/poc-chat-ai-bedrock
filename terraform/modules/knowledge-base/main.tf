# Get current AWS region
data "aws_region" "current" {}

# S3 Bucket for Knowledge Base Documents
resource "aws_s3_bucket" "kb_documents" {
  bucket = var.s3_bucket_name

  tags = merge(
    var.tags,
    {
      Name = var.s3_bucket_name
    }
  )
}

# Enable versioning on documents bucket
resource "aws_s3_bucket_versioning" "kb_documents" {
  bucket = aws_s3_bucket.kb_documents.id

  versioning_configuration {
    status = "Enabled"
  }
}

# Enable encryption on documents bucket
resource "aws_s3_bucket_server_side_encryption_configuration" "kb_documents" {
  bucket = aws_s3_bucket.kb_documents.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

# Block public access to documents bucket
resource "aws_s3_bucket_public_access_block" "kb_documents" {
  bucket = aws_s3_bucket.kb_documents.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# S3 Bucket for Vector Storage
resource "aws_s3_bucket" "kb_vectors" {
  bucket = var.s3_vector_bucket_name

  tags = merge(
    var.tags,
    {
      Name = var.s3_vector_bucket_name
    }
  )
}

# Enable versioning on vector bucket
resource "aws_s3_bucket_versioning" "kb_vectors" {
  bucket = aws_s3_bucket.kb_vectors.id

  versioning_configuration {
    status = "Enabled"
  }
}

# Enable encryption on vector bucket
resource "aws_s3_bucket_server_side_encryption_configuration" "kb_vectors" {
  bucket = aws_s3_bucket.kb_vectors.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

# Block public access to vector bucket
resource "aws_s3_bucket_public_access_block" "kb_vectors" {
  bucket = aws_s3_bucket.kb_vectors.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# Create Knowledge Base with S3 Vectors using AWS CLI
# Note: Terraform AWS provider doesn't support S3 Vectors yet, so we use null_resource
resource "null_resource" "kb_with_s3_vectors" {
  triggers = {
    kb_name           = var.knowledge_base_name
    role_arn          = var.kb_role_arn
    embedding_model   = var.embedding_model
    vector_bucket_arn = aws_s3_bucket.kb_vectors.arn
    region            = data.aws_region.current.name
  }

  provisioner "local-exec" {
    command = <<-EOT
      # Create Knowledge Base with S3 Vectors
      KB_ID=$(aws bedrock-agent create-knowledge-base \
        --name "${var.knowledge_base_name}" \
        --role-arn "${var.kb_role_arn}" \
        --knowledge-base-configuration '{
          "type": "VECTOR",
          "vectorKnowledgeBaseConfiguration": {
            "embeddingModelArn": "arn:aws:bedrock:${data.aws_region.current.name}::foundation-model/${var.embedding_model}"
          }
        }' \
        --storage-configuration '{
          "type": "S3_VECTORS",
          "s3VectorsConfiguration": {
            "vectorBucketArn": "${aws_s3_bucket.kb_vectors.arn}"
          }
        }' \
        --region ${data.aws_region.current.name} \
        --query 'knowledgeBase.knowledgeBaseId' \
        --output text 2>/dev/null || \
        aws bedrock-agent list-knowledge-bases \
          --region ${data.aws_region.current.name} \
          --query "knowledgeBaseSummaries[?name=='${var.knowledge_base_name}'].knowledgeBaseId | [0]" \
          --output text)
      
      echo "$KB_ID" > ${path.module}/.kb_id
      
      # Create S3 Data Source
      aws bedrock-agent create-data-source \
        --knowledge-base-id "$KB_ID" \
        --name "${var.knowledge_base_name}-s3-data-source" \
        --data-source-configuration '{
          "type": "S3",
          "s3Configuration": {
            "bucketArn": "${aws_s3_bucket.kb_documents.arn}"
          }
        }' \
        --region ${data.aws_region.current.name} 2>/dev/null || true
    EOT
  }

  provisioner "local-exec" {
    when    = destroy
    command = <<-EOT
      if [ -f ${path.module}/.kb_id ]; then
        KB_ID=$(cat ${path.module}/.kb_id)
        
        # Delete data sources first
        DATA_SOURCES=$(aws bedrock-agent list-data-sources \
          --knowledge-base-id "$KB_ID" \
          --region ${self.triggers.region} \
          --query 'dataSourceSummaries[].dataSourceId' \
          --output text 2>/dev/null || echo "")
        
        for DS_ID in $DATA_SOURCES; do
          aws bedrock-agent delete-data-source \
            --knowledge-base-id "$KB_ID" \
            --data-source-id "$DS_ID" \
            --region ${self.triggers.region} 2>/dev/null || true
        done
        
        # Wait a bit for data sources to be deleted
        sleep 5
        
        # Delete knowledge base
        aws bedrock-agent delete-knowledge-base \
          --knowledge-base-id "$KB_ID" \
          --region ${self.triggers.region} 2>/dev/null || true
        
        rm -f ${path.module}/.kb_id
      fi
    EOT
  }

  depends_on = [
    aws_s3_bucket.kb_documents,
    aws_s3_bucket.kb_vectors
  ]
}

# Data source to get the created Knowledge Base details
data "external" "kb_details" {
  program = ["bash", "-c", <<-EOT
    if [ -f ${path.module}/.kb_id ]; then
      KB_ID=$(cat ${path.module}/.kb_id)
      aws bedrock-agent get-knowledge-base \
        --knowledge-base-id "$KB_ID" \
        --region ${data.aws_region.current.name} \
        --query '{id: knowledgeBase.knowledgeBaseId, arn: knowledgeBase.knowledgeBaseArn, name: knowledgeBase.name}' \
        --output json 2>/dev/null || echo '{"id":"","arn":"","name":""}'
    else
      echo '{"id":"","arn":"","name":""}'
    fi
  EOT
  ]

  depends_on = [null_resource.kb_with_s3_vectors]
}

# Data source to get data source ID
data "external" "data_source_details" {
  program = ["bash", "-c", <<-EOT
    if [ -f ${path.module}/.kb_id ]; then
      KB_ID=$(cat ${path.module}/.kb_id)
      DS_ID=$(aws bedrock-agent list-data-sources \
        --knowledge-base-id "$KB_ID" \
        --region ${data.aws_region.current.name} \
        --query 'dataSourceSummaries[0].dataSourceId' \
        --output text 2>/dev/null || echo "")
      echo "{\"data_source_id\":\"$DS_ID\"}"
    else
      echo '{"data_source_id":""}'
    fi
  EOT
  ]

  depends_on = [null_resource.kb_with_s3_vectors]
}
