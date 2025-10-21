package lzval

import (
	"context"
	"iter"
)

type (
	// ImmutableMap represents a map of string keys to values.
	ImmutableMap struct {
		m LoadableMap
	}

	// ImmutableSlice represents a slice of values.
	ImmutableSlice struct {
		s LoadableSlice
	}
)

// Get retrieves a ImmutableValue representing the value associated with the
// given key in the ImmutableMap. It returns nil if the field does not exist
// or if the receiver is nil.
func (m *ImmutableMap) Get(key string) (ValueLoader, bool) {
	if m == nil {
		return nil, false
	}
	v, ok := m.m[key]
	return v, ok
}

// Len returns the number of fields in the ImmutableMap.
// If the ImmutableMap is nil, it returns 0.
func (m *ImmutableMap) Len() int {
	if m == nil {
		return 0
	}
	return len(m.m)
}

// All iterates over all key, value pairs in the ImmutableMap.
// Like iterating over a vanilla Go map, the order of
// pairs is non-deterministic.
func (m *ImmutableMap) All() iter.Seq2[string, ValueLoader] {
	return func(yield func(string, ValueLoader) bool) {
		if m == nil {
			return
		}
		for k, v := range m.m {
			if !yield(k, v) {
				return
			}
		}
	}
}

func (m *ImmutableMap) Mutable() *Map {
	if m == nil {
		return nil
	}
	return &Map{base: m}
}

func (m *ImmutableMap) RecursiveLoad(ctx context.Context) (map[string]any, error) {
	return recursiveLoadMap(ctx, m.m)
}

// At retrieves the ImmutableValue representing the value at the specified
// index. Like a vanilla Go slice, if the index is out of bounds, it will panic.
func (is *ImmutableSlice) At(index int) ValueLoader {
	if is == nil {
		_ = []struct{}{}[index] // induce equivalent panic
		return nil
	}
	return is.s[index]
}

// Len returns the number of elements in the ImmutableSlice.
// If the ImmutableSlice is nil, it returns 0.
func (is *ImmutableSlice) Len() int {
	if is == nil {
		return 0
	}
	return len(is.s)
}

// SubSlice returns a new ImmutableSlice that is a slice of the original ImmutableSlice,
// equivalent to calling `mySlice[l:r]`. Like a vanilla Go slice, if an index
// is out of bounds, it will panic.
func (is *ImmutableSlice) SubSlice(l, r int) *ImmutableSlice {
	if is == nil {
		_ = []struct{}{}[l:r] // induce equivalent panic
		return nil
	}
	return &ImmutableSlice{s: is.s[l:r]}
}

// All iterates over all elements in the ImmutableSlice.
func (is *ImmutableSlice) All() iter.Seq2[int, ValueLoader] {
	return func(yield func(int, ValueLoader) bool) {
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

func (m *ImmutableSlice) Mutable() *Slice {
	if m == nil {
		return nil
	}
	return &Slice{base: m}
}

func (m *ImmutableSlice) RecursiveLoad(ctx context.Context) ([]any, error) {
	return recursiveLoadSlice(ctx, m.s)
}
