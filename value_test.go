package lzval

import (
	"encoding/json"
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
)

func TestLazyValue(t *testing.T) {
	assertNilValue := func(t *testing.T, v *Value) {
		assert.Nil(t, v, "expected nil Value")
		assert.Equal(t, TypeNonexistent, v.Type(), "expected Type() return TypeNonexistent, instead got %s", v.Type())
		assert.Nil(t, v.Value(), "expected Get() return nil, instead got %v", v.Value())
	}

	t.Run("if nil, return nil Value", func(t *testing.T) {
		var lv *LazyValue
		v, err := lv.Load()
		require.NoError(t, err)
		assertNilValue(t, v)

		vAny, err := lv.RecursiveLoad()
		require.NoError(t, err)
		assert.Nil(t, vAny)
	})

	t.Run("if empty Payload, return nil Value", func(t *testing.T) {
		lv := &LazyValue{Payload: []byte{}}
		v, err := lv.Load()
		require.NoError(t, err)
		assertNilValue(t, v)

		// assert that this result is cached
		lv.Payload = []byte(`"foo"`)
		v, err = lv.Load()
		require.NoError(t, err)
		assertNilValue(t, v)

		lv = &LazyValue{Payload: []byte{}}
		vAny, err := lv.RecursiveLoad()
		require.NoError(t, err)
		assert.Nil(t, vAny)

		lv = &LazyValue{Payload: nil}
		v, err = lv.Load()
		require.NoError(t, err)
		assertNilValue(t, v)
		vAny, err = lv.RecursiveLoad()
		require.NoError(t, err)
		assert.Nil(t, vAny)

		// assert that this result is cached
		lv.Payload = []byte(`"foo"`)
		v, err = lv.Load()
		require.NoError(t, err)
		assertNilValue(t, v)

		lv = &LazyValue{Payload: nil}
		vAny, err = lv.RecursiveLoad()
		require.NoError(t, err)
		assert.Nil(t, vAny)
	})

	t.Run("if no Codec, return error", func(t *testing.T) {
		lv := &LazyValue{Payload: []byte(`"foo"`)}
		_, err := lv.Load()
		require.ErrorIs(t, err, errNoCodec)

		// assert that this result is cached
		lv.Codec = NewMockCodec(t)
		_, err = lv.Load()
		require.ErrorIs(t, err, errNoCodec)

		lv = &LazyValue{Payload: []byte(`"foo"`)}
		_, err = lv.RecursiveLoad()
		require.ErrorIs(t, err, errNoCodec)
	})

	t.Run("if Codec Decode returns an error, propagate it", func(t *testing.T) {
		codec := NewMockCodec(t)
		b := []byte(`"foo"`)
		myErr := errors.New("decode error")
		codec.On("Decode", b).Return(nil, myErr)
		lv := &LazyValue{Payload: []byte(`"foo"`), Codec: codec}
		_, err := lv.Load()
		require.ErrorIs(t, err, myErr, "expected myErr from Load()")

		// assert that this result is cached
		codec2 := NewMockCodec(t)
		myErr2 := errors.New("decode error2")
		codec2.On("Decode", b).Return(nil, myErr2).Maybe() // should not be called anyway
		lv.Codec = codec2
		_, err = lv.Load()
		require.ErrorIs(t, err, myErr, "expected myErr from Load()")

		codec = NewMockCodec(t)
		codec.On("Decode", b).Return(nil, myErr)
		lv = &LazyValue{Payload: []byte(`"foo"`), Codec: codec}
		_, err = lv.RecursiveLoad()
		assert.ErrorIs(t, err, myErr)
	})

	t.Run("happy path: decode null", func(t *testing.T) {
		codec := NewMockCodec(t)
		codec.On("Decode", []byte(`foo`)).Return(nil, nil)
		codec.On("Decode", []byte(`foo2`)).Return(true, nil).Maybe() // should not be called anyway
		lv := &LazyValue{Payload: []byte(`foo`), Codec: codec}
		v, err := lv.Load()
		require.NoError(t, err)
		assert.Equal(t, TypeNull, v.Type(), "expected Type() return TypeNull, instead got %s", v.Type())
		assert.Nil(t, v.Value(), "expected Get() return nil, instead got %v", v.Value())

		// assert that this result is cached
		lv.Payload = []byte(`foo2`)
		v, err = lv.Load()
		require.NoError(t, err)
		assert.Equal(t, TypeNull, v.Type(), "expected Type() return TypeNull, instead got %s", v.Type())
		assert.Nil(t, v.Value(), "expected Get() return nil, instead got %v", v.Value())

		codec = NewMockCodec(t)
		codec.On("Decode", []byte(`foo`)).Return(nil, nil)
		lv = &LazyValue{Payload: []byte(`foo`), Codec: codec}
		v2, err := lv.RecursiveLoad()
		require.NoError(t, err)
		assert.Nil(t, v2)
	})

	t.Run("happy path: decode boolean", func(t *testing.T) {
		codec := NewMockCodec(t)
		codec.On("Decode", []byte(`foo`)).Return(true, nil)
		codec.On("Decode", []byte(`foo2`)).Return(nil, nil).Maybe() // should not be called anyway
		lv := &LazyValue{Payload: []byte(`foo`), Codec: codec}
		v, err := lv.Load()
		require.NoError(t, err)
		assert.Equal(t, TypeBoolean, v.Type(), "expected Type() return TypeBoolean, instead got %s", v.Type())
		assert.Equal(t, true, v.Value(), "expected Get() return true, instead got %v", v.Value())
		assert.Equal(t, true, v.Boolean(), "expected Boolean() return true, instead got %v", v.Boolean())

		// assert that this result is cached
		lv.Payload = []byte(`foo2`)
		v, err = lv.Load()
		require.NoError(t, err)
		assert.Equal(t, TypeBoolean, v.Type(), "expected Type() return TypeBoolean, instead got %s", v.Type())
		assert.Equal(t, true, v.Value(), "expected Get() return true, instead got %v", v.Value())
		assert.Equal(t, true, v.Boolean(), "expected Boolean() return true, instead got %v", v.Boolean())

		codec = NewMockCodec(t)
		codec.On("Decode", []byte(`foo`)).Return(true, nil)
		lv = &LazyValue{Payload: []byte(`foo`), Codec: codec}
		v2, err := lv.RecursiveLoad()
		require.NoError(t, err)
		assert.Equal(t, true, v2)
	})

	t.Run("happy path: decode number", func(t *testing.T) {
		codec := NewMockCodec(t)
		codec.On("Decode", []byte(`foo`)).Return(float64(123.456), nil)
		codec.On("Decode", []byte(`foo2`)).Return(float64(789.012), nil).Maybe() // should not be called anyway
		lv := &LazyValue{Payload: []byte(`foo`), Codec: codec}
		v, err := lv.Load()
		require.NoError(t, err)
		assert.Equal(t, TypeNumber, v.Type(), "expected Type() return TypeNumber, instead got %s", v.Type())
		assert.Equal(t, float64(123.456), v.Value(), "expected Get() return 123.456, instead got %v", v.Value())
		assert.Equal(t, float64(123.456), v.Number(), "expected Number() return 123.456, instead got %v", v.Number())

		// assert that this result is cached
		lv.Payload = []byte(`foo2`)
		v, err = lv.Load()
		require.NoError(t, err)
		assert.Equal(t, TypeNumber, v.Type(), "expected Type() return TypeNumber, instead got %s", v.Type())
		assert.Equal(t, float64(123.456), v.Value(), "expected Get() return 123.456, instead got %v", v.Value())
		assert.Equal(t, float64(123.456), v.Number(), "expected Number() return 123.456, instead got %v", v.Number())

		codec = NewMockCodec(t)
		codec.On("Decode", []byte(`foo`)).Return(123.456, nil)
		lv = &LazyValue{Payload: []byte(`foo`), Codec: codec}
		v2, err := lv.RecursiveLoad()
		require.NoError(t, err)
		assert.Equal(t, 123.456, v2)
	})

	t.Run("happy path: decode string", func(t *testing.T) {
		codec := NewMockCodec(t)
		codec.On("Decode", []byte(`foo`)).Return("bar", nil)
		codec.On("Decode", []byte(`foo2`)).Return("bar2", nil).Maybe() // should not be called anyway
		lv := &LazyValue{Payload: []byte(`foo`), Codec: codec}
		v, err := lv.Load()
		require.NoError(t, err)
		assert.Equal(t, TypeString, v.Type(), "expected Type() return TypeString, instead got %s", v.Type())
		assert.Equal(t, "bar", v.Value(), "expected Get() return \"bar\", instead got %v", v.Value())
		assert.Equal(t, "bar", v.String(), "expected String() return \"bar\", instead got %q", v.String())

		// assert that this result is cached
		lv.Payload = []byte(`foo2`)
		v, err = lv.Load()
		require.NoError(t, err)
		assert.Equal(t, TypeString, v.Type(), "expected Type() return TypeString, instead got %s", v.Type())
		assert.Equal(t, "bar", v.Value(), "expected Get() return \"bar\", instead got %v", v.Value())
		assert.Equal(t, "bar", v.String(), "expected String() return \"bar\", instead got %q", v.String())

		codec = NewMockCodec(t)
		codec.On("Decode", []byte(`foo`)).Return("bar", nil)
		lv = &LazyValue{Payload: []byte(`foo`), Codec: codec}
		v2, err := lv.RecursiveLoad()
		require.NoError(t, err)
		assert.Equal(t, "bar", v2)
	})

	t.Run("happy path: decode slice", func(t *testing.T) {
		codec := NewMockCodec(t)
		a := []*LazyValue{new(LazyValue), new(LazyValue)}
		assertEqual := func(t *testing.T, laAny any) {
			t.Helper()
			la, ok := laAny.(*LazySlice)
			require.True(t, ok, "expected LazyMap type, got %T", laAny)
			assert.Equal(t, 2, la.Len(), "expected object to have 2 elements, instead got %d", la.Len())
			assert.Equal(t, a[0], la.At(0), "expected At(0) to return %v, instead got %v",
				a[0], la.At(0))
			assert.Equal(t, a[1], la.At(1), "expected At(1) to return %v, instead got %v",
				a[1], la.At(1))
		}
		a2 := []*LazyValue{new(LazyValue)}
		codec.On("Decode", []byte(`foo`)).Return(a, nil)
		codec.On("Decode", []byte(`foo2`)).Return(a2, nil).Maybe() // should not be called anyway
		lv := &LazyValue{Payload: []byte(`foo`), Codec: codec}
		v, err := lv.Load()
		require.NoError(t, err)
		assert.Equal(t, TypeSlice, v.Type(), "expected Type() return TypeSlice, instead got %s", v.Type())
		assertEqual(t, v.Value())
		assertEqual(t, v.Slice())

		// assert that this result is cached
		lv.Payload = []byte(`foo2`)
		v, err = lv.Load()
		require.NoError(t, err)
		assert.Equal(t, TypeSlice, v.Type(), "expected Type() return TypeSlice, instead got %s", v.Type())
		assertEqual(t, v.Value())
		assertEqual(t, v.Slice())
	})

	t.Run("happy path: decode object", func(t *testing.T) {
		codec := NewMockCodec(t)
		m := map[string]*LazyValue{"key1": new(LazyValue), "key2": new(LazyValue)}
		assertEqual := func(t *testing.T, loAny any) {
			t.Helper()
			lo, ok := loAny.(*LazyMap)
			require.True(t, ok, "expected LazyMap type, got %T", loAny)
			assert.Equal(t, 2, lo.Len(), "expected object to have 2 keys, instead got %d", lo.Len())
			assert.Equal(t, m["key1"], lo.Get("key1"), "expected Get(\"key1\") to return %v, instead got %v",
				m["key1"], lo.Get("key1"))
			assert.Equal(t, m["key2"], lo.Get("key2"), "expected Get(\"key2\") to return %v, instead got %v",
				m["key2"], lo.Get("key2"))
		}
		m2 := map[string]*LazyValue{"key3": new(LazyValue)}
		codec.On("Decode", []byte(`foo`)).Return(m, nil)
		codec.On("Decode", []byte(`foo2`)).Return(m2, nil).Maybe() // should not be called anyway
		lv := &LazyValue{Payload: []byte(`foo`), Codec: codec}
		v, err := lv.Load()
		require.NoError(t, err)
		assert.Equal(t, TypeMap, v.Type(), "expected Type() return TypeMap, instead got %s", v.Type())
		assertEqual(t, v.Value())
		assertEqual(t, v.Map())

		// assert that this result is cached
		lv.Payload = []byte(`foo2`)
		v, err = lv.Load()
		require.NoError(t, err)
		assert.Equal(t, TypeMap, v.Type(), "expected Type() return TypeMap, instead got %s", v.Type())
		assertEqual(t, v.Value())
		assertEqual(t, v.Map())
	})

	t.Run("if nil, UnmarshalJSON errors", func(t *testing.T) {
		var lv *LazyValue
		err := lv.UnmarshalJSON([]byte(`"foo"`))
		require.Error(t, err)
	})

	t.Run("UnmarshalJSON happy path", func(t *testing.T) {
		lv := &LazyValue{}
		err := lv.UnmarshalJSON([]byte(`"foo"`))
		require.NoError(t, err)
		assert.Equal(t, []byte(`"foo"`), lv.Payload, "expected Payload to be %q, instead got %q", `"foo"`, lv.Payload)
		assert.Equal(t, JSONCodec{}, lv.Codec, "expected Codec to be JSONCodec{}, instead got %T", lv.Codec)
	})

	t.Run("Load is concurrency safe", func(t *testing.T) {
		codec := NewMockCodec(t)
		n := 1000
		codec.On("Decode", []byte(`"foo"`)).Return("foo", nil).Times(1)
		lv := &LazyValue{Payload: []byte(`"foo"`), Codec: codec}
		var eg errgroup.Group
		for range n {
			eg.Go(func() error {
				v, err := lv.Load()
				if err != nil {
					return err
				}
				assert.Equal(t, TypeString, v.Type(), "expected Type() return TypeString, instead got %s", v.Type())
				assert.Equal(t, "foo", v.Value(), "expected Get() return \"foo\", instead got %v", v.Value())
				assert.Equal(t, "foo", v.String(), "expected String() return \"foo\", instead got %q", v.String())
				return nil
			})
		}
		require.NoError(t, eg.Wait())
	})
}

