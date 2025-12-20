package green

import (
	"encoding/json"
	"fmt"
	"iter"
	"maps"
	"slices"
	"sync"
)

type (
	// Value has a concrete type which is either Map, Slice, or a literal Go
	// type, in contrast to ImmutableValue.
	Value any

	// Map provides a mutable map of key-value pairs.
	//
	// The methods for Map are NOT SAFE for concurrent use. To safely read from
	// a Map concurrently, first convert it to an ImmutableMap via Immutable().
	Map struct {
		base *ImmutableMap
		// overwrites are writes that override values in base.
		// It might contain raw Go containers or green containers.
		// We update to mutable green containers the first time
		// an element is accessed.
		overwrites map[string]any
		parents    []reportable
		// dirty tracks whether this map or a nested container has been mutated
		// since creation.
		dirty bool
		// len is tracked manually as the Map is mutated to provide O(1) Len()
		// calls.
		len int
	}

	// Slice provides a mutable slice of values.
	//
	// The methods for Slice are NOT SAFE for concurrent use. To safely read
	// from a Slice concurrently, first convert it to an ImmutableSlice via
	// Immutable().
	Slice struct {
		base *ImmutableSlice
		// overwrites are writes that overrides values in base.
		overwrites map[int]any
		// overwriteOffset is the offset to apply to overwrite keys to map them
		// to the base slice's indexes. This is used to support ReSlice/SubSlice
		// which may have trimmed elements from the front of base.
		overwriteOffset int
		// appends are elements inserted at the end via Push.
		appends []any
		// prepends are elements inserted at the beginning via PushFront. To
		// optimize for Go slices' native append functionality, this slice
		// contains the front elements in reverse order.
		prepends []any
		parents  []reportable
		// dirty tracks whether this slice or a nested container has been
		// mutated since creation.
		dirty bool
	}
)

// Get retrieves a Value for the value associated with the given key in the Map
// and a boolean indicating whether a value for that key exists. If the Map is
// nil, this always returns (nil, false).
//
// If Get would find a container value (immutable, mutable, or native Go), it
// wraps that value in a mutable container before returning it. Thus, the only
// containers this function can return are *green.Map and *green.Slice.
//
// This has O(1) average time complexity.
func (m *Map) Get(key string) (Value, bool) {
	if m == nil {
		return nil, false
	}

	if v, ok := m.overwrites[key]; ok {
		if isDeleted(v) {
			return nil, false
		}
		v = m.handleNewValue(key, v)
		return v, true
	}

	v, ok := m.base.Get(key)
	if !ok {
		return nil, false
	}

	v = m.handleNewValue(key, v)
	return v, true
}

// Has returns whether the Map contains a value for the given key. If the Map is
// nil, this always returns false.
//
// This has O(1) time complexity.
func (m *Map) Has(key string) bool {
	if m == nil {
		return false
	}

	if v, ok := m.overwrites[key]; ok {
		return !isDeleted(v)
	}

	return m.base.Has(key)
}

// Set sets the value for the given key in the Map. If the Map is nil, this
// panics.
//
// Values passed into the Set function should not be mutated after being set.
// Normal Go values can be passed in, along with immutable and mutable
// containers.
//
// This has O(1) average time complexity.
func (m *Map) Set(key string, val any) {
	if m == nil {
		panic("*green.Map: assignment to entry in nil map")
	}

	keyExisted := m.Has(key)
	m.setOverwrite(key, val)
	if !keyExisted {
		m.len++
	}
	m.reportDirty()
}

// Delete removes the value for the given key in the Map. If the Map is nil,
// this is a no-op.
//
// This has O(1) average time complexity.
func (m *Map) Delete(key string) {
	if m == nil {
		return
	}

	keyExisted := m.Has(key)
	m.setOverwrite(key, deleted)
	if keyExisted {
		m.len--
		m.reportDirty()
	}
}

// Len returns the number of fields in the Map. If the Map is nil, it returns 0.
//
// This has O(1) time complexity.
func (m *Map) Len() int {
	if m == nil {
		return 0
	}

	return m.len
}

