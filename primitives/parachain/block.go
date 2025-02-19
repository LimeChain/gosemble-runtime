package parachain

import (
	"bytes"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/primitives/types"
)

type BlockData struct {
	Block        types.Block
	CompactProof StorageProof
}

func (bd BlockData) Encode(buffer *bytes.Buffer) error {
	return sc.EncodeEach(buffer, bd.Block, bd.CompactProof)
}

func (bd BlockData) Bytes() []byte {
	return sc.EncodedBytes(bd)
}
