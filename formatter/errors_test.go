package formatter

import (
	"errors"
	"fmt"
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
)

func Test_mapErrors(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected []string
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: []string{},
		},
		{
			name:     "single error",
			err:      fmt.Errorf("test error"),
			expected: []string{"error: test error"},
		},
		{
			name: "joined errors",
			err: errors.Join(
				fmt.Errorf("error 1"),
				fmt.Errorf("error 2"),
			),
			expected: []string{"error: error 1", "error: error 2"},
		},
		{
			name: "hcl diagnostics",
			err: hcl.Diagnostics{
				{
					Severity: hcl.DiagError,
					Summary:  "test summary",
					Detail:   "test detail",
					Subject: &hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1},
						End:      hcl.Pos{Line: 1, Column: 5},
					},
				},
			},
			expected: []string{"diagnostic: test summary - test detail"},
		},
		{
			name: "mixed errors",
			err: errors.Join(
				fmt.Errorf("generic error"),
				hcl.Diagnostics{
					{
						Severity: hcl.DiagError,
						Summary:  "hcl error",
						Detail:   "detail",
						Subject: &hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 1},
							End:      hcl.Pos{Line: 1, Column: 5},
						},
					},
				},
			),
			expected: []string{"error: generic error", "diagnostic: hcl error - detail"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := mapErrors(tt.err, errorMapper[string]{
				diagnostic: func(diag *hcl.Diagnostic) string {
					return fmt.Sprintf("diagnostic: %s - %s", diag.Summary, diag.Detail)
				},
				error: func(err error) string {
					return fmt.Sprintf("error: %s", err.Error())
				},
			})

			if len(results) != len(tt.expected) {
				t.Fatalf("expected %d results, got %d", len(tt.expected), len(results))
			}

			for i, expected := range tt.expected {
				if results[i] != expected {
					t.Errorf("result[%d]: expected %q, got %q", i, expected, results[i])
				}
			}
		})
	}
}
