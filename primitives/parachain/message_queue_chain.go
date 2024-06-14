package parachain

import (
	"bytes"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/primitives/io"
	"github.com/LimeChain/gosemble/primitives/types"
)

type MessageQueueChain struct {
	RelayHash types.H256
}

func (mqc MessageQueueChain) Encode(buffer *bytes.Buffer) error {
	return sc.EncodeEach(buffer, mqc.RelayHash)
}

func DecodeMessageQueueChain(buffer *bytes.Buffer) (MessageQueueChain, error) {
	relayHash, err := types.DecodeH256(buffer)
	if err != nil {
		return MessageQueueChain{}, err
	}

	return MessageQueueChain{relayHash}, nil
}

func (mqc MessageQueueChain) Bytes() []byte {
	return sc.EncodedBytes(mqc)
}

func (mqc *MessageQueueChain) ExtendDownward(downwardMessage InboundDownwardMessage, hashing io.Hashing) error {
	prev := mqc.RelayHash
	payload := append(prev.Bytes(), downwardMessage.SentAt.Bytes()...)
	payload = append(payload, hashing.Blake256(downwardMessage.Msg.Bytes())...)

	newHash := hashing.Blake256(payload)

	h256, err := types.NewH256(sc.BytesToFixedSequenceU8(newHash)...)
	if err != nil {
		return err
	}
	mqc.RelayHash = h256

	return nil
}
