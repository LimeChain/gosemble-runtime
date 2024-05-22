package parachain

import (
	"bytes"
	sc "github.com/LimeChain/goscale"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

type ValidationParams struct {
	ParentHead             sc.Sequence[sc.U8]
	BlockData              sc.Sequence[sc.U8]
	RelayParentBlockNumber RelayChainBlockNumber
	RelayParentStorageRoot primitives.H256
}

func (vp ValidationParams) Encode(buffer *bytes.Buffer) error {
	return sc.EncodeEach(buffer,
		vp.ParentHead,
		vp.BlockData,
		vp.RelayParentBlockNumber,
		vp.RelayParentStorageRoot,
	)
}

func DecodeValidationParams(buffer *bytes.Buffer) (ValidationParams, error) {
	parentHead, err := sc.DecodeSequence[sc.U8](buffer)
	if err != nil {
		return ValidationParams{}, err
	}
	blockData, err := sc.DecodeSequence[sc.U8](buffer)
	if err != nil {
		return ValidationParams{}, err
	}
	relayParentBlockNumber, err := sc.DecodeU32(buffer)
	if err != nil {
		return ValidationParams{}, err
	}
	relayParentStorageRoot, err := primitives.DecodeH256(buffer)
	if err != nil {
		return ValidationParams{}, err
	}

	return ValidationParams{
		ParentHead:             parentHead,
		BlockData:              blockData,
		RelayParentBlockNumber: relayParentBlockNumber,
		RelayParentStorageRoot: relayParentStorageRoot,
	}, nil
}

func (vp ValidationParams) Bytes() []byte {
	return sc.EncodedBytes(vp)
}
