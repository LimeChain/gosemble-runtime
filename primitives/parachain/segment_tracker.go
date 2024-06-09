package parachain

import (
	"bytes"
	"errors"
	sc "github.com/LimeChain/goscale"
	"reflect"
)

type SegmentTracker struct {
	UsedBandwidth         UsedBandwidth
	HrmpWatermark         sc.Option[RelayChainBlockNumber]
	ConsumedGoAheadSignal sc.Option[sc.U8]
}

func (st SegmentTracker) Encode(buffer *bytes.Buffer) error {
	return sc.EncodeEach(buffer, st.UsedBandwidth, st.HrmpWatermark, st.ConsumedGoAheadSignal)
}

func DecodeSegmentTracker(buffer *bytes.Buffer) (SegmentTracker, error) {
	usedBandwidth, err := DecodeUsedBandwidth(buffer)
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
		UsedBandwidth:         usedBandwidth,
		HrmpWatermark:         hrmpWatermark,
		ConsumedGoAheadSignal: consumedGoAheadSignal,
	}, nil
}

func (st SegmentTracker) Bytes() []byte {
	return sc.EncodedBytes(st)
}

// Subtract removes previously added block from the tracker.
func (st *SegmentTracker) Subtract(block *Ancestor) error {
	err := st.UsedBandwidth.Subtract(&block.UsedBandwidth)
	if err != nil {
		return err
	}

	if block.ConsumedGoAheadSignal.HasValue {
		// This is the same signal stored in the tracker.
		signalInSegment := st.ConsumedGoAheadSignal
		st.ConsumedGoAheadSignal = sc.NewOption[sc.U8](nil)
		if !reflect.DeepEqual(signalInSegment, block.ConsumedGoAheadSignal) {
			return errors.New("mismatching consumed GoAheadSignal")
		}
	}
	// Watermark doesn't need to be updated since this is always dropped
	// from the tail of the segment.
	return nil
}
