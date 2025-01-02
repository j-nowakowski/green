package codec

import (
	"bytes"
	"encoding/json"
	"errors"
)

type (
	JSONCodec struct{}

	// rawJSON is used to catch the bytes of
	// values inside arrays or objects
	// and stop downstream decoding. I'm using
	// this instead of json.RawMessage because
	// the latter creates a deep copy of the bytes,
	// whereas rawJSON takes a shallow copy.
	//
	// TODO: This should be benchmarked.
	rawJSON struct {
		b []byte
	}
)

func (JSONCodec) Encode(v Value) ([]byte, error) {
	return json.Marshal(v)
}

func (JSONCodec) Decode(b []byte) (Value, error) {
	switch {
	case bytes.HasPrefix(b, []byte(`{`)):
		// object
		var a map[string]rawJSON
		if err := json.Unmarshal(b, &a); err != nil {
			return nil, err
		}
		v := make(Object, len(a))
		for k, r := range a {
			v[k] = []byte(r.b)
		}
		return v, nil
	case bytes.HasPrefix(b, []byte(`[`)):
		// array
		var a []rawJSON
		if err := json.Unmarshal(b, &a); err != nil {
			return nil, err
		}
		v := make(Array, len(a))
		for i, r := range a {
			v[i] = []byte(r.b)
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

func (m *rawJSON) MarshalJSON() ([]byte, error) {
	if m == nil {
		return []byte("null"), nil
	}
	return m.b, nil
}

func (m *rawJSON) UnmarshalJSON(b []byte) error {
	if m == nil {
		return errors.New("rawJSON: UnmarshalJSON on nil pointer")
	}
	m.b = b
	return nil
}
