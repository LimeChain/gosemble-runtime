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
	expectedBytesUnincludedSegment, _ = hex.DecodeString("040500000006000000040a0000000500000006000000013aa96b0149b6ca3688878bdbd19464448624136398e3ce45b9e755d3ab61355c0105")
)

var (
	targetUnincludedSegment = UnincludedSegment{
		Ancestors: sc.Sequence[Ancestor]{
			{
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
			},
		},
	}
)

func Test_UnincludedSegment_Encode(t *testing.T) {
	buffer := &bytes.Buffer{}

	err := targetUnincludedSegment.Encode(buffer)
	assert.NoError(t, err)
	assert.Equal(t, expectedBytesUnincludedSegment, buffer.Bytes())
}

func Test_UnincludedSegment_Decode(t *testing.T) {
	buf := bytes.NewBuffer(expectedBytesUnincludedSegment)

	result, err := DecodeUnincludedSegment(buf)
	assert.NoError(t, err)
	assert.Equal(t, targetUnincludedSegment, result)
}

func Test_UnincludedSegment_Bytes(t *testing.T) {
	assert.Equal(t, expectedBytesUnincludedSegment, targetUnincludedSegment.Bytes())
}
