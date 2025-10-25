package green

// Equal is a green-container-aware equality check between two values. A
// container is equivalent to another container if the outputs of their Export
// functions are deeply equal. A container is equivalent to a native Go type if
// the output of its Export function is deeply equal to that native Go type.
// Consequently, an ImmutableMap can be found to be equal to another
// ImmutableMap, a Map, or a map[string]any, given that they represent the same
// data. This function leverages pointer equality to avoid checking all nodes in
// the graph of the containers can it can be avoided. A mutable container can be
// efficiently found equal to its deriving immutable container if it has not
// been modified since derivation!
//
// This function has O(n) time complexity in the worst case, where n is the
// number of nodes in the container graphs. In the best case, where pointer
// equality can be used to short-circuit the check, it has O(1) time complexity.
func Equal(a, b any) bool {
	switch a := a.(type) {
	case *ImmutableMap:
		return a.Equal(b)
	case *ImmutableSlice:
		return a.Equal(b)
	case *Map:
		return a.Equal(b)
	case *Slice:
		return a.Equal(b)
	default:
		return a == b
	}
}

func equalImmuteMap(a *ImmutableMap, b any) bool {
	switch b := b.(type) {
	case *ImmutableMap:
		return equalImmuteMapToImmuteMap(a, b)
	case *Map:
		return equalImmuteMapToMap(a, b)
	case map[string]any:
		return equalImmuteMapToGoMap(a, b)
	default:
		return false
	}
}

func equalImmuteMapToImmuteMap(a, b *ImmutableMap) bool {
	if a == b {
		return true
	}
	if a.Len() != b.Len() {
		return false
	}
	for k, aValue := range a.All() {
		bValue, ok := b.Get(k)
		if !ok {
			return false
		}
		if !Equal(aValue, bValue) {
			return false
		}
	}
	return true
}

func equalImmuteMapToMap(a *ImmutableMap, b *Map) bool {
	if !b.dirty && b.base == a {
		return true
	}
	if a.Len() != b.Len() {
		return false
	}
	for k, aValue := range a.All() {
		bValue, ok := b.Get(k)
		if !ok {
			return false
		}
		if !Equal(aValue, bValue) {
			return false
		}
	}
	return true
}

func equalImmuteMapToGoMap(a *ImmutableMap, b map[string]any) bool {
	if a.Len() != len(b) {
		return false
	}
	for k, aValue := range a.All() {
		bValue, ok := b[k]
		if !ok {
			return false
		}
		if !Equal(aValue, bValue) {
			return false
		}
	}
	return true
}

func equalImmuteSlice(a *ImmutableSlice, b any) bool {
	switch b := b.(type) {
	case *ImmutableSlice:
		return equalImmuteSliceToImmuteSlice(a, b)
	case *Slice:
		return equalImmuteSliceToSlice(a, b)
	case []any:
		return equalImmuteSliceToGoSlice(a, b)
	default:
		return false
	}
}

func equalImmuteSliceToImmuteSlice(a, b *ImmutableSlice) bool {
	if a == b {
		return true
	}
	if a.Len() != b.Len() {
		return false
	}
	for i := range a.Len() {
		if !Equal(a.At(i), b.At(i)) {
			return false
		}
	}
	return true
}

func equalImmuteSliceToSlice(a *ImmutableSlice, b *Slice) bool {
	if !b.dirty && b.base == a {
		return true
	}
	if a.Len() != b.Len() {
		return false
	}
	for i := range a.Len() {
		if !Equal(a.At(i), b.At(i)) {
			return false
		}
	}
	return true
}

func equalImmuteSliceToGoSlice(a *ImmutableSlice, b []any) bool {
	if a.Len() != len(b) {
		return false
	}
	for i := range a.Len() {
		if !Equal(a.At(i), b[i]) {
			return false
		}
	}
	return true
}

func equalMap(a *Map, b any) bool {
	switch b := b.(type) {
	case *ImmutableMap:
		return equalImmuteMapToMap(b, a)
	case *Map:
		return equalMapToMap(a, b)
	case map[string]any:
		return equalMapToGoMap(a, b)
	default:
		return false
	}
}

func equalMapToMap(a, b *Map) bool {
	if a == b {
		return true
	}
	if a.Len() != b.Len() {
		return false
	}
	for k, aValue := range a.All() {
		bValue, ok := b.Get(k)
		if !ok {
			return false
		}
		if !Equal(aValue, bValue) {
			return false
		}
	}
	return true
}

func equalMapToGoMap(a *Map, b map[string]any) bool {
	if a.Len() != len(b) {
		return false
	}
	for k, aValue := range a.All() {
		bValue, ok := b[k]
		if !ok {
			return false
		}
		if !Equal(aValue, bValue) {
			return false
		}
	}
	return true
}

func equalSlice(a *Slice, b any) bool {
	switch b := b.(type) {
	case *ImmutableSlice:
		return equalImmuteSliceToSlice(b, a)
	case *Slice:
		return equalSliceToSlice(a, b)
	case []any:
		return equalSliceToGoSlice(a, b)
	default:
		return false
	}
}

func equalSliceToSlice(a, b *Slice) bool {
	if a == b {
		return true
	}
	if a.Len() != b.Len() {
		return false
	}
	for i := range a.Len() {
		if !Equal(a.At(i), b.At(i)) {
			return false
		}
	}
	return true
}

func equalSliceToGoSlice(a *Slice, b []any) bool {
	if a.Len() != len(b) {
		return false
	}
	for i := range a.Len() {
		if !Equal(a.At(i), b[i]) {
			return false
		}
	}
	return true
}
