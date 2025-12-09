# Implementation Plan: S3 Vectors Fix for Terraform Knowledge Base Module

This implementation plan breaks down the S3 Vectors fix into discrete, actionable coding tasks. Each task builds incrementally and references specific requirements from the requirements document.

## Task List

- [x] 1. Backup current state and document existing resources
  - Create backup directory: `terraform/modules/knowledge-base/backup/`
  - Save current Terraform state: `terraform state pull > backup/state-backup.json`
  - Document current resource IDs and ARNs
  - Save current main.tf, variables.tf, outputs.tf to backup directory
  - Document current Knowledge Base ID and Data Source ID for potential import
  - _Requirements: 9.1, 9.2, 9.3_

- [x] 2. Update provider configuration in knowledge-base module
  - Open `terraform/modules/knowledge-base/main.tf`
  - Remove AWSCC provider from required_providers block
  - Update AWS provider version constraint to ">= 6.25.0"
  - Remove any awscc provider configuration blocks
  - Add comment explaining why AWS provider 6.25.0+ is required
  - _Requirements: 1.1, 1.2, 1.5_

- [x] 3. Replace S3 vector bucket resources with standard S3 bucket
  - In `terraform/modules/knowledge-base/main.tf`, locate aws_s3vectors_vector_bucket resource
  - Replace with standard aws_s3_bucket resource named "vectors"
  - Add aws_s3_bucket_versioning resource for vectors bucket
  - Add aws_s3_bucket_server_side_encryption_configuration with AES256
  - Add aws_s3_bucket_public_access_block with all options set to true
  - Ensure bucket name follows pattern: "${var.project_name}-kb-vectors-${var.environment}"
  - Apply tags from var.tags with additional Name and Purpose tags
  - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5_

- [x] 4. Create vector index resource
  - In `terraform/modules/knowledge-base/main.tf`, locate aws_s3vectors_index resource
  - Replace with aws_bedrockagent_vector_index resource
  - Set index_name to "${var.project_name}-kb-index-${var.environment}"
  - Set vector_bucket_name to aws_s3_bucket.vectors.id (not ARN)
  - Set dimension to 1536 for Titan Embeddings G1
  - Set data_type to "float32"
  - Set distance_metric to "cosine"
  - Add explicit depends_on for aws_s3_bucket.vectors
  - Apply tags from var.tags
  - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5_

- [x] 5. Replace Knowledge Base resource with AWS provider version
  - In `terraform/modules/knowledge-base/main.tf`, locate awscc_bedrock_knowledge_base resource
  - Replace with aws_bedrockagent_knowledge_base resource
  - Update knowledge_base_configuration block structure for AWS provider
  - Update storage_configuration block with type = "S3_VECTORS"
  - Add s3_vectors_configuration block with vector_bucket_arn and index_arn
  - Set vector_bucket_arn to aws_s3_bucket.vectors.arn
  - Set index_arn to aws_bedrockagent_vector_index.main.arn
  - Add explicit depends_on for aws_bedrockagent_vector_index.main
  - Maintain all existing variables (name, description, role_arn, embedding_model)
  - _Requirements: 1.3, 4.1, 4.2, 4.3, 4.4_

- [x] 6. Replace Data Source resource with AWS provider version
  - In `terraform/modules/knowledge-base/main.tf`, locate awscc_bedrock_data_source resource
  - Replace with aws_bedrockagent_data_source resource
  - Update data_source_configuration block structure for AWS provider
  - Maintain S3 configuration with bucket_arn reference
  - Add explicit depends_on for aws_bedrockagent_knowledge_base.main
  - Maintain all existing variables (name, description)
  - _Requirements: 7.2, 7.3, 7.4_

- [x] 7. Update IAM permissions for S3 Vectors operations
  - Open `terraform/modules/iam/main.tf`
  - Locate Knowledge Base IAM role policy
  - Update S3VectorsAccess statement to include s3:DeleteObject action
  - Update S3VectorsIndexAccess statement with correct action names
  - Change s3vectors:Query to bedrock:Query
  - Change s3vectors:PutVector to bedrock:PutVector
  - Change s3vectors:GetVector to bedrock:GetVector
  - Change s3vectors:DeleteVector to bedrock:DeleteVector
  - Update Resource to reference aws_bedrockagent_vector_index.main.arn
  - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5, 5.6, 5.7, 5.8_

- [x] 8. Update module outputs to maintain backward compatibility
  - Open `terraform/modules/knowledge-base/outputs.tf`
  - Verify all output names remain unchanged
  - Update output values to reference new resource names
  - Update knowledge_base_id to use aws_bedrockagent_knowledge_base.main.id
  - Update knowledge_base_arn to use aws_bedrockagent_knowledge_base.main.arn
  - Update data_source_id to use aws_bedrockagent_data_source.s3.id
  - Update vectors_bucket_name to use aws_s3_bucket.vectors.id
  - Update vectors_bucket_arn to use aws_s3_bucket.vectors.arn
  - Update index_name to use aws_bedrockagent_vector_index.main.index_name
  - Update index_arn to use aws_bedrockagent_vector_index.main.arn
  - _Requirements: 9.1, 9.2_


