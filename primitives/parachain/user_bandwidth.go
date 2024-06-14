package parachain

import (
	"bytes"

	sc "github.com/LimeChain/goscale"
)

type UsedBandwidth struct {
	UmpMsgCount   sc.U32
	UmpTotalBytes sc.U32
	HrmpOutgoing  sc.Dictionary[sc.U32, HrmpChannelUpdate]
}

func (ub UsedBandwidth) Encode(buffer *bytes.Buffer) error {
	return sc.EncodeEach(buffer, ub.UmpMsgCount, ub.UmpTotalBytes, ub.HrmpOutgoing)
}

func DecodeUsedBandwidth(buffer *bytes.Buffer) (UsedBandwidth, error) {
	msgCount, err := sc.DecodeU32(buffer)
	if err != nil {
		return UsedBandwidth{}, err
	}

	umpTotalBytes, err := sc.DecodeU32(buffer)
	if err != nil {
		return UsedBandwidth{}, err
	}

	outgoing, err := decodeHrmpOutgoing(buffer)
	if err != nil {
		return UsedBandwidth{}, err
	}

	return UsedBandwidth{
		UmpMsgCount:   msgCount,
		UmpTotalBytes: umpTotalBytes,
		HrmpOutgoing:  outgoing,
	}, nil
}

func (ub UsedBandwidth) Bytes() []byte {
	return sc.EncodedBytes(ub)
}

func (ub *UsedBandwidth) Subtract(other *UsedBandwidth) error {
	ub.UmpMsgCount -= other.UmpMsgCount
	ub.UmpTotalBytes -= other.UmpTotalBytes

	newHrmpOutgoing := sc.Dictionary[sc.U32, HrmpChannelUpdate]{}
	for i, channel := range other.HrmpOutgoing {
		entry, ok := ub.HrmpOutgoing[i]
		if !ok {

		}
		entry.Subtract(channel)
		if !entry.IsEmpty() {
			newHrmpOutgoing[i] = entry
		}
	}

	ub.HrmpOutgoing = newHrmpOutgoing

	return nil
}

func decodeHrmpOutgoing(buffer *bytes.Buffer) (sc.Dictionary[sc.U32, HrmpChannelUpdate], error) {
	v, err := sc.DecodeCompact[sc.U128](buffer)
	if err != nil {
		return sc.Dictionary[sc.U32, HrmpChannelUpdate]{}, err
	}
	size := int(v.ToBigInt().Int64())

	result := sc.Dictionary[sc.U32, HrmpChannelUpdate]{}

	for i := 0; i < size; i++ {
		key, err := sc.DecodeU32(buffer)
		if err != nil {
			return sc.Dictionary[sc.U32, HrmpChannelUpdate]{}, err
		}
		value, err := DecodeHrmpChannelUpdate(buffer)
		if err != nil {
			return sc.Dictionary[sc.U32, HrmpChannelUpdate]{}, err
		}

		result[key] = value
	}

	return result, nil
}
