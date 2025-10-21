package lzval

// import (
// 	"crypto/rand"
// 	"encoding/json"
// 	"errors"
// 	"fmt"
// 	"sync"
// 	"testing"

// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/require"
// 	"golang.org/x/sync/errgroup"
// )

// func TestLazyValue(t *testing.T) {
// 	assertNilValue := func(t *testing.T, v ImmutableValue) {
// 		assert.Nil(t, v, "expected nil Value")
// 	}

// 	t.Run("if nil, return nil Value", func(t *testing.T) {
// 		var lv *ByteLoader
// 		v, err := lv.Load(t.Context())
// 		require.NoError(t, err)
// 		assertNilValue(t, v)

// 		vAny, err := lv.RecursiveLoad(t.Context())
// 		require.NoError(t, err)
// 		assert.Nil(t, vAny)
// 	})

// 	t.Run("if empty Payload, return nil Value", func(t *testing.T) {
// 		lv := &ByteLoader{Payload: []byte{}}
// 		v, err := lv.Load(t.Context())
// 		require.NoError(t, err)
// 		assertNilValue(t, v)

// 		// assert that this result is cached
// 		lv.Payload = []byte(`"foo"`)
// 		v, err = lv.Load(t.Context())
// 		require.NoError(t, err)
// 		assertNilValue(t, v)

// 		lv = &ByteLoader{Payload: []byte{}}
// 		vAny, err := lv.RecursiveLoad(t.Context())
// 		require.NoError(t, err)
// 		assert.Nil(t, vAny)

// 		lv = &ByteLoader{Payload: nil}
// 		v, err = lv.Load(t.Context())
// 		require.NoError(t, err)
// 		assertNilValue(t, v)
// 		vAny, err = lv.RecursiveLoad(t.Context())
// 		require.NoError(t, err)
// 		assert.Nil(t, vAny)

// 		// assert that this result is cached
// 		lv.Payload = []byte(`"foo"`)
// 		v, err = lv.Load(t.Context())
// 		require.NoError(t, err)
// 		assertNilValue(t, v)

// 		lv = &ByteLoader{Payload: nil}
// 		vAny, err = lv.RecursiveLoad(t.Context())
// 		require.NoError(t, err)
// 		assert.Nil(t, vAny)
// 	})

// 	t.Run("if no Codec, return error", func(t *testing.T) {
// 		lv := &ByteLoader{Payload: []byte(`"foo"`)}
// 		_, err := lv.Load(t.Context())
// 		require.ErrorIs(t, err, errNoCodec)

// 		// assert that this result is cached
// 		lv.Codec = NewMockCodec(t)
// 		_, err = lv.Load(t.Context())
// 		require.ErrorIs(t, err, errNoCodec)

// 		lv = &ByteLoader{Payload: []byte(`"foo"`)}
// 		_, err = lv.RecursiveLoad(t.Context())
// 		require.ErrorIs(t, err, errNoCodec)
// 	})

// 	t.Run("if Codec Decode returns an error, propagate it", func(t *testing.T) {
// 		codec := NewMockCodec(t)
// 		b := []byte(`"foo"`)
// 		myErr := errors.New("decode error")
// 		codec.On("Decode", b).Return(nil, myErr)
// 		lv := &ByteLoader{Payload: []byte(`"foo"`), Codec: codec}
// 		_, err := lv.Load(t.Context())
// 		require.ErrorIs(t, err, myErr, "expected myErr from Load(t.Context())")

// 		// assert that this result is cached
// 		codec2 := NewMockCodec(t)
// 		myErr2 := errors.New("decode error2")
// 		codec2.On("Decode", b).Return(nil, myErr2).Maybe() // should not be called anyway
// 		lv.Codec = codec2
// 		_, err = lv.Load(t.Context())
// 		require.ErrorIs(t, err, myErr, "expected myErr from Load(t.Context())")

// 		codec = NewMockCodec(t)
// 		codec.On("Decode", b).Return(nil, myErr)
// 		lv = &ByteLoader{Payload: []byte(`"foo"`), Codec: codec}
// 		_, err = lv.RecursiveLoad(t.Context())
// 		assert.ErrorIs(t, err, myErr)
// 	})

