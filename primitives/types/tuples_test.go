package types

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	tuple2U64Value = Tuple2U64{
		First:  1,
		Second: 2,
	}
)

var (
	tuple2U64Bytes = []byte{
		0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
		0x2, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	}
)

func Test_Tuple2U64_Encode(t *testing.T) {
	buffer := &bytes.Buffer{}

	err := tuple2U64Value.Encode(buffer)

	assert.NoError(t, err)
	assert.Equal(t, tuple2U64Bytes, buffer.Bytes())
}

func Test_Tuple2U64_Bytes(t *testing.T) {
	assert.Equal(t, tuple2U64Bytes, tuple2U64Value.Bytes())
}

func Test_DecodeTuple2U64(t *testing.T) {
	buffer := bytes.NewBuffer(tuple2U64Bytes)

	decodedTuple2U64, err := DecodeTuple2U64(buffer)

	assert.NoError(t, err)
	assert.Equal(t, tuple2U64Value, decodedTuple2U64)
}