// All returns an iterator over all key, value pairs in the Map. Like iterating
// over a native Go map, the order of pairs is non-deterministic. Unlike a
// native Go map, the order of iteration is not from a uniform random
// distribution due to how the underlying data is stored, so do not rely on this
// function for full and secure randomization. This function yields nothing if
// the Map is nil. See the Get function for details on the types of values
// yielded.
//
// This has O(k') average time complexity, where k' is the number of key-value
// pairs in the Map which get iterated over.
func (m *Map) All() iter.Seq2[string, Value] {
	return func(yield func(string, Value) bool) {
		if m == nil {
			return
		}
		for k, v := range m.overwrites {
			if isDeleted(v) {
				continue
			}
			v = m.handleNewValue(k, v)
			if !yield(k, v) {
				return
			}
		}
		for k, v := range m.base.All() {
			if _, overwritten := m.overwrites[k]; overwritten {
				continue
			}
			v = m.handleNewValue(k, v)
			if !yield(k, v) {
				return
			}
		}
	}
}

// Immutable returns an immutable version of the Map. Subsequent mutations to
// the Map do not affect the returned ImmutableMap. If the Map is nil, this
// returns nil.
//
// This has O(k) time complexity, where k is the total number of dirty nodes in
// the graph representing the underlying value.
func (m *Map) Immutable() *ImmutableMap {
	if m == nil {
		return nil
	}
	if !m.dirty {
		return m.base
	}

	im := make(map[string]any, m.Len())
	for k, v := range m.overwrites { // we don't call m.All() because that eagerly wraps as Values
		switch v := v.(type) {
		case *Map:
			im[k] = v.Immutable()
		case *Slice:
			im[k] = v.Immutable()
		case deletedType:
		default:
			im[k] = v
		}
	}
	for k, v := range m.base.All() {
		if _, overwritten := m.overwrites[k]; overwritten {
			continue
		}
		im[k] = v
	}
	return &ImmutableMap{base: im}
}

// Clone returns a shallow copy of the Map. Subsequent mutations to the clone do
// not affect the original Map, and vice versa. However, nested containers are
// shared between the original and the clone, so mutations to nested containers
// affect both. If the Map is nil, this returns nil.
//
// This has O(k+w+d) time complexity, where k is the number of key-value pairs
// in the base map, w is the number of distinct keys Set on the Map, and d is
// the number of distinct keys Deleted from the Map.
func (m *Map) Clone() *Map {
	if m == nil {
		return nil
	}

	for range m.All() {
	}

	m2 := &Map{
		base:       m.base,
		overwrites: maps.Clone(m.overwrites),
		dirty:      m.dirty,
		len:        m.len,
	}

	for _, v := range m.All() {
		switch vv := v.(type) {
		case *Slice:
			vv.parents = append(vv.parents, m2)
		case *Map:
			vv.parents = append(vv.parents, m2)
		}
	}

	return m2
}

// Export returns a deep copy of the Map, with all values converted to the
// native Go types of the underlying values. Modifying this map does not affect
// the Map, nor any values used as inputs, nor any values returned from future
// calls to Export. If the Map is nil, this returns nil.
//
// This has O(n) time complexity, where n is the total number of nodes in the
// graph representing the underlying value.
func (m *Map) Export() map[string]any {
	if m == nil {
		return nil
	}

	m2 := make(map[string]any, m.Len())
	for k, v := range m.All() {
		m2[k] = Export(v)
	}
	return m2
}

func (m *Map) Equal(other any) bool {
	return equalMap(m, other)
}

func (m *Map) MarshalJSON() ([]byte, error) {
	if !m.dirty {
		return m.base.MarshalJSON()
	}
	tmpMap := make(map[string]any, m.Len())
	for k, v := range m.All() {
		tmpMap[k] = v
	}
	return json.Marshal(tmpMap)
}

// At retrieves the Value at the specified index. Like a native Go slice, if the
// index is out of bounds, this function panics.
//
// If At would find a container value (immutable, mutable, or native Go), it
// wraps that value in a mutable container before returning it. Thus, the only
// containers this function can return are *green.Map and *green.Slice.
//
// This has O(1) average time complexity.
func (s *Slice) At(index int) Value {
	if index < 0 {
		panic(fmt.Sprintf("*green.Slice.At: index out of range [%d]", index))
	}
	if index >= s.Len() {
		panic(fmt.Sprintf("*green.Slice.At: index out of range [%d] with length %d", index, s.Len()))
	}

	// in prepends?
	if index < len(s.prepends) {
		prependIndex := s.prependIndex(index)
		v := s.prepends[prependIndex]
		v, ok := asNewMutableContainer(v, s)
		if ok {
			s.prepends[prependIndex] = v
		}
		return v
	}
	index -= len(s.prepends)

	// in base/overwrites?
	if index < s.base.Len() {
		v, ok := s.getOverride(index)
		if ok {
			v, ok := asNewMutableContainer(v, s)
			if ok {
				s.setOverwrite(index, v)
			}
			return v
		}
		v, ok = asNewMutableContainer(s.base.At(index), s)
		if ok {
			s.setOverwrite(index, v)
		}
		return v
	}

	// in appends
	index -= s.base.Len()
	v := s.appends[index]
	v, ok := asNewMutableContainer(v, s)
	if ok {
		s.appends[index] = v
	}
	return v
}

