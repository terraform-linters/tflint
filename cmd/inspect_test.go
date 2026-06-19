package cmd

import (
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/terraform-linters/tflint/tflint"
)

func Test_checkRunners_boundsConcurrency(t *testing.T) {
	cases := []struct {
		name    string
		runners int
		workers int
		// wantPeak is the expected number of checks in flight at once:
		// min(workers, runners).
		wantPeak int32
	}{
		{name: "serial", runners: 8, workers: 1, wantPeak: 1},
		{name: "bounded below runner count", runners: 12, workers: 3, wantPeak: 3},
		{name: "workers exceed runner count", runners: 2, workers: 8, wantPeak: 2},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			runners := make([]*tflint.Runner, tc.runners)

			var inflight, max int32
			proceed := make(chan struct{})
			done := make(chan error, 1)

			go func() {
				done <- checkRunners(runners, tc.workers, func(*tflint.Runner) error {
					cur := atomic.AddInt32(&inflight, 1)
					for {
						old := atomic.LoadInt32(&max)
						if cur <= old || atomic.CompareAndSwapInt32(&max, old, cur) {
							break
						}
					}
					<-proceed
					atomic.AddInt32(&inflight, -1)
					return nil
				})
			}()

			// Wait until the expected number of checks are simultaneously in
			// flight. An unbounded implementation would blow past wantPeak here.
			deadline := time.After(2 * time.Second)
			for atomic.LoadInt32(&inflight) < tc.wantPeak {
				select {
				case <-deadline:
					t.Fatalf("only %d checks became in-flight, expected %d", atomic.LoadInt32(&inflight), tc.wantPeak)
				default:
					time.Sleep(time.Millisecond)
				}
			}
			// Give any unbounded extra goroutines a window to start before sampling.
			time.Sleep(20 * time.Millisecond)
			peak := atomic.LoadInt32(&inflight)
			close(proceed)

			if err := <-done; err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if peak > int32(tc.workers) {
				t.Fatalf("expected at most %d concurrent checks, observed %d", tc.workers, peak)
			}
			if max > int32(tc.workers) {
				t.Fatalf("expected peak concurrency to stay within %d, observed %d", tc.workers, max)
			}
		})
	}
}

func Test_checkRunners_returnsFirstError(t *testing.T) {
	runners := make([]*tflint.Runner, 5)
	wantErr := errors.New("boom")

	var calls int32
	err := checkRunners(runners, 2, func(*tflint.Runner) error {
		if atomic.AddInt32(&calls, 1) == 1 {
			return wantErr
		}
		return nil
	})

	if !errors.Is(err, wantErr) {
		t.Fatalf("expected %v, got %v", wantErr, err)
	}
	// Every runner is still drained even after an error.
	if got := atomic.LoadInt32(&calls); got != int32(len(runners)) {
		t.Fatalf("expected all %d runners to be checked, got %d", len(runners), got)
	}
}

func Test_Options_runnerWorkers(t *testing.T) {
	workers2 := 2
	workersZero := 0

	cases := []struct {
		name string
		opts Options
		want int
	}{
		{name: "no-parallel-runners forces serial", opts: Options{NoParallelRunners: true}, want: 1},
		{name: "no-parallel-runners wins over max-workers", opts: Options{NoParallelRunners: true, MaxWorkers: &workers2}, want: 1},
		{name: "max-workers sets the bound", opts: Options{MaxWorkers: &workers2}, want: 2},
		// A non-positive max-workers falls back to the CPU count.
		{name: "non-positive max-workers falls back to NumCPU", opts: Options{MaxWorkers: &workersZero}, want: 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			want := tc.want
			if want == 0 {
				want = tc.opts.maxWorkers()
			}
			if got := tc.opts.runnerWorkers(); got != want {
				t.Fatalf("expected runnerWorkers to return %d, got %d", want, got)
			}
		})
	}
}
