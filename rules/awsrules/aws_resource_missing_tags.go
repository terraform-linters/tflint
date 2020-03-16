package awsrules

import (
	"fmt"
	"log"
	"sort"
	"strings"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/terraform/configs"
	"github.com/terraform-linters/tflint/tflint"
	"github.com/zclconf/go-cty/cty"
)

// AwsResourceMissingTagsRule checks whether the resource is tagged correctly
type AwsResourceMissingTagsRule struct {
	resourceTypes []string
}

type awsResourceTagsRuleConfig struct {
	Tags    []string `hcl:"tags"`
	Exclude []string `hcl:"exclude,optional"`
}

const (
	tagsAttributeName = "tags"
	tagBlockName      = "tag"
)

// NewAwsResourceMissingTagsRule returns new rules for all resources that support tags
func NewAwsResourceMissingTagsRule() *AwsResourceMissingTagsRule {
	resourceTypes := []string{
		"aws_accessanalyzer_analyzer",
		"aws_acm_certificate",
		"aws_acmpca_certificate_authority",
		"aws_alb",
		"aws_alb_target_group",
		"aws_ami",
		"aws_ami_copy",
		"aws_ami_from_instance",
		"aws_api_gateway_api_key",
		"aws_api_gateway_client_certificate",
		"aws_api_gateway_domain_name",
		"aws_api_gateway_rest_api",
		"aws_api_gateway_stage",
		"aws_api_gateway_usage_plan",
		"aws_api_gateway_vpc_link",
		"aws_apigatewayv2_api",
		"aws_appmesh_mesh",
		"aws_appmesh_route",
		"aws_appmesh_virtual_node",
		"aws_appmesh_virtual_router",
		"aws_appmesh_virtual_service",
		"aws_appsync_graphql_api",
		"aws_athena_workgroup",
		"aws_autoscaling_group",
		"aws_backup_plan",
		"aws_backup_vault",
		"aws_cloud9_environment_ec2",
		"aws_cloudformation_stack",
		"aws_cloudformation_stack_set",
		"aws_cloudfront_distribution",
		"aws_cloudhsm_v2_cluster",
		"aws_cloudtrail",
		"aws_cloudwatch_event_rule",
		"aws_cloudwatch_log_group",
		"aws_cloudwatch_metric_alarm",
		"aws_codebuild_project",
		"aws_codecommit_repository",
		"aws_codepipeline",
		"aws_codepipeline_webhook",
		"aws_codestarnotifications_notification_rule",
		"aws_cognito_identity_pool",
		"aws_cognito_user_pool",
		"aws_config_aggregate_authorization",
		"aws_config_config_rule",
		"aws_config_configuration_aggregator",
		"aws_customer_gateway",
		"aws_datapipeline_pipeline",
		"aws_datasync_agent",
		"aws_datasync_location_efs",
		"aws_datasync_location_nfs",
		"aws_datasync_location_s3",
		"aws_datasync_location_smb",
		"aws_datasync_task",
		"aws_dax_cluster",
		"aws_db_cluster_snapshot",
		"aws_db_event_subscription",
		"aws_db_instance",
		"aws_db_option_group",
		"aws_db_parameter_group",
		"aws_db_security_group",
		"aws_db_snapshot",
		"aws_db_subnet_group",
		"aws_default_network_acl",
		"aws_default_route_table",
		"aws_default_security_group",
		"aws_default_subnet",
		"aws_default_vpc",
		"aws_default_vpc_dhcp_options",
		"aws_directory_service_directory",
		"aws_dlm_lifecycle_policy",
		"aws_dms_endpoint",
		"aws_dms_replication_instance",
		"aws_dms_replication_subnet_group",
		"aws_dms_replication_task",
		"aws_docdb_cluster",
		"aws_docdb_cluster_instance",
		"aws_docdb_cluster_parameter_group",
		"aws_docdb_subnet_group",
		"aws_dx_connection",
		"aws_dx_hosted_private_virtual_interface_accepter",
		"aws_dx_hosted_public_virtual_interface_accepter",
		"aws_dx_hosted_transit_virtual_interface_accepter",
		"aws_dx_lag",
		"aws_dx_private_virtual_interface",
		"aws_dx_public_virtual_interface",
		"aws_dx_transit_virtual_interface",
		"aws_dynamodb_table",
		"aws_ebs_snapshot",
		"aws_ebs_snapshot_copy",
		"aws_ebs_volume",
		"aws_ec2_capacity_reservation",
		"aws_ec2_client_vpn_endpoint",
		"aws_ec2_fleet",
		"aws_ec2_traffic_mirror_filter",
		"aws_ec2_traffic_mirror_session",
		"aws_ec2_traffic_mirror_target",
		"aws_ec2_transit_gateway",
		"aws_ec2_transit_gateway_route_table",
		"aws_ec2_transit_gateway_vpc_attachment",
		"aws_ec2_transit_gateway_vpc_attachment_accepter",
		"aws_ecr_repository",
		"aws_ecs_capacity_provider",
		"aws_ecs_cluster",
		"aws_ecs_service",
		"aws_ecs_task_definition",
		"aws_efs_file_system",
		"aws_eip",
		"aws_eks_cluster",
		"aws_eks_fargate_profile",
		"aws_eks_node_group",
		"aws_elastic_beanstalk_application",
		"aws_elastic_beanstalk_application_version",
		"aws_elastic_beanstalk_environment",
		"aws_elasticache_cluster",
		"aws_elasticache_replication_group",
		"aws_elasticsearch_domain",
		"aws_elb",
		"aws_emr_cluster",
		"aws_flow_log",
		"aws_fsx_lustre_file_system",
		"aws_fsx_windows_file_system",
		"aws_gamelift_alias",
		"aws_gamelift_build",
		"aws_gamelift_fleet",
		"aws_gamelift_game_session_queue",
		"aws_glacier_vault",
		"aws_globalaccelerator_accelerator",
		"aws_glue_crawler",
		"aws_glue_job",
		"aws_glue_trigger",
		"aws_iam_role",
		"aws_iam_user",
		"aws_inspector_resource_group",
		"aws_instance",
		"aws_internet_gateway",
		"aws_key_pair",
		"aws_kinesis_analytics_application",
		"aws_kinesis_firehose_delivery_stream",
		"aws_kinesis_stream",
		"aws_kms_external_key",
		"aws_kms_key",
		"aws_lambda_function",
		"aws_launch_template",
		"aws_lb",
		"aws_lb_target_group",
		"aws_licensemanager_license_configuration",
		"aws_lightsail_instance",
		"aws_media_convert_queue",
		"aws_media_package_channel",
		"aws_media_store_container",
		"aws_mq_broker",
		"aws_mq_configuration",
		"aws_msk_cluster",
		"aws_nat_gateway",
		"aws_neptune_cluster",
		"aws_neptune_cluster_instance",
		"aws_neptune_cluster_parameter_group",
		"aws_neptune_event_subscription",
		"aws_neptune_parameter_group",
		"aws_neptune_subnet_group",
		"aws_network_acl",
		"aws_network_interface",
		"aws_opsworks_stack",
		"aws_organizations_account",
		"aws_pinpoint_app",
		"aws_placement_group",
		"aws_qldb_ledger",
		"aws_ram_resource_share",
		"aws_rds_cluster",
		"aws_rds_cluster_endpoint",
		"aws_rds_cluster_instance",
		"aws_rds_cluster_parameter_group",
		"aws_redshift_cluster",
		"aws_redshift_event_subscription",
		"aws_redshift_parameter_group",
		"aws_redshift_snapshot_copy_grant",
		"aws_redshift_snapshot_schedule",
		"aws_redshift_subnet_group",
		"aws_resourcegroups_group",
		"aws_route53_health_check",
		"aws_route53_resolver_endpoint",
		"aws_route53_resolver_rule",
		"aws_route53_zone",
		"aws_route_table",
		"aws_s3_bucket",
		"aws_s3_bucket_object",
		"aws_sagemaker_endpoint",
		"aws_sagemaker_endpoint_configuration",
		"aws_sagemaker_model",
		"aws_sagemaker_notebook_instance",
		"aws_secretsmanager_secret",
		"aws_security_group",
		"aws_servicecatalog_portfolio",
		"aws_sfn_activity",
		"aws_sfn_state_machine",
		"aws_sns_topic",
		"aws_spot_instance_request",
		"aws_sqs_queue",
		"aws_ssm_activation",
		"aws_ssm_document",
		"aws_ssm_maintenance_window",
		"aws_ssm_parameter",
		"aws_ssm_patch_baseline",
		"aws_storagegateway_cached_iscsi_volume",
		"aws_storagegateway_gateway",
		"aws_storagegateway_nfs_file_share",
		"aws_storagegateway_smb_file_share",
		"aws_subnet",
		"aws_swf_domain",
		"aws_transfer_server",
		"aws_transfer_user",
		"aws_vpc",
		"aws_vpc_dhcp_options",
		"aws_vpc_endpoint",
		"aws_vpc_endpoint_service",
		"aws_vpc_peering_connection",
		"aws_vpc_peering_connection_accepter",
		"aws_vpn_connection",
		"aws_vpn_gateway",
		"aws_waf_rate_based_rule",
		"aws_waf_rule",
		"aws_waf_rule_group",
		"aws_waf_web_acl",
		"aws_wafregional_rate_based_rule",
		"aws_wafregional_rule",
		"aws_wafregional_rule_group",
		"aws_wafregional_web_acl",
		"aws_workspaces_directory",
		"aws_workspaces_ip_group",
	}
	return &AwsResourceMissingTagsRule{
		resourceTypes: resourceTypes,
	}
}

