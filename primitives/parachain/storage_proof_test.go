package parachain

import (
	"bytes"
	"encoding/hex"
	"testing"

	sc "github.com/LimeChain/goscale"

	"github.com/stretchr/testify/assert"
)

var (
	expectedBytesStorageProof, _ = hex.DecodeString("08100405060718010203040506")
)

var (
	targetStorageProof = StorageProof{
		TrieNodes: sc.Sequence[sc.Sequence[sc.U8]]{
			sc.Sequence[sc.U8]{
				4, 5, 6, 7,
			},
			sc.Sequence[sc.U8]{
				1, 2, 3, 4, 5, 6,
			},
		},
	}
)

func Test_StorageProof_Encode(t *testing.T) {
	buffer := &bytes.Buffer{}

	err := targetStorageProof.Encode(buffer)
	assert.NoError(t, err)
	assert.Equal(t, expectedBytesStorageProof, buffer.Bytes())
}

func Test_StorageProof_Decode(t *testing.T) {
	buf := bytes.NewBuffer(expectedBytesStorageProof)

	result, err := DecodeStorageProof(buf)
	assert.NoError(t, err)
	assert.Equal(t, targetStorageProof, result)
}

func Test_StorageProof_Bytes(t *testing.T) {
	assert.Equal(t, expectedBytesStorageProof, targetStorageProof.Bytes())
}
