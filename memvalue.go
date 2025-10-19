package lzval

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

type (
	// MemValue, short for "memoized value", wraps around a byte slice
	// which it can decode into a Value on demand. Once decoded, the
	// value is memoized for future calls. Decoding is handled by the
	// a Codec. The methods on MemValue are safe to be called concurrently.
	MemValue struct {
		// Payload holds the raw bytes of the Value.
		// We assume that this slice is not mutated after being set;
		// violating this assumption may lead to unexpected behavior.
		Payload []byte

		// Codec informs how the payload is to be decoded
		// and how values are to be encoded.
		// It should not be nil or mutated after being set.
		Codec Codec

		val  Value
		err  error
		once sync.Once
	}
)

// Load decodes the MemValue's Payload into a Value.
// If the receiver is nil or has no Payload, it returns nil.
// An error is returned if the Codec is nil or if decoding fails.
// If the MemValue has already been resolved, it returns the cached
// Value (or error, if any).
func (mv *MemValue) Load(ctx context.Context) (Value, error) {
	if mv == nil {
		return nil, nil
	}
	mv.once.Do(func() {
		if len(mv.Payload) == 0 {
			return
		}
		if mv.Codec == nil {
			mv.err = errNoCodec
			return
		}
		var v any
		v, mv.err = mv.Codec.Decode(ctx, mv.Payload)
		if mv.err != nil {
			return
		}
		switch v := v.(type) {
		case MemSlice:
			mv.val = &ImmutableSlice{s: v}
		case MemMap:
			mv.val = &ImmutableMap{m: v}
		default:
			mv.val = v
		}
	})
	return mv.val, mv.err
}

// RecursiveLoad calls Load on the value and each of its child
// values recursively, returning a value of pure Go types.
// The values returned are all newly constructed and the caller
// is free to mutate them.
func (mv *MemValue) RecursiveLoad(ctx context.Context) (any, error) {
	return recursiveLoad(ctx, mv)
}

var (
	errNoCodec = errors.New("codec should not be nil")
)

func recursiveLoad(ctx context.Context, mv *MemValue) (any, error) {
	if mv == nil {
		return nil, nil
	}
	switch mv := mv.val.(type) {
	case MemMap:
		return recursiveLoadMap(ctx, mv)
	case MemSlice:
		return recursiveLoadSlice(ctx, mv)
	default:
		return nil, nil
	}
}

func recursiveLoadMap(ctx context.Context, mm MemMap) (map[string]any, error) {
	if mm == nil {
		return nil, nil
	}
	m := make(map[string]any, len(mm))
	for k, v := range mm {
		lv, err := recursiveLoad(ctx, v)
		if err != nil {
			return nil, fmt.Errorf("load %q: %w", k, err)
		}
		m[k] = lv
	}
	return m, nil
}

func recursiveLoadSlice(ctx context.Context, ms MemSlice) ([]any, error) {
	if ms == nil {
		return nil, nil
	}
	s := make([]any, len(ms))
	for i, v := range ms {
		lv, err := recursiveLoad(ctx, v)
		if err != nil {
			return nil, fmt.Errorf("load %d: %w", i, err)
		}
		s[i] = lv
	}
	return s, nil
}
