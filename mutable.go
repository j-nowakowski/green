package green

import (
	"fmt"
	"iter"
	"maps"
	"slices"
)

type (
	// Value has a concrete type which is either Map, Slice, or a literal Go
	// type. Values are mutable.
	Value any

	// Map represents an mutable map of key-value pairs.
	//
	// The methods of mutableMap are NOT SAFE for concurrent use.
	Map struct {
		base       *ImmutableMap
		overwrites map[string]any
		deletions  map[string]struct{}
		parent     reportable
		// cloneParent tracks the parent of the original copy of the map when
		// the map was created via Clone(). It is used to signal dirty state to
		// the original parent when nested, shared data is mutated, but to avoid
		// making such signals if the mutation is on the clone only.
		cloneParent reportable
		dirty       bool
		len         int
	}

	// Slice represents an mutable slice of values.
	//
	// The methods of mutableSlice are NOT SAFE for concurrent use.
	Slice struct {
		base            *ImmutableSlice
		overwrites      map[int]any
		overwriteOffset int
		appends         []any
		// prepends are inserted in reverse order
		prepends []any
		parent   reportable
		// cloneParent tracks the parent of the original copy of the slice when
		// the slice was created via Clone(). It is used to signal dirty state
		// to the original parent when nested, shared data is mutated, but to
		// avoid making such signals if the mutation is on the clone only.
		cloneParent reportable
		dirty       bool
	}

	reportable interface {
		reportDirty()
	}
)

// ExportValue converts an Value into its native Go type. For Map and Slice
// types, this performs a deep copy of the entire structure.
//
// This has O(n) time complexity, where n is the total number of nodes in the
// graph representing the underlying value.
func ExportValue(v Value) any {
	switch v := v.(type) {
	case *Map:
		return v.Export()
	case *Slice:
		return v.Export()
	default:
		return v
	}
}

// EqualValues compares two Values for deep equality. It returns true if they
// are deeply equal, false otherwise. It optimizes for the cases where both
// values (or nested values) are the same Map or Slice instance, checked by
// pointer equality. Even if two values are different instances, they are still
// considered equal if their contents are deeply equal.
//
// This has O(n) time complexity, where n is the total number of nodes among
// both a and b in the graph representing their underlying values. In practice,
// this might be much faster due to pointer equality optimizations.
func EqualValues(a, b Value) bool {
	switch va := a.(type) {
	case *Map:
		vb, ok := b.(*Map)
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
			if !EqualValues(vaValue, vbValue) {
				return false
			}
		}
		return true
	case *Slice:
		vb, ok := b.(*Slice)
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
			if !EqualValues(vaValue, vbValue) {
				return false
			}
		}
		return true
	default:
		return a == b
	}
}

func (m *Map) Set(key string, val any) {
	if m == nil {
		panic("*green.Map: assignment to entry in nil map")
	}

	keyExisted := m.Has(key)
	m.addOverwrite(key, val)
	m.reportDirty()
	delete(m.deletions, key)
	if !keyExisted {
		m.len++
	}
}

func (m *Map) Has(key string) bool {
	if m == nil {
		return false
	}

	if _, ok := m.overwrites[key]; ok {
		return true
	}
	if _, ok := m.deletions[key]; ok {
		return false
	}
	return m.base.Has(key)
}

func (m *Map) Delete(key string) {
	if m == nil {
		return
	}

	keyExisted := m.Has(key)
	m.addDeletion(key)
	m.reportDirty()
	delete(m.overwrites, key)
	if keyExisted {
		m.len--
	}
}

func (m *Map) Len() int {
	if m == nil {
		return 0
	}

	return m.len
}

func (m *Map) Get(key string) (Value, bool) {
	if m == nil {
		return nil, false
	}

	if v, ok := m.overwrites[key]; ok {
		v = m.handleBaseValue(key, v)
		return v, true
	}

	v, ok := m.base.Get(key)
	if !ok {
		return nil, false
	}

	if _, deleted := m.deletions[key]; deleted {
		return nil, false
	}

	v = m.handleBaseValue(key, v)
	return v, true
}

func (m *Map) All() iter.Seq2[string, Value] {
	return func(yield func(string, Value) bool) {
		if m == nil {
			return
		}
		for k, v := range m.overwrites {
			v = m.handleBaseValue(k, v)
			if !yield(k, v) {
				return
			}
		}
		for k, v := range m.base.All() {
			if _, deleted := m.deletions[k]; deleted {
				continue
			}
			if _, overwritten := m.overwrites[k]; overwritten {
				continue
			}
			v = m.handleBaseValue(k, v)
			if !yield(k, v) {
				return
			}
		}
	}
}

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
		default:
			im[k] = v
		}
	}
	for k, v := range m.base.All() {
		if _, deleted := m.deletions[k]; deleted {
			continue
		}
		if _, overwritten := m.overwrites[k]; overwritten {
			continue
		}
		im[k] = v
	}
	return &ImmutableMap{base: im, len: len(im), copied: true}
}

func (m *Map) Clone() *Map {
	if m == nil {
		return nil
	}

	for range m.All() {
		// force wrapping of immediately nested values
	}

	return &Map{
		base:        m.base,
		overwrites:  maps.Clone(m.overwrites),
		deletions:   maps.Clone(m.deletions),
		cloneParent: m.cloneParent,
		dirty:       m.dirty,
		len:         m.len,
	}
}

// Export returns a deep copy of the map, with all values converted to the
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
		m2[k] = ExportValue(v)
	}
	return m2
}

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
		s.addOverwrite(index, val)
		return
	}

	// in appends
	index -= s.base.Len()
	s.appends[index] = val
}

