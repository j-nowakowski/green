package lzval

import "testing"

func TestJSONNode(t *testing.T) {
	TestCodec(t, JSONCodec{})
}

// func FuzzJSONNode(f *testing.F) {

// }
