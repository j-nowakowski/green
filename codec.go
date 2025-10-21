package lzval

import "context"

//go:generate mockery
type (
	LoadableSlice = []ValueLoader
	LoadableMap   = map[string]ValueLoader
	DecodeValue   = any

	Codec interface {
		Encode(context.Context, any) ([]byte, error)
		Decode(context.Context, []byte) (DecodeValue, error)
	}
)
