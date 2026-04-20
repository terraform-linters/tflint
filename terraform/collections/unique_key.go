// SPDX-License-Identifier: MPL-2.0

package collections

// UniqueKey is a comparable identifier for values of T.
//
// Implementations are expected to use a comparable concrete type so they can
// safely act as keys in the Set storage map.
type UniqueKey[T any] interface {
	IsUniqueKey(T)
}

// UniqueKeyer is implemented by values that can derive their own stable key.
type UniqueKeyer[T any] interface {
	UniqueKey() UniqueKey[T]
}

// comparableKey adapts any comparable value into a UniqueKey without changing
// its equality semantics.
type comparableKey[T comparable] struct {
	value T
}

func (comparableKey[T]) IsUniqueKey(T) {}

func comparableKeyFunc[T comparable](value T) UniqueKey[T] {
	return comparableKey[T]{value: value}
}
