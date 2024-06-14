package parachain

import (
	"bytes"

	sc "github.com/LimeChain/goscale"
)

type InboundHrmpMessage struct {
	SentAt RelayChainBlockNumber
	Data   sc.Sequence[sc.U8]
}

func (ihm InboundHrmpMessage) Encode(buffer *bytes.Buffer) error {
	return sc.EncodeEach(buffer, ihm.SentAt, ihm.Data)
}

func (ihm InboundHrmpMessage) Bytes() []byte {
	return sc.EncodedBytes(ihm)
}
