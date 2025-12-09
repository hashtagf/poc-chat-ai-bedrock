# Migration Guide: AWSCC Provider to AWS Provider for S3 Vectors

This guide provides step-by-step instructions for migrating the Knowledge Base module from the AWSCC provider to the AWS provider (6.25.0+) with native S3 Vectors support.

## Overview

**What's Changing:**
- Provider: AWSCC → AWS Provider 6.25.0+
- Resources: `awscc_bedrock_*` → `aws_bedrockagent_*`
- S3 Vectors: Custom resources → Native AWS provider resources

**Why Migrate:**
- Better provider support and documentation
- Simplified configuration
- Improved stability and compatibility
- Native S3 Vectors support in AWS provider

**Migration Impact:**
- Resources will be replaced (destroyed and recreated)
- Vector embeddings will be lost and need re-ingestion
- Knowledge Base ID may change (depending on migration option)
- Downtime during migration (minutes to hours depending on data volume)

---

## Pre-Migration Checklist

Complete these steps **before** starting the migration:

### 1. Backup Current State

```bash
# Navigate to your environment directory
cd terraform/environments/dev  # or staging/prod

# Pull and save current Terraform state
terraform state pull > backup-state-$(date +%Y%m%d-%H%M%S).json

# Verify backup was created
ls -lh backup-state-*.json
```

### 2. Document Current Resources

```bash
# Save current resource configuration
terraform show > current-resources-$(date +%Y%m%d-%H%M%S).txt

# List all resources in state
terraform state list > state-resources-$(date +%Y%m%d-%H%M%S).txt

# Document current outputs
terraform output > current-outputs-$(date +%Y%m%d-%H%M%S).txt
```

### 3. Record Critical Resource IDs

```bash
# Save Knowledge Base ID
echo "KNOWLEDGE_BASE_ID=$(terraform output -raw knowledge_base_id)" > migration-ids.env

# Save Data Source ID
echo "DATA_SOURCE_ID=$(terraform output -raw data_source_id)" >> migration-ids.env

# Save bucket names
echo "DOCUMENTS_BUCKET=$(terraform output -raw documents_bucket_name)" >> migration-ids.env
echo "VECTORS_BUCKET=$(terraform output -raw vectors_bucket_name)" >> migration-ids.env

# Review saved IDs
cat migration-ids.env
```

### 4. Backup Source Documents

```bash
# Download all documents from S3 (if not already backed up)
DOCS_BUCKET=$(terraform output -raw documents_bucket_name)
mkdir -p backup-documents
aws s3 sync s3://${DOCS_BUCKET}/ backup-documents/

# Verify backup
ls -lh backup-documents/
```

### 5. Document Application Configuration

```bash
# Save current .env file
cp ../../../.env ../../../.env.backup-$(date +%Y%m%d-%H%M%S)

# Note: Vector embeddings cannot be exported directly
# They will need to be regenerated through re-ingestion
```

### 6. Notify Stakeholders

⚠️ **Important:** Inform your team about:
- Planned downtime window
- Expected duration (typically 30-60 minutes)
- Knowledge Base will be unavailable during migration
- Application will need configuration updates post-migration

---

## Migration Options

Choose the migration approach that best fits your environment:

- **Option 1 (Clean Slate)**: Recommended for dev/staging environments
- **Option 2 (In-Place)**: Recommended for production with existing data

---

## Option 1: Clean Slate Migration (Destroy and Recreate)

**Best for:** Development and staging environments where downtime is acceptable.

**Pros:**
- Simplest approach
- Clean state with no legacy configuration
- Guaranteed consistency

**Cons:**
- Complete downtime during migration
- All vector embeddings lost (requires re-ingestion)
- Knowledge Base ID will change

### Step 1: Destroy Existing Resources

```bash
cd terraform/environments/dev  # or your target environment

# Review what will be destroyed
terraform plan -destroy

# Destroy existing resources
terraform destroy

# Confirm all resources are removed
terraform state list  # Should be empty or show only non-KB resources
```

### Step 2: Update Module Code

The module code has already been updated in tasks 2-8. Verify the changes:

```bash
cd ../../modules/knowledge-base

# Verify AWS provider version
grep -A 5 "required_providers" main.tf

# Should show:
# aws = {
#   source  = "hashicorp/aws"
#   version = ">= 6.25.0"
# }

# Verify no AWSCC provider references
grep -i "awscc" main.tf  # Should return no results
```

