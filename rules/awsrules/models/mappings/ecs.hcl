import = "aws-sdk-go/models/apis/ecs/2014-11-13/api-2.json"

mapping "aws_ecs_cluster" {
  name = String
  tags = Tags
}

mapping "aws_ecs_service" {
  name                               = String
  task_definition                    = String
  desired_count                      = BoxedInteger
  launch_type                        = LaunchType
  platform_version                   = String
  scheduling_strategy                = SchedulingStrategy
  cluster                            = String
  iam_role                           = String
  deployment_controller              = DeploymentController
  deployment_maximum_percent         = BoxedInteger
  deployment_minimum_healthy_percent = BoxedInteger
  enable_ecs_managed_tags            = Boolean
  propagate_tags                     = PropagateTags
  ordered_placement_strategy         = PlacementStrategies
  health_check_grace_period_seconds  = BoxedInteger
  load_balancer                      = LoadBalancers
  placement_constraints              = PlacementConstraints
  network_configuration              = NetworkConfiguration
  service_registries                 = ServiceRegistries
  tags                               = Tags
}

mapping "aws_ecs_task_definition" {
  family                   = String
  container_definitions    = ContainerDefinitions
  task_role_arn            = String
  execution_role_arn       = String
  network_mode             = NetworkMode
  ipc_mode                 = IpcMode
  pid_mode                 = PidMode
  volume                   = VolumeList
  placement_constraints    = TaskDefinitionPlacementConstraints
  cpu                      = String
  memory                   = String
  requires_compatibilities = CompatibilityList
  tags                     = Tags
}

test "aws_ecs_service" "launch_type" {
  ok = "FARGATE"
  ng = "POD"
}

test "aws_ecs_service" "propagate_tags" {
  ok = "SERVICE"
  ng = "CONTAINER"
}

test "aws_ecs_service" "scheduling_strategy" {
  ok = "REPLICA"
  ng = "SERVER"
}

test "aws_ecs_task_definition" "ipc_mode" {
  ok = "host"
  ng = "vpc"
}

test "aws_ecs_task_definition" "network_mode" {
  ok = "bridge"
  ng = "vpc"
}

test "aws_ecs_task_definition" "pid_mode" {
  ok = "task"
  ng = "awsvpc"
}
