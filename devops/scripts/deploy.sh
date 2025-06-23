#!/bin/bash

# EcoCI Deployment Script
# This script deploys the EcoCI infrastructure and applications

set -e

# Configuration
ENVIRONMENT=${1:-staging}
AWS_REGION=${2:-us-west-2}
ACTION=${3:-plan}

PROJECT_NAME="ecoci"
TERRAFORM_DIR="devops/terraform/environments/${ENVIRONMENT}"

echo "EcoCI Deployment Script"
echo "Environment: ${ENVIRONMENT}"
echo "AWS Region: ${AWS_REGION}"
echo "Action: ${ACTION}"

# Check prerequisites
echo "Checking prerequisites..."

# Check tools
REQUIRED_TOOLS=("terraform" "aws" "kubectl" "helm")
for tool in "${REQUIRED_TOOLS[@]}"; do
    if ! command -v "$tool" &> /dev/null; then
        echo "Error: $tool is not installed"
        exit 1
    fi
done

# Check AWS credentials
if ! aws sts get-caller-identity &> /dev/null; then
    echo "Error: AWS credentials not configured"
    exit 1
fi

# Navigate to Terraform directory
if [[ ! -d "$TERRAFORM_DIR" ]]; then
    echo "Error: Terraform directory not found: $TERRAFORM_DIR"
    exit 1
fi

cd "$TERRAFORM_DIR"

# Initialize Terraform
echo "Initializing Terraform..."
terraform init

# Validate Terraform configuration
echo "Validating Terraform configuration..."
terraform validate

case "$ACTION" in
    "plan")
        echo "Running Terraform plan..."
        terraform plan
        ;;
    "apply")
        echo "Running Terraform apply..."
        terraform apply -auto-approve
        
        # Update kubeconfig after successful apply
        echo "Updating kubeconfig..."
        CLUSTER_NAME=$(terraform output -raw cluster_name)
        aws eks update-kubeconfig --region "$AWS_REGION" --name "$CLUSTER_NAME"
        
        echo "Deployment completed successfully!"
        echo ""
        echo "Cluster: $CLUSTER_NAME"
        echo "Region: $AWS_REGION"
        echo ""
        echo "Next steps:"
        echo "1. Update Kubernetes manifests with account ID and certificate ARN"
        echo "2. Deploy applications: kubectl apply -f ../../k8s/"
        echo "3. Check cluster status: kubectl get nodes"
        ;;
    "destroy")
        echo "WARNING: This will destroy all resources in ${ENVIRONMENT} environment!"
        read -p "Are you sure? (y/N): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            echo "Running Terraform destroy..."
            terraform destroy -auto-approve
            echo "Infrastructure destroyed."
        else
            echo "Destroy cancelled."
        fi
        ;;
    *)
        echo "Invalid action: $ACTION"
        echo "Valid actions: plan, apply, destroy"
        exit 1
        ;;
esac