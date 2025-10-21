package lzval

// type (
// 	// ByteLoader wraps around a byte slice which it can decode into an
// 	// ImmutableValue on demand. Once decoded, the value is memoized for future
// 	// calls. Decoding is handled by the Codec. The methods on ByteLoader are
// 	// safe to be called concurrently.
// 	ByteLoader struct {
// 		// Payload holds the raw bytes to be loaded.
// 		// We assume that this slice is not mutated after being set;
// 		// violating this assumption may lead to unexpected behavior.
// 		Payload []byte

// 		// Codec informs how the payload is to be decoded
// 		// and how values are to be encoded.
// 		// It should not be nil or mutated after being set.
// 		Codec Codec

// 		val  ImmutableValue
// 		err  error
// 		once sync.Once
// 	}
// )

// // Load decodes the ByteLoader's Payload into an ImmutableValue.
// // If the receiver is nil or has no Payload, it returns nil.
// // An error is returned if the Codec is nil or if decoding fails.
// // If the ByteLoader has already been resolved, it returns the cached
// // ImmutableValue (or error, if any).
// func (bl *ByteLoader) Load(ctx context.Context) (ImmutableValue, error) {
// 	if bl == nil {
// 		return nil, nil
// 	}
// 	bl.once.Do(func() {
// 		if len(bl.Payload) == 0 {
// 			return
// 		}
// 		if bl.Codec == nil {
// 			bl.err = errNoCodec
// 			return
// 		}
// 		var v any
// 		v, bl.err = bl.Codec.Decode(ctx, bl.Payload)
// 		if bl.err != nil {
// 			return
// 		}
// 		switch v := v.(type) {
// 		case LoadableSlice:
// 			bl.val = &ImmutableSlice{s: v}
// 		case LoadableMap:
// 			bl.val = &ImmutableMap{m: v}
// 		default:
// 			bl.val = v
// 		}
// 	})
// 	return bl.val, bl.err
// }

// // RecursiveLoad calls Load on the value and each of its child
// // values recursively, returning a value of pure Go types.
// // The values returned are all newly constructed and the caller
// // is free to mutate them.
// func (bl *ByteLoader) RecursiveLoad(ctx context.Context) (any, error) {
// 	return recursiveLoad(ctx, bl)
// }

// var (
// 	errNoCodec = errors.New("codec should not be nil")
// )

// func recursiveLoad(ctx context.Context, v ImmutableValue) (any, error) {
// 	switch v := v.(type) {
// 	case LoadableMap:
// 		return recursiveLoadMap(ctx, v)
// 	case LoadableSlice:
// 		return recursiveLoadSlice(ctx, v)
// 	default:
// 		return nil, nil
// 	}
// }

// func recursiveLoadMap(ctx context.Context, mm LoadableMap) (map[string]any, error) {
// 	if mm == nil {
// 		return nil, nil
// 	}
// 	m := make(map[string]any, len(mm))
// 	for k, v := range mm {
// 		lv, err := v.Load(ctx)
// 		if err != nil {
// 			return nil, fmt.Errorf("load elem at key %q: %w", k, err)
// 		}
// 		fv, err := recursiveLoad(ctx, lv)
// 		if err != nil {
// 			return nil, err
// 		}
// 		m[k] = fv
// 	}
// 	return m, nil
// }

// func recursiveLoadSlice(ctx context.Context, ms LoadableSlice) ([]any, error) {
// 	if ms == nil {
// 		return nil, nil
// 	}
// 	s := make([]any, len(ms))
// 	for i, v := range ms {
// 		lv, err := v.Load(ctx)
// 		if err != nil {
// 			return nil, fmt.Errorf("load elem at index %d: %w", i, err)
// 		}
// 		fv, err := recursiveLoad(ctx, lv)
// 		if err != nil {
// 			return nil, err
// 		}
// 		s[i] = fv
// 	}
// 	return s, nil
// }
