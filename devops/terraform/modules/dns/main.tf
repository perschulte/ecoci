# DNS Module for EcoCI Infrastructure (Route53 + ACM)

locals {
  # Create the subdomain for the environment
  zone_name = var.environment == "production" ? var.domain_name : "${var.environment}.${var.domain_name}"
}

# Route53 Hosted Zone
resource "aws_route53_zone" "main" {
  name = local.zone_name

  tags = merge(var.tags, {
    Name        = "${var.project_name}-${var.environment}-zone"
    Environment = var.environment
  })
}

# ACM Certificate for the domain and its subdomains
resource "aws_acm_certificate" "main" {
  domain_name               = local.zone_name
  subject_alternative_names = ["*.${local.zone_name}"]
  validation_method         = "DNS"

  lifecycle {
    create_before_destroy = true
  }

  tags = merge(var.tags, {
    Name        = "${var.project_name}-${var.environment}-cert"
    Environment = var.environment
  })
}

# Route53 records for ACM certificate validation
resource "aws_route53_record" "cert_validation" {
  for_each = {
    for dvo in aws_acm_certificate.main.domain_validation_options : dvo.domain_name => {
      name   = dvo.resource_record_name
      record = dvo.resource_record_value
      type   = dvo.resource_record_type
    }
  }

  allow_overwrite = true
  name            = each.value.name
  records         = [each.value.record]
  ttl             = 60
  type            = each.value.type
  zone_id         = aws_route53_zone.main.zone_id
}

# ACM certificate validation
resource "aws_acm_certificate_validation" "main" {
  certificate_arn         = aws_acm_certificate.main.arn
  validation_record_fqdns = [for record in aws_route53_record.cert_validation : record.fqdn]
  
  timeouts {
    create = "5m"
  }
}

# Health check for monitoring
resource "aws_route53_health_check" "main" {
  count                           = var.enable_health_check ? 1 : 0
  fqdn                           = local.zone_name
  port                           = 443
  type                           = "HTTPS_STR_MATCH"
  resource_path                  = "/healthz"
  failure_threshold              = "3"
  request_interval               = "30"
  search_string                  = "ok"
  cloudwatch_logs_region         = var.aws_region
  insufficient_data_health_status = "Failure"

  tags = merge(var.tags, {
    Name = "${var.project_name}-${var.environment}-health-check"
  })
}

# CloudWatch alarm for health check
resource "aws_cloudwatch_metric_alarm" "health_check" {
  count               = var.enable_health_check ? 1 : 0
  alarm_name          = "${var.project_name}-${var.environment}-health-check-failed"
  comparison_operator = "LessThanThreshold"
  evaluation_periods  = "2"
  metric_name         = "HealthCheckStatus"
  namespace           = "AWS/Route53"
  period              = "60"
  statistic           = "Minimum"
  threshold           = "1"
  alarm_description   = "This metric monitors whether the health check is passing"
  alarm_actions       = var.alarm_actions

  dimensions = {
    HealthCheckId = aws_route53_health_check.main[0].id
  }

  tags = var.tags
}

# Route53 Resolver Query Log Configuration (optional for debugging)
resource "aws_route53_resolver_query_log_config" "main" {
  count           = var.enable_query_logging ? 1 : 0
  name            = "${var.project_name}-${var.environment}-query-logs"
  destination_arn = aws_cloudwatch_log_group.dns_query_logs[0].arn

  tags = var.tags
}

resource "aws_cloudwatch_log_group" "dns_query_logs" {
  count             = var.enable_query_logging ? 1 : 0
  name              = "/aws/route53/${var.project_name}-${var.environment}"
  retention_in_days = 7

  tags = var.tags
}

# Associate query log config with VPC (if provided)
resource "aws_route53_resolver_query_log_config_association" "main" {
  count                        = var.enable_query_logging && var.vpc_id != null ? 1 : 0
  resolver_query_log_config_id = aws_route53_resolver_query_log_config.main[0].id
  resource_id                  = var.vpc_id
}