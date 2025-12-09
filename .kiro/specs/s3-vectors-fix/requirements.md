# Requirements Document

## Introduction

This document specifies the requirements for fixing the Terraform knowledge base module to properly use S3 Vectors storage according to AWS best practices and the latest Terraform AWS provider documentation. The current implementation uses the AWSCC provider for S3 Vectors, but there are compatibility and configuration issues that need to be resolved to ensure proper integration with Amazon Bedrock Knowledge Bases.

## Glossary

- **S3 Vectors**: Amazon S3's vector storage capability for cost-effective vector embeddings storage
- **Vector Bucket**: An S3 bucket specifically configured for storing vector embeddings
- **Vector Index**: A searchable index structure for vector similarity queries
- **AWSCC Provider**: AWS Cloud Control API provider for Terraform (community provider)
- **AWS Provider**: Official HashiCorp AWS provider for Terraform
- **Knowledge Base**: Amazon Bedrock feature for retrieval-augmented generation (RAG)
- **Embedding Model**: AI model that converts text into vector representations
- **Storage Configuration**: The configuration block that defines where Knowledge Base vectors are stored

## Requirements

### Requirement 1

**User Story:** As a DevOps engineer, I want to use the correct Terraform provider for S3 Vectors, so that the infrastructure is stable and follows AWS best practices.

#### Acceptance Criteria

1. WHEN reviewing provider documentation THEN the system SHALL use the AWS provider version 6.25.0 or higher for S3 Vectors support
2. WHEN configuring S3 Vectors THEN the system SHALL use native AWS provider resources instead of AWSCC provider
3. WHEN the Knowledge Base is created THEN the system SHALL use the aws_bedrockagent_knowledge_base resource from the AWS provider
4. WHEN provider versions are specified THEN the system SHALL pin exact versions to prevent breaking changes
5. WHEN multiple providers are used THEN the system SHALL document why each provider is necessary

### Requirement 2

**User Story:** As a DevOps engineer, I want the S3 vector bucket configured correctly, so that Bedrock can store and query vector embeddings.

#### Acceptance Criteria

1. WHEN creating the vector bucket THEN the system SHALL use aws_s3_bucket resource with appropriate naming
2. WHEN configuring the vector bucket THEN the system SHALL enable server-side encryption with AES-256
3. WHEN configuring the vector bucket THEN the system SHALL enable versioning for data protection
4. WHEN configuring the vector bucket THEN the system SHALL block all public access
5. WHEN creating the vector bucket THEN the system SHALL apply consistent tags matching other resources

### Requirement 3

**User Story:** As a DevOps engineer, I want the vector index configured properly, so that semantic search queries return accurate results.

#### Acceptance Criteria

1. WHEN creating the vector index THEN the system SHALL configure dimensions matching the embedding model (1536 for Titan Embeddings G1)
2. WHEN creating the vector index THEN the system SHALL use cosine distance metric for similarity calculations
3. WHEN creating the vector index THEN the system SHALL use float32 data type for vector storage
4. WHEN creating the vector index THEN the system SHALL associate the index with the vector bucket
5. WHEN the index is created THEN the system SHALL output the index ARN for reference

### Requirement 4

**User Story:** As a DevOps engineer, I want the Knowledge Base storage configuration to use S3 Vectors, so that vector storage costs are minimized.

#### Acceptance Criteria

1. WHEN configuring Knowledge Base storage THEN the system SHALL set storage type to "S3_VECTORS"
2. WHEN configuring S3 Vectors storage THEN the system SHALL provide the vector bucket ARN
3. WHEN configuring S3 Vectors storage THEN the system SHALL provide the vector index ARN
4. WHEN the Knowledge Base is created THEN the system SHALL verify the storage configuration is valid
5. WHEN using S3 Vectors THEN the system SHALL document cost savings compared to OpenSearch Serverless

### Requirement 5

**User Story:** As a DevOps engineer, I want IAM permissions configured correctly for S3 Vectors, so that Bedrock can read and write vector embeddings.

#### Acceptance Criteria

1. WHEN the Knowledge Base IAM role is created THEN the system SHALL grant s3:GetObject permission on the vector bucket
2. WHEN the Knowledge Base IAM role is created THEN the system SHALL grant s3:PutObject permission on the vector bucket
3. WHEN the Knowledge Base IAM role is created THEN the system SHALL grant s3:DeleteObject permission on the vector bucket
4. WHEN the Knowledge Base IAM role is created THEN the system SHALL grant s3:ListBucket permission on the vector bucket
5. WHEN S3 Vectors are used THEN the system SHALL grant s3vectors:Query permission on the index ARN
6. WHEN S3 Vectors are used THEN the system SHALL grant s3vectors:PutVector permission on the index ARN
7. WHEN S3 Vectors are used THEN the system SHALL grant s3vectors:GetVector permission on the index ARN
8. WHEN S3 Vectors are used THEN the system SHALL grant s3vectors:DeleteVector permission on the index ARN

### Requirement 6

**User Story:** As a DevOps engineer, I want clear documentation on S3 Vectors setup, so that I can troubleshoot issues and understand the configuration.

#### Acceptance Criteria

1. WHEN reviewing module documentation THEN the system SHALL explain why S3 Vectors is chosen over OpenSearch
2. WHEN reviewing module documentation THEN the system SHALL document the cost comparison ($5-10/month vs $700/month)
3. WHEN reviewing module documentation THEN the system SHALL explain the vector index configuration parameters
4. WHEN reviewing module documentation THEN the system SHALL provide troubleshooting steps for common issues
5. WHEN reviewing module documentation THEN the system SHALL include example queries for testing the Knowledge Base

### Requirement 7

**User Story:** As a DevOps engineer, I want the module to handle resource dependencies correctly, so that resources are created in the proper order.

#### Acceptance Criteria

1. WHEN creating resources THEN the system SHALL create the vector bucket before the vector index
2. WHEN creating resources THEN the system SHALL create the vector index before the Knowledge Base
3. WHEN creating resources THEN the system SHALL create IAM roles before the Knowledge Base
4. WHEN creating resources THEN the system SHALL use explicit depends_on where implicit dependencies are insufficient
5. WHEN destroying resources THEN the system SHALL handle cleanup in reverse order without errors

### Requirement 8

**User Story:** As a developer, I want to easily ingest documents into the Knowledge Base, so that I can test RAG functionality.

#### Acceptance Criteria

1. WHEN documents are uploaded to S3 THEN the system SHALL provide a script to trigger ingestion
2. WHEN ingestion is triggered THEN the system SHALL start an ingestion job via AWS API
3. WHEN ingestion is running THEN the system SHALL monitor job status and report progress
4. WHEN ingestion completes THEN the system SHALL display statistics (documents processed, vectors created)
5. WHEN ingestion fails THEN the system SHALL display error messages and suggest remediation steps

### Requirement 9

**User Story:** As a DevOps engineer, I want the Terraform configuration to be compatible with existing environments, so that I can upgrade without breaking changes.

#### Acceptance Criteria

1. WHEN upgrading the module THEN the system SHALL maintain backward compatibility with existing variable names
2. WHEN upgrading the module THEN the system SHALL maintain backward compatibility with existing output names
3. WHEN upgrading the module THEN the system SHALL provide migration instructions if breaking changes are necessary
4. WHEN applying changes THEN the system SHALL use terraform plan to preview changes before applying
5. WHEN resources need replacement THEN the system SHALL document which resources will be recreated