// 	t.Run("happy path: decode null", func(t *testing.T) {
// 		codec := NewMockCodec(t)
// 		codec.On("Decode", []byte(`foo`)).Return(nil, nil)
// 		codec.On("Decode", []byte(`foo2`)).Return(true, nil).Maybe() // should not be called anyway
// 		lv := &ByteLoader{Payload: []byte(`foo`), Codec: codec}
// 		v, err := lv.Load(t.Context())
// 		require.NoError(t, err)
// 		assert.Nil(t, v, "expected Get() return nil, instead got %v", v)

// 		// assert that this result is cached
// 		lv.Payload = []byte(`foo2`)
// 		v, err = lv.Load(t.Context())
// 		require.NoError(t, err)
// 		assert.Nil(t, v, "expected Get() return nil, instead got %v", v)

// 		codec = NewMockCodec(t)
// 		codec.On("Decode", []byte(`foo`)).Return(nil, nil)
// 		lv = &ByteLoader{Payload: []byte(`foo`), Codec: codec}
// 		v2, err := lv.RecursiveLoad(t.Context())
// 		require.NoError(t, err)
// 		assert.Nil(t, v2)
// 	})

// 	t.Run("happy path: decode boolean", func(t *testing.T) {
// 		codec := NewMockCodec(t)
// 		codec.On("Decode", []byte(`foo`)).Return(true, nil)
// 		codec.On("Decode", []byte(`foo2`)).Return(nil, nil).Maybe() // should not be called anyway
// 		lv := &ByteLoader{Payload: []byte(`foo`), Codec: codec}
// 		v, err := lv.Load(t.Context())
// 		require.NoError(t, err)
// 		assert.Equal(t, true, v, "expected Get() return true, instead got %v", v)

// 		// assert that this result is cached
// 		lv.Payload = []byte(`foo2`)
// 		v, err = lv.Load(t.Context())
// 		require.NoError(t, err)
// 		assert.Equal(t, true, v, "expected Get() return true, instead got %v", v)

// 		codec = NewMockCodec(t)
// 		codec.On("Decode", []byte(`foo`)).Return(true, nil)
// 		lv = &ByteLoader{Payload: []byte(`foo`), Codec: codec}
// 		v2, err := lv.RecursiveLoad(t.Context())
// 		require.NoError(t, err)
// 		assert.Equal(t, true, v2)
// 	})

// 	t.Run("happy path: decode number", func(t *testing.T) {
// 		codec := NewMockCodec(t)
// 		codec.On("Decode", []byte(`foo`)).Return(float64(123.456), nil)
// 		codec.On("Decode", []byte(`foo2`)).Return(float64(789.012), nil).Maybe() // should not be called anyway
// 		lv := &ByteLoader{Payload: []byte(`foo`), Codec: codec}
// 		v, err := lv.Load(t.Context())
// 		require.NoError(t, err)
// 		assert.Equal(t, float64(123.456), v, "expected Get() return 123.456, instead got %v", v)

// 		// assert that this result is cached
// 		lv.Payload = []byte(`foo2`)
// 		v, err = lv.Load(t.Context())
// 		require.NoError(t, err)
// 		assert.Equal(t, float64(123.456), v, "expected Get() return 123.456, instead got %v", v)

// 		codec = NewMockCodec(t)
// 		codec.On("Decode", []byte(`foo`)).Return(123.456, nil)
// 		lv = &ByteLoader{Payload: []byte(`foo`), Codec: codec}
// 		v2, err := lv.RecursiveLoad(t.Context())
// 		require.NoError(t, err)
// 		assert.Equal(t, 123.456, v2)
// 	})

// 	t.Run("happy path: decode string", func(t *testing.T) {
// 		codec := NewMockCodec(t)
// 		codec.On("Decode", []byte(`foo`)).Return("bar", nil)
// 		codec.On("Decode", []byte(`foo2`)).Return("bar2", nil).Maybe() // should not be called anyway
// 		lv := &ByteLoader{Payload: []byte(`foo`), Codec: codec}
// 		v, err := lv.Load(t.Context())
// 		require.NoError(t, err)
// 		assert.Equal(t, "bar", v, "expected Get() return \"bar\", instead got %v", v)

