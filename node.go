package lzval

import (
	"errors"
	"fmt"
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
		once *sync.Once
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
		valArray   *LazyArray
		valObject  *LazyObject
	}

	// ValueType represents the type of the value held by a Value.
	ValueType int

	// LazyObject represents a map of string keys to values.
	LazyObject struct {
		m Object
	}

	// LazyArray represents a slice of values.
	LazyArray struct {
		a Array
	}
)

const (
	TypeNonexistent ValueType = iota
	TypeNull
	TypeNumber
	TypeString
	TypeBoolean
	TypeArray
	TypeObject
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
	case TypeArray:
		return "array"
	case TypeObject:
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

// Value returns the value of the Value.
// The concrete type of the returned value depends on the Value type:
//   - TypeNumber: float64
//   - TypeBoolean: bool
//   - TypeString: string
//   - TypeObject: *LazyObject
//   - TypeArray: *LazyArray
//   - TypeNull: nil
//
// If the Value is nil, it returns nil.
func (n *Value) Value() DecodeValue {
	if n == nil {
		return nil
	}
	switch n.t {
	case TypeNumber:
		return n.valNumber
	case TypeBoolean:
		return n.valBoolean
	case TypeString:
		return n.valString
	case TypeObject:
		return n.valObject
	case TypeArray:
		return n.valArray
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

// Array returns the value as LazyArray if its type is TypeArray.
// Otherwise, it returns nil.
func (v *Value) Array() *LazyArray {
	if v == nil {
		return nil
	}
	switch v.t {
	case TypeArray:
		return v.valArray
	default:
		return nil
	}
}

// Object returns the value as LazyObject if its type is TypeObject.
// Otherwise, it returns nil.
func (v *Value) Object() *LazyObject {
	if v == nil {
		return nil
	}
	switch v.t {
	case TypeArray:
		return v.valObject
	default:
		return nil
	}
}

// Field retrieves a LazyValue representing the value at the given field name
// in the LazyObject. It returns nil if the receiver is nil or if the
// field does not exist.
func (lo *LazyObject) Field(name string) *LazyValue {
	if lo == nil {
		return nil
	}
	Value, ok := lo.m[name]
	if !ok {
		return nil
	}
	return Value
}

// Len returns the number of fields in the LazyObject.
// If the LazyObject is nil, it returns 0.
func (lo *LazyObject) Len() int {
	if lo == nil {
		return 0
	}
	return len(lo.m)
}

// Len returns the number of elements in the LazyArray.
// If the LazyArray is nil, it returns 0.
func (la *LazyArray) Len() int {
	if la == nil {
		return 0
	}
	return len(la.a)
}

// Element retrieves the LazyValue representing the value at the specified
// index. Like a vanilla Go slice, if the index is out of bounds, it will panic.
func (la *LazyArray) Element(i int) *LazyValue {
	if la == nil {
		_ = []struct{}{}[i] // induce equivalent panic
		return nil
	}
	return la.a[i]
}

// SubArray returns a new LazyArray that is a slice of the original LazyArray,
// equivalent to calling `mySlice[l:r]`. Like a vanilla Go slice, if an index
// is out of bounds, it will panic.
func (la *LazyArray) SubArray(l, r int) *LazyArray {
	if la == nil {
		_ = []struct{}{}[l:r] // induce equivalent panic
		return nil
	}
	return &LazyArray{a: la.a[l:r]}
}

// Resolve decodes the LazyValue's Payload into a Value.
// If the receiver is nil or has no Payload, it returns nil.
// An error is returned if the Codec is nil or if decoding fails.
// If the LazyValue has already been resolved, it returns the cached
// Value (or error, if any).
func (lv *LazyValue) Resolve() (*Value, error) {
	lv.once.Do(func() {
		if len(lv.Payload) == 0 {
			return
		}
		if lv.Codec == nil {
			lv.err = errors.New("codec should not be nil")
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
		case Array:
			lv.val.t = TypeArray
			lv.val.valArray = &LazyArray{a: v}
		case Object:
			lv.val.t = TypeObject
			lv.val.valObject = &LazyObject{m: v}
		default:
			lv.err = fmt.Errorf("unexpected type %T", v)
			lv.val = nil
			return
		}
	})
	return lv.val, lv.err
}

func (lv *LazyValue) UnmarshalJSON(b []byte) error {
	if lv == nil {
		return errors.New("*LazyValue: UnmarshalJSON on nil pointer")
	}
	lv.Payload = b
	lv.Codec = JSONCodec{}
	return nil
}
