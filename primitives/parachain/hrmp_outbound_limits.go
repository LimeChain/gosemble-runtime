package parachain

import (
	"bytes"

	sc "github.com/LimeChain/goscale"
)

type HrmpOutboundLimits struct {
	BytesRemaining    sc.U32
	MessagesRemaining sc.U32
}

func (hol HrmpOutboundLimits) Encode(buffer *bytes.Buffer) error {
	return sc.EncodeEach(buffer, hol.BytesRemaining, hol.MessagesRemaining)
}

func DecodeHrmpOutboundLimits(buffer *bytes.Buffer) (HrmpOutboundLimits, error) {
	bytesRemaining, err := sc.DecodeU32(buffer)
	if err != nil {
		return HrmpOutboundLimits{}, err
	}

	messagesRemaining, err := sc.DecodeU32(buffer)
	if err != nil {
		return HrmpOutboundLimits{}, err
	}

	return HrmpOutboundLimits{
		BytesRemaining:    bytesRemaining,
		MessagesRemaining: messagesRemaining,
	}, nil
}

func (hol HrmpOutboundLimits) Bytes() []byte {
	return sc.EncodedBytes(hol)
}
