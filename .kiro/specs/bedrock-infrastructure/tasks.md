# Implementation Plan: Bedrock Infrastructure with Terraform

This implementation plan breaks down the Terraform infrastructure into discrete, actionable coding tasks. Each task builds incrementally on previous tasks and references specific requirements from the requirements document.

## Task List

- [ ] 1. Set up Terraform project structure and state backend
  - Create directory structure: `terraform/modules/`, `terraform/environments/dev/`, `terraform/environments/staging/`, `terraform/environments/prod/`
  - Create `terraform/modules/state-backend/` module with variables.tf, main.tf, outputs.tf
  - Implement S3 bucket for Terraform state with versioning and encryption
  - Implement DynamoDB table for state locking with LockID primary key
  - Create `terraform/backend.tf` configuration file for remote state
  - _Requirements: 4.1, 4.2, 4.3, 4.4_

- [ ] 1.1 Write property test for state backend tagging
  - **Property 1: Universal Resource Tagging**
  - **Validates: Requirements 5.1, 5.2, 5.3, 5.4**

- [ ] 2. Create IAM module for Bedrock permissions
  - Create `terraform/modules/iam/` with variables.tf, main.tf, outputs.tf
  - Define input variables: project_name, environment, foundation_model_arn, embedding_model_arn, s3_bucket_arn, opensearch_collection_arn, tags
  - Implement Agent IAM role with trust policy for bedrock.amazonaws.com
  - Implement Agent IAM policy granting bedrock:InvokeModel on foundation model ARN
  - Implement Knowledge Base IAM role with trust policy for bedrock.amazonaws.com
  - Implement Knowledge Base IAM policy granting s3:GetObject, s3:ListBucket, aoss:APIAccessAll, bedrock:InvokeModel
  - Define outputs: agent_role_arn, agent_role_name, kb_role_arn, kb_role_name
  - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5_

- [ ] 2.1 Write property test for IAM least privilege compliance
  - **Property 2: IAM Least Privilege Compliance**
  - **Validates: Requirements 3.5**

- [ ] 3. Create Bedrock Agent module
  - Create `terraform/modules/bedrock-agent/` with variables.tf, main.tf, outputs.tf
  - Define input variables: agent_name, foundation_model, agent_instruction, agent_role_arn, idle_session_ttl, tags
  - Implement aws_bedrockagent_agent resource with foundation model configuration
  - Implement aws_bedrockagent_agent_alias resource for DRAFT version
  - Implement terraform_data resource with local-exec provisioner for agent preparation
  - Implement time_sleep resource for agent initialization delay
  - Define outputs: agent_id, agent_arn, agent_alias_id, agent_alias_arn
  - _Requirements: 1.1, 1.2, 1.3, 1.4_

- [ ] 3.1 Write property test for agent tagging
  - **Property 1: Universal Resource Tagging**
  - **Validates: Requirements 5.1, 5.2, 5.3, 5.4**

- [ ] 4. Create Knowledge Base module with S3 and OpenSearch
  - Create `terraform/modules/knowledge-base/` with variables.tf, main.tf, outputs.tf
  - Define input variables: knowledge_base_name, embedding_model, kb_role_arn, s3_bucket_name, opensearch_collection_name, vector_index_name, vector_field, text_field, metadata_field, tags
  - Implement aws_s3_bucket resource for knowledge base documents
  - Implement aws_s3_bucket_versioning to enable versioning
  - Implement aws_s3_bucket_server_side_encryption_configuration for AES-256 encryption
  - Implement aws_s3_bucket_public_access_block to prevent public access
  - Implement aws_opensearchserverless_security_policy for encryption
  - Implement aws_opensearchserverless_security_policy for network access
  - Implement aws_opensearchserverless_access_policy for data access
  - Implement aws_opensearchserverless_collection resource for vector storage
  - Implement aws_bedrockagent_knowledge_base resource with embedding model
  - Implement aws_bedrockagent_data_source resource connecting S3 to Knowledge Base
  - Define outputs: knowledge_base_id, knowledge_base_arn, s3_bucket_name, s3_bucket_arn, opensearch_collection_endpoint, data_source_id
  - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5, 2.6_

