# Knowledge Base Ingestion Test Results

## Test Execution Date
December 9, 2025

## Test Objective
Test the document ingestion workflow for the S3 Vectors-based Knowledge Base according to task 13 requirements.

## Test Steps Performed

### 1. Test Document Upload ✅
- **Status**: SUCCESS
- **Action**: Created and uploaded test document to S3 documents bucket
- **Document**: `test-document.txt` (1,491 bytes)
- **Bucket**: `bedrock-chat-poc-kb-docs-dev`
- **Verification**: Document successfully uploaded and visible in S3

```bash
aws s3 ls s3://bedrock-chat-poc-kb-docs-dev/
# Output: 2025-12-10 00:03:39       1491 test-document.txt
```

### 2. Ingestion Script Execution ❌
- **Status**: FAILED
- **Action**: Attempted to run ingestion script
- **Script**: `terraform/modules/knowledge-base/scripts/ingest-kb.sh`
- **Error**: ValidationException from AWS Bedrock

## Error Details

### Error Message
```
An error occurred (ValidationException) when calling the StartIngestionJob operation: 
Knowledge base role arn:aws:iam::315183407016:role/bedrock-chat-poc-kb-role-dev is not 
able to call specified bedrock embedding model 
arn:aws:bedrock:ap-southeast-1::foundation-model/amazon.titan-embed-text-v1: 
The provided model identifier is invalid. 
(Service: BedrockRuntime, Status Code: 400, Request ID: eac84393-61e1-47da-9799-63ad6a3cb8c2)
```

### Root Cause Analysis

**Issue**: The Knowledge Base is configured to use `amazon.titan-embed-text-v1` embedding model, but this model is **not available** in the `ap-southeast-1` region.

**Evidence**:
1. Listing available embedding models in ap-southeast-1:
   ```bash
   aws bedrock list-foundation-models --region ap-southeast-1 \
     --output json | jq -r '.modelSummaries[] | select(.modelId | contains("embed")) | .modelId'
   ```
   **Result**: Only Cohere models available:
   - `cohere.embed-v4:0`
   - `cohere.embed-english-v3`
   - `cohere.embed-multilingual-v3`

2. Attempting to get Titan model details:
   ```bash
   aws bedrock get-foundation-model --model-identifier amazon.titan-embed-text-v1 \
     --region ap-southeast-1
   ```
   **Result**: `ValidationException: The provided model identifier is invalid.`

3. Current Knowledge Base configuration:
   ```json
   {
     "embeddingModelArn": "arn:aws:bedrock:ap-southeast-1::foundation-model/amazon.titan-embed-text-v1"
   }
   ```

### Impact on Requirements

**Requirements 8.1-8.4** (Document Ingestion Workflow):
- ❌ 8.1: Cannot upload and trigger ingestion due to invalid model configuration
- ❌ 8.2: Ingestion job cannot start due to model validation failure
- ❌ 8.3: Cannot monitor ingestion progress (job never starts)
- ❌ 8.4: Cannot verify statistics (ingestion blocked)

## Infrastructure Configuration Issues

### Current State
- **Knowledge Base ID**: HVKLKEBTZT
- **Data Source ID**: ZOIDPA3WXB
- **Region**: ap-southeast-1
- **Configured Model**: amazon.titan-embed-text-v1 (INVALID for this region)
- **Storage Type**: S3_VECTORS
- **Status**: ACTIVE (but non-functional for ingestion)

### IAM Permissions
The IAM role has correct permissions:
```json
{
  "Sid": "BedrockInvokeModel",
  "Effect": "Allow",
  "Action": ["bedrock:InvokeModel"],
  "Resource": "arn:aws:bedrock:ap-southeast-1::foundation-model/amazon.titan-embed-text-v1"
}
```
However, the resource ARN points to a non-existent model in this region.

## Resolution Options

### Option 1: Update Knowledge Base to Use Cohere Model (Recommended)
Update the Terraform configuration to use an available embedding model:

**Changes Required**:
1. Update `terraform/modules/knowledge-base/main.tf`:
   ```hcl
   embedding_model_arn = "arn:aws:bedrock:${data.aws_region.current.id}::foundation-model/cohere.embed-english-v3"
   ```

2. Update vector index dimensions (Cohere uses 1024 dimensions vs Titan's 1536):
   ```hcl
   dimension = 1024  # Changed from 1536
   ```

3. Update IAM policy resource ARN:
   ```hcl
   Resource = "arn:aws:bedrock:${data.aws_region.current.id}::foundation-model/cohere.embed-english-v3"
   ```

4. Apply Terraform changes:
   ```bash
   cd terraform/environments/dev
   terraform apply
   ```

**Note**: This will require recreating the Knowledge Base and re-ingesting all documents.

### Option 2: Use a Different AWS Region
Deploy the infrastructure in a region where Titan models are available:
- us-east-1 (N. Virginia)
- us-west-2 (Oregon)
- eu-west-1 (Ireland)

**Changes Required**:
1. Update `terraform/environments/dev/terraform.tfvars`:
   ```hcl
   aws_region = "us-east-1"
   ```

2. Re-deploy infrastructure:
   ```bash
   terraform destroy
   terraform apply
   ```

### Option 3: Request Model Access (If Available)
Check if Titan models can be enabled in ap-southeast-1:
1. Go to AWS Bedrock Console → Model access
2. Request access to Amazon Titan Embedding models
3. Wait for approval (if available in region)

**Note**: Based on API responses, Titan models appear to be fundamentally unavailable in ap-southeast-1, not just requiring access approval.

## Recommendations

### Immediate Action
**Update the Knowledge Base configuration to use Cohere embedding model** (Option 1) because:
1. Cohere models are readily available in ap-southeast-1
2. Cohere Embed English v3 provides good quality embeddings (1024 dimensions)
3. No region change required
4. Minimal infrastructure changes needed

### Implementation Steps
1. Update Terraform configuration files as described in Option 1
2. Run `terraform plan` to review changes
3. Run `terraform apply` to update infrastructure
4. Re-upload test document
5. Re-run ingestion script
6. Verify ingestion completes successfully

### Testing After Fix
Once the model configuration is corrected, re-run this test:
```bash
# 1. Upload test document
aws s3 cp test-document.txt s3://bedrock-chat-poc-kb-docs-dev/

# 2. Run ingestion
cd terraform/modules/knowledge-base/scripts
ENVIRONMENT=dev ./ingest-kb.sh

# 3. Verify vectors in S3
aws s3 ls s3://bedrock-chat-poc-kb-vectors-dev/ --recursive

# 4. Test query
aws bedrock-agent-runtime retrieve \
  --knowledge-base-id HVKLKEBTZT \
  --retrieval-query text="What are the benefits of S3 Vectors?" \
  --region ap-southeast-1
```

## Conclusion

The ingestion test **cannot be completed** with the current infrastructure configuration due to an invalid embedding model for the ap-southeast-1 region. The Knowledge Base was created with a Titan embedding model ARN that is not available in this region.

**Task Status**: BLOCKED - Requires infrastructure configuration update before ingestion can be tested.

**Next Steps**: 
1. Update Terraform configuration to use Cohere embedding model
2. Apply infrastructure changes
3. Re-test ingestion workflow

## Related Files
- Test document: `test-document.txt`
- Ingestion script: `terraform/modules/knowledge-base/scripts/ingest-kb.sh`
- Terraform config: `terraform/modules/knowledge-base/main.tf`
- Environment config: `terraform/environments/dev/terraform.tfvars`