### Step 3: Initialize with New Provider

```bash
cd ../../environments/dev

# Upgrade providers
terraform init -upgrade

# Verify AWS provider version
terraform version
# Should show: provider registry.terraform.io/hashicorp/aws v6.25.0 or higher
```

### Step 4: Plan and Apply

```bash
# Review the plan
terraform plan -out=tfplan

# Verify expected resources will be created:
# - aws_s3_bucket.vectors
# - aws_bedrockagent_vector_index.main
# - aws_bedrockagent_knowledge_base.main
# - aws_bedrockagent_data_source.s3

# Apply the changes
terraform apply tfplan

# Save new outputs
terraform output > new-outputs.txt
```

### Step 5: Verify Resources in AWS Console

```bash
# Get new Knowledge Base ID
NEW_KB_ID=$(terraform output -raw knowledge_base_id)
echo "New Knowledge Base ID: ${NEW_KB_ID}"

# Verify Knowledge Base exists
aws bedrock-agent get-knowledge-base \
  --knowledge-base-id ${NEW_KB_ID} \
  --region ap-southeast-1

# Verify storage type is S3_VECTORS
aws bedrock-agent get-knowledge-base \
  --knowledge-base-id ${NEW_KB_ID} \
  --region ap-southeast-1 \
  --query 'knowledgeBase.storageConfiguration.type' \
  --output text
# Should output: S3_VECTORS
```

### Step 6: Re-upload Documents and Trigger Ingestion

```bash
# Upload documents to new bucket
DOCS_BUCKET=$(terraform output -raw documents_bucket_name)
aws s3 sync backup-documents/ s3://${DOCS_BUCKET}/

# Verify upload
aws s3 ls s3://${DOCS_BUCKET}/ --recursive

# Trigger ingestion
cd ../../modules/knowledge-base/scripts
./ingest-kb.sh

# Monitor ingestion progress
# The script will display status updates
```

### Step 7: Update Application Configuration

```bash
cd ../../../../

# Update .env file with new Knowledge Base ID
NEW_KB_ID=$(cd terraform/environments/dev && terraform output -raw knowledge_base_id)
sed -i.bak "s/BEDROCK_KNOWLEDGE_BASE_ID=.*/BEDROCK_KNOWLEDGE_BASE_ID=${NEW_KB_ID}/" .env

# Verify update
grep BEDROCK_KNOWLEDGE_BASE_ID .env
```

### Step 8: Test End-to-End

```bash
# Test Knowledge Base query
aws bedrock-agent-runtime retrieve \
  --knowledge-base-id ${NEW_KB_ID} \
  --retrieval-query text="test query" \
  --region ap-southeast-1

# Start backend application
cd backend
go run cmd/server/main.go

# In another terminal, test the application
# (Use your application's test procedures)
```

---

## Option 2: In-Place Migration (State Manipulation)

**Best for:** Production environments where you want to minimize downtime and preserve resource IDs.

**Pros:**
- Potentially preserve Knowledge Base ID
- Minimal downtime if successful
- Existing vector embeddings may be preserved

**Cons:**
- More complex process
- Higher risk of state corruption
- May still require resource replacement
- Requires careful state manipulation

⚠️ **Warning:** This approach requires advanced Terraform knowledge. Test in dev/staging first.

### Step 1: Analyze Current State

```bash
cd terraform/environments/prod  # or your target environment

# List all resources
terraform state list

# Identify AWSCC resources to migrate:
# - awscc_bedrock_knowledge_base.main
# - awscc_bedrock_data_source.s3
# - Any aws_s3vectors_* resources
```

### Step 2: Remove AWSCC Resources from State (Don't Destroy)

```bash
# Remove Knowledge Base from state (keeps it in AWS)
terraform state rm awscc_bedrock_knowledge_base.main

# Remove Data Source from state (keeps it in AWS)
terraform state rm awscc_bedrock_data_source.s3

# Remove any S3 Vectors resources if present
terraform state rm aws_s3vectors_vector_bucket.vectors 2>/dev/null || true
terraform state rm aws_s3vectors_index.main 2>/dev/null || true

# Verify removal
terraform state list | grep -E "(awscc|s3vectors)"
# Should return no results
```

