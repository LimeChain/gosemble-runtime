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

func newStorage() *storage {
	hashing := io.NewHashing()
	defaultAccountInfo := types.DefaultAccountInfo()

	return &storage{
		Account:            support.NewHashStorageMapWithDefault[types.AccountId](keySystem, keyAccount, hashing.Blake128, types.DecodeAccountInfo, &defaultAccountInfo),
		BlockWeight:        support.NewHashStorageValue(keySystem, keyBlockWeight, types.DecodeConsumedWeight),
		BlockHash:          support.NewHashStorageMap[sc.U64, types.Blake2bHash](keySystem, keyBlockHash, hashing.Twox64, types.DecodeBlake2bHash),
		BlockNumber:        support.NewHashStorageValue(keySystem, keyNumber, sc.DecodeU64),
		AllExtrinsicsLen:   support.NewHashStorageValue(keySystem, keyAllExtrinsicsLen, sc.DecodeU32),
		ExtrinsicIndex:     support.NewSimpleStorageValue(keyExtrinsicIndex, sc.DecodeU32),
		ExtrinsicData:      support.NewHashStorageMap[sc.U32, sc.Sequence[sc.U8]](keySystem, keyExtrinsicData, hashing.Twox64, sc.DecodeSequence[sc.U8]),
		ExtrinsicCount:     support.NewHashStorageValue(keySystem, keyExtrinsicCount, sc.DecodeU32),
		ParentHash:         support.NewHashStorageValue(keySystem, keyParentHash, types.DecodeBlake2bHash),
		Digest:             support.NewHashStorageValue(keySystem, keyDigest, types.DecodeDigest),
		Events:             support.NewHashStorageValue(keySystem, keyEvents, func(*bytes.Buffer) (types.EventRecord, error) { return types.EventRecord{}, nil }),
		EventCount:         support.NewHashStorageValue(keySystem, keyEventCount, sc.DecodeU32),
		EventTopics:        support.NewHashStorageMap[types.H256, sc.VaryingData](keySystem, keyEventTopics, hashing.Blake128, func(buffer *bytes.Buffer) (sc.VaryingData, error) { return sc.NewVaryingData(), nil }),
		LastRuntimeUpgrade: support.NewHashStorageValue(keySystem, keyLastRuntimeUpgrade, types.DecodeLastRuntimeUpgradeInfo),
		ExecutionPhase:     support.NewHashStorageValue(keySystem, keyExecutionPhase, types.DecodeExtrinsicPhase),
		HeapPages:          support.NewSimpleStorageValue(keyHeapPages, sc.DecodeU64),
		Code:               support.NewRawStorageValue(keyCode),
		AuthorizedUpgrade:  support.NewHashStorageValue(keySystem, keyAuthorizedUpgrade, DecodeCodeUpgradeAuthorization),
	}
}
