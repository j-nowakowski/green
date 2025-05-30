package lzval

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
)

func TestLazyValue(t *testing.T) {
	assertNilValue := func(t *testing.T, v *Value) {
		assert.Nil(t, v, "expected nil Value")
		assert.Equal(t, TypeNonexistent, v.Type(), "expected Type() return TypeNonexistent, instead got %s", v.Type())
		assert.Nil(t, v.Get(), "expected Get() return nil, instead got %v", v.Get())
	}

	t.Run("if nil, return nil Value", func(t *testing.T) {
		var lv *LazyValue
		v, err := lv.Load()
		require.NoError(t, err)
		assertNilValue(t, v)
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

		lv = &LazyValue{Payload: nil}
		v, err = lv.Load()
		require.NoError(t, err)
		assertNilValue(t, v)

		// assert that this result is cached
		lv.Payload = []byte(`"foo"`)
		v, err = lv.Load()
		require.NoError(t, err)
		assertNilValue(t, v)
	})

	t.Run("if no Codec, return error", func(t *testing.T) {
		lv := &LazyValue{Payload: []byte(`"foo"`)}
		_, err := lv.Load()
		require.ErrorIs(t, err, errNoCodec)

		// assert that this result is cached
		lv.Codec = NewMockCodec(t)
		_, err = lv.Load()
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
	})

	t.Run("happy path: decode null", func(t *testing.T) {
		codec := NewMockCodec(t)
		codec.On("Decode", []byte(`foo`)).Return(nil, nil)
		codec.On("Decode", []byte(`foo2`)).Return(true, nil).Maybe() // should not be called anyway
		lv := &LazyValue{Payload: []byte(`foo`), Codec: codec}
		v, err := lv.Load()
		require.NoError(t, err)
		assert.Equal(t, TypeNull, v.Type(), "expected Type() return TypeNull, instead got %s", v.Type())
		assert.Nil(t, v.Get(), "expected Get() return nil, instead got %v", v.Get())

		// assert that this result is cached
		lv.Payload = []byte(`foo2`)
		v, err = lv.Load()
		require.NoError(t, err)
		assert.Equal(t, TypeNull, v.Type(), "expected Type() return TypeNull, instead got %s", v.Type())
		assert.Nil(t, v.Get(), "expected Get() return nil, instead got %v", v.Get())
	})

	t.Run("happy path: decode boolean", func(t *testing.T) {
		codec := NewMockCodec(t)
		codec.On("Decode", []byte(`foo`)).Return(true, nil)
		codec.On("Decode", []byte(`foo2`)).Return(nil, nil).Maybe() // should not be called anyway
		lv := &LazyValue{Payload: []byte(`foo`), Codec: codec}
		v, err := lv.Load()
		require.NoError(t, err)
		assert.Equal(t, TypeBoolean, v.Type(), "expected Type() return TypeBoolean, instead got %s", v.Type())
		assert.Equal(t, true, v.Get(), "expected Get() return true, instead got %v", v.Get())
		assert.Equal(t, true, v.Boolean(), "expected Boolean() return true, instead got %v", v.Boolean())

		// assert that this result is cached
		lv.Payload = []byte(`foo2`)
		v, err = lv.Load()
		require.NoError(t, err)
		assert.Equal(t, TypeBoolean, v.Type(), "expected Type() return TypeBoolean, instead got %s", v.Type())
		assert.Equal(t, true, v.Get(), "expected Get() return true, instead got %v", v.Get())
		assert.Equal(t, true, v.Boolean(), "expected Boolean() return true, instead got %v", v.Boolean())
	})

	t.Run("happy path: decode number", func(t *testing.T) {
		codec := NewMockCodec(t)
		codec.On("Decode", []byte(`foo`)).Return(float64(123.456), nil)
		codec.On("Decode", []byte(`foo2`)).Return(float64(789.012), nil).Maybe() // should not be called anyway
		lv := &LazyValue{Payload: []byte(`foo`), Codec: codec}
		v, err := lv.Load()
		require.NoError(t, err)
		assert.Equal(t, TypeNumber, v.Type(), "expected Type() return TypeNumber, instead got %s", v.Type())
		assert.Equal(t, float64(123.456), v.Get(), "expected Get() return 123.456, instead got %v", v.Get())
		assert.Equal(t, float64(123.456), v.Number(), "expected Number() return 123.456, instead got %v", v.Number())

		// assert that this result is cached
		lv.Payload = []byte(`foo2`)
		v, err = lv.Load()
		require.NoError(t, err)
		assert.Equal(t, TypeNumber, v.Type(), "expected Type() return TypeNumber, instead got %s", v.Type())
		assert.Equal(t, float64(123.456), v.Get(), "expected Get() return 123.456, instead got %v", v.Get())
		assert.Equal(t, float64(123.456), v.Number(), "expected Number() return 123.456, instead got %v", v.Number())
	})

	t.Run("happy path: decode string", func(t *testing.T) {
		codec := NewMockCodec(t)
		codec.On("Decode", []byte(`foo`)).Return("bar", nil)
		codec.On("Decode", []byte(`foo2`)).Return("bar2", nil).Maybe() // should not be called anyway
		lv := &LazyValue{Payload: []byte(`foo`), Codec: codec}
		v, err := lv.Load()
		require.NoError(t, err)
		assert.Equal(t, TypeString, v.Type(), "expected Type() return TypeString, instead got %s", v.Type())
		assert.Equal(t, "bar", v.Get(), "expected Get() return \"bar\", instead got %v", v.Get())
		assert.Equal(t, "bar", v.String(), "expected String() return \"bar\", instead got %q", v.String())

		// assert that this result is cached
		lv.Payload = []byte(`foo2`)
		v, err = lv.Load()
		require.NoError(t, err)
		assert.Equal(t, TypeString, v.Type(), "expected Type() return TypeString, instead got %s", v.Type())
		assert.Equal(t, "bar", v.Get(), "expected Get() return \"bar\", instead got %v", v.Get())
		assert.Equal(t, "bar", v.String(), "expected String() return \"bar\", instead got %q", v.String())
	})

	t.Run("happy path: decode array", func(t *testing.T) {
		codec := NewMockCodec(t)
		a := []*LazyValue{new(LazyValue), new(LazyValue)}
		assertEqual := func(t *testing.T, laAny any) {
			t.Helper()
			la, ok := laAny.(*LazyArray)
			require.True(t, ok, "expected LazyObject type, got %T", laAny)
			assert.Equal(t, 2, la.Len(), "expected object to have 2 elements, instead got %d", la.Len())
			assert.Equal(t, a[0], la.Element(0), "expected Element(0) to return %v, instead got %v",
				a[0], la.Element(0))
			assert.Equal(t, a[1], la.Element(1), "expected Element(1) to return %v, instead got %v",
				a[1], la.Element(1))
		}
		a2 := []*LazyValue{new(LazyValue)}
		codec.On("Decode", []byte(`foo`)).Return(a, nil)
		codec.On("Decode", []byte(`foo2`)).Return(a2, nil).Maybe() // should not be called anyway
		lv := &LazyValue{Payload: []byte(`foo`), Codec: codec}
		v, err := lv.Load()
		require.NoError(t, err)
		assert.Equal(t, TypeArray, v.Type(), "expected Type() return TypeArray, instead got %s", v.Type())
		assertEqual(t, v.Get())
		assertEqual(t, v.Array())

		// assert that this result is cached
		lv.Payload = []byte(`foo2`)
		v, err = lv.Load()
		require.NoError(t, err)
		assert.Equal(t, TypeArray, v.Type(), "expected Type() return TypeArray, instead got %s", v.Type())
		assertEqual(t, v.Get())
		assertEqual(t, v.Array())
	})

	t.Run("happy path: decode object", func(t *testing.T) {
		codec := NewMockCodec(t)
		m := map[string]*LazyValue{"key1": new(LazyValue), "key2": new(LazyValue)}
		assertEqual := func(t *testing.T, loAny any) {
			t.Helper()
			lo, ok := loAny.(*LazyObject)
			require.True(t, ok, "expected LazyObject type, got %T", loAny)
			assert.Equal(t, 2, lo.Len(), "expected object to have 2 keys, instead got %d", lo.Len())
			assert.Equal(t, m["key1"], lo.Field("key1"), "expected Field(\"key1\") to return %v, instead got %v",
				m["key1"], lo.Field("key1"))
			assert.Equal(t, m["key2"], lo.Field("key2"), "expected Field(\"key2\") to return %v, instead got %v",
				m["key2"], lo.Field("key2"))
		}
		m2 := map[string]*LazyValue{"key3": new(LazyValue)}
		codec.On("Decode", []byte(`foo`)).Return(m, nil)
		codec.On("Decode", []byte(`foo2`)).Return(m2, nil).Maybe() // should not be called anyway
		lv := &LazyValue{Payload: []byte(`foo`), Codec: codec}
		v, err := lv.Load()
		require.NoError(t, err)
		assert.Equal(t, TypeObject, v.Type(), "expected Type() return TypeObject, instead got %s", v.Type())
		assertEqual(t, v.Get())
		assertEqual(t, v.Object())

		// assert that this result is cached
		lv.Payload = []byte(`foo2`)
		v, err = lv.Load()
		require.NoError(t, err)
		assert.Equal(t, TypeObject, v.Type(), "expected Type() return TypeObject, instead got %s", v.Type())
		assertEqual(t, v.Get())
		assertEqual(t, v.Object())
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
				assert.Equal(t, "foo", v.Get(), "expected Get() return \"foo\", instead got %v", v.Get())
				assert.Equal(t, "foo", v.String(), "expected String() return \"foo\", instead got %q", v.String())
				return nil
			})
		}
		require.NoError(t, eg.Wait())
	})
}

