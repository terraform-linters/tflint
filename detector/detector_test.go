package detector

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/hashicorp/hcl/hcl/parser"
	"github.com/hashicorp/hcl/hcl/token"
	"github.com/k0kubun/pp"
	"github.com/wata727/tflint/config"
	"github.com/wata727/tflint/evaluator"
	"github.com/wata727/tflint/logger"
	"github.com/wata727/tflint/schema"
)

func TestDetect(t *testing.T) {
	type Config struct {
		IgnoreRule   string
		IgnoreModule string
	}

	cases := []struct {
		Name   string
		Config Config
		Result int
	}{
		{
			Name: "detect template and module",
			Config: Config{
				IgnoreRule:   "",
				IgnoreModule: "",
			},
			Result: 2,
		},
		{
			Name: "ignore module",
			Config: Config{
				IgnoreRule:   "",
				IgnoreModule: "./tf_aws_ec2_instance",
			},
			Result: 1,
		},
		{
			Name: "ignore rule",
			Config: Config{
				IgnoreRule:   "test_rule",
				IgnoreModule: "",
			},
			Result: 0,
		},
	}

	detectorFactories = []string{"CreateTestDetector"}

	for _, tc := range cases {
		prev, _ := filepath.Abs(".")
		dir, _ := os.Getwd()
		defer os.Chdir(prev)
		testDir := dir + "/test-fixtures"
		os.Chdir(testDir)

		src := `
resource "aws_instance" "web" {
    instance_type = "t2.micro"
}

module "ec2_instance" {
    source = "./tf_aws_ec2_instance"
    ami = "ami-12345"
    num = "1"
}`
		templates := make(map[string]*ast.File)
		templates["text.tf"], _ = parser.Parse([]byte(src))
		files := map[string][]byte{"text.tf": []byte(src)}
		schema, _ := schema.Make(files)

		c := config.Init()
		c.SetIgnoreRule(tc.Config.IgnoreRule)
		c.SetIgnoreModule(tc.Config.IgnoreModule)
		evalConfig, _ := evaluator.NewEvaluator(templates, schema, []*ast.File{}, c)
		d := &Detector{
			Schema:     schema,
			Config:     c,
			EvalConfig: evalConfig,
			Logger:     logger.Init(false),
		}

		issues := d.Detect()
		if len(issues) != tc.Result {
			t.Fatalf("\nBad: %d\nExpected: %d\n\ntestcase: %s", len(issues), tc.Result, tc.Name)
		}
	}
}

func TestEvalToString(t *testing.T) {
	type Input struct {
		Src  string
		File string
	}

	cases := []struct {
		Name   string
		Input  Input
		Result string
		Error  bool
	}{
		{
			Name: "return string",
			Input: Input{
				Src: "${var.text}",
				File: `
variable "text" {
    default = "result"
}`,
			},
			Result: "result",
			Error:  false,
		},
		{
			Name: "not string",
			Input: Input{
				Src: "${var.text}",
				File: `
variable "text" {
    default = ["result"]
}`,
			},
			Result: "",
			Error:  true,
		},
		{
			Name: "not evaluable",
			Input: Input{
				Src:  "${aws_instance.app}",
				File: `variable "text" {}`,
			},
			Result: "",
			Error:  true,
		},
	}

	for _, tc := range cases {
		templates := make(map[string]*ast.File)
		templates["text.tf"], _ = parser.Parse([]byte(tc.Input.File))

		evalConfig, _ := evaluator.NewEvaluator(templates, []*schema.Template{}, []*ast.File{}, config.Init())
		d := &Detector{
			EvalConfig: evalConfig,
		}

		result, err := d.evalToString(tc.Input.Src)
		if tc.Error && err == nil {
			t.Fatalf("\nshould be happen error.\n\ntestcase: %s", tc.Name)
			continue
		}
		if !tc.Error && err != nil {
			t.Fatalf("\nshould not be happen error.\nError: %s\n\ntestcase: %s", err, tc.Name)
			continue
		}

		if result != tc.Result {
			t.Fatalf("\nBad: %s\nExpected: %s\n\ntestcase: %s", result, tc.Result, tc.Name)
		}
	}
}

