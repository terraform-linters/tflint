package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"runtime"
	"sync"
	"time"

	"github.com/terraform-linters/tflint/tflint"
)

// worker is a struct to store the result of each directory
type worker struct {
	dir    string
	stdout io.Reader
	stderr io.Reader
	err    error
}

func (cli *CLI) inspectParallel(opts Options) int {
	workingDirs, err := findWorkingDirs(opts)
	if err != nil {
		cli.formatter.Print(tflint.Issues{}, fmt.Errorf("Failed to find workspaces; %w", err), map[string][]byte{})
		return ExitCodeError
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go cli.registerShutdownHandler(cancel)

	workers, err := spawnWorkers(ctx, workingDirs, opts)
	if err != nil {
		cli.formatter.Print(tflint.Issues{}, fmt.Errorf("Failed to perform workers; %w", err), map[string][]byte{})
		return ExitCodeError
	}

	issues := tflint.Issues{}
	var canceled bool

	for worker := range workers {
		stdout, err := io.ReadAll(worker.stdout)
		if err != nil {
			cli.formatter.PrintErrorParallel(fmt.Errorf("Failed to read stdout in %s; %w", worker.dir, err), cli.sources)
			continue
		}
		stderr, err := io.ReadAll(worker.stderr)
		if err != nil {
			cli.formatter.PrintErrorParallel(fmt.Errorf("Failed to read stderr in %s; %w", worker.dir, err), cli.sources)
			continue
		}
		if worker.err != nil {
			// If the worker is canceled, suppress the error message.
			if errors.Is(worker.err, context.Canceled) {
				canceled = true
				continue
			}

			log.Printf("[DEBUG] Failed to run in %s; %s; stdout=%s", worker.dir, worker.err, stdout)
			cli.formatter.PrintErrorParallel(fmt.Errorf("Failed to run in %s; %w\n\n%s", worker.dir, worker.err, stderr), cli.sources)
			continue
		}

		var workerIssues tflint.Issues
		if err := json.Unmarshal(stdout, &workerIssues); err != nil {
			panic(fmt.Errorf("failed to parse issues in %s; %s; stdout=%s; stderr=%s", worker.dir, err, stdout, stderr))
		}
		issues = append(issues, workerIssues...)

		if len(stderr) > 0 {
			// Regardless of format, output to stderr is synchronized.
			cli.formatter.PrettyPrintStderr(fmt.Sprintf("An output to stderr found in %s\n\n%s\n", worker.dir, stderr))
		}
	}

	if canceled {
		// If the worker is canceled, suppress the error message.
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

// Spawn workers to run in parallel for each directory.
// A worker is a process that runs itself as a child process.
// The number of parallelism is controlled by --max-workers flag. The default is the number of CPUs.
func spawnWorkers(ctx context.Context, workingDirs []string, opts Options) (<-chan worker, error) {
	self, err := os.Executable()
	if err != nil {
		return nil, err
	}

	maxWorkers := runtime.NumCPU()
	if opts.MaxWorkers != nil {
		if c := *opts.MaxWorkers; c > 0 {
			maxWorkers = c
		}
	}

	ch := make(chan worker)
	semaphore := make(chan struct{}, maxWorkers)

	go func() {
		defer close(ch)

		var wg sync.WaitGroup
		for _, wd := range workingDirs {
			wg.Add(1)
			go func(wd string) {
				defer wg.Done()
				spawnWorker(ctx, self, wd, opts, ch, semaphore)
			}(wd)
		}
		wg.Wait()
	}()

	return ch, nil
}

// Spawn a worker process for the given directory.
// When the process is complete, send the results to the given channel.
// If the context is canceled, the started process will be interrupted.
func spawnWorker(ctx context.Context, executable string, workingDir string, opts Options, ch chan<- worker, semaphore chan struct{}) {
	// Blocks from exceeding the maximum number of workers
	select {
	case semaphore <- struct{}{}:
		defer func() {
			<-semaphore
		}()
	case <-ctx.Done():
		log.Printf("[DEBUG] Worker in %s is canceled\n", workingDir)
		ch <- worker{dir: workingDir, stdout: new(bytes.Buffer), stderr: new(bytes.Buffer), err: ctx.Err()}
		return
	}

	stdout, stderr := new(bytes.Buffer), new(bytes.Buffer)
	cmd := exec.CommandContext(ctx, executable, opts.toWorkerCommands(workingDir)...)
	cmd.Stdout, cmd.Stderr = stdout, stderr
	cmd.Cancel = func() error {
		log.Printf("[DEBUG] Worker in %s is terminated\n", workingDir)
		return cmd.Process.Signal(os.Interrupt)
	}
	cmd.WaitDelay = 3 * time.Second
	err := cmd.Run()
	if ctx.Err() != nil {
		// If the context is canceled, return the context error instead of the command error.
		err = ctx.Err()
	}

	ch <- worker{dir: workingDir, stdout: stdout, stderr: stderr, err: err}
}
