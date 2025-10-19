package lzval

type (
	Map[T any] struct {
		base     *ImmutableMap[T]
		m        map[string]MutableValue
		parent   reportable
		dirty    bool
		hydrated bool
	}

	Slice[T any] struct {
		base     *ImmutableSlice[T]
		a        []MutableValue
		parent   reportable
		dirty    bool
		hydrated bool
	}

	reportable interface {
		reportDirty()
	}
)

func (m *Map[T]) Get(key string) *MutableValue {
	m.hydrate()
	return nil
}

func (m *Map[T]) hydrate() {
	if m.hydrated {
		return
	}
	m.hydrated = true
	if m.m == nil {
		return
	}

	m.m = make(map[string]MutableValue, m.base.Len())
	for k, v := range m.base.All() {
		vm := MutableValue{
			val:    v,
			parent: m,
		}
		m.m[k] = vm
	}

	m.hydrated = true
}

func (m *Map[T]) reportWritten() {
	if !m.dirty {
		m.dirty = true
		if m.parent != nil {
			m.parent.reportDirty()
		}
	}
}

func (mm *Slice) reportWritten() {
	if !mm.dirty {
		mm.dirty = true
		if mm.parent != nil {
			mm.parent.reportDirty()
		}
	}
}