func TestLazyMap(t *testing.T) {
	t.Run("Get with non-empty map", func(t *testing.T) {
		codec := NewMockCodec(t)
		m := map[string]*LazyValue{"key1": new(LazyValue), "key2": new(LazyValue)}
		codec.On("Decode", []byte(`foo`)).Return(m, nil)
		lz := &LazyValue{Payload: []byte(`foo`), Codec: codec}
		v, err := lz.Load()
		require.NoError(t, err)
		lo := v.Map()
		require.NotNil(t, lo)
		assert.Equal(t, len(m), lo.Len(), "expected object to have %d keys, instead got %d", len(m), lo.Len())
		assert.Equal(t, m["key1"], lo.Get("key1"), "expected Get(\"key1\") to return %v, instead got %v",
			m["key1"], lo.Get("key1"))
		assert.Equal(t, m["key2"], lo.Get("key2"), "expected Get(\"key2\") to return %v, instead got %v",
			m["key2"], lo.Get("key2"))
		assert.Nil(t, lo.Get("key3"), "expected Get(\"key3\") to return nil, instead got %v", lo.Get("key1"))
	})

	t.Run("Get with nil map", func(t *testing.T) {
		codec := NewMockCodec(t)
		m := map[string]*LazyValue(nil)
		codec.On("Decode", []byte(`foo`)).Return(m, nil)
		lz := &LazyValue{Payload: []byte(`foo`), Codec: codec}
		v, err := lz.Load()
		require.NoError(t, err)
		lo := v.Map()
		require.NotNil(t, lo)
		assert.Equal(t, len(m), lo.Len(), "expected object to have %d keys, instead got %d", len(m), lo.Len())
		assert.Nil(t, lo.Get("key1"), "expected Get(\"key1\") to return nil, instead got %v", lo.Get("key1"))
	})

	t.Run("if nil, Len returns 0", func(t *testing.T) {
		var lv *LazyMap
		assert.Equal(t, 0, lv.Len(), "expected Len() on nil LazyMap to return 0, instead got %d", lv.Len())
	})

	t.Run("if nil, Get returns nil", func(t *testing.T) {
		var lv *LazyMap
		assert.Nil(t, lv.Get("foo"), "expected Get(\"foo\") on nil LazyMap to return nil, instead got %v",
			lv.Get("foo"))
	})

	t.Run("RecursiveLoad", func(t *testing.T) {
		m := map[string]any{
			"a": nil,
			"b": 3.14,
			"c": true,
			"d": "foo",
			"e": map[string]any{"k": "v"},
			"f": []any{"e1", "e2"},
		}
		b, err := json.Marshal(m)
		require.NoError(t, err)

		lv := NewJSON(b)
		gotAny, err := lv.RecursiveLoad()
		require.NoError(t, err)
		assert.Equal(t, m, gotAny)

		lv = NewJSON(b)
		v, err := lv.Load()
		require.NoError(t, err)
		require.Equal(t, TypeMap, lv.val.Type())
		lm := v.Map()
		gotM, err := lm.RecursiveLoad()
		require.NoError(t, err)
		assert.Equal(t, m, gotM)
	})
}

