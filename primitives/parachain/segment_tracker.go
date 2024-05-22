package parachain

import (
	"bytes"
	sc "github.com/LimeChain/goscale"
)

type SegmentTracker struct {
	UserBandwidth         UserBandwidth
	HrmpWatermark         sc.Option[RelayChainBlockNumber]
	ConsumedGoAheadSignal sc.Option[sc.U8]
}

func (st SegmentTracker) Encode(buffer *bytes.Buffer) error {
	return sc.EncodeEach(buffer, st.UserBandwidth, st.HrmpWatermark, st.ConsumedGoAheadSignal)
}

func DecodeSegmentTracker(buffer *bytes.Buffer) (SegmentTracker, error) {
	userBandwidth, err := DecodeUserBandwidth(buffer)
	if err != nil {
		return SegmentTracker{}, err
	}

	hrmpWatermark, err := sc.DecodeOptionWith(buffer, sc.DecodeU32)
	if err != nil {
		return SegmentTracker{}, err
	}

	consumedGoAheadSignal, err := sc.DecodeOptionWith(buffer, DecodeUpgradeGoAhead)
	if err != nil {
		return SegmentTracker{}, err
	}

	return SegmentTracker{
		UserBandwidth:         userBandwidth,
		HrmpWatermark:         hrmpWatermark,
		ConsumedGoAheadSignal: consumedGoAheadSignal,
	}, nil
}

func (st SegmentTracker) Bytes() []byte {
	return sc.EncodedBytes(st)
}
