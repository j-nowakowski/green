package lzval

// import (
// 	"testing"

// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/require"
// )

// func TestCodec(t *testing.T, codec Codec) {
// 	tests := []struct {
// 		name           string
// 		value          any
// 		expectOverride *any
// 	}{
// 		{name: "null", value: nil},
// 		{name: "boolean true", value: true},
// 		{name: "boolean false", value: false},
// 		{name: "number positive int", value: float64(1)},
// 		{name: "number negative int", value: float64(-1)},
// 		{name: "number zero", value: float64(0)},
// 		{name: "number positive float", value: -123.456},
// 		{name: "number negative float", value: -123.456},
// 		{name: "empty string", value: ""},
// 		{name: "string", value: "foo"},
// 		{name: "slice", value: []any{float64(1), "two", false, nil, map[string]any{"foo": "bar"}, []any{"foo", "bar"}}},
// 		{name: "nil slice", value: []any(nil), expectOverride: ptr[any](nil)},
// 		{name: "empty slice", value: []any{}},
// 		{name: "object", value: map[string]any{"key1": "value1", "key2": float64(2), "key3": true, "key4": nil, "key5": map[string]any{"foo": "bar"}, "key6": []any{"foo", "bar"}}},
// 		{name: "nil object", value: map[string]any(nil), expectOverride: ptr[any](nil)},
// 		{name: "empty object", value: map[string]any{}},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			b, err := codec.Encode(t.Context(), tt.value)
// 			require.NoError(t, err)
// 			v, err := codec.Decode(t.Context(), b)
// 			require.NoError(t, err)
// 			if tt.expectOverride != nil {
// 				assertEqual(t, *tt.expectOverride, v)
// 			} else {
// 				assertEqual(t, tt.value, v)
// 			}
// 		})
// 	}
// }

// func assertEqual(t *testing.T, expected any, actual DecodeValue) {
// 	t.Helper()
// 	switch expect := expected.(type) {
// 	case nil, bool, float64, string:
// 		assert.Equal(t, expect, actual)
// 	case []any:
// 		assertEqualSlice(t, expect, actual)
// 	case map[string]any:
// 		assertEqualMap(t, expect, actual)
// 	default:
// 		t.Fatalf("unexpected type %T", expect)
// 	}
// }

// func assertEqualLazy(t *testing.T, expected any, actual any) {
// 	t.Helper()
// 	switch expect := expected.(type) {
// 	case nil, bool, float64, string:
// 		assert.Equal(t, expect, actual)
// 	case []any:
// 		assertEqualLazySlice(t, expect, actual)
// 	case map[string]any:
// 		assertEqualLazyMap(t, expect, actual)
// 	default:
// 		t.Fatalf("unexpected type %T", expect)
// 	}
// }

// func assertEqualSlice(t *testing.T, expected []any, actualAny any) {
// 	t.Helper()
// 	actual, ok := actualAny.(LoadableSlice)
// 	require.True(t, ok, "expected Slice type, got %T", actualAny)
// 	require.Len(t, actual, len(expected))
// 	for i, v := range expected {
// 		val, err := actual[i].Load(t.Context())
// 		require.NoError(t, err)
// 		assertEqualLazy(t, v, val)
// 	}
// }

// func assertEqualLazySlice(t *testing.T, expected []any, actualAny any) {
// 	t.Helper()
// 	actual, ok := actualAny.(*ImmutableSlice)
// 	require.True(t, ok, "expected Slice type, got %T", actualAny)
// 	require.Equal(t, actual.Len(), len(expected))
// 	for i, v := range expected {
// 		val, err := actual.At(i).Load(t.Context())
// 		require.NoError(t, err)
// 		assertEqualLazy(t, v, val)
// 	}
// }

// func assertEqualMap(t *testing.T, expected map[string]any, actualAny any) {
// 	t.Helper()
// 	actual, ok := actualAny.(LoadableMap)
// 	require.True(t, ok, "expected Map type, got %T", actualAny)
// 	require.Len(t, actual, len(expected))
// 	for k, v := range expected {
// 		field, ok := actual[k]
// 		require.True(t, ok, "expected field %s to exist in object", k)
// 		val, err := field.Load(t.Context())
// 		require.NoError(t, err)
// 		assertEqualLazy(t, v, val)
// 	}
// }

// func assertEqualLazyMap(t *testing.T, expected map[string]any, actualAny any) {
// 	t.Helper()
// 	actual, ok := actualAny.(*ImmutableMap)
// 	require.True(t, ok, "expected Map type, got %T", actualAny)
// 	require.Equal(t, actual.Len(), len(expected))
// 	for k, v := range expected {
// 		got, ok := actual.Get(k)
// 		require.True(t, ok, "expected field %s to exist in object", k)
// 		val, err := got.Load(t.Context())
// 		require.NoError(t, err)
// 		assertEqualLazy(t, v, val)
// 	}
// }

// func ptr[T any](v T) *T {
// 	return &v
// }