// Name returns the rule name
func (r *AwsResourceMissingTagsRule) Name() string {
	return "aws_resource_missing_tags"
}

// Enabled returns whether the rule is enabled by default
func (r *AwsResourceMissingTagsRule) Enabled() bool {
	return false
}

// Severity returns the rule severity
func (r *AwsResourceMissingTagsRule) Severity() string {
	return tflint.NOTICE
}

// Link returns the rule reference link
func (r *AwsResourceMissingTagsRule) Link() string {
	return tflint.ReferenceLink(r.Name())
}

// Check checks resources for missing tags
func (r *AwsResourceMissingTagsRule) Check(runner *tflint.Runner) error {
	config := awsResourceTagsRuleConfig{}
	if err := runner.DecodeRuleConfig(r.Name(), &config); err != nil {
		return err
	}

	for _, resourceType := range r.resourceTypes {
		// Skip this resource if its type is excluded in configuration
		if stringInSlice(resourceType, config.Exclude) {
			continue
		}

		// Special handling for tags on aws_autoscaling_group resources
		if resourceType == "aws_autoscaling_group" {
			r.checkAwsAutoScalingGroups(runner, config)
			continue
		}

		for _, resource := range runner.LookupResourcesByType(resourceType) {
			body, _, diags := resource.Config.PartialContent(&hcl.BodySchema{
				Attributes: []hcl.AttributeSchema{
					{
						Name: tagsAttributeName,
					},
				},
			})
			if diags.HasErrors() {
				return diags
			}

			if attribute, ok := body.Attributes[tagsAttributeName]; ok {
				log.Printf("[DEBUG] Walk `%s` attribute", resource.Type+"."+resource.Name+"."+tagsAttributeName)
				err := runner.WithExpressionContext(attribute.Expr, func() error {
					var err error
					resourceTags := make(map[string]string)
					err = runner.EvaluateExpr(attribute.Expr, &resourceTags)
					return runner.EnsureNoError(err, func() error {
						r.emitIssue(runner, resourceTags, config, attribute.Expr.Range())
						return nil
					})
				})
				if err != nil {
					return err
				}
			} else {
				log.Printf("[DEBUG] Walk `%s` resource", resource.Type+"."+resource.Name)
				r.emitIssue(runner, map[string]string{}, config, resource.DeclRange)
			}
		}
	}
	return nil
}