// 		// assert that this result is cached
// 		lv.Payload = []byte(`foo2`)
// 		v, err = lv.Load(t.Context())
// 		require.NoError(t, err)
// 		assert.Equal(t, "bar", v, "expected Get() return \"bar\", instead got %v", v)

// 		codec = NewMockCodec(t)
// 		codec.On("Decode", []byte(`foo`)).Return("bar", nil)
// 		lv = &ByteLoader{Payload: []byte(`foo`), Codec: codec}
// 		v2, err := lv.RecursiveLoad(t.Context())
// 		require.NoError(t, err)
// 		assert.Equal(t, "bar", v2)
// 	})

// 	t.Run("happy path: decode slice", func(t *testing.T) {
// 		codec := NewMockCodec(t)
// 		a := []*ByteLoader{new(ByteLoader), new(ByteLoader)}
// 		assertEqual := func(t *testing.T, laAny any) {
// 			t.Helper()
// 			la, ok := laAny.(*ImmutableSlice)
// 			require.True(t, ok, "expected ImmutableSlice type, got %T", laAny)
// 			assert.Equal(t, 2, la.Len(), "expected object to have 2 elements, instead got %d", la.Len())
// 			assert.Equal(t, a[0], la.At(0), "expected At(0) to return %v, instead got %v",
// 				a[0], la.At(0))
// 			assert.Equal(t, a[1], la.At(1), "expected At(1) to return %v, instead got %v",
// 				a[1], la.At(1))
// 		}
// 		a2 := []*ByteLoader{new(ByteLoader)}
// 		codec.On("Decode", []byte(`foo`)).Return(a, nil)
// 		codec.On("Decode", []byte(`foo2`)).Return(a2, nil).Maybe() // should not be called anyway
// 		lv := &ByteLoader{Payload: []byte(`foo`), Codec: codec}
// 		v, err := lv.Load(t.Context())
// 		require.NoError(t, err)
// 		assertEqual(t, v)

// 		// assert that this result is cached
// 		lv.Payload = []byte(`foo2`)
// 		v, err = lv.Load(t.Context())
// 		require.NoError(t, err)
// 		assertEqual(t, v)
// 	})

// 	t.Run("happy path: decode object", func(t *testing.T) {
// 		codec := NewMockCodec(t)
// 		m := map[string]*ByteLoader{"key1": new(ByteLoader), "key2": new(ByteLoader)}
// 		assertEqual := func(t *testing.T, loAny any) {
// 			t.Helper()
// 			lo, ok := loAny.(*ImmutableMap)
// 			require.True(t, ok, "expected ImmutableMap type, got %T", loAny)
// 			assert.Equal(t, 2, lo.Len(), "expected object to have 2 keys, instead got %d", lo.Len())
// 			got, ok := lo.Get("key1")
// 			require.True(t, ok, "expected field key1 to exist in object")
// 			assert.Equal(t, m["key1"], got, "expected Get(\"key1\") to return %v, instead got %v",
// 				m["key1"], got)
// 			got, ok = lo.Get("key2")
// 			require.True(t, ok, "expected field key2 to exist in object")
// 			assert.Equal(t, m["key2"], got, "expected Get(\"key2\") to return %v, instead got %v",
// 				m["key2"], got)
// 		}
// 		m2 := map[string]*ByteLoader{"key3": new(ByteLoader)}
// 		codec.On("Decode", []byte(`foo`)).Return(m, nil)
// 		codec.On("Decode", []byte(`foo2`)).Return(m2, nil).Maybe() // should not be called anyway
// 		lv := &ByteLoader{Payload: []byte(`foo`), Codec: codec}
// 		v, err := lv.Load(t.Context())
// 		require.NoError(t, err)
// 		assertEqual(t, v)

// 		// assert that this result is cached
// 		lv.Payload = []byte(`foo2`)
// 		v, err = lv.Load(t.Context())
// 		require.NoError(t, err)
// 		assertEqual(t, v)
// 	})

