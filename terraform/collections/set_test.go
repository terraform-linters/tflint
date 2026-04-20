// SPDX-License-Identifier: MPL-2.0

package collections

import "testing"

func TestSetTracksUniqueness(t *testing.T) {
	set := NewSet[sampleKey]()

	if got := set.Len(); got != 0 {
		t.Fatalf("initial length = %d, want 0", got)
	}

	set.Add(sampleKey("a"))
	set.Add(sampleKey("a"), sampleKey("b"))

	if got := set.Len(); got != 2 {
		t.Fatalf("length after inserts = %d, want 2", got)
	}
	if !set.Has(sampleKey("a")) {
		t.Fatal("expected set to contain a")
	}
	if !set.Has(sampleKey("b")) {
		t.Fatal("expected set to contain b")
	}

	set.Remove(sampleKey("a"))

	if got := set.Len(); got != 1 {
		t.Fatalf("length after removal = %d, want 1", got)
	}
	if set.Has(sampleKey("a")) {
		t.Fatal("did not expect set to contain removed value")
	}
}

func TestZeroValueSetSupportsReadOnlyCalls(t *testing.T) {
	var set Set[string]

	if got := set.Len(); got != 0 {
		t.Fatalf("zero-value length = %d, want 0", got)
	}
	if set.Has("anything") {
		t.Fatal("zero-value set unexpectedly reported membership")
	}
}
