package parachain

import (
	"bytes"
	"encoding/hex"
	"testing"

	primitives "github.com/LimeChain/gosemble/primitives/types"

	sc "github.com/LimeChain/goscale"
	"github.com/stretchr/testify/assert"
)

var (
	expectedBytesValidationParams, _ = hex.DecodeString("100304050610050607080a0000003aa96b0149b6ca3688878bdbd19464448624136398e3ce45b9e755d3ab61355c")
)

var (
	targetValidationParams = ValidationParams{
		ParentHead:             sc.Sequence[sc.U8]{3, 4, 5, 6},
		BlockData:              sc.Sequence[sc.U8]{5, 6, 7, 8},
		RelayParentBlockNumber: 10,
		RelayParentStorageRoot: primitives.H256{FixedSequence: sc.BytesToFixedSequenceU8(relayParentStorageRoot)},
	}
)

func Test_ValidationParams_Encode(t *testing.T) {
	buffer := &bytes.Buffer{}

	err := targetValidationParams.Encode(buffer)
	assert.NoError(t, err)
	assert.Equal(t, expectedBytesValidationParams, buffer.Bytes())
}

func Test_ValidationParams_Decode(t *testing.T) {
	buf := bytes.NewBuffer(expectedBytesValidationParams)

	result, err := DecodeValidationParams(buf)
	assert.NoError(t, err)
	assert.Equal(t, targetValidationParams, result)
}

func Test_ValidationParams_Bytes(t *testing.T) {
	assert.Equal(t, expectedBytesValidationParams, targetValidationParams.Bytes())
}
