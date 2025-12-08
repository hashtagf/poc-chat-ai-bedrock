# Implementation Plan: Bedrock Infrastructure with Terraform

This implementation plan breaks down the Terraform infrastructure into discrete, actionable coding tasks. Each task builds incrementally on previous tasks and references specific requirements from the requirements document.

## Task List

- [x] 1. Set up Terraform project structure and state backend
  - Create directory structure: `terraform/modules/`, `terraform/environments/dev/`, `terraform/environments/staging/`, `terraform/environments/prod/`
  - Create `terraform/modules/state-backend/` module with variables.tf, main.tf, outputs.tf
  - Implement S3 bucket for Terraform state with versioning and encryption
  - Configure S3 native locking for state management (automatic with S3 backend)
  - Create `terraform/backend.tf` configuration file for remote state
  - Add .gitignore entries for Terraform files (.terraform/, *.tfstate, *.tfstate.backup, .terraform.lock.hcl)
  - _Requirements: 4.1, 4.2, 4.3, 4.4_

- [ ] 1.1 Write property test for state backend tagging
  - **Property 1: Universal Resource Tagging**
  - **Validates: Requirements 5.1, 5.2, 5.3, 5.4**

- [x] 2. Create IAM module for Bedrock permissions
  - Create `terraform/modules/iam/` with variables.tf, main.tf, outputs.tf
  - Define input variables: project_name, environment, foundation_model_arn, embedding_model_arn, s3_bucket_arn, s3_vector_bucket_arn, tags
  - Implement Agent IAM role with trust policy for bedrock.amazonaws.com
  - Implement Agent IAM policy granting bedrock:InvokeModel on foundation model ARN
  - Implement Knowledge Base IAM role with trust policy for bedrock.amazonaws.com
  - Implement Knowledge Base IAM policy granting s3:GetObject, s3:ListBucket on documents bucket, s3:PutObject, s3:GetObject, s3:ListBucket on vector bucket, and bedrock:InvokeModel
  - Define outputs: agent_role_arn, agent_role_name, kb_role_arn, kb_role_name
  - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5_

- [ ]* 2.1 Write property test for IAM least privilege compliance
  - **Property 2: IAM Least Privilege Compliance**
  - **Validates: Requirements 3.5**

- [x] 3. Create Bedrock Agent module
  - Create `terraform/modules/bedrock-agent/` with variables.tf, main.tf, outputs.tf
  - Define input variables: agent_name, foundation_model, agent_instruction, agent_role_arn, idle_session_ttl, tags
  - Implement aws_bedrockagent_agent resource with foundation model configuration
  - Implement aws_bedrockagent_agent_alias resource for DRAFT version
  - Implement terraform_data resource with local-exec provisioner for agent preparation
  - Implement time_sleep resource for agent initialization delay
  - Define outputs: agent_id, agent_arn, agent_alias_id, agent_alias_arn
  - _Requirements: 1.1, 1.2, 1.3, 1.4_

- [ ]* 3.1 Write property test for agent tagging
  - **Property 1: Universal Resource Tagging**
  - **Validates: Requirements 5.1, 5.2, 5.3, 5.4**

- [x] 4. Create Knowledge Base module with S3 vector store
  - Create `terraform/modules/knowledge-base/` with variables.tf, main.tf, outputs.tf
  - Define input variables: knowledge_base_name, embedding_model, kb_role_arn, s3_bucket_name, s3_vector_bucket_name, tags
  - Implement aws_s3_bucket resource for knowledge base documents
  - Implement aws_s3_bucket_versioning to enable versioning on documents bucket
  - Implement aws_s3_bucket_server_side_encryption_configuration for AES-256 encryption on documents bucket
  - Implement aws_s3_bucket_public_access_block to prevent public access to documents bucket
  - Implement aws_s3_bucket resource for vector storage
  - Implement aws_s3_bucket_versioning to enable versioning on vector bucket
  - Implement aws_s3_bucket_server_side_encryption_configuration for AES-256 encryption on vector bucket
  - Implement aws_s3_bucket_public_access_block to prevent public access to vector bucket
  - Implement aws_bedrockagent_knowledge_base resource with S3 vector store configuration
  - Implement aws_bedrockagent_data_source resource connecting S3 documents to Knowledge Base
  - Define outputs: knowledge_base_id, knowledge_base_arn, s3_bucket_name, s3_bucket_arn, s3_vector_bucket_name, s3_vector_bucket_arn, data_source_id
  - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5, 2.6, 2.7_

- [ ]* 4.1 Write property test for knowledge base resource tagging
  - **Property 1: Universal Resource Tagging**
  - **Validates: Requirements 5.1, 5.2, 5.3, 5.4**

- [x] 5. Create VPC module for production deployment
  - Create `terraform/modules/vpc/` with variables.tf, main.tf, outputs.tf
  - Define input variables: vpc_name, vpc_cidr, availability_zones, private_subnet_cidrs, public_subnet_cidrs, enable_nat_gateway, single_nat_gateway, tags
  - Implement aws_vpc resource with DNS support enabled
  - Implement aws_subnet resources for private subnets across multiple AZs
  - Implement aws_subnet resources for public subnets across multiple AZs
  - Implement aws_internet_gateway for public subnet internet access
  - Implement aws_eip resources for NAT gateways
  - Implement aws_nat_gateway resources in public subnets
  - Implement aws_route_table resources for private and public subnets
  - Implement aws_route_table_association resources
  - Implement aws_security_group for VPC endpoints allowing HTTPS egress
  - Implement aws_vpc_endpoint for Bedrock Agent Runtime (interface type)
  - Implement aws_vpc_endpoint for S3 (gateway type)
  - Define outputs: vpc_id, private_subnet_ids, public_subnet_ids, bedrock_vpc_endpoint_id, s3_vpc_endpoint_id, security_group_id
  - _Requirements: 9.1, 9.2, 9.3, 9.4, 9.5, 9.7_

