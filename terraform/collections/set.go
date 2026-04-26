// SPDX-License-Identifier: MPL-2.0

package collections

import "iter"

// Set stores unique values using a caller-supplied identity function.
//
// The zero value behaves like an empty read-only set. Callers must construct a
// set with NewSet, NewSetFunc, or NewSetCmp before mutating it.
type Set[T any] struct {
	members map[UniqueKey[T]]T
	key     func(T) UniqueKey[T]
}

// NewSet constructs a set for values that know how to derive their own key.
func NewSet[T UniqueKeyer[T]](values ...T) Set[T] {
	return NewSetFunc(T.UniqueKey, values...)
}

// NewSetFunc constructs a set using keyFunc to define value identity.
func NewSetFunc[T any](keyFunc func(T) UniqueKey[T], values ...T) Set[T] {
	set := Set[T]{
		members: make(map[UniqueKey[T]]T),
		key:     keyFunc,
	}
	set.Add(values...)
	return set
}

// NewSetCmp constructs a set for comparable values using Go equality.
func NewSetCmp[T comparable](values ...T) Set[T] {
	return NewSetFunc(comparableKeyFunc[T], values...)
}

// Has reports whether value is present in the set.
func (s Set[T]) Has(value T) bool {
	if len(s.members) == 0 {
		return false
	}
	_, ok := s.members[s.key(value)]
	return ok
}

// Add inserts or replaces values by their identity key.
func (s Set[T]) Add(values ...T) {
	for _, value := range values {
		s.members[s.key(value)] = value
	}
}

// Remove deletes value if an equivalent entry exists.
func (s Set[T]) Remove(value T) {
	delete(s.members, s.key(value))
}

// All returns an iterator over the current set members.
func (s Set[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) {
		for _, value := range s.members {
			if !yield(value) {
				return
			}
		}
	}
}

// Len returns the number of unique values in the set.
func (s Set[T]) Len() int {
	return len(s.members)
}
