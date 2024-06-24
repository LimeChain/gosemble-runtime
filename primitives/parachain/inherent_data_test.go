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
	expectedBytesInherentData, _ = hex.DecodeString("0c05060a040000003aa96b0149b6ca3688878bdbd19464448624136398e3ce45b9e755d3ab61355c02000000040c010203040a0000000c020304040a000000040b0000000c0b0c0d")
)

var (
	relayParentStorageRoot = common.MustHexToHash("0x3aa96b0149b6ca3688878bdbd19464448624136398e3ce45b9e755d3ab61355c").ToBytes()
	targetInherentData     = InherentData{
		ValidationData: PersistedValidationData{
			ParentHead:             sc.Sequence[sc.U8]{5, 6, 10},
			RelayParentNumber:      4,
			RelayParentStorageRoot: primitives.H256{FixedSequence: sc.BytesToFixedSequenceU8(relayParentStorageRoot)},
			MaxPovSize:             2,
		},
		RelayChainState: StorageProof{
			TrieNodes: sc.Sequence[sc.Sequence[sc.U8]]{
				sc.Sequence[sc.U8]{
					1, 2, 3,
				},
			},
		},
		DownwardMessages: sc.Sequence[InboundDownwardMessage]{
			{
				SentAt: 10,
				Msg:    sc.Sequence[sc.U8]{2, 3, 4},
			},
		},
		HorizontalMessages: HorizontalMessages{
			messages: sc.Dictionary[sc.U32, sc.Sequence[InboundDownwardMessage]]{
				10: sc.Sequence[InboundDownwardMessage]{
					{
						SentAt: 11,
						Msg:    sc.Sequence[sc.U8]{11, 12, 13},
					},
				},
			},
		},
	}
)

func Test_InherentData_Encode(t *testing.T) {
	buffer := &bytes.Buffer{}

	err := targetInherentData.Encode(buffer)
	assert.NoError(t, err)
	assert.Equal(t, expectedBytesInherentData, buffer.Bytes())
}

func Test_InherentData_Decode(t *testing.T) {
	buf := bytes.NewBuffer(expectedBytesInherentData)

	result, err := DecodeInherentData(buf)
	assert.NoError(t, err)
	assert.Equal(t, targetInherentData, result)
}

func Test_InherentData_Bytes(t *testing.T) {
	assert.Equal(t, expectedBytesInherentData, targetInherentData.Bytes())
}
