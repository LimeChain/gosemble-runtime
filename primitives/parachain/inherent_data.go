package parachain

import (
	"bytes"

	sc "github.com/LimeChain/goscale"
)

type InherentData struct {
	ValidationData     PersistedValidationData
	RelayChainState    StorageProof
	DownwardMessages   sc.Sequence[InboundDownwardMessage]
	HorizontalMessages HorizontalMessages
}

func (id InherentData) Encode(buffer *bytes.Buffer) error {
	return sc.EncodeEach(buffer,
		id.ValidationData,
		id.RelayChainState,
		id.DownwardMessages,
		id.HorizontalMessages)
}

func DecodeInherentData(buffer *bytes.Buffer) (InherentData, error) {
	validationData, err := DecodePersistedValidationData(buffer)
	if err != nil {
		return InherentData{}, err
	}

	storageProof, err := DecodeStorageProof(buffer)
	if err != nil {
		return InherentData{}, err
	}

	downwardMessages, err := sc.DecodeSequenceWith(buffer, DecodeInboundDownwardMessage)
	if err != nil {
		return InherentData{}, err
	}

	horizontalMessages, err := DecodeHorizontalMessages(buffer)
	if err != nil {
		return InherentData{}, err
	}

	return InherentData{
		ValidationData:     validationData,
		RelayChainState:    storageProof,
		DownwardMessages:   downwardMessages,
		HorizontalMessages: horizontalMessages,
	}, nil
}

func (id InherentData) Bytes() []byte {
	return sc.EncodedBytes(id)
}
