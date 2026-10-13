package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/terraform-linters/tflint/formatter"
)

func TestLocalAndRPCParity(t *testing.T) {
	const native = `
variable "untyped" { default = "hello" }
# tflint-ignore: terraform_typed_variables
variable "ignored" { default = "ignored" }
variable "items" {
  type = list(string)
  default = ["hello"]
}
locals {
  interpolation = "${var.untyped}"
  index = var.items.0
  empty = var.items == []
  duplicate = { key = "first", key = "second" }
  lookup = lookup({ key = "value" }, "key")
}
output "result" { value = local.interpolation }
module "remote" { source = "git::https://example.com/modules.git" }
`
	for _, tc := range []struct {
		name   string
		files  map[string]string
		args   []string
		status int
	}{
		{"recommended", map[string]string{"main.tf": native}, nil, 2},
		{"fix", map[string]string{"main.tf": native}, []string{"--fix"}, 2},
		{"all", map[string]string{"main.tf": native, ".tflint.hcl": "config { call_module_type = \"none\" }\nplugin \"terraform\" {\nenabled = true\npreset = \"all\"\n}"}, nil, 2},
		{"json", map[string]string{"main.tf.json": `{"variable":{"name":{"default":"hello"}},"output":{"name":{"value":"${var.name}"}}}`}, nil, 2},
		{"json array", map[string]string{"main.tf.json": `[{"variable":{"name":{"default":"hello"}}},{"output":{"name":{"value":"${var.name}"}}}]`}, nil, 2},
		{"suppressed", map[string]string{"main.tf": "# tflint-ignore: terraform_typed_variables\nvariable \"ignored\" {}"}, []string{"--only=terraform_typed_variables"}, 0},
		{"modules", map[string]string{"main.tf": "module \"child\" { source = \"./child\" }", "child/main.tf": "variable \"unused\" { default = \"x\" }", ".tflint.hcl": "config { call_module_type = \"local\" }"}, nil, 2},
		{"malformed", map[string]string{"main.tf": "variable \"broken\" {"}, nil, 1},
		{"invalid rule config", map[string]string{"main.tf": native, ".tflint.hcl": "config { call_module_type = \"none\" }\nrule \"terraform_module_pinned_source\" {\nenabled = true\nunexpected = true\n}"}, nil, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			type result struct {
				Status int
				Issues []string
				Errors []string
				Files  map[string]string
			}
			run := func(rpc string) result {
				if err := os.RemoveAll(root); err != nil {
					t.Fatal(err)
				}
				files := map[string]string{".tflint.hcl": "config { call_module_type = \"none\" }"}
				for name, source := range tc.files {
					files[name] = source
				}
				for name, source := range files {
					path := filepath.Join(root, name)
					if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
						t.Fatal(err)
					}
					if err := os.WriteFile(path, []byte(source), 0644); err != nil {
						t.Fatal(err)
					}
				}
				command := exec.Command("tflint", append([]string{"--format=json"}, tc.args...)...)
				command.Dir = root
				for _, env := range os.Environ() {
					if !strings.HasPrefix(env, "TFLINT_BUNDLED_RPC=") {
						command.Env = append(command.Env, env)
					}
				}
				command.Env = append(command.Env, "TFLINT_BUNDLED_RPC="+rpc)
				var stdout, stderr bytes.Buffer
				command.Stdout, command.Stderr = &stdout, &stderr
				_ = command.Run()
				if command.ProcessState == nil {
					t.Fatalf("failed to start tflint: %s", &stderr)
				}
				got := result{Status: command.ProcessState.ExitCode(), Files: map[string]string{}}
				if got.Status != tc.status {
					t.Fatalf("rpc=%s status=%d: %s %s", rpc, got.Status, &stdout, &stderr)
				}
				var output formatter.JSONOutput
				if err := json.Unmarshal(stdout.Bytes(), &output); err != nil {
					t.Fatalf("%s: %s %s", err, &stdout, &stderr)
				}
				for _, issue := range output.Issues {
					encoded, _ := json.Marshal(issue)
					got.Issues = append(got.Issues, string(encoded))
				}
				for _, diagnostic := range output.Errors {
					encoded, _ := json.Marshal(diagnostic)
					got.Errors = append(got.Errors, string(encoded))
				}
				slices.Sort(got.Issues)
				slices.Sort(got.Errors)
				for name := range files {
					source, err := os.ReadFile(filepath.Join(root, name))
					if err != nil {
						t.Fatal(err)
					}
					got.Files[name] = string(source)
				}
				return got
			}
			remote, local := run("1"), run("0")
			if diff := cmp.Diff(remote, local); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
