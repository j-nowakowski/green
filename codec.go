package lzval

type (
	// DecodeValue has concrete type bool, float64, string, Array, Object, or is nil.
	DecodeValue any
	Array       = []*LazyValue
	Object      = map[string]*LazyValue

	Codec interface {
		Encode(any) ([]byte, error)
		Decode([]byte) (DecodeValue, error)
	}
)