func TestLazySlice(t *testing.T) {
	t.Run("non-empty slice", func(t *testing.T) {
		codec := NewMockCodec(t)
		a := []*LazyValue{new(LazyValue), new(LazyValue), new(LazyValue)}
		codec.On("Decode", []byte(`foo`)).Return(a, nil)
		lz := &LazyValue{Payload: []byte(`foo`), Codec: codec}
		v, err := lz.Load()
		require.NoError(t, err)
		la := v.Slice()
		require.NotNil(t, la)
		assert.Equal(t, len(a), la.Len(), "expected slice to have %d elements, instead got %d", len(a), la.Len())
		assert.Equal(t, a[0], la.At(0), "expected At(0) to return %v, instead got %v",
			a[0], la.At(0))
		assert.Equal(t, a[1], la.At(1), "expected At(1) to return %v, instead got %v",
			a[1], la.At(1))
		assert.Equal(t, a[2], la.At(2), "expected At(2) to return %v, instead got %v",
			a[2], la.At(2))
		assert.Panics(t, func() {
			la.At(-1)
		})
		assert.Panics(t, func() {
			la.At(3)
		})
		n := 0
		for i, v := range la.All() {
			n++
			assert.Equal(t, a[i], v, "expected All()[%d] to return %v, instead got %v", i, a[i], v)
		}
		assert.Equal(t, len(a), n, "expected All() to iterate %d elements, instead iterated %d", len(a), n)

		// subslice
		la2 := la.SubSlice(1, 3)
		assert.Equal(t, 2, la2.Len(), "expected slice to have %d elements, instead got %d", 2, la2.Len())
		assert.Equal(t, a[1], la2.At(0), "expected At(0) to return %v, instead got %v",
			a[1], la2.At(0))
		assert.Equal(t, a[2], la2.At(1), "expected At(1) to return %v, instead got %v",
			a[2], la2.At(1))
		assert.Panics(t, func() {
			la.SubSlice(0, 4)
		})
		assert.Panics(t, func() {
			la.SubSlice(2, 1)
		})
		n = 0
		for i, v := range la2.All() {
			n++
			assert.Equal(t, a[i+1], v, "expected All()[%d] to return %v, instead got %v", i, a[i+1], v)
		}
		assert.Equal(t, la2.Len(), n, "expected All() to iterate %d elements, instead iterated %d", la2.Len(), n)
	})

	t.Run("nil slice", func(t *testing.T) {
		codec := NewMockCodec(t)
		a := []*LazyValue(nil)
		codec.On("Decode", []byte(`foo`)).Return(a, nil)
		lz := &LazyValue{Payload: []byte(`foo`), Codec: codec}
		v, err := lz.Load()
		require.NoError(t, err)
		la := v.Slice()
		require.NotNil(t, la)
		assert.Equal(t, len(a), la.Len(), "expected slice to have %d elements, instead got %d", len(a), la.Len())
		assert.Panics(t, func() {
			la.At(-1)
		})
		assert.Panics(t, func() {
			la.At(0)
		})

		// subslice
		la2 := la.SubSlice(0, 0)
		assert.Equal(t, 0, la2.Len(), "expected slice to have %d elements, instead got %d", 0, la2.Len())
		assert.Panics(t, func() {
			la.SubSlice(0, 1)
		})
		assert.Panics(t, func() {
			la.SubSlice(2, 1)
		})
	})

	t.Run("if nil, SubSlice handles", func(t *testing.T) {
		var lv *LazySlice
		assert.Panics(t, func() {
			lv.SubSlice(0, 1)
		})
		assert.Nil(t, lv.SubSlice(0, 0)) // special case
	})

	t.Run("if nil, Len returns 0", func(t *testing.T) {
		var lv *LazySlice
		assert.Equal(t, 0, lv.Len(), "expected Len() on nil LazySlice to return 0, instead got %d", lv.Len())
	})

	t.Run("if nil, At panics", func(t *testing.T) {
		assert.Panics(t, func() {
			var lv *LazySlice
			lv.At(0)
		})
	})

	t.Run("RecursiveLoad", func(t *testing.T) {
		s := []any{nil, 3.14, true, "foo", map[string]any{"k": "v"}, []any{"e1", "e2"}}
		b, err := json.Marshal(s)
		require.NoError(t, err)

		lv := NewJSON(b)
		gotAny, err := lv.RecursiveLoad()
		require.NoError(t, err)
		assert.Equal(t, s, gotAny)

		lv = NewJSON(b)
		v, err := lv.Load()
		require.NoError(t, err)
		require.Equal(t, TypeSlice, lv.val.Type())
		ls := v.Slice()
		gotS, err := ls.RecursiveLoad()
		require.NoError(t, err)
		assert.Equal(t, s, gotS)
	})
}

