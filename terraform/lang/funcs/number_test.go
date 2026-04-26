// SPDX-License-Identifier: MPL-2.0

package funcs

import (
	"testing"

	"github.com/terraform-linters/tflint-plugin-sdk/terraform/lang/marks"
	"github.com/zclconf/go-cty/cty"
)

func TestLog(t *testing.T) {
	tests := []struct {
		name string
		num  cty.Value
		base cty.Value
		want cty.Value
	}{
		{
			name: "base 10 of 1",
			num:  cty.NumberFloatVal(1),
			base: cty.NumberFloatVal(10),
			want: cty.NumberFloatVal(0),
		},
		{
			name: "base 10 of 10",
			num:  cty.NumberFloatVal(10),
			base: cty.NumberFloatVal(10),
			want: cty.NumberFloatVal(1),
		},
		{
			name: "zero numerator",
			num:  cty.NumberFloatVal(0),
			base: cty.NumberFloatVal(10),
			want: cty.NegativeInfinity,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := Log(test.num, test.base)
			assertCtyResult(t, got, err, test.want, "")
		})
	}
}

func TestPow(t *testing.T) {
	tests := []struct {
		name  string
		num   cty.Value
		power cty.Value
		want  cty.Value
	}{
		{
			name:  "positive exponent",
			num:   cty.NumberFloatVal(3),
			power: cty.NumberFloatVal(2),
			want:  cty.NumberFloatVal(9),
		},
		{
			name:  "negative exponent",
			num:   cty.NumberFloatVal(2),
			power: cty.NumberFloatVal(-2),
			want:  cty.NumberFloatVal(0.25),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := Pow(test.num, test.power)
			assertCtyResult(t, got, err, test.want, "")
		})
	}
}

func TestSignum(t *testing.T) {
	tests := []struct {
		name string
		num  cty.Value
		want cty.Value
	}{
		{name: "zero", num: cty.NumberFloatVal(0), want: cty.NumberFloatVal(0)},
		{name: "positive", num: cty.NumberFloatVal(12), want: cty.NumberFloatVal(1)},
		{name: "negative", num: cty.NumberFloatVal(-29), want: cty.NumberFloatVal(-1)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := Signum(test.num)
			assertCtyResult(t, got, err, test.want, "")
		})
	}
}

func TestParseInt(t *testing.T) {
	tests := []struct {
		name    string
		num     cty.Value
		base    cty.Value
		want    cty.Value
		wantErr string
	}{
		{
			name: "decimal",
			num:  cty.StringVal("128"),
			base: cty.NumberIntVal(10),
			want: cty.NumberIntVal(128),
		},
		{
			name: "hexadecimal",
			num:  cty.StringVal("FF00"),
			base: cty.NumberIntVal(16),
			want: cty.NumberIntVal(65280),
		},
		{
			name: "propagates marks",
			num:  cty.StringVal("128").Mark("trace"),
			base: cty.NumberIntVal(10).Mark(marks.Sensitive),
			want: cty.NumberIntVal(128).WithMarks(cty.NewValueMarks("trace", marks.Sensitive)),
		},
		{
			name: "unknown base keeps marks",
			num:  cty.StringVal("128").Mark(marks.Sensitive),
			base: cty.UnknownVal(cty.Number).Mark("base"),
			want: cty.UnknownVal(cty.Number).RefineNotNull().WithMarks(cty.NewValueMarks(marks.Sensitive, "base")),
		},
		{
			name:    "invalid base",
			num:     cty.StringVal("128"),
			base:    cty.NumberIntVal(1),
			wantErr: "base must be a whole number between 2 and 62 inclusive",
		},
		{
			name:    "invalid integer",
			num:     cty.StringVal("wat").Mark(marks.Sensitive),
			base:    cty.NumberIntVal(10),
			wantErr: "cannot parse (sensitive value) as a base 10 integer",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := ParseInt(test.num, test.base)
			assertCtyResult(t, got, err, test.want, test.wantErr)
		})
	}
}
