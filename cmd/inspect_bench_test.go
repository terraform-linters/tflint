package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/spf13/afero"
	sdk "github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/terraform-linters/tflint/plugin"
	"github.com/terraform-linters/tflint/terraform"
	"github.com/terraform-linters/tflint/tflint"
	"github.com/zclconf/go-cty/cty"
)

// benchModuleCounts is the fixture size swept by the benchmarks. The break
// dimension for the per-runner storm is the number of module-call runners, so
// the suite scales it across orders of magnitude.
var benchModuleCounts = []int{1, 10, 100, 1000}

// benchLimits sweeps the per-runner concurrency bound: serial, bounded to the
// number of CPUs, and unbounded (the pre-bound behavior).
var benchLimits = []struct {
	name  string
	limit int
}{
	{name: "serial", limit: 1},
	{name: "cpu", limit: runtime.GOMAXPROCS(0)},
	{name: "unbounded", limit: -1},
}

// BenchmarkCheckRunners measures the per-runner check fan-out in isolation with
// a synthetic check of fixed latency. The deterministic "peak-concurrency"
// metric is the authoritative signal: it records the maximum number of checks
// running at once. With a finite limit, peak concurrency is capped at the limit
// regardless of how many runners exist; unbounded, it equals the runner count.
func BenchmarkCheckRunners(b *testing.B) {
	const checkLatency = 50 * time.Microsecond

	for _, lim := range benchLimits {
		for _, modules := range benchModuleCounts {
			runners := make([]*tflint.Runner, modules)
			b.Run(fmt.Sprintf("limit=%s/modules=%d", lim.name, modules), func(b *testing.B) {
				var calls, inflight, peak atomic.Int64
				check := func(_ *tflint.Runner) error {
					updatePeakConcurrency(&peak, inflight.Add(1))
					calls.Add(1)
					time.Sleep(checkLatency)
					inflight.Add(-1)
					return nil
				}

				b.ReportAllocs()
				iterations := 0
				for b.Loop() {
					if err := checkRunners(runners, lim.limit, check); err != nil {
						b.Fatal(err)
					}
					iterations++
				}

				b.ReportMetric(float64(peak.Load()), "peak-concurrency")
				if iterations > 0 {
					b.ReportMetric(float64(calls.Load())/float64(iterations), "check-calls/op")
				}
			})
		}
	}
}

// BenchmarkBuildRunners measures runner construction over a generated fixture.
// Runner construction is serial today, so this establishes the baseline cost
// against which any future parallelization is judged.
func BenchmarkBuildRunners(b *testing.B) {
	for _, modules := range benchModuleCounts {
		b.Run(fmt.Sprintf("modules=%d", modules), func(b *testing.B) {
			dir := b.TempDir()
			generateModulesFixture(b, dir, modules)
			b.Chdir(dir)
			config := tflint.EmptyConfig()

			b.ReportAllocs()
			var runnerCount int
			for b.Loop() {
				loader, err := terraform.NewLoader(afero.Afero{Fs: afero.NewOsFs()}, dir)
				if err != nil {
					b.Fatal(err)
				}
				_, moduleRunners, err := tflint.BuildRunners(loader, config, dir, ".")
				if err != nil {
					b.Fatal(err)
				}
				runnerCount = len(moduleRunners)
			}

			b.ReportMetric(float64(runnerCount), "runners")
		})
	}
}

