// SPDX-License-Identifier: MPL-2.0

package funcs

import (
	"testing"

	"github.com/terraform-linters/tflint-plugin-sdk/terraform/lang/marks"
	"github.com/zclconf/go-cty/cty"
)

func TestSensitiveMarksTheOuterValue(t *testing.T) {
	cases := []cty.Value{
		cty.NumberIntVal(1),
		cty.UnknownVal(cty.String),
		cty.NullVal(cty.String),
		cty.DynamicVal,
		cty.ListVal([]cty.Value{cty.NumberIntVal(1)}),
		cty.NumberIntVal(1).Mark(marks.Sensitive),
		cty.ListVal([]cty.Value{cty.NumberIntVal(1).Mark(marks.Sensitive)}),
	}

	for _, input := range cases {
		got, err := Sensitive(input)
		if err != nil {
			t.Fatalf("Sensitive(%s) returned error: %v", input.GoString(), err)
		}
		if !got.HasMark(marks.Sensitive) {
			t.Fatalf("Sensitive(%s) did not mark the result sensitive", input.GoString())
		}

		gotRaw, _ := got.Unmark()
		wantRaw, _ := input.Unmark()
		if !gotRaw.RawEquals(wantRaw) {
			t.Fatalf("Sensitive(%s) changed the underlying value", input.GoString())
		}
	}
}

func TestNonsensitiveRemovesOnlyTheOuterSensitiveMark(t *testing.T) {
	cases := []cty.Value{
		cty.NumberIntVal(1).Mark(marks.Sensitive),
		cty.DynamicVal.Mark(marks.Sensitive),
		cty.UnknownVal(cty.String).Mark(marks.Sensitive),
		cty.NullVal(cty.EmptyObject).Mark(marks.Sensitive),
		cty.ListVal([]cty.Value{cty.NumberIntVal(1).Mark(marks.Sensitive)}).Mark(marks.Sensitive),
		cty.NumberIntVal(1),
		cty.NullVal(cty.String),
		cty.DynamicVal,
		cty.UnknownVal(cty.String),
	}

	for _, input := range cases {
		got, err := Nonsensitive(input)
		if err != nil {
			t.Fatalf("Nonsensitive(%s) returned error: %v", input.GoString(), err)
		}
		if got.HasMark(marks.Sensitive) {
			t.Fatalf("Nonsensitive(%s) left the outer mark in place", input.GoString())
		}

		wantRaw, _ := input.Unmark()
		if !got.RawEquals(wantRaw) {
			t.Fatalf("Nonsensitive(%s) changed the underlying value", input.GoString())
		}
	}
}

func TestIssensitiveReflectsMarksAndUnknowns(t *testing.T) {
	cases := []struct {
		name  string
		input cty.Value
		want  cty.Value
	}{
		{name: "marked number", input: cty.NumberIntVal(1).Mark(marks.Sensitive), want: cty.True},
		{name: "plain number", input: cty.NumberIntVal(1), want: cty.False},
		{name: "marked dynamic", input: cty.DynamicVal.Mark(marks.Sensitive), want: cty.True},
		{name: "marked unknown string", input: cty.UnknownVal(cty.String).Mark(marks.Sensitive), want: cty.True},
		{name: "marked null object", input: cty.NullVal(cty.EmptyObject).Mark(marks.Sensitive), want: cty.True},
		{name: "plain null", input: cty.NullVal(cty.String), want: cty.False},
		{name: "dynamic unknown", input: cty.DynamicVal, want: cty.UnknownVal(cty.Bool)},
		{name: "unknown string", input: cty.UnknownVal(cty.String), want: cty.UnknownVal(cty.Bool)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Issensitive(tc.input)
			if err != nil {
				t.Fatalf("Issensitive(%s) returned error: %v", tc.input.GoString(), err)
			}
			if !got.RawEquals(tc.want) {
				t.Fatalf("Issensitive(%s) = %s, want %s", tc.input.GoString(), got.GoString(), tc.want.GoString())
			}
		})
	}
}
