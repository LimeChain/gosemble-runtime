package parachain

import (
	"bytes"
	sc "github.com/LimeChain/goscale"
)

type ValidationResult struct {
	HeadData                  sc.Sequence[sc.U8]
	NewValidationCode         sc.Option[sc.Sequence[sc.U8]]
	UpwardMessages            sc.Sequence[UpwardMessage]       // Convert to bounded vec (aka to have max limit, max 16 * 1024)
	HorizontalMessages        sc.Sequence[OutboundHrmpMessage] // Convert to bounded vec (aka to have max limit, max 16 * 1024)
	ProcessedDownwardMessages sc.U32
	HrmpWatermark             RelayChainBlockNumber
}

func (vr ValidationResult) Encode(buffer *bytes.Buffer) error {
	return sc.EncodeEach(buffer,
		vr.HeadData,
		vr.NewValidationCode,
		vr.UpwardMessages,
		vr.HorizontalMessages,
		vr.ProcessedDownwardMessages,
		vr.HrmpWatermark,
	)
}

func DecodeValidationResult(buffer *bytes.Buffer) (ValidationResult, error) {
	headData, err := sc.DecodeSequence[sc.U8](buffer)
	if err != nil {
		return ValidationResult{}, err
	}

	newValidationCode, err := sc.DecodeOptionWith(buffer, sc.DecodeSequence[sc.U8])
	if err != nil {
		return ValidationResult{}, err
	}

	upwardMessages, err := sc.DecodeSequenceWith(buffer, sc.DecodeSequence[sc.U8])
	if err != nil {
		return ValidationResult{}, err
	}

	horizontalMessages, err := sc.DecodeSequenceWith(buffer, DecodeOutboundHrmpMessage)
	if err != nil {
		return ValidationResult{}, err
	}

	processedDownwawrdMessages, err := sc.DecodeU32(buffer)
	if err != nil {
		return ValidationResult{}, err
	}

	hrmpWatermark, err := sc.DecodeU32(buffer)
	if err != nil {
		return ValidationResult{}, err
	}

	return ValidationResult{
		HeadData:                  headData,
		NewValidationCode:         newValidationCode,
		UpwardMessages:            upwardMessages,
		HorizontalMessages:        horizontalMessages,
		ProcessedDownwardMessages: processedDownwawrdMessages,
		HrmpWatermark:             hrmpWatermark,
	}, nil
}

func (vr ValidationResult) Bytes() []byte {
	return sc.EncodedBytes(vr)
}
