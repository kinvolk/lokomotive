resource "aws_lb_listener" "ingress-http" {
  load_balancer_arn = var.lb_arn
  protocol          = "TCP"
  port              = var.lb_http_port

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.workers-http.arn
  }
}

resource "aws_lb_listener" "ingress-https" {
  load_balancer_arn = var.lb_arn
  protocol          = "TCP"
  port              = var.lb_https_port

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.workers-https.arn
  }
}

resource "aws_lb_target_group" "workers-http" {
  name        = "${var.cluster_name}-${var.pool_name}-http"
  vpc_id      = var.vpc_id
  target_type = "instance"

  protocol = "TCP"
  port     = 80

  # HTTP health check for ingress
  health_check {
    protocol = "HTTP"
    port     = 10254
    path     = "/healthz"

    # NLBs required to use same healthy and unhealthy thresholds
    healthy_threshold   = 3
    unhealthy_threshold = 3

    # Interval between health checks required to be 10 or 30
    interval = 10
  }
}

resource "aws_lb_target_group" "workers-https" {
  name        = "${var.cluster_name}-${var.pool_name}-https"
  vpc_id      = var.vpc_id
  target_type = "instance"

  protocol = "TCP"
  port     = 443

  # HTTP health check for ingress
  health_check {
    protocol = "HTTP"
    port     = 10254
    path     = "/healthz"

    # NLBs required to use same healthy and unhealthy thresholds
    healthy_threshold   = 3
    unhealthy_threshold = 3

    # Interval between health checks required to be 10 or 30
    interval = 10
  }
}
