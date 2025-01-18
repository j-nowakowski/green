package lznode

import (
	"errors"
	"fmt"
)

type (
	Node struct {
		t          NodeType
		b          []byte
		codec      Codec
		valBoolean bool
		valNumber  float64
		valString  string
		valArray   *ArrayNode
		valObject  *ObjectNode
	}

	NodeType int

	ObjectNode struct {
		m Object
	}

	ArrayNode struct {
		a Array
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

func NewNode(b []byte, codec Codec) *Node {
	return &Node{
		b:     b,
		codec: codec,
	}
}

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

func (n *Node) Array() (*ArrayNode, error) {
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

func (n *Node) Object() (*ObjectNode, error) {
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

func (n *ObjectNode) GetField(key string) *Node {
	if n == nil {
		return &Node{} // nonexistent node
	}
	node, ok := n.m[key]
	if !ok {
		return &Node{} // nonexistent node
	}
	return node
}

func (n *ArrayNode) Len() int {
	if n == nil {
		return 0
	}
	return len(n.a)
}

func (n *ArrayNode) GetElement(i int) *Node {
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
	case Array:
		n.t = TypeArray
		n.valArray = &ArrayNode{
			a: v,
		}
	case Object:
		n.t = TypeObject
		n.valObject = &ObjectNode{
			m: v,
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
