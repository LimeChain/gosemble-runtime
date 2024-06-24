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
	expectedBytesMessagingStateSnapshot, _ = hex.DecodeString("0000000000000000000000000000000000000000000000000000000000000000030000000400000004020000000100000002000000030000000400000005000000000401000000060000000500000004000000030000000200000000")
)

var (
	targetMessagingSnapshot = MessagingStateSnapshot{
		DmqMqcHead: primitives.H256{FixedSequence: constants.ZeroAccountId.FixedSequence},
		RelayDispatchQueueRemainingCapacity: RelayDispatchQueueRemainingCapacity{
			RemainingCount: 3,
			RemainingSize:  4,
		},
		IngressChannels: sc.Sequence[Channel]{
			{
				ParachainId: 2,
				AbridgedHRMPChannel: AbridgedHRMPChannel{
					MaxCapacity:    1,
					MaxTotalSize:   2,
					MaxMessageSize: 3,
					MsgCount:       4,
					TotalSize:      5,
					MqcHead:        sc.Option[primitives.H256]{},
				},
			},
		},
		EgressChannels: sc.Sequence[Channel]{
			{
				ParachainId: 1,
				AbridgedHRMPChannel: AbridgedHRMPChannel{
					MaxCapacity:    6,
					MaxTotalSize:   5,
					MaxMessageSize: 4,
					MsgCount:       3,
					TotalSize:      2,
					MqcHead:        sc.Option[primitives.H256]{},
				},
			},
		},
	}
)

func Test_MessagingStateSnapshot_Encode(t *testing.T) {
	buffer := &bytes.Buffer{}

	err := targetMessagingSnapshot.Encode(buffer)
	assert.NoError(t, err)
	assert.Equal(t, expectedBytesMessagingStateSnapshot, buffer.Bytes())
}

func Test_MessagingStateSnapshot_Decode(t *testing.T) {
	buf := bytes.NewBuffer(expectedBytesMessagingStateSnapshot)

	result, err := DecodeMessagingStateSnapshot(buf)
	assert.NoError(t, err)
	assert.Equal(t, targetMessagingSnapshot, result)
}

func Test_MessagingStateSnapshot_Bytes(t *testing.T) {
	assert.Equal(t, expectedBytesMessagingStateSnapshot, targetMessagingSnapshot.Bytes())
}
