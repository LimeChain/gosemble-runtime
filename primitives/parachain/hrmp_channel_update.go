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