// BenchmarkInspectFanout drives the per-runner check fan-out end-to-end over a
// generated fixture. Each check evaluates a root-context expression through a
// real plugin GRPCServer (the shared-rootRunner path that issue #2094 concerns)
// and then sleeps to model the remote latency of an I/O-bound deep check. The
// "peak-concurrency" metric shows how many checks run at once: bounded, it is
// capped at the limit; unbounded, it equals the number of module runners, which
// is the storm this work bounds.
func BenchmarkInspectFanout(b *testing.B) {
	const remoteLatency = time.Millisecond

	for _, lim := range benchLimits {
		if lim.limit == 1 {
			// Serial is covered by BenchmarkCheckRunners; skip it here to keep
			// the end-to-end sweep focused on bounded vs unbounded.
			continue
		}
		for _, modules := range benchModuleCounts {
			b.Run(fmt.Sprintf("limit=%s/modules=%d", lim.name, modules), func(b *testing.B) {
				dir := b.TempDir()
				generateModulesFixture(b, dir, modules)
				b.Chdir(dir)

				loader, err := terraform.NewLoader(afero.Afero{Fs: afero.NewOsFs()}, dir)
				if err != nil {
					b.Fatal(err)
				}
				rootRunner, moduleRunners, err := tflint.BuildRunners(loader, tflint.EmptyConfig(), dir, ".")
				if err != nil {
					b.Fatal(err)
				}
				files := loader.Files()
				expr := parseBenchExpr(b, "local.tag")
				wantType := cty.String

				var inflight, peak atomic.Int64
				check := func(runner *tflint.Runner) error {
					updatePeakConcurrency(&peak, inflight.Add(1))
					defer inflight.Add(-1)

					server := plugin.NewGRPCServer(runner, rootRunner, files, nil)
					if _, err := server.EvaluateExpr(expr, sdk.EvaluateExprOption{ModuleCtx: sdk.RootModuleCtxType, WantType: &wantType}); err != nil {
						return err
					}
					time.Sleep(remoteLatency)
					return nil
				}

				b.ReportAllocs()
				for b.Loop() {
					if err := checkRunners(moduleRunners, lim.limit, check); err != nil {
						b.Fatal(err)
					}
				}

				b.ReportMetric(float64(peak.Load()), "peak-concurrency")
				b.ReportMetric(float64(len(moduleRunners)), "runners")
			})
		}
	}
}

// updatePeakConcurrency records n as the new peak if it exceeds the current one.
func updatePeakConcurrency(peak *atomic.Int64, n int64) {
	for {
		current := peak.Load()
		if n <= current || peak.CompareAndSwap(current, n) {
			return
		}
	}
}

func parseBenchExpr(tb testing.TB, src string) hcl.Expression {
	tb.Helper()
	expr, diags := hclsyntax.ParseExpression([]byte(src), "bench.tf", hcl.InitialPos)
	if diags.HasErrors() {
		tb.Fatal(diags)
	}
	return expr
}

// generateModulesFixture writes a Terraform configuration to dir: a root module
// that calls a child module `modules` times, plus the .terraform/modules
// manifest the loader uses to resolve those calls. Each call expands to one
// module runner, so the fixture yields exactly `modules` runners. The root and
// child modules each declare interdependent locals so checks exercise the real
// evaluator rather than a trivial expression.
func generateModulesFixture(tb testing.TB, dir string, modules int) {
	tb.Helper()

	childDir := filepath.Join(dir, "child")
	if err := os.MkdirAll(childDir, 0o755); err != nil {
		tb.Fatal(err)
	}
	writeBenchFile(tb, filepath.Join(childDir, "main.tf"), `
variable "instance_type" {
  type = string
}

locals {
  base    = "prefix-${var.instance_type}"
  derived = "${local.base}-suffix"
}

resource "aws_instance" "this" {
  instance_type = local.derived
}
`)

	type manifestEntry struct {
		Key    string
		Source string
		Dir    string
	}
	manifest := struct{ Modules []manifestEntry }{
		Modules: []manifestEntry{{Key: "", Source: "", Dir: "."}},
	}

	var root strings.Builder
	root.WriteString(`
locals {
  region = "us-east-1"
  env    = "prod"
  tag    = "${local.region}-${local.env}"
}
`)
	for i := range modules {
		name := fmt.Sprintf("mod%d", i)
		fmt.Fprintf(&root, "\nmodule %q {\n  source        = \"./child\"\n  instance_type = \"t2.micro\"\n}\n", name)
		manifest.Modules = append(manifest.Modules, manifestEntry{Key: name, Source: "./child", Dir: "child"})
	}
	writeBenchFile(tb, filepath.Join(dir, "main.tf"), root.String())

	manifestDir := filepath.Join(dir, ".terraform", "modules")
	if err := os.MkdirAll(manifestDir, 0o755); err != nil {
		tb.Fatal(err)
	}
	data, err := json.Marshal(manifest)
	if err != nil {
		tb.Fatal(err)
	}
	writeBenchFile(tb, filepath.Join(manifestDir, "modules.json"), string(data))
}

func writeBenchFile(tb testing.TB, path, content string) {
	tb.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		tb.Fatal(err)
	}
}
