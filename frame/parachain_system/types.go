package parachain_system

import (
	"bytes"
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/primitives/parachain"
)

type CollationInfo struct {
	UpwardMessages            sc.Sequence[parachain.UpwardMessage]
	HorizontalMessages        sc.Sequence[parachain.OutboundHrmpMessage]
	ValidationCode            sc.Option[sc.Sequence[sc.U8]]
	ProcessedDownwardMessages sc.U32
	HrmpWatermark             parachain.RelayChainBlockNumber
	HeadData                  parachain.HeadData
}

func (ci CollationInfo) Encode(buffer *bytes.Buffer) error {
	return sc.EncodeEach(buffer,
		ci.UpwardMessages,
		ci.HorizontalMessages,
		ci.ValidationCode,
		ci.ProcessedDownwardMessages,
		ci.HrmpWatermark,
		ci.HeadData,
	)
}

func DecodeCollationInfo(buffer *bytes.Buffer) (CollationInfo, error) {
	upwardMessages, err := parachain.DecodeUpwardMessages(buffer)
	if err != nil {
		return CollationInfo{}, err
	}
	horizontalMessages, err := parachain.DecodeOutboundHrmpMessages(buffer)
	if err != nil {
		return CollationInfo{}, err
	}
	validationCode, err := sc.DecodeOptionWith(buffer, sc.DecodeSequence[sc.U8])
	if err != nil {
		return CollationInfo{}, err
	}
	processedDownwardMessages, err := sc.DecodeU32(buffer)
	if err != nil {
		return CollationInfo{}, err
	}
	hrmpWatermark, err := sc.DecodeU32(buffer)
	if err != nil {
		return CollationInfo{}, err
	}
	headData, err := sc.DecodeSequence[sc.U8](buffer)
	if err != nil {
		return CollationInfo{}, err
	}

	return CollationInfo{
		UpwardMessages:            upwardMessages,
		HorizontalMessages:        horizontalMessages,
		ValidationCode:            validationCode,
		ProcessedDownwardMessages: processedDownwardMessages,
		HrmpWatermark:             hrmpWatermark,
		HeadData:                  headData,
	}, nil
}

func (ci CollationInfo) Bytes() []byte {
	return sc.EncodedBytes(ci)
}
