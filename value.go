package lzval

import (
	"errors"
	"fmt"
	"iter"
	"sync"
)

type (
	// LazyValue represents a read-only byte slice payload which encodes
	// a value, along with a Codec which defines how to decode the
	// payload into said value. The methods on LazyValue are safe to be
	// called concurrently from multiple goroutines.
	LazyValue struct {
		// Payload holds the raw bytes of the Value.
		// We assume that this slice is not mutated after being set;
		// violating this assumption may lead to unexpected behavior.
		Payload []byte

		// Codec informs how the payload is to be decoded
		// and how values are to be encoded.
		// It should not be nil or mutated after being set.
		Codec Codec

		val  *Value
		err  error
		once sync.Once
	}

	// Value represents an immutable decoded value from a LazyValue.
	// The methods on Value are safe to be called concurrently from
	// multiple goroutines.
	Value struct {
		t ValueType
		// todo: benchmark if it's any better to just use an `any` to store val
		valBoolean bool
		valNumber  float64
		valString  string
		valSlice   *LazySlice
		valMap     *LazyMap
	}

	// ValueType represents the type of the value held by a Value.
	ValueType int

	// LazyMap represents a map of string keys to values.
	LazyMap struct {
		m Map
	}

	// LazySlice represents a slice of values.
	LazySlice struct {
		s Slice
	}
)

const (
	TypeNonexistent ValueType = iota
	TypeNull
	TypeNumber
	TypeString
	TypeBoolean
	TypeSlice
	TypeMap
)

// String returns a string representation of the ValueType.
func (t ValueType) String() string {
	switch t {
	case TypeNull:
		return "null"
	case TypeNumber:
		return "number"
	case TypeString:
		return "string"
	case TypeBoolean:
		return "boolean"
	case TypeSlice:
		return "slice"
	case TypeMap:
		return "object"
	default:
		return "nonexistent"
	}
}

// Type returns the type of the Value.
// If the Value is nil, it returns TypeNonexistent.
func (n *Value) Type() ValueType {
	if n == nil {
		return TypeNonexistent
	}
	return n.t
}

// Get returns the concrete value.
// The Go type of the returned value depends on the Value's Type:
//   - TypeNumber: float64
//   - TypeBoolean: bool
//   - TypeString: string
//   - TypeMap: *LazyMap
//   - TypeSlice: *LazySlice
//   - TypeNull: nil
//
// If the Value is nil, it returns nil.
func (v *Value) Get() DecodeValue {
	if v == nil {
		return nil
	}
	switch v.t {
	case TypeNumber:
		return v.valNumber
	case TypeBoolean:
		return v.valBoolean
	case TypeString:
		return v.valString
	case TypeMap:
		return v.valMap
	case TypeSlice:
		return v.valSlice
	default:
		return nil
	}
}

// Number returns the numeric value if its type is TypeNumber.
// Otherwise, it returns 0.
func (v *Value) Number() float64 {
	if v == nil {
		return 0
	}
	switch v.t {
	case TypeNumber:
		return v.valNumber
	default:
		return 0
	}
}

// String returns the string value if its type is TypeString.
// Otherwise, it returns empty string.
func (v *Value) String() string {
	if v == nil {
		return ""
	}
	switch v.t {
	case TypeString:
		return v.valString
	default:
		return ""
	}
}

// Boolean returns the value as boolean if its type is TypeBoolean.
// Otherwise, it returns false.
func (v *Value) Boolean() bool {
	if v == nil {
		return false
	}
	switch v.t {
	case TypeBoolean:
		return v.valBoolean
	default:
		return false
	}
}

// Slice returns the value as LazySlice if its type is TypeSlice.
// Otherwise, it returns nil.
func (v *Value) Slice() *LazySlice {
	if v == nil {
		return nil
	}
	switch v.t {
	case TypeSlice:
		return v.valSlice
	default:
		return nil
	}
}

// Map returns the value as LazyMap if its type is TypeMap.
// Otherwise, it returns nil.
func (v *Value) Map() *LazyMap {
	if v == nil {
		return nil
	}
	switch v.t {
	case TypeMap:
		return v.valMap
	default:
		return nil
	}
}

// Get retrieves a LazyValue representing the value associated with the
// given key in the LazyMap. It returns nil if the field does not exist
// or if the receiver is nil.
func (lm *LazyMap) Get(key string) *LazyValue {
	if lm == nil {
		return nil
	}
	Value, ok := lm.m[key]
	if !ok {
		return nil
	}
	return Value
}

// Len returns the number of fields in the LazyMap.
// If the LazyMap is nil, it returns 0.
func (lm *LazyMap) Len() int {
	if lm == nil {
		return 0
	}
	return len(lm.m)
}

// All iterates over all key, value pairs in the LazyMap.
// Like iterating over a vanilla Go map, the order of
// pairs is non-deterministic.
func (lm *LazyMap) All() iter.Seq2[string, *LazyValue] {
	return func(yield func(string, *LazyValue) bool) {
		if lm == nil {
			return
		}
		for k, v := range lm.m {
			if !yield(k, v) {
				return
			}
		}
	}
}

// RecursiveLoad calls Load on each field in the map and
// each of their child values recursively, returning a value
// of pure Go types. The values returned are all newly
// constructed and the caller is free to mutate them.
func (lm *LazyMap) RecursiveLoad() (map[string]any, error) {
	return recursiveLoadMap(lm)
}

