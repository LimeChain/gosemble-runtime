package parachain

import (
	"bytes"

	sc "github.com/LimeChain/goscale"
)

type RelayDispatchQueueRemainingCapacity struct {
	RemainingCount sc.U32
	RemainingSize  sc.U32
}

func (rdqrc RelayDispatchQueueRemainingCapacity) Encode(buffer *bytes.Buffer) error {
	return sc.EncodeEach(buffer, rdqrc.RemainingCount, rdqrc.RemainingSize)
}

func DecodeRelayDispatchQueueRemainingCapacity(buffer *bytes.Buffer) (RelayDispatchQueueRemainingCapacity, error) {
	remainingCount, err := sc.DecodeU32(buffer)
	if err != nil {
		return RelayDispatchQueueRemainingCapacity{}, err
	}

	remainingSize, err := sc.DecodeU32(buffer)
	if err != nil {
		return RelayDispatchQueueRemainingCapacity{}, err
	}

	return RelayDispatchQueueRemainingCapacity{
		RemainingCount: remainingCount,
		RemainingSize:  remainingSize,
	}, nil
}

func (rdqrc RelayDispatchQueueRemainingCapacity) Bytes() []byte {
	return sc.EncodedBytes(rdqrc)
}
