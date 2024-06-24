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
	expectedBytesMessageQueueChain, _ = hex.DecodeString("3aa96b0149b6ca3688878bdbd19464448624136398e3ce45b9e755d3ab61355c")
)

var (
	messageQueueChainRelayHash = common.MustHexToHash("0x3aa96b0149b6ca3688878bdbd19464448624136398e3ce45b9e755d3ab61355c").ToBytes()
	targetMessageQueueChain    = MessageQueueChain{
		RelayHash: primitives.H256{FixedSequence: sc.BytesToFixedSequenceU8(messageQueueChainRelayHash)},
	}
)

func Test_MessageQueueChain_Encode(t *testing.T) {
	buffer := &bytes.Buffer{}

	err := targetMessageQueueChain.Encode(buffer)
	assert.NoError(t, err)
	assert.Equal(t, expectedBytesMessageQueueChain, buffer.Bytes())
}

func Test_MessageQueueChain_Decode(t *testing.T) {
	buf := bytes.NewBuffer(expectedBytesMessageQueueChain)

	result, err := DecodeMessageQueueChain(buf)
	assert.NoError(t, err)
	assert.Equal(t, targetMessageQueueChain, result)
}

func Test_MessageQueueChain_Bytes(t *testing.T) {
	assert.Equal(t, expectedBytesMessageQueueChain, targetMessageQueueChain.Bytes())
}
