package parachain

import (
	"bytes"
	"encoding/hex"
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/stretchr/testify/assert"
)

var (
	expectedBytesUserBandwidth, _ = hex.DecodeString("0500000006000000040a0000000500000006000000")
)

var (
	targetUserBandwidth = UsedBandwidth{
		UmpMsgCount:   5,
		UmpTotalBytes: 6,
		HrmpOutgoing: sc.Dictionary[sc.U32, HrmpChannelUpdate]{
			10: {
				MsgCount:   5,
				TotalBytes: 6,
			},
		},
	}
)

func Test_UserBandwidth_Encode(t *testing.T) {
	buffer := &bytes.Buffer{}

	err := targetUserBandwidth.Encode(buffer)
	assert.NoError(t, err)
	assert.Equal(t, expectedBytesUserBandwidth, buffer.Bytes())
}

func Test_UserBandwidth_Decode(t *testing.T) {
	buf := bytes.NewBuffer(expectedBytesUserBandwidth)

	result, err := DecodeUsedBandwidth(buf)
	assert.NoError(t, err)
	assert.Equal(t, targetUserBandwidth, result)
}

func Test_UserBandwidth_Bytes(t *testing.T) {
	assert.Equal(t, expectedBytesUserBandwidth, targetUserBandwidth.Bytes())
}
