package cmd

import (
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/terraform-linters/tflint/tflint"
)

func Test_checkRunners(t *testing.T) {
	checkErr := errors.New("check failed")

	for _, tc := range []struct {
		name    string
		limit   int
		runners int
		fail    bool
	}{
		{name: "serial success", limit: 1, runners: 8, fail: false},
		{name: "serial failure", limit: 1, runners: 8, fail: true},
		{name: "bounded success", limit: 4, runners: 8, fail: false},
		{name: "bounded failure", limit: 4, runners: 8, fail: true},
		{name: "unbounded success", limit: 8, runners: 8, fail: false},
		{name: "no runners", limit: 4, runners: 0, fail: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			runners := make([]*tflint.Runner, tc.runners)
			var calls, inflight, peak atomic.Int64
			check := func(_ *tflint.Runner) error {
				n := inflight.Add(1)
				for {
					p := peak.Load()
					if n <= p || peak.CompareAndSwap(p, n) {
						break
					}
				}
				calls.Add(1)
				// Hold briefly so concurrent checks overlap, exercising the bound.
				time.Sleep(time.Millisecond)
				inflight.Add(-1)
				if tc.fail {
					return checkErr
				}
				return nil
			}

			err := checkRunners(runners, tc.limit, check)

			// errgroup waits for every goroutine, so all runners are checked even
			// when one fails.
			if got := int(calls.Load()); got != tc.runners {
				t.Errorf("expected check to run for all %d runners, but ran %d", tc.runners, got)
			}

			// Concurrency never exceeds the limit.
			if got := int(peak.Load()); got > tc.limit {
				t.Errorf("peak concurrency %d exceeded limit %d", got, tc.limit)
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
