package parachain

import (
	"bytes"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/primitives/types"
)

type AbridgedHRMPChannel struct {
	MaxCapacity    sc.U32
	MaxTotalSize   sc.U32
	MaxMessageSize sc.U32
	MsgCount       sc.U32
	TotalSize      sc.U32
	MqcHead        sc.Option[types.H256]
}

func (ahc AbridgedHRMPChannel) Encode(buffer *bytes.Buffer) error {
	return sc.EncodeEach(buffer,
		ahc.MaxCapacity,
		ahc.MaxTotalSize,
		ahc.MaxMessageSize,
		ahc.MsgCount,
		ahc.TotalSize,
		ahc.MqcHead,
	)
}

func DecodeAbridgedHRMPChannel(buffer *bytes.Buffer) (AbridgedHRMPChannel, error) {
	maxCapacity, err := sc.DecodeU32(buffer)
	if err != nil {
		return AbridgedHRMPChannel{}, err
	}

	maxTotalSize, err := sc.DecodeU32(buffer)
	if err != nil {
		return AbridgedHRMPChannel{}, err
	}

	maxMessageSize, err := sc.DecodeU32(buffer)
	if err != nil {
		return AbridgedHRMPChannel{}, err
	}

	msgCount, err := sc.DecodeU32(buffer)
	if err != nil {
		return AbridgedHRMPChannel{}, err
	}

	totalSize, err := sc.DecodeU32(buffer)
	if err != nil {
		return AbridgedHRMPChannel{}, err
	}

	mqcHead, err := sc.DecodeOptionWith(buffer, types.DecodeH256)
	if err != nil {
		return AbridgedHRMPChannel{}, err
	}

	return AbridgedHRMPChannel{
		MaxCapacity:    maxCapacity,
		MaxTotalSize:   maxTotalSize,
		MaxMessageSize: maxMessageSize,
		MsgCount:       msgCount,
		TotalSize:      totalSize,
		MqcHead:        mqcHead,
	}, nil
}

func (ahc AbridgedHRMPChannel) Bytes() []byte {
	return sc.EncodedBytes(ahc)
}
