package cmd

import (
	"context"
	"fmt"
	"io"
	"maps"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"

	"github.com/spf13/afero"
	"github.com/terraform-linters/tflint/formatter"
	"github.com/terraform-linters/tflint/terraform"
	"github.com/terraform-linters/tflint/tflint"
)

type worker struct {
	dir     string
	issues  tflint.Issues
	sources map[string][]byte
	err     error
}

func (cli *CLI) inspectParallel(opts Options) int {
	workingDirs, err := findWorkingDirs(opts)
	if err != nil {
		cli.formatter.Print(tflint.Issues{}, fmt.Errorf("Failed to find workspaces; %w", err), map[string][]byte{})
		return ExitCodeError
	}

	parser := terraform.NewParser(afero.NewOsFs())
	modules := workingDirs[:0]
	for _, dir := range workingDirs {
		if parser.IsConfigDir("", dir) {
			modules = append(modules, dir)
		}
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	workers := cli.spawnWorkers(ctx, modules, opts)
	issues := tflint.Issues{}
	for result := range workers {
		maps.Copy(cli.sources, result.sources)
		if result.err != nil {
			cli.formatter.PrintErrorParallel(fmt.Errorf("Failed to run in %s; %w", result.dir, result.err), result.sources)
			continue
		}
		fillNoRangeIssueFilenames(result.dir, result.issues)
		issues = append(issues, result.issues...)
	}
	if ctx.Err() != nil {
		return ExitCodeError
	}
	var force bool
	if opts.Force != nil {
		force = *opts.Force
	}
	if err := cli.formatter.PrintParallel(issues, cli.sources); err != nil {
		return ExitCodeError
	}
	if len(issues) > 0 && !force && exceedsMinimumFailure(issues, opts.MinimumFailureSeverity) {
		return ExitCodeIssuesFound
	}
	return ExitCodeOK
}

func fillNoRangeIssueFilenames(dir string, issues tflint.Issues) {
	for _, issue := range issues {
		if issue.Range.Filename == "" {
			issue.Range.Filename = dir
		}
		for idx := range issue.Callers {
			if issue.Callers[idx].Filename == "" {
				issue.Callers[idx].Filename = dir
			}
		}
	}
}

func (cli *CLI) spawnWorkers(ctx context.Context, dirs []string, opts Options) <-chan worker {
	maxWorkers := runtime.NumCPU()
	if opts.MaxWorkers != nil && *opts.MaxWorkers > 0 {
		maxWorkers = *opts.MaxWorkers
	}
	maxWorkers = min(maxWorkers, len(dirs))
	jobs := make(chan string)
	results := make(chan worker)
	var wg sync.WaitGroup
	for range maxWorkers {
		wg.Go(func() {
			for dir := range jobs {
				if ctx.Err() != nil {
					return
				}
				child := &CLI{
					outStream:          io.Discard,
					errStream:          io.Discard,
					originalWorkingDir: cli.originalWorkingDir,
					sources:            map[string][]byte{},
					formatter:          &formatter.Formatter{},
				}
				workerOpts := opts
				workerOpts.Chdir = dir
				workerOpts.ActAsWorker = true
				issues, changes, err := child.inspectDirectory(ctx, workerOpts)
				if err == nil && opts.Fix && ctx.Err() == nil {
					err = writeChanges(changes)
				}
				sources := child.sources
				if child.loader != nil {
					sources = child.loader.Sources()
				}
				results <- worker{dir: dir, issues: issues, sources: sources, err: err}
			}
		})
	}
	go func() {
		defer close(jobs)
		for _, dir := range dirs {
			select {
			case jobs <- dir:
			case <-ctx.Done():
				return
			}
		}
	}()
	go func() {
		wg.Wait()
		close(results)
	}()
	return results
}
