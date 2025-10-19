package lzval

import (
	"iter"
)

type (
	// ImmutableMap represents a map of string keys to values.
	ImmutableMap[T any] struct {
		m map[string]T
	}

	// ImmutableSlice represents a slice of values.
	ImmutableSlice[T any] struct {
		s []T
	}
)

// Get retrieves a ImmutableValue representing the value associated with the
// given key in the ImmutableMap. It returns nil if the field does not exist
// or if the receiver is nil.
func (im *ImmutableMap[T]) Get(key string) (T, bool) {
	if im == nil {
		var t T
		return t, false
	}
	v, ok := im.m[key]
	return v, ok
}

// Len returns the number of fields in the ImmutableMap.
// If the ImmutableMap is nil, it returns 0.
func (im *ImmutableMap[T]) Len() int {
	if im == nil {
		return 0
	}
	return len(im.m)
}

// All iterates over all key, value pairs in the ImmutableMap.
// Like iterating over a vanilla Go map, the order of
// pairs is non-deterministic.
func (im *ImmutableMap[T]) All() iter.Seq2[string, T] {
	return func(yield func(string, T) bool) {
		if im == nil {
			return
		}
		for k, v := range im.m {
			if !yield(k, v) {
				return
			}
		}
	}
}

func (im *ImmutableMap[T]) Mutable() *Map[T] {
	if im == nil {
		return nil
	}
	return &Map[T]{
		base: im,
	}
}

// At retrieves the ImmutableValue representing the value at the specified
// index. Like a vanilla Go slice, if the index is out of bounds, it will panic.
func (is *ImmutableSlice[T]) At(index int) T {
	if is == nil {
		_ = []struct{}{}[index] // induce equivalent panic
		return *new(T)
	}
	return is.s[index]
}

// Len returns the number of elements in the ImmutableSlice.
// If the ImmutableSlice is nil, it returns 0.
func (is *ImmutableSlice[T]) Len() int {
	if is == nil {
		return 0
	}
	return len(is.s)
}

// SubSlice returns a new ImmutableSlice that is a slice of the original ImmutableSlice,
// equivalent to calling `mySlice[l:r]`. Like a vanilla Go slice, if an index
// is out of bounds, it will panic.
func (is *ImmutableSlice[T]) SubSlice(l, r int) *ImmutableSlice[T] {
	if is == nil {
		_ = []struct{}{}[l:r] // induce equivalent panic
		return nil
	}
	return &ImmutableSlice[T]{s: is.s[l:r]}
}

// All iterates over all elements in the ImmutableSlice.
func (is *ImmutableSlice[T]) All() iter.Seq2[int, T] {
	return func(yield func(int, T) bool) {
		if is == nil {
			return
		}
		for i, v := range is.s {
			if !yield(i, v) {
				return
			}
		}
	}
}
