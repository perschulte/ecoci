# DNS Module Variables

variable "project_name" {
  description = "Name of the project"
  type        = string
}

variable "environment" {
  description = "Environment name"
  type        = string
}

variable "domain_name" {
  description = "Base domain name for the project"
  type        = string
}

variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "us-west-2"
}

variable "vpc_id" {
  description = "VPC ID for Route53 resolver query logging"
  type        = string
  default     = null
}

variable "enable_health_check" {
  description = "Enable Route53 health check"
  type        = bool
  default     = true
}

variable "enable_query_logging" {
  description = "Enable Route53 resolver query logging"
  type        = bool
  default     = false
}

variable "alarm_actions" {
  description = "List of ARNs to notify when health check alarm triggers"
  type        = list(string)
  default     = []
}

variable "tags" {
  description = "Additional tags to apply to resources"
  type        = map(string)
  default     = {}
}