# EcoCI Infrastructure Variables

variable "aws_region" {
  description = "AWS region for infrastructure deployment"
  type        = string
  default     = "us-west-2"
}

variable "environment" {
  description = "Environment name (staging, production)"
  type        = string
  validation {
    condition     = contains(["staging", "production"], var.environment)
    error_message = "Environment must be either 'staging' or 'production'."
  }
}

variable "project_name" {
  description = "Name of the project"
  type        = string
  default     = "ecoci"
}

variable "domain_name" {
  description = "Base domain name for the project"
  type        = string
  default     = "ecoci.dev"
}

variable "vpc_cidr" {
  description = "CIDR block for VPC"
  type        = string
  default     = "10.0.0.0/16"
}

variable "availability_zones" {
  description = "Availability zones to use for subnets"
  type        = list(string)
  default     = ["us-west-2a", "us-west-2b"]
}

variable "eks_cluster_version" {
  description = "Kubernetes version for EKS cluster"
  type        = string
  default     = "1.28"
}

variable "eks_node_instance_types" {
  description = "EC2 instance types for EKS worker nodes"
  type        = list(string)
  default     = ["t3.medium"]
}

variable "eks_node_desired_capacity" {
  description = "Desired number of worker nodes"
  type        = number
  default     = 2
}

variable "eks_node_max_capacity" {
  description = "Maximum number of worker nodes"
  type        = number
  default     = 4
}

variable "eks_node_min_capacity" {
  description = "Minimum number of worker nodes"
  type        = number
  default     = 1
}

variable "enable_nat_gateway" {
  description = "Should be true to provision NAT Gateway"
  type        = bool
  default     = true
}

variable "single_nat_gateway" {
  description = "Should be true to provision single NAT Gateway for all private subnets"
  type        = bool
  default     = true
}

variable "github_org" {
  description = "GitHub organization name for OIDC provider"
  type        = string
  default     = "ecoci-org"
}

variable "github_repo" {
  description = "GitHub repository name for OIDC provider"
  type        = string
  default     = "ecoci"
}

variable "tags" {
  description = "Additional tags to apply to all resources"
  type        = map(string)
  default     = {}
}