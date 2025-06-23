# DNS Module Outputs

output "zone_id" {
  description = "Route53 hosted zone ID"
  value       = aws_route53_zone.main.zone_id
}

output "zone_arn" {
  description = "Route53 hosted zone ARN"
  value       = aws_route53_zone.main.arn
}

output "zone_name" {
  description = "Route53 hosted zone name"
  value       = aws_route53_zone.main.name
}

output "name_servers" {
  description = "Name servers for the hosted zone"
  value       = aws_route53_zone.main.name_servers
}

output "certificate_arn" {
  description = "ARN of the ACM certificate"
  value       = aws_acm_certificate_validation.main.certificate_arn
}

output "certificate_domain_name" {
  description = "Domain name of the ACM certificate"
  value       = aws_acm_certificate.main.domain_name
}

output "certificate_subject_alternative_names" {
  description = "Subject alternative names of the ACM certificate"
  value       = aws_acm_certificate.main.subject_alternative_names
}

output "certificate_status" {
  description = "Status of the ACM certificate"
  value       = aws_acm_certificate.main.status
}

output "health_check_id" {
  description = "Route53 health check ID"
  value       = var.enable_health_check ? aws_route53_health_check.main[0].id : null
}

output "health_check_arn" {
  description = "Route53 health check ARN"
  value       = var.enable_health_check ? aws_route53_health_check.main[0].arn : null
}

output "query_log_config_id" {
  description = "Route53 resolver query log configuration ID"
  value       = var.enable_query_logging ? aws_route53_resolver_query_log_config.main[0].id : null
}

output "query_log_config_arn" {
  description = "Route53 resolver query log configuration ARN"
  value       = var.enable_query_logging ? aws_route53_resolver_query_log_config.main[0].arn : null
}