package lzval

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMutable(t *testing.T) {
	t.Run("MutableMap", func(t *testing.T) {
		t.Run("basic mutations on map and nested", func(t *testing.T) {
			nm := map[string]any{
				"nk1": "nv1",
				"nk2": "nv2",
			}
			ns := []any{"ne1", "ne2"}
			m := map[string]any{
				"k1": "v1",
				"k2": nm,
				"k3": ns,
				"k4": "v4",
			}
			mOriginal := deepCopy(m).(map[string]any)

			im := NewImmutableMap(m)

			mut1 := im.Mutable()
			mut2 := im.Mutable()
			mut1Clone := mut1.Clone()
			assert.Equal(t, 4, mut1.Len())

			// check k1 before mutating
			require.True(t, mut1.Has("k1"))
			v1, ok := mut1.Get("k1")
			require.True(t, ok)
			assert.Equal(t, "v1", v1)

			// check k2 before mutating
			require.True(t, mut1.Has("k2"))
			v2, ok := mut1.Get("k2")
			require.True(t, ok)
			v2Map, ok := v2.(*Map)
			require.True(t, ok)
			assert.Equal(t, 2, v2Map.Len())
			require.True(t, v2Map.Has("nk1"))
			nv1, ok := v2Map.Get("nk1")
			require.True(t, ok)
			assert.Equal(t, "nv1", nv1)
			require.True(t, v2Map.Has("nk2"))
			nv2, ok := v2Map.Get("nk2")
			require.True(t, ok)
			assert.Equal(t, "nv2", nv2)

			// check k3 before mutating
			require.True(t, mut1.Has("k3"))
			v3, ok := mut1.Get("k3")
			require.True(t, ok)
			v3Slice, ok := v3.(*Slice)
			require.True(t, ok)
			require.Equal(t, 2, v3Slice.Len())
			ne1 := v3Slice.At(0)
			assert.Equal(t, "ne1", ne1)
			ne2 := v3Slice.At(1)
			assert.Equal(t, "ne2", ne2)

			require.False(t, mut1.Has("k5"))
			_, ok = mut1.Get("k5")
			require.False(t, ok)

			mut1FromAll := map[string]any{}
			for k, v := range mut1.All() {
				mut1FromAll[k] = ExportValue(v)
			}
			assert.Equal(t, m, mut1FromAll)
			assert.Equal(t, mOriginal, mut1FromAll)

			// delete non-existent key
			mut1.Delete("non-existent-key")
			assert.Equal(t, 4, mut1.Len())

			// mutate all values
			mut1.Set("k1", "v1-modified")
			mut1.Delete("k4")
			assert.Equal(t, 3, mut1.Len())
			mut1.Set("k5", "v5-new")
			v2Map.Set("nk1", "nv1-modified")
			v2Map.Set("nk3", "nv3-new")
			v3Slice.Set(0, "ne1-modified")
			v3Slice.Push("ne3-new")
			assert.Equal(t, 4, mut1.Len())

			expectMut1 := map[string]any{
				"k1": "v1-modified",
				"k2": map[string]any{
					"nk1": "nv1-modified",
					"nk2": "nv2",
					"nk3": "nv3-new",
				},
				"k3": []any{"ne1-modified", "ne2", "ne3-new"},
				"k5": "v5-new",
			}
			assert.Equal(t, expectMut1, mut1.Export())
			assert.True(t, EqualValues(NewImmutableMap(expectMut1).Mutable(), mut1))
			assert.Equal(t, mOriginal, m)

			require.True(t, mut1.Has("k1"))
			v1After, ok := mut1.Get("k1")
			require.True(t, ok)
			assert.Equal(t, "v1-modified", v1After)

			require.True(t, mut1.Has("k2"))
			v2After, ok := mut1.Get("k2")
			require.True(t, ok)
			v2MapAfter, ok := v2After.(*Map)
			require.True(t, ok)
			require.True(t, v2MapAfter.Has("nk1"))
			nv1After, ok := v2MapAfter.Get("nk1")
			require.True(t, ok)
			assert.Equal(t, "nv1-modified", nv1After)

			require.True(t, v2MapAfter.Has("nk3"))
			v3After, ok := mut1.Get("k3")
			require.True(t, ok)
			v3SliceAfter, ok := v3After.(*Slice)
			require.True(t, ok)
			require.Equal(t, 3, v3SliceAfter.Len())
			ne1After := v3SliceAfter.At(0)
			assert.Equal(t, "ne1-modified", ne1After)

			require.True(t, mut1.Has("k5"))
			v5After, ok := mut1.Get("k5")
			require.True(t, ok)
			assert.Equal(t, "v5-new", v5After)

			mut1AfterFromAll := map[string]any{}
			for k, v := range mut1.All() {
				mut1AfterFromAll[k] = ExportValue(v)
			}
			assert.Equal(t, expectMut1, mut1AfterFromAll)

			assert.Equal(t, m, mut2.Export())
			assert.True(t, EqualValues(NewImmutableMap(m).Mutable(), mut2))

			expectMut1Clone := map[string]any{
				"k1": "v1",
				"k2": map[string]any{
					"nk1": "nv1-modified",
					"nk2": "nv2",
					"nk3": "nv3-new",
				},
				"k3": []any{"ne1-modified", "ne2", "ne3-new"},
				"k4": "v4",
			}
			assert.Equal(t, expectMut1Clone, mut1Clone.Export())
			assert.True(t, EqualValues(NewImmutableMap(expectMut1Clone).Mutable(), mut1Clone))

			assert.Equal(t, mOriginal, im.Export())
		})

		t.Run("instance management", func(t *testing.T) {
			nmm := NewImmutableMap(map[string]any{}).Mutable()
			nms := NewImmutableSlice([]any{}).Mutable()
			m := map[string]any{
				"k1": map[string]any{"nk1": "nv1"},
				"k2": []any{"ne1"},
				"k3": nmm,
				"k4": nms,
			}

			im := NewImmutableMap(m)
			mut := im.Mutable()

			v1, ok := mut.Get("k1")
			require.True(t, ok)
			v1Map, ok := v1.(*Map)
			require.True(t, ok, "%T: %v", v1, v1)
			v1Again, ok := mut.Get("k1")
			require.True(t, ok)
			v1MapAgain, ok := v1Again.(*Map)
			require.True(t, ok)
			assert.Same(t, v1Map, v1MapAgain)

			v2, ok := mut.Get("k2")
			require.True(t, ok)
			v2Slice, ok := v2.(*Slice)
			require.True(t, ok, "%T: %v", v2, v2)
			v2Again, ok := mut.Get("k2")
			require.True(t, ok)
			v2SliceAgain, ok := v2Again.(*Slice)
			require.True(t, ok)
			assert.Same(t, v2Slice, v2SliceAgain)

			v3, ok := mut.Get("k3")
			require.True(t, ok)
			v3Map, ok := v3.(*Map)
			require.True(t, ok, "%T: %v", v3, v3)
			assert.Same(t, nmm, v3Map)

			v4, ok := mut.Get("k4")
			require.True(t, ok)
			v4Slice, ok := v4.(*Slice)
			require.True(t, ok, "%T: %v", v4, v4)
			assert.Same(t, nms, v4Slice)
		})

		t.Run("Swapping mutability", func(t *testing.T) {
			m := map[string]any{
				"k1": "v1",
				"k2": map[string]any{"nk1": "nv1"},
				"k3": []any{"ne1"},
			}

			im := NewImmutableMap(m)

			mut := im.Mutable()
			mut.Set("k1", "v1-modified")

			im2 := mut.Immutable()

			mut.Set("k1", "v1-modified-again")

			require.True(t, im2.Has("k1"))
			v1, ok := im2.Get("k1")
			require.True(t, ok)
			assert.Equal(t, "v1-modified", v1)

			require.True(t, mut.Has("k1"))
			v1Mut, ok := mut.Get("k1")
			require.True(t, ok)
			assert.Equal(t, "v1-modified-again", v1Mut)

			mut3 := im2.Mutable()
			v2, ok := mut3.Get("k2")
			require.True(t, ok)
			v2Map, ok := v2.(*Map)
			require.True(t, ok)
			v2Map.Set("nk1", "nv1-modified")
			v3, ok := mut3.Get("k3")
			require.True(t, ok)
			v3Slice, ok := v3.(*Slice)
			require.True(t, ok)
			v3Slice.Set(0, "ne1-modified")

			im4 := mut3.Immutable()
			got, ok := im4.Get("k2")
			require.True(t, ok)
			_, ok = got.(*ImmutableMap)
			require.True(t, ok)
			gotSlice, ok := im4.Get("k3")
			require.True(t, ok)
			_, ok = gotSlice.(*ImmutableSlice)
			require.True(t, ok)

			expectMut3 := map[string]any{
				"k1": "v1-modified",
				"k2": map[string]any{"nk1": "nv1-modified"},
				"k3": []any{"ne1-modified"},
			}
			assert.Equal(t, expectMut3, mut3.Export())

			expectImm2 := map[string]any{
				"k1": "v1-modified",
				"k2": map[string]any{"nk1": "nv1"},
				"k3": []any{"ne1"},
			}
			assert.Equal(t, expectImm2, im2.Export())

			expectImm := map[string]any{
				"k1": "v1",
				"k2": map[string]any{"nk1": "nv1"},
				"k3": []any{"ne1"},
			}
			assert.Equal(t, expectImm, im.Export())
		})

		t.Run("map is dirty from nested map mutation", func(t *testing.T) {
			m := map[string]any{
				"k2": map[string]any{"nk1": "nv1"},
			}
			im := NewImmutableMap(m)
			mut := im.Mutable()
			v2, ok := mut.Get("k2")
			require.True(t, ok)
			v2Map, ok := v2.(*Map)
			require.True(t, ok)
			v2Map.Set("nk1", "nv1-modified")

			im2 := mut.Immutable()

			expectIm2 := map[string]any{
				"k2": map[string]any{"nk1": "nv1-modified"},
			}
			assert.Equal(t, expectIm2, im2.Export())
		})

		t.Run("map is dirty from nested slice mutation", func(t *testing.T) {
			m := map[string]any{
				"k2": []any{"ne1"},
			}
			im := NewImmutableMap(m)
			mut := im.Mutable()
			v2, ok := mut.Get("k2")
			require.True(t, ok)
			v2Slice, ok := v2.(*Slice)
			require.True(t, ok)
			v2Slice.Set(0, "ne1-modified")

			im2 := mut.Immutable()

			expectIm2 := map[string]any{
				"k2": []any{"ne1-modified"},
			}
			assert.Equal(t, expectIm2, im2.Export())
		})
	})

	t.Run("MutableSlice", func(t *testing.T) {
		t.Run("basic mutations on slice and nested", func(t *testing.T) {
			nm := map[string]any{
				"nk1": "nv1",
				"nk2": "nv2",
			}
			ns := []any{"ne1", "ne2"}
			s := []any{
				"e1",
				nm,
				ns,
				"e4",
			}
			sOriginal := deepCopy(s).([]any)

			is := NewImmutableSlice(s)

			mut1 := is.Mutable()
			mut2 := is.Mutable()
			mut1Clone := mut1.Clone()
			assert.Equal(t, 4, mut1.Len())

			// check index 0 before mutating
			v0 := mut1.At(0)
			assert.Equal(t, "e1", v0)

			// check index 1 before mutating
			v1 := mut1.At(1)
			v1Map, ok := v1.(*Map)
			require.True(t, ok)
			assert.Equal(t, 2, v1Map.Len())
			require.True(t, v1Map.Has("nk1"))
			nv1, ok := v1Map.Get("nk1")
			require.True(t, ok)
			assert.Equal(t, "nv1", nv1)
			require.True(t, v1Map.Has("nk2"))
			nv2, ok := v1Map.Get("nk2")
			require.True(t, ok)
			assert.Equal(t, "nv2", nv2)

			// check index 2 before mutating
			v2 := mut1.At(2)
			v2Slice, ok := v2.(*Slice)
			require.True(t, ok)
			require.Equal(t, 2, v2Slice.Len())
			ne1 := v2Slice.At(0)
			assert.Equal(t, "ne1", ne1)
			ne2 := v2Slice.At(1)
			assert.Equal(t, "ne2", ne2)

			// check index 3 before mutating
			v3 := mut1.At(3)
			assert.Equal(t, "e4", v3)

			mut1FromAll := make([]any, mut1.Len())
			for i, v := range mut1.All() {
				mut1FromAll[i] = ExportValue(v)
			}
			assert.Equal(t, s, mut1FromAll)
			assert.Equal(t, sOriginal, mut1FromAll)

			// mutate all values
			mut1.Set(0, "e1-modified")
			v1Map.Set("nk1", "nv1-modified")
			v1Map.Set("nk3", "nv3-new")
			v2Slice.Set(0, "ne1-modified")
			v2Slice.Push("ne3-new")
			mut1.Push("e5-new")
			mut1.PushFront("eo-new")
			assert.Equal(t, 6, mut1.Len())

			expectMut1 := []any{
				"eo-new",
				"e1-modified",
				map[string]any{
					"nk1": "nv1-modified",
					"nk2": "nv2",
					"nk3": "nv3-new",
				},
				[]any{"ne1-modified", "ne2", "ne3-new"},
				"e4",
				"e5-new",
			}
			assert.Equal(t, expectMut1, mut1.Export())
			assert.True(t, EqualValues(NewImmutableSlice(expectMut1).Mutable(), mut1))
			assert.Equal(t, sOriginal, s)

			e0 := mut1.At(0)
			assert.Equal(t, "eo-new", e0)
			v1After := mut1.At(1)
			assert.Equal(t, "e1-modified", v1After)
			v2After := mut1.At(2)
			v2MapAfter, ok := v2After.(*Map)
			require.True(t, ok)
			require.True(t, v2MapAfter.Has("nk1"))
			nv1After, ok := v2MapAfter.Get("nk1")
			require.True(t, ok)
			assert.Equal(t, "nv1-modified", nv1After)
			v3After := mut1.At(3)
			v3SliceAfter, ok := v3After.(*Slice)
			require.True(t, ok)
			require.Equal(t, 3, v3SliceAfter.Len())
			ne1After := v3SliceAfter.At(0)
			assert.Equal(t, "ne1-modified", ne1After)
			v4After := mut1.At(4)
			assert.Equal(t, "e4", v4After)

			mut1AfterFromAll := make([]any, mut1.Len())
			for i, v := range mut1.All() {
				mut1AfterFromAll[i] = ExportValue(v)
			}
			assert.Equal(t, expectMut1, mut1AfterFromAll)

			assert.Equal(t, s, mut2.Export())
			assert.True(t, EqualValues(NewImmutableSlice(s).Mutable(), mut2))
			expectMut1Clone := []any{
				"e1",
				map[string]any{
					"nk1": "nv1-modified",
					"nk2": "nv2",
					"nk3": "nv3-new",
				},
				[]any{"ne1-modified", "ne2", "ne3-new"},
				"e4",
			}
			assert.Equal(t, expectMut1Clone, mut1Clone.Export())
			assert.True(t, EqualValues(NewImmutableSlice(expectMut1Clone).Mutable(), mut1Clone))

			assert.Equal(t, sOriginal, is.Export())
		})

		t.Run("instance management", func(t *testing.T) {
			nmm := NewImmutableMap(map[string]any{}).Mutable()
			nms := NewImmutableSlice([]any{}).Mutable()
			s := []any{
				map[string]any{"nk1": "nv1"},
				[]any{"ne1"},
				nmm,
				nms,
			}

			is := NewImmutableSlice(s)
			mut := is.Mutable()

			v1 := mut.At(0)
			v1Map, ok := v1.(*Map)
			require.True(t, ok)
			v1Again := mut.At(0)
			v1MapAgain, ok := v1Again.(*Map)
			require.True(t, ok)
			assert.Same(t, v1Map, v1MapAgain)

			v2 := mut.At(1)
			v2Slice, ok := v2.(*Slice)
			require.True(t, ok)
			v2Again := mut.At(1)
			v2SliceAgain, ok := v2Again.(*Slice)
			require.True(t, ok)
			assert.Same(t, v2Slice, v2SliceAgain)
			v3 := mut.At(2)
			v3Map, ok := v3.(*Map)
			require.True(t, ok)
			assert.Same(t, nmm, v3Map)
			v4 := mut.At(3)
			v4Slice, ok := v4.(*Slice)
			require.True(t, ok)
			assert.Same(t, nms, v4Slice)
		})

		t.Run("Swapping mutability", func(t *testing.T) {
			s := []any{
				"e1",
				map[string]any{"nk1": "nv1"},
				[]any{"ne1"},
			}

			is := NewImmutableSlice(s)

			mut := is.Mutable()
			mut.Set(0, "e1-modified")

			is2 := mut.Immutable()

			mut.Set(0, "e1-modified-again")

			v1 := is2.At(0)
			assert.Equal(t, "e1-modified", v1)

			v1Mut := mut.At(0)
			assert.Equal(t, "e1-modified-again", v1Mut)

			mut3 := is2.Mutable()
			v2 := mut3.At(1)
			v2Map, ok := v2.(*Map)
			require.True(t, ok)
			v2Map.Set("nk1", "nv1-modified")
			v3 := mut3.At(2)
			v3Slice, ok := v3.(*Slice)
			require.True(t, ok)
			v3Slice.Set(0, "ne1-modified")

			is4 := mut3.Immutable()
			got := is4.At(1)
			_, ok = got.(*ImmutableMap)
			require.True(t, ok)
			gotSlice := is4.At(2)
			_, ok = gotSlice.(*ImmutableSlice)
			require.True(t, ok)

			expectMut3 := []any{
				"e1-modified",
				map[string]any{"nk1": "nv1-modified"},
				[]any{"ne1-modified"},
			}
			assert.Equal(t, expectMut3, mut3.Export())

			expectIs2 := []any{
				"e1-modified",
				map[string]any{"nk1": "nv1"},
				[]any{"ne1"},
			}
			assert.Equal(t, expectIs2, is2.Export())

			expectIs := []any{
				"e1",
				map[string]any{"nk1": "nv1"},
				[]any{"ne1"},
			}
			assert.Equal(t, expectIs, is.Export())
		})

		t.Run("slice is dirty from nested map mutation", func(t *testing.T) {
			s := []any{
				map[string]any{"nk1": "nv1"},
			}
			is := NewImmutableSlice(s)
			mut := is.Mutable()
			v1 := mut.At(0)
			v1Map, ok := v1.(*Map)
			require.True(t, ok)
			v1Map.Set("nk1", "nv1-modified")

			is2 := mut.Immutable()

			expectIs2 := []any{
				map[string]any{"nk1": "nv1-modified"},
			}
			assert.Equal(t, expectIs2, is2.Export())
		})

		t.Run("slice is dirty from nested slice mutation", func(t *testing.T) {
			s := []any{
				[]any{"ne1"},
			}
			is := NewImmutableSlice(s)
			mut := is.Mutable()
			v1 := mut.At(0)
			v1Slice, ok := v1.(*Slice)
			require.True(t, ok)
			v1Slice.Set(0, "ne1-modified")

			is2 := mut.Immutable()

			expectIs2 := []any{
				[]any{"ne1-modified"},
			}
			assert.Equal(t, expectIs2, is2.Export())
		})

		t.Run("PushFront, Push, ReSlice, SubSlice", func(t *testing.T) {
			base := []any{
				"e1",
				map[string]any{"k2": "e2"},
				"e3",
				[]any{"e4"},
				"e5",
			}
			baseIS := NewImmutableSlice(base)
			baseMut := baseIS.Mutable()

			e0Map := map[string]any{"k0": "e0"}
			eMinus2Slice := []any{"e-2"}
			baseMut.PushFront(e0Map)
			baseMut.PushFront("e-1")
			baseMut.PushFront(eMinus2Slice)

			e6Slice := []any{"e6"}
			e8Map := map[string]any{"k8": "e8"}
			baseMut.Push(e6Slice)
			baseMut.Push("e7")
			baseMut.Push(e8Map)

			expectAfterPushes := []any{
				[]any{"e-2"},
				"e-1",
				map[string]any{"k0": "e0"},
				"e1",
				map[string]any{"k2": "e2"},
				"e3",
				[]any{"e4"},
				"e5",
				[]any{"e6"},
				"e7",
				map[string]any{"k8": "e8"},
			}
			assert.Equal(t, expectAfterPushes, baseMut.Export())
			baseMutFromAll := make([]any, baseMut.Len())
			for i, v := range baseMut.All() {
				baseMutFromAll[i] = ExportValue(v)
			}
			assert.Equal(t, expectAfterPushes, baseMutFromAll)

			baseMutClone := baseMut.Clone()
			baseImmute2 := baseMut.Immutable()

			baseMut.PushFront("start")
			baseMut.Push("end")
			require.Equal(t, 13, baseMut.Len())
			eMinus2SliceM := mustGetSliceFromSlice(t, 1, baseMut)
			eMinus2SliceM.Set(0, "e-2-modified")
			e0MapM := mustGetMapFromSlice(t, 3, baseMut)
			e0MapM.Set("k0", "e0-modified")
			e6MapM := mustGetSliceFromSlice(t, 9, baseMut)
			e6MapM.Set(0, "e6-modified")
			e8MapM := mustGetMapFromSlice(t, 11, baseMut)
			e8MapM.Set("k8", "e8-modified")

			expectAfterMoreMutations := []any{
				"start",
				[]any{"e-2-modified"},
				"e-1",
				map[string]any{"k0": "e0-modified"},
				"e1",
				map[string]any{"k2": "e2"},
				"e3",
				[]any{"e4"},
				"e5",
				[]any{"e6-modified"},
				"e7",
				map[string]any{"k8": "e8-modified"},
				"end",
			}
			assert.Equal(t, expectAfterMoreMutations, baseMut.Export())

			expectBaseMutClone := []any{
				[]any{"e-2-modified"},
				"e-1",
				map[string]any{"k0": "e0-modified"},
				"e1",
				map[string]any{"k2": "e2"},
				"e3",
				[]any{"e4"},
				"e5",
				[]any{"e6-modified"},
				"e7",
				map[string]any{"k8": "e8-modified"},
			}
			assert.Equal(t, expectBaseMutClone, baseMutClone.Export())

			assert.Equal(t, expectAfterPushes, baseImmute2.Export())

			subSlice := baseMut.SubSlice(1, 12)
			expectSubSlice := []any{
				[]any{"e-2-modified"},
				"e-1",
				map[string]any{"k0": "e0-modified"},
				"e1",
				map[string]any{"k2": "e2"},
				"e3",
				[]any{"e4"},
				"e5",
				[]any{"e6-modified"},
				"e7",
				map[string]any{"k8": "e8-modified"},
			}
			assert.Equal(t, expectSubSlice, subSlice.Export())

			baseMut.ReSlice(1, 12)
			assert.Equal(t, expectSubSlice, baseMut.Export())

			// the subslice and reslice should be looking at same underlying data
			eMinus2SliceSub := mustGetSliceFromSlice(t, 0, subSlice)
			eMinus2SliceReslice := mustGetSliceFromSlice(t, 0, baseMut)
			assert.Same(t, eMinus2SliceSub, eMinus2SliceReslice)

			e8MapSub := mustGetMapFromSlice(t, 10, subSlice)
			e8MapReslice := mustGetMapFromSlice(t, 10, baseMut)
			assert.Same(t, e8MapSub, e8MapReslice)

			// update e-1, e3, and e7 in subslice and verify reflected in reslice
			subSlice.Set(1, "e-1-modified")
			subSlice.Set(5, "e3-modified")
			subSlice.Set(9, "e7-modified")

			expectAfterSubsliceMutations := []any{
				[]any{"e-2-modified"},
				"e-1-modified",
				map[string]any{"k0": "e0-modified"},
				"e1",
				map[string]any{"k2": "e2"},
				"e3-modified",
				[]any{"e4"},
				"e5",
				[]any{"e6-modified"},
				"e7-modified",
				map[string]any{"k8": "e8-modified"},
			}
			assert.Equal(t, expectAfterSubsliceMutations, subSlice.Export())
			assert.Equal(t, expectAfterSubsliceMutations, baseMut.Export())
		})

		t.Run("ReSlice/SubSlice completely removing prepends/appends", func(t *testing.T) {
			s := []any{
				[]any{"e5"},
				[]any{"e6"},
				[]any{"e7"},
				[]any{"e8"},
			}
			is := NewImmutableSlice(s)

			mut := is.Mutable()
			mut.PushFront("e4")
			mut.Push("e9")
			subslice := mut.SubSlice(2, 4)
			mut.ReSlice(2, 4)

			expect := []any{
				[]any{"e6"},
				[]any{"e7"},
			}
			assert.Equal(t, expect, subslice.Export())
			assert.Equal(t, expect, mut.Export())

			s0 := mustGetSliceFromSlice(t, 0, mut)
			assert.Equal(t, []any{"e6"}, s0.Export())
			s1 := mustGetSliceFromSlice(t, 1, mut)
			assert.Equal(t, []any{"e7"}, s1.Export())

			ss0 := mustGetSliceFromSlice(t, 0, subslice)
			assert.Equal(t, []any{"e6"}, ss0.Export())
			ss1 := mustGetSliceFromSlice(t, 1, subslice)
			assert.Equal(t, []any{"e7"}, ss1.Export())
		})

		t.Run("instance management in reslice/subslice", func(t *testing.T) {
			base := []any{
				map[string]any{"k3": "e3"},
				[]any{"e4"},
			}
			baseIS := NewImmutableSlice(base)
			baseMut := baseIS.Mutable()
			baseMut.PushFront([]any{"e2"})
			baseMut.PushFront(map[string]any{"k1": "e1"})
			baseMut.PushFront("front")
			baseMut.Push(map[string]any{"k5": "e5"})
			baseMut.Push([]any{"e6"})
			baseMut.Push("end")

			subslice := baseMut.SubSlice(1, 7)

			// assert sameness of all instances
			m1s := mustGetMapFromSlice(t, 0, subslice)
			m1 := mustGetMapFromSlice(t, 1, baseMut)
			s2s := mustGetSliceFromSlice(t, 1, subslice)
			s2 := mustGetSliceFromSlice(t, 2, baseMut)
			m3s := mustGetMapFromSlice(t, 2, subslice)
			m3 := mustGetMapFromSlice(t, 3, baseMut)
			s4s := mustGetSliceFromSlice(t, 3, subslice)
			s4 := mustGetSliceFromSlice(t, 4, baseMut)
			m5s := mustGetMapFromSlice(t, 4, subslice)
			m5 := mustGetMapFromSlice(t, 5, baseMut)
			s6s := mustGetSliceFromSlice(t, 5, subslice)
			s6 := mustGetSliceFromSlice(t, 6, baseMut)

			baseMut.ReSlice(1, 7)
			m1r := mustGetMapFromSlice(t, 0, baseMut)
			s2r := mustGetSliceFromSlice(t, 1, baseMut)
			m3r := mustGetMapFromSlice(t, 2, baseMut)
			s4r := mustGetSliceFromSlice(t, 3, baseMut)
			m5r := mustGetMapFromSlice(t, 4, baseMut)
			s6r := mustGetSliceFromSlice(t, 5, baseMut)

			assert.Same(t, m1s, m1)
			assert.Same(t, m1s, m1r)
			assert.Same(t, s2s, s2)
			assert.Same(t, s2s, s2r)
			assert.Same(t, m3s, m3)
			assert.Same(t, m3s, m3r)
			assert.Same(t, s4s, s4)
			assert.Same(t, s4s, s4r)
			assert.Same(t, m5s, m5)
			assert.Same(t, m5s, m5r)
			assert.Same(t, s6s, s6)
			assert.Same(t, s6s, s6r)
		})

		t.Run("instance management in clone", func(t *testing.T) {
			base := []any{
				map[string]any{"k3": "e3"},
				[]any{"e4"},
			}
			baseIS := NewImmutableSlice(base)
			baseMut := baseIS.Mutable()
			baseMut.PushFront([]any{"e2"})
			baseMut.PushFront(map[string]any{"k1": "e1"})
			baseMut.Push(map[string]any{"k5": "e5"})
			baseMut.Push([]any{"e6"})

			baseMutClone := baseMut.Clone()

			// assert sameness of all instances
			m1c := mustGetMapFromSlice(t, 0, baseMutClone)
			m1 := mustGetMapFromSlice(t, 0, baseMut)
			s2c := mustGetSliceFromSlice(t, 1, baseMutClone)
			s2 := mustGetSliceFromSlice(t, 1, baseMut)
			m3c := mustGetMapFromSlice(t, 2, baseMutClone)
			m3 := mustGetMapFromSlice(t, 2, baseMut)
			s4c := mustGetSliceFromSlice(t, 3, baseMutClone)
			s4 := mustGetSliceFromSlice(t, 3, baseMut)
			m5c := mustGetMapFromSlice(t, 4, baseMutClone)
			m5 := mustGetMapFromSlice(t, 4, baseMut)
			s6c := mustGetSliceFromSlice(t, 5, baseMutClone)
			s6 := mustGetSliceFromSlice(t, 5, baseMut)

			assert.Same(t, m1c, m1)
			assert.Same(t, s2c, s2)
			assert.Same(t, m3c, m3)
			assert.Same(t, s4c, s4)
			assert.Same(t, m5c, m5)
			assert.Same(t, s6c, s6)
		})

		t.Run("ReSlice/SubSlice completely removing append and base", func(t *testing.T) {
			s := []any{
				[]any{"e3"},
				[]any{"e4"},
			}
			is := NewImmutableSlice(s)
			mut := is.Mutable()
			mut.PushFront("e2")
			mut.PushFront("e1")
			mut.Push("e5")
			mut.Push("e6")
			mut.Push("e7")

			subslice := mut.SubSlice(5, 6)
			require.Equal(t, 1, subslice.Len())
			assert.Equal(t, "e6", subslice.At(0))

			mut.ReSlice(5, 6)
			require.Equal(t, 1, mut.Len())
			assert.Equal(t, "e6", mut.At(0))
		})

		t.Run("ReSlice/Subslice compltely removing prepend and base", func(t *testing.T) {
			s := []any{
				[]any{"e3"},
				[]any{"e4"},
			}
			is := NewImmutableSlice(s)
			mut := is.Mutable()
			mut.PushFront("e2")
			mut.PushFront("e1")
			mut.PushFront("e0")
			mut.Push("e5")
			mut.Push("e6")

			subslice := mut.SubSlice(1, 2)
			require.Equal(t, 1, subslice.Len())
			assert.Equal(t, "e1", subslice.At(0))

			mut.ReSlice(1, 2)
			require.Equal(t, 1, mut.Len())
			assert.Equal(t, "e1", mut.At(0))
		})
	})
}

func mustGetMapFromMap(t *testing.T, key string, m *Map) *Map {
	v, ok := m.Get(key)
	require.True(t, ok)
	vMap, ok := v.(*Map)
	require.True(t, ok, "%T: %v", v, v)
	return vMap
}

func mustGetSliceFromMap(t *testing.T, key string, m *Map) *Slice {
	v, ok := m.Get(key)
	require.True(t, ok)
	vSlice, ok := v.(*Slice)
	require.True(t, ok, "%T: %v", v, v)
	return vSlice
}

func mustGetMapFromSlice(t *testing.T, index int, s *Slice) *Map {
	v := s.At(index)
	vMap, ok := v.(*Map)
	require.True(t, ok, "%T: %v", v, v)
	return vMap
}

func mustGetSliceFromSlice(t *testing.T, index int, s *Slice) *Slice {
	v := s.At(index)
	vSlice, ok := v.(*Slice)
	require.True(t, ok, "%T: %v", v, v)
	return vSlice
}
