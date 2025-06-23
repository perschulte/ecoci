# EcoCI DevOps Infrastructure

This directory contains the Infrastructure as Code (IaC), CI/CD pipelines, and deployment configurations for the EcoCI project.

## Architecture Overview

EcoCI uses a cloud-native architecture deployed on AWS with the following components:

- **AWS EKS**: Managed Kubernetes cluster for container orchestration
- **Route53**: DNS management and domain routing
- **ACM**: TLS certificate management
- **ECR**: Container image registry
- **S3**: Artifact storage for builds
- **VPC**: Network isolation and security

## Directory Structure

```
devops/
├── terraform/                 # Infrastructure as Code
│   ├── modules/              # Reusable Terraform modules
│   │   ├── vpc/              # VPC and networking
│   │   ├── eks/              # EKS cluster configuration
│   │   ├── dns/              # Route53 and ACM
│   │   └── security/         # IAM, security groups, OIDC
│   ├── environments/         # Environment-specific configurations
│   │   └── staging/          # Staging environment
│   ├── main.tf               # Root configuration
│   ├── variables.tf          # Input variables
│   └── outputs.tf            # Output values
├── k8s/                      # Kubernetes manifests
│   ├── base/                 # Base resources (namespaces, RBAC)
│   └── staging/              # Staging-specific deployments
├── .github/workflows/        # CI/CD pipelines
│   ├── cli-build.yml         # CLI wheel building
│   ├── container-build.yml   # Container image building
│   └── deploy-staging.yml    # Staging deployment
└── README.md                 # This file
```

## Prerequisites

Before deploying the infrastructure, ensure you have:

1. **AWS Account** with appropriate permissions
2. **Domain**: `ecoci.dev` domain configured in Route53 (or your preferred domain)
3. **GitHub Repository**: With OIDC provider configured
4. **Terraform State Backend**: S3 bucket and DynamoDB table for state management

### Required Tools

