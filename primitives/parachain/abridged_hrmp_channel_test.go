package parachain

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/ChainSafe/gossamer/lib/common"
	sc "github.com/LimeChain/goscale"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
)

var (
	expectedBytesAbridgedHrmpChannel, _ = hex.DecodeString("0100000002000000030000000400000005000000013aa96b0149b6ca3688878bdbd19464448624136398e3ce45b9e755d3ab61355c")
)

var (
	head                      = common.MustHexToHash("0x3aa96b0149b6ca3688878bdbd19464448624136398e3ce45b9e755d3ab61355c").ToBytes()
	targetAbridgedHrmpChannel = AbridgedHRMPChannel{
		MaxCapacity:    1,
		MaxTotalSize:   2,
		MaxMessageSize: 3,
		MsgCount:       4,
		TotalSize:      5,
		MqcHead: sc.Option[primitives.H256]{
			HasValue: true,
			Value:    primitives.H256{FixedSequence: sc.BytesToFixedSequenceU8(head)},
		},
	}
)

func Test_AbridgedHrmpChannel_Encode(t *testing.T) {
	buffer := &bytes.Buffer{}

	err := targetAbridgedHrmpChannel.Encode(buffer)

	assert.NoError(t, err)
	assert.Equal(t, expectedBytesAbridgedHrmpChannel, buffer.Bytes())
}

func Test_AbridgedHrmpChannel_Decode(t *testing.T) {
	buf := bytes.NewBuffer(expectedBytesAbridgedHrmpChannel)

	result, err := DecodeAbridgedHRMPChannel(buf)
	assert.NoError(t, err)

	assert.Equal(t, targetAbridgedHrmpChannel, result)
}

func Test_AbridgedHrmpChannel_Bytes(t *testing.T) {
	assert.Equal(t, expectedBytesAbridgedHrmpChannel, targetAbridgedHrmpChannel.Bytes())
}