type AwsAutoscalingGroupTag struct {
	Key               string `cty:"key"`
	Value             string `cty:"value"`
	PropagateAtLaunch bool   `cty:"propagate_at_launch"`
}

// checkAwsAutoScalingGroups handles the special case for tags on AutoScaling Groups
// See: https://github.com/terraform-providers/terraform-provider-aws/blob/master/aws/autoscaling_tags.go
func (r *AwsResourceMissingTagsRule) checkAwsAutoScalingGroups(runner *tflint.Runner, config awsResourceTagsRuleConfig) error {
	resourceType := "aws_autoscaling_group"

	for _, resource := range runner.LookupResourcesByType(resourceType) {
		asgTagBlockTags, tagBlockLocation, err := r.checkAwsAutoScalingGroupsTag(runner, config, resource)
		if err != nil {
			return err
		}

		asgTagsAttributeTags, tagsAttributeLocation, err := r.checkAwsAutoScalingGroupsTags(runner, config, resource)
		if err != nil {
			return err
		}

		var location hcl.Range
		tags := make(map[string]string)
		switch {
		case len(asgTagBlockTags) > 0 && len(asgTagsAttributeTags) > 0:
			issue := fmt.Sprintf("Only tag block or tags attribute may be present, but found both")
			runner.EmitIssue(r, issue, resource.DeclRange)
			return nil
		case len(asgTagBlockTags) == 0 && len(asgTagsAttributeTags) == 0:
			r.emitIssue(runner, map[string]string{}, config, resource.DeclRange)
			return nil
		case len(asgTagBlockTags) > 0 && len(asgTagsAttributeTags) == 0:
			tags = asgTagBlockTags
			location = tagBlockLocation
		case len(asgTagBlockTags) == 0 && len(asgTagsAttributeTags) > 0:
			tags = asgTagsAttributeTags
			location = tagsAttributeLocation
		}

		return runner.EnsureNoError(err, func() error {
			r.emitIssue(runner, tags, config, location)
			return nil
		})
	}
	return nil
}

