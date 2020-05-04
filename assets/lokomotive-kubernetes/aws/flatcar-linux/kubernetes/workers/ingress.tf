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
  vpc_id      = var.vpc_id
  target_type = "instance"

  protocol = "TCP"
  port     = 30080

  health_check {
    protocol = "TCP"
    port     = 30080

    # NLBs required to use same healthy and unhealthy thresholds
    healthy_threshold   = 3
    unhealthy_threshold = 3

    # Interval between health checks required to be 10 or 30
    interval = 10
  }

  tags = {
    ClusterName = var.cluster_name
    PoolName    = var.pool_name
  }
}

resource "aws_lb_target_group" "workers-https" {
  vpc_id      = var.vpc_id
  target_type = "instance"

  protocol = "TCP"
  port     = 30443

  health_check {
    protocol = "TCP"
    port     = 30443

    # NLBs required to use same healthy and unhealthy thresholds
    healthy_threshold   = 3
    unhealthy_threshold = 3

    # Interval between health checks required to be 10 or 30
    interval = 10
  }

  tags = {
    ClusterName = var.cluster_name
    PoolName    = var.pool_name
  }
}