func TestEvalToStringTokens(t *testing.T) {
	type Input struct {
		Src  token.Token
		File string
	}

	cases := []struct {
		Name   string
		Input  Input
		Result []token.Token
		Error  bool
	}{
		{
			Name: "return list",
			Input: Input{
				Src: token.Token{
					Text: "${var.array}",
					Pos: token.Pos{
						Line: 14,
					},
				},
				File: `
variable "array" {
    default = ["result1", "result2"]
}`,
			},
			Result: []token.Token{
				{
					Text: "result1",
					Pos: token.Pos{
						Line: 14,
					},
				},
				{
					Text: "result2",
					Pos: token.Pos{
						Line: 14,
					},
				},
			},
			Error: false,
		},
		{
			Name: "not list",
			Input: Input{
				Src: token.Token{
					Text: "${var.array}",
					Pos: token.Pos{
						Line: 14,
					},
				},
				File: `
variable "array" {
    default = "result"
}`,
			},
			Result: []token.Token{},
			Error:  true,
		},
		{
			Name: "not evaluable",
			Input: Input{
				Src: token.Token{
					Text: "${var.array}",
					Pos: token.Pos{
						Line: 14,
					},
				},
				File: `variable "array" {}`,
			},
			Result: []token.Token{},
			Error:  true,
		},
	}

	for _, tc := range cases {
		templates := make(map[string]*ast.File)
		templates["text.tf"], _ = parser.Parse([]byte(tc.Input.File))

		evalConfig, _ := evaluator.NewEvaluator(templates, []*schema.Template{}, []*ast.File{}, config.Init())
		d := &Detector{
			EvalConfig: evalConfig,
		}

		result, err := d.evalToStringTokens(tc.Input.Src)
		if tc.Error && err == nil {
			t.Fatalf("\nshould be happen error.\n\ntestcase: %s", tc.Name)
			continue
		}
		if !tc.Error && err != nil {
			t.Fatalf("\nshould not be happen error.\nError: %s\n\ntestcase: %s", err, tc.Name)
			continue
		}

		if !reflect.DeepEqual(result, tc.Result) {
			t.Fatalf("\nBad: %s\nExpected: %s\n\ntestcase: %s", pp.Sprint(result), pp.Sprint(tc.Result), tc.Name)
		}
	}
}

func TestIsSkip(t *testing.T) {
	type Input struct {
		RuleName          string
		File              string
		DeepCheckMode     bool
		DeepCheckDetector bool
		TargetType        string
		Target            string
	}

	cases := []struct {
		Name   string
		Input  Input
		Result bool
	}{
		{
			Name: "return false when enabled deep checking",
			Input: Input{
				RuleName: "aws_instance_invalid_type",
				File: `
resource "aws_instance" "web" {
    ami = "ami-12345"
}`,
				DeepCheckMode:     true,
				DeepCheckDetector: true,
				TargetType:        "resource",
				Target:            "aws_instance",
			},
			Result: false,
		},
		{
			Name: "return true when disabled deep checking",
			Input: Input{
				RuleName: "aws_instance_invalid_type",
				File: `
resource "aws_instance" "web" {
    ami = "ami-12345"
}`,
				DeepCheckMode:     false,
				DeepCheckDetector: true,
				TargetType:        "resource",
				Target:            "aws_instance",
			},
			Result: true,
		},
		{
			Name: "return false when disabled deep checking but not deep check detector",
			Input: Input{
				RuleName: "aws_instance_invalid_type",
				File: `
resource "aws_instance" "web" {
    ami = "ami-12345"
}`,
				DeepCheckMode:     false,
				DeepCheckDetector: false,
				TargetType:        "resource",
				Target:            "aws_instance",
			},
			Result: false,
		},
		{
			Name: "return false when enabled deep checking and not deep check detector",
			Input: Input{
				RuleName: "aws_instance_invalid_type",
				File: `
resource "aws_instance" "web" {
    ami = "ami-12345"
}`,
				DeepCheckMode:     true,
				DeepCheckDetector: false,
				TargetType:        "resource",
				Target:            "aws_instance",
			},
			Result: false,
		},
		{
			Name: "return true when target resources are not found",
			Input: Input{
				RuleName: "aws_instance_invalid_type",
				File: `
resource "aws_instance" "web" {
    ami = "ami-12345"
}`,
				DeepCheckMode:     true,
				DeepCheckDetector: true,
				TargetType:        "resource",
				Target:            "aws_db_instance",
			},
			Result: true,
		},
		{
			Name: "return false when modules are found",
			Input: Input{
				RuleName: "aws_instance_invalid_type",
				File: `
module "ec2_instance" {
    source = "./ec2_instance"
}`,
				DeepCheckMode:     true,
				DeepCheckDetector: true,
				TargetType:        "module",
			},
			Result: false,
		},
		{
			Name: "return true when target modules are not found",
			Input: Input{
				RuleName: "aws_instance_invalid_type",
				File: `
resource "aws_instance" "web" {
    ami = "ami-12345"
}`,
				DeepCheckMode:     true,
				DeepCheckDetector: true,
				TargetType:        "module",
			},
			Result: true,
		},
	}

	for _, tc := range cases {
		templates := make(map[string]*ast.File)
		templates["text.tf"], _ = parser.Parse([]byte(tc.Input.File))
		files := map[string][]byte{"test.tf": []byte(tc.Input.File)}
		schema, _ := schema.Make(files)

		d := &Detector{
			Schema: schema,
			Config: config.Init(),
			Logger: logger.Init(false),
		}
		d.Config.DeepCheck = tc.Input.DeepCheckMode

		result := d.isSkip(
			tc.Input.RuleName,
			tc.Input.DeepCheckDetector,
			tc.Input.TargetType,
			tc.Input.Target,
		)
		if result != tc.Result {
			t.Fatalf("\nBad: %t\nExpected: %t\n\ntestcase: %s", result, tc.Result, tc.Name)
		}
	}
}

func TestHasError(t *testing.T) {
	d := &Detector{Error: false}

	if d.HasError() {
		t.Fatal("If no error has occurred, should return false.")
	}
	d.Error = true
	if !d.HasError() {
		t.Fatal("If an error has occurred, should return true.")
	}
}