// 	t.Run("Load is concurrency safe", func(t *testing.T) {
// 		codec := NewMockCodec(t)
// 		n := 1000
// 		codec.On("Decode", []byte(`"foo"`)).Return("foo", nil).Times(1)
// 		lv := &ByteLoader{Payload: []byte(`"foo"`), Codec: codec}
// 		var eg errgroup.Group
// 		for range n {
// 			eg.Go(func() error {
// 				v, err := lv.Load(t.Context())
// 				if err != nil {
// 					return err
// 				}
// 				assert.Equal(t, "foo", v, "expected Get() return \"foo\", instead got %v", v)
// 				return nil
// 			})
// 		}
// 		require.NoError(t, eg.Wait())
// 	})
// }

// func TestImmutableMap(t *testing.T) {
// 	t.Run("Get with non-empty map", func(t *testing.T) {
// 		codec := NewMockCodec(t)
// 		m := map[string]*ByteLoader{"key1": new(ByteLoader), "key2": new(ByteLoader)}
// 		codec.On("Decode", []byte(`foo`)).Return(m, nil)
// 		lz := &ByteLoader{Payload: []byte(`foo`), Codec: codec}
// 		v, err := lz.Load(t.Context())
// 		require.NoError(t, err)
// 		lo, ok := v.(*ImmutableMap)
// 		require.True(t, ok, "expected ImmutableMap type, got %T", v)
// 		require.NotNil(t, lo)
// 		assert.Equal(t, len(m), lo.Len(), "expected object to have %d keys, instead got %d", len(m), lo.Len())
// 		got, ok := lo.Get("key1")
// 		require.True(t, ok, "expected field key1 to exist in object")
// 		assert.Equal(t, m["key1"], got, "expected Get(\"key1\") to return %v, instead got %v",
// 			m["key1"], got)
// 		got, ok = lo.Get("key2")
// 		require.True(t, ok, "expected field key2 to exist in object")
// 		assert.Equal(t, m["key2"], got, "expected Get(\"key2\") to return %v, instead got %v",
// 			m["key2"], got)
// 		got, ok = lo.Get("key3")
// 		require.True(t, ok, "expected field key3 to exist in object")
// 		assert.Nil(t, got, "expected Get(\"key3\") to return nil, instead got %v", got)
// 	})

// 	t.Run("Get with nil map", func(t *testing.T) {
// 		codec := NewMockCodec(t)
// 		m := map[string]*ByteLoader(nil)
// 		codec.On("Decode", []byte(`foo`)).Return(m, nil)
// 		lz := &ByteLoader{Payload: []byte(`foo`), Codec: codec}
// 		v, err := lz.Load(t.Context())
// 		require.NoError(t, err)
// 		lo, ok := v.(*ImmutableMap)
// 		require.True(t, ok, "expected ImmutableMap type, got %T", v)
// 		require.NotNil(t, lo)
// 		assert.Equal(t, len(m), lo.Len(), "expected object to have %d keys, instead got %d", len(m), lo.Len())
// 		got, ok := lo.Get("key1")
// 		require.True(t, ok, "expected key1 to exist")
// 		assert.Nil(t, got, "expected Get(\"key1\") to return nil, instead got %v", got)
// 	})

// 	t.Run("if nil, Len returns 0", func(t *testing.T) {
// 		var lv *ImmutableMap
// 		assert.Equal(t, 0, lv.Len(), "expected Len() on nil ImmutableMap to return 0, instead got %d", lv.Len())
// 	})

// 	t.Run("if nil, Get returns nil", func(t *testing.T) {
// 		var lv *ImmutableMap
// 		got, ok := lv.Get("foo")
// 		assert.Nil(t, got, "expected Get(\"foo\") on nil ImmutableMap to return nil, instead got %v", got)
// 		assert.False(t, ok, "expected Get(\"foo\") on nil ImmutableMap to return ok=false, instead got ok=%v", ok)
// 	})

// 	t.Run("RecursiveLoad", func(t *testing.T) {
// 		m := map[string]any{
// 			"a": nil,
// 			"b": 3.14,
// 			"c": true,
// 			"d": "foo",
// 			"e": map[string]any{"k": "v"},
// 			"f": []any{"e1", "e2"},
// 		}
// 		b, err := json.Marshal(m)
// 		require.NoError(t, err)

