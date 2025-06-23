# EcoCI Infrastructure Outputs

output "vpc_id" {
  description = "ID of the VPC"
  value       = module.vpc.vpc_id
}

output "vpc_cidr_block" {
  description = "CIDR block of the VPC"
  value       = module.vpc.vpc_cidr_block
}

output "private_subnet_ids" {
  description = "List of IDs of private subnets"
  value       = module.vpc.private_subnet_ids
}

output "public_subnet_ids" {
  description = "List of IDs of public subnets"
  value       = module.vpc.public_subnet_ids
}

output "eks_cluster_id" {
  description = "EKS cluster ID"
  value       = module.eks.cluster_id
}

output "eks_cluster_arn" {
  description = "EKS cluster ARN"
  value       = module.eks.cluster_arn
}

output "eks_cluster_endpoint" {
  description = "EKS cluster API server endpoint"
  value       = module.eks.cluster_endpoint
}

output "eks_cluster_name" {
  description = "EKS cluster name"
  value       = module.eks.cluster_name
}

output "eks_cluster_certificate_authority_data" {
  description = "Base64 encoded certificate data required to communicate with the cluster"
  value       = module.eks.cluster_certificate_authority_data
}

output "eks_cluster_security_group_id" {
  description = "Security group ID attached to the EKS cluster"
  value       = module.eks.cluster_security_group_id
}

output "eks_node_groups" {
  description = "EKS node groups"
  value       = module.eks.eks_managed_node_groups
}

output "route53_zone_id" {
  description = "Route53 hosted zone ID"
  value       = module.dns.zone_id
}

output "route53_zone_name" {
  description = "Route53 hosted zone name"
  value       = module.dns.zone_name
}

output "acm_certificate_arn" {
  description = "ARN of the ACM certificate"
  value       = module.dns.certificate_arn
}

output "github_oidc_provider_arn" {
  description = "ARN of the GitHub OIDC provider"
  value       = module.security.github_oidc_provider_arn
}

output "github_actions_role_arn" {
  description = "ARN of the GitHub Actions IAM role"
  value       = module.security.github_actions_role_arn
}

output "alb_security_group_id" {
  description = "Security group ID for Application Load Balancer"
  value       = module.security.alb_security_group_id
}