// Set sets the value at the specified index in the Slice. Like a native Go
// slice, if the index is out of bounds, this function panics.
//
// Values passed into the Set function should not be mutated after being set.
// Normal Go values can be passed in, along with immutable and mutable
// containers.
//
// This has O(1) average time complexity.
func (s *Slice) Set(index int, val any) {
	if index < 0 {
		panic(fmt.Sprintf("*green.Slice.Set: index out of range [%d]", index))
	}
	if index >= s.Len() {
		panic(fmt.Sprintf("*green.Slice.Set: index out of range [%d] with length %d", index, s.Len()))
	}

	s.reportDirty()

	// in prepends?
	if index < len(s.prepends) {
		s.prepends[s.prependIndex(index)] = val
		return
	}
	index -= len(s.prepends)

	// in overwrites?
	if index < s.base.Len() {
		s.setOverwrite(index, val)
		return
	}

	// in appends
	index -= s.base.Len()
	s.appends[index] = val
}

// Len returns the number of elements in the Slice. If the Slice is nil, it
// returns 0.
//
// This has O(1) time complexity.
func (s *Slice) Len() int {
	return len(s.prepends) + len(s.appends) + s.base.Len()
}

// Push appends the given value to the end of the Slice. Note that this has
// in-place semantics, it does not return a new Slice like native Go slices'
// append function. If the Slice is nil, this panics.
//
// This has O(1) average time complexity (the same complexity as Go's native
// append function).
func (s *Slice) Push(val any) {
	if s == nil {
		panic("*green.Slice.Push: push to nil slice")
	}

	s.appends = append(s.appends, val)
	s.reportDirty()
}

// PushFront prepends the given value to the front of the Slice. Note that this
// has in-place semantics; it does not return a new Slice. If the Slice is nil,
// this panics.
//
// This has O(1) average time complexity (the same complexity as Go's native
// append function).
func (s *Slice) PushFront(val any) {
	if s == nil {
		panic("*green.Slice.PushFront: push-front to nil slice")
	}

	s.prepends = append(s.prepends, val)
	s.reportDirty()
}

// ReSlice adjusts the bounds of the Slice to the given left index (inclusive)
// and right index (exclusive). Note that this has in-slice semantics; it does
// not return a new Slice. Like a native Go slice, if the indexes are out of
// bounds, or if left > right, this function panics. No values are copied, so
// existing references to containers within the before the ReSlice call are
// still shared.
//
// This has O(left-right) average time complexity.
func (s *Slice) ReSlice(left, right int) {
	s2 := s.subSlice(left, right, "ReSlice")
	if s2 == s {
		return
	}
	*s = *s2
	s.reportDirty()
}

// SubSlice returns a new Slice representing the shallow subslice of the
// original Slice from the given left index (inclusive) to the right index
// (exclusive). Like a native Go slice, if the indexes are out of bounds, or if
// left > right, this function panics. No values are copied, so existing
// references to containers within the Slice are still shared with the SubSlice.
//
// This has O(left-right) average time complexity.
func (s *Slice) SubSlice(left, right int) *Slice {
	return s.subSlice(left, right, "SubSlice")
}

// All returns an iterator over all index, value pairs in the Slice in order.
// This function yields nothing if the Slice is nil. See the At function for
// details on the types of values yielded.
//
// This has O(k') average time complexity, where k' is the number of elements in
// the Slice which get iterated over.
func (s *Slice) All() iter.Seq2[int, Value] {
	return func(yield func(int, Value) bool) {
		if s == nil {
			return
		}
		for i := s.prependIndex(0); i >= 0; i-- {
			v := s.prepends[i]
			v, ok := asNewMutableContainer(v, s)
			if ok {
				s.prepends[i] = v
			}
			if !yield(s.prependIndex(i), v) {
				return
			}
		}
		for i, v := range s.base.All() {
			if v2, ok := s.getOverride(i); ok {
				v = v2
			}
			v, ok := asNewMutableContainer(v, s)
			if ok {
				s.setOverwrite(i, v)
			}
			if !yield(i+len(s.prepends), v) {
				return
			}
		}
		for i, v := range s.appends {
			v, ok := asNewMutableContainer(v, s)
			if ok {
				s.appends[i] = v
			}
			if !yield(i+len(s.prepends)+s.base.Len(), v) {
				return
			}
		}
	}
}

