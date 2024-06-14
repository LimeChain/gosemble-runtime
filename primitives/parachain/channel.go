package parachain

import (
	"bytes"

	sc "github.com/LimeChain/goscale"
)

type Channel struct {
	ParachainId         sc.U32
	AbridgedHRMPChannel AbridgedHRMPChannel
}

func (c Channel) Encode(buffer *bytes.Buffer) error {
	return sc.EncodeEach(buffer, c.ParachainId, c.AbridgedHRMPChannel)
}

func DecodeChannel(buffer *bytes.Buffer) (Channel, error) {
	parachainId, err := sc.DecodeU32(buffer)
	if err != nil {
		return Channel{}, err
	}

	ac, err := DecodeAbridgedHRMPChannel(buffer)
	if err != nil {
		return Channel{}, err
	}

	return Channel{
		ParachainId:         parachainId,
		AbridgedHRMPChannel: ac,
	}, nil
}

func (c Channel) Bytes() []byte {
	return sc.EncodedBytes(c)
}
