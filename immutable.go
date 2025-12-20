package green

import (
	"encoding/json"
	"fmt"
	"iter"
	"sync"
)

type (
	// ImmutableValue has a concrete type is either ImmutableMap,
	// ImmutableSlice, or a literal Go type, in contrast to Value, which
	// contains mutable Map and Slice types.
	ImmutableValue any

	// ImmutableMap provides an immutable map of key-value pairs.
	//
	// ImmutableMap methods are safe for concurrent use.
	ImmutableMap struct {
		base          map[string]any
		subContainers map[string]ImmutableValue
		mu            sync.Mutex
		jsonBytes     []byte
		jsonError     error
		jsonMarshal   sync.Once
	}

	// ImmutableSlice provides a slice of values.
	//
	// ImmutableSlice methods are safe for concurrent use.
	ImmutableSlice struct {
		base          []any
		subContainers map[int]ImmutableValue
		mu            sync.Mutex
		jsonBytes     []byte
		jsonError     error
		jsonMarshal   sync.Once
	}
)

// NewImmutableMap wraps a map containing only native Go types and returns an
// ImmutableMap which grants read-only access to it and all nested values. The
// map should not be modified after being passed into this function. No
// operations on the ImmutableMap modify the original map.
//
// This has O(1) time complexity.
func NewImmutableMap(m map[string]any) *ImmutableMap {
	return &ImmutableMap{base: m}
}

// NewImmutableSlice wraps a slice containing only native Go types and returns
// an ImmutableSlice which grants read-only access to it and all nested values.
// The slice should not be modified after being passed into this function. No
// operations on the ImmutableSlice modify the original slice.
//
// This has O(1) time complexity.
func NewImmutableSlice(s []any) *ImmutableSlice {
	return &ImmutableSlice{base: s}
}

// ExportImmutableValue converts an ImmutableValue into its native Go type. For
// ImmutableMap and ImmutableSlice types, this performs a deep copy of the
// entire structure.
//
// This has O(n) time complexity, where n is the total number of nodes in the
// graph representing the underlying value.
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