func TestLazyObject(t *testing.T) {
	t.Run("Field with non-empty map", func(t *testing.T) {
		codec := NewMockCodec(t)
		m := map[string]*LazyValue{"key1": new(LazyValue), "key2": new(LazyValue)}
		codec.On("Decode", []byte(`foo`)).Return(m, nil)
		lz := &LazyValue{Payload: []byte(`foo`), Codec: codec}
		v, err := lz.Load()
		require.NoError(t, err)
		lo := v.Object()
		require.NotNil(t, lo)
		assert.Equal(t, len(m), lo.Len(), "expected object to have %d keys, instead got %d", len(m), lo.Len())
		assert.Equal(t, m["key1"], lo.Field("key1"), "expected Field(\"key1\") to return %v, instead got %v",
			m["key1"], lo.Field("key1"))
		assert.Equal(t, m["key2"], lo.Field("key2"), "expected Field(\"key2\") to return %v, instead got %v",
			m["key2"], lo.Field("key2"))
		assert.Nil(t, lo.Field("key3"), "expected Field(\"key3\") to return nil, instead got %v", lo.Field("key1"))
	})

	t.Run("Field with nil map", func(t *testing.T) {
		codec := NewMockCodec(t)
		m := map[string]*LazyValue(nil)
		codec.On("Decode", []byte(`foo`)).Return(m, nil)
		lz := &LazyValue{Payload: []byte(`foo`), Codec: codec}
		v, err := lz.Load()
		require.NoError(t, err)
		lo := v.Object()
		require.NotNil(t, lo)
		assert.Equal(t, len(m), lo.Len(), "expected object to have %d keys, instead got %d", len(m), lo.Len())
		assert.Nil(t, lo.Field("key1"), "expected Field(\"key1\") to return nil, instead got %v", lo.Field("key1"))
	})

	t.Run("if nil, Len returns 0", func(t *testing.T) {
		var lv *LazyObject
		assert.Equal(t, 0, lv.Len(), "expected Len() on nil LazyObject to return 0, instead got %d", lv.Len())
	})

	t.Run("if nil, Field returns nil", func(t *testing.T) {
		var lv *LazyObject
		assert.Nil(t, lv.Field("foo"), "expected Field(\"foo\") on nil LazyObject to return nil, instead got %v",
			lv.Field("foo"))
	})
}