### Step 3: Update Module Code

Ensure the module code has been updated (tasks 2-8 completed).

### Step 4: Initialize with New Provider

```bash
# Upgrade providers
terraform init -upgrade

# Verify AWS provider version
terraform version
```

### Step 5: Import Existing Resources

```bash
# Source the migration IDs
source migration-ids.env

# Import Knowledge Base
terraform import aws_bedrockagent_knowledge_base.main ${KNOWLEDGE_BASE_ID}

# Import Data Source
terraform import aws_bedrockagent_data_source.s3 ${KNOWLEDGE_BASE_ID},${DATA_SOURCE_ID}

# Note: S3 buckets and vector index may need to be recreated
# Check the plan to see what needs to be imported vs recreated
```

### Step 6: Review Plan

```bash
# Generate plan
terraform plan -out=tfplan

# Carefully review the plan:
# - Resources with "~" will be updated in-place (good)
# - Resources with "-/+" will be replaced (may cause downtime)
# - Resources with "+" will be created (expected for new resources)

# Save plan for review
terraform show tfplan > migration-plan.txt
```

### Step 7: Apply Changes Carefully

```bash
# If plan looks acceptable, apply
terraform apply tfplan

# If plan shows unexpected replacements, STOP and:
# 1. Review the differences between old and new resource configurations
# 2. Adjust the Terraform code to match existing resource attributes
# 3. Re-run terraform plan until changes are minimal
```

### Step 8: Verify No Unexpected Changes

```bash
# Run plan again - should show no changes
terraform plan

# If changes are shown, investigate and resolve before proceeding
```

---

## Post-Migration Validation

Complete these steps after either migration option:

### 1. Verify Terraform State

```bash
cd terraform/environments/dev  # or your environment

# Verify all resources are in state
terraform state list

# Should include:
# - aws_s3_bucket.vectors
# - aws_bedrockagent_vector_index.main
# - aws_bedrockagent_knowledge_base.main
# - aws_bedrockagent_data_source.s3

# Verify no AWSCC resources remain
terraform state list | grep awscc
# Should return no results
```

### 2. Verify AWS Resources

```bash
# Get Knowledge Base ID
KB_ID=$(terraform output -raw knowledge_base_id)

# Check Knowledge Base configuration
aws bedrock-agent get-knowledge-base \
  --knowledge-base-id ${KB_ID} \
  --region ap-southeast-1 \
  --output json | jq '.knowledgeBase.storageConfiguration'

# Should show:
# {
#   "type": "S3_VECTORS",
#   "s3VectorsConfiguration": {
#     "vectorBucketArn": "arn:aws:s3:::...",
#     "indexArn": "arn:aws:bedrock:..."
#   }
# }

# Check Data Source
DATA_SOURCE_ID=$(terraform output -raw data_source_id)
aws bedrock-agent get-data-source \
  --knowledge-base-id ${KB_ID} \
  --data-source-id ${DATA_SOURCE_ID} \
  --region ap-southeast-1

# Check S3 buckets exist
aws s3 ls $(terraform output -raw documents_bucket_name)
aws s3 ls $(terraform output -raw vectors_bucket_name)

# Check vector index
INDEX_ARN=$(terraform output -raw index_arn)
echo "Vector Index ARN: ${INDEX_ARN}"
```

### 3. Test Document Ingestion

```bash
# Upload a test document
TEST_DOC="test-migration.txt"
echo "This is a test document for migration validation." > ${TEST_DOC}

DOCS_BUCKET=$(terraform output -raw documents_bucket_name)
aws s3 cp ${TEST_DOC} s3://${DOCS_BUCKET}/

# Trigger ingestion
cd ../../modules/knowledge-base/scripts
./ingest-kb.sh

# Wait for ingestion to complete (monitor script output)
```

### 4. Test Knowledge Base Queries

```bash
# Test retrieve operation
aws bedrock-agent-runtime retrieve \
  --knowledge-base-id ${KB_ID} \
  --retrieval-query text="test document migration" \
  --region ap-southeast-1 \
  --output json | jq '.retrievalResults'

# Should return results including the test document

# Test with multiple queries
aws bedrock-agent-runtime retrieve \
  --knowledge-base-id ${KB_ID} \
  --retrieval-query text="your domain-specific query" \
  --region ap-southeast-1
```

