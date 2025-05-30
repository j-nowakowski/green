package lzval

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCodec(t *testing.T, codec Codec) {
	tests := []struct {
		name           string
		value          any
		expectOverride *any
	}{
		{name: "null", value: nil},
		{name: "boolean true", value: true},
		{name: "boolean false", value: false},
		{name: "number positive int", value: float64(1)},
		{name: "number negative int", value: float64(-1)},
		{name: "number zero", value: float64(0)},
		{name: "number positive float", value: -123.456},
		{name: "number negative float", value: -123.456},
		{name: "empty string", value: ""},
		{name: "string", value: "foo"},
		{name: "array", value: []any{float64(1), "two", false, nil, map[string]any{"foo": "bar"}, []any{"foo", "bar"}}},
		{name: "nil array", value: []any(nil), expectOverride: ptr[any](nil)},
		{name: "empty array", value: []any{}},
		{name: "object", value: map[string]any{"key1": "value1", "key2": float64(2), "key3": true, "key4": nil, "key5": map[string]any{"foo": "bar"}, "key6": []any{"foo", "bar"}}},
		{name: "nil object", value: map[string]any(nil), expectOverride: ptr[any](nil)},
		{name: "empty object", value: map[string]any{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := codec.Encode(tt.value)
			require.NoError(t, err)
			v, err := codec.Decode(b)
			require.NoError(t, err)
			if tt.expectOverride != nil {
				assertEqual(t, *tt.expectOverride, v)
			} else {
				assertEqual(t, tt.value, v)
			}
		})
	}
}

func assertEqual(t *testing.T, expected any, actual DecodeValue) {
	t.Helper()
	switch expect := expected.(type) {
	case nil, bool, float64, string:
		assert.Equal(t, expect, actual)
	case []any:
		assertEqualArray(t, expect, actual)
	case map[string]any:
		assertEqualObject(t, expect, actual)
	default:
		t.Fatalf("unexpected type %T", expect)
	}
}

func assertEqualLazy(t *testing.T, expected any, actual any) {
	t.Helper()
	switch expect := expected.(type) {
	case nil, bool, float64, string:
		assert.Equal(t, expect, actual)
	case []any:
		assertEqualLazyArray(t, expect, actual)
	case map[string]any:
		assertEqualLazyObject(t, expect, actual)
	default:
		t.Fatalf("unexpected type %T", expect)
	}
}

func assertEqualArray(t *testing.T, expected []any, actualAny any) {
	t.Helper()
	actual, ok := actualAny.(Array)
	require.True(t, ok, "expected Array type, got %T", actualAny)
	require.Len(t, actual, len(expected))
	for i, v := range expected {
		val, err := actual[i].Load()
		require.NoError(t, err)
		assertEqualLazy(t, v, val.Get())
	}
}

func assertEqualLazyArray(t *testing.T, expected []any, actualAny any) {
	t.Helper()
	actual, ok := actualAny.(*LazyArray)
	require.True(t, ok, "expected Array type, got %T", actualAny)
	require.Equal(t, actual.Len(), len(expected))
	for i, v := range expected {
		val, err := actual.Element(i).Load()
		require.NoError(t, err)
		assertEqualLazy(t, v, val.Get())
	}
}

func assertEqualObject(t *testing.T, expected map[string]any, actualAny any) {
	t.Helper()
	actual, ok := actualAny.(Object)
	require.True(t, ok, "expected Object type, got %T", actualAny)
	require.Len(t, actual, len(expected))
	for k, v := range expected {
		field, ok := actual[k]
		require.True(t, ok, "expected field %s to exist in object", k)
		val, err := field.Load()
		require.NoError(t, err)
		assertEqualLazy(t, v, val.Get())
	}
}

func assertEqualLazyObject(t *testing.T, expected map[string]any, actualAny any) {
	t.Helper()
	actual, ok := actualAny.(*LazyObject)
	require.True(t, ok, "expected Object type, got %T", actualAny)
	require.Equal(t, actual.Len(), len(expected))
	for k, v := range expected {
		val, err := actual.Field(k).Load()
		require.NoError(t, err)
		assertEqualLazy(t, v, val.Get())
	}
}

func ptr[T any](v T) *T {
	return &v
}