func TestLazyArray(t *testing.T) {
	t.Run("Element with non-empty array", func(t *testing.T) {
		codec := NewMockCodec(t)
		a := []*LazyValue{new(LazyValue), new(LazyValue), new(LazyValue)}
		codec.On("Decode", []byte(`foo`)).Return(a, nil)
		lz := &LazyValue{Payload: []byte(`foo`), Codec: codec}
		v, err := lz.Load()
		require.NoError(t, err)
		la := v.Array()
		require.NotNil(t, la)
		assert.Equal(t, len(a), la.Len(), "expected array to have %d elements, instead got %d", len(a), la.Len())
		assert.Equal(t, a[0], la.Element(0), "expected Element(0) to return %v, instead got %v",
			a[0], la.Element(0))
		assert.Equal(t, a[1], la.Element(1), "expected Element(1) to return %v, instead got %v",
			a[1], la.Element(1))
		assert.Equal(t, a[2], la.Element(2), "expected Element(2) to return %v, instead got %v",
			a[2], la.Element(2))
		assert.Panics(t, func() {
			la.Element(-1)
		})
		assert.Panics(t, func() {
			la.Element(3)
		})

		// subarray
		la2 := la.SubArray(1, 3)
		assert.Equal(t, 2, la2.Len(), "expected array to have %d elements, instead got %d", 2, la2.Len())
		assert.Equal(t, a[1], la2.Element(0), "expected Element(0) to return %v, instead got %v",
			a[1], la2.Element(0))
		assert.Equal(t, a[2], la2.Element(1), "expected Element(1) to return %v, instead got %v",
			a[2], la2.Element(1))
		assert.Panics(t, func() {
			la.SubArray(0, 4)
		})
		assert.Panics(t, func() {
			la.SubArray(2, 1)
		})
	})

	t.Run("Element with nil array", func(t *testing.T) {
		codec := NewMockCodec(t)
		a := []*LazyValue(nil)
		codec.On("Decode", []byte(`foo`)).Return(a, nil)
		lz := &LazyValue{Payload: []byte(`foo`), Codec: codec}
		v, err := lz.Load()
		require.NoError(t, err)
		la := v.Array()
		require.NotNil(t, la)
		assert.Equal(t, len(a), la.Len(), "expected array to have %d elements, instead got %d", len(a), la.Len())
		assert.Panics(t, func() {
			la.Element(-1)
		})
		assert.Panics(t, func() {
			la.Element(0)
		})

		// subarray
		la2 := la.SubArray(0, 0)
		assert.Equal(t, 0, la2.Len(), "expected array to have %d elements, instead got %d", 0, la2.Len())
		assert.Panics(t, func() {
			la.SubArray(0, 1)
		})
		assert.Panics(t, func() {
			la.SubArray(2, 1)
		})
	})

	t.Run("if nil, SubArray handles", func(t *testing.T) {
		var lv *LazyArray
		assert.Panics(t, func() {
			lv.SubArray(0, 1)
		})
		assert.Nil(t, lv.SubArray(0, 0)) // special case
	})

	t.Run("if nil, Len returns 0", func(t *testing.T) {
		var lv *LazyArray
		assert.Equal(t, 0, lv.Len(), "expected Len() on nil LazyArray to return 0, instead got %d", lv.Len())
	})

	t.Run("if nil, Element panics", func(t *testing.T) {
		assert.Panics(t, func() {
			var lv *LazyArray
			lv.Element(0)
		})
	})
}

func TestValue(t *testing.T) {

}
