package parachain

import (
	"bytes"
	sc "github.com/LimeChain/goscale"
)

type UserBandwidth struct {
	UmpMsgCount   sc.U32
	UmpTotalBytes sc.U32
	HrmpOutgoing  sc.Dictionary[sc.U32, HrmpChannelUpdate]
}

func (ub UserBandwidth) Encode(buffer *bytes.Buffer) error {
	return sc.EncodeEach(buffer, ub.UmpMsgCount, ub.UmpTotalBytes, ub.HrmpOutgoing)
}

func DecodeUserBandwidth(buffer *bytes.Buffer) (UserBandwidth, error) {
	msgCount, err := sc.DecodeU32(buffer)
	if err != nil {
		return UserBandwidth{}, err
	}

	umpTotalBytes, err := sc.DecodeU32(buffer)
	if err != nil {
		return UserBandwidth{}, err
	}

	outgoing, err := decodeHrmpOutgoing(buffer)
	if err != nil {
		return UserBandwidth{}, err
	}

	return UserBandwidth{
		UmpMsgCount:   msgCount,
		UmpTotalBytes: umpTotalBytes,
		HrmpOutgoing:  outgoing,
	}, nil
}

func (ub UserBandwidth) Bytes() []byte {
	return sc.EncodedBytes(ub)
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