- [Terraform](https://www.terraform.io/downloads.html) >= 1.5
- [AWS CLI](https://aws.amazon.com/cli/) >= 2.0
- [kubectl](https://kubernetes.io/docs/tasks/tools/) >= 1.28
- [Helm](https://helm.sh/docs/intro/install/) >= 3.12

## Deployment Guide

### 1. Set Up Terraform Backend

Create the S3 bucket and DynamoDB table for Terraform state:

```bash
# Create S3 bucket for Terraform state
aws s3 mb s3://ecoci-staging-terraform-state --region us-west-2

# Enable versioning
aws s3api put-bucket-versioning \\
  --bucket ecoci-staging-terraform-state \\
  --versioning-configuration Status=Enabled

# Create DynamoDB table for state locking
aws dynamodb create-table \\
  --table-name ecoci-staging-terraform-locks \\
  --attribute-definitions AttributeName=LockID,AttributeType=S \\
  --key-schema AttributeName=LockID,KeyType=HASH \\
  --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5 \\
  --region us-west-2
```

### 2. Configure GitHub Secrets

Set up the following secrets in your GitHub repository:

- `AWS_GITHUB_ACTIONS_ROLE_ARN`: ARN of the GitHub Actions IAM role
- `ARTIFACTS_BUCKET`: Name of the S3 bucket for build artifacts

### 3. Deploy Infrastructure

```bash
cd devops/terraform/environments/staging

# Initialize Terraform
terraform init

# Plan the deployment
terraform plan

# Apply the infrastructure
terraform apply
```

### 4. Update Kubernetes Manifests

After infrastructure deployment, update the placeholder values in the Kubernetes manifests:

```bash
# Get AWS Account ID
ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)

# Get Certificate ARN
CERT_ARN=$(terraform output -raw acm_certificate_arn)

# Update Kubernetes manifests
sed -i "s/ACCOUNT_ID/$ACCOUNT_ID/g" ../k8s/staging/*.yaml
sed -i "s/CERTIFICATE_ARN/$CERT_ARN/g" ../k8s/staging/*.yaml
```

### 5. Configure kubectl

```bash
# Update kubeconfig
aws eks update-kubeconfig --region us-west-2 --name ecoci-staging-eks
```

### 6. Deploy Applications

The applications will be automatically deployed via GitHub Actions when changes are pushed to the main branch. You can also manually deploy using:

```bash
# Apply base resources
kubectl apply -f devops/k8s/base/

# Apply staging resources
kubectl apply -f devops/k8s/staging/
```

## CI/CD Pipelines

### CLI Build Pipeline (`cli-build.yml`)

Triggers: Changes to `cli/**` directory

Steps:
1. Lint and test Python code
2. Run security scans (Bandit, Safety)
3. Build Python wheel
4. Upload to S3 artifacts bucket
5. Create GitHub release (on tags)

### Container Build Pipeline (`container-build.yml`)

Triggers: Changes to `auth-api/**` or `badge-service/**` directories

Steps:
1. Detect changed services
2. Build and push Docker images to ECR
3. Run vulnerability scans with Trivy
4. Upload security results to GitHub Security tab

### Staging Deployment Pipeline (`deploy-staging.yml`)

Triggers: Push to main branch

Steps:
1. Run Terraform plan (on PRs)
2. Deploy Kubernetes manifests
3. Install/upgrade AWS Load Balancer Controller
4. Deploy applications with health checks
5. Run smoke tests

## Monitoring and Observability

The infrastructure includes Grafana Agent for metrics and logs collection:

- **Metrics**: Scraped from application `/metrics` endpoints
- **Logs**: Collected from Kubernetes pods
- **Grafana Cloud**: Ready for integration (credentials required)

### Setting up Grafana Cloud

1. Create a Grafana Cloud account
2. Get your Prometheus and Loki credentials
3. Create a Kubernetes secret:

```bash
kubectl create secret generic grafana-cloud-credentials \\
  --namespace ecoci-staging \\
  --from-literal=prometheus-user=YOUR_PROMETHEUS_USER \\
  --from-literal=prometheus-api-key=YOUR_PROMETHEUS_API_KEY \\
  --from-literal=loki-user=YOUR_LOKI_USER \\
  --from-literal=loki-api-key=YOUR_LOKI_API_KEY
```

## Cost Optimization

The staging environment is configured for cost optimization:

- **Single NAT Gateway**: Shared across availability zones
- **t3.medium instances**: Right-sized for development workloads
- **Single node cluster**: Minimum viable setup
- **ECR lifecycle policies**: Automatic cleanup of old images
- **S3 lifecycle policies**: Automatic cleanup of old artifacts

## Security Features

- **OIDC Authentication**: GitHub Actions uses workload identity (no long-lived secrets)
- **Least Privilege IAM**: Minimal required permissions
- **Network Security**: Private subnets for worker nodes
- **Container Security**: Non-root containers, read-only filesystems
- **Vulnerability Scanning**: Automated security scans in CI/CD
- **TLS Everywhere**: ACM certificates for all endpoints

## Troubleshooting

### Common Issues

1. **Terraform State Lock**: If Terraform is stuck, check DynamoDB table and remove stale locks
2. **EKS Access**: Ensure your AWS credentials have EKS access and the cluster is in the correct region
3. **Load Balancer**: AWS Load Balancer Controller needs proper IAM permissions
4. **DNS Resolution**: Ensure Route53 hosted zone is properly configured

### Useful Commands

```bash
# Check cluster status
kubectl get nodes
kubectl get pods --all-namespaces

# View application logs
kubectl logs -f deployment/auth-api -n ecoci-staging
kubectl logs -f deployment/badge-service -n ecoci-staging

# Check ingress status
kubectl get ingress -n ecoci-staging

# Debug load balancer controller
kubectl logs -f deployment/aws-load-balancer-controller -n kube-system
```

## Cleanup

To destroy the infrastructure:

```bash
cd devops/terraform/environments/staging
terraform destroy
```

**Warning**: This will delete all resources including data. Make sure to backup any important data before destroying.

## Next Steps

- Set up production environment
- Implement blue-green deployments
- Add automated testing in staging
- Configure alerts and monitoring dashboards
- Implement disaster recovery procedures