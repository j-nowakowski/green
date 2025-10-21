package lzval

// type (
// 	Map struct {
// 		base     *ImmutableMap
// 		m        map[string]MutableValueLoader
// 		parent   reportable
// 		dirty    bool
// 		hydrated bool
// 	}

// 	Slice struct {
// 		base     *ImmutableSlice
// 		s        []MutableValueLoader
// 		parent   reportable
// 		dirty    bool
// 		hydrated bool
// 	}

// 	literalLoader struct {
// 		loadable Loadable[any]
// 		parent   reportable
// 		once     sync.Once
// 		mv       Value
// 		err      error
// 	}

// 	immutableLiteralLoader struct {
// 		loadable Loadable[any]
// 		once     sync.Once
// 		mv       ImmutableValue
// 		err      error
// 	}

// 	alwaysLoader struct {
// 		v any
// 	}

// 	reportable interface {
// 		reportDirty()
// 	}

// 	mutableLoader struct {
// 		parent   reportable
// 		loadable Loadable[ImmutableValue]
// 		once     sync.Once
// 		v        Value
// 		err      error
// 	}
// )

// func (s *Map) Get(key string) (MutableValueLoader, bool) {
// 	s.hydrate()

// 	val, ok := s.m[key]
// 	return val, ok
// }

// func (s *Map) Set(key string, v any) {
// 	s.hydrate()

// 	s.m[key] = &literalLoader{
// 		loadable: &alwaysLoader{v: v},
// 		parent:   s,
// 	}
// 	s.reportDirty()
// }

// func (s *Map) SetLoader(key string, v MutableValueLoader) {
// 	s.hydrate()

// 	s.m[key] = &literalLoader{
// 		loadable: v,
// 		parent:   s,
// 	}
// 	s.reportDirty()
// }

// func (s *Map) Len() int {
// 	if !s.hydrated {
// 		return s.base.Len()
// 	} else {
// 		return len(s.m)
// 	}
// }

// func (s *Map) Delete(key string) {
// 	s.hydrate()

// 	delete(s.m, key)
// 	s.reportDirty()
// }

// func (s *Map) All() iter.Seq2[string, ValueLoader] {
// 	s.hydrate()

// 	return func(yield func(string, ValueLoader) bool) {
// 		if s == nil {
// 			return
// 		}
// 		for k, v := range s.m {
// 			if !yield(k, v) {
// 				return
// 			}
// 		}
// 	}
// }

// func (s *Map) Immutable() *ImmutableMap {
// 	if s == nil {
// 		return nil
// 	}
// 	if !s.dirty {
// 		return s.base
// 	}

// 	im := &ImmutableMap{
// 		m: make(map[string]ValueLoader, len(s.m)),
// 	}
// 	for k, v := range s.m {
// 		switch v := v.(type) {
// 		case *alwaysLoader:
// 			im.m[k] = &immutableLiteralLoader{loadable: v}
// 		case *literalLoader:
// 			im.m[k] = &immutableLiteralLoader{loadable: v.loadable}
// 		case *mutableLoader:
// 			switch vv := v.v.(type) {
// 			case *Map:
// 				im.m[k] = &alwaysLoader{v: vv.Immutable()}
// 			case *Slice:
// 				im.m[k] = &alwaysLoader{v: vv.Immutable()}
// 			default:
// 				im.m[k] = v.loadable
// 			}
// 		default:
// 			panic("unreachable")
// 		}
// 	}
// 	return im
// }

// func (s *Slice) At(index int) MutableValueLoader {
// 	s.hydrate()

// 	return s.s[index]
// }

// func (s *Slice) Set(index int, v any) {
// 	s.hydrate()

// 	s.s[index] = &literalLoader{
// 		loadable: &alwaysLoader{v: v},
// 		parent:   s,
// 	}
// 	s.reportDirty()
// }

// func (s *Slice) SetLoader(index int, v MutableValueLoader) {
// 	s.hydrate()

// 	s.s[index] = &literalLoader{
// 		loadable: v,
// 		parent:   s,
// 	}
// 	s.reportDirty()
// }

// func (s *Slice) Append(v any) {
// 	s.hydrate()

// 	s.s = append(s.s, &literalLoader{
// 		loadable: &alwaysLoader{v: v},
// 		parent:   s,
// 	})
// 	s.reportDirty()
// }

// func (s *Slice) AppendLoader(v MutableValueLoader) {
// 	s.hydrate()

// 	s.s = append(s.s, &literalLoader{
// 		loadable: v,
// 		parent:   s,
// 	})
// 	s.reportDirty()
// }

// func (s *Slice) Len() int {
// 	if !s.hydrated {
// 		return s.base.Len()
// 	} else {
// 		return len(s.s)
// 	}
// }

// func (s *Slice) All() iter.Seq2[int, ValueLoader] {
// 	s.hydrate()

// 	return func(yield func(int, ValueLoader) bool) {
// 		if s == nil {
// 			return
// 		}
// 		for i, v := range s.s {
// 			if !yield(i, v) {
// 				return
// 			}
// 		}
// 	}
// }

// func (s *Slice) Immutable() *ImmutableSlice {
// 	if s == nil {
// 		return nil
// 	}
// 	if !s.dirty {
// 		return s.base
// 	}