// 		lv := NewJSONByteLoader(b)
// 		gotAny, err := lv.RecursiveLoad(t.Context())
// 		require.NoError(t, err)
// 		assert.Equal(t, m, gotAny)

// 		lv = NewJSONByteLoader(b)
// 		v, err := lv.Load(t.Context())
// 		require.NoError(t, err)
// 		lm, ok := v.(*ImmutableMap)
// 		require.True(t, ok, "expected ImmutableMap type, got %T", v)
// 		gotM, err := lm.RecursiveLoad(t.Context())
// 		require.NoError(t, err)
// 		assert.Equal(t, m, gotM)
// 	})
// }

// func TestLazySlice(t *testing.T) {
// 	t.Run("non-empty slice", func(t *testing.T) {
// 		codec := NewMockCodec(t)
// 		a := []*ByteLoader{new(ByteLoader), new(ByteLoader), new(ByteLoader)}
// 		codec.On("Decode", []byte(`foo`)).Return(a, nil)
// 		lz := &ByteLoader{Payload: []byte(`foo`), Codec: codec}
// 		v, err := lz.Load(t.Context())
// 		require.NoError(t, err)
// 		la, ok := v.(*ImmutableSlice)
// 		require.True(t, ok, "expected LazySlice type, got %T", v)
// 		require.NotNil(t, la)
// 		assert.Equal(t, len(a), la.Len(), "expected slice to have %d elements, instead got %d", len(a), la.Len())
// 		assert.Equal(t, a[0], la.At(0), "expected At(0) to return %v, instead got %v",
// 			a[0], la.At(0))
// 		assert.Equal(t, a[1], la.At(1), "expected At(1) to return %v, instead got %v",
// 			a[1], la.At(1))
// 		assert.Equal(t, a[2], la.At(2), "expected At(2) to return %v, instead got %v",
// 			a[2], la.At(2))
// 		assert.Panics(t, func() {
// 			la.At(-1)
// 		})
// 		assert.Panics(t, func() {
// 			la.At(3)
// 		})
// 		n := 0
// 		for i, v := range la.All() {
// 			n++
// 			assert.Equal(t, a[i], v, "expected All()[%d] to return %v, instead got %v", i, a[i], v)
// 		}
// 		assert.Equal(t, len(a), n, "expected All() to iterate %d elements, instead iterated %d", len(a), n)

// 		// subslice
// 		la2 := la.SubSlice(1, 3)
// 		assert.Equal(t, 2, la2.Len(), "expected slice to have %d elements, instead got %d", 2, la2.Len())
// 		assert.Equal(t, a[1], la2.At(0), "expected At(0) to return %v, instead got %v",
// 			a[1], la2.At(0))
// 		assert.Equal(t, a[2], la2.At(1), "expected At(1) to return %v, instead got %v",
// 			a[2], la2.At(1))
// 		assert.Panics(t, func() {
// 			la.SubSlice(0, 4)
// 		})
// 		assert.Panics(t, func() {
// 			la.SubSlice(2, 1)
// 		})
// 		n = 0
// 		for i, v := range la2.All() {
// 			n++
// 			assert.Equal(t, a[i+1], v, "expected All()[%d] to return %v, instead got %v", i, a[i+1], v)
// 		}
// 		assert.Equal(t, la2.Len(), n, "expected All() to iterate %d elements, instead iterated %d", la2.Len(), n)
// 	})

// 	t.Run("nil slice", func(t *testing.T) {
// 		codec := NewMockCodec(t)
// 		a := []*ByteLoader(nil)
// 		codec.On("Decode", []byte(`foo`)).Return(a, nil)
// 		lz := &ByteLoader{Payload: []byte(`foo`), Codec: codec}
// 		v, err := lz.Load(t.Context())
// 		require.NoError(t, err)
// 		la, ok := v.(*ImmutableSlice)
// 		require.True(t, ok, "expected LazySlice type, got %T", v)
// 		require.NotNil(t, la)
// 		assert.Equal(t, len(a), la.Len(), "expected slice to have %d elements, instead got %d", len(a), la.Len())
// 		assert.Panics(t, func() {
// 			la.At(-1)
// 		})
// 		assert.Panics(t, func() {
// 			la.At(0)
// 		})

