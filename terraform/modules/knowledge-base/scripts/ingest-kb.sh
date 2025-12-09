#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Get Terraform outputs
cd "$(dirname "$0")/../../../environments/${ENVIRONMENT:-dev}"

KB_ID=$(terraform output -raw knowledge_base_id 2>/dev/null)
DS_ID=$(terraform output -raw data_source_id 2>/dev/null)
REGION=$(terraform output -raw aws_region 2>/dev/null || echo "ap-southeast-1")

if [ -z "$KB_ID" ] || [ "$KB_ID" = "" ]; then
  echo -e "${RED}✗ Knowledge Base not found${NC}"
  echo "Run 'terraform apply' first"
  exit 1
fi

echo -e "${GREEN}=== Knowledge Base Ingestion ===${NC}"
echo "Knowledge Base ID: $KB_ID"
echo "Data Source ID: $DS_ID"
echo "Region: $REGION"
echo ""

# Start ingestion job
echo -e "${YELLOW}Starting ingestion job...${NC}"

JOB_ID=$(aws bedrock-agent start-ingestion-job \
  --knowledge-base-id "$KB_ID" \
  --data-source-id "$DS_ID" \
  --region "$REGION" \
  --query 'ingestionJob.ingestionJobId' \
  --output text)

echo "Job ID: $JOB_ID"
echo ""

# Monitor progress
echo "Monitoring ingestion progress..."
MAX_ATTEMPTS=60
ATTEMPT=0

while [ $ATTEMPT -lt $MAX_ATTEMPTS ]; do
  RESPONSE=$(aws bedrock-agent get-ingestion-job \
    --knowledge-base-id "$KB_ID" \
    --data-source-id "$DS_ID" \
    --ingestion-job-id "$JOB_ID" \
    --region "$REGION" \
    --output json)
  
  STATUS=$(echo "$RESPONSE" | jq -r '.ingestionJob.status')
  
  if [ "$STATUS" = "COMPLETE" ]; then
    STATS=$(echo "$RESPONSE" | jq -r '.ingestionJob.statistics')
    echo ""
    echo -e "${GREEN}✓ Ingestion completed successfully${NC}"
    echo ""
    echo "Statistics:"
    echo "$STATS" | jq '.'
    exit 0
  elif [ "$STATUS" = "FAILED" ]; then
    echo ""
    echo -e "${RED}✗ Ingestion failed${NC}"
    echo "$RESPONSE" | jq -r '.ingestionJob.failureReasons[]?'
    exit 1
  fi
  
  echo -ne "\rStatus: $STATUS (${ATTEMPT}s elapsed)"
  sleep 5
  ATTEMPT=$((ATTEMPT+5))
done

echo ""
echo -e "${YELLOW}⚠ Ingestion still in progress after ${MAX_ATTEMPTS}s${NC}"
echo "Check status manually:"
echo "  aws bedrock-agent get-ingestion-job \\"
echo "    --knowledge-base-id $KB_ID \\"
echo "    --data-source-id $DS_ID \\"
echo "    --ingestion-job-id $JOB_ID \\"
echo "    --region $REGION"