// 	im := &ImmutableSlice{
// 		s: make([]ValueLoader, len(s.s)),
// 	}
// 	for i, v := range s.s {
// 		switch v := v.(type) {
// 		case *alwaysLoader:
// 			im.s[i] = &immutableLiteralLoader{loadable: v}
// 		case *literalLoader:
// 			im.s[i] = &immutableLiteralLoader{loadable: v.loadable}
// 		case *mutableLoader:
// 			switch vv := v.v.(type) {
// 			case *Map:
// 				im.s[i] = &alwaysLoader{v: vv.Immutable()}
// 			case *Slice:
// 				im.s[i] = &alwaysLoader{v: vv.Immutable()}
// 			default:
// 				im.s[i] = v.loadable
// 			}
// 		default:
// 			panic("unreachable")
// 		}
// 	}
// 	return im
// }

// func (s *Map) hydrate() {
// 	if s == nil {
// 		return
// 	}

// 	if s.hydrated {
// 		return
// 	}
// 	s.hydrated = true

// 	if s.m == nil {
// 		return
// 	}
// 	s.m = make(map[string]MutableValueLoader, s.base.Len())
// 	for k, v := range s.base.m {
// 		s.m[k] = &mutableLoader{
// 			parent:   s,
// 			loadable: v,
// 		}
// 	}
// }

// func (s *Map) reportDirty() {
// 	if !s.dirty {
// 		s.dirty = true
// 		if s.parent != nil {
// 			s.parent.reportDirty()
// 		}
// 	}
// }

// func (s *Slice) hydrate() {
// 	if s == nil {
// 		return
// 	}

// 	if s.hydrated {
// 		return
// 	}
// 	s.hydrated = true

// 	if s.s == nil {
// 		return
// 	}
// 	s.s = make([]MutableValueLoader, s.base.Len())
// 	for i, v := range s.base.s {
// 		s.s[i] = &mutableLoader{
// 			parent:   s,
// 			loadable: v,
// 		}
// 	}
// }

// func (mm *Slice) reportDirty() {
// 	if !mm.dirty {
// 		mm.dirty = true
// 		if mm.parent != nil {
// 			mm.parent.reportDirty()
// 		}
// 	}
// }

// func (ml *mutableLoader) Load(ctx context.Context) (Value, error) {
// 	ml.once.Do(func() {
// 		var v any
// 		v, ml.err = ml.loadable.Load(ctx)
// 		if ml.err != nil {
// 			return
// 		}

// 		switch v := v.(type) {
// 		case *ImmutableMap:
// 			ml.v = &Map{
// 				base:   v,
// 				parent: ml.parent,
// 			}
// 		case *ImmutableSlice:
// 			ml.v = &Slice{
// 				base:   v,
// 				parent: ml.parent,
// 			}
// 		default:
// 			ml.v = v
// 		}
// 	})
// 	return ml.v, ml.err
// }

// func (mll *literalLoader) Load(ctx context.Context) (Value, error) {
// 	mll.once.Do(func() {
// 		var v any
// 		v, mll.err = mll.loadable.Load(ctx)
// 		if mll.err != nil {
// 			return
// 		}
// 		switch v := v.(type) {
// 		case map[string]any:
// 			m := make(map[string]MutableValueLoader, len(v))
// 			mm := &Map{m: m}
// 			mll.mv = mm
// 			for k, val := range v {
// 				m[k] = &literalLoader{
// 					loadable: &alwaysLoader{v: val},
// 					parent:   mm,
// 				}
// 			}
// 		case []any:
// 			s := make([]MutableValueLoader, len(v))
// 			ms := &Slice{s: s}
// 			mll.mv = ms
// 			for i, val := range v {
// 				s[i] = &literalLoader{
// 					loadable: &alwaysLoader{v: val},
// 					parent:   ms,
// 				}
// 			}
// 		case *Map, *Slice:
// 			mll.mv = v
// 		case *ImmutableMap:
// 			mll.mv = v.Mutable()
// 		case *ImmutableSlice:
// 			mll.mv = v.Mutable()
// 		default:
// 			mll.mv = v
// 		}
// 	})
// 	return mll.mv, nil
// }

// func (ll *immutableLiteralLoader) Load(ctx context.Context) (ImmutableValue, error) {
// 	ll.once.Do(func() {
// 		var v any
// 		v, ll.err = ll.loadable.Load(ctx)
// 		if ll.err != nil {
// 			return
// 		}
// 		switch v := v.(type) {
// 		case map[string]any:
// 			m := make(map[string]ValueLoader, len(v))
// 			mm := &Map{m: m}
// 			ll.mv = mm
// 			for k, val := range v {
// 				m[k] = &immutableLiteralLoader{
// 					loadable: &alwaysLoader{v: val},
// 				}
// 			}
// 		case []any:
// 			s := make([]MutableValueLoader, len(v))
// 			ms := &Slice{s: s}
// 			ll.mv = ms
// 			for i, val := range v {
// 				s[i] = &immutableLiteralLoader{
// 					loadable: &alwaysLoader{v: val},
// 				}
// 			}
// 		case *ImmutableMap, *ImmutableSlice:
// 			ll.mv = v
// 		case *Map:
// 			ll.mv = v.Immutable()
// 		case *Slice:
// 			ll.mv = v.Immutable()
// 		default:
// 			ll.mv = v
// 		}
// 	})
// 	return ll.mv, nil
// }

// func (al *alwaysLoader) Load(_ context.Context) (any, error) {
// 	return al.v, nil
// }
