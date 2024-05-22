package parachain

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/ChainSafe/gossamer/lib/common"
	"github.com/ChainSafe/gossamer/pkg/trie"
	"github.com/ChainSafe/gossamer/pkg/trie/db"
	"github.com/ChainSafe/gossamer/pkg/trie/node"
	"github.com/ChainSafe/gossamer/pkg/trie/pools"
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants"
	"github.com/LimeChain/gosemble/primitives/io"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/LimeChain/gosemble/utils/decoder"
	"strings"
)

var (
	errEmptyProof       = errors.New("proof slice empty")
	errRootNodeNotFound = errors.New("root node not found in proof")
)

var (
	keyCurrentSlot             = common.MustHexToBytes("0x1cb6f36e027abb2091cfb5110ab5087f06155b3cd9a8c9e5e9a23fd5dc13a5ed")
	keyPrefixGoAhead           = common.MustHexToBytes("0xcd710b30bd2eab0352ddcc26417aa1949e94c040f5e73d9b7addd6cb603d15d3")
	keyPrefixRestrictionSignal = common.MustHexToBytes("0xcd710b30bd2eab0352ddcc26417aa194f27bbb460270642b5bcaf032ea04d56a")
	keyActiveConfig            = common.MustHexToBytes("0x06de3d8a54d27e44a9d5ce189618f22db4b49d95320d9021994c850f25b8e385")
)

type RelayChainStateProof struct {
	ParachainId sc.U32
	Trie        *trie.Trie
	hashing     io.Hashing
}

func NewRelayChainStateProof(parachainId sc.U32, relayChainHash primitives.H256, proof StorageProof, hashing io.Hashing) (RelayChainStateProof, error) {
	database, err := db.NewMemoryDBFromProof(proof.ToBytes())
	if err != nil {
		return RelayChainStateProof{}, err
	}

	t, err := BuildTrie(proof.ToBytes(), relayChainHash.Bytes(), database)
	if err != nil {
		return RelayChainStateProof{}, err
	}

	return RelayChainStateProof{
		ParachainId: parachainId,
		Trie:        t,
		hashing:     hashing,
	}, nil
}

func (rlcsp RelayChainStateProof) ReadSlot() (sc.U64, error) {
	value := rlcsp.Trie.Get(keyCurrentSlot)
	if value == nil {
		return 0, NewErrorStateProofSlot(ReadEntryErrorProof)
	}

	currentSlot, err := sc.DecodeU64(bytes.NewBuffer(value))
	if err != nil {
		return 0, NewErrorStateProofSlot(ReadEntryErrorDecode)
	}

	return currentSlot, nil
}

func (rlcsp RelayChainStateProof) ReadUpgradeGoAheadSignal() (sc.Option[sc.U8], error) {
	hashParachainId := rlcsp.hashing.Twox64(rlcsp.ParachainId.Bytes())

	key := append(keyPrefixGoAhead, hashParachainId...)
	key = append(key, rlcsp.ParachainId.Bytes()...)

	value := rlcsp.Trie.Get(key)
	if value == nil {
		return sc.NewOption[sc.U8](nil), nil
	}

	goAhead, err := DecodeUpgradeGoAhead(bytes.NewBuffer(value))
	if err != nil {
		return sc.Option[sc.U8]{}, NewErrorStateProofUpgradeGoAhead(ReadEntryErrorDecode)
	}

	return sc.NewOption[sc.U8](goAhead), nil
}

func (rlcsp RelayChainStateProof) ReadRestrictionSignal() (sc.Option[sc.U8], error) {
	hashParachainId := rlcsp.hashing.Twox64(rlcsp.ParachainId.Bytes())

	key := append(keyPrefixRestrictionSignal, hashParachainId...)
	key = append(key, rlcsp.ParachainId.Bytes()...)

	value := rlcsp.Trie.Get(key)
	if value == nil {
		return sc.NewOption[sc.U8](nil), nil
	}

	goAhead, err := DecodeUpgradeRestrictionSignal(bytes.NewBuffer(value))
	if err != nil {
		return sc.Option[sc.U8]{}, NewErrorStateProofUpgradeRestriction(ReadEntryErrorDecode)
	}

	return sc.NewOption[sc.U8](goAhead), nil
}

func (rlcsp RelayChainStateProof) ReadAbridgedHostConfiguration() (AbridgedHostConfiguration, error) {
	value := rlcsp.Trie.Get(keyActiveConfig)
	if value == nil {
		return AbridgedHostConfiguration{}, NewErrorStateProofConfig(ReadEntryErrorProof)
	}

	ahc, err := DecodeAbridgeHostConfiguration(bytes.NewBuffer(value))
	if err != nil {
		return AbridgedHostConfiguration{}, NewErrorStateProofConfig(ReadEntryErrorDecode)
	}

	return ahc, nil
}