// Immutable returns an immutable version of the Slice. Subsequent mutations to
// the Slice do not affect the returned ImmutableSlice. If the Slice is nil,
// this returns nil.
//
// This has O(k) time complexity, where k is the total number of dirty nodes in
// the graph representing the underlying value.
func (s *Slice) Immutable() *ImmutableSlice {
	if s == nil {
		return nil
	}
	if !s.dirty {
		return s.base
	}

	is := make([]any, s.Len())
	j := 0
	addElement := func(v any) {
		switch v := v.(type) {
		case *Map:
			is[j] = v.Immutable()
		case *Slice:
			is[j] = v.Immutable()
		default:
			is[j] = v
		}
		j++
	}

	// we don't call m.All() because that eagerly wraps as Values
	for i := s.prependIndex(0); i >= 0; i-- {
		v := s.prepends[i]
		addElement(v)
	}
	for i, v := range s.base.All() {
		if v2, ok := s.getOverride(i); ok {
			v = v2
		}
		addElement(v)
	}
	for _, v := range s.appends {
		addElement(v)
	}

	return &ImmutableSlice{base: is}
}

// Clone returns a shallow copy of the Slice. Subsequent mutations to the clone
// do not affect the original Slice, and vice versa. However, nested containers
// are shared between the original and the clone, so mutations to nested
// containers affect both. If the Slice is nil, this returns nil.
//
// This has O(k) time complexity, where k is the number of elements in the
// Slice.
func (s *Slice) Clone() *Slice {
	if s == nil {
		return nil
	}

	for range s.All() {
	}

	s2 := &Slice{
		base:            s.base,
		overwrites:      maps.Clone(s.overwrites),
		overwriteOffset: s.overwriteOffset,
		appends:         slices.Clone(s.appends),
		prepends:        slices.Clone(s.prepends),
		dirty:           s.dirty,
	}

	for _, v := range s.All() {
		switch vv := v.(type) {
		case *Slice:
			vv.parents = append(vv.parents, s2)
		case *Map:
			vv.parents = append(vv.parents, s2)
		}
	}

	return s2
}

// Export returns a deep copy of the slice, with all values converted to the
// native Go types of the underlying values. Modifying this slice does not
// affect the Slice, nor any values used as inputs, nor any values returned from
// future calls to Export. If the Slice is nil, this returns nil.
//
// This has O(n) time complexity, where n is the total number of nodes in the
// graph representing the underlying value.
func (s *Slice) Export() []any {
	if s == nil {
		return nil
	}

	s2 := make([]any, s.Len())
	for i, v := range s.All() {
		s2[i] = Export(v)
	}
	return s2
}

func (s *Slice) Equal(other any) bool {
	return equalSlice(s, other)
}

func (s *Slice) MarshalJSON() ([]byte, error) {
	if !s.dirty {
		return s.base.MarshalJSON()
	}
	tmpSlice := make([]any, s.Len())
	for i, v := range s.All() {
		tmpSlice[i] = v
	}
	return json.Marshal(tmpSlice)
}

type (
	// reportable is an interface which mutable containers implement. It is used
	// to signal dirty state to parent containers.
	reportable interface {
		reportDirty()
	}
)

func asNewMutableContainer(v any, parent reportable) (any, bool) {
	switch v := v.(type) {
	case *ImmutableMap:
		v2 := v.Mutable()
		v2.parents = []reportable{parent}
		return v2, true
	case *ImmutableSlice:
		v2 := v.Mutable()
		v2.parents = []reportable{parent}
		return v2, true
	case map[string]any:
		v2 := NewImmutableMap(v).Mutable()
		v2.parents = []reportable{parent}
		return v2, true
	case []any:
		v2 := NewImmutableSlice(v).Mutable()
		v2.parents = []reportable{parent}
		return v2, true
	default:
		return v, false
	}
}

func (m *Map) handleNewValue(k string, v any) any {
	v, ok := asNewMutableContainer(v, m)
	if ok {
		m.setOverwrite(k, v)
	}
	return v
}

