package parachain

import (
	"bytes"
	"encoding/hex"
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/stretchr/testify/assert"
)

var (
	expectedBytesSegmentTracker, _ = hex.DecodeString("01000000020000000001050000000101")
)

var (
	targetSegmentTracker = SegmentTracker{
		UsedBandwidth: UsedBandwidth{
			UmpMsgCount:   1,
			UmpTotalBytes: 2,
			HrmpOutgoing:  sc.Dictionary[sc.U32, HrmpChannelUpdate]{},
		},
		HrmpWatermark: sc.Option[RelayChainBlockNumber]{
			HasValue: true,
			Value:    5,
		},
		ConsumedGoAheadSignal: sc.Option[sc.U8]{
			HasValue: true,
			Value:    1,
		},
	}
)

func Test_SegmentTracker_Encode(t *testing.T) {
	buffer := &bytes.Buffer{}

	err := targetSegmentTracker.Encode(buffer)
	assert.NoError(t, err)
	assert.Equal(t, expectedBytesSegmentTracker, buffer.Bytes())
}

func Test_SegmentTracker_Decode(t *testing.T) {
	buf := bytes.NewBuffer(expectedBytesSegmentTracker)

	result, err := DecodeSegmentTracker(buf)
	assert.NoError(t, err)
	assert.Equal(t, targetSegmentTracker, result)
}

func Test_SegmentTracker_Bytes(t *testing.T) {
	assert.Equal(t, expectedBytesSegmentTracker, targetSegmentTracker.Bytes())
}
