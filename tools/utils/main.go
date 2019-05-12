package utils

import (
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/iancoleman/strcase"
)

// ToCamel converts a string to CamelCase
func ToCamel(str string) string {
	exceptions := map[string]string{
		"ami":               "AMI",
		"api":               "API",
		"db":                "DB",
		"alb":               "ALB",
		"elb":               "ELB",
		"elasticache":       "ElastiCache",
		"iam":               "IAM",
		"account_id":        "AccountID",
		"subnet_id":         "SubnetID",
		"cluster_id":        "ClusterID",
		"url":               "URL",
		"uri":               "URI",
		"http":              "HTTP",
		"ip_address":        "IPAddress",
		"statement_id":      "StatementID",
		"target_id":         "TargetID",
		"key_id":            "KeyID",
		"pool_id":           "PoolID",
		"directory_id":      "DirectoryID",
		"dns":               "DNS",
		"build_id":          "BuildID",
		"detector_id":       "DetectorID",
		"ssh":               "SSH",
		"blueprint_id":      "BlueprintID",
		"bundle_id":         "BundleID",
		"static_ip":         "StaticIP",
		"parent_id":         "ParentID",
		"policy_id":         "PolicyID",
		"zone_id":           "ZoneID",
		"health_check_id":   "HealthCheckID",
		"delegation_set_id": "DelegationSetID",
		"rule_id":           "RuleID",
		"vpc_id":            "VpcID",
		"endpoint_id":       "EndpointID",
		"acl":               "ACL",
		"secret_id":         "SecretID",
		"tls":               "TLS",
		"instance_id":       "InstanceID",
		"window_id":         "WindowID",
		"baseline_id":       "BaselineID",
		"disk_id":           "DiskID",
		"interface_id":      "InterfaceID",
		"snapshot_id":       "SnapshotID",
		"server_id":         "ServerID",
		"sql":               "SQL",
		"xss":               "XSS",
	}
	for pattern, conv := range exceptions {
		str = strings.Replace(str, "_"+pattern+"_", "_"+conv+"_", -1)
		str = strings.Replace(str, pattern+"_", conv+"_", -1)
		str = strings.Replace(str, "_"+pattern, "_"+conv, -1)
	}
	return strcase.ToCamel(str)
}

// GenerateFile generates a new file from the passed template and metadata
func GenerateFile(fileName string, tmplName string, meta interface{}) {
	file, err := os.Create(fileName)
	if err != nil {
		panic(err)
	}

	tmpl := template.Must(template.ParseFiles(tmplName))
	err = tmpl.Execute(file, meta)
	if err != nil {
		panic(err)
	}
}

// GenerateFileWithLogs generates a new file from the passed template and metadata
// The difference from GenerateFile function is to output logs
func GenerateFileWithLogs(fileName string, tmplName string, meta interface{}) {
	GenerateFile(fileName, tmplName, meta)
	fmt.Printf("Create: %s\n", fileName)
}