// At retrieves the LazyValue representing the value at the specified
// index. Like a vanilla Go slice, if the index is out of bounds, it will panic.
func (ls *LazySlice) At(index int) *LazyValue {
	if ls == nil {
		_ = []struct{}{}[index] // induce equivalent panic
		return nil
	}
	return ls.s[index]
}

// Len returns the number of elements in the LazySlice.
// If the LazySlice is nil, it returns 0.
func (ls *LazySlice) Len() int {
	if ls == nil {
		return 0
	}
	return len(ls.s)
}

// SubSlice returns a new LazySlice that is a slice of the original LazySlice,
// equivalent to calling `mySlice[l:r]`. Like a vanilla Go slice, if an index
// is out of bounds, it will panic.
func (ls *LazySlice) SubSlice(l, r int) *LazySlice {
	if ls == nil {
		_ = []struct{}{}[l:r] // induce equivalent panic
		return nil
	}
	return &LazySlice{s: ls.s[l:r]}
}

// All iterates over all elements in the LazySlice.
func (ls *LazySlice) All() iter.Seq2[int, *LazyValue] {
	return func(yield func(int, *LazyValue) bool) {
		if ls == nil {
			return
		}
		for i, v := range ls.s {
			if !yield(i, v) {
				return
			}
		}
	}
}

// RecursiveLoad calls Load on each element in the slice and
// each of their child values recursively, returning a value
// of pure Go types. The values returned are all newly
// constructed and the caller is free to mutate them.
func (ls *LazySlice) RecursiveLoad() ([]any, error) {
	return recursiveLoadSlice(ls)
}

// Load decodes the LazyValue's Payload into a Value.
// If the receiver is nil or has no Payload, it returns nil.
// An error is returned if the Codec is nil or if decoding fails.
// If the LazyValue has already been resolved, it returns the cached
// Value (or error, if any).
func (lv *LazyValue) Load() (*Value, error) {
	if lv == nil {
		return nil, nil
	}
	lv.once.Do(func() {
		if len(lv.Payload) == 0 {
			return
		}
		if lv.Codec == nil {
			lv.err = errNoCodec
			return
		}
		var v any
		v, lv.err = lv.Codec.Decode(lv.Payload)
		if lv.err != nil {
			return
		}
		lv.val = new(Value)
		switch v := v.(type) {
		case nil:
			lv.val.t = TypeNull
		case float64:
			lv.val.t = TypeNumber
			lv.val.valNumber = v
		case bool:
			lv.val.t = TypeBoolean
			lv.val.valBoolean = v
		case string:
			lv.val.t = TypeString
			lv.val.valString = v
		case Slice:
			lv.val.t = TypeSlice
			lv.val.valSlice = &LazySlice{s: v}
		case Map:
			lv.val.t = TypeMap
			lv.val.valMap = &LazyMap{m: v}
		default:
			lv.err = fmt.Errorf("unexpected type %T", v)
			lv.val = nil
			return
		}
	})
	return lv.val, lv.err
}

// RecursiveLoad calls Load on the value and each of its child
// values recursively, returning a value of pure Go types.
// The values returned are all newly constructed and the caller
// is free to mutate them.
func (lv *LazyValue) RecursiveLoad() (any, error) {
	v, err := lv.Load()
	if err != nil {
		return nil, err
	}
	return recursiveLoad(v)
}

func (lv *LazyValue) UnmarshalJSON(b []byte) error {
	if lv == nil {
		return errors.New("*LazyValue: UnmarshalJSON on nil pointer")
	}
	lv.Payload = b
	lv.Codec = JSONCodec{}
	return nil
}

var (
	errNoCodec = errors.New("codec should not be nil")
)

func recursiveLoad(v *Value) (any, error) {
	// we duplicate logic from the Get func
	// to avoid redundant checks, though this
	// results in us inspecting the internals
	if v == nil {
		return nil, nil
	}
	switch v.t {
	case TypeNumber:
		return v.valNumber, nil
	case TypeBoolean:
		return v.valBoolean, nil
	case TypeString:
		return v.valString, nil
	case TypeMap:
		return recursiveLoadMap(v.valMap)
	case TypeSlice:
		return recursiveLoadSlice(v.valSlice)
	default:
		return nil, nil
	}
}

func recursiveLoadMap(lm *LazyMap) (map[string]any, error) {
	if lm == nil {
		return nil, nil
	}
	m := make(map[string]any, lm.Len())
	for k, lv := range lm.All() {
		vv, err := lv.Load()
		if err != nil {
			return nil, fmt.Errorf("load %q: %w", k, err)
		}
		dv, err := recursiveLoad(vv)
		if err != nil {
			return nil, fmt.Errorf("load %q: %w", k, err)
		}
		m[k] = dv
	}
	return m, nil
}

func recursiveLoadSlice(ls *LazySlice) ([]any, error) {
	if ls == nil {
		return nil, nil
	}
	s := make([]any, ls.Len())
	for i, lv := range ls.All() {
		vv, err := lv.Load()
		if err != nil {
			return nil, fmt.Errorf("load %d: %w", i, err)
		}
		dv, err := recursiveLoad(vv)
		if err != nil {
			return nil, fmt.Errorf("load %d: %w", i, err)
		}
		s[i] = dv
	}
	return s, nil
}
