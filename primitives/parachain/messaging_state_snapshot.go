package parachain

import (
	"bytes"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/primitives/types"
)

type MessagingStateSnapshot struct {
	DmqMqcHead                          types.H256
	RelayDispatchQueueRemainingCapacity RelayDispatchQueueRemainingCapacity
	IngressChannels                     sc.Sequence[Channel]
	EgressChannels                      sc.Sequence[Channel]
}

func (mss MessagingStateSnapshot) Encode(buffer *bytes.Buffer) error {
	return sc.EncodeEach(buffer,
		mss.DmqMqcHead,
		mss.RelayDispatchQueueRemainingCapacity,
		mss.IngressChannels,
		mss.EgressChannels,
	)
}

func DecodeMessagingStateSnapshot(buffer *bytes.Buffer) (MessagingStateSnapshot, error) {
	dmqMqcHead, err := types.DecodeH256(buffer)
	if err != nil {
		return MessagingStateSnapshot{}, err
	}

	relayDispatchQueueRemainingCapacity, err := DecodeRelayDispatchQueueRemainingCapacity(buffer)
	if err != nil {
		return MessagingStateSnapshot{}, err
	}

	ingressChannels, err := sc.DecodeSequenceWith(buffer, DecodeChannel)
	if err != nil {
		return MessagingStateSnapshot{}, err
	}

	egressChannels, err := sc.DecodeSequenceWith(buffer, DecodeChannel)
	if err != nil {
		return MessagingStateSnapshot{}, err
	}

	return MessagingStateSnapshot{
		DmqMqcHead:                          dmqMqcHead,
		RelayDispatchQueueRemainingCapacity: relayDispatchQueueRemainingCapacity,
		IngressChannels:                     ingressChannels,
		EgressChannels:                      egressChannels,
	}, nil
}

func (mss MessagingStateSnapshot) Bytes() []byte {
	return sc.EncodedBytes(mss)
}