func BenchmarkValue(b *testing.B) {
	/*
		Benchmarking found that the current implementation is slightly superior
		to an alternative implementation that uses a single `any` field
		and type assertions.

		BenchmarkValue/latency:_current_value_impl-8         	85498476	        13.87 ns/op	      16 B/op	       1 allocs/op
		BenchmarkValue/latency:_alt_value_impl-8             	83965011	        14.28 ns/op	      16 B/op	       1 allocs/op
		BenchmarkValue/throughput:_current_value_impl-8      	 8560072	       140.5 ns/op	      16 B/op	       1 allocs/op
		BenchmarkValue/throughput:_alt_value_impl-8          	 8550756	       140.7 ns/op	      16 B/op	       1 allocs/op
	*/

	b.Run("latency: current value impl", func(b *testing.B) {
		for b.Loop() {
			v := &Value{
				t:         TypeString,
				valString: "foo",
			}
			_ = v.String()
			_ = v.Value()
		}
	})

	b.Run("latency: alt value impl", func(b *testing.B) {
		for b.Loop() {
			v := &altValue{
				t:   TypeString,
				val: "foo",
			}
			_ = v.String()
			_ = v.Value()
		}
	})

	b.Run("throughput: current value impl", func(b *testing.B) {
		var wg sync.WaitGroup
		for b.Loop() {
			wg.Add(1)
			go func() {
				defer wg.Done()
				v := &Value{
					t:         TypeString,
					valString: "foo",
				}
				_ = v.String()
				_ = v.Value()
			}()
		}
		wg.Wait()
	})

	b.Run("throughput: alt value impl", func(b *testing.B) {
		var wg sync.WaitGroup
		for b.Loop() {
			wg.Add(1)
			go func() {
				defer wg.Done()
				v := &altValue{
					t:   TypeString,
					val: "foo",
				}
				_ = v.String()
				_ = v.Value()
			}()
		}
		wg.Wait()
	})
}

type altValue struct {
	t   ValueType
	val any
}

func (v *altValue) Value() DecodeValue {
	if v == nil {
		return nil
	}
	switch v.t {
	case TypeNumber:
		vv, _ := v.val.(float64)
		return vv
	case TypeBoolean:
		vv, _ := v.val.(bool)
		return vv
	case TypeString:
		vv, _ := v.val.(string)
		return vv
	case TypeMap:
		vv, _ := v.val.(*LazyMap)
		return vv
	case TypeSlice:
		vv, _ := v.val.(*LazySlice)
		return vv
	default:
		return nil
	}
}

func (v *altValue) String() string {
	if v == nil {
		return ""
	}
	switch v.t {
	case TypeString:
		vv, _ := v.val.(string)
		return vv
	default:
		return ""
	}
}
