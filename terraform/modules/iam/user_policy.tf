# IAM Policy for User Access to Bedrock Agent Services
# This policy allows users (like flamedev) to invoke Bedrock Agents and access Knowledge Bases

# Data source to get the current user (flamedev)
data "aws_iam_user" "bedrock_user" {
  user_name = "flamedev"
}

# IAM Policy for Bedrock Agent Runtime Access
resource "aws_iam_policy" "bedrock_agent_user_policy" {
  name        = "${var.project_name}-bedrock-user-policy-${var.environment}"
  description = "Policy for user access to Bedrock Agent Runtime services"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "bedrock-agent-runtime:InvokeAgent",
          "bedrock-agent-runtime:InvokeAgentStream",
          "bedrock-agent-runtime:CreateSession",
          "bedrock-agent-runtime:GetSession",
          "bedrock-agent-runtime:DeleteSession"
        ]
        Resource = [
          "arn:aws:bedrock:${local.region}:${local.account_id}:agent/${var.bedrock_agent_id}",
          "arn:aws:bedrock:${local.region}:${local.account_id}:agent-alias/${var.bedrock_agent_id}/${var.bedrock_agent_alias_id}"
        ]
      },
      {
        Effect = "Allow"
        Action = [
          "bedrock:Retrieve",
          "bedrock:RetrieveAndGenerate"
        ]
        Resource = [
          "arn:aws:bedrock:${local.region}:${local.account_id}:knowledge-base/${var.bedrock_knowledge_base_id}"
        ]
      },
      {
        Effect = "Allow"
        Action = [
          "bedrock:InvokeModel",
          "bedrock:InvokeModelWithResponseStream"
        ]
        Resource = [
          "arn:aws:bedrock:${local.region}::foundation-model/${var.foundation_model_id}"
        ]
      }
    ]
  })

  tags = merge(
    var.tags,
    {
      Name = "${var.project_name}-bedrock-user-policy-${var.environment}"
    }
  )
}

# Attach the policy to the user
resource "aws_iam_user_policy_attachment" "bedrock_user_policy_attachment" {
  user       = data.aws_iam_user.bedrock_user.user_name
  policy_arn = aws_iam_policy.bedrock_agent_user_policy.arn
}