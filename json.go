package lzval

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
)

type (
	JSONCodec struct{}
)

// NewJSON prepares a LazyValue that will decode JSON data lazily.
func NewJSON(b json.RawMessage) *MemValue {
	return &MemValue{
		Payload: b,
		Codec:   JSONCodec{},
	}
}

func (JSONCodec) Decode(_ context.Context, b []byte) (any, error) {
	b = bytes.TrimSpace(b)
	switch {
	case bytes.HasPrefix(b, []byte{'{'}):
		return splitJSONMap(b)
	case bytes.HasPrefix(b, []byte{'['}):
		return splitJSONArray(b)
	default:
		var v any
		// TODO: might be faster to generate code for this too
		if err := json.Unmarshal(b, &v); err != nil {
			return nil, err
		}
		return v, nil
	}
}

func (JSONCodec) Encode(_ context.Context, v any) ([]byte, error) {
	return json.Marshal(v)
}

func splitJSONArray(data []byte) (MemSlice, error) {
	if !bytes.HasSuffix(data, []byte{']'}) {
		return nil, errors.New("input is not a JSON array")
	}

	elems := MemSlice{}
	start := 1
	depthSquare := 0
	depthCurly := 0
	inString := false
	escape := false
	for i := 1; i < len(data)-1; i++ {
		b := data[i]

		if inString {
			switch {
			case escape:
				escape = false
			case b == '\\':
				escape = true
			case b == '"':
				inString = false
			}
			continue
		}

		switch b {
		case '"':
			inString = true
		case '[':
			depthSquare++
		case '{':
			depthCurly++
		case '}':
			if depthCurly == 0 {
				return nil, errors.New("unmatched curly brace in JSON array")
			}
			depthCurly--
		case ']':
			if depthSquare == 0 {
				return nil, errors.New("unmatched square bracket in JSON array")
			}
			depthSquare--
		case ',':
			if depthSquare == 0 && depthCurly == 0 {
				elems = append(elems, &MemValue{Payload: data[start:i], Codec: JSONCodec{}})
				start = i + 1
			}
		}
	}
	if len(data)-1 > start {
		elems = append(elems, &MemValue{Payload: data[start : len(data)-1], Codec: JSONCodec{}})
	}

	return elems, nil
}

func splitJSONMap(data []byte) (MemMap, error) {
	if !bytes.HasSuffix(data, []byte{'}'}) {
		return nil, errors.New("input is not a JSON object")
	}

	m := MemMap{}
	start := 1
	depthSquare := 0
	depthCurly := 0
	inString := false
	escape := false
	afterColon := false
	var key string
	for i := 1; i < len(data)-1; i++ {
		b := data[i]

		if inString {
			switch {
			case escape:
				escape = false
			case b == '\\':
				escape = true
			case b == '"':
				inString = false
			}
			continue
		}

		switch b {
		case '"':
			inString = true
		case '[':
			depthSquare++
		case '{':
			depthCurly++
		case '}':
			if depthCurly == 0 {
				return nil, errors.New("unmatched curly brace in JSON array")
			}
			depthCurly--
		case ']':
			if depthSquare == 0 {
				return nil, errors.New("unmatched square bracket in JSON array")
			}
			depthSquare--
		case ':':
			if depthSquare == 0 && depthCurly == 0 {
				if afterColon {
					return nil, errors.New("unexpected ':' in JSON object")
				}
				keyBytes := bytes.TrimSpace(data[start:i])
				if len(keyBytes) < 2 || keyBytes[0] != '"' || keyBytes[len(keyBytes)-1] != '"' {
					return nil, errors.New("invalid JSON object key")
				}
				key = string(keyBytes[1 : len(keyBytes)-1])
				afterColon = true
				start = i + 1
			}
		case ',':
			if depthSquare == 0 && depthCurly == 0 {
				if !afterColon {
					return nil, errors.New("missing ':' in JSON object")
				}
				m[key] = &MemValue{Payload: data[start:i], Codec: JSONCodec{}}
				afterColon = false
				start = i + 1
			}
		}
	}
	if len(data)-1 > start {
		if !afterColon {
			return nil, errors.New("missing ':' in JSON object")
		}
		m[key] = &MemValue{Payload: data[start : len(data)-1], Codec: JSONCodec{}}
	}

	return m, nil
}
