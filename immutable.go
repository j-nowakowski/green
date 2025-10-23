package green

import (
	"fmt"
	"iter"
	"maps"
	"slices"
)

/*
	To-do items:
	- Instead of one-time shallow copy, consider
	  using a map which tracks overwrites. This was
	  the original design but it got too confusing handling
	  nested vanilla Go values vs ImmutableValues/Values.
	  This should be doable. After this change, concurrency
	  safety might be easier to implement.
	- Optimizations which avoid unnecessary wrapping and
	  then unwrapping Immutable/Mutables
	- Concurrency safety
*/

type (
	// ImmutableValue has a concrete type is either ImmutableMap,
	// ImmutableSlice, or a literal Go type.
	ImmutableValue any

	// ImmutableMap represents an immutable map of key-value pairs.
	//
	// The methods of ImmutableMap are NOT SAFE for concurrent use.
	// This is a planned future enhancement.
	ImmutableMap struct {
		base   map[string]any
		copied bool
		len    int // to avoid needing lock on Len() (lock not currently implemented)
	}

	// ImmutableSlice represents a slice of values.
	//
	// The methods of ImmutableMap are NOT SAFE for concurrent use.
	// This is a planned future enhancement.
	ImmutableSlice struct {
		base   []any
		copied bool
		len    int // to avoid needing lock on Len() (lock not currently implemented)
	}
)

// NewImmutableMap wraps a map containing only native Go types
// and returns an ImmutableMap which grants read-only access to it
// and all nested values. The map should not be modified after
// being passed into this function. No operations on the ImmutableMap
// modify the original map.
//
// This has O(1) time complexity.
func NewImmutableMap(m map[string]any) *ImmutableMap {
	return &ImmutableMap{base: m, len: len(m)}
}

// NewImmutableSlice wraps a slice containing only native Go types
// and returns an ImmutableSlice which grants read-only access to it
// and all nested values. The slice should not be modified after
// being passed into this function. No operations on the ImmutableSlice
// modify the original slice.
//
// This has O(1) time complexity.
func NewImmutableSlice(s []any) *ImmutableSlice {
	return &ImmutableSlice{base: s, len: len(s)}
}

// ExportImmutableValue converts an ImmutableValue into its native Go type.
// For ImmutableMap and ImmutableSlice types, this performs a deep copy
// of the entire structure.
//
// This has O(n) time complexity, where n is the total number of nodes
// in the graph representing the underlying value.
func ExportImmutableValue(iv ImmutableValue) any {
	switch v := iv.(type) {
	case *ImmutableMap:
		return v.Export()
	case *ImmutableSlice:
		return v.Export()
	default:
		return v
	}
}

// EqualImmutableValues compares two ImmutableValues for deep equality.
// It returns true if they are deeply equal, false otherwise.
// It optimizes for the cases where both values (or nested values) are the
// same ImmutableMap or ImmutableSlice instance, checked by pointer equality.
// Even if two values are different instances, they are still considered
// equal if their contents are deeply equal.
//
// This has O(n) time complexity, where n is the total number of nodes among
// both a and b in the graph representing their underlying values. In practice,
// this might be much faster due to pointer equality optimizations.
func EqualImmutableValues(a, b ImmutableValue) bool {
	switch va := a.(type) {
	case *ImmutableMap:
		vb, ok := b.(*ImmutableMap)
		if !ok {
			return false
		}
		if va == vb {
			// shortcut: same pointer, must be identical
			return true
		}
		if va.Len() != vb.Len() {
			return false
		}
		for k, vaValue := range va.All() {
			vbValue, ok := vb.Get(k)
			if !ok {
				return false
			}
			if !EqualImmutableValues(vaValue, vbValue) {
				return false
			}
		}
		return true
	case *ImmutableSlice:
		vb, ok := b.(*ImmutableSlice)
		if !ok {
			return false
		}
		if va == vb {
			// shortcut: same pointer, must be identical
			return true
		}
		if va.Len() != vb.Len() {
			return false
		}
		for i, vaValue := range va.All() {
			vbValue := vb.At(i)
			if !EqualImmutableValues(vaValue, vbValue) {
				return false
			}
		}
		return true
	default:
		return a == b
	}
}

