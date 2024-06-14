package parachain

import (
	"bytes"

	sc "github.com/LimeChain/goscale"
)

type DownwardMessage = sc.Sequence[sc.U8]

type InboundDownwardMessage struct {
	SentAt RelayChainBlockNumber
	Msg    DownwardMessage
}

func (idm InboundDownwardMessage) Encode(buffer *bytes.Buffer) error {
	return sc.EncodeEach(buffer, idm.SentAt, idm.Msg)
}

func DecodeInboundDownwardMessage(buffer *bytes.Buffer) (InboundDownwardMessage, error) {
	sentAt, err := sc.DecodeU32(buffer)
	if err != nil {
		return InboundDownwardMessage{}, err
	}

	msg, err := sc.DecodeSequence[sc.U8](buffer)
	if err != nil {
		return InboundDownwardMessage{}, err
	}

	return InboundDownwardMessage{
		SentAt: sentAt,
		Msg:    msg,
	}, nil
}

func (idm InboundDownwardMessage) Bytes() []byte {
	return sc.EncodedBytes(idm)
}
