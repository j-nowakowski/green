package green

// Export converts a potentially green-container into its native Go type. For
// such containers, this performs a deep copy.
//
// This has O(n) time complexity, where n is the total number of nodes in the
// graph representing the underlying value.
func Export(v any) any {
	switch v := v.(type) {
	case *Map:
		return v.Export()
	case *Slice:
		return v.Export()
	default:
		return v
	}
}
