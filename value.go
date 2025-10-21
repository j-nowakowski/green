package lzval

type (
	// ImmutableValue has a concrete type is either ImmutableMap,
	// ImmutableSlice, or a literal Go type.
	ImmutableValue = any

	// Value has a concrete type which is either
	// Map, Slice, or a literal Go type.
	Value = any

	ValueLoader        = Loadable[ImmutableValue]
	MutableValueLoader = Loadable[Value]
)
