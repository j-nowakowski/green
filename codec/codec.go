package codec

type (
	// Value has concrete type bool, float64, string, Array, Object, or is nil.
	Value  any
	Array  = [][]byte
	Object = map[string][]byte

	Codec interface {
		Encode(Value) ([]byte, error)
		Decode([]byte) (Value, error)
	}
)