func (m *Map) reportDirty() {
	if !m.dirty {
		m.dirty = true
		var wg sync.WaitGroup
		for _, p := range m.parents {
			wg.Add(1)
			go func(p reportable) {
				defer wg.Done()
				p.reportDirty()
			}(p)
		}
		wg.Wait()
	}
}

func (m *Map) setOverwrite(k string, v any) {
	if m.overwrites == nil {
		m.overwrites = make(map[string]any)
	}
	m.overwrites[k] = v
}

func (s *Slice) reportDirty() {
	if !s.dirty {
		s.dirty = true
		var wg sync.WaitGroup
		for _, p := range s.parents {
			wg.Add(1)
			go func(p reportable) {
				defer wg.Done()
				p.reportDirty()
			}(p)
		}
		wg.Wait()
	}
}

func (s *Slice) setOverwrite(i int, v any) {
	if s.overwrites == nil {
		s.overwrites = make(map[int]any)
	}
	s.overwrites[i+s.overwriteOffset] = v
}

func (s *Slice) prependIndex(i int) int {
	return len(s.prepends) - 1 - i
}

func (s *Slice) subSlice(l, r int, funcName string) *Slice {
	if l < 0 {
		panic(fmt.Sprintf("*green.Slice.%s: index out of range [%d]", funcName, l))
	}
	if r > s.Len() {
		panic(fmt.Sprintf("*green.Slice.%s: index out of range [%d] with length %d", funcName, r, s.Len()))
	}
	if l > r {
		panic(fmt.Sprintf("*green.Slice.%s: slice bounds out of range [%d:%d]", funcName, l, r))
	}

	if l == 0 && r == s.Len()-1 {
		return s
	}

	var (
		newPrepends        []any
		newBase            *ImmutableSlice
		newOverwriteOffset int
		newAppends         []any
	)

	// trim from the front
	lToTrim := l
	if lToTrim > 0 {
		origPrependsLen := len(s.prepends)
		newPrepends = s.prepends[:max(0, s.prependIndex(lToTrim)+1)]
		lToTrim -= origPrependsLen - len(newPrepends)
	}

	// trim from the back
	rToTrim := s.Len() - r
	if rToTrim > 0 {
		origAppendsLen := len(s.appends)
		newAppends = s.appends[:max(0, len(s.appends)-rToTrim)]
		rToTrim -= origAppendsLen - len(newAppends)
	}

	// adjust base
	if lToTrim > 0 || rToTrim > 0 {
		// adjust overwrites offset
		newOverwriteOffset = s.overwriteOffset + lToTrim

		// resize the base
		lBaseToTrim := min(lToTrim, s.base.Len())
		rBaseToTrim := min(rToTrim, s.base.Len()-lBaseToTrim)
		newBase = s.base.SubSlice(lBaseToTrim, s.base.Len()-rBaseToTrim)

		lToTrim -= lBaseToTrim
		rToTrim -= rBaseToTrim

		// adjust prepends/appends if we overflow from base (only 1 branch can
		// be true)
		if lToTrim > 0 {
			newAppends = newAppends[lToTrim:]
		} else if rToTrim > 0 {
			newPrepends = newPrepends[rToTrim:]
		}
	} else {
		newBase = s.base
	}

	// force wrapping of immediately nested values
	for i, v := range newPrepends {
		v, ok := asNewMutableContainer(v, s)
		if ok {
			newPrepends[i] = v // updates underlying array too
		}
	}
	for i, v := range newAppends {
		v, ok := asNewMutableContainer(v, s)
		if ok {
			newAppends[i] = v // updates underlying array too
		}
	}
	for i, v := range newBase.All() {
		if v2, ok := s.getOverride(i + newOverwriteOffset); ok {
			v = v2
		}
		v, ok := asNewMutableContainer(v, s)
		if ok {
			s.setOverwrite(i+newOverwriteOffset, v)
		}
	}

	return &Slice{
		base:            newBase,
		overwrites:      s.overwrites,
		overwriteOffset: newOverwriteOffset,
		appends:         newAppends,
		prepends:        newPrepends,
		parents:         s.parents,
		dirty:           s.dirty,
	}
}

func (s *Slice) getOverride(i int) (any, bool) {
	v, ok := s.overwrites[i+s.overwriteOffset]
	return v, ok
}

type deletedType struct{}

var deleted = deletedType{}

func isDeleted(v any) bool {
	_, ok := v.(deletedType)
	return ok
}