- [ ]* 5.1 Write property test for VPC resource tagging
  - **Property 1: Universal Resource Tagging**
  - **Validates: Requirements 5.1, 5.2, 5.3, 5.4**

- [x] 6. Create development environment configuration
  - Create `terraform/environments/dev/main.tf` that instantiates all modules
  - Create `terraform/environments/dev/variables.tf` with all required variable definitions
  - Create `terraform/environments/dev/outputs.tf` mapping to application environment variables
  - Create `terraform/environments/dev/terraform.tfvars` with dev-specific values
  - Create `terraform/environments/dev/backend.tf` with S3 backend configuration
  - Configure module calls: iam, bedrock-agent, knowledge-base (VPC module not included for dev)
  - Set enable_vpc = false for development environment
  - Define outputs: bedrock_agent_id, bedrock_agent_alias_id, bedrock_knowledge_base_id, s3_bucket_name, aws_region
  - _Requirements: 1.5, 6.1, 6.2, 6.3, 6.4, 6.5, 7.1, 7.2, 7.3_

- [x] 7. Create staging environment configuration
  - Create `terraform/environments/staging/main.tf` (similar to dev)
  - Create `terraform/environments/staging/variables.tf` (copy from dev)
  - Create `terraform/environments/staging/outputs.tf` (copy from dev)
  - Create `terraform/environments/staging/backend.tf` with S3 backend configuration
  - Create `terraform/environments/staging/terraform.tfvars` with staging-specific values
  - Configure staging-specific resource names and tags
  - Set enable_vpc = false for staging environment
  - _Requirements: 7.1, 7.2, 7.3_

- [x] 8. Create production environment configuration
  - Create `terraform/environments/prod/main.tf` (copy from dev with VPC module added)
  - Create `terraform/environments/prod/variables.tf` (copy from dev with VPC variables added)
  - Create `terraform/environments/prod/outputs.tf` (copy from dev with VPC outputs added)
  - Create `terraform/environments/prod/backend.tf` with S3 backend configuration
  - Create `terraform/environments/prod/terraform.tfvars` with production-specific values
  - Configure production-specific resource names and tags
  - Set enable_vpc = true and configure VPC parameters
  - Add VPC module instantiation with multi-AZ configuration
  - Configure stricter security settings (multi-AZ NAT gateways)
  - _Requirements: 7.1, 7.2, 7.3, 7.4, 9.7_

- [x] 9. Implement variable validation rules
  - Add validation blocks to environment variables.tf for CIDR block formats
  - Add validation for foundation model ID format (must start with provider name)
  - Add validation for embedding model ID format (must start with provider name)
  - Add validation for environment values (must be dev, staging, or prod)
  - Add validation for AWS region format (must match AWS region pattern)
  - _Requirements: 7.3_

- [x] 10. Create deployment documentation
  - Create `terraform/README.md` with deployment instructions
  - Document state backend bootstrap process (manual first-time setup)
  - Document environment deployment workflow (init, plan, apply)
  - Document how to update application configuration from Terraform outputs
  - Document Knowledge Base document upload and sync process
  - Include troubleshooting section for common errors (state locking, permissions, etc.)
  - Document required AWS CLI version and Terraform version
  - _Requirements: 6.1, 6.2, 6.3, 6.4, 6.5_

- [ ]* 11. Write Terratest integration tests
  - Create `terraform/test/` directory with Go test files
  - Set up Go module with Terratest dependencies
  - Write test for Bedrock Agent module (verify agent creation, outputs)
  - Write test for Knowledge Base module (verify S3 buckets, KB creation)
  - Write test for IAM module (verify roles and policies)
  - Write test for VPC module (verify subnets, endpoints, security groups)
  - Write test for full dev environment deployment
  - Implement cleanup with defer terraform.Destroy
  - _Requirements: All requirements validated through infrastructure testing_

- [x] 12. Checkpoint - Validate infrastructure deployment
  - Ensure all Terraform modules validate successfully with `terraform validate`
  - Ensure all Terraform code is formatted with `terraform fmt`
  - Deploy to test AWS account and verify all resources are created
  - Verify Terraform outputs match expected format
  - Ask the user if questions arise

## Notes

- Tasks marked with * are optional property-based tests and integration tests
- Core implementation tasks (1-10, 12) are required for functional infrastructure
- Each task includes references to specific requirements being addressed
- Property-based tests validate universal properties across all resources
- Integration tests validate actual AWS resource creation and configuration
- The implementation follows a bottom-up approach: modules first, then environments
- VPC module is created but only used in production environment
- S3 native locking is automatic with S3 backend (no DynamoDB table needed)
- All tasks are currently not started - no terraform/ directory exists in the workspace yet
