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

	detectors = map[string]string{
		"test_rule": "CreateTestDetector",
	}

	for _, tc := range cases {
		prev, _ := filepath.Abs(".")
		dir, _ := os.Getwd()
		defer os.Chdir(prev)
		testDir := dir + "/test-fixtures"
		os.Chdir(testDir)

		listMap := make(map[string]*ast.ObjectList)
		root, _ := parser.Parse([]byte(`
resource "aws_instance" {}

module "ec2_instance" {
    source = "./tf_aws_ec2_instance"
    ami = "ami-12345"
    num = "1"
}`))
		list, _ := root.Node.(*ast.ObjectList)
		listMap["text.tf"] = list

		c := config.Init()
		c.SetIgnoreRule(tc.Config.IgnoreRule)
		c.SetIgnoreModule(tc.Config.IgnoreModule)
		evalConfig, _ := evaluator.NewEvaluator(listMap, map[string]*ast.File{}, c)
		d := &Detector{
			ListMap:    listMap,
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

func TestHclObjectKeyText(t *testing.T) {
	cases := []struct {
		Name   string
		Src    string
		Result string
	}{
		{
			Name: "return key name",
			Src: `
resource "aws_instance" "web" {
    instance_type = "t2.micro"
}`,
			Result: "web",
		},
	}

	for _, tc := range cases {
		root, _ := parser.Parse([]byte(tc.Src))
		list, _ := root.Node.(*ast.ObjectList)
		item := list.Filter("resource", "aws_instance").Items[0]

		result := hclObjectKeyText(item)
		if result != tc.Result {
			t.Fatalf("\nBad: %s\nExpected: %s\n\ntestcase: %s", result, tc.Result, tc.Name)
		}
	}
}

func TestHclLiteralToken(t *testing.T) {
	type Input struct {
		File string
		Key  string
	}

	cases := []struct {
		Name   string
		Input  Input
		Result token.Token
		Error  bool
	}{
		{
			Name: "return literal token",
			Input: Input{
				File: `
resource "aws_instance" "web" {
    instance_type = "t2.micro"
}`,
				Key: "instance_type",
			},
			Result: token.Token{
				Type: 9,
				Pos: token.Pos{
					Filename: "",
					Offset:   47,
					Line:     3,
					Column:   21,
				},
				Text: "\"t2.micro\"",
				JSON: false,
			},
			Error: false,
		},
		{
			Name: "happen error when value is list",
			Input: Input{
				File: `
resource "aws_instance" "web" {
    instance_type = ["t2.micro"]
}`,
				Key: "instance_type",
			},
			Result: token.Token{},
			Error:  true,
		},
		{
			Name: "happen error when value is map",
			Input: Input{
				File: `
resource "aws_instance" "web" {
    instance_type = {
        default = "t2.micro"
    }
}`,
				Key: "instance_type",
			},
			Result: token.Token{},
			Error:  true,
		},
		{
			Name: "happen error when key not found",
			Input: Input{
				File: `
resource "aws_instance" "web" {
    instance_type = "t2.micro"
}`,
				Key: "ami_id",
			},
			Result: token.Token{},
			Error:  true,
		},
	}

	for _, tc := range cases {
		root, _ := parser.Parse([]byte(tc.Input.File))
		list, _ := root.Node.(*ast.ObjectList)
		item := list.Filter("resource", "aws_instance").Items[0]

		result, err := hclLiteralToken(item, tc.Input.Key)
		if tc.Error && err == nil {
			t.Fatalf("\nshould be happen error.\n\ntestcase: %s", tc.Name)
			continue
		}
		if !tc.Error && err != nil {
			t.Fatalf("\nshould not be happen error.\nError: %s\n\ntestcase: %s", err, tc.Name)
			continue
		}

		if result.Text != tc.Result.Text {
			t.Fatalf("\nBad: %s\nExpected: %s\n\ntestcase: %s", pp.Sprint(result), pp.Sprint(tc.Result), tc.Name)
		}
	}
}

func TestHclLiteralListToken(t *testing.T) {
	type Input struct {
		File string
		Key  string
	}

	cases := []struct {
		Name   string
		Input  Input
		Result []token.Token
		Error  bool
	}{
		{
			Name: "return literal tokens",
			Input: Input{
				File: `
resource "aws_instance" "web" {
    vpc_security_group_ids = [
        "sg-1234abcd",
        "sg-abcd1234",
    ]
}`,
				Key: "vpc_security_group_ids",
			},
			Result: []token.Token{
				token.Token{
					Type: 9,
					Pos: token.Pos{
						Filename: "",
						Offset:   72,
						Line:     4,
						Column:   9,
					},
					Text: "\"sg-1234abcd\"",
					JSON: false,
				},
				token.Token{
					Type: 9,
					Pos: token.Pos{
						Filename: "",
						Offset:   95,
						Line:     5,
						Column:   9,
					},
					Text: "\"sg-abcd1234\"",
					JSON: false,
				},
			},
			Error: false,
		},
		{
			Name: "happen error when value is literal",
			Input: Input{
				File: `
resource "aws_instance" "web" {
    vpc_security_group_ids = "sg-1234abcd"
}`,
				Key: "vpc_security_group_ids",
			},
			Result: []token.Token{},
			Error:  true,
		},
		{
			Name: "happen error when value is map",
			Input: Input{
				File: `
resource "aws_instance" "web" {
    vpc_security_group_ids = {
        first  = "sg-1234abcd"
        second = "sg-abcd1234"
    }
}`,
				Key: "vpc_security_group_ids",
			},
			Result: []token.Token{},
			Error:  true,
		},
		{
			Name: "happen error when key not found",
			Input: Input{
				File: `
resource "aws_instance" "web" {
    vpc_security_group_ids = [
        "sg-1234abcd",
        "sg-abcd1234",
    ]
}`,
				Key: "instance_type",
			},
			Result: []token.Token{},
			Error:  true,
		},
	}

	for _, tc := range cases {
		root, _ := parser.Parse([]byte(tc.Input.File))
		list, _ := root.Node.(*ast.ObjectList)
		item := list.Filter("resource", "aws_instance").Items[0]

		result, err := hclLiteralListToken(item, tc.Input.Key)
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

func TestHclObjectItems(t *testing.T) {
	type Input struct {
		File string
		Key  string
	}

	cases := []struct {
		Name   string
		Input  Input
		Result []*ast.ObjectItem
		Error  bool
	}{
		{
			Name: "return object items",
			Input: Input{
				File: `
resource "aws_instance" "web" {
    root_block_device = {
        volume_size = "16"
    }
}`,
				Key: "root_block_device",
			},
			Result: []*ast.ObjectItem{
				&ast.ObjectItem{
					Keys: []*ast.ObjectKey{},
					Assign: token.Pos{
						Filename: "",
						Offset:   55,
						Line:     3,
						Column:   23,
					},
					Val: &ast.ObjectType{
						Lbrace: token.Pos{
							Filename: "",
							Offset:   57,
							Line:     3,
							Column:   25,
						},
						Rbrace: token.Pos{
							Filename: "",
							Offset:   90,
							Line:     5,
							Column:   5,
						},
						List: &ast.ObjectList{
							Items: []*ast.ObjectItem{
								&ast.ObjectItem{
									Keys: []*ast.ObjectKey{
										&ast.ObjectKey{
											Token: token.Token{
												Type: 4,
												Pos: token.Pos{
													Filename: "",
													Offset:   67,
													Line:     4,
													Column:   9,
												},
												Text: "volume_size",
												JSON: false,
											},
										},
									},
									Assign: token.Pos{
										Filename: "",
										Offset:   79,
										Line:     4,
										Column:   21,
									},
									Val: &ast.LiteralType{
										Token: token.Token{
											Type: 9,
											Pos: token.Pos{
												Filename: "",
												Offset:   81,
												Line:     4,
												Column:   23,
											},
											Text: "\"16\"",
											JSON: false,
										},
										LineComment: (*ast.CommentGroup)(nil),
									},
									LeadComment: (*ast.CommentGroup)(nil),
									LineComment: (*ast.CommentGroup)(nil),
								},
							},
						},
					},
					LeadComment: (*ast.CommentGroup)(nil),
					LineComment: (*ast.CommentGroup)(nil),
				},
			},
			Error: false,
		},
		{
			Name: "happen error when key not found",
			Input: Input{
				File: `
resource "aws_instance" "web" {
    root_block_device = {
        volume_size = "16"
    }
}`,
				Key: "ami_id",
			},
			Result: []*ast.ObjectItem{},
			Error:  true,
		},
	}

	for _, tc := range cases {
		root, _ := parser.Parse([]byte(tc.Input.File))
		list, _ := root.Node.(*ast.ObjectList)
		item := list.Filter("resource", "aws_instance").Items[0]

		result, err := hclObjectItems(item, tc.Input.Key)
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

func TestIsKeyNotFound(t *testing.T) {
	type Input struct {
		File string
		Key  string
	}

	cases := []struct {
		Name   string
		Input  Input
		Result bool
	}{
		{
			Name: "key found",
			Input: Input{
				File: `
resource "aws_instance" "web" {
    instance_type = "t2.micro"
}`,
				Key: "instance_type",
			},
			Result: false,
		},
		{
			Name: "happen error when value is list",
			Input: Input{
				File: `
resource "aws_instance" "web" {
    instance_type = "t2.micro"
}`,
				Key: "iam_instance_profile",
			},
			Result: true,
		},
	}

	for _, tc := range cases {
		root, _ := parser.Parse([]byte(tc.Input.File))
		list, _ := root.Node.(*ast.ObjectList)
		item := list.Filter("resource", "aws_instance").Items[0]
		result := IsKeyNotFound(item, tc.Input.Key)

		if result != tc.Result {
			t.Fatalf("\nBad: %t\nExpected: %t\n\ntestcase: %s", result, tc.Result, tc.Name)
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
		listMap := make(map[string]*ast.ObjectList)
		root, _ := parser.Parse([]byte(tc.Input.File))
		list, _ := root.Node.(*ast.ObjectList)
		listMap["text.tf"] = list

		evalConfig, _ := evaluator.NewEvaluator(listMap, map[string]*ast.File{}, config.Init())
		d := &Detector{
			ListMap:    listMap,
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
				token.Token{
					Text: "result1",
					Pos: token.Pos{
						Line: 14,
					},
				},
				token.Token{
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
		listMap := make(map[string]*ast.ObjectList)
		root, _ := parser.Parse([]byte(tc.Input.File))
		list, _ := root.Node.(*ast.ObjectList)
		listMap["text.tf"] = list

		evalConfig, _ := evaluator.NewEvaluator(listMap, map[string]*ast.File{}, config.Init())
		d := &Detector{
			ListMap:    listMap,
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
		File              string
		DeepCheckMode     bool
		DeepCheckDetector bool
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
				File: `
resource "aws_instance" {
    ami = "ami-12345"
}`,
				DeepCheckMode:     true,
				DeepCheckDetector: true,
				Target:            "aws_instance",
			},
			Result: false,
		},
		{
			Name: "return true when disabled deep checking",
			Input: Input{
				File: `
resource "aws_instance" {
    ami = "ami-12345"
}`,
				DeepCheckMode:     false,
				DeepCheckDetector: true,
				Target:            "aws_instance",
			},
			Result: true,
		},
		{
			Name: "return false when disabled deep checking but not deep check detector",
			Input: Input{
				File: `
resource "aws_instance" {
    ami = "ami-12345"
}`,
				DeepCheckMode:     false,
				DeepCheckDetector: false,
				Target:            "aws_instance",
			},
			Result: false,
		},
		{
			Name: "return false when enabled deep checking and not deep check detector",
			Input: Input{
				File: `
resource "aws_instance" {
    ami = "ami-12345"
}`,
				DeepCheckMode:     true,
				DeepCheckDetector: false,
				Target:            "aws_instance",
			},
			Result: false,
		},
		{
			Name: "return true when target resources are not found",
			Input: Input{
				File: `
resource "aws_instance" {
    ami = "ami-12345"
}`,
				DeepCheckMode:     true,
				DeepCheckDetector: true,
				Target:            "aws_db_instance",
			},
			Result: true,
		},
	}

	for _, tc := range cases {
		listMap := make(map[string]*ast.ObjectList)
		root, _ := parser.Parse([]byte(tc.Input.File))
		list, _ := root.Node.(*ast.ObjectList)
		listMap["text.tf"] = list

		d := &Detector{
			ListMap: listMap,
			Config:  config.Init(),
			Logger:  logger.Init(false),
		}
		d.Config.DeepCheck = tc.Input.DeepCheckMode

		result := d.isSkip(tc.Input.DeepCheckDetector, tc.Input.Target)
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