// 		// subslice
// 		la2 := la.SubSlice(0, 0)
// 		assert.Equal(t, 0, la2.Len(), "expected slice to have %d elements, instead got %d", 0, la2.Len())
// 		assert.Panics(t, func() {
// 			la.SubSlice(0, 1)
// 		})
// 		assert.Panics(t, func() {
// 			la.SubSlice(2, 1)
// 		})
// 	})

// 	t.Run("if nil, SubSlice handles", func(t *testing.T) {
// 		var lv *ImmutableSlice
// 		assert.Panics(t, func() {
// 			lv.SubSlice(0, 1)
// 		})
// 		assert.Nil(t, lv.SubSlice(0, 0)) // special case
// 	})

// 	t.Run("if nil, Len returns 0", func(t *testing.T) {
// 		var lv *ImmutableSlice
// 		assert.Equal(t, 0, lv.Len(), "expected Len() on nil LazySlice to return 0, instead got %d", lv.Len())
// 	})

// 	t.Run("if nil, At panics", func(t *testing.T) {
// 		assert.Panics(t, func() {
// 			var lv *ImmutableSlice
// 			lv.At(0)
// 		})
// 	})

// 	t.Run("RecursiveLoad", func(t *testing.T) {
// 		s := []any{nil, 3.14, true, "foo", map[string]any{"k": "v"}, []any{"e1", "e2"}}
// 		b, err := json.Marshal(s)
// 		require.NoError(t, err)

// 		lv := NewJSONByteLoader(b)
// 		gotAny, err := lv.RecursiveLoad(t.Context())
// 		require.NoError(t, err)
// 		assert.Equal(t, s, gotAny)

// 		lv = NewJSONByteLoader(b)
// 		v, err := lv.Load(t.Context())
// 		require.NoError(t, err)
// 		ls, ok := v.(*ImmutableSlice)
// 		require.True(t, ok, "expected LazySlice type, got %T", v)
// 		gotS, err := ls.RecursiveLoad(t.Context())
// 		require.NoError(t, err)
// 		assert.Equal(t, s, gotS)
// 	})
// }

// func TestJSONLazyValueAdHoc(t *testing.T) {
// 	bigMapSize := 1000
// 	m := map[string]any{
// 		"a": nil,
// 		"b": 3.14,
// 		"c": true,
// 		"d": "foo",
// 		"e": map[string]any{"k": "v"},
// 		"f": []any{"e1", "e2"},
// 		"g": map[string]any{
// 			"a": nil,
// 			"b": 3.14,
// 			"c": true,
// 			"d": "foo",
// 			"e": map[string]any{"k": "v"},
// 			"f": []any{"e1", "e2"},
// 			"g": map[string]any{
// 				"a": nil,
// 				"b": 3.14,
// 				"c": true,
// 				"d": "foo",
// 				"e": map[string]any{"k": "v"},
// 				"f": []any{"e1", "e2"},
// 			},
// 		},
// 		"h": map[string]any{
// 			"a": nil,
// 			"b": 3.14,
// 			"c": true,
// 			"d": "foo",
// 			"e": map[string]any{"k": "v"},
// 			"f": []any{"e1", "e2"},
// 			"g": map[string]any{
// 				"a": nil,
// 				"b": 3.14,
// 				"c": true,
// 				"d": "foo",
// 				"e": map[string]any{"k": "v"},
// 				"f": []any{"e1", "e2"},
// 			},
// 		},
// 	}
// 	bigMap := make(map[string]any, bigMapSize)
// 	for range bigMapSize {
// 		bigMap[rand.Text()] = rand.Text()
// 	}
// 	m["aa"] = bigMap
// 	m["zz"] = bigMap
// 	bsBig, err := json.Marshal(m)
// 	requireNoError(t, err)

