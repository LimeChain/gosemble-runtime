package system

import (
	"bytes"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/frame/support"
	"github.com/LimeChain/gosemble/primitives/io"
	"github.com/LimeChain/gosemble/primitives/types"
)

var (
	keySystem             = []byte("System")
	keyAccount            = []byte("Account")
	keyAllExtrinsicsLen   = []byte("AllExtrinsicsLen")
	keyBlockHash          = []byte("BlockHash")
	keyBlockWeight        = []byte("BlockWeight")
	keyCode               = []byte(":code")
	keyAuthorizedUpgrade  = []byte("AuthorizedUpgrade")
	keyDigest             = []byte("Digest")
	keyEventCount         = []byte("EventCount")
	keyEvents             = []byte("Events")
	keyEventTopics        = []byte("EventTopics")
	keyExecutionPhase     = []byte("ExecutionPhase")
	keyExtrinsicCount     = []byte("ExtrinsicCount")
	keyExtrinsicData      = []byte("ExtrinsicData")
	keyExtrinsicIndex     = []byte(":extrinsic_index")
	keyHeapPages          = []byte(":heappages")
	keyLastRuntimeUpgrade = []byte("LastRuntimeUpgrade")
	keyNumber             = []byte("Number")
	keyParentHash         = []byte("ParentHash")
)

type storage struct {
	Account            support.StorageMap[types.AccountId, types.AccountInfo]
	BlockWeight        support.StorageValue[types.ConsumedWeight]
	BlockHash          support.StorageMap[sc.U64, types.Blake2bHash]
	BlockNumber        support.StorageValue[sc.U64]
	AllExtrinsicsLen   support.StorageValue[sc.U32]
	ExtrinsicIndex     support.StorageValue[sc.U32]
	ExtrinsicData      support.StorageMap[sc.U32, sc.Sequence[sc.U8]]
	ExtrinsicCount     support.StorageValue[sc.U32]
	ParentHash         support.StorageValue[types.Blake2bHash]
	Digest             support.StorageValue[types.Digest]
	Events             support.StorageValue[types.EventRecord] // This calls only Append and Kill
	EventCount         support.StorageValue[sc.U32]
	EventTopics        support.StorageMap[types.H256, sc.VaryingData]
	LastRuntimeUpgrade support.StorageValue[types.LastRuntimeUpgradeInfo]
	ExecutionPhase     support.StorageValue[types.ExtrinsicPhase]
	HeapPages          support.StorageValue[sc.U64]
	Code               support.StorageRawValue
	AuthorizedUpgrade  support.StorageValue[CodeUpgradeAuthorization]
}

func newStorage(s io.Storage) *storage {
	hashing := io.NewHashing()

	return &storage{
		Account:            support.NewHashStorageMap[types.AccountId](s, keySystem, keyAccount, hashing.Blake128, types.DecodeAccountInfo),
		BlockWeight:        support.NewHashStorageValue(s, keySystem, keyBlockWeight, types.DecodeConsumedWeight),
		BlockHash:          support.NewHashStorageMap[sc.U64, types.Blake2bHash](s, keySystem, keyBlockHash, hashing.Twox64, types.DecodeBlake2bHash),
		BlockNumber:        support.NewHashStorageValue(s, keySystem, keyNumber, sc.DecodeU64),
		AllExtrinsicsLen:   support.NewHashStorageValue(s, keySystem, keyAllExtrinsicsLen, sc.DecodeU32),
		ExtrinsicIndex:     support.NewSimpleStorageValue(s, keyExtrinsicIndex, sc.DecodeU32),
		ExtrinsicData:      support.NewHashStorageMap[sc.U32, sc.Sequence[sc.U8]](s, keySystem, keyExtrinsicData, hashing.Twox64, sc.DecodeSequence[sc.U8]),
		ExtrinsicCount:     support.NewHashStorageValue(s, keySystem, keyExtrinsicCount, sc.DecodeU32),
		ParentHash:         support.NewHashStorageValue(s, keySystem, keyParentHash, types.DecodeBlake2bHash),
		Digest:             support.NewHashStorageValue(s, keySystem, keyDigest, types.DecodeDigest),
		Events:             support.NewHashStorageValue(s, keySystem, keyEvents, func(*bytes.Buffer) (types.EventRecord, error) { return types.EventRecord{}, nil }),
		EventCount:         support.NewHashStorageValue(s, keySystem, keyEventCount, sc.DecodeU32),
		EventTopics:        support.NewHashStorageMap[types.H256, sc.VaryingData](s, keySystem, keyEventTopics, hashing.Blake128, func(buffer *bytes.Buffer) (sc.VaryingData, error) { return sc.NewVaryingData(), nil }),
		LastRuntimeUpgrade: support.NewHashStorageValue(s, keySystem, keyLastRuntimeUpgrade, types.DecodeLastRuntimeUpgradeInfo),
		ExecutionPhase:     support.NewHashStorageValue(s, keySystem, keyExecutionPhase, types.DecodeExtrinsicPhase),
		HeapPages:          support.NewSimpleStorageValue(s, keyHeapPages, sc.DecodeU64),
		Code:               support.NewRawStorageValue(s, keyCode),
		AuthorizedUpgrade:  support.NewHashStorageValue(s, keySystem, keyAuthorizedUpgrade, DecodeCodeUpgradeAuthorization),
	}
}
