resource "aws_iam_instance_profile" "ecs" {
  name  = "${var.app_name}-ecs-instance"
  roles = ["${aws_iam_role.ecs_instance.name}"]
}

resource "aws_iam_policy_attachment" "ecs_instance" {
  name       = "${var.app_name}-ecs-instance"
  roles      = ["${aws_iam_role.ecs_instance.name}"]
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonEC2ContainerServiceforEC2Role"
}

resource "aws_iam_role" "ecs_instance" {
  name = "${var.app_name}-ecs-instance"
  path = "/"

  assume_role_policy = <<EOF
{
    "Version": "2008-10-17",
    "Statement": [
      {
        "Action": "sts:AssumeRole",
        "Principal": {
          "Service": "ec2.amazonaws.com"
        },
        "Effect": "Allow",
        "Sid": ""
      }
    ]
}
EOF
}

resource "aws_security_group" "ecs_instance" {
  name        = "${var.app_name}-ecs-instance"
  description = "container security group for ${var.app_name}"
  vpc_id      = "${var.vpc}"

  ingress {
    from_port       = 0
    to_port         = 65535
    protocol        = "TCP"
    security_groups = ["${aws_security_group.ecs_alb.id}"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_iam_policy_attachment" "fleet" {
  name       = "${var.app_name}-fleet"
  roles      = ["${aws_iam_role.fleet.name}"]
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonEC2SpotFleetRole"
}

resource "aws_iam_role" "fleet" {
  name = "${var.app_name}-fleet"

  assume_role_policy = <<EOF
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "spotfleet.amazonaws.com",
          "ec2.amazonaws.com"
        ]
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_spot_fleet_request" "main" {
  iam_fleet_role                      = "${aws_iam_role.fleet.arn}"
  spot_price                          = "${var.spot_prices[0]}"
  allocation_strategy                 = "${var.strategy}"
  target_capacity                     = "${var.instance_count}"
  terminate_instances_with_expiration = true
  valid_until                         = "${var.valid_until}"

  launch_specification {
    ami                    = "${var.ami}"
    instance_type          = "${var.instance_type}"
    spot_price             = "${var.spot_prices[0]}"
    subnet_id              = "${var.subnets[0]}"
    vpc_security_group_ids = ["${aws_security_group.ecs_instance.id}"]
    iam_instance_profile   = "${aws_iam_instance_profile.ecs.name}"
    key_name               = "${var.key_name}"

    root_block_device = {
      volume_type = "gp2"
      volume_size = "${var.volume_size}"
    }

    user_data = <<USER_DATA
#!/bin/bash
echo ECS_CLUSTER=${aws_ecs_cluster.main.name} >> /etc/ecs/ecs.config
USER_DATA
  }

  launch_specification {
    ami                    = "${var.ami}"
    instance_type          = "${var.instance_type}"
    spot_price             = "${var.spot_prices[1]}"
    subnet_id              = "${var.subnets[1]}"
    vpc_security_group_ids = ["${aws_security_group.ecs_instance.id}"]
    iam_instance_profile   = "${aws_iam_instance_profile.ecs.name}"
    key_name               = "${var.key_name}"

    root_block_device = {
      volume_type = "gp2"
      volume_size = "${var.volume_size}"
    }

    user_data = <<USER_DATA
#!/bin/bash
echo ECS_CLUSTER=${aws_ecs_cluster.main.name} >> /etc/ecs/ecs.config
USER_DATA
  }

  depends_on = ["aws_iam_policy_attachment.fleet"]
}
