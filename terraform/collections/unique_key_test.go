// SPDX-License-Identifier: MPL-2.0

package collections

import "testing"

type sampleKey string

var _ UniqueKey[sampleKey] = sampleKey("")
var _ UniqueKeyer[sampleKey] = sampleKey("")

func (sampleKey) IsUniqueKey(sampleKey) {}

func (k sampleKey) UniqueKey() UniqueKey[sampleKey] {
	return k
}

func TestSampleKeyImplementsCollectionKeyContracts(t *testing.T) {
	if got, want := sampleKey("example").UniqueKey(), UniqueKey[sampleKey](sampleKey("example")); got != want {
		t.Fatalf("unexpected unique key: got %v want %v", got, want)
	}
}
