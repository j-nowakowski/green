package lznode

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNode(t *testing.T, node *Node) {
	t.Run("test node", func(t *testing.T) {
		codec := node.codec
		require.NotNil(t, codec)
		t.Run("null value", func(t *testing.T) {

		})
	})

}

func mustEncode(t *testing.T, codec Codec, v any) []byte {
	b, err := codec.Encode(v)
	require.NoError(t, err)
	return b
}