- [x] 9. Validate Terraform configuration
  - Run `terraform fmt -recursive` to format all files
  - Run `terraform validate` in the knowledge-base module directory
  - Run `terraform validate` in each environment directory (dev, staging, prod)
  - Fix any validation errors that appear
  - Verify no references to AWSCC provider remain
  - Verify all resource references are updated
  - _Requirements: 1.1, 1.2, 1.3, 1.4_

- [x] 10. Update module README documentation
  - Open `terraform/modules/knowledge-base/README.md`
  - Update Prerequisites section to require AWS Provider >= 6.25.0
  - Remove any references to AWSCC provider
  - Update Features section to clarify S3 Vectors implementation
  - Update Architecture diagram to show aws_bedrockagent_* resources
  - Add Cost Optimization section with S3 Vectors vs OpenSearch comparison
  - Update example usage code blocks with new resource types
  - Add troubleshooting section for common migration issues
  - Add note about embedding model dimensions (1536 for Titan, 1024 for Cohere)
  - _Requirements: 6.1, 6.2, 6.3, 6.4, 6.5_

- [x] 11. Create migration guide document
  - Create `terraform/modules/knowledge-base/MIGRATION.md`
  - Document pre-migration checklist (backup state, document resources)
  - Document Option 1: Clean slate migration (destroy and recreate)
  - Document Option 2: In-place migration (state manipulation)
  - Include step-by-step commands for each option
  - Document post-migration validation steps
  - Document rollback procedure
  - Add warnings about data loss and downtime
  - _Requirements: 9.3, 9.4, 9.5_

- [x] 12. Test configuration in development environment
  - Navigate to `terraform/environments/dev/`
  - Run `terraform init -upgrade` to upgrade providers
  - Run `terraform plan` and review the changes
  - Document which resources will be replaced vs updated
  - If acceptable, run `terraform apply` to deploy changes
  - Verify all resources are created successfully in AWS Console
  - Check S3 buckets exist with correct configuration
  - Check vector index exists with correct parameters
  - Check Knowledge Base exists with S3_VECTORS storage type
  - _Requirements: 4.4, 7.5, 9.4_

- [x] 13. Test document ingestion workflow
  - Upload a test document to the documents S3 bucket
  - Run the ingestion script: `cd terraform/modules/knowledge-base/scripts && ./ingest-kb.sh`
  - Verify ingestion job starts successfully
  - Monitor ingestion job progress
  - Verify ingestion completes without errors
  - Check that vectors are created in the S3 vectors bucket
  - Verify statistics are displayed (documents processed, vectors created)
  - _Requirements: 8.1, 8.2, 8.3, 8.4_

- [x] 14. Test Knowledge Base query functionality
  - Use AWS CLI to test retrieve operation
  - Run: `aws bedrock-agent-runtime retrieve --knowledge-base-id <kb-id> --retrieval-query text="test query" --region ap-southeast-1`
  - Verify query returns results
  - Verify results include citations to source documents
  - Test with multiple different queries
  - Verify query latency is acceptable (sub-second)
  - _Requirements: 4.4_

- [x] 15. Update application configuration
  - Navigate to project root directory
  - Run `cd terraform/environments/dev && terraform output -raw knowledge_base_id`
  - Update `.env` file with new Knowledge Base ID (should be same if in-place migration)
  - Verify backend application can connect to Knowledge Base
  - Test end-to-end RAG workflow through the application
  - Verify chat responses include knowledge base context
  - _Requirements: 9.1, 9.2_

- [x] 16. Checkpoint - Validate complete deployment
  - Ensure all Terraform resources are created successfully
  - Ensure Knowledge Base ingestion works correctly
  - Ensure Knowledge Base queries return results
  - Ensure application integration works end-to-end
  - Document any issues encountered and resolutions
  - Ask the user if questions arise

- [ ]* 17. Write Terratest validation tests
  - Create `terraform/modules/knowledge-base/test/` directory
  - Create Go test file: `knowledge_base_test.go`
  - Write test to validate provider version is 6.25.0+
  - Write test to validate S3 bucket configuration (encryption, versioning, public access block)
  - Write test to validate vector index configuration (dimensions, distance metric, data type)
  - Write test to validate Knowledge Base storage type is S3_VECTORS
  - Write test to validate IAM permissions include all required actions
  - Write test to validate all resources have required tags (Property 1)
  - Run tests with `go test -v -timeout 30m`
  - _Requirements: All requirements validated through infrastructure testing_

- [ ]* 18. Update root Terraform documentation
  - Open `terraform/README.md`
  - Update Prerequisites section with AWS Provider 6.25.0+ requirement
  - Update Quick Start section with `terraform init -upgrade` step
  - Add S3 Vectors Benefits section highlighting cost savings
  - Update troubleshooting section with migration-related issues
  - Add link to module-specific MIGRATION.md
  - _Requirements: 6.1, 6.2, 6.3_

## Notes

- Tasks marked with * are optional testing and documentation tasks
- Core implementation tasks (1-16) are required for functional infrastructure
- Task 12-15 involve actual deployment and testing in AWS
- Migration can be done as clean slate (destroy/recreate) or in-place (state manipulation)
- Backup state before making any changes (Task 1)
- Test in development environment before applying to staging/production
- The ingestion script (ingest-kb.sh) should already exist and work with new resources
- All variable and output names remain the same for backward compatibility
- Property 1 (Universal Resource Tagging) is validated in optional Task 17