// 	lv := NewJSONByteLoader(bsBig)
// 	v, err := lv.Load(t.Context())
// 	requireNoError(t, err)
// 	lvm, ok := v.(*ImmutableMap)
// 	require.True(t, ok, "expected ImmutableMap type, got %T", v)
// 	lvg, ok := lvm.Get("g")
// 	require.True(t, ok, "expected field g to exist in object")
// 	vg, err := lvg.Load(t.Context())
// 	requireNoError(t, err)
// 	lvgm, ok := vg.(*ImmutableMap)
// 	require.True(t, ok, "expected ImmutableMap type, got %T", vg)
// 	lvgg, ok := lvgm.Get("g")
// 	require.True(t, ok, "expected field g to exist in object")
// 	vgg, err := lvgg.Load(t.Context())
// 	requireNoError(t, err)
// 	vggm, ok := vgg.(*ImmutableMap)
// 	require.True(t, ok, "expected ImmutableMap type, got %T", vgg)
// 	lvggf, ok := vggm.Get("f")
// 	require.True(t, ok, "expected field f to exist in object")
// 	vggf, err := lvggf.Load(t.Context())
// 	requireNoError(t, err)
// 	vggfSlice, ok := vggf.(*ImmutableSlice)
// 	require.True(t, ok, "expected LazySlice type, got %T", vggf)
// 	elems := make([]string, vggfSlice.Len())
// 	for i, lve := range vggfSlice.All() {
// 		ve, err := lve.Load(t.Context())
// 		requireNoError(t, err)
// 		elems[i], ok = ve.(string)
// 		require.True(t, ok, "expected string type, got %T", ve)
// 	}
// 	if len(elems) != 2 || elems[0] != "e1" || elems[1] != "e2" {
// 		t.Fatalf("expected elems to be [\"e1\", \"e2\"], instead got %v", elems)
// 	}
// }

// func BenchmarkJSONLazyValue(b *testing.B) {
// 	// value at `g.g.f` we will be trying to read from the payload
// 	const throughputChunkSize = 125

// 	sizesToTest := []int{1, 10, 100, 500, 1000, 2500, 5000, 7500, 10000}

// 	for _, bigMapSize := range sizesToTest {
// 		b.Run(fmt.Sprintf("object size %d", bigMapSize), func(b *testing.B) {
// 			m := map[string]any{
// 				"a": nil,
// 				"b": 3.14,
// 				"c": true,
// 				"d": "foo",
// 				"e": map[string]any{"k": "v"},
// 				"f": []any{"e1", "e2"},
// 				"g": map[string]any{
// 					"a": nil,
// 					"b": 3.14,
// 					"c": true,
// 					"d": "foo",
// 					"e": map[string]any{"k": "v"},
// 					"f": []any{"e1", "e2"},
// 					"g": map[string]any{
// 						"a": nil,
// 						"b": 3.14,
// 						"c": true,
// 						"d": "foo",
// 						"e": map[string]any{"k": "v"},
// 						"f": []any{"e1", "e2"},
// 					},
// 				},
// 				"h": map[string]any{
// 					"a": nil,
// 					"b": 3.14,
// 					"c": true,
// 					"d": "foo",
// 					"e": map[string]any{"k": "v"},
// 					"f": []any{"e1", "e2"},
// 					"g": map[string]any{
// 						"a": nil,
// 						"b": 3.14,
// 						"c": true,
// 						"d": "foo",
// 						"e": map[string]any{"k": "v"},
// 						"f": []any{"e1", "e2"},
// 					},
// 				},
// 			}
// 			bigMap := make(map[string]any, bigMapSize)
// 			for range bigMapSize {
// 				bigMap[rand.Text()] = rand.Text()
// 			}
// 			m["aa"] = bigMap
// 			m["zz"] = bigMap
// 			bsBig, err := json.Marshal(m)
// 			requireNoError(b, err)

