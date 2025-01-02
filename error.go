package lznode

import (
	"errors"
	"fmt"
)

type (
	DifferentTypeError struct {
		Expected NodeType
		Actual   NodeType
	}
)

func (err DifferentTypeError) Error() string {
	return fmt.Sprintf("expected type %s, actual type is %s", err.Expected, err.Actual)
}

func IsNonexistentNodeError(err error) bool {
	var typeErr DifferentTypeError
	if !errors.As(err, &typeErr) {
		return false
	}
	return typeErr.Actual == TypeNonexistent
}
