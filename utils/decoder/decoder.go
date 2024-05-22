package decoder

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/ChainSafe/gossamer/lib/common"
	"github.com/ChainSafe/gossamer/pkg/trie/node"
	sc "github.com/LimeChain/goscale"
)

var (
	// ErrDecodeStorageValue is defined since no sentinel error is defined
	// in the scale package.
	ErrDecodeStorageValue        = errors.New("cannot decode storage value")
	ErrDecodeHashedStorageValue  = errors.New("cannot decode hashed storage value")
	ErrDecodeHashedValueTooShort = errors.New("hashed storage value too short")
	ErrReadChildrenBitmap        = errors.New("cannot read children bitmap")
	// ErrDecodeChildHash is defined since no sentinel error is defined
	// in the scale package.
	ErrDecodeChildHash = errors.New("cannot decode child hash")
)

const (
	// ChildrenCapacity is the maximum number of children in a branch node.
	ChildrenCapacity = 16
)

// TODO: move in another folder
// DecodeNode decodes a node from a reader.
// The encoding format is documented in the README.md
// of this package, and specified in the Polkadot spec at
// https://spec.polkadot.network/#sect-state-storage
// For branch decoding, see the comments on decodeBranch.
// For leaf decoding, see the comments on decodeLeaf.
func DecodeNode(reader *bytes.Buffer) (n *node.Node, err error) {
	variant, partialKeyLength, err := decodeHeader(reader)
	if err != nil {
		return nil, fmt.Errorf("decoding header: %w", err)
	}

	switch variant {
	case emptyVariant:
		return nil, nil
	case leafVariant, leafWithHashedValueVariant:
		n, err = decodeLeaf(reader, variant, partialKeyLength)
		if err != nil {
			return nil, fmt.Errorf("cannot decode leaf: %w", err)
		}
		return n, nil
	case branchVariant, branchWithValueVariant, branchWithHashedValueVariant:
		n, err = decodeBranch(reader, variant, partialKeyLength)
		if err != nil {
			return nil, fmt.Errorf("cannot decode branch: %w", err)
		}
		return n, nil
	default:
		// this is a programming error, an unknown node variant should be caught by decodeHeader.
		panic(fmt.Sprintf("not implemented for node variant %08b", variant))
	}
}

// decodeBranch reads from a reader and decodes to a node branch.
// Note that since the encoded branch stores the hash of the children nodes, we are not
// reconstructing the child nodes from the encoding. This function instead stubs where the
// children are known to be with an empty leaf. The children nodes hashes are then used to
// find other storage values using the persistent database.
func decodeBranch(reader *bytes.Buffer, variant variant, partialKeyLength uint16) (*node.Node, error) {
	result := &node.Node{
		Children: make([]*node.Node, ChildrenCapacity),
	}

	partialKey, err := decodeKey(reader, partialKeyLength)
	if err != nil {
		return nil, fmt.Errorf("cannot decode key: %w", err)
	}
	result.PartialKey = partialKey

	childrenBitmap := make([]byte, 2)
	_, err = reader.Read(childrenBitmap)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrReadChildrenBitmap, err)
	}

	switch variant {
	case branchWithValueVariant:
		storageValue, err := sc.DecodeSequence[sc.U8](reader)
		if err != nil {
			return nil, fmt.Errorf("%w: %s", ErrDecodeStorageValue, err)
		}
		result.StorageValue = sc.SequenceU8ToBytes(storageValue)
	case branchWithHashedValueVariant:
		hashedValue, err := decodeHashedValue(reader)
		if err != nil {
			return nil, err
		}
		result.StorageValue = hashedValue
		result.IsHashedValue = true
	default:
		// Ignored
	}

	for i := 0; i < ChildrenCapacity; i++ {
		if (childrenBitmap[i/8]>>(i%8))&1 != 1 {
			continue
		}
		seqHash, err := sc.DecodeSequence[sc.U8](reader)
		if err != nil {
			return nil, fmt.Errorf("%w: at index %d: %s",
				ErrDecodeChildHash, i, err)
		}
		hash := sc.SequenceU8ToBytes(seqHash)

		childNode := &node.Node{
			MerkleValue: hash,
		}
		if len(hash) < common.HashLength {
			// Handle inlined nodes
			reader := bytes.NewBuffer(hash)
			childNode, err = DecodeNode(reader)
			if err != nil {
				return nil, fmt.Errorf("decoding inlined child at index %d: %w", i, err)
			}
			result.Descendants += childNode.Descendants
		}

		result.Descendants++
		result.Children[i] = childNode
	}

	return result, nil
}

// decodeLeaf reads from a reader and decodes to a leaf node.
func decodeLeaf(reader *bytes.Buffer, variant variant, partialKeyLength uint16) (*node.Node, error) {
	result := &node.Node{}

	partialKey, err := decodeKey(reader, partialKeyLength)
	if err != nil {
		return nil, fmt.Errorf("cannot decode key: %w", err)
	}
	result.PartialKey = partialKey

	if variant == leafWithHashedValueVariant {
		hashedValue, err := decodeHashedValue(reader)
		if err != nil {
			return nil, err
		}
		result.StorageValue = hashedValue
		result.IsHashedValue = true
		return result, nil
	}

	storageValue, err := sc.DecodeSequence[sc.U8](reader)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrDecodeStorageValue, err)
	}
	result.StorageValue = sc.SequenceU8ToBytes(storageValue)

	return result, nil
}

func decodeHashedValue(reader *bytes.Buffer) ([]byte, error) {
	buffer := make([]byte, common.HashLength)
	n, err := reader.Read(buffer)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrDecodeStorageValue, err)
	}
	if n < common.HashLength {
		return nil, fmt.Errorf("%w: expected %d, got: %d", ErrDecodeHashedValueTooShort, common.HashLength, n)
	}

	return buffer, nil
}
