# Security Module Outputs

output "github_oidc_provider_arn" {
  description = "ARN of the GitHub OIDC provider"
  value       = aws_iam_openid_connect_provider.github.arn
}

output "github_actions_role_arn" {
  description = "ARN of the GitHub Actions IAM role"
  value       = aws_iam_role.github_actions.arn
}

output "github_actions_role_name" {
  description = "Name of the GitHub Actions IAM role"
  value       = aws_iam_role.github_actions.name
}

output "artifacts_bucket_name" {
  description = "Name of the S3 bucket for build artifacts"
  value       = aws_s3_bucket.artifacts.bucket
}

output "artifacts_bucket_arn" {
  description = "ARN of the S3 bucket for build artifacts"
  value       = aws_s3_bucket.artifacts.arn
}

output "alb_security_group_id" {
  description = "Security group ID for Application Load Balancer"
  value       = aws_security_group.alb.id
}

output "alb_security_group_arn" {
  description = "Security group ARN for Application Load Balancer"
  value       = aws_security_group.alb.arn
}

output "app_security_group_id" {
  description = "Security group ID for applications"
  value       = aws_security_group.app.id
}

output "app_security_group_arn" {
  description = "Security group ARN for applications"
  value       = aws_security_group.app.arn
}

output "ecr_auth_api_repository_url" {
  description = "URL of the ECR repository for auth-api"
  value       = aws_ecr_repository.auth_api.repository_url
}

output "ecr_auth_api_repository_arn" {
  description = "ARN of the ECR repository for auth-api"
  value       = aws_ecr_repository.auth_api.arn
}

output "ecr_badge_service_repository_url" {
  description = "URL of the ECR repository for badge-service"
  value       = aws_ecr_repository.badge_service.repository_url
}

output "ecr_badge_service_repository_arn" {
  description = "ARN of the ECR repository for badge-service"
  value       = aws_ecr_repository.badge_service.arn
}