// 			benchLazy := func(b *testing.B, bs []byte) {
// 				lv := NewJSONByteLoader(bs)
// 				v, err := lv.Load(b.Context())
// 				requireNoError(b, err)
// 				lvm, ok := v.(*ImmutableMap)
// 				require.True(b, ok, "expected ImmutableMap type, got %T", v)
// 				lvg, ok := lvm.Get("g")
// 				require.True(b, ok, "expected field g to exist in object")
// 				vg, err := lvg.Load(b.Context())
// 				requireNoError(b, err)
// 				lvgm, ok := vg.(*ImmutableMap)
// 				require.True(b, ok, "expected ImmutableMap type, got %T", vg)
// 				lvgg, ok := lvgm.Get("g")
// 				require.True(b, ok, "expected field g to exist in object")
// 				vgg, err := lvgg.Load(b.Context())
// 				requireNoError(b, err)
// 				vggm, ok := vgg.(*ImmutableMap)
// 				require.True(b, ok, "expected ImmutableMap type, got %T", vgg)
// 				lvggf, ok := vggm.Get("f")
// 				require.True(b, ok, "expected field f to exist in object")
// 				vggf, err := lvggf.Load(b.Context())
// 				requireNoError(b, err)
// 				vggfSlice, ok := vggf.(*ImmutableSlice)
// 				require.True(b, ok, "expected LazySlice type, got %T", vggf)
// 				elems := make([]string, vggfSlice.Len())
// 				for i, lve := range vggfSlice.All() {
// 					ve, err := lve.Load(b.Context())
// 					requireNoError(b, err)
// 					elems[i], ok = ve.(string)
// 					require.True(b, ok, "expected string type, got %T", ve)
// 				}
// 				if len(elems) != 2 || elems[0] != "e1" || elems[1] != "e2" {
// 					b.Fatalf("expected elems to be [\"e1\", \"e2\"], instead got %v", elems)
// 				}
// 			}

// 			_ = benchLazy

// 			benchFullUnmarshal := func(b *testing.B, bs []byte) {
// 				var vAny any
// 				requireNoError(b, json.Unmarshal(bs, &vAny))
// 				v, ok := vAny.(map[string]any)
// 				if !ok {
// 					b.Fatal("not map")
// 				}
// 				vg, ok := v["g"].(map[string]any)
// 				if !ok {
// 					b.Fatal("not map2")
// 				}
// 				vgg, ok := vg["g"].(map[string]any)
// 				if !ok {
// 					b.Fatal("not map3")
// 				}
// 				vggf, ok := vgg["f"].([]any)
// 				if !ok {
// 					b.Fatal("not slice")
// 				}
// 				if len(vggf) != 2 {
// 					b.Fatal("not length 2")
// 				}
// 				elems := make([]string, len(vggf))
// 				elems[0], ok = vggf[0].(string)
// 				if !ok {
// 					b.Fatal("not string0")
// 				}
// 				elems[1], ok = vggf[1].(string)
// 				if !ok {
// 					b.Fatal("not string1")
// 				}
// 			}
// 			_ = benchFullUnmarshal

// 			benchmarkLatencyThroughput(b, "JSONLazyValue", throughputChunkSize, func(b *testing.B) {
// 				benchLazy(b, bsBig)
// 			})
// 			benchmarkLatencyThroughput(b, "JSONUnmarshal", throughputChunkSize, func(b *testing.B) {
// 				benchFullUnmarshal(b, bsBig)
// 			})
// 		})
// 	}
// }

// func requireNoError(b testing.TB, err error) {
// 	if err != nil {
// 		b.Fatalf("unexpected error: %v", err)
// 	}
// }

// func benchmarkLatencyThroughput(b *testing.B, name string, throughputChunkSize int, benchFn func(b *testing.B)) {
// 	b.ResetTimer()
// 	b.Run(name+":latency", func(b *testing.B) {
// 		for b.Loop() {
// 			benchFn(b)
// 		}
// 	})

// 	b.ResetTimer()
// 	b.Run(fmt.Sprintf("%s:throughput(%d)", name, throughputChunkSize), func(b *testing.B) {
// 		for b.Loop() {
// 			var wg sync.WaitGroup
// 			for range throughputChunkSize {
// 				wg.Add(1)
// 				go func() {
// 					defer wg.Done()
// 					benchFn(b)
// 				}()
// 			}
// 			wg.Wait()
// 		}
// 	})
// }
