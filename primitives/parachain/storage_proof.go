package parachain

import (
	"bytes"

	sc "github.com/LimeChain/goscale"
)

type StorageProof struct {
	TrieNodes sc.Sequence[sc.Sequence[sc.U8]]
}

func (sp StorageProof) Encode(buffer *bytes.Buffer) error {
	return sc.EncodeEach(buffer, sp.TrieNodes)
}

func DecodeStorageProof(buffer *bytes.Buffer) (StorageProof, error) {
	trieNodes, err := sc.DecodeSequenceWith(buffer, sc.DecodeSequence[sc.U8])
	if err != nil {
		return StorageProof{}, err
	}

	return StorageProof{trieNodes}, nil
}

func (sp StorageProof) Bytes() []byte {
	return sc.EncodedBytes(sp)
}

func (sp StorageProof) ToBytes() [][]byte {
	var result [][]byte

	for _, trieNode := range sp.TrieNodes {
		result = append(result, sc.SequenceU8ToBytes(trieNode))
	}

	return result
}
