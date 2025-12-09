#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
REGION="${AWS_REGION:-ap-southeast-1}"
ENVIRONMENT="${ENVIRONMENT:-dev}"
PROJECT_NAME="bedrock-chat-poc"

KB_NAME="${PROJECT_NAME}-kb-${ENVIRONMENT}"
DOCS_BUCKET="${PROJECT_NAME}-kb-docs-${ENVIRONMENT}"
VECTORS_BUCKET="${PROJECT_NAME}-kb-vectors-${ENVIRONMENT}"
IAM_ROLE="${PROJECT_NAME}-kb-role-${ENVIRONMENT}"

echo -e "${GREEN}=== Amazon Bedrock Knowledge Base Setup ===${NC}"
echo "Environment: $ENVIRONMENT"
echo "Region: $REGION"
echo ""

# Step 1: Deploy Terraform infrastructure
echo -e "${YELLOW}Step 1: Deploying Terraform infrastructure...${NC}"
cd "$(dirname "$0")/../../environments/${ENVIRONMENT}"
terraform apply -auto-approve

echo -e "${GREEN}✓ Infrastructure deployed${NC}"
echo ""

# Step 2: Create Knowledge Base
echo -e "${YELLOW}Step 2: Creating Knowledge Base...${NC}"

# Get IAM role ARN
ROLE_ARN=$(aws iam get-role \
  --role-name "$IAM_ROLE" \
  --region "$REGION" \
  --query 'Role.Arn' \
  --output text)

echo "Using IAM Role: $ROLE_ARN"

# Create Knowledge Base
KB_ID=$(aws bedrock-agent create-knowledge-base \
  --name "$KB_NAME" \
  --description "Knowledge base for chat POC (${ENVIRONMENT} environment)" \
  --role-arn "$ROLE_ARN" \
  --knowledge-base-configuration "type=VECTOR,vectorKnowledgeBaseConfiguration={embeddingModelArn=arn:aws:bedrock:${REGION}::foundation-model/amazon.titan-embed-text-v1}" \
  --storage-configuration "type=S3,s3Configuration={bucketArn=arn:aws:s3:::${VECTORS_BUCKET}}" \
  --region "$REGION" \
  --tags "Environment=${ENVIRONMENT},Project=${PROJECT_NAME},ManagedBy=CLI" \
  --query 'knowledgeBase.knowledgeBaseId' \
  --output text)

echo -e "${GREEN}✓ Knowledge Base created: $KB_ID${NC}"
echo ""

# Wait for Knowledge Base to be ready
echo "Waiting for Knowledge Base to be ready..."
sleep 5

# Step 3: Create Data Source
echo -e "${YELLOW}Step 3: Creating Data Source...${NC}"

DS_ID=$(aws bedrock-agent create-data-source \
  --knowledge-base-id "$KB_ID" \
  --name "${KB_NAME}-s3-data-source" \
  --data-source-configuration "type=S3,s3Configuration={bucketArn=arn:aws:s3:::${DOCS_BUCKET}}" \
  --region "$REGION" \
  --query 'dataSource.dataSourceId' \
  --output text)

echo -e "${GREEN}✓ Data Source created: $DS_ID${NC}"
echo ""

# Step 4: Update .env file
echo -e "${YELLOW}Step 4: Updating .env file...${NC}"

ENV_FILE="$(dirname "$0")/../../../.env"

if [ -f "$ENV_FILE" ]; then
  # Update existing entry or append
  if grep -q "^BEDROCK_KNOWLEDGE_BASE_ID=" "$ENV_FILE"; then
    sed -i.bak "s/^BEDROCK_KNOWLEDGE_BASE_ID=.*/BEDROCK_KNOWLEDGE_BASE_ID=${KB_ID}/" "$ENV_FILE"
  else
    echo "BEDROCK_KNOWLEDGE_BASE_ID=${KB_ID}" >> "$ENV_FILE"
  fi
else
  echo "BEDROCK_KNOWLEDGE_BASE_ID=${KB_ID}" > "$ENV_FILE"
fi

echo -e "${GREEN}✓ .env file updated${NC}"
echo ""

# Step 5: Upload test document
echo -e "${YELLOW}Step 5: Uploading test document...${NC}"

TEST_FILE=$(mktemp)
cat > "$TEST_FILE" << EOF
# Test Document for Knowledge Base

This is a test document to verify the knowledge base setup.

## Key Information
- Environment: ${ENVIRONMENT}
- Knowledge Base ID: ${KB_ID}
- Created: $(date)

The knowledge base is now ready to ingest documents and answer questions.
EOF

aws s3 cp "$TEST_FILE" "s3://${DOCS_BUCKET}/test-document.txt" --region "$REGION"
rm "$TEST_FILE"

echo -e "${GREEN}✓ Test document uploaded${NC}"
echo ""

# Step 6: Start ingestion job
echo -e "${YELLOW}Step 6: Starting ingestion job...${NC}"

JOB_ID=$(aws bedrock-agent start-ingestion-job \
  --knowledge-base-id "$KB_ID" \
  --data-source-id "$DS_ID" \
  --region "$REGION" \
  --query 'ingestionJob.ingestionJobId' \
  --output text)

echo "Ingestion Job ID: $JOB_ID"
echo "Waiting for ingestion to complete..."

# Poll ingestion status
MAX_ATTEMPTS=30
ATTEMPT=0
while [ $ATTEMPT -lt $MAX_ATTEMPTS ]; do
  STATUS=$(aws bedrock-agent get-ingestion-job \
    --knowledge-base-id "$KB_ID" \
    --data-source-id "$DS_ID" \
    --ingestion-job-id "$JOB_ID" \
    --region "$REGION" \
    --query 'ingestionJob.status' \
    --output text)
  
  if [ "$STATUS" = "COMPLETE" ]; then
    echo -e "${GREEN}✓ Ingestion completed successfully${NC}"
    break
  elif [ "$STATUS" = "FAILED" ]; then
    echo -e "${RED}✗ Ingestion failed${NC}"
    exit 1
  fi
  
  echo "Status: $STATUS (attempt $((ATTEMPT+1))/$MAX_ATTEMPTS)"
  sleep 10
  ATTEMPT=$((ATTEMPT+1))
done

if [ $ATTEMPT -eq $MAX_ATTEMPTS ]; then
  echo -e "${YELLOW}⚠ Ingestion still in progress. Check status manually.${NC}"
fi

echo ""
echo -e "${GREEN}=== Setup Complete! ===${NC}"
echo ""
echo "Knowledge Base Details:"
echo "  ID: $KB_ID"
echo "  Name: $KB_NAME"
echo "  Data Source ID: $DS_ID"
echo "  Documents Bucket: s3://${DOCS_BUCKET}"
echo "  Vectors Bucket: s3://${VECTORS_BUCKET}"
echo ""
echo "Next steps:"
echo "  1. Upload your documents to: s3://${DOCS_BUCKET}"
echo "  2. Run ingestion: ./ingest-kb.sh"
echo "  3. Start your application with the updated .env file"
echo ""
echo "Cost estimate: ~\$5-10/month (vs \$700/month for OpenSearch)"
