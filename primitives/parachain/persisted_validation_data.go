package parachain

import (
	"bytes"
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/primitives/types"
)

type PersistedValidationData struct {
	ParentHead             HeadData
	RelayParentNumber      RelayChainBlockNumber
	RelayParentStorageRoot types.H256
	MaxPovSize             sc.U32
}

func (pvd PersistedValidationData) Encode(buffer *bytes.Buffer) error {
	return sc.EncodeEach(buffer,
		pvd.ParentHead,
		pvd.RelayParentNumber,
		pvd.RelayParentStorageRoot,
		pvd.MaxPovSize)
}

func DecodePersistedValidationData(buffer *bytes.Buffer) (PersistedValidationData, error) {
	parentHead, err := sc.DecodeSequence[sc.U8](buffer)
	if err != nil {
		return PersistedValidationData{}, err
	}

	relayParentNumber, err := sc.DecodeU32(buffer)
	if err != nil {
		return PersistedValidationData{}, err
	}

	relayParentStorageRoot, err := types.DecodeH256(buffer)
	if err != nil {
		return PersistedValidationData{}, err
	}

	maxPovSize, err := sc.DecodeU32(buffer)
	if err != nil {
		return PersistedValidationData{}, err
	}

	return PersistedValidationData{
		parentHead,
		relayParentNumber,
		relayParentStorageRoot,
		maxPovSize,
	}, nil
}

func (pvd PersistedValidationData) Bytes() []byte {
	return sc.EncodedBytes(pvd)
}
