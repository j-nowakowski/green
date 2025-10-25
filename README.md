# `green`: Copy-On-Write Container Types for Go

Values of type `map[string]any` and `[]any`, referred to as *container types* in
this package, are used ubiquitously in free format objects and arrays. However,
these types do not have any access controls; you cannot enforce that a map or
slice is not mutated by other code. The result is that owners of container
entities are forced to deep copy their entities prior to passing in to
potentially mutating code, which is expensive.

For example, an application wide cache might serve the results of recent queries
for future asynchronous requests to leverage. If the cache returns a native Go
map or slice and one of the future asynchronous queries mutates it, then the
other requests will see mutated data, leading to hard-to-track bugs.

`green` introduces **immutable** and **mutable** container types with methods
which are semantically similar to their Go equivalents. As the names suggest,
immutable maps and slices can be read from but not modified. Nested containers
within immutables are themselves immutable. Immutable values can derive mutable
views of the underlying immutable value. Mutable values can be modified without
modifying the underlying immutable value nor other mutable views derived from
that same value. When modifications occur, a **copy-on-write** approach is used
to copy only the necessary parent containers. Finally, mutable views can be
canonized back to an immutable value.

## Examples

Probing an immutable map:

```go
m := map[string]any{
    "breed": "Great Pyrenees",
    "tricks": []any{"sit", "shake"}
}
im := green.NewImmutableMap(m)

breed, ok := im.Get("breed")
if !ok {
    // handle
}
fmt.Println(breed) // prints "Great Pyrenees"

tricks, ok := im.Get("tricks")
if !ok {
    // handle
}
tricksSlice, ok := tricks.(*green.ImmutableSlice)
if !ok {
    // handle
}
fmt.Println(tricksSlice.At(1)) // prints "shake"
```

Deriving and modifying mutable views:

```go
m := map[string]any{
    "breed": "Great Pyrenees",
    "tricks": []any{"sit", "shake"}
}
im := green.NewImmutableMap(m)

mm1 := im.Mutable() // derive a mutable
mm2 := im.Mutable() // derive another mutable

mm1.Set("age", 6)           // update first mutable
fmt.Println(mm1.Has("age")) // prints "true"
fmt.Println(mm2.Has("age")) // prints "false"
fmt.Println(im.Has("age"))  // prints "false"
fmt.Println(len(m))         // prints "2"

tricks, ok := mm1.Get("tricks")
if !ok {
    // handle
}
tricksSlice, ok := tricks.(*green.Slice) // mutable slice
if !ok {
    // handle
}
tricksSlice.Set(1, "laydown")
tricksSlice.Push("speak")

im1 := mm1.Immutable() // canonize a new immutable
```
