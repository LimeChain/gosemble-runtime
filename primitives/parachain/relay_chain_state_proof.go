package parachain

import (
	"bytes"
	"errors"

	"github.com/ChainSafe/gossamer/lib/common"
	"github.com/ChainSafe/gossamer/pkg/trie/db"
	"github.com/ChainSafe/gossamer/pkg/trie/inmemory"
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants"
	"github.com/LimeChain/gosemble/primitives/io"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/LimeChain/gosemble/utils/decoder"
)

var (
	errRootNodeNotFound = errors.New("root node not found in proof")
)

var (
	keyCurrentSlot             = common.MustHexToBytes("0x1cb6f36e027abb2091cfb5110ab5087f06155b3cd9a8c9e5e9a23fd5dc13a5ed")
	keyPrefixDmqMqcHead        = common.MustHexToBytes("0x63f78c98723ddc9073523ef3beefda0c4d7fefc408aac59dbfe80a72ac8e3ce5")
	keyPrefixGoAhead           = common.MustHexToBytes("0xcd710b30bd2eab0352ddcc26417aa1949e94c040f5e73d9b7addd6cb603d15d3")
	keyPrefixParaHead          = common.MustHexToBytes("0xcd710b30bd2eab0352ddcc26417aa1941b3c252fcb29d88eff4f3de5de4476c3")
	keyPrefixRestrictionSignal = common.MustHexToBytes("0xcd710b30bd2eab0352ddcc26417aa194f27bbb460270642b5bcaf032ea04d56a")
	keyActiveConfig            = common.MustHexToBytes("0x06de3d8a54d27e44a9d5ce189618f22db4b49d95320d9021994c850f25b8e385")
)

type RelayChainStateProof interface {
	ReadSlot() (sc.U64, error)
	ReadUpgradeGoAheadSignal() (sc.Option[sc.U8], error)
	ReadRestrictionSignal() (sc.Option[sc.U8], error)
	ReadAbridgedHostConfiguration() (AbridgedHostConfiguration, error)
	ReadIncludedParaHeadHash() sc.Option[sc.FixedSequence[sc.U8]]
	ReadMessagingStateSnapshot(ahc AbridgedHostConfiguration) (MessagingStateSnapshot, error)
}

type relayChainStateProof struct {
	ParachainId sc.U32
	Trie        *inmemory.InMemoryTrie
	hashing     io.Hashing
}

func NewRelayChainStateProof(parachainId sc.U32, relayChainHash primitives.H256, proof StorageProof, hashing io.Hashing) (RelayChainStateProof, error) {
	database, err := db.NewMemoryDBFromProof(proof.ToBytes())
	if err != nil {
		return relayChainStateProof{}, err
	}

	trie, err := BuildTrie(relayChainHash.Bytes(), database)
	if err != nil {
		return relayChainStateProof{}, err
	}

	return relayChainStateProof{
		ParachainId: parachainId,
		Trie:        trie,
		hashing:     hashing,
	}, nil
}

func (rlcsp relayChainStateProof) ReadSlot() (sc.U64, error) {
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

func (rlcsp relayChainStateProof) ReadUpgradeGoAheadSignal() (sc.Option[sc.U8], error) {
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

func (rlcsp relayChainStateProof) ReadRestrictionSignal() (sc.Option[sc.U8], error) {
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

func (rlcsp relayChainStateProof) ReadAbridgedHostConfiguration() (AbridgedHostConfiguration, error) {
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

func (rlcsp relayChainStateProof) ReadIncludedParaHeadHash() sc.Option[sc.FixedSequence[sc.U8]] {
	hashParachainId := rlcsp.hashing.Twox64(rlcsp.ParachainId.Bytes())

	key := append(keyPrefixParaHead, hashParachainId...)
	key = append(key, rlcsp.ParachainId.Bytes()...)

	value := rlcsp.Trie.Get(key)
	if value == nil {
		return sc.NewOption[sc.FixedSequence[sc.U8]](nil)
	}

	paraHead, err := sc.DecodeSequence[sc.U8](bytes.NewBuffer(value))
	if err != nil {
		return sc.NewOption[sc.FixedSequence[sc.U8]](nil)
	}

	paraHeadHash := rlcsp.hashing.Blake256(sc.SequenceU8ToBytes(paraHead))

	return sc.NewOption[sc.FixedSequence[sc.U8]](sc.BytesToFixedSequenceU8(paraHeadHash))
}

func (rlcsp relayChainStateProof) ReadMessagingStateSnapshot(ahc AbridgedHostConfiguration) (MessagingStateSnapshot, error) {
	// TODO: read and populate messaging state snapshot from the state proof.

	return MessagingStateSnapshot{
		DmqMqcHead:                          primitives.H256{FixedSequence: constants.ZeroAccountId.FixedSequence},
		RelayDispatchQueueRemainingCapacity: RelayDispatchQueueRemainingCapacity{},
		IngressChannels:                     nil,
		EgressChannels:                      nil,
	}, nil
}

// BuildTrie sets a partial trie based on the proof slice of encoded nodes.
func BuildTrie(rootHash []byte, db db.Database) (t *inmemory.InMemoryTrie, err error) {
	// buildTrie sets a partial trie based on the proof slice of encoded nodes.
	if _, err := db.Get(rootHash); err != nil {
		return nil, NewErrorStateProofRootMismatch()
	}

	tr := inmemory.NewEmptyTrie()
	err = tr.LoadWithDecoder(db, common.BytesToHash(rootHash), decoder.DecodeNode)

	if err != nil {
		return nil, err
	}

	return tr, nil
}