func (rlcsp RelayChainStateProof) ReadMessagingStateSnapshot(ahc AbridgedHostConfiguration) (MessagingStateSnapshot, error) {
	// TODO:
	return MessagingStateSnapshot{
		DmqMqcHead:                          primitives.H256{FixedSequence: constants.ZeroAccountId.FixedSequence},
		RelayDispatchQueueRemainingCapacity: RelayDispatchQueueRemainingCapacity{},
		IngressChannels:                     nil,
		EgressChannels:                      nil,
	}, nil
}

func BuildTrie(encodedProofNodes [][]byte, rootHash []byte, db db.Database) (t *trie.Trie, err error) {
	if len(encodedProofNodes) == 0 {
		return nil, fmt.Errorf("%w: for Merkle root hash 0x%x",
			errEmptyProof, rootHash)
	}

	digestToEncoding := make(map[string][]byte, len(encodedProofNodes))

	// note we can use a buffer from the pool since
	// the calculated root hash digest is not used after
	// the function completes.
	buffer := pools.DigestBuffers.Get().(*bytes.Buffer)
	defer pools.DigestBuffers.Put(buffer)

	// This loop does two things:
	// 1. It finds the root node by comparing it with the root hash and decodes it.
	// 2. It stores other encoded nodes in a mapping from their encoding digest to
	//    their encoding. They are only decoded later if the root or one of its
	//    descendant nodes reference their hash digest.
	var root *node.Node
	for _, encodedProofNode := range encodedProofNodes {
		// Note all encoded proof nodes are one of the following:
		// - trie root node
		// - child trie root node
		// - child node with an encoding larger than 32 bytes
		// In all cases, their Merkle value is the encoding hash digest,
		// so we use MerkleValueRoot to force hashing the node in case
		// it is a root node smaller or equal to 32 bytes.
		buffer.Reset()
		err = node.MerkleValueRoot(encodedProofNode, buffer)
		if err != nil {
			return nil, fmt.Errorf("calculating node hash: %w", err)
		}
		digest := buffer.Bytes()

		if root != nil || !bytes.Equal(digest, rootHash) {
			// root node already found or the hash doesn't match the root hash.
			digestToEncoding[string(digest)] = encodedProofNode
			continue
			// Note: no need to add the root node to the map of hash to encoding
		}

		root, err = decoder.DecodeNode(bytes.NewBuffer(encodedProofNode))
		//root, err = decoder.DecodeNode(bytes.NewBuffer(encodedProofNode))
		if err != nil {
			return nil, fmt.Errorf("decoding root node: %w", err)
		}
		// The built proof trie is not used with a database, but just in case
		// it becomes used with a database in the future, we set the dirty flag
		// to true.
		root.Dirty = true
	}

	if root == nil {
		proofHashDigests := make([]string, 0, len(digestToEncoding))
		for hashDigestString := range digestToEncoding {
			hashDigestHex := common.BytesToHex([]byte(hashDigestString))
			proofHashDigests = append(proofHashDigests, hashDigestHex)
		}
		return nil, fmt.Errorf("%w: for root hash 0x%x in proof hash digests %s",
			errRootNodeNotFound, rootHash, strings.Join(proofHashDigests, ", "))
	}

	err = loadProof(digestToEncoding, root)
	if err != nil {
		return nil, fmt.Errorf("loading proof: %w", err)
	}

	return trie.NewTrie(root, db), nil
}

// loadProof is a recursive function that will create all the trie paths based
// on the map from node hash digest to node encoding, starting from the node `n`.
func loadProof(digestToEncoding map[string][]byte, n *node.Node) (err error) {
	if n.Kind() != node.Branch {
		return nil
	}

	branch := n
	for i, child := range branch.Children {
		if child == nil {
			continue
		}

		merkleValue := child.MerkleValue
		encoding, ok := digestToEncoding[string(merkleValue)]

		if !ok {
			inlinedChild := len(child.StorageValue) > 0 || child.HasChild()
			if inlinedChild {
				// The built proof trie is not used with a database, but just in case
				// it becomes used with a database in the future, we set the dirty flag
				// to true.
				child.Dirty = true
			} else {
				// hash not found and the child is not inlined,
				// so clear the child from the branch.
				branch.Descendants -= 1 + child.Descendants
				branch.Children[i] = nil
				if !branch.HasChild() {
					// Convert branch to a leaf if all its children are nil.
					branch.Children = nil
				}
			}
			continue
		}

		child, err := decoder.DecodeNode(bytes.NewBuffer(encoding))
		//child, err := decoder.DecodeNode(bytes.NewBuffer(encoding))
		if err != nil {
			return fmt.Errorf("decoding child node for hash digest 0x%x: %w",
				merkleValue, err)
		}

		// The built proof trie is not used with a database, but just in case
		// it becomes used with a database in the future, we set the dirty flag
		// to true.
		child.Dirty = true

		branch.Children[i] = child
		branch.Descendants += child.Descendants
		err = loadProof(digestToEncoding, child)
		if err != nil {
			return err // do not wrap error since this is recursive
		}
	}

	return nil
}
