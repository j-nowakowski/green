package lzval

import "testing"

func TestJSONCodec(t *testing.T) {
	TestCodec(t, JSONCodec{})
}

// func FuzzJSONCodec(f *testing.F) {

// }
