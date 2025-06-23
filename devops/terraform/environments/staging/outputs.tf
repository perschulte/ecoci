# EcoCI Staging Environment Outputs

# VPC Outputs
output "vpc_id" {
  description = "ID of the VPC"
  value       = module.ecoci_staging.vpc_id
}

output "private_subnet_ids" {
  description = "List of IDs of private subnets" 
  value       = module.ecoci_staging.private_subnet_ids
}

output "public_subnet_ids" {
  description = "List of IDs of public subnets"
  value       = module.ecoci_staging.public_subnet_ids
}

# EKS Outputs
output "cluster_name" {
  description = "EKS cluster name"
  value       = module.ecoci_staging.eks_cluster_name
}

output "cluster_endpoint" {
  description = "EKS cluster API server endpoint"
  value       = module.ecoci_staging.eks_cluster_endpoint
  sensitive   = true
}

output "cluster_ca_certificate" {
  description = "Base64 encoded certificate data required to communicate with the cluster"
  value       = module.ecoci_staging.eks_cluster_certificate_authority_data
  sensitive   = true
}

# DNS Outputs
output "domain_zone_id" {
  description = "Route53 hosted zone ID"
  value       = module.ecoci_staging.route53_zone_id
}

output "domain_name" {
  description = "Domain name for staging environment"
  value       = module.ecoci_staging.route53_zone_name
}

output "certificate_arn" {
  description = "ARN of the ACM certificate"
  value       = module.ecoci_staging.acm_certificate_arn
}

# Security Outputs
output "github_actions_role_arn" {
  description = "ARN of the GitHub Actions IAM role"
  value       = module.ecoci_staging.github_actions_role_arn
}

# ECR Repository URLs (for CI/CD)
output "auth_api_ecr_url" {
  description = "ECR repository URL for auth-api"
  value       = module.ecoci_staging.ecr_auth_api_repository_url
}

output "badge_service_ecr_url" {
  description = "ECR repository URL for badge-service"
  value       = module.ecoci_staging.ecr_badge_service_repository_url
}

# Update kubeconfig command
output "kubectl_config_command" {
  description = "Command to update kubeconfig for this cluster"
  value       = "aws eks update-kubeconfig --region us-west-2 --name ${module.ecoci_staging.eks_cluster_name}"
}