package lzval

type (
	// Value has a concrete type is either ImmutableMap,
	// ImmutableSlice, or a literal Go type.
	Value any

	// MutableValue has a concrete type which is either
	// MutableMap, MutableSlice, or a literal Go type.
	MutableValue any
)

// type (
// 	// Value is an value. Its methods are safe to be called concurrently.
// 	Value struct {
// 		val any
// 	}

// 	// ValueType represents the type of the value held by a Value.
// 	ValueType int
// )

// const (
// 	TypeLiteral ValueType = iota
// 	TypeSlice
// 	TypeMap
// )

// // String returns a string representation of the ValueType.
// func (t ValueType) String() string {
// 	switch t {
// 	case TypeSlice:
// 		return "slice"
// 	case TypeMap:
// 		return "object"
// 	default:
// 		return "literal"
// 	}
// }

// // Type returns the type of the Value.
// // If the Value is nil, it returns TypeNonexistent.
// func (n *Value) Type() ValueType {
// 	if n == nil {
// 		return TypeLiteral
// 	}
// 	switch n.val.(type) {
// 	case *ImmutableMap:
// 		return TypeMap
// 	case *ImmutableSlice:
// 		return TypeSlice
// 	default:
// 		return TypeLiteral
// 	}
// }

// // Value returns the concrete value.
// // The Go type of the returned value depends on the Value's Type:
// //   - TypeNumber: float64
// //   - TypeBoolean: bool
// //   - TypeString: string
// //   - TypeMap: *ImmutableMap
// //   - TypeSlice: *ImmutableSlice
// //   - TypeNull: nil
// //
// // If the Value is nil, it returns nil.
// func (v *Value) Literal() any {
// 	if n == nil {
// 		return TypeLiteral
// 	}
// 	switch n.val.(type) {
// 	case *ImmutableMap:
// 		return TypeMap
// 	case *ImmutableSlice:
// 		return TypeSlice
// 	default:
// 		return TypeLiteral
// 	}
// }

// // Number returns the numeric value if its type is TypeNumber.
// // Otherwise, it returns 0.
// func (v *Value) Number() float64 {
// 	if v == nil {
// 		return 0
// 	}
// 	switch v.t {
// 	case TypeNumber:
// 		return v.valNumber
// 	default:
// 		return 0
// 	}
// }

// // String returns the string value if its type is TypeString.
// // Otherwise, it returns empty string.
// func (v *Value) String() string {
// 	if v == nil {
// 		return ""
// 	}
// 	switch v.t {
// 	case TypeString:
// 		return v.valString
// 	default:
// 		return ""
// 	}
// }

// // Boolean returns the value as boolean if its type is TypeBoolean.
// // Otherwise, it returns false.
// func (v *Value) Boolean() bool {
// 	if v == nil {
// 		return false
// 	}
// 	switch v.t {
// 	case TypeBoolean:
// 		return v.valBoolean
// 	default:
// 		return false
// 	}
// }

// // Slice returns the value as ImmutableSlice if its type is TypeSlice.
// // Otherwise, it returns nil.
// func (v *Value) Slice() *ImmutableSlice {
// 	if v == nil {
// 		return nil
// 	}
// 	switch v.t {
// 	case TypeSlice:
// 		return v.valSlice
// 	default:
// 		return nil
// 	}
// }

// // Map returns the value as ImmutableMap if its type is TypeMap.
// // Otherwise, it returns nil.
// func (v *Value) Map() *ImmutableMap {
// 	if v == nil {
// 		return nil
// 	}
// 	switch v.t {
// 	case TypeMap:
// 		return v.valMap
// 	default:
// 		return nil
// 	}
// }
