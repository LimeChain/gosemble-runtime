package parachain

import (
	"bytes"
	"encoding/hex"
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/stretchr/testify/assert"
)

var (
	expectedBytesCollationInfo, _ = hex.DecodeString("040c01020304030000000c030405010c060708010000000200000008090a")
)

var (
	targetCollationInfo = CollationInfo{
		UpwardMessages: sc.Sequence[UpwardMessage]{
			sc.Sequence[sc.U8]{1, 2, 3},
		},
		HorizontalMessages: sc.Sequence[OutboundHrmpMessage]{
			{
				Id:   3,
				Data: sc.Sequence[sc.U8]{3, 4, 5},
			},
		},
		ValidationCode: sc.Option[sc.Sequence[sc.U8]]{
			HasValue: true,
			Value:    sc.Sequence[sc.U8]{6, 7, 8},
		},
		ProcessedDownwardMessages: 1,
		HrmpWatermark:             2,
		HeadData:                  sc.Sequence[sc.U8]{9, 10},
	}
)

func Test_CollationInfo_Encode(t *testing.T) {
	buffer := &bytes.Buffer{}

	err := targetCollationInfo.Encode(buffer)
	assert.NoError(t, err)
	assert.Equal(t, expectedBytesCollationInfo, buffer.Bytes())
}

func Test_CollationInfo_Decode(t *testing.T) {
	buf := bytes.NewBuffer(expectedBytesCollationInfo)

	result, err := DecodeCollationInfo(buf)
	assert.NoError(t, err)
	assert.Equal(t, targetCollationInfo, result)
}

func Test_CollationInfo_Bytes(t *testing.T) {
	assert.Equal(t, expectedBytesCollationInfo, targetCollationInfo.Bytes())
}