// Get retrieves an ImmutableValue for the value associated with the
// given key in the ImmutableMap and a boolean indicating whether a
// value for that key exists. If the ImmutableMap is nil, this always
// returns (nil, false).
//
// This has O(1) average time complexity. The first time any container
// object is fetched (and we have to convert it to an Immutable object),
// this function performs O(k) work to shallow copy the underlying map,
// where k is the number of key-value pairs in that map (importantly, nested
// objects are not copied). This cost is amortized over future calls to Get,
// All, SubSlice, and Export.
func (m *ImmutableMap) Get(key string) (ImmutableValue, bool) {
	if m == nil {
		return nil, false
	}

	v, ok := m.base[key]
	if !ok {
		return nil, false
	}
	return m.handleBaseValue(key, v), true
}

// Has returns whether the ImmutableMap contains a value for the given key.
// If the ImmutableMap is nil, this always returns false.
//
// This has O(1) time complexity. It never triggers one-time shallow copies.
func (m *ImmutableMap) Has(key string) bool {
	if m == nil {
		return false
	}

	_, ok := m.base[key]
	return ok
}

// Len returns the number of fields in the ImmutableMap.
// If the ImmutableMap is nil, it returns 0.
//
// This has O(1) time complexity.
func (m *ImmutableMap) Len() int {
	if m == nil {
		return 0
	}

	return m.len
}

// All iterates over all key, value pairs in the ImmutableMap.
// Like iterating over a native Go map, the order of
// pairs is non-deterministic. This function returns and yields nothing
// if the ImmutableMap is nil.
//
// This has O(k') average time complexity, where k' is the number of key-value
// pairs in the map which get iterated over. Similar to Get, the first time
// this iterates over a container object, a one-time shallow copy is performed
// which takes O(k) time, where k is the total number of key-value pairs in
// that map.
func (m *ImmutableMap) All() iter.Seq2[string, ImmutableValue] {
	return func(yield func(string, ImmutableValue) bool) {
		if m == nil {
			return
		}

		for k, v := range m.base {
			v = m.handleBaseValue(k, v)
			if !yield(k, v) {
				return
			}
		}
	}
}

// Mutable returns a mutable version of the ImmutableSlice.
//
// This has O(1) time complexity.
func (m *ImmutableMap) Mutable() *Map {
	if m == nil {
		return nil
	}

	return &Map{base: m, len: m.Len()}
}

// Export returns a deep copy of the map, with all values converted to
// the native Go types of the underlying values. Modifying this map
// does not affect the ImmutableMap, nor any values used as inputs,
// nor any values returned from future calls to Export. If the ImmutableMap
// is nil, this returns nil.
//
// This has O(n) time complexity, where n is the total number of nodes
// in the graph representing the underlying value.
func (m *ImmutableMap) Export() map[string]any {
	if m == nil {
		return nil
	}

	m2 := make(map[string]any, m.Len())
	for k, v := range m.All() {
		m2[k] = ExportImmutableValue(v)
	}
	return m2
}

// At retrieves the ImmutableValue for the value at the specified index.
// Like a native Go slice, if the index is out of bounds, this function panics.
//
// This has O(1) average time complexity. The first time any container
// object is fetched (and we have to convert it to an Immutable object),
// this function performs O(k) work to shallow copy the underlying slice,
// where k is the number of elements in that slice (importantly, nested
// objects are not copied). This cost is amortized over future calls to At,
// All, and Export.
func (s *ImmutableSlice) At(index int) ImmutableValue {
	if index < 0 {
		panic(fmt.Sprintf("*green.ImmutableSlice.At: index out of range [%d]", index))
	}
	if index >= s.Len() {
		panic(fmt.Sprintf("*green.ImmutableSlice.At: index out of range [%d] with length %d", index, s.Len()))
	}

	return s.handleBaseValue(index)
}

