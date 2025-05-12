package plugin

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hashicorp/go-version"
	hcl "github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/spf13/afero"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/plugin"
	"github.com/terraform-linters/tflint-plugin-sdk/terraform/lang/marks"
	sdk "github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/terraform-linters/tflint/tflint"
	"github.com/zclconf/go-cty/cty"
	"maps"
)

var SDKVersion = version.Must(version.NewVersion(plugin.SDKVersion))

func TestGetModuleContent(t *testing.T) {
	runner := tflint.TestRunner(t, map[string]string{"main.tf": `
resource "aws_instance" "foo" {
	instance_type = "t2.micro"

	ebs_block_device {}
	dynamic "ebs_block_device" {
		for_each = toset(["foo"])
		content {}
	}
}
resource "aws_instance" "baz" {
	count         = 0
	instance_type = "t3.nano"

	ebs_block_device {}
	dynamic "ebs_block_device" {
		for_each = toset(["foo"])
		content {}
	}
}`})
	rootRunner := tflint.TestRunner(t, map[string]string{"main.tf": `
resource "aws_instance" "bar" {
	instance_type = "m5.2xlarge"
}`})

	server := NewGRPCServer(runner, rootRunner, runner.Files(), SDKVersion)

	tests := []struct {
		Name string
		Args func() (*hclext.BodySchema, sdk.GetModuleContentOption)
		Want *hclext.BodyContent
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
								Blocks:     []hclext.BlockSchema{{Type: "ebs_block_device"}},
							},
						},
					},
				}, sdk.GetModuleContentOption{ModuleCtx: sdk.SelfModuleCtxType, Hint: sdk.GetModuleContentHint{ResourceType: "aws_instance"}}
			},
			Want: &hclext.BodyContent{
				Blocks: hclext.Blocks{
					{
						Type:   "resource",
						Labels: []string{"aws_instance", "foo"},
						Body: &hclext.BodyContent{
							Attributes: hclext.Attributes{"instance_type": &hclext.Attribute{Name: "instance_type"}},
							Blocks: hclext.Blocks{
								{
									Type: "ebs_block_device",
									Body: &hclext.BodyContent{Attributes: hclext.Attributes{}, Blocks: hclext.Blocks{}},
								},
								{
									Type: "ebs_block_device",
									Body: &hclext.BodyContent{Attributes: hclext.Attributes{}, Blocks: hclext.Blocks{}},
								},
							},
						},
					},
				},
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
				}, sdk.GetModuleContentOption{ModuleCtx: sdk.RootModuleCtxType, Hint: sdk.GetModuleContentHint{ResourceType: "aws_instance"}}
			},
			Want: &hclext.BodyContent{
				Blocks: hclext.Blocks{
					{
						Type:   "resource",
						Labels: []string{"aws_instance", "bar"},
						Body: &hclext.BodyContent{
							Attributes: hclext.Attributes{"instance_type": &hclext.Attribute{Name: "instance_type"}},
							Blocks:     hclext.Blocks{},
						},
					},
				},
			},
		},
		{
			Name: "expand mode none",
			Args: func() (*hclext.BodySchema, sdk.GetModuleContentOption) {
				return &hclext.BodySchema{
					Blocks: []hclext.BlockSchema{
						{
							Type:       "resource",
							LabelNames: []string{"type", "name"},
							Body: &hclext.BodySchema{
								Attributes: []hclext.AttributeSchema{{Name: "instance_type"}},
								Blocks:     []hclext.BlockSchema{{Type: "ebs_block_device"}},
							},
						},
					},
				}, sdk.GetModuleContentOption{ModuleCtx: sdk.SelfModuleCtxType, ExpandMode: sdk.ExpandModeNone, Hint: sdk.GetModuleContentHint{ResourceType: "aws_instance"}}
			},
			Want: &hclext.BodyContent{
				Blocks: hclext.Blocks{
					{
						Type:   "resource",
						Labels: []string{"aws_instance", "baz"},
						Body: &hclext.BodyContent{
							Attributes: hclext.Attributes{"instance_type": &hclext.Attribute{Name: "instance_type"}},
							Blocks: hclext.Blocks{
								{
									Type: "ebs_block_device",
									Body: &hclext.BodyContent{Attributes: hclext.Attributes{}, Blocks: hclext.Blocks{}},
								},
							},
						},
					},
					{
						Type:   "resource",
						Labels: []string{"aws_instance", "foo"},
						Body: &hclext.BodyContent{
							Attributes: hclext.Attributes{"instance_type": &hclext.Attribute{Name: "instance_type"}},
							Blocks: hclext.Blocks{
								{
									Type: "ebs_block_device",
									Body: &hclext.BodyContent{Attributes: hclext.Attributes{}, Blocks: hclext.Blocks{}},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			got, diags := server.GetModuleContent(test.Args())
			if diags.HasErrors() {
				t.Fatalf("failed to call GetModuleContent: %s", diags)
			}

			opts := cmp.Options{
				cmpopts.IgnoreFields(hclext.Block{}, "TypeRange", "LabelRanges"),
				cmpopts.IgnoreFields(hclext.Attribute{}, "Expr", "NameRange"),
				cmpopts.IgnoreFields(hcl.Range{}, "Start", "End", "Filename"),
				cmpopts.SortSlices(func(i, j *hclext.Block) bool {
					return i.DefRange.String() < j.DefRange.String()
				}),
			}

			if diff := cmp.Diff(got, test.Want, opts); diff != "" {
				t.Errorf("diff: %s", diff)
			}
		})
	}
}

func TestGetFile(t *testing.T) {
	tests := []struct {
		Name    string
		Arg     string
		Changes map[string][]byte
		Want    string
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
		{
			Name: "get file from root module",
			Arg:  "test_on_root1.tf",
			Want: `
resource "aws_instance" "foo" {
	instance_type = "t2.nano"
}`,
		},
		{
			Name: "get autofixed file",
			Arg:  "test1.tf",
			Changes: map[string][]byte{
				"test1.tf": []byte(`
resource "aws_instance" "foo" {
	instance_type = "t3.nano"
}`),
			},
			Want: `
resource "aws_instance" "foo" {
	instance_type = "t3.nano"
}`,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
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
			rootRunner := tflint.TestRunner(t, map[string]string{
				"test_on_root1.tf": `
resource "aws_instance" "foo" {
	instance_type = "t2.nano"
}`,
			})
			files := runner.Files()
			maps.Copy(files, rootRunner.Files())

			server := NewGRPCServer(runner, rootRunner, files, SDKVersion)

			if diags := runner.ApplyChanges(test.Changes); diags.HasErrors() {
				t.Fatal(diags)
			}

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

	server := NewGRPCServer(runner, rootRunner, runner.Files(), SDKVersion)

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
	config := []byte(`
rule "test_in_file" {
	enabled = true
	foo = "bar"
}`)
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	if err := fs.WriteFile(".tflint.hcl", config, os.ModePerm); err != nil {
		t.Fatal(err)
	}
	fileConfig, err := tflint.LoadConfig(fs, ".tflint.hcl")
	if err != nil {
		t.Fatalf("failed to load test config: %s", err)
	}

	// config from CLI
	cliConfig := tflint.EmptyConfig()
	cliConfig.Rules["test_in_cli"] = &tflint.RuleConfig{Name: "test_in_cli", Enabled: true, Body: nil}

	fileConfig.Merge(cliConfig)
	runner := tflint.TestRunnerWithConfig(t, map[string]string{}, fileConfig)

	server := NewGRPCServer(runner, nil, runner.Files(), SDKVersion)

	// default error check helper
	neverHappend := func(err error) bool { return err != nil }

	tests := []struct {
		Name     string
		Args     func() (string, *hclext.BodySchema)
		Want     *hclext.BodyContent
		ErrCheck func(error) bool
	}{
		{
			Name: `get "test_in_file" rule`,
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
			Want:     &hclext.BodyContent{},
			ErrCheck: neverHappend,
		},
		{
			Name: "get rule enabled by CLI without required attribute",
			Args: func() (string, *hclext.BodySchema) {
				return "test_in_cli", &hclext.BodySchema{
					Attributes: []hclext.AttributeSchema{{Name: "foo"}},
				}
			},
			Want:     &hclext.BodyContent{Blocks: hclext.Blocks{}, Attributes: hclext.Attributes{}},
			ErrCheck: neverHappend,
		},
		{
			Name: "get rule enabled by CLI with required attribute",
			Args: func() (string, *hclext.BodySchema) {
				return "test_in_cli", &hclext.BodySchema{
					Attributes: []hclext.AttributeSchema{{Name: "foo", Required: true}},
				}
			},
			Want: nil,
			ErrCheck: func(err error) bool {
				return err == nil || err.Error() != "This rule cannot be enabled with the --enable-rule option because it lacks the required configuration"
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			content, sources, err := server.GetRuleConfigContent(test.Args())
			if test.ErrCheck(err) {
				t.Fatalf("failed to call GetRuleConfigContent: %s", err)
			}

			if string(sources[".tflint.hcl"]) != string(config) {
				t.Fatalf("failed to match returned file: %s", sources[".tflint.hcl"])
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
}

variable "sensitive" {
	sensitive = true
	default   = "foo"
}

variable "ephemeral" {
	ephemeral = true
	default   = "foo"
}

variable "no_default" {}

variable "null" {
	type    = string
	default = null
}
`})
	rootRunner := tflint.TestRunner(t, map[string]string{"main.tf": `
variable "foo" {
	default = "baz"
}`})

	server := NewGRPCServer(runner, rootRunner, runner.Files(), SDKVersion)

	sdkv21 := version.Must(version.NewVersion("0.21.0"))

	// test util functions
	hclExpr := func(expr string) hcl.Expression {
		file, diags := hclsyntax.ParseConfig(fmt.Appendf(nil, `expr = %s`, expr), "test.tf", hcl.InitialPos)
		if diags.HasErrors() {
			panic(diags)
		}
		attributes, diags := file.Body.JustAttributes()
		if diags.HasErrors() {
			panic(diags)
		}
		return attributes["expr"].Expr
	}
	// default error check helper
	neverHappend := func(err error) bool { return err != nil }

	tests := []struct {
		Name       string
		Args       func() (hcl.Expression, sdk.EvaluateExprOption)
		SDKVersion *version.Version
		Want       cty.Value
		ErrCheck   func(error) bool
	}{
		{
			Name: "self module context",
			Args: func() (hcl.Expression, sdk.EvaluateExprOption) {
				return hclExpr(`var.foo`), sdk.EvaluateExprOption{WantType: &cty.String, ModuleCtx: sdk.SelfModuleCtxType}
			},
			Want:     cty.StringVal("bar"),
			ErrCheck: neverHappend,
		},
		{
			Name: "root module context",
			Args: func() (hcl.Expression, sdk.EvaluateExprOption) {
				return hclExpr(`var.foo`), sdk.EvaluateExprOption{WantType: &cty.String, ModuleCtx: sdk.RootModuleCtxType}
			},
			Want:     cty.StringVal("baz"),
			ErrCheck: neverHappend,
		},
		{
			Name: "sensitive value",
			Args: func() (hcl.Expression, sdk.EvaluateExprOption) {
				return hclExpr(`var.sensitive`), sdk.EvaluateExprOption{WantType: &cty.String, ModuleCtx: sdk.SelfModuleCtxType}
			},
			Want:     cty.StringVal("foo").Mark(marks.Sensitive),
			ErrCheck: neverHappend,
		},
		{
			Name: "sensitive value (SDK v0.21)",
			Args: func() (hcl.Expression, sdk.EvaluateExprOption) {
				return hclExpr(`var.sensitive`), sdk.EvaluateExprOption{WantType: &cty.String, ModuleCtx: sdk.SelfModuleCtxType}
			},
			Want:       cty.StringVal("foo").Mark(marks.Sensitive),
			SDKVersion: sdkv21,
			ErrCheck:   neverHappend,
		},
		{
			Name: "no default",
			Args: func() (hcl.Expression, sdk.EvaluateExprOption) {
				return hclExpr(`var.no_default`), sdk.EvaluateExprOption{WantType: &cty.String, ModuleCtx: sdk.SelfModuleCtxType}
			},
			Want:     cty.UnknownVal(cty.String),
			ErrCheck: neverHappend,
		},
		{
			Name: "null",
			Args: func() (hcl.Expression, sdk.EvaluateExprOption) {
				return hclExpr(`var.null`), sdk.EvaluateExprOption{WantType: &cty.String, ModuleCtx: sdk.SelfModuleCtxType}
			},
			Want:     cty.NullVal(cty.String),
			ErrCheck: neverHappend,
		},
		{
			Name: "ephemeral value",
			Args: func() (hcl.Expression, sdk.EvaluateExprOption) {
				return hclExpr(`var.ephemeral`), sdk.EvaluateExprOption{WantType: &cty.String, ModuleCtx: sdk.SelfModuleCtxType}
			},
			Want:     cty.StringVal("foo").Mark(marks.Ephemeral),
			ErrCheck: neverHappend,
		},
		{
			Name: "ephemeral value (SDK v0.21)",
			Args: func() (hcl.Expression, sdk.EvaluateExprOption) {
				return hclExpr(`var.ephemeral`), sdk.EvaluateExprOption{WantType: &cty.String, ModuleCtx: sdk.SelfModuleCtxType}
			},
			Want:       cty.NullVal(cty.NilType),
			SDKVersion: sdkv21,
			ErrCheck: func(err error) bool {
				return err == nil || !errors.Is(err, sdk.ErrSensitive)
			},
		},
		{
			Name: "ephemeral value in object (SDK v0.21)",
			Args: func() (hcl.Expression, sdk.EvaluateExprOption) {
				ty := cty.Object(map[string]cty.Type{"value": cty.String})
				return hclExpr(`{ value = var.ephemeral }`), sdk.EvaluateExprOption{WantType: &ty, ModuleCtx: sdk.SelfModuleCtxType}
			},
			Want:       cty.NullVal(cty.NilType),
			SDKVersion: sdkv21,
			ErrCheck: func(err error) bool {
				return err == nil || !errors.Is(err, sdk.ErrSensitive)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			if test.SDKVersion == nil {
				test.SDKVersion = SDKVersion
			}
			server.clientSDKVersion = test.SDKVersion

			got, err := server.EvaluateExpr(test.Args())
			if test.ErrCheck(err) {
				t.Fatalf("failed to call EvaluateExpr: %s", err)
			}

			if got.GoString() != test.Want.GoString() {
				t.Errorf(`expected to get %s, but got %s`, test.Want.GoString(), got.GoString())
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
		Args func() (sdk.Rule, string, hcl.Range, bool)
		Want int
	}{
		{
			Name: "on expr",
			Args: func() (sdk.Rule, string, hcl.Range, bool) {
				return &testRule{}, "error", exprRange, false
			},
			Want: 1,
		},
		{
			Name: "on non-expr",
			Args: func() (sdk.Rule, string, hcl.Range, bool) {
				return &testRule{}, "error", resourceDefRange, false
			},
			Want: 1,
		},
		{
			Name: "on another file",
			Args: func() (sdk.Rule, string, hcl.Range, bool) {
				return &testRule{}, "error", hcl.Range{Filename: "not_found.tf"}, false
			},
			Want: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			runner := tflint.TestRunner(t, map[string]string{"main.tf": config})

			server := NewGRPCServer(runner, nil, runner.Files(), SDKVersion)

			_, err := server.EmitIssue(test.Args())
			if err != nil {
				t.Fatalf("failed to call EmitIssue: %s", err)
			}

			if len(runner.Issues) != test.Want {
				t.Errorf("expected to %d issues, but got %d issues", test.Want, len(runner.Issues))
			}
		})
	}
}

func TestApplyChanges(t *testing.T) {
	tests := []struct {
		name    string
		files   map[string]string
		changes map[string][]byte
		want    map[string][]byte
	}{
		{
			name: "change file",
			files: map[string]string{
				"main.tf": `
resource "aws_instance" "foo" {
	instance_type = "t2.micro"
}`,
				"variables.tf": `variable "foo" {}`,
			},
			changes: map[string][]byte{
				"main.tf": []byte(`
resource "aws_instance" "foo" {
	instance_type = "t3.nano"
}`),
			},
			want: map[string][]byte{
				"main.tf": []byte(`
resource "aws_instance" "foo" {
	instance_type = "t3.nano"
}`),
				"variables.tf": []byte(`variable "foo" {}`),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			runner := tflint.TestRunner(t, test.files)

			server := NewGRPCServer(runner, nil, runner.Files(), SDKVersion)

			err := server.ApplyChanges(test.changes)
			if err != nil {
				t.Fatalf("failed to call ApplyChanges: %s", err)
			}

			got := server.GetFiles(sdk.SelfModuleCtxType)
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("diff: %s", diff)
			}
		})
	}
}
