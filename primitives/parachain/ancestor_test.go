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
	expectedBytesAncestor, _ = hex.DecodeString("0500000006000000040a0000000500000006000000013aa96b0149b6ca3688878bdbd19464448624136398e3ce45b9e755d3ab61355c0105")
)

var (
	paraHeadHash   = common.MustHexToHash("0x3aa96b0149b6ca3688878bdbd19464448624136398e3ce45b9e755d3ab61355c").ToBytes()
	targetAncestor = Ancestor{
		UsedBandwidth: UsedBandwidth{
			UmpMsgCount:   5,
			UmpTotalBytes: 6,
			HrmpOutgoing: sc.Dictionary[sc.U32, HrmpChannelUpdate]{
				10: {
					MsgCount:   5,
					TotalBytes: 6,
				},
			},
		},
		ParaHeadHash: sc.Option[primitives.H256]{
			HasValue: true,
			Value:    primitives.H256{FixedSequence: sc.BytesToFixedSequenceU8(paraHeadHash)},
		},
		ConsumedGoAheadSignal: sc.Option[sc.U8]{
			HasValue: true,
			Value:    5,
		},
	}
)

func Test_Ancestor_Encode(t *testing.T) {
	buffer := &bytes.Buffer{}

	err := targetAncestor.Encode(buffer)

	assert.NoError(t, err)
	assert.Equal(t, expectedBytesAncestor, buffer.Bytes())
}

func Test_Ancestor_Decode(t *testing.T) {
	buf := bytes.NewBuffer(expectedBytesAncestor)

	result, err := DecodeAncestor(buf)
	assert.NoError(t, err)

	assert.Equal(t, targetAncestor, result)
}

func Test_Ancestor_Bytes(t *testing.T) {
	assert.Equal(t, expectedBytesAncestor, targetAncestor.Bytes())
}
