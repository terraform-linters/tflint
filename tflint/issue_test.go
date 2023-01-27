package tflint

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	hcl "github.com/hashicorp/hcl/v2"
	sdk "github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

func Test_NewSeverity(t *testing.T) {
	tests := []struct {
		Name        string
		Sev         string
		ExpectedSev sdk.Severity
		ExpectedErr string
	}{
		{
			Name:        "new error severity",
			Sev:         "error",
			ExpectedSev: sdk.ERROR,
			ExpectedErr: "<nil>",
		},
		{
			Name:        "new warning severity",
			Sev:         "warning",
			ExpectedSev: sdk.WARNING,
			ExpectedErr: "<nil>",
		},
		{
			Name:        "new notice severity",
			Sev:         "notice",
			ExpectedSev: sdk.NOTICE,
			ExpectedErr: "<nil>",
		},
		{
			Name:        "unrecognized severity",
			Sev:         "test",
			ExpectedSev: sdk.NOTICE,
			ExpectedErr: "test is not a recognized severity",
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			s, err := NewSeverity(test.Sev)

			if !cmp.Equal(s, test.ExpectedSev) {
				t.Fatalf("Failed: diff=%s", cmp.Diff(s, test.ExpectedSev))
			}

			if !cmp.Equal(fmt.Sprint(err), test.ExpectedErr) {
				t.Fatalf("Failed: diff=%s", cmp.Diff(fmt.Sprint(err), test.ExpectedErr))
			}
		})
	}
}

func Test_SeverityToInt32(t *testing.T) {
	tests := []struct {
		Name        string
		Sev         sdk.Severity
		ExpectedInt int32
		ExpectedErr string
	}{
		{
			Name:        "convert error severity",
			Sev:         sdk.ERROR,
			ExpectedInt: 2,
			ExpectedErr: "<nil>",
		},
		{
			Name:        "convert warning severity",
			Sev:         sdk.WARNING,
			ExpectedInt: 1,
			ExpectedErr: "<nil>",
		},
		{
			Name:        "convert notice severity",
			Sev:         sdk.NOTICE,
			ExpectedInt: 0,
			ExpectedErr: "<nil>",
		},
		{
			Name:        "convert unrecognized severity",
			Sev:         10,
			ExpectedInt: 0,
			ExpectedErr: "Unknown",
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			s, err := SeverityToInt32(test.Sev)

			if !cmp.Equal(s, test.ExpectedInt) {
				t.Fatalf("Failed: diff=%s", cmp.Diff(s, test.ExpectedInt))
			}

			if !strings.Contains(fmt.Sprint(err), test.ExpectedErr) {
				t.Fatalf("Failed: diff=%s", cmp.Diff(fmt.Sprint(err), test.ExpectedErr))
			}
		})
	}
}

func Test_Sort(t *testing.T) {
	issues := Issues{
		{
			Rule:    &testRule{},
			Message: "test",
			Range: hcl.Range{
				Filename: "test2.tf",
				Start:    hcl.Pos{Line: 1, Column: 1},
				End:      hcl.Pos{Line: 1, Column: 2},
			},
		},
		{
			Rule:    &testRule{},
			Message: "test",
			Range: hcl.Range{
				Filename: "test1.tf",
				Start:    hcl.Pos{Line: 2, Column: 1},
				End:      hcl.Pos{Line: 2, Column: 2},
			},
		},
		{
			Rule:    &testRule{},
			Message: "test",
			Range: hcl.Range{
				Filename: "test1.tf",
				Start:    hcl.Pos{Line: 1, Column: 1},
				End:      hcl.Pos{Line: 1, Column: 2},
			},
		},
		{
			Rule:    &testRule{},
			Message: "test",
			Range: hcl.Range{
				Filename: "test1.tf",
				Start:    hcl.Pos{Line: 1, Column: 1},
				End:      hcl.Pos{Line: 2, Column: 3},
			},
		},
		{
			Rule:    &testRule{},
			Message: "test",
			Range: hcl.Range{
				Filename: "test1.tf",
				Start:    hcl.Pos{Line: 2, Column: 2},
				End:      hcl.Pos{Line: 2, Column: 3},
			},
		},
		{
			Rule:    &testRule{},
			Message: "test",
			Range: hcl.Range{
				Filename: "test1.tf",
				Start:    hcl.Pos{Line: 1, Column: 1},
				End:      hcl.Pos{Line: 2, Column: 4},
			},
		},
	}

	expected := Issues{
		{
			Rule:    &testRule{},
			Message: "test",
			Range: hcl.Range{
				Filename: "test1.tf",
				Start:    hcl.Pos{Line: 1, Column: 1},
				End:      hcl.Pos{Line: 2, Column: 4},
			},
		},
		{
			Rule:    &testRule{},
			Message: "test",
			Range: hcl.Range{
				Filename: "test1.tf",
				Start:    hcl.Pos{Line: 1, Column: 1},
				End:      hcl.Pos{Line: 2, Column: 3},
			},
		},
		{
			Rule:    &testRule{},
			Message: "test",
			Range: hcl.Range{
				Filename: "test1.tf",
				Start:    hcl.Pos{Line: 1, Column: 1},
				End:      hcl.Pos{Line: 1, Column: 2},
			},
		},
		{
			Rule:    &testRule{},
			Message: "test",
			Range: hcl.Range{
				Filename: "test1.tf",
				Start:    hcl.Pos{Line: 2, Column: 1},
				End:      hcl.Pos{Line: 2, Column: 2},
			},
		},
		{
			Rule:    &testRule{},
			Message: "test",
			Range: hcl.Range{
				Filename: "test1.tf",
				Start:    hcl.Pos{Line: 2, Column: 2},
				End:      hcl.Pos{Line: 2, Column: 3},
			},
		},
		{
			Rule:    &testRule{},
			Message: "test",
			Range: hcl.Range{
				Filename: "test2.tf",
				Start:    hcl.Pos{Line: 1, Column: 1},
				End:      hcl.Pos{Line: 1, Column: 2},
			},
		},
	}

	got := issues.Sort()
	if !cmp.Equal(got, expected) {
		t.Fatalf("Failed: diff=%s", cmp.Diff(got, expected))
	}
}
