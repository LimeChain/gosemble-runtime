package parachain

import (
	"bytes"
	"encoding/hex"
	"testing"

	sc "github.com/LimeChain/goscale"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
)

var (
	expectedBytesChannel, _ = hex.DecodeString("03000000010000000200000003000000040000000500000000")
)

var (
	targetChannel = Channel{
		ParachainId: 3,
		AbridgedHRMPChannel: AbridgedHRMPChannel{
			MaxCapacity:    1,
			MaxTotalSize:   2,
			MaxMessageSize: 3,
			MsgCount:       4,
			TotalSize:      5,
			MqcHead: sc.Option[primitives.H256]{
				HasValue: false,
			},
		},
	}
)

func Test_Channel_Encode(t *testing.T) {
	buffer := &bytes.Buffer{}

	err := targetChannel.Encode(buffer)
	assert.NoError(t, err)
	assert.Equal(t, expectedBytesChannel, buffer.Bytes())
}

func Test_Channel_Decode(t *testing.T) {
	buf := bytes.NewBuffer(expectedBytesChannel)

	result, err := DecodeChannel(buf)
	assert.NoError(t, err)
	assert.Equal(t, targetChannel, result)
}

func Test_Channel_Bytes(t *testing.T) {
	assert.Equal(t, expectedBytesChannel, targetChannel.Bytes())
}
