package lzval

import "context"

type (
	Loadable[T any] interface {
		Load(context.Context) (T, error)
	}
)