- [ ] 4.1 Write property test for knowledge base resource tagging
  - **Property 1: Universal Resource Tagging**
  - **Validates: Requirements 5.1, 5.2, 5.3, 5.4**

- [ ] 5. Create VPC module for production deployment
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

- [ ] 5.1 Write property test for VPC resource tagging
  - **Property 1: Universal Resource Tagging**
  - **Validates: Requirements 5.1, 5.2, 5.3, 5.4**

- [ ] 6. Create development environment configuration
  - Create `terraform/environments/dev/main.tf` that instantiates all modules
  - Create `terraform/environments/dev/variables.tf` with all required variable definitions
  - Create `terraform/environments/dev/outputs.tf` mapping to application environment variables
  - Create `terraform/environments/dev/terraform.tfvars` with dev-specific values
  - Configure module calls: iam, bedrock-agent, knowledge-base
  - Set enable_vpc = false for development environment
  - Define outputs: bedrock_agent_id, bedrock_agent_alias_id, bedrock_knowledge_base_id, s3_bucket_name, aws_region
  - _Requirements: 1.5, 6.1, 6.2, 6.3, 6.4, 6.5, 7.1, 7.2, 7.3_

- [ ] 7. Create staging environment configuration
  - Create `terraform/environments/staging/terraform.tfvars` with staging-specific values
  - Use same main.tf, variables.tf, outputs.tf structure as dev
  - Configure staging-specific resource names and tags
  - Set enable_vpc = false for staging environment
  - _Requirements: 7.1, 7.2, 7.3_

- [ ] 8. Create production environment configuration
  - Create `terraform/environments/prod/terraform.tfvars` with production-specific values
  - Use same main.tf, variables.tf, outputs.tf structure as dev
  - Configure production-specific resource names and tags
  - Set enable_vpc = true and configure VPC parameters
  - Add VPC module instantiation with multi-AZ configuration
  - Configure stricter security settings (multi-AZ NAT, standby replicas)
  - _Requirements: 7.1, 7.2, 7.3, 7.4, 9.7_

- [ ] 9. Implement variable validation rules
  - Add validation blocks to variables.tf for CIDR block formats
  - Add validation for foundation model ID format
  - Add validation for embedding model ID format
  - Add validation for environment values (dev, staging, prod)
  - Add validation for AWS region format
  - _Requirements: 7.3_

- [ ] 10. Create deployment documentation
  - Create `terraform/README.md` with deployment instructions
  - Document state backend bootstrap process
  - Document environment deployment workflow
  - Document how to update application configuration from Terraform outputs
  - Document Knowledge Base document upload and sync process
  - Include troubleshooting section for common errors
  - _Requirements: 6.1, 6.2, 6.3, 6.4, 6.5_

- [ ] 11. Write Terratest integration tests
  - Create `terraform/test/` directory with Go test files
  - Write test for Bedrock Agent module (verify agent creation, outputs)
  - Write test for Knowledge Base module (verify S3, OpenSearch, KB creation)
  - Write test for IAM module (verify roles and policies)
  - Write test for VPC module (verify subnets, endpoints, security groups)
  - Write test for full dev environment deployment
  - Implement cleanup with defer terraform.Destroy
  - _Requirements: All requirements validated through infrastructure testing_

- [ ] 12. Checkpoint - Validate infrastructure deployment
  - Ensure all Terraform modules validate successfully with `terraform validate`
  - Ensure all Terraform code is formatted with `terraform fmt`
  - Deploy to test AWS account and verify all resources are created
  - Verify Terraform outputs match expected format
  - Ask the user if questions arise

## Notes

- All tasks are required for comprehensive infrastructure implementation
- Each task includes references to specific requirements being addressed
- Property-based tests validate universal properties across all resources
- Integration tests validate actual AWS resource creation and configuration
- The implementation follows a bottom-up approach: modules first, then environments
- VPC module is created but only used in production environment
