package plugin

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	hcl "github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	sdk "github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/terraform-linters/tflint/tflint"
	"github.com/zclconf/go-cty/cty"
)

func TestGetModuleContent(t *testing.T) {
	runner := tflint.TestRunner(t, map[string]string{"main.tf": `
resource "aws_instance" "foo" {
	instance_type = "t2.micro"
}`})
	rootRunner := tflint.TestRunner(t, map[string]string{"main.tf": `
resource "aws_instance" "bar" {
	instance_type = "m5.2xlarge"
}`})

	server := NewGRPCServer(runner, rootRunner, map[string][]byte{})

	tests := []struct {
		Name string
		Args func() (*hclext.BodySchema, sdk.GetModuleContentOption)
		Want func() (*hclext.BodyContent, hcl.Diagnostics)
	}{
		{
			Name: "self module context",
			Args: func() (*hclext.BodySchema, sdk.GetModuleContentOption) {
				return &hclext.BodySchema{
					Blocks: []hclext.BlockSchema{
						{
							Type:       "resource",
							LabelNames: []string{"type", "name"},
							Body: &hclext.BodySchema{
								Attributes: []hclext.AttributeSchema{{Name: "instance_type"}},
							},
						},
					},
				}, sdk.GetModuleContentOption{ModuleCtx: sdk.SelfModuleCtxType}
			},
			Want: func() (*hclext.BodyContent, hcl.Diagnostics) {
				return runner.GetModuleContent(&hclext.BodySchema{
					Blocks: []hclext.BlockSchema{
						{
							Type:       "resource",
							LabelNames: []string{"type", "name"},
							Body: &hclext.BodySchema{
								Attributes: []hclext.AttributeSchema{{Name: "instance_type"}},
							},
						},
					},
				}, sdk.GetModuleContentOption{})
			},
		},
		{
			Name: "root module context",
			Args: func() (*hclext.BodySchema, sdk.GetModuleContentOption) {
				return &hclext.BodySchema{
					Blocks: []hclext.BlockSchema{
						{
							Type:       "resource",
							LabelNames: []string{"type", "name"},
							Body: &hclext.BodySchema{
								Attributes: []hclext.AttributeSchema{{Name: "instance_type"}},
							},
						},
					},
				}, sdk.GetModuleContentOption{ModuleCtx: sdk.RootModuleCtxType}
			},
			Want: func() (*hclext.BodyContent, hcl.Diagnostics) {
				return rootRunner.GetModuleContent(&hclext.BodySchema{
					Blocks: []hclext.BlockSchema{
						{
							Type:       "resource",
							LabelNames: []string{"type", "name"},
							Body: &hclext.BodySchema{
								Attributes: []hclext.AttributeSchema{{Name: "instance_type"}},
							},
						},
					},
				}, sdk.GetModuleContentOption{})
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			want, diags := test.Want()
			if diags.HasErrors() {
				t.Fatalf("failed to get want: %s", diags)
			}

			got, diags := server.GetModuleContent(test.Args())
			if diags.HasErrors() {
				t.Fatalf("failed to call GetModuleContent: %s", diags)
			}

			opts := cmp.Options{
				cmp.Comparer(func(x, y cty.Value) bool {
					return x.GoString() == y.GoString()
				}),
			}
			if diff := cmp.Diff(got, want, opts); diff != "" {
				t.Errorf("diff: %s", diff)
			}
		})
	}
}

func TestGetFile(t *testing.T) {
	runner := tflint.TestRunner(t, map[string]string{
		"test1.tf": `
resource "aws_instance" "foo" {
	instance_type = "t2.micro"
}`,
		"test2.tf": `
resource "aws_instance" "bar" {
	instance_type = "m5.2xlarge"
}`,
	})

	server := NewGRPCServer(runner, nil, map[string][]byte{})

	tests := []struct {
		Name string
		Arg  string
		Want string
	}{
		{
			Name: "get test1.tf",
			Arg:  "test1.tf",
			Want: `
resource "aws_instance" "foo" {
	instance_type = "t2.micro"
}`,
		},
		{
			Name: "get test2.tf",
			Arg:  "test2.tf",
			Want: `
resource "aws_instance" "bar" {
	instance_type = "m5.2xlarge"
}`,
		},
		{
			Name: "file not found",
			Arg:  "test3.tf",
			Want: "",
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			file, err := server.GetFile(test.Arg)
			if err != nil {
				t.Fatalf("failed to call GetFile: %s", err)
			}

			var got string
			if file != nil {
				got = string(file.Bytes)
			}

			if got != test.Want {
				t.Errorf("unexpected file: %s", got)
			}
		})
	}
}