// Get retrieves an ImmutableValue for the value associated with the given key
// in the ImmutableMap and a boolean indicating whether a value for that key
// exists. If the ImmutableMap is nil, this always returns (nil, false).
//
// If Get would find a container value (immutable, mutable, or native Go), it
// wraps that value in an immutable container before returning it. Thus, the
// only containers this function can return are *green.ImmutableMap and
// *green.ImmutableSlice.
//
// This has O(1) average time complexity.
func (m *ImmutableMap) Get(key string) (ImmutableValue, bool) {
	if m == nil {
		return nil, false
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if v, ok := m.subContainers[key]; ok {
		return v, true
	}

	vBase, ok := m.base[key]
	if !ok {
		return nil, false
	}

	v, ok := isContainer(vBase)
	if ok {
		if m.subContainers == nil {
			m.subContainers = make(map[string]ImmutableValue)
		}
		m.subContainers[key] = v
	}

	return v, true
}

// Has returns whether the ImmutableMap contains a value for the given key. If
// the ImmutableMap is nil, this always returns false.
//
// This has O(1) time complexity.
func (m *ImmutableMap) Has(key string) bool {
	if m == nil {
		return false
	}

	_, ok := m.base[key]
	return ok
}

// Len returns the number of fields in the ImmutableMap. If the ImmutableMap is
// nil, it returns 0.
//
// This has O(1) time complexity.
func (m *ImmutableMap) Len() int {
	if m == nil {
		return 0
	}

	return len(m.base)
}

// All returns an iterator over all key, value pairs in the ImmutableMap. Like
// iterating over a native Go map, the order of pairs is non-deterministic. This
// function yields nothing if the ImmutableMap is nil. See the Get function for
// details on the types of values yielded.
//
// This has O(k') average time complexity, where k' is the number of key-value
// pairs in the map which get iterated over.
func (m *ImmutableMap) All() iter.Seq2[string, ImmutableValue] {
	return func(yield func(string, ImmutableValue) bool) {
		if m == nil {
			return
		}

		for k := range m.base {
			v, _ := m.Get(k)
			if !yield(k, v) {
				return
			}
		}
	}
}

// Mutable derives a mutable version of the ImmutableMap. Subsequent mutations
// to the returned Map do not affect the ImmutableMap. If the ImmutableMap is
// nil, this returns nil.
//
// This has O(1) time complexity.
func (m *ImmutableMap) Mutable() *Map {
	if m == nil {
		return nil
	}

	return &Map{base: m, len: m.Len()}
}

// Export returns a deep copy of the map, with all values converted to the
// native Go types of the underlying values. Modifying this map does not affect
// the ImmutableMap, nor any values used as inputs, nor any values returned from
// future calls to Export. If the ImmutableMap is nil, this returns nil.
//
// This has O(n) time complexity, where n is the total number of nodes in the
// graph representing the underlying value.
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

func (m *ImmutableMap) Equal(other any) bool {
	return equalImmuteMap(m, other)
}

func (m *ImmutableMap) MarshalJSON() ([]byte, error) {
	m.jsonMarshal.Do(func() {
		tmpMap := make(map[string]any, m.Len())
		for k, v := range m.All() {
			tmpMap[k] = v
		}
		m.jsonBytes, m.jsonError = json.Marshal(tmpMap)
	})
	return m.jsonBytes, m.jsonError
}

// At retrieves the ImmutableValue at the specified index. Like a native Go
// slice, if the index is out of bounds, this function panics.
//
// If At would find a container value (immutable, mutable, or native Go), it
// wraps that value in an immutable container before returning it. Thus, the
// only containers this function can return are *green.ImmutableMap and
// *green.ImmutableSlice.
//
// This has O(1) average time complexity.
func (s *ImmutableSlice) At(index int) ImmutableValue {
	if index < 0 {
		panic(fmt.Sprintf("*green.ImmutableSlice.At: index out of range [%d]", index))
	}
	if index >= s.Len() {
		panic(fmt.Sprintf("*green.ImmutableSlice.At: index out of range [%d] with length %d", index, s.Len()))
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if v, ok := s.subContainers[index]; ok {
		return v
	}

	vBase := s.base[index]
	v, ok := isContainer(vBase)
	if ok {
		if s.subContainers == nil {
			s.subContainers = make(map[int]ImmutableValue)
		}
		s.subContainers[index] = v
	}

	return v
}

// Len returns the number of elements in the ImmutableSlice. If the
// ImmutableSlice is nil, it returns 0.
//
// This has O(1) time complexity.
func (s *ImmutableSlice) Len() int {
	if s == nil {
		return 0
	}

	return len(s.base)
}

// SubSlice returns a new ImmutableSlice representing the subslice of the
// original ImmutableSlice from the given left index (inclusive) to the right
// index (exclusive). Like a native Go slice, if the indexes are out of bounds,
// or if left > right, this function panics.
//
// This has O(left-right) average time complexity.
func (s *ImmutableSlice) SubSlice(left, right int) *ImmutableSlice {
	if left < 0 {
		panic(fmt.Sprintf("*green.ImmutableSlice.SubSlice: index out of range [%d]", left))
	}
	if right > s.Len() {
		panic(fmt.Sprintf("*green.ImmutableSlice.SubSlice: index out of range [%d] with length %d", right, s.Len()))
	}
	if left > right {
		panic(fmt.Sprintf("*green.ImmutableSlice.SubSlice: slice bounds out of range [%d:%d]", left, right))
	}

	if left == 0 && right == s.Len() {
		return s
	}

	subBase := make([]any, right-left)
	for i := 0; i < right-left; i++ {
		subBase[i] = s.At(left + i)
	}

	return &ImmutableSlice{base: subBase}
}

// All returns an iterator over all elements in the ImmutableSlice in order.
// This has O(k) time complexity, where k is the number of elements in the slice
// which were iterated over. See the At function for details on the types of
// values yielded.
//
// This has O(k') average time complexity, where k' is the number of elements in
// the slice which get iterated over.
func (s *ImmutableSlice) All() iter.Seq2[int, ImmutableValue] {
	return func(yield func(int, ImmutableValue) bool) {
		if s == nil {
			return
		}

		for i := range s.base {
			v := s.At(i)
			if !yield(i, v) {
				return
			}
		}
	}
}

// Mutable derives a mutable version of the ImmutableSlice. Subsequent mutations
// to the returned Slice do not affect the ImmutableSlice. If the ImmutableSlice
// is nil, this returns nil.
//
// This has O(1) time complexity.
func (s *ImmutableSlice) Mutable() *Slice {
	if s == nil {
		return nil
	}

	return &Slice{base: s}
}

// Export returns a deep copy of the slice, with all values converted to the
// native Go types of the underlying values. Modifying this slice does not
// affect the ImmutableSlice, nor any values used as inputs, nor any values
// returned from future calls to Export. If the ImmutableSlice is nil, this
// returns nil.
//
// This has O(n) time complexity, where n is the total number of nodes in the
// graph representing the underlying value.
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

func (s *ImmutableSlice) Equal(other any) bool {
	return equalImmuteSlice(s, other)
}

func (s *ImmutableSlice) MarshalJSON() ([]byte, error) {
	s.jsonMarshal.Do(func() {
		tmpSlice := make([]any, s.Len())
		for i, v := range s.All() {
			tmpSlice[i] = v
		}
		s.jsonBytes, s.jsonError = json.Marshal(tmpSlice)
	})
	return s.jsonBytes, s.jsonError
}

func isContainer(v any) (ImmutableValue, bool) {
	switch vv := v.(type) {
	case map[string]any:
		iv := NewImmutableMap(vv)
		return iv, true
	case []any:
		is := NewImmutableSlice(vv)
		return is, true
	default:
		return vv, false
	}
}
