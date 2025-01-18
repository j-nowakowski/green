package lznode

type (
	// Value has concrete type bool, float64, string, Array, Object, or is nil.
	Value  any
	Array  = []*Node
	Object = map[string]*Node

	Codec interface {
		Encode(Value) ([]byte, error)
		Decode([]byte) (Value, error)
	}
)