func (s *Slice) Len() int {
	return len(s.prepends) + len(s.appends) + s.base.Len()
}

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
		v, ok := handleBaseValue(v, s)
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
			v, ok := handleBaseValue(v, s)
			if ok {
				s.addOverwrite(index, v)
			}
			return v
		}
		v, ok = handleBaseValue(s.base.At(index), s)
		if ok {
			s.addOverwrite(index, v)
		}
		return v
	}

	// in appends
	index -= s.base.Len()
	v := s.appends[index]
	v, ok := handleBaseValue(v, s)
	if ok {
		s.appends[index] = v
	}
	return v
}

func (s *Slice) All() iter.Seq2[int, Value] {
	return func(yield func(int, Value) bool) {
		if s == nil {
			return
		}
		for i := s.prependIndex(0); i >= 0; i-- {
			v := s.prepends[i]
			v, ok := handleBaseValue(v, s)
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
			v, ok := handleBaseValue(v, s)
			if ok {
				s.addOverwrite(i, v)
			}
			if !yield(i+len(s.prepends), v) {
				return
			}
		}
		for i, v := range s.appends {
			v, ok := handleBaseValue(v, s)
			if ok {
				s.appends[i] = v
			}
			if !yield(i+len(s.prepends)+s.base.Len(), v) {
				return
			}
		}
	}
}

func (s *Slice) Push(val any) {
	s.reportDirty()
	s.appends = append(s.appends, val)
}

func (s *Slice) PushFront(val any) {
	s.reportDirty()
	s.prepends = append(s.prepends, val)
}

func (s *Slice) ReSlice(l, r int) {
	s2 := s.subSlice(l, r, "ReSlice")
	*s = *s2
}

func (s *Slice) SubSlice(l, r int) *Slice {
	return s.subSlice(l, r, "SubSlice")
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
		s2[i] = ExportValue(v)
	}
	return s2
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
		v, ok := handleBaseValue(v, s)
		if ok {
			newPrepends[i] = v // updates underlying array too
		}
	}
	for i, v := range newAppends {
		v, ok := handleBaseValue(v, s)
		if ok {
			newAppends[i] = v // updates underlying array too
		}
	}
	for i, v := range newBase.All() {
		if v2, ok := s.getOverride(i + newOverwriteOffset); ok {
			v = v2
		}
		v, ok := handleBaseValue(v, s)
		if ok {
			s.addOverwrite(i+newOverwriteOffset, v)
		}
	}

	return &Slice{
		base:            newBase,
		overwrites:      s.overwrites,
		overwriteOffset: newOverwriteOffset,
		appends:         newAppends,
		prepends:        newPrepends,
		parent:          s.parent,
		cloneParent:     s.cloneParent,
		dirty:           s.dirty,
	}
}

func (s *Slice) Clone() *Slice {
	if s == nil {
		return nil
	}

	for range s.All() {
		// force wrapping of immediately nested values
	}

	return &Slice{
		base:            s.base,
		overwrites:      maps.Clone(s.overwrites),
		overwriteOffset: s.overwriteOffset,
		appends:         slices.Clone(s.appends),
		prepends:        slices.Clone(s.prepends),
		parent:          s.parent,
		cloneParent:     s.cloneParent,
		dirty:           s.dirty,
	}
}

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

	return &ImmutableSlice{base: is, len: len(is), copied: true}
}

func (s *Slice) getOverride(i int) (any, bool) {
	v, ok := s.overwrites[i+s.overwriteOffset]
	return v, ok
}

func handleBaseValue(v any, parent reportable) (any, bool) {
	switch v := v.(type) {
	case *ImmutableMap:
		v2 := v.Mutable()
		v2.parent = parent
		v2.cloneParent = parent
		return v2, true
	case *ImmutableSlice:
		v2 := v.Mutable()
		v2.parent = parent
		v2.cloneParent = parent
		return v2, true
	case map[string]any:
		v2 := NewImmutableMap(v).Mutable()
		v2.parent = parent
		v2.cloneParent = parent
		return v2, true
	case []any:
		v2 := NewImmutableSlice(v).Mutable()
		v2.parent = parent
		v2.cloneParent = parent
		return v2, true
	default:
		return v, false
	}
}

func (m *Map) handleBaseValue(k string, v any) any {
	v, ok := handleBaseValue(v, m)
	if ok {
		m.addOverwrite(k, v)
	}
	return v
}

func (m *Map) reportDirty() {
	if !m.dirty {
		m.dirty = true
		if m.parent != nil {
			m.parent.reportDirty()
		}
		if m.cloneParent != nil {
			m.cloneParent.reportDirty()
		}
	}
}

func (m *Map) addOverwrite(k string, v any) {
	if m.overwrites == nil {
		m.overwrites = make(map[string]any)
	}
	m.overwrites[k] = v
}

func (m *Map) addDeletion(k string) {
	if m.deletions == nil {
		m.deletions = make(map[string]struct{})
	}
	m.deletions[k] = struct{}{}
}

func (s *Slice) reportDirty() {
	if !s.dirty {
		s.dirty = true
		if s.parent != nil {
			s.parent.reportDirty()
		}
		if s.cloneParent != nil {
			s.cloneParent.reportDirty()
		}
	}
}

func (s *Slice) addOverwrite(i int, v any) {
	if s.overwrites == nil {
		s.overwrites = make(map[int]any)
	}
	s.overwrites[i+s.overwriteOffset] = v
}

func (s *Slice) prependIndex(i int) int {
	return len(s.prepends) - 1 - i
}
