package parachain

import (
	"bytes"

	sc "github.com/LimeChain/goscale"
)

type HorizontalMessages struct {
	messages sc.Dictionary[sc.U32, sc.Sequence[InboundDownwardMessage]]
}

func (hm HorizontalMessages) Encode(buffer *bytes.Buffer) error {
	return sc.EncodeEach(buffer, hm.messages)
}

func DecodeHorizontalMessages(buffer *bytes.Buffer) (HorizontalMessages, error) {
	v, err := sc.DecodeCompact[sc.U128](buffer)
	if err != nil {
		return HorizontalMessages{}, err
	}
	size := int(v.ToBigInt().Int64())

	result := sc.Dictionary[sc.U32, sc.Sequence[InboundDownwardMessage]]{}

	for i := 0; i < size; i++ {
		key, err := sc.DecodeU32(buffer)
		if err != nil {
			return HorizontalMessages{}, err
		}
		value, err := sc.DecodeSequenceWith(buffer, DecodeInboundDownwardMessage)
		if err != nil {
			return HorizontalMessages{}, err
		}

		result[key] = value
	}

	return HorizontalMessages{
		messages: result,
	}, nil
}

func (hm HorizontalMessages) Bytes() []byte {
	return sc.EncodedBytes(hm)
}

func (hm HorizontalMessages) UnprocessedMessages(number RelayChainBlockNumber) HorizontalMessages {
	messages := sc.Dictionary[sc.U32, sc.Sequence[InboundDownwardMessage]]{}
	for paraChainId, horizontalMessages := range hm.messages {
		for _, horizontalMessage := range horizontalMessages {
			resultMessages := sc.Sequence[InboundDownwardMessage]{}
			if horizontalMessage.SentAt > number {
				resultMessages = append(resultMessages, horizontalMessage)
			}

			messages[paraChainId] = resultMessages
		}
	}

	return HorizontalMessages{messages: messages}
}
