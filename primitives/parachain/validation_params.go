package parachain

import (
	"bytes"
	sc "github.com/LimeChain/goscale"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

type ValidationParams struct {
	ParentHead       sc.Sequence[sc.U8]
	BlockData        sc.Sequence[sc.U8]
	RelayBlockNumber RelayChainBlockNumber
	RelayStorageRoot primitives.H256
}

func (vp ValidationParams) Encode(buffer *bytes.Buffer) error {
	return sc.EncodeEach(buffer,
		vp.ParentHead,
		vp.BlockData,
		vp.RelayBlockNumber,
		vp.RelayStorageRoot,
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
	relayBlockNumber, err := sc.DecodeU32(buffer)
	if err != nil {
		return ValidationParams{}, err
	}
	relayStorageRoot, err := primitives.DecodeH256(buffer)
	if err != nil {
		return ValidationParams{}, err
	}

	return ValidationParams{
		ParentHead:       parentHead,
		BlockData:        blockData,
		RelayBlockNumber: relayBlockNumber,
		RelayStorageRoot: relayStorageRoot,
	}, nil
}

func (vp ValidationParams) Bytes() []byte {
	return sc.EncodedBytes(vp)
}
