# EcoCI Staging Environment Configuration

terraform {
  backend "s3" {
    bucket = "ecoci-staging-terraform-state"
    key    = "staging/terraform.tfstate"
    region = "us-west-2"
    
    # Enable state locking
    dynamodb_table = "ecoci-staging-terraform-locks"
    encrypt        = true
  }
}

module "ecoci_staging" {
  source = "../.."

  # Environment Configuration
  environment         = "staging"
  project_name       = "ecoci"
  aws_region         = "us-west-2"
  domain_name        = "ecoci.dev"

  # Network Configuration
  vpc_cidr           = "10.0.0.0/16"
  availability_zones = ["us-west-2a", "us-west-2b"]
  enable_nat_gateway = true
  single_nat_gateway = true  # Cost optimization for staging

  # EKS Configuration (cost-optimized for staging)
  eks_cluster_version     = "1.28"
  eks_node_instance_types = ["t3.medium"]
  eks_node_desired_capacity = 1    # Single node for staging
  eks_node_max_capacity     = 2
  eks_node_min_capacity     = 1

  # GitHub Configuration
  github_org  = "ecoci-org"
  github_repo = "ecoci"

  # Additional tags
  tags = {
    Environment   = "staging"
    Project       = "EcoCI"
    CostCenter   = "Engineering"
    Owner        = "DevOps"
    Terraform    = "true"
  }
}