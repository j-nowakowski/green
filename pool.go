package green

// import "sync"

// type (
// 	set struct {
// 		m map[string]struct{}
// 	}
// )

// var (
// 	setPool = sync.Pool{
// 		New: func() any { return &set{} },
// 	}

// 	objPool = sync.Pool{
// 		New: func() any { return make(map[string]any, 0) },
// 	}
// )

// func getSet() *set {
// 	return setPool.Get().(*set)
// }

// func putSet(s *set) {
// 	clear(s.m)
// 	setPool.Put(s)
// }

// func getObj() map[string]any {
// 	return objPool.Get().(map[string]any)
// }

// func putObj(m map[string]any) {
// 	clear(m)
// 	objPool.Put(m)
// }

// func (s *set) add(key string) {
// 	if s.m == nil {
// 		s.m = make(map[string]struct{})
// 	}
// 	s.m[key] = struct{}{}
// }

// func (s *set) has(key string) bool {
// 	_, ok := s.m[key]
// 	return ok
// }
