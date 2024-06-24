package parachain

import (
	"bytes"
	"encoding/hex"
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
)

var (
	expectedBytesPersistedValidationData, _ = hex.DecodeString("0c01020301000000000000000000000000000000000000000000000000000000000000000000000102000000")
)

var (
	targetPersistedValidationData = PersistedValidationData{
		ParentHead:             sc.Sequence[sc.U8]{1, 2, 3},
		RelayParentNumber:      1,
		RelayParentStorageRoot: primitives.H256{FixedSequence: constants.OneAccountId.FixedSequence},
		MaxPovSize:             2,
	}
)

func Test_PersistedValidationData_Encode(t *testing.T) {
	buffer := &bytes.Buffer{}

	err := targetPersistedValidationData.Encode(buffer)
	assert.NoError(t, err)
	assert.Equal(t, expectedBytesPersistedValidationData, buffer.Bytes())
}

func Test_PersistedValidationData_Decode(t *testing.T) {
	buf := bytes.NewBuffer(expectedBytesPersistedValidationData)

	result, err := DecodePersistedValidationData(buf)
	assert.NoError(t, err)
	assert.Equal(t, targetPersistedValidationData, result)
}

func Test_PersistedValidationData_Bytes(t *testing.T) {
	assert.Equal(t, expectedBytesPersistedValidationData, targetPersistedValidationData.Bytes())
}
