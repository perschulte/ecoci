# EKS Module Outputs

output "cluster_id" {
  description = "EKS cluster ID"
  value       = aws_eks_cluster.main.cluster_id
}

output "cluster_arn" {
  description = "EKS cluster ARN"
  value       = aws_eks_cluster.main.arn
}

output "cluster_name" {
  description = "EKS cluster name"
  value       = aws_eks_cluster.main.name
}

output "cluster_endpoint" {
  description = "EKS cluster API server endpoint"
  value       = aws_eks_cluster.main.endpoint
}

output "cluster_version" {
  description = "EKS cluster Kubernetes version"
  value       = aws_eks_cluster.main.version
}

output "cluster_platform_version" {
  description = "EKS cluster platform version"
  value       = aws_eks_cluster.main.platform_version
}

output "cluster_certificate_authority_data" {
  description = "Base64 encoded certificate data required to communicate with the cluster"
  value       = aws_eks_cluster.main.certificate_authority[0].data
}

output "cluster_security_group_id" {
  description = "Security group ID attached to the EKS cluster"
  value       = aws_eks_cluster.main.vpc_config[0].cluster_security_group_id
}

output "cluster_iam_role_name" {
  description = "IAM role name associated with EKS cluster"
  value       = aws_iam_role.eks_cluster.name
}

output "cluster_iam_role_arn" {
  description = "IAM role ARN associated with EKS cluster"
  value       = aws_iam_role.eks_cluster.arn
}

output "cluster_oidc_issuer_url" {
  description = "The URL on the EKS cluster OIDC Issuer"
  value       = aws_eks_cluster.main.identity[0].oidc[0].issuer
}

output "cluster_primary_security_group_id" {
  description = "The cluster primary security group ID created by EKS"
  value       = aws_eks_cluster.main.vpc_config[0].cluster_security_group_id
}

output "eks_managed_node_groups" {
  description = "Map of attribute maps for all EKS managed node groups created"
  value = {
    main = {
      node_group_name = aws_eks_node_group.main.node_group_name
      node_group_arn  = aws_eks_node_group.main.arn
      node_group_status = aws_eks_node_group.main.status
      capacity_type   = aws_eks_node_group.main.capacity_type
      instance_types  = aws_eks_node_group.main.instance_types
      ami_type       = aws_eks_node_group.main.ami_type
      node_role_arn  = aws_eks_node_group.main.node_role_arn
      scaling_config = aws_eks_node_group.main.scaling_config
    }
  }
}

output "node_security_group_id" {
  description = "ID of the EKS node shared security group"
  value       = aws_security_group.eks_nodes.id
}

output "node_security_group_arn" {
  description = "ARN of the EKS node shared security group"
  value       = aws_security_group.eks_nodes.arn
}

output "oidc_provider_arn" {
  description = "The ARN of the OIDC Identity Provider if enabled"
  value       = var.enable_irsa ? aws_iam_openid_connect_provider.eks.arn : null
}

output "cluster_addons" {
  description = "Map of attribute maps for all EKS cluster addons enabled"
  value = {
    coredns = {
      addon_name    = aws_eks_addon.coredns.addon_name
      addon_version = aws_eks_addon.coredns.addon_version
      status        = aws_eks_addon.coredns.status
    }
    kube-proxy = {
      addon_name    = aws_eks_addon.kube_proxy.addon_name
      addon_version = aws_eks_addon.kube_proxy.addon_version
      status        = aws_eks_addon.kube_proxy.status
    }
    vpc-cni = {
      addon_name    = aws_eks_addon.vpc_cni.addon_name
      addon_version = aws_eks_addon.vpc_cni.addon_version
      status        = aws_eks_addon.vpc_cni.status
    }
    aws-ebs-csi-driver = {
      addon_name    = aws_eks_addon.ebs_csi_driver.addon_name
      addon_version = aws_eks_addon.ebs_csi_driver.addon_version
      status        = aws_eks_addon.ebs_csi_driver.status
    }
  }
}