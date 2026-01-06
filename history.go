package green

// import (
// 	"iter"
// )

// type (
// 	historicMap struct {
// 		base         map[string]any
// 		writeHistory []map[string]any
// 		length       int
// 		dirty        bool
// 	}

// 	deleted struct{}
// )

// func (b *historicMap) get(k string) (any, bool) {
// 	for i := len(b.writeHistory) - 1; i >= 0; i-- {
// 		v, ok := b.writeHistory[i][k]
// 		if !ok {
// 			continue
// 		}
// 		if isDeleted(v) {
// 			return nil, false
// 		}
// 		return v, true
// 	}
// 	v, ok := b.base[k]
// 	if !ok || isDeleted(v) {
// 		return nil, false
// 	}

// 	return v, ok
// }

// func (b *historicMap) has(k string) bool {
// 	for i := len(b.writeHistory) - 1; i >= 0; i-- {
// 		v, ok := b.writeHistory[i][k]
// 		if !ok {
// 			continue
// 		}
// 		if isDeleted(v) {
// 			return false
// 		}
// 		return true
// 	}
// 	return false
// }

// func (b *historicMap) len() int {
// 	return b.length
// }

// func (b *historicMap) buildHistory(m map[string]any) historicMap {
// 	maps := make([]map[string]any, len(b.writeHistory)+1)
// 	copy(maps, b.writeHistory)
// 	maps[len(b.writeHistory)] = m
// 	return historicMap{
// 		writeHistory: maps,
// 		length:       b.length + len(m),
// 	}
// }

// func (b *historicMap) all() iter.Seq2[string, any] {
// 	return func(yield func(string, any) bool) {
// 		if b == nil {
// 			return
// 		}

// 		seen := getSet()
// 		defer putSet(seen)

// 		for i := len(b.writeHistory) - 1; i >= 0; i-- {
// 			m := b.writeHistory[i]
// 			for k, v := range m {
// 				if _, ok := seen[k]; ok {
// 					continue
// 				}
// 				seen[k] = struct{}{}
// 				if isDeleted(v) {
// 					continue
// 				}
// 				if !yield(k, v) {
// 					return
// 				}
// 			}
// 		}
// 	}
// }

// func isDeleted(v any) bool {
// 	return v == (deleted{})
// }
