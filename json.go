package lzval

import (
	"bytes"
	"encoding/json"
)

type (
	JSONCodec struct{}
)

func NewRawJSONNode(b []byte) *LazyValue {
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
