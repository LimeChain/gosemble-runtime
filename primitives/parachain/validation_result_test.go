package parachain

import (
	"bytes"
	"encoding/hex"
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/stretchr/testify/assert"
)

var (
	expectedBytesValidationResult, _ = hex.DecodeString("140102030305011002030405041001020304040b000000140a0b0c0d0e0a0000000b000000")
)

var (
	targetValidationResult = ValidationResult{
		HeadData: sc.Sequence[sc.U8]{1, 2, 3, 3, 5},
		NewValidationCode: sc.Option[sc.Sequence[sc.U8]]{
			HasValue: true,
			Value:    sc.Sequence[sc.U8]{2, 3, 4, 5},
		},
		UpwardMessages: sc.Sequence[UpwardMessage]{
			sc.Sequence[sc.U8]{1, 2, 3, 4},
		},
		HorizontalMessages: sc.Sequence[OutboundHrmpMessage]{
			{
				Id:   11,
				Data: sc.Sequence[sc.U8]{10, 11, 12, 13, 14},
			},
		},
		ProcessedDownwardMessages: 10,
		HrmpWatermark:             11,
	}
)

func Test_ValidationResult_Encode(t *testing.T) {
	buffer := &bytes.Buffer{}

	err := targetValidationResult.Encode(buffer)
	assert.NoError(t, err)
	assert.Equal(t, expectedBytesValidationResult, buffer.Bytes())
}

func Test_ValidationResult_Decode(t *testing.T) {
	buf := bytes.NewBuffer(expectedBytesValidationResult)

	result, err := DecodeValidationResult(buf)
	assert.NoError(t, err)
	assert.Equal(t, targetValidationResult, result)
}

func Test_ValidationResult_Bytes(t *testing.T) {
	assert.Equal(t, expectedBytesValidationResult, targetValidationResult.Bytes())
}
