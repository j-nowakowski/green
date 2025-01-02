package lzdecode

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
)

type (
	Node interface {
		Type() (NodeType, error)
		Value() (any, error)
		Number() (float64, error)
		String() (string, error)
		Boolean() (bool, error)
		Array() (ArrayNode, error)
		Object() (ObjectNode, error)
	}

	ObjectNode interface {
		GetField(key string) Node
	}

	ArrayNode interface {
		Len() int
		GetElement(i int) Node
	}

	JSONNode struct {
		t          NodeType
		b          []byte
		valBoolean bool
		valNumber  float64
		valString  string
		valArray   *JSONArrayNode
		valObject  *JSONObjectNode
	}

	NodeType int

	DifferentTypeError struct {
		Expected NodeType
		Actual   NodeType
	}

	JSONObjectNode struct {
		m map[string]*JSONNode
	}

	JSONArrayNode struct {
		a []*JSONNode
	}
)

const (
	Nonexistent NodeType = iota
	Null
	Number
	String
	Boolean
	Array
	Object
)

func (t NodeType) String() string {
	switch t {
	case Null:
		return "null"
	case Number:
		return "number"
	case String:
		return "string"
	case Boolean:
		return "boolean"
	case Array:
		return "array"
	case Object:
		return "object"
	default:
		return "nonexistent"
	}
}

var (
	ErrDifferentType = errors.New("different type")
)

func NewJSONNode(b []byte) *JSONNode {
	n := &JSONNode{
		b: b,
	}
	return n
}

func (n *JSONNode) Type() (NodeType, error) {
	if err := n.hydrate(); err != nil {
		return Nonexistent, err
	}
	return n.t, nil
}

func (n *JSONNode) Value() (any, error) {
	if err := n.hydrate(); err != nil {
		return nil, err
	}
	switch n.t {
	case Number:
		return n.valNumber, nil
	case Boolean:
		return n.valBoolean, nil
	case String:
		return n.valString, nil
	case Object:
		return n.valObject, nil
	case Array:
		return n.valArray, nil
	default:
		return nil, nil
	}
}

func (n *JSONNode) Number() (float64, error) {
	if err := n.hydrate(); err != nil {
		return 0, err
	}
	switch n.t {
	case Number:
		return n.valNumber, nil
	default:
		return 0, DifferentTypeError{Expected: Number, Actual: n.t}
	}
}

func (n *JSONNode) String() (string, error) {
	if err := n.hydrate(); err != nil {
		return "", err
	}
	switch n.t {
	case String:
		return n.valString, nil
	default:
		return "", DifferentTypeError{Expected: String, Actual: n.t}
	}
}

func (n *JSONNode) Boolean() (bool, error) {
	if err := n.hydrate(); err != nil {
		return false, err
	}
	switch n.t {
	case Boolean:
		return n.valBoolean, nil
	default:
		return false, DifferentTypeError{Expected: Boolean, Actual: n.t}
	}
}

func (n *JSONNode) Array() (ArrayNode, error) {
	if err := n.hydrate(); err != nil {
		return nil, err
	}
	switch n.t {
	case Array:
		return n.valArray, nil
	default:
		return nil, DifferentTypeError{Expected: Array, Actual: n.t}
	}
}

func (n *JSONNode) Object() (ObjectNode, error) {
	if err := n.hydrate(); err != nil {
		return nil, err
	}
	switch n.t {
	case Array:
		return n.valObject, nil
	default:
		return nil, DifferentTypeError{Expected: Object, Actual: n.t}
	}
}

func (n *JSONObjectNode) GetField(key string) Node {
	// n.m will be hydrated at this point
	node, ok := n.m[key]
	if !ok {
		return &JSONNode{} // nonexistent node
	}
	return node
}

func (n *JSONArrayNode) Len() int {
	return len(n.a)
}

func (n *JSONArrayNode) GetElement(i int) Node {
	return n.a[i] // allow this to panic if out of bounds
}

func (n *JSONNode) UnmarshalJSON(b []byte) error {
	n.b = b
	return nil
}

func (n *JSONNode) MarshalJSON() ([]byte, error) {
	return n.b, nil
}

func (n *JSONNode) hydrate() error {
	if n.t != Nonexistent {
		return nil
	}
	switch {
	case bytes.HasPrefix(n.b, []byte(`{`)):
		// object
		var a map[string]*JSONNode
		if err := json.Unmarshal(n.b, &a); err != nil {
			return err
		}
		n.valObject = &JSONObjectNode{
			m: a,
		}
		n.t = Object
	case bytes.HasPrefix(n.b, []byte(`[`)):
		// array
		var a []*JSONNode
		if err := json.Unmarshal(n.b, &a); err != nil {
			return err
		}
		n.valArray = &JSONArrayNode{
			a: a,
		}
		n.t = Array
	default:
		// some scalar value
		var a any
		if err := json.Unmarshal(n.b, &a); err != nil {
			return err
		}
		switch a := a.(type) {
		case nil:
			n.t = Null
		case float64:
			n.t = Number
			n.valNumber = a
		case bool:
			n.t = Boolean
			n.valBoolean = a
		case string:
			n.t = String
			n.valString = a
		default:
			return fmt.Errorf("unexpected type %T", a)
		}
	}
	return nil
}

func (err DifferentTypeError) Error() string {
	return fmt.Sprintf("expected type %s, actual type is %s", err.Expected, err.Actual)
}

func IsNonexistentNodeError(err error) bool {
	var typeErr DifferentTypeError
	if !errors.As(err, &typeErr) {
		return false
	}
	return typeErr.Actual == Nonexistent
}

// func GetSubfield(n Node, keys ...string) (Node, error) {
// 	for _, k := range keys {

// 	}
// }