func TestGetFiles(t *testing.T) {
	runner := tflint.TestRunner(t, map[string]string{"main.tf": `
resource "aws_instance" "foo" {
	instance_type = "t2.micro"
}`})
	rootRunner := tflint.TestRunner(t, map[string]string{"main.tf": `
resource "aws_instance" "bar" {
	instance_type = "m5.2xlarge"
}`})

	server := NewGRPCServer(runner, rootRunner, map[string][]byte{})

	tests := []struct {
		Name string
		Arg  sdk.ModuleCtxType
		Want map[string]string
	}{
		{
			Name: "self module context",
			Arg:  sdk.SelfModuleCtxType,
			Want: map[string]string{"main.tf": `
resource "aws_instance" "foo" {
	instance_type = "t2.micro"
}`},
		},
		{
			Name: "root module context",
			Arg:  sdk.RootModuleCtxType,
			Want: map[string]string{"main.tf": `
resource "aws_instance" "bar" {
	instance_type = "m5.2xlarge"
}`},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			files := server.GetFiles(test.Arg)

			got := map[string]string{}
			for name, file := range files {
				got[name] = string(file)
			}

			if diff := cmp.Diff(got, test.Want); diff != "" {
				t.Errorf("diff: %s", diff)
			}
		})
	}
}

func TestGetRuleConfigContent(t *testing.T) {
	// config from file
	config := `
rule "test_in_file" {
	enabled = true
	foo = "bar"
}`
	configFile := filepath.Join(t.TempDir(), ".tflint.hcl")
	if err := os.WriteFile(configFile, []byte(config), 0755); err != nil {
		t.Fatalf("failed to create config file: %s", err)
	}
	fileConfig, err := tflint.LoadConfig(configFile)
	if err != nil {
		t.Fatalf("failed to load test config: %s", err)
	}

	// config from CLI
	cliConfig := tflint.EmptyConfig()
	cliConfig.Rules["test_in_cli"] = &tflint.RuleConfig{Name: "test_in_cli", Enabled: true, Body: hcl.EmptyBody()}

	runner := tflint.TestRunnerWithConfig(t, map[string]string{}, fileConfig.Merge(cliConfig))

	server := NewGRPCServer(runner, nil, map[string][]byte{})

	// default error check helper
	neverHappend := func(err error) bool { return err != nil }

	tests := []struct {
		Name     string
		Args     func() (string, *hclext.BodySchema)
		Want     *hclext.BodyContent
		ErrCheck func(error) bool
	}{
		{
			Name: "get `test_in_file` rule",
			Args: func() (string, *hclext.BodySchema) {
				return "test_in_file", &hclext.BodySchema{
					Attributes: []hclext.AttributeSchema{{Name: "foo"}},
				}
			},
			Want: &hclext.BodyContent{
				Attributes: hclext.Attributes{
					"foo": &hclext.Attribute{Name: "foo"},
				},
				Blocks: hclext.Blocks{},
			},
			ErrCheck: neverHappend,
		},
		{
			Name: "rule not found",
			Args: func() (string, *hclext.BodySchema) {
				return "not_found", &hclext.BodySchema{
					Attributes: []hclext.AttributeSchema{{Name: "foo"}},
				}
			},
			Want: nil,
			ErrCheck: func(err error) bool {
				return err == nil || err.Error() != "rule `not_found` is not found in config"
			},
		},
		{
			Name: "get rule enabled by CLI",
			Args: func() (string, *hclext.BodySchema) {
				return "test_in_cli", &hclext.BodySchema{
					Attributes: []hclext.AttributeSchema{{Name: "foo"}},
				}
			},
			Want: nil,
			ErrCheck: func(err error) bool {
				return err == nil || err.Error() != "This rule cannot be enabled with the `--enable-rule` option because it lacks the required configuration"
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			content, file, err := server.GetRuleConfigContent(test.Args())
			if test.ErrCheck(err) {
				t.Fatalf("failed to call GetRuleConfigContent: %s", err)
			}

			var gotFile string
			if file != nil {
				gotFile = string(file.Bytes)
			}
			if gotFile != config {
				t.Fatalf("failed to match returned file: %s", gotFile)
			}

			opts := cmp.Options{
				cmpopts.IgnoreFields(hclext.Attribute{}, "Expr", "Range", "NameRange"),
			}
			if diff := cmp.Diff(content, test.Want, opts); diff != "" {
				t.Errorf("diff: %s", diff)
			}
		})
	}
}