### 5. Verify Application Integration

```bash
cd ../../../../

# Verify .env configuration
grep BEDROCK_KNOWLEDGE_BASE_ID .env

# Start backend
cd backend
go run cmd/server/main.go &
BACKEND_PID=$!

# Wait for startup
sleep 5

# Test health endpoint
curl http://localhost:8080/health

# Test chat endpoint (adjust based on your API)
# curl -X POST http://localhost:8080/api/chat \
#   -H "Content-Type: application/json" \
#   -d '{"message": "test query"}'

# Stop backend
kill ${BACKEND_PID}
```

### 6. Monitor Costs

```bash
# Check S3 storage costs
aws s3 ls s3://$(terraform output -raw vectors_bucket_name) --recursive --summarize

# Expected: Much lower than OpenSearch Serverless (~$5-10/month vs ~$700/month)
```

### 7. Document Migration Results

```bash
# Create migration report
cat > migration-report-$(date +%Y%m%d).txt <<EOF
Migration Date: $(date)
Environment: $(basename $(pwd))
Migration Option: [Clean Slate / In-Place]

Pre-Migration:
- Old KB ID: ${OLD_KB_ID:-N/A}
- Old Provider: AWSCC

Post-Migration:
- New KB ID: $(terraform output -raw knowledge_base_id)
- New Provider: AWS $(terraform version | grep provider.registry.terraform.io/hashicorp/aws)
- Storage Type: S3_VECTORS

Validation Results:
- Terraform state: ✓
- AWS resources: ✓
- Document ingestion: ✓
- Query functionality: ✓
- Application integration: ✓

Issues Encountered:
[Document any issues and resolutions]

EOF

cat migration-report-$(date +%Y%m%d).txt
```

---

## Rollback Procedure

If migration fails or issues are discovered, follow these steps to rollback:

### Immediate Rollback (During Migration)

If you encounter issues during migration and need to abort:

```bash
# Stop any in-progress operations
# Press Ctrl+C if terraform apply is running

# Restore state from backup
cd terraform/environments/dev  # or your environment

# Find your backup file
ls -lt backup-state-*.json | head -1

# Restore state
terraform state push backup-state-YYYYMMDD-HHMMSS.json

# Verify state restoration
terraform state list

# Re-initialize with old provider
terraform init

# Verify plan shows no changes
terraform plan
```

### Rollback After Completed Migration

If issues are discovered after migration is complete:

#### Option A: Rollback Code and State

```bash
# 1. Restore Terraform state
cd terraform/environments/dev
terraform state push backup-state-YYYYMMDD-HHMMSS.json

# 2. Revert code changes
cd ../../modules/knowledge-base
git checkout HEAD~1 -- main.tf variables.tf outputs.tf

# 3. Re-initialize with old provider
cd ../../environments/dev
terraform init

# 4. Verify plan
terraform plan

# 5. If resources need to be recreated, apply
terraform apply
```

#### Option B: Destroy and Recreate from Backup

```bash
# 1. Destroy new resources
cd terraform/environments/dev
terraform destroy

# 2. Restore old code
cd ../../modules/knowledge-base
git checkout HEAD~1 -- main.tf variables.tf outputs.tf

# 3. Re-initialize
cd ../../environments/dev
terraform init

# 4. Recreate with old configuration
terraform apply

# 5. Restore documents
DOCS_BUCKET=$(terraform output -raw documents_bucket_name)
aws s3 sync backup-documents/ s3://${DOCS_BUCKET}/

# 6. Trigger ingestion
cd ../../modules/knowledge-base/scripts
./ingest-kb.sh
```

### Update Application Configuration

```bash
# Restore old .env file
cd ../../../../
cp .env.backup-YYYYMMDD-HHMMSS .env

# Verify configuration
grep BEDROCK_KNOWLEDGE_BASE_ID .env

# Restart application
cd backend
go run cmd/server/main.go
```

---

## Troubleshooting

### Common Issues and Solutions

#### Issue: Provider Version Not Available

**Error:**
```
Error: Failed to query available provider packages
Could not retrieve the list of available versions for provider hashicorp/aws
```

