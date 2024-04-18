package parachain

import (
	"bytes"
	sc "github.com/LimeChain/goscale"
)

type OutboundHrmpMessage struct {
	Id   sc.U32
	Data sc.Sequence[sc.U8]
}

func (ohm OutboundHrmpMessage) Encode(buffer *bytes.Buffer) error {
	return sc.EncodeEach(buffer, ohm.Id, ohm.Data)
}

func DecodeOutboundHrmpMessage(buffer *bytes.Buffer) (OutboundHrmpMessage, error) {
	id, err := sc.DecodeU32(buffer)
	if err != nil {
		return OutboundHrmpMessage{}, err
	}

	data, err := sc.DecodeSequence[sc.U8](buffer)
	if err != nil {
		return OutboundHrmpMessage{}, err
	}

	return OutboundHrmpMessage{
		Id:   id,
		Data: data,
	}, nil
}

func (ohm OutboundHrmpMessage) Bytes() []byte {
	return sc.EncodedBytes(ohm)
}

func DecodeOutboundHrmpMessages(buffer *bytes.Buffer) (sc.Sequence[OutboundHrmpMessage], error) {
	return sc.DecodeSequenceWith(buffer, DecodeOutboundHrmpMessage)
}