func TestEvaluateExpr(t *testing.T) {
	runner := tflint.TestRunner(t, map[string]string{"main.tf": `
variable "foo" {
	default = "bar"
}`})
	rootRunner := tflint.TestRunner(t, map[string]string{"main.tf": `
variable "foo" {
	default = "baz"
}`})

	server := NewGRPCServer(runner, rootRunner, map[string][]byte{})

	// test util functions
	hclExpr := func(expr string) hcl.Expression {
		file, diags := hclsyntax.ParseConfig([]byte(fmt.Sprintf(`expr = %s`, expr)), "test.tf", hcl.InitialPos)
		if diags.HasErrors() {
			panic(diags)
		}
		attributes, diags := file.Body.JustAttributes()
		if diags.HasErrors() {
			panic(diags)
		}
		return attributes["expr"].Expr
	}

	tests := []struct {
		Name string
		Args func() (hcl.Expression, sdk.EvaluateExprOption)
		Want cty.Value
	}{
		{
			Name: "self module context",
			Args: func() (hcl.Expression, sdk.EvaluateExprOption) {
				return hclExpr(`var.foo`), sdk.EvaluateExprOption{WantType: &cty.String, ModuleCtx: sdk.SelfModuleCtxType}
			},
			Want: cty.StringVal("bar"),
		},
		{
			Name: "root module context",
			Args: func() (hcl.Expression, sdk.EvaluateExprOption) {
				return hclExpr(`var.foo`), sdk.EvaluateExprOption{WantType: &cty.String, ModuleCtx: sdk.RootModuleCtxType}
			},
			Want: cty.StringVal("baz"),
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			got, err := server.EvaluateExpr(test.Args())
			if err != nil {
				t.Fatalf("failed to call EvaluateExpr: %s", err)
			}

			if got.GoString() != test.Want.GoString() {
				t.Errorf("expected to get `%s`, but got `%s`", test.Want.GoString(), got.GoString())
			}
		})
	}
}

type testRule struct {
	sdk.DefaultRule
}

func (r *testRule) Name() string           { return "test_rule" }
func (r *testRule) Enabled() bool          { return true }
func (r *testRule) Severity() sdk.Severity { return sdk.ERROR }
func (r *testRule) Check(sdk.Runner) error { return nil }

func TestEmitIssue(t *testing.T) {
	// calculate ranges
	config := `
resource "aws_instance" "foo" {
	instance_type = "t2.micro"
}`
	file, diags := hclsyntax.ParseConfig([]byte(config), "main.tf", hcl.InitialPos)
	if diags.HasErrors() {
		t.Fatalf("failed to parse config file: %s", diags)
	}
	content, diags := file.Body.Content(&hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{{Type: "resource", LabelNames: []string{"type", "name"}}},
	})
	if diags.HasErrors() {
		t.Fatalf("failed to extract config content: %s", diags)
	}
	block := content.Blocks[0]
	content, diags = block.Body.Content(&hcl.BodySchema{Attributes: []hcl.AttributeSchema{{Name: "instance_type"}}})
	if diags.HasErrors() {
		t.Fatalf("failed to extract config nested content: %s", diags)
	}

	resourceDefRange := block.DefRange
	exprRange := content.Attributes["instance_type"].Expr.Range()

	tests := []struct {
		Name string
		Args func() (sdk.Rule, string, hcl.Range)
		Want int
	}{
		{
			Name: "on expr",
			Args: func() (sdk.Rule, string, hcl.Range) {
				return &testRule{}, "error", exprRange
			},
			Want: 1,
		},
		{
			Name: "on non-expr",
			Args: func() (sdk.Rule, string, hcl.Range) {
				return &testRule{}, "error", resourceDefRange
			},
			Want: 1,
		},
		{
			Name: "on another file",
			Args: func() (sdk.Rule, string, hcl.Range) {
				return &testRule{}, "error", hcl.Range{Filename: "not_found.tf"}
			},
			Want: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			runner := tflint.TestRunner(t, map[string]string{"main.tf": config})

			server := NewGRPCServer(runner, nil, map[string][]byte{})

			err := server.EmitIssue(test.Args())
			if err != nil {
				t.Fatalf("failed to call EmitIssue: %s", err)
			}

			if len(runner.Issues) != test.Want {
				t.Errorf("expected to %d issues, but got %d issues", test.Want, len(runner.Issues))
			}
		})
	}
}
