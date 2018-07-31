resource "aws_security_group" "ecs_alb" {
  description = "Balancer for ${var.app_name}"

  vpc_id = "${var.vpc}"
  name   = "${var.app_name}-alb-sg"

  ingress {
    protocol    = "tcp"
    from_port   = "${var.app_port}"
    to_port     = "${var.app_port}"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_alb" "main" {
  name            = "${var.app_name}"
  subnets         = "${var.subnets}"
  security_groups = ["${aws_security_group.ecs_alb.id}"]
}

resource "aws_alb_target_group" "main" {
  name     = "${var.app_name}"
  port     = "${var.app_port}"
  protocol = "HTTP"
  vpc_id   = "${var.vpc}"

  depends_on = [
    "aws_alb.main"
  ]
}

resource "aws_alb_listener" "https" {
  count             = "${var.https ? 1 : 0}"
  load_balancer_arn = "${aws_alb.main.id}"
  port              = "443"
  protocol          = "HTTPS"
  ssl_policy        = "${var.app_ssl_policy}"
  certificate_arn   = "${var.app_certificate_arn}"

  default_action {
    target_group_arn = "${aws_alb_target_group.main.id}"
    type             = "forward"
  }
}

resource "aws_alb_listener" "http" {
  count             = "${var.https ? 0 : 1}"
  load_balancer_arn = "${aws_alb.main.id}"
  port              = "80"
  protocol          = "HTTP"

  default_action {
    target_group_arn = "${aws_alb_target_group.main.id}"
    type             = "forward"
  }
}
