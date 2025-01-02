package lznode

import (
	"errors"
	"fmt"

	"github.com/j-nowakowski/lznode/codec"
)

type (
	Node struct {
		t          NodeType
		b          []byte
		codec      codec.Codec
		valBoolean bool
		valNumber  float64
		valString  string
		valArray   *Array
		valObject  *Object
	}

	NodeType int

	Object struct {
		m map[string]*Node
	}

	Array struct {
		a []*Node
	}
)

const (
	TypeNonexistent NodeType = iota
	TypeNull
	TypeNumber
	TypeString
	TypeBoolean
	TypeArray
	TypeObject
)

func (t NodeType) String() string {
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

var jsonCodec = codec.JSONCodec{}

func NewJSONNode(b []byte) *Node {
	n := &Node{
		b:     b,
		codec: jsonCodec,
	}
	return n
}

func (n *Node) Type() (NodeType, error) {
	if n == nil {
		return TypeNonexistent, errors.New("*Node.Type: called on nil pointer")
	}
	if err := n.hydrate(); err != nil {
		return TypeNonexistent, err
	}
	return n.t, nil
}

func (n *Node) Value() (any, error) {
	if n == nil {
		return nil, errors.New("*Node.Value: called on nil pointer")
	}
	if err := n.hydrate(); err != nil {
		return nil, err
	}
	switch n.t {
	case TypeNumber:
		return n.valNumber, nil
	case TypeBoolean:
		return n.valBoolean, nil
	case TypeString:
		return n.valString, nil
	case TypeObject:
		return n.valObject, nil
	case TypeArray:
		return n.valArray, nil
	default:
		return nil, nil
	}
}

func (n *Node) Number() (float64, error) {
	if n == nil {
		return 0, errors.New("*Node.Number: called on nil pointer")
	}
	if err := n.hydrate(); err != nil {
		return 0, err
	}
	switch n.t {
	case TypeNumber:
		return n.valNumber, nil
	default:
		return 0, DifferentTypeError{Expected: TypeNumber, Actual: n.t}
	}
}

func (n *Node) String() (string, error) {
	if n == nil {
		return "", errors.New("*Node.String: called on nil pointer")
	}
	if err := n.hydrate(); err != nil {
		return "", err
	}
	switch n.t {
	case TypeString:
		return n.valString, nil
	default:
		return "", DifferentTypeError{Expected: TypeString, Actual: n.t}
	}
}

func (n *Node) Boolean() (bool, error) {
	if n == nil {
		return false, errors.New("*Node.Boolean: called on nil pointer")
	}
	if err := n.hydrate(); err != nil {
		return false, err
	}
	switch n.t {
	case TypeBoolean:
		return n.valBoolean, nil
	default:
		return false, DifferentTypeError{Expected: TypeBoolean, Actual: n.t}
	}
}

func (n *Node) Array() (*Array, error) {
	if n == nil {
		return nil, errors.New("*Node.Array: called on nil pointer")
	}
	if err := n.hydrate(); err != nil {
		return nil, err
	}
	switch n.t {
	case TypeArray:
		return n.valArray, nil
	default:
		return nil, DifferentTypeError{Expected: TypeArray, Actual: n.t}
	}
}

func (n *Node) Object() (*Object, error) {
	if n == nil {
		return nil, errors.New("*Node.Object: called on nil pointer")
	}
	if err := n.hydrate(); err != nil {
		return nil, err
	}
	switch n.t {
	case TypeArray:
		return n.valObject, nil
	default:
		return nil, DifferentTypeError{Expected: TypeObject, Actual: n.t}
	}
}

func (n *Object) GetField(key string) *Node {
	if n == nil {
		return &Node{} // nonexistent node
	}
	node, ok := n.m[key]
	if !ok {
		return &Node{} // nonexistent node
	}
	return node
}

func (n *Array) Len() int {
	if n == nil {
		return 0
	}
	return len(n.a)
}

func (n *Array) GetElement(i int) *Node {
	return n.a[i] // allow this to panic if out of bounds
}

func (n *Node) hydrate() error {
	if n.t != TypeNonexistent {
		return nil
	}
	v, err := n.codec.Decode(n.b)
	if err != nil {
		return err
	}
	switch v := v.(type) {
	case nil:
		n.t = TypeNull
	case float64:
		n.t = TypeNumber
		n.valNumber = v
	case bool:
		n.t = TypeBoolean
		n.valBoolean = v
	case string:
		n.t = TypeString
		n.valString = v
	case codec.Array:
		n.t = TypeArray
		n.valArray = &Array{
			a: make([]*Node, len(v)),
		}
		for i, b := range v {
			n.valArray.a[i] = &Node{
				b:     b,
				codec: n.codec,
			}
		}
	case codec.Object:
		n.t = TypeObject
		n.valObject = &Object{
			m: make(map[string]*Node, len(v)),
		}
		for k, b := range v {
			n.valObject.m[k] = &Node{
				b:     b,
				codec: n.codec,
			}
		}
	default:
		return fmt.Errorf("unexpected type %T", v)
	}
	return nil
}

// func GetSubfield(n Node, keys ...string) (Node, error) {
// 	for _, k := range keys {

// 	}
// }
