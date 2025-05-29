package lzval

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCodec(t *testing.T, codec Codec) {
	tests := []struct {
		name  string
		value any
	}{
		{"null", nil},
		{"boolean true", true},
		{"boolean false", false},
		{"number positive int", float64(1)},
		{"number negative int", float64(-1)},
		{"number zero", float64(0)},
		{"number positive float", -123.456},
		{"number negative float", -123.456},
		{"empty string", ""},
		{"string", "foo"},
		{"array", []any{float64(1), "two", false, nil, map[string]any{"foo": "bar"}, []any{"foo", "bar"}}},
		{"nil array", []any(nil)},
		{"empty array", []any{}},
		{"object", map[string]any{"key1": "value1", "key2": float64(2), "key3": true, "key4": nil, "key5": map[string]any{"foo": "bar"}, "key6": []any{"foo", "bar"}}},
		{"nil object", map[string]any(nil)},
		{"empty object", map[string]any{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := codec.Encode(tt.value)
			require.NoError(t, err)
			v, err := codec.Decode(b)
			require.NoError(t, err)
			assertEqual(t, tt.value, v)
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

func assertEqualArray(t *testing.T, expected []any, actualAny any) {
	t.Helper()
	actual, ok := actualAny.(Array)
	require.True(t, ok, "expected Array type, got %T", actualAny)
	require.Len(t, actual, len(expected))
	for i, v := range expected {
		node, err := actual[i].Resolve()
		require.NoError(t, err)
		assertEqual(t, v, node.Value())
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
		node, err := field.Resolve()
		require.NoError(t, err)
		assertEqual(t, v, node.Value())
	}
}
