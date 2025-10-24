package green

import (
	"fmt"
	"sync"
	"testing"
)

func BenchmarkMutable(b *testing.B) {
	numCopies := []int{1, 11, 21, 31, 41, 51}
	// numCopies := []int{10}
	numEventSizes := []int{101}
	// numEventSizes := []int{1, 101, 201, 301, 401, 501}
	concurrency := 32

	payloads := make([]map[string]any, len(numEventSizes))
	for i, eventSize := range numEventSizes {
		bigMap := make(map[string]any, eventSize)
		for j := range eventSize {
			bigMap[fmt.Sprintf("key-%d", j)] = fmt.Sprintf("value-%d", j)
		}
		v := map[string]any{
			"bigMap":     bigMap,
			"last_name":  nil,
			"arms":       2,
			"first_name": "Adam",
			"details": map[string]any{
				"city": "cityname",
				"age":  30,
			},
			"pets": []any{"cat", "dog", "fish"},
		}
		payloads[i] = v
	}

	copyMethod := func(b *testing.B, v map[string]any) {
		v2 := deepCopy(v)
		vMap, ok := v2.(map[string]any)
		if !ok {
			b.Fatal("expected map[string]any")
		}
		vDetailsMap, ok := vMap["details"].(map[string]any)
		if !ok {
			b.Fatal("expected map[string]any for details")
		}
		vDetailsMap["age"] = 31
		vPetsSlice, ok := vMap["pets"].([]any)
		if !ok {
			b.Fatal("expected []any for pets")
		}
		vPetsSlice[0] = "hamster"
	}

	greenMethod := func(b *testing.B, im *ImmutableMap) {
		vMut := im.Mutable()
		vDetails, ok := vMut.Get("details")
		if !ok {
			b.Fatal("expected details key")
		}
		vDetailsMut, ok := vDetails.(*Map)
		if !ok {
			b.Fatal("expected *Map for details")
		}
		vDetailsMut.Set("age", 31)
		vPets, ok := vMut.Get("pets")
		if !ok {
			b.Fatal("expected pets key")
		}
		vPetsMut, ok := vPets.(*Slice)
		if !ok {
			b.Fatal("expected *Slice for pets")
		}
		vPetsMut.Set(0, "hamster")
		_ = vMut.Immutable()
	}

	for _, numCopy := range numCopies {
		for i, eventSize := range numEventSizes {
			v := payloads[i]
			b.Run(fmt.Sprintf("deepcopy_numCopies:%d_eventSize:%d_concurrency:%d", numCopy, eventSize, concurrency),
				func(b *testing.B) {
					for b.Loop() {
						var wg sync.WaitGroup
						for range concurrency {
							wg.Add(1)
							go func() {
								defer wg.Done()
								for range numCopy {
									copyMethod(b, v)
								}
							}()
						}
						wg.Wait()
					}
				},
			)
		}
	}
	for _, numCopy := range numCopies {
		for i, eventSize := range numEventSizes {
			v := payloads[i]
			b.Run(fmt.Sprintf("green_numCopies:%d_eventSize:%d_concurrency:%d", numCopy, eventSize, concurrency),
				func(b *testing.B) {
					for b.Loop() {
						var wg sync.WaitGroup
						for range concurrency {
							wg.Add(1)
							go func() {
								defer wg.Done()
								for range numCopy {
									greenMethod(b, NewImmutableMap(v))
								}
							}()
						}
						wg.Wait()
					}
				},
			)
		}
	}
}
