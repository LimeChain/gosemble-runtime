package parachain

import (
	"bytes"
	sc "github.com/LimeChain/goscale"
)

type HrmpChannelUpdate struct {
	MsgCount   sc.U32
	TotalBytes sc.U32
}

func (hcu HrmpChannelUpdate) Encode(buffer *bytes.Buffer) error {
	return sc.EncodeEach(buffer, hcu.MsgCount, hcu.TotalBytes)
}

func DecodeHrmpChannelUpdate(buffer *bytes.Buffer) (HrmpChannelUpdate, error) {
	msgCount, err := sc.DecodeU32(buffer)
	if err != nil {
		return HrmpChannelUpdate{}, err
	}

	totalBytes, err := sc.DecodeU32(buffer)
	if err != nil {
		return HrmpChannelUpdate{}, err
	}

	return HrmpChannelUpdate{
		MsgCount:   msgCount,
		TotalBytes: totalBytes,
	}, nil
}

func (hcu HrmpChannelUpdate) Bytes() []byte {
	return sc.EncodedBytes(hcu)
}

func (hcu *HrmpChannelUpdate) Subtract(other HrmpChannelUpdate) {
	hcu.MsgCount -= other.MsgCount
	hcu.TotalBytes -= other.TotalBytes
}

func (hcu HrmpChannelUpdate) IsEmpty() bool {
	return hcu.TotalBytes == 0 && hcu.MsgCount == 0
}
