package lzval

import "context"

//go:generate mockery --name=Codec
type (
	MemSlice = []*MemValue
	MemMap   = map[string]*MemValue

	Codec interface {
		Encode(context.Context, any) ([]byte, error)
		Decode(context.Context, []byte) (any, error)
	}
)