// Len returns the number of elements in the ImmutableSlice.
// If the ImmutableSlice is nil, it returns 0.
//
// This has O(1) time complexity.
func (s *ImmutableSlice) Len() int {
	if s == nil {
		return 0
	}

	return s.len
}

// SubSlice returns a new ImmutableSlice that is a subslice of the
// receiver ImmutableSlice, equivalent to calling `mySlice[l:r]`.
// Like a native Go slice, if an index is out of bounds, this function
// panics.
//
// This has O(l-r) average time complexity. Similar to At, the first time
// this slices over a container object, a one-time shallow copy is performed
// which takes O(k) time, where k is the total number of key-value pairs in that map.
func (s *ImmutableSlice) SubSlice(l, r int) *ImmutableSlice {
	if l < 0 {
		panic(fmt.Sprintf("*green.ImmutableSlice.SubSlice: index out of range [%d]", l))
	}
	if r > s.Len() {
		panic(fmt.Sprintf("*green.ImmutableSlice.SubSlice: index out of range [%d] with length %d", r, s.Len()))
	}
	if l > r {
		panic(fmt.Sprintf("*green.ImmutableSlice.SubSlice: slice bounds out of range [%d:%d]", l, r))
	}

	if l == 0 && r == s.Len() {
		return s
	}

	subBase := make([]any, r-l)
	for i := 0; i < r-l; i++ {
		subBase[i] = s.handleBaseValue(l + i)
	}

	return &ImmutableSlice{base: subBase, len: r - l}
}

// All iterates over all elements in the ImmutableSlice in order.
// This has O(k) time complexity, where k is the number of elements in the slice
// which were iterated over.
//
// This has O(k') average time complexity, where k' is the number of elements
// in the slice which get iterated over. Similar to At, the first time this
// iterates over a container object, a one-time shallow copy is performed
// which takes O(k) time, where k is the total number of elements in that slice.
func (s *ImmutableSlice) All() iter.Seq2[int, ImmutableValue] {
	return func(yield func(int, ImmutableValue) bool) {
		if s == nil {
			return
		}

		for i := range s.base {
			v := s.handleBaseValue(i)
			if !yield(i, v) {
				return
			}
		}
	}
}

// Mutable returns a mutable version of the ImmutableSlice.
// This has O(1) time complexity.
func (s *ImmutableSlice) Mutable() *Slice {
	if s == nil {
		return nil
	}

	return &Slice{base: s}
}

// Export returns a deep copy of the slice, with all values converted to
// the native Go types of the underlying values. Modifying this slice
// does not affect the ImmutableSlice, nor any values used as inputs,
// nor any values returned from future calls to Export. If the ImmutableSlice
// is nil, this returns nil.
//
// This has O(n) time complexity, where n is the total number of nodes
// in the graph representing the underlying value.
func (s *ImmutableSlice) Export() []any {
	if s == nil {
		return nil
	}

	s2 := make([]any, s.Len())
	for i, v := range s.All() {
		s2[i] = ExportImmutableValue(v)
	}
	return s2
}

func (m *ImmutableMap) copyBaseOnce() {
	if m.copied {
		return
	}

	m.base = maps.Clone(m.base)
	m.copied = true
}

func (m *ImmutableMap) handleBaseValue(k string, v any) ImmutableValue {
	switch vv := v.(type) {
	case map[string]any:
		iv := NewImmutableMap(vv)
		m.copyBaseOnce()
		m.base[k] = iv
		return iv
	case []any:
		is := NewImmutableSlice(vv)
		m.copyBaseOnce()
		m.base[k] = is
		return is
	default:
		return vv
	}
}

func (s *ImmutableSlice) copyBaseOnce() {
	if s.copied {
		return
	}

	s.base = slices.Clone(s.base)
	s.copied = true
}

func (s *ImmutableSlice) handleBaseValue(index int) ImmutableValue {
	switch vv := s.base[index].(type) {
	case map[string]any:
		iv := NewImmutableMap(vv)
		s.copyBaseOnce()
		s.base[index] = iv
		return iv
	case []any:
		is := NewImmutableSlice(vv)
		s.copyBaseOnce()
		s.base[index] = is
		return is
	default:
		return vv
	}
}
