#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Configuration
REGION="${AWS_REGION:-ap-southeast-1}"
ENVIRONMENT="${ENVIRONMENT:-dev}"
PROJECT_NAME="bedrock-chat-poc"

KB_NAME="${PROJECT_NAME}-kb-${ENVIRONMENT}"
DOCS_BUCKET="${PROJECT_NAME}-kb-docs-${ENVIRONMENT}"
VECTORS_BUCKET="${PROJECT_NAME}-kb-vectors-${ENVIRONMENT}"

echo -e "${RED}=== Knowledge Base Cleanup ===${NC}"
echo "This will delete:"
echo "  - Knowledge Base: $KB_NAME"
echo "  - All data sources"
echo "  - S3 bucket contents"
echo "  - Terraform infrastructure"
echo ""
read -p "Are you sure? (yes/no): " CONFIRM

if [ "$CONFIRM" != "yes" ]; then
  echo "Cancelled"
  exit 0
fi

echo ""

# Get Knowledge Base ID
KB_ID=$(aws bedrock-agent list-knowledge-bases \
  --region "$REGION" \
  --query "knowledgeBaseSummaries[?name=='${KB_NAME}'].knowledgeBaseId | [0]" \
  --output text 2>/dev/null || echo "")

if [ -z "$KB_ID" ] || [ "$KB_ID" = "None" ]; then
  echo -e "${YELLOW}⚠ Knowledge Base not found, skipping deletion${NC}"
else
  echo "Knowledge Base ID: $KB_ID"
  
  # Get Data Source ID
  DS_ID=$(aws bedrock-agent list-data-sources \
    --knowledge-base-id "$KB_ID" \
    --region "$REGION" \
    --query 'dataSourceSummaries[0].dataSourceId' \
    --output text 2>/dev/null || echo "")
  
  # Delete Data Source
  if [ -n "$DS_ID" ] && [ "$DS_ID" != "None" ]; then
    echo -e "${YELLOW}Deleting data source...${NC}"
    aws bedrock-agent delete-data-source \
      --knowledge-base-id "$KB_ID" \
      --data-source-id "$DS_ID" \
      --region "$REGION" 2>/dev/null || true
    echo -e "${GREEN}✓ Data source deleted${NC}"
    sleep 5
  fi
  
  # Delete Knowledge Base
  echo -e "${YELLOW}Deleting knowledge base...${NC}"
  aws bedrock-agent delete-knowledge-base \
    --knowledge-base-id "$KB_ID" \
    --region "$REGION" 2>/dev/null || true
  echo -e "${GREEN}✓ Knowledge base deleted${NC}"
  sleep 5
fi

# Empty S3 buckets
echo -e "${YELLOW}Emptying S3 buckets...${NC}"

for BUCKET in "$DOCS_BUCKET" "$VECTORS_BUCKET"; do
  if aws s3 ls "s3://${BUCKET}" --region "$REGION" 2>/dev/null; then
    echo "Emptying: $BUCKET"
    aws s3 rm "s3://${BUCKET}" --recursive --region "$REGION" 2>/dev/null || true
    echo -e "${GREEN}✓ Bucket emptied: $BUCKET${NC}"
  fi
done

# Destroy Terraform infrastructure
echo -e "${YELLOW}Destroying Terraform infrastructure...${NC}"
cd "$(dirname "$0")/../../environments/${ENVIRONMENT}"
terraform destroy -auto-approve

echo ""
echo -e "${GREEN}=== Cleanup Complete! ===${NC}"
echo ""
echo "All resources have been deleted."
echo "Remember to remove BEDROCK_KNOWLEDGE_BASE_ID from your .env file"
