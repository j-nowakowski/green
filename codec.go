package lzval

//go:generate mockery --name=Codec
type (
	// DecodeValue has concrete type bool, float64, string, Slice, Map, or is nil.
	DecodeValue any
	Slice       = []*LazyValue
	Map         = map[string]*LazyValue

	Codec interface {
		Encode(any) ([]byte, error)
		Decode([]byte) (DecodeValue, error)
	}
)
