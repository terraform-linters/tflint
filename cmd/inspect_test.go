package cmd

import (
	"errors"
	"sync/atomic"
	"testing"

	"github.com/terraform-linters/tflint/tflint"
)

func Test_checkRunners(t *testing.T) {
	checkErr := errors.New("check failed")

	for _, tc := range []struct {
		name     string
		parallel bool
		runners  int
		fail     bool
		// wantAllChecked is whether every runner is deterministically checked.
		// In parallel mode the first error is returned without waiting for the
		// remaining goroutines, so on failure the number of checks actually run
		// is not deterministic.
		wantAllChecked bool
	}{
		{name: "parallel success", parallel: true, runners: 8, fail: false, wantAllChecked: true},
		{name: "serial success", parallel: false, runners: 8, fail: false, wantAllChecked: true},
		{name: "serial failure", parallel: false, runners: 8, fail: true, wantAllChecked: true},
		{name: "no runners", parallel: true, runners: 0, fail: true, wantAllChecked: true},
		{name: "parallel failure", parallel: true, runners: 8, fail: true, wantAllChecked: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			runners := make([]*tflint.Runner, tc.runners)
			var calls atomic.Int64
			check := func(_ *tflint.Runner) error {
				calls.Add(1)
				if tc.fail {
					return checkErr
				}
				return nil
			}

			err := checkRunners(runners, tc.parallel, check)

			if tc.wantAllChecked {
				if got := int(calls.Load()); got != tc.runners {
					t.Errorf("expected check to run for all %d runners, but ran %d", tc.runners, got)
				}
			}

			if tc.fail && tc.runners > 0 {
				if !errors.Is(err, checkErr) {
					t.Errorf("expected error %v, got %v", checkErr, err)
				}
			} else if err != nil {
				t.Errorf("expected no error, got %v", err)
			}
		})
	}
}
