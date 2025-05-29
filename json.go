package lzval

import (
	"bytes"
	"encoding/json"
)

type (
	// JSONCodec is a codec for encoding and decoding JSON data.
	// It simply wraps the standard library's JSON encoding/decoding,
	// but stops arrays and objects from being decoded recursively.
	JSONCodec struct{}
)

// NewJSONLazyValue prepares a LazyValue that will decode JSON data lazily.
func NewJSONLazyValue(b json.RawMessage) *LazyValue {
	return &LazyValue{
		Payload: b,
		Codec:   JSONCodec{},
	}
}

func (JSONCodec) Encode(v any) ([]byte, error) {
	return json.Marshal(v)
}

func (JSONCodec) Decode(b []byte) (DecodeValue, error) {
	b = bytes.TrimSpace(b)
	switch {
	case bytes.HasPrefix(b, []byte(`{`)):
		var v map[string]*LazyValue
		if err := json.Unmarshal(b, &v); err != nil {
			return nil, err
		}
		return v, nil
	case bytes.HasPrefix(b, []byte(`[`)):
		var v []*LazyValue
		if err := json.Unmarshal(b, &v); err != nil {
			return nil, err
		}
		return v, nil
	default:
		var a any
		if err := json.Unmarshal(b, &a); err != nil {
			return nil, err
		}
		return a, nil
	}
}
