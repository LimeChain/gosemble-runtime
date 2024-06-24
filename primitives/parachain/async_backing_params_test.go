package parachain

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	expectedBytesAsyncBackingParams, _ = hex.DecodeString("0400000005000000")
)

var (
	targetAsyncBackingParams = AsyncBackingParams{
		MaxCandidateDepth:  4,
		AllowedAncestryLen: 5,
	}
)

func Test_AsyncBackingParams_Encode(t *testing.T) {
	buffer := &bytes.Buffer{}

	err := targetAsyncBackingParams.Encode(buffer)
	assert.NoError(t, err)
	assert.Equal(t, expectedBytesAsyncBackingParams, buffer.Bytes())
}

func Test_AsyncBackingParams_Decode(t *testing.T) {
	buf := bytes.NewBuffer(expectedBytesAsyncBackingParams)

	result, err := DecodeAsyncBackingParams(buf)
	assert.NoError(t, err)
	assert.Equal(t, targetAsyncBackingParams, result)
}

func Test_AsyncBackingParams_Bytes(t *testing.T) {
	assert.Equal(t, expectedBytesAsyncBackingParams, targetAsyncBackingParams.Bytes())
}