// checkAwsAutoScalingGroupsTag checks tag{} blocks on aws_autoscaling_group resources
func (r *AwsResourceMissingTagsRule) checkAwsAutoScalingGroupsTag(runner *tflint.Runner, config awsResourceTagsRuleConfig, resource *configs.Resource) (map[string]string, hcl.Range, error) {
	tags := make(map[string]string)
	body, _, diags := resource.Config.PartialContent(&hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{
				Type: tagBlockName,
			},
		},
	})
	if diags.HasErrors() {
		return tags, (hcl.Range{}), diags
	}

	for _, tagBlock := range body.Blocks {
		attributes, diags := tagBlock.Body.JustAttributes()
		if diags.HasErrors() {
			return tags, tagBlock.DefRange, diags
		}

		var key string
		err := runner.EvaluateExpr(attributes["key"].Expr, &key)
		if err != nil {
			return tags, tagBlock.DefRange, diags
		}

		var val string
		err = runner.EvaluateExpr(attributes["value"].Expr, &val)
		if err != nil {
			return tags, tagBlock.DefRange, diags
		}
		tags[key] = val
	}

	return tags, resource.DeclRange, nil
}

// checkAwsAutoScalingGroupsTag checks the tags attribute on aws_autoscaling_group resources
func (r *AwsResourceMissingTagsRule) checkAwsAutoScalingGroupsTags(runner *tflint.Runner, config awsResourceTagsRuleConfig, resource *configs.Resource) (map[string]string, hcl.Range, error) {
	tags := make(map[string]string)
	body, _, diags := resource.Config.PartialContent(&hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{
			{
				Name: tagsAttributeName,
			},
		},
	})
	if diags.HasErrors() {
		return tags, (hcl.Range{}), diags
	}

	attribute, ok := body.Attributes[tagsAttributeName]
	if ok {
		err := runner.WithExpressionContext(attribute.Expr, func() error {
			wantType := cty.List(cty.Object(map[string]cty.Type{
				"key":                 cty.String,
				"value":               cty.String,
				"propagate_at_launch": cty.Bool,
			}))
			var asgTags []AwsAutoscalingGroupTag
			err := runner.EvaluateExprType(attribute.Expr, &asgTags, wantType)
			if err != nil {
				return err
			}
			for _, tag := range asgTags {
				tags[tag.Key] = tag.Value
			}
			return nil
		})
		if err != nil {
			return tags, attribute.Expr.Range(), err
		}
		return tags, attribute.Expr.Range(), nil
	}

	return tags, resource.DeclRange, nil
}

func (r *AwsResourceMissingTagsRule) emitIssue(runner *tflint.Runner, tags map[string]string, config awsResourceTagsRuleConfig, location hcl.Range) {
	var missing []string
	for _, tag := range config.Tags {
		if _, ok := tags[tag]; !ok {
			missing = append(missing, fmt.Sprintf("\"%s\"", tag))
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		wanted := strings.Join(missing, ", ")
		issue := fmt.Sprintf("The resource is missing the following tags: %s.", wanted)
		runner.EmitIssue(r, issue, location)
	}
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
