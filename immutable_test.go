package lzval

import (
	"maps"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImmutable(t *testing.T) {
	t.Run("ImmutableMap", func(t *testing.T) {
		key2Map := map[string]any{
			"nestedKey1": 42,
		}
		key3Slice := []any{1}
		source := map[string]any{
			"key1": "value1",
			"key2": key2Map,
			"key3": key3Slice,
		}
		sourceOriginal := deepCopy(source)

		sourceImmutable := NewImmutableMap(source)

		// assert length 3
		assert.Equal(t, 3, sourceImmutable.Len())

		// assert key1 has same string value
		got1, ok := sourceImmutable.Get("key1")
		require.True(t, ok)
		assert.Equal(t, "value1", got1)

		// assert key2 is a map with correct nested value
		got2, ok := sourceImmutable.Get("key2")
		require.True(t, ok)
		got2Map, ok := got2.(*ImmutableMap)
		require.True(t, ok)
		require.Equal(t, 1, got2Map.Len())
		nestedValue, ok := got2Map.Get("nestedKey1")
		require.True(t, ok)
		assert.Equal(t, 42, nestedValue)

		// assert that key2 map reference is the same across calls
		got2Again, ok := sourceImmutable.Get("key2")
		require.True(t, ok)
		got2AgainMap, ok := got2Again.(*ImmutableMap)
		require.True(t, ok)
		assert.True(t, got2Map == got2AgainMap, "Get should return same ImmutableMap instance")

		// assert that key3 is a slice with correct nested value
		got3, ok := sourceImmutable.Get("key3")
		require.True(t, ok)
		got3Slice, ok := got3.(*ImmutableSlice)
		require.True(t, ok)
		assert.Equal(t, 1, got3Slice.Len())
		val0 := got3Slice.At(0)
		assert.Equal(t, 1, val0)

		// assert that key3 slice reference is the same across calls
		got3Again, ok := sourceImmutable.Get("key3")
		require.True(t, ok)
		got3SliceAgain, ok := got3Again.(*ImmutableSlice)
		require.True(t, ok)
		assert.True(t, got3Slice == got3SliceAgain, "Get should return same ImmutableSlice instance")

		// test All iterator
		collected := maps.Collect(sourceImmutable.All())
		assert.Equal(t, 3, len(collected))

		// key1 from All should match original
		assert.Equal(t, "value1", collected["key1"])

		// key2 from All should match original
		collected2, ok := collected["key2"].(*ImmutableMap)
		require.True(t, ok)
		assert.True(t, collected2 == got2, "Get and All should return same ImmutableMap instance")

		// key3 from All should match original
		collected3, ok := collected["key3"].(*ImmutableSlice)
		require.True(t, ok)
		assert.True(t, collected3 == got3Slice, "Get and All should return same ImmutableSlice instance")

		// All order should be non-deterministic. Run 100 times and find that at least one time is different
		var firstKey string
		for k := range sourceImmutable.All() {
			firstKey = k
			break
		}
		success := false
		for range 100 {
			var firstKeyAgain string
			for k := range sourceImmutable.All() {
				firstKeyAgain = k
				break
			}
			if firstKeyAgain != firstKey {
				success = true
				break
			}
		}
		assert.True(t, success, "All() order should be non-deterministic")

		// test Export
		sourceExported := sourceImmutable.Export()
		assert.Equal(t, sourceOriginal, sourceExported)
		sourceExportedCopy := deepCopy(sourceExported)

		// source should not have been mutated by previous operations
		assert.Equal(t, sourceOriginal, source)

		// Export should return a deep copy
		source["foo"] = "bar"
		key2Map["foo"] = "bar"
		key3Slice[0] = "bar"
		assert.Equal(t, sourceExported, sourceExportedCopy)
	})

	t.Run("ImmutableSlice", func(t *testing.T) {
		nestedMap := map[string]any{
			"nestedKey1": 42,
		}
		nestedSlice := []any{1}
		source := []any{
			"value1",
			nestedMap,
			nestedSlice,
		}
		sourceOriginal := deepCopy(source)

		sourceImmutable := NewImmutableSlice(source)

		// assert length 3
		assert.Equal(t, 3, sourceImmutable.Len())

		// assert index 0 has same string value
		got0 := sourceImmutable.At(0)
		assert.Equal(t, "value1", got0)

		// assert index 1 is a map with correct nested value
		got1 := sourceImmutable.At(1)
		got1Map, ok := got1.(*ImmutableMap)
		require.True(t, ok)
		require.Equal(t, 1, got1Map.Len())
		nestedValue, ok := got1Map.Get("nestedKey1")
		require.True(t, ok)
		assert.Equal(t, 42, nestedValue)

		// assert that index 1 map reference is the same across calls
		got1Again := sourceImmutable.At(1)
		got1AgainMap, ok := got1Again.(*ImmutableMap)
		require.True(t, ok)
		assert.True(t, got1Map == got1AgainMap, "At should return same ImmutableMap instance")

		// assert index 2 is a slice with correct nested value
		got2 := sourceImmutable.At(2)
		got2Slice, ok := got2.(*ImmutableSlice)
		require.True(t, ok)
		assert.Equal(t, 1, got2Slice.Len())
		val0 := got2Slice.At(0)
		assert.Equal(t, 1, val0)

		// assert that index 2 slice reference is the same across calls
		got2Again := sourceImmutable.At(2)
		got2SliceAgain, ok := got2Again.(*ImmutableSlice)
		require.True(t, ok)
		assert.True(t, got2Slice == got2SliceAgain, "At should return same ImmutableSlice instance")

		// test All iterator (preserves order for slices)
		collected := make([]any, 0, sourceImmutable.Len())
		for _, v := range sourceImmutable.All() {
			collected = append(collected, v)
		}
		require.Equal(t, 3, len(collected))
		assert.Equal(t, "value1", collected[0])
		all1, ok := collected[1].(*ImmutableMap)
		require.True(t, ok)
		assert.True(t, all1 == got1Map, "At and All should return same ImmutableMap instance")
		all2, ok := collected[2].(*ImmutableSlice)
		require.True(t, ok)
		assert.True(t, all2 == got2Slice, "At and All should return same ImmutableSlice instance")

		// test Export
		sourceExported := sourceImmutable.Export()
		assert.Equal(t, sourceOriginal, sourceExported)
		sourceExportedCopy := deepCopy(sourceExported)

		// source should not have been mutated by previous operations
		assert.Equal(t, sourceOriginal, source)

		// Export should return a deep copy
		source[0] = "foo"
		nestedMap["foo"] = "bar"
		nestedSlice[0] = "bar"
		assert.Equal(t, sourceExported, sourceExportedCopy)

		// SubSlice
		map1 := map[string]any{"a": 1}
		slice1 := []any{2, 3}
		src := []any{
			"zero",
			map1,
			slice1,
			"three",
			map[string]any{"b": 4},
		}
		srcCopy := deepCopy(src).([]any)
		ims := NewImmutableSlice(src)

		// take sub-slice [1:4] (Go-style: start inclusive, end exclusive)
		sub := ims.SubSlice(1, 4)
		require.Equal(t, 3, sub.Len())

		// check contents and types
		v0 := sub.At(0)
		m, ok := v0.(*ImmutableMap)
		require.True(t, ok)
		valA, ok := m.Get("a")
		require.True(t, ok)
		assert.Equal(t, 1, valA)

		v1 := sub.At(1)
		s1, ok := v1.(*ImmutableSlice)
		require.True(t, ok)
		assert.Equal(t, 2, s1.Len())
		assert.Equal(t, 2, s1.At(0))
		assert.Equal(t, 3, s1.At(1))

		v2 := sub.At(2)
		assert.Equal(t, "three", v2)

		// pointer equality: elements returned from SubSlice should be the same instances
		origV0 := ims.At(1)
		origMap, ok := origV0.(*ImmutableMap)
		require.True(t, ok)
		assert.True(t, origMap == m, "SubSlice map should be same instance as original")

		origV1 := ims.At(2)
		origSlice, ok := origV1.(*ImmutableSlice)
		require.True(t, ok)
		assert.True(t, origSlice == s1, "SubSlice slice should be same instance as original")

		// Exported sub-slice should match the expected raw slice
		expected := deepCopy(srcCopy[1:4]).([]any)
		exported := sub.Export()
		assert.Equal(t, expected, exported)

		// original should remain unchanged after operations
		assert.Equal(t, srcCopy, src)

		// ensure Export returns a deep copy (mutate original and nested values)
		map1["a"] = 100
		slice1[0] = 99
		assert.Equal(t, expected, exported)
	})

	t.Run("EqualImmutableValues_ImmutableMap", func(t *testing.T) {
		key2Map := map[string]any{
			"nestedKey1": 42,
		}
		key3Slice := []any{1}
		m1 := map[string]any{
			"key1": "value1",
			"key2": key2Map,
			"key3": key3Slice,
		}
		m1Copy := deepCopy(m1).(map[string]any)
		im1 := NewImmutableMap(m1)
		im1Copy := NewImmutableMap(m1Copy)

		assert.True(t, EqualImmutableValues(im1, im1))
		assert.True(t, EqualImmutableValues(im1, im1Copy), "ImmutableMaps with same content should be equal")

		assert.False(t, EqualImmutableValues(im1, "foo"))
		assert.False(t, EqualImmutableValues(im1, NewImmutableSlice([]any{1})))

		assert.False(t, EqualImmutableValues(im1, NewImmutableMap(map[string]any{
			"key1": "different",
			"key2": key2Map,
			"key3": key3Slice,
		})))

		assert.False(t, EqualImmutableValues(im1, NewImmutableMap(map[string]any{
			"key1": "value1",
			"key2": key2Map,
			"key3": key3Slice,
			"foo":  "bar",
		})))

		assert.False(t, EqualImmutableValues(im1, NewImmutableMap(map[string]any{
			"key1": "value1",
			"key2": map[string]any{
				"nestedKey1": 42,
				"nestedKey2": 100,
			},
			"key3": key3Slice,
		})))

		assert.False(t, EqualImmutableValues(im1, NewImmutableMap(map[string]any{
			"key1": "value1",
			"key2": map[string]any{
				"nestedKey1": 43,
			},
			"key3": key3Slice,
		})))

		assert.False(t, EqualImmutableValues(im1, NewImmutableMap(map[string]any{
			"key1": "value1",
			"key2": map[string]any{
				"nestedKey1": 42,
			},
			"key3": []any{1, 2},
		})))

		assert.False(t, EqualImmutableValues(im1, NewImmutableMap(map[string]any{
			"key1": "value1",
			"key2": map[string]any{
				"nestedKey1": 42,
			},
			"key3": []any{2},
		})))
	})

	t.Run("EqualImmutableValues_ImmutableSlice", func(t *testing.T) {
		nestedMap := map[string]any{
			"nestedKey1": 42,
		}
		nestedSlice := []any{1}
		s1 := []any{
			"value1",
			nestedMap,
			nestedSlice,
		}
		s1Copy := deepCopy(s1).([]any)

		is1 := NewImmutableSlice(s1)
		is1Copy := NewImmutableSlice(s1Copy)

		assert.True(t, EqualImmutableValues(is1, is1))
		assert.True(t, EqualImmutableValues(is1, is1Copy), "ImmutableSlices with same content should be equal")

		assert.False(t, EqualImmutableValues(is1, "foo"))
		assert.False(t, EqualImmutableValues(is1, NewImmutableMap(map[string]any{"foo": "bar"})))

		// different first element
		assert.False(t, EqualImmutableValues(is1, NewImmutableSlice([]any{
			"different",
			nestedMap,
			nestedSlice,
		})))

		// extra element
		assert.False(t, EqualImmutableValues(is1, NewImmutableSlice([]any{
			"value1",
			nestedMap,
			nestedSlice,
			"foo",
		})))

		// nested map has extra key
		assert.False(t, EqualImmutableValues(is1, NewImmutableSlice([]any{
			"value1",
			map[string]any{
				"nestedKey1": 42,
				"nestedKey2": 100,
			},
			nestedSlice,
		})))

		// nested map value changed
		assert.False(t, EqualImmutableValues(is1, NewImmutableSlice([]any{
			"value1",
			map[string]any{
				"nestedKey1": 43,
			},
			nestedSlice,
		})))

		// nested slice longer
		assert.False(t, EqualImmutableValues(is1, NewImmutableSlice([]any{
			"value1",
			nestedMap,
			[]any{1, 2},
		})))

		// nested slice different element
		assert.False(t, EqualImmutableValues(is1, NewImmutableSlice([]any{
			"value1",
			nestedMap,
			[]any{2},
		})))
	})
}

func deepCopy(a any) any {
	switch a := a.(type) {
	case map[string]any:
		m := make(map[string]any, len(a))
		for k, v := range a {
			m[k] = deepCopy(v)
		}
		return m
	case []any:
		s := make([]any, len(a))
		for i, v := range a {
			s[i] = deepCopy(v)
		}
		return s
	default:
		return a
	}
}