**Solution:**
```bash
# Clear provider cache
rm -rf .terraform/
rm .terraform.lock.hcl

# Re-initialize
terraform init -upgrade
```

#### Issue: Resource Already Exists

**Error:**
```
Error: Resource already exists
A resource with the ID "xxx" already exists
```

**Solution:**
```bash
# Import the existing resource
terraform import aws_bedrockagent_knowledge_base.main <knowledge-base-id>

# Or remove from AWS and let Terraform recreate
aws bedrock-agent delete-knowledge-base \
  --knowledge-base-id <knowledge-base-id> \
  --region ap-southeast-1
```

#### Issue: State Lock Error

**Error:**
```
Error: Error acquiring the state lock
```

**Solution:**
```bash
# Check who has the lock
aws dynamodb get-item \
  --table-name terraform-state-lock \
  --key '{"LockID":{"S":"your-state-file-path"}}'

# If lock is stale, force unlock (use with caution)
terraform force-unlock <lock-id>
```

#### Issue: Ingestion Fails After Migration

**Error:**
```
Error: Ingestion job failed with status: FAILED
```

**Solution:**
```bash
# Check IAM permissions
aws bedrock-agent get-knowledge-base \
  --knowledge-base-id ${KB_ID} \
  --region ap-southeast-1 \
  --query 'knowledgeBase.roleArn'

# Verify role has required permissions
aws iam get-role-policy \
  --role-name <role-name> \
  --policy-name <policy-name>

# Check CloudWatch logs for detailed error
aws logs tail /aws/bedrock/knowledgebases/${KB_ID} --follow
```

#### Issue: Queries Return No Results

**Possible Causes:**
1. Ingestion not complete
2. Vector index configuration mismatch
3. Embedding model dimension mismatch

**Solution:**
```bash
# Check ingestion status
aws bedrock-agent list-ingestion-jobs \
  --knowledge-base-id ${KB_ID} \
  --data-source-id ${DATA_SOURCE_ID} \
  --region ap-southeast-1

# Verify vector index configuration
terraform output index_arn

# Check embedding model dimensions
# Titan Embeddings G1: 1536 dimensions
# Cohere Embed: 1024 dimensions

# Re-trigger ingestion if needed
cd terraform/modules/knowledge-base/scripts
./ingest-kb.sh
```

---

## Best Practices

### Before Migration

1. **Test in Lower Environments First**
   - Always migrate dev → staging → production
   - Validate each environment before proceeding

2. **Schedule Maintenance Window**
   - Plan for 2-4 hours of downtime
   - Notify all stakeholders
   - Have rollback plan ready

3. **Document Everything**
   - Save all resource IDs
   - Backup all configurations
   - Keep detailed notes of each step

### During Migration

1. **Use Version Control**
   - Commit code changes before migration
   - Tag the pre-migration state
   - Create a migration branch

2. **Monitor Closely**
   - Watch Terraform output carefully
   - Check AWS Console for resource creation
   - Monitor CloudWatch logs

3. **Validate Each Step**
   - Don't rush through steps
   - Verify each resource before proceeding
   - Test functionality incrementally

### After Migration

1. **Monitor Performance**
   - Check query latency
   - Monitor error rates
   - Track costs

2. **Update Documentation**
   - Document new resource IDs
   - Update runbooks
   - Share lessons learned

3. **Clean Up**
   - Remove backup files after validation period
   - Archive migration logs
   - Update team documentation

---

## Additional Resources

- [AWS Provider Documentation](https://registry.terraform.io/providers/hashicorp/aws/latest/docs)
- [Bedrock Agent Resources](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/bedrockagent_knowledge_base)
- [S3 Vectors Documentation](https://docs.aws.amazon.com/bedrock/latest/userguide/knowledge-base-setup.html)
- [Terraform State Management](https://developer.hashicorp.com/terraform/language/state)

---

## Support

If you encounter issues not covered in this guide:

1. Check the [TROUBLESHOOTING.md](../../../TROUBLESHOOTING.md) file
2. Review Terraform and AWS provider documentation
3. Check CloudWatch logs for detailed error messages
4. Consult with your DevOps team or AWS support

---

**Last Updated:** December 2024  
**Terraform Version:** >= 1.5.0  
**AWS Provider Version:** >= 6.25.0
