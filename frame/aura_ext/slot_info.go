package aura_ext

import (
	"bytes"
	sc "github.com/LimeChain/goscale"
)

type SlotInfo struct {
	Slot     sc.U64
	Authored sc.U32
}

func (si SlotInfo) Encode(buffer *bytes.Buffer) error {
	return sc.EncodeEach(buffer, si.Slot, si.Authored)
}

func DecodeSlotInfo(buffer *bytes.Buffer) (SlotInfo, error) {
	slot, err := sc.DecodeU64(buffer)
	if err != nil {
		return SlotInfo{}, err
	}

	authored, err := sc.DecodeU32(buffer)
	if err != nil {
		return SlotInfo{}, err
	}

	return SlotInfo{
		Slot:     slot,
		Authored: authored,
	}, nil
}

func (si SlotInfo) Bytes() []byte {
	return sc.EncodedBytes(si)
}
