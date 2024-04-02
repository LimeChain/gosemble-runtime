package system

import (
	"errors"
	"math"
	"reflect"
	"testing"

	"github.com/ChainSafe/gossamer/lib/common"
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants"
	"github.com/LimeChain/gosemble/constants/metadata"
	"github.com/LimeChain/gosemble/mocks"
	"github.com/LimeChain/gosemble/primitives/log"
	"github.com/LimeChain/gosemble/primitives/types"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	moduleId = 0
)

var (
	blockNum   = sc.U64(1)
	eventCount = sc.U32(1)

	accountInfo = primitives.AccountInfo{
		Nonce:       1,
		Consumers:   2,
		Providers:   3,
		Sufficients: 4,
		Data: primitives.AccountData{
			Free:     sc.NewU128(5),
			Reserved: sc.NewU128(6),
			Frozen:   sc.NewU128(7),
		},
	}
	blockHashCount = sc.U64(5)
	blockWeights   = primitives.BlockWeights{
		BaseBlock: primitives.Weight{
			RefTime:   1,
			ProofSize: 2,
		},
		MaxBlock: primitives.Weight{
			RefTime:   3,
			ProofSize: 4,
		},
		PerClass: primitives.PerDispatchClassWeightsPerClass{
			Normal: primitives.WeightsPerClass{
				BaseExtrinsic: primitives.Weight{
					RefTime:   5,
					ProofSize: 6,
				},
			},
			Operational: primitives.WeightsPerClass{
				BaseExtrinsic: primitives.Weight{
					RefTime:   7,
					ProofSize: 8,
				},
			},
			Mandatory: primitives.WeightsPerClass{
				BaseExtrinsic: primitives.Weight{
					RefTime:   9,
					ProofSize: 10,
				},
			},
		},
	}
	blockLength = primitives.BlockLength{
		Max: primitives.PerDispatchClassU32{
			Normal:      11,
			Operational: 12,
			Mandatory:   13,
		},
	}
	dbWeight = primitives.RuntimeDbWeight{
		Read:  1,
		Write: 2,
	}
	baseWeight = primitives.WeightFromParts(567, 123)
	version    = primitives.RuntimeVersion{
		SpecName:           "test-spec",
		ImplName:           "test-impl",
		AuthoringVersion:   1,
		SpecVersion:        2,
		ImplVersion:        3,
		TransactionVersion: 4,
		StateVersion:       5,
	}
	parentHash = primitives.Blake2bHash{
		FixedSequence: sc.BytesToFixedSequenceU8(
			common.MustHexToHash("0x88dc3417d5058ec4b4503e0c12ea1a0a89be200fe98922423d4334014fa6b0ff").ToBytes(),
		)}
	blockNumber     = sc.U64(5)
	digest          = testDigest()
	targetAccountId = constants.ZeroAccountId
	moduleConstants = newConstants(primitives.BlockHashCount{U32: sc.U32(blockHashCount)}, blockWeights, blockLength, dbWeight, version)
)

var (
	unknownTransactionNoUnsignedValidator = primitives.NewTransactionValidityError(primitives.NewUnknownTransactionNoUnsignedValidator())
	mdGenerator                           = primitives.NewMetadataTypeGenerator()
	errPanic                              = errors.New("panic")
)

var (
	expectErrorsDefinition = primitives.NewMetadataTypeDefinitionVariant(
		sc.Sequence[primitives.MetadataDefinitionVariant]{
			primitives.NewMetadataDefinitionVariant(
				"InvalidSpecName",
				sc.Sequence[primitives.MetadataTypeDefinitionField]{},
				ErrorInvalidSpecName,
				"The name of specification does not match between the current runtime and the new runtime."),
			primitives.NewMetadataDefinitionVariant(
				"SpecVersionNeedsToIncrease",
				sc.Sequence[primitives.MetadataTypeDefinitionField]{},
				ErrorSpecVersionNeedsToIncrease,
				"The specification version is not allowed to decrease between the current runtime and the new runtime."),
			primitives.NewMetadataDefinitionVariant(
				"FailedToExtractRuntimeVersion",
				sc.Sequence[primitives.MetadataTypeDefinitionField]{},
				ErrorFailedToExtractRuntimeVersion,
				"Failed to extract the runtime version from the new runtime.  Either calling `Core_version` or decoding `RuntimeVersion` failed."),
			primitives.NewMetadataDefinitionVariant(
				"NonDefaultComposite",
				sc.Sequence[primitives.MetadataTypeDefinitionField]{},
				ErrorNonDefaultComposite,
				"Suicide called when the account has non-default composite data."),
			primitives.NewMetadataDefinitionVariant(
				"NonZeroRefCount",
				sc.Sequence[primitives.MetadataTypeDefinitionField]{},
				ErrorNonZeroRefCount,
				"There is a non-zero reference count preventing the account from being purged."),
			primitives.NewMetadataDefinitionVariant(
				"CallFiltered",
				sc.Sequence[primitives.MetadataTypeDefinitionField]{},
				ErrorCallFiltered,
				"The origin filter prevent the call to be dispatched.",
			),
			primitives.NewMetadataDefinitionVariant(
				"NothingAuthorized",
				sc.Sequence[primitives.MetadataTypeDefinitionField]{},
				ErrorNothingAuthorized,
				"No upgrade authorized.",
			),
			primitives.NewMetadataDefinitionVariant(
				"Unauthorized",
				sc.Sequence[primitives.MetadataTypeDefinitionField]{},
				ErrorUnauthorized,
				"The submitted code is not authorized.",
			),
		},
	)
)

var (
	mockStorageAccount            *mocks.StorageMap[primitives.AccountId, primitives.AccountInfo]
	mockStorageBlockWeight        *mocks.StorageValue[primitives.ConsumedWeight]
	mockStorageBlockHash          *mocks.StorageMap[sc.U64, primitives.Blake2bHash]
	mockStorageBlockNumber        *mocks.StorageValue[sc.U64]
	mockStorageAllExtrinsicsLen   *mocks.StorageValue[sc.U32]
	mockStorageExtrinsicIndex     *mocks.StorageValue[sc.U32]
	mockStorageExtrinsicData      *mocks.StorageMap[sc.U32, sc.Sequence[sc.U8]]
	mockStorageExtrinsicCount     *mocks.StorageValue[sc.U32]
	mockStorageParentHash         *mocks.StorageValue[primitives.Blake2bHash]
	mockStorageDigest             *mocks.StorageValue[primitives.Digest]
	mockStorageEvents             *mocks.StorageValue[primitives.EventRecord]
	mockStorageEventCount         *mocks.StorageValue[sc.U32]
	mockStorageEventTopics        *mocks.StorageMap[primitives.H256, sc.VaryingData]
	mockStorageLastRuntimeUpgrade *mocks.StorageValue[primitives.LastRuntimeUpgradeInfo]
	mockStorageExecutionPhase     *mocks.StorageValue[primitives.ExtrinsicPhase]
	mockStorageHeapPages          *mocks.StorageValue[sc.U64]
	mockStorageCode               *mocks.RawStorageValue
	mockStorageAuthorizedUpgrade  *mocks.StorageValue[CodeUpgradeAuthorization]

	mockIoStorage *mocks.IoStorage
	mockIoHashing *mocks.IoHashing
	mockIoTrie    *mocks.IoTrie
	mockIoMisc    *mocks.IoMisc

	mockTypeMutateAccountInfo       = mock.AnythingOfType("func(*types.AccountInfo) (goscale.Encodable, error)")
	mockTypeMutateOptionAccountInfo = mock.AnythingOfType("func(*goscale.Option[github.com/LimeChain/gosemble/primitives/types.AccountInfo]) (goscale.Encodable, error)")
)

func Test_Module_GetIndex(t *testing.T) {
	assert.Equal(t, sc.U8(moduleId), setupModule().GetIndex())
}

func Test_Module_Functions(t *testing.T) {
	target := setupModule()
	functions := target.Functions()

	assert.Equal(t, 11, len(functions))
}

func Test_Module_PreDispatch(t *testing.T) {
	target := setupModule()
	mockCall := new(mocks.Call)

	result, err := target.PreDispatch(mockCall)

	assert.Nil(t, err)
	assert.Equal(t, sc.Empty{}, result)
}

func Test_Module_ValidateUnsigned(t *testing.T) {
	target := setupModule()
	mockCall := new(mocks.Call)

	result, err := target.ValidateUnsigned(primitives.TransactionSource{}, mockCall)

	assert.Equal(t, unknownTransactionNoUnsignedValidator, err)
	assert.Equal(t, primitives.ValidTransaction{}, result)
}

func Test_Module_BlockHashCount(t *testing.T) {
	assert.Equal(t, blockHashCount, sc.U64(setupModule().BlockHashCount().U32))
}

func Test_Module_BlockLength(t *testing.T) {
	assert.Equal(t, blockLength, setupModule().BlockLength())
}

func Test_Module_BlockWeights(t *testing.T) {
	assert.Equal(t, blockWeights, setupModule().BlockWeights())
}

func Test_Module_DbWeight(t *testing.T) {
	assert.Equal(t, dbWeight, setupModule().DbWeight())
}

func Test_Module_Version(t *testing.T) {
	assert.Equal(t, version, setupModule().Version())
}

func Test_Module_StorageDigest(t *testing.T) {
	target := setupModule()

	mockStorageDigest.On("Get").Return(digest, nil)

	result, err := target.StorageDigest()
	assert.Nil(t, err)

	assert.Equal(t, digest, result)
	mockStorageDigest.AssertCalled(t, "Get")
}

func Test_Module_StorageBlockWeight(t *testing.T) {
	blockWeight := primitives.ConsumedWeight{
		Normal:      primitives.WeightFromParts(1, 2),
		Operational: primitives.WeightFromParts(3, 4),
		Mandatory:   primitives.WeightFromParts(5, 6),
	}
	target := setupModule()

	mockStorageBlockWeight.On("Get").Return(blockWeight, nil)

	result, err := target.StorageBlockWeight()
	assert.Nil(t, err)

	assert.Equal(t, blockWeight, result)
	mockStorageBlockWeight.AssertCalled(t, "Get")
}

func Test_Module_StorageBlockWeightSet(t *testing.T) {
	blockWeight := primitives.ConsumedWeight{
		Normal:      primitives.WeightFromParts(1, 2),
		Operational: primitives.WeightFromParts(3, 4),
		Mandatory:   primitives.WeightFromParts(5, 6),
	}
	target := setupModule()

	mockStorageBlockWeight.On("Put", blockWeight).Return()

	target.StorageBlockWeightSet(blockWeight)

	mockStorageBlockWeight.AssertCalled(t, "Put", blockWeight)
}

func Test_Module_StorageBlockHash(t *testing.T) {
	key := sc.U64(0)
	target := setupModule()

	mockStorageBlockHash.On("Get", key).Return(parentHash, nil)

	result, err := target.StorageBlockHash(key)
	assert.Nil(t, err)

	assert.Equal(t, parentHash, result)
	mockStorageBlockHash.AssertCalled(t, "Get", key)
}

func Test_Module_StorageBlockHashSet(t *testing.T) {
	key := sc.U64(0)
	target := setupModule()

	mockStorageBlockHash.On("Put", key, parentHash).Return()

	target.StorageBlockHashSet(key, parentHash)

	mockStorageBlockHash.AssertCalled(t, "Put", key, parentHash)
}

func Test_Module_StorageBlockHashExists(t *testing.T) {
	key := sc.U64(0)
	target := setupModule()

	mockStorageBlockHash.On("Exists", key).Return(true)

	result := target.StorageBlockHashExists(key)

	assert.Equal(t, true, result)
	mockStorageBlockHash.AssertCalled(t, "Exists", key)
}

func Test_Module_StorageBlockNumber(t *testing.T) {
	target := setupModule()

	mockStorageBlockNumber.On("Get").Return(blockNumber, nil)

	result, err := target.StorageBlockNumber()
	assert.Nil(t, err)

	assert.Equal(t, blockNumber, result)
	mockStorageBlockNumber.AssertCalled(t, "Get")
}

func Test_Module_StorageBlockNumberSet(t *testing.T) {
	target := setupModule()

	mockStorageBlockNumber.On("Put", blockNumber).Return()

	target.StorageBlockNumberSet(blockNumber)

	mockStorageBlockNumber.AssertCalled(t, "Put", blockNumber)
}

func Test_Module_StorageLastRuntimeUpgrade(t *testing.T) {
	lrui := primitives.LastRuntimeUpgradeInfo{
		SpecVersion: sc.Compact{Number: sc.U32(1)},
		SpecName:    "test",
	}
	target := setupModule()

	mockStorageLastRuntimeUpgrade.On("Get").Return(lrui, nil)

	result, err := target.StorageLastRuntimeUpgrade()
	assert.Nil(t, err)

	assert.Equal(t, lrui, result)
	mockStorageLastRuntimeUpgrade.AssertCalled(t, "Get")
}

func Test_Module_StorageLastRuntimeUpgradeSet(t *testing.T) {
	lrui := primitives.LastRuntimeUpgradeInfo{
		SpecVersion: sc.Compact{Number: sc.U32(1)},
		SpecName:    "test",
	}
	target := setupModule()

	mockStorageLastRuntimeUpgrade.On("Put", lrui).Return()

	target.StorageLastRuntimeUpgradeSet(lrui)

	mockStorageLastRuntimeUpgrade.AssertCalled(t, "Put", lrui)
}

func Test_Module_StorageAccount(t *testing.T) {
	target := setupModule()

	mockStorageAccount.On("Get", targetAccountId).Return(accountInfo, nil)

	result, err := target.StorageAccount(targetAccountId)
	assert.Nil(t, err)

	assert.Equal(t, accountInfo, result)
	mockStorageAccount.AssertCalled(t, "Get", targetAccountId)
}

func Test_Module_StorageAccountSet(t *testing.T) {
	target := setupModule()

	mockStorageAccount.On("Put", targetAccountId, accountInfo).Return()

	target.StorageAccountSet(targetAccountId, accountInfo)

	mockStorageAccount.AssertCalled(t, "Put", targetAccountId, accountInfo)
}

func Test_Module_StorageAllExtrinsicLen(t *testing.T) {
	extrinsicLen := sc.U32(2)
	target := setupModule()

	mockStorageAllExtrinsicsLen.On("Get").Return(extrinsicLen, nil)

	result, err := target.StorageAllExtrinsicsLen()
	assert.Nil(t, err)

	assert.Equal(t, extrinsicLen, result)
	mockStorageAllExtrinsicsLen.AssertCalled(t, "Get")
}

func Test_Module_StorageAllExtrinsicLenSet(t *testing.T) {
	extrinsicLen := sc.U32(2)
	target := setupModule()

	mockStorageAllExtrinsicsLen.On("Put", extrinsicLen).Return()

	target.StorageAllExtrinsicsLenSet(extrinsicLen)

	mockStorageAllExtrinsicsLen.AssertCalled(t, "Put", extrinsicLen)
}

func Test_Module_Initialize(t *testing.T) {
	target := setupModule()
	executionPhase := primitives.NewExtrinsicPhaseInitialization()

	mockStorageExecutionPhase.On("Put", executionPhase).Return()
	mockStorageExtrinsicIndex.On("Put", sc.U32(0)).Return()
	mockStorageBlockNumber.On("Put", blockNumber).Return()
	mockStorageDigest.On("Put", digest).Return()
	mockStorageParentHash.On("Put", parentHash).Return()
	mockStorageBlockHash.On("Put", blockNumber-1, parentHash).Return()
	mockStorageBlockWeight.On("Clear").Return()

	target.Initialize(blockNumber, parentHash, digest)

	mockStorageExecutionPhase.AssertCalled(t, "Put", executionPhase)
	mockStorageExtrinsicIndex.AssertCalled(t, "Put", sc.U32(0))
	mockStorageBlockNumber.AssertCalled(t, "Put", blockNumber)
	mockStorageDigest.AssertCalled(t, "Put", digest)
	mockStorageParentHash.AssertCalled(t, "Put", parentHash)
	mockStorageBlockHash.AssertCalled(t, "Put", blockNumber-1, parentHash)
	mockStorageBlockWeight.AssertCalled(t, "Clear")
}

func Test_Module_RegisterExtraWeightUnchecked(t *testing.T) {
	blockWeight := primitives.ConsumedWeight{
		Normal:      primitives.WeightFromParts(1, 2),
		Operational: primitives.WeightFromParts(3, 4),
		Mandatory:   primitives.WeightFromParts(5, 6),
	}
	weight := primitives.WeightFromParts(7, 8)
	class := primitives.NewDispatchClassNormal()
	target := setupModule()
	expectCurrentWeight := primitives.ConsumedWeight{
		Normal:      blockWeight.Normal.Add(weight),
		Operational: blockWeight.Operational,
		Mandatory:   blockWeight.Mandatory,
	}

	mockStorageBlockWeight.On("Get").Return(blockWeight, nil)
	mockStorageBlockWeight.On("Put", expectCurrentWeight)

	target.RegisterExtraWeightUnchecked(weight, class)

	mockStorageBlockWeight.AssertCalled(t, "Get")
	mockStorageBlockWeight.AssertCalled(t, "Put", expectCurrentWeight)
}

func Test_Module_RegisterExtraWeightUnchecked_BlockWeight_Error(t *testing.T) {
	target := setupModule()

	weight := primitives.WeightFromParts(7, 8)
	class := primitives.NewDispatchClassNormal()

	mockStorageBlockWeight.On("Get").Return(primitives.ConsumedWeight{}, errPanic)

	err := target.RegisterExtraWeightUnchecked(weight, class)
	assert.Equal(t, errPanic, err)

	mockStorageBlockWeight.AssertCalled(t, "Get")
}

func Test_Module_RegisterExtraWeightUnchecked_Accrue_Error(t *testing.T) {
	target := setupModule()

	weight := primitives.WeightFromParts(7, 8)
	class := primitives.DispatchClass{VaryingData: sc.NewVaryingData(sc.U8(99))}

	mockStorageBlockWeight.On("Get").Return(primitives.ConsumedWeight{}, nil)

	err := target.RegisterExtraWeightUnchecked(weight, class)
	assert.Equal(t, "not a valid 'DispatchClass' type", err.Error())

	mockStorageBlockWeight.AssertCalled(t, "Get")
}

func Test_Module_NoteFinishedInitialize(t *testing.T) {
	executionPhase := primitives.NewExtrinsicPhaseApply(sc.U32(0))
	target := setupModule()

	mockStorageExecutionPhase.On("Put", executionPhase).Return()

	target.NoteFinishedInitialize()

	mockStorageExecutionPhase.AssertCalled(t, "Put", executionPhase)
}

func Test_Module_NoteExtrinsic(t *testing.T) {
	extrinsicBytes := []byte("test")
	extrinsicIndex := sc.U32(1)
	target := setupModule()

	mockStorageExtrinsicIndex.On("Get").Return(extrinsicIndex, nil)
	mockStorageExtrinsicData.On("Put", extrinsicIndex, sc.BytesToSequenceU8(extrinsicBytes)).Return()

	target.NoteExtrinsic(extrinsicBytes)

	mockStorageExtrinsicIndex.AssertCalled(t, "Get")
	mockStorageExtrinsicData.AssertCalled(t, "Put", extrinsicIndex, sc.BytesToSequenceU8(extrinsicBytes))
}

func Test_Module_NoteAppliedExtrinsic_ExtrinsicSuccess(t *testing.T) {
	blockNum := sc.U64(5)
	eventCount := sc.U32(0)
	extrinsicIndex := sc.U32(1)
	postInfo := primitives.PostDispatchInfo{
		ActualWeight: sc.NewOption[primitives.Weight](nil),
		PaysFee:      primitives.PaysYes,
	}
	dispatchInfo := primitives.DispatchInfo{
		Class:   primitives.NewDispatchClassNormal(),
		PaysFee: primitives.PaysYes,
	}
	expectDispatchInfo := primitives.DispatchInfo{
		Weight:  blockWeights.PerClass.Normal.BaseExtrinsic,
		Class:   primitives.NewDispatchClassNormal(),
		PaysFee: primitives.PaysYes,
	}
	expectEventRecord := primitives.EventRecord{
		Phase:  primitives.NewExtrinsicPhaseInitialization(),
		Event:  newEventExtrinsicSuccess(moduleId, expectDispatchInfo),
		Topics: []primitives.H256{},
	}

	target := setupModule()

	mockStorageBlockNumber.On("Get").Return(blockNum, nil)
	mockStorageExecutionPhase.On("Get").Return(primitives.NewExtrinsicPhaseInitialization(), nil)
	mockStorageEventCount.On("Get").Return(eventCount, nil)
	mockStorageEventCount.On("Put", eventCount+1).Return()
	mockStorageEvents.On("Append", expectEventRecord).Return()

	mockStorageExtrinsicIndex.On("Get").Return(extrinsicIndex, nil)
	mockStorageExtrinsicIndex.On("Put", extrinsicIndex+1).Return()
	mockStorageExecutionPhase.On("Put", primitives.NewExtrinsicPhaseApply(extrinsicIndex+1)).Return()

	target.NoteAppliedExtrinsic(postInfo, nil, dispatchInfo)

	mockStorageBlockNumber.AssertNumberOfCalls(t, "Get", 1)
	mockStorageExecutionPhase.AssertCalled(t, "Get")
	mockStorageEventCount.AssertCalled(t, "Get")
	mockStorageEventCount.AssertCalled(t, "Put", eventCount+1)
	mockStorageEvents.AssertCalled(t, "Append", expectEventRecord)
	mockStorageEventTopics.AssertNotCalled(t, "Append")

	mockStorageExtrinsicIndex.AssertCalled(t, "Get")
	mockStorageExecutionPhase.AssertCalled(t, "Put", primitives.NewExtrinsicPhaseApply(extrinsicIndex+1))
}

func Test_Module_NoteAppliedExtrinsic_ExtrinsicFailed(t *testing.T) {
	blockNum := sc.U64(5)
	eventCount := sc.U32(0)
	extrinsicIndex := sc.U32(1)

	dispatchErr := primitives.NewDispatchErrorCorruption()
	dispatchInfo := primitives.DispatchInfo{
		Class:   primitives.NewDispatchClassNormal(),
		PaysFee: primitives.PaysYes,
	}
	expectDispatchInfo := primitives.DispatchInfo{
		Weight:  blockWeights.PerClass.Normal.BaseExtrinsic,
		Class:   primitives.NewDispatchClassNormal(),
		PaysFee: primitives.PaysYes,
	}
	expectEventRecord := primitives.EventRecord{
		Phase:  primitives.NewExtrinsicPhaseInitialization(),
		Event:  newEventExtrinsicFailed(moduleId, dispatchErr, expectDispatchInfo),
		Topics: []primitives.H256{},
	}

	target := setupModule()

	mockStorageBlockNumber.On("Get").Return(blockNum, nil)
	mockStorageExecutionPhase.On("Get").Return(primitives.NewExtrinsicPhaseInitialization(), nil)
	mockStorageEventCount.On("Get").Return(eventCount, nil)
	mockStorageEventCount.On("Put", eventCount+1).Return()
	mockStorageEvents.On("Append", expectEventRecord).Return()

	mockStorageExtrinsicIndex.On("Get").Return(extrinsicIndex, nil)
	mockStorageExtrinsicIndex.On("Put", extrinsicIndex+1).Return()
	mockStorageExecutionPhase.On("Put", primitives.NewExtrinsicPhaseApply(extrinsicIndex+1)).Return()

	target.NoteAppliedExtrinsic(primitives.PostDispatchInfo{}, dispatchErr, dispatchInfo)

	mockStorageBlockNumber.AssertNumberOfCalls(t, "Get", 2)
	mockStorageExecutionPhase.AssertCalled(t, "Get")
	mockStorageEventCount.AssertCalled(t, "Get")
	mockStorageEventCount.AssertCalled(t, "Put", eventCount+1)
	mockStorageEvents.AssertCalled(t, "Append", expectEventRecord)
	mockStorageEventTopics.AssertNotCalled(t, "Append")

	mockStorageExtrinsicIndex.AssertCalled(t, "Get")
	mockStorageExecutionPhase.AssertCalled(t, "Put", primitives.NewExtrinsicPhaseApply(extrinsicIndex+1))
}

func Test_Module_Finalize_RemovePreviousHash(t *testing.T) {
	target := setupModule()
	blockNumber := sc.U64(7)

	extrinsicCount := sc.U32(1)
	extrinsicDataBytes := []byte("extrinsicDataBytes")
	extrinsicRoot := common.MustHexToHash("0x3aa96b0149b6ca3688878bdbd19464448624136398e3ce45b9e755d3ab61355a").ToBytes()
	expectExtrinsicRoot := primitives.H256{FixedSequence: sc.BytesToFixedSequenceU8(extrinsicRoot)}
	storageRoot := common.MustHexToHash("0x3aa96b0149b6ca3688878bdbd19464448624136398e3ce45b9e755d3ab61355b").ToBytes()
	expectStorageRoot := primitives.H256{FixedSequence: sc.BytesToFixedSequenceU8(storageRoot)}
	blakeArgs := append(sc.ToCompact(uint64(extrinsicCount)).Bytes(), extrinsicDataBytes...)

	expectResult := primitives.Header{
		ParentHash:     parentHash,
		Number:         blockNumber,
		StateRoot:      expectStorageRoot,
		ExtrinsicsRoot: expectExtrinsicRoot,
		Digest:         digest,
	}

	mockStorageExecutionPhase.On("Clear").Return()
	mockStorageAllExtrinsicsLen.On("Clear").Return()

	mockStorageBlockNumber.On("Get").Return(blockNumber, nil)
	mockStorageParentHash.On("Get").Return(parentHash, nil)
	mockStorageDigest.On("Get").Return(digest, nil)
	mockStorageExtrinsicCount.On("Take").Return(extrinsicCount, nil)
	mockStorageExtrinsicData.On("TakeBytes", sc.U32(0)).Return(extrinsicDataBytes, nil)
	mockIoTrie.On("Blake2256OrderedRoot", blakeArgs, int32(constants.StorageVersion)).Return(extrinsicRoot)
	mockStorageBlockHash.On("Remove", sc.U64(1)).Return()
	mockIoStorage.On("Root", int32(version.StateVersion)).Return(storageRoot)

	result, err := target.Finalize()
	assert.Nil(t, err)

	assert.Equal(t, expectResult, result)

	mockStorageExecutionPhase.AssertCalled(t, "Clear")
	mockStorageAllExtrinsicsLen.AssertCalled(t, "Clear")

	mockStorageBlockNumber.AssertCalled(t, "Get")
	mockStorageParentHash.AssertCalled(t, "Get")
	mockStorageDigest.AssertCalled(t, "Get")
	mockStorageExtrinsicCount.AssertCalled(t, "Take")
	mockStorageExtrinsicData.AssertCalled(t, "TakeBytes", sc.U32(0))
	mockIoTrie.AssertCalled(t, "Blake2256OrderedRoot", blakeArgs, int32(constants.StorageVersion))
	mockStorageBlockHash.AssertCalled(t, "Remove", sc.U64(1))
	mockIoStorage.AssertCalled(t, "Root", int32(version.StateVersion))
}

func Test_Module_Finalize_Success(t *testing.T) {
	target := setupModule()
	extrinsicCount := sc.U32(1)
	extrinsicDataBytes := []byte("extrinsicDataBytes")
	extrinsicRoot := common.MustHexToHash("0x3aa96b0149b6ca3688878bdbd19464448624136398e3ce45b9e755d3ab61355a").ToBytes()
	expectExtrinsicRoot := primitives.H256{FixedSequence: sc.BytesToFixedSequenceU8(extrinsicRoot)}
	storageRoot := common.MustHexToHash("0x3aa96b0149b6ca3688878bdbd19464448624136398e3ce45b9e755d3ab61355b").ToBytes()
	expectStorageRoot := primitives.H256{FixedSequence: sc.BytesToFixedSequenceU8(storageRoot)}
	blakeArgs := append(sc.ToCompact(uint64(extrinsicCount)).Bytes(), extrinsicDataBytes...)

	expectResult := primitives.Header{
		ParentHash:     parentHash,
		Number:         blockNumber,
		StateRoot:      expectStorageRoot,
		ExtrinsicsRoot: expectExtrinsicRoot,
		Digest:         digest,
	}

	mockStorageExecutionPhase.On("Clear").Return()
	mockStorageAllExtrinsicsLen.On("Clear").Return()

	mockStorageBlockNumber.On("Get").Return(blockNumber, nil)
	mockStorageParentHash.On("Get").Return(parentHash, nil)
	mockStorageDigest.On("Get").Return(digest, nil)
	mockStorageExtrinsicCount.On("Take").Return(extrinsicCount, nil)
	mockStorageExtrinsicData.On("TakeBytes", sc.U32(0)).Return(extrinsicDataBytes, nil)
	mockIoTrie.On("Blake2256OrderedRoot", blakeArgs, int32(constants.StorageVersion)).Return(extrinsicRoot)
	mockIoStorage.On("Root", int32(version.StateVersion)).Return(storageRoot)

	result, err := target.Finalize()
	assert.Nil(t, err)

	assert.Equal(t, expectResult, result)

	mockStorageExecutionPhase.AssertCalled(t, "Clear")
	mockStorageAllExtrinsicsLen.AssertCalled(t, "Clear")

	mockStorageBlockNumber.AssertCalled(t, "Get")
	mockStorageParentHash.AssertCalled(t, "Get")
	mockStorageDigest.AssertCalled(t, "Get")
	mockStorageExtrinsicCount.AssertCalled(t, "Take")
	mockStorageExtrinsicData.AssertCalled(t, "TakeBytes", sc.U32(0))
	mockIoTrie.AssertCalled(t, "Blake2256OrderedRoot", blakeArgs, int32(constants.StorageVersion))
	mockStorageBlockHash.AssertNotCalled(t, "Remove", mock.Anything)
	mockIoStorage.AssertCalled(t, "Root", int32(version.StateVersion))
}

func Test_Module_NoteFinishedExtrinsics(t *testing.T) {
	extrinsicIndex := sc.U32(4)
	target := setupModule()

	mockStorageExtrinsicIndex.On("Take").Return(extrinsicIndex, nil)
	mockStorageExtrinsicCount.On("Put", extrinsicIndex).Return()
	mockStorageExecutionPhase.On("Put", primitives.NewExtrinsicPhaseFinalization()).Return()

	target.NoteFinishedExtrinsics()

	mockStorageExtrinsicIndex.AssertCalled(t, "Take")
	mockStorageExtrinsicCount.AssertCalled(t, "Put", extrinsicIndex)
	mockStorageExecutionPhase.AssertCalled(t, "Put", primitives.NewExtrinsicPhaseFinalization())
}

func Test_Module_ResetEvents(t *testing.T) {
	target := setupModule()

	mockStorageEvents.On("Clear").Return()
	mockStorageEventCount.On("Clear").Return()
	mockStorageEventTopics.On("Clear", sc.U32(math.MaxUint32))

	target.ResetEvents()

	mockStorageEvents.AssertCalled(t, "Clear")
	mockStorageEventCount.AssertCalled(t, "Clear")
	mockStorageEventTopics.AssertCalled(t, "Clear", sc.U32(math.MaxUint32))
}

func Test_Module_CanDecProviders_ZeroConsumer(t *testing.T) {
	target := setupModule()
	accountInfo := primitives.AccountInfo{}

	mockStorageAccount.On("Get", targetAccountId).Return(accountInfo, nil)

	result, err := target.CanDecProviders(targetAccountId)
	assert.Nil(t, err)
	assert.Equal(t, true, result)

	mockStorageAccount.AssertCalled(t, "Get", targetAccountId)
}

func Test_Module_CanDecProviders_Providers(t *testing.T) {
	target := setupModule()
	accountInfo := primitives.AccountInfo{
		Consumers: 2,
		Providers: 3,
	}

	mockStorageAccount.On("Get", targetAccountId).Return(accountInfo, nil)

	result, err := target.CanDecProviders(targetAccountId)
	assert.Nil(t, err)
	assert.Equal(t, true, result)

	mockStorageAccount.AssertCalled(t, "Get", targetAccountId)
}

func Test_Module_CanDecProviders_False(t *testing.T) {
	target := setupModule()
	accountInfo := primitives.AccountInfo{
		Consumers: 2,
	}

	mockStorageAccount.On("Get", targetAccountId).Return(accountInfo, nil)

	result, err := target.CanDecProviders(targetAccountId)
	assert.Nil(t, err)
	assert.Equal(t, false, result)

	mockStorageAccount.AssertCalled(t, "Get", targetAccountId)
}

func Test_Module_TryMutateExists_Error(t *testing.T) {
	target := setupModule()
	expectedErr := primitives.NewDispatchErrorBadOrigin()

	accountInfo := primitives.AccountInfo{}
	f := func(account *primitives.AccountData) (sc.Encodable, error) {
		return nil, expectedErr
	}

	mockStorageAccount.On("Get", targetAccountId).Return(accountInfo, nil)

	_, err := target.TryMutateExists(targetAccountId, f)

	assert.Equal(t, expectedErr, err)

	mockStorageAccount.AssertCalled(t, "Get", targetAccountId)
	mockStorageAccount.AssertNotCalled(t,
		"Mutate",
		targetAccountId,
		mockTypeMutateAccountInfo)
}

func Test_Module_TryMutateExists_NoProviding(t *testing.T) {
	target := setupModule()
	expectedResult := sc.NewU128(5)

	accountInfo := primitives.AccountInfo{}
	f := func(account *primitives.AccountData) (sc.Encodable, error) {
		return expectedResult, nil
	}

	mockStorageAccount.On("Get", targetAccountId).Return(accountInfo, nil)

	result, err := target.TryMutateExists(targetAccountId, f)
	assert.Nil(t, err)

	assert.Equal(t, expectedResult, result)

	mockStorageAccount.AssertCalled(t, "Get", targetAccountId)
	mockStorageAccount.AssertNotCalled(t,
		"Mutate",
		targetAccountId,
		mockTypeMutateAccountInfo)
}

func Test_Module_TryMutateExists_WasProviding_NoLongerProviding_DecRefStatus_Success(t *testing.T) {
	target := setupModule()
	expectedResult := sc.NewU128(5)

	accountInfo := primitives.AccountInfo{
		Data: primitives.AccountData{
			Free: sc.NewU128(1),
		},
	}
	f := func(account *primitives.AccountData) (sc.Encodable, error) {
		account.Free = primitives.Balance{}
		return expectedResult, nil
	}

	mockStorageAccount.On("Get", targetAccountId).Return(accountInfo, nil)
	mockStorageAccount.
		On(
			"TryMutateExists",
			targetAccountId,
			mockTypeMutateOptionAccountInfo).
		Return(primitives.DecRefStatusExists, nil)

	result, err := target.TryMutateExists(targetAccountId, f)
	assert.Nil(t, err)

	assert.Equal(t, expectedResult, result)

	mockStorageAccount.AssertCalled(t, "Get", targetAccountId)
	mockStorageAccount.
		AssertCalled(t,
			"TryMutateExists",
			targetAccountId,
			mockTypeMutateOptionAccountInfo)
	mockStorageAccount.AssertNotCalled(t,
		"Mutate",
		targetAccountId,
		mockTypeMutateAccountInfo)
}

func Test_Module_TryMutateExists_WasProviding_NoLongerProviding_Error(t *testing.T) {
	target := setupModule()
	expectedErr := primitives.NewDispatchErrorCannotLookup()

	accountInfo := primitives.AccountInfo{
		Data: primitives.AccountData{
			Free: sc.NewU128(1),
		},
	}
	f := func(account *primitives.AccountData) (sc.Encodable, error) {
		account.Free = primitives.Balance{}
		return sc.Empty{}, nil
	}

	mockStorageAccount.On("Get", targetAccountId).Return(accountInfo, nil)
	mockStorageAccount.
		On(
			"TryMutateExists",
			targetAccountId,
			mockTypeMutateOptionAccountInfo).
		Return(sc.Empty{}, expectedErr)

	_, err := target.TryMutateExists(targetAccountId, f)

	assert.Equal(t, expectedErr, err)

	mockStorageAccount.AssertCalled(t, "Get", targetAccountId)
	mockStorageAccount.
		AssertCalled(t,
			"TryMutateExists",
			targetAccountId,
			mockTypeMutateOptionAccountInfo)
	mockStorageAccount.AssertNotCalled(t,
		"Mutate",
		targetAccountId,
		mockTypeMutateAccountInfo)
}

func Test_Module_TryMutateExists_WasNotProviding_IsProviding(t *testing.T) {
	target := setupModule()

	expectedResult := sc.NewU128(5)
	accountInfo := primitives.AccountInfo{
		Data: primitives.AccountData{},
	}
	f := func(account *primitives.AccountData) (sc.Encodable, error) {
		account.Free = sc.NewU128(5)
		return expectedResult, nil
	}

	mockStorageAccount.On("Get", targetAccountId).Return(accountInfo, nil)
	mockStorageAccount.On(
		"Mutate",
		targetAccountId,
		mockTypeMutateAccountInfo).
		Return(primitives.IncRefStatusExisted, nil).Once()
	mockStorageAccount.On(
		"Mutate",
		targetAccountId,
		mockTypeMutateAccountInfo).
		Return(sc.NewU128(2), nil).Once()

	result, err := target.TryMutateExists(targetAccountId, f)
	assert.Nil(t, err)

	assert.Equal(t, expectedResult, result)

	mockStorageAccount.AssertCalled(t, "Get", targetAccountId)
	mockStorageAccount.AssertNumberOfCalls(t, "Mutate", 2)
	mockStorageAccount.AssertCalled(t,
		"Mutate",
		targetAccountId,
		mockTypeMutateAccountInfo)
}

func Test_Module_TryMutateExists_WasProviding_IsProviding_Success(t *testing.T) {
	target := setupModule()

	expectedResult := sc.NewU128(5)
	accountInfo := primitives.AccountInfo{
		Data: primitives.AccountData{
			Free: sc.NewU128(1),
		},
	}
	f := func(*primitives.AccountData) (sc.Encodable, error) {
		return expectedResult, nil
	}

	mockStorageAccount.On("Get", targetAccountId).Return(accountInfo, nil)
	mockStorageAccount.On(
		"Mutate",
		targetAccountId,
		mockTypeMutateAccountInfo).
		Return(sc.Empty{}, nil)

	result, err := target.TryMutateExists(targetAccountId, f)
	assert.Nil(t, err)

	assert.Equal(t, expectedResult, result)

	mockStorageAccount.AssertCalled(t, "Get", targetAccountId)
	mockStorageAccount.AssertNumberOfCalls(t, "Mutate", 1)
	mockStorageAccount.AssertCalled(t,
		"Mutate",
		targetAccountId,
		mockTypeMutateAccountInfo)
}

func Test_Module_TryMutateExists_GetAccount_Error(t *testing.T) {
	target := setupModule()
	expectedErr := primitives.NewDispatchErrorBadOrigin()

	accountInfo := primitives.AccountInfo{}
	f := func(account *primitives.AccountData) (sc.Encodable, error) {
		return nil, nil
	}

	mockStorageAccount.On("Get", targetAccountId).Return(accountInfo, expectedErr)

	_, err := target.TryMutateExists(targetAccountId, f)

	assert.Equal(t, expectedErr, err)
	mockStorageAccount.AssertCalled(t, "Get", targetAccountId)
	mockStorageAccount.AssertNotCalled(t,
		"Mutate",
		targetAccountId,
		mockTypeMutateAccountInfo)
}

func Test_Module_TryMutateExists_incProviders_Error(t *testing.T) {
	target := setupModule()

	expectedErr := errors.New("err")
	expectedResult := sc.NewU128(5)
	accountInfo := primitives.AccountInfo{
		Data: primitives.AccountData{},
	}
	f := func(account *primitives.AccountData) (sc.Encodable, error) {
		account.Free = sc.NewU128(5)
		return expectedResult, nil
	}

	mockStorageAccount.On("Get", targetAccountId).Return(accountInfo, nil)
	mockStorageAccount.On(
		"Mutate",
		targetAccountId,
		mockTypeMutateAccountInfo).
		Return(primitives.IncRefStatusExisted, expectedErr)

	_, err := target.TryMutateExists(targetAccountId, f)

	assert.Equal(t, expectedErr, err)

	mockStorageAccount.AssertCalled(t, "Get", targetAccountId)
	mockStorageAccount.AssertCalled(t, "Mutate", targetAccountId, mockTypeMutateAccountInfo)
}

func Test_Module_TryMutateExists_AccountMutate_Error(t *testing.T) {
	target := setupModule()

	expectedErr := errors.New("err")
	expectedResult := sc.NewU128(5)
	accountInfo := primitives.AccountInfo{
		Data: primitives.AccountData{},
	}
	f := func(account *primitives.AccountData) (sc.Encodable, error) {
		account.Free = sc.NewU128(5)
		return expectedResult, nil
	}

	mockStorageAccount.On("Get", targetAccountId).Return(accountInfo, nil)
	mockStorageAccount.On(
		"Mutate",
		targetAccountId,
		mockTypeMutateAccountInfo).
		Return(primitives.IncRefStatusExisted, nil).Once()
	mockStorageAccount.On(
		"Mutate",
		targetAccountId,
		mockTypeMutateAccountInfo).
		Return(sc.NewU128(2), expectedErr).Once()

	_, err := target.TryMutateExists(targetAccountId, f)

	assert.Equal(t, expectedErr, err)

	mockStorageAccount.AssertCalled(t, "Get", targetAccountId)
	mockStorageAccount.AssertNumberOfCalls(t, "Mutate", 2)
	mockStorageAccount.AssertCalled(t,
		"Mutate",
		targetAccountId,
		mockTypeMutateAccountInfo)
}

func Test_Module_incrementProviders_RefStatusCreated(t *testing.T) {
	accountInfo := &primitives.AccountInfo{}
	expectedResult := primitives.IncRefStatusCreated
	target := setupModule()

	mockStorageBlockNumber.On("Get").Return(sc.U64(0), nil)

	result := target.incrementProviders(targetAccountId, accountInfo)

	assert.Equal(t, expectedResult, result)
	assert.Equal(t, sc.U32(1), accountInfo.Providers)

	mockStorageBlockNumber.AssertCalled(t, "Get")
	mockStorageExecutionPhase.AssertNotCalled(t, "Get")
	mockStorageEventCount.AssertNotCalled(t, "Get")
	mockStorageEventCount.AssertNotCalled(t, "Put", mock.Anything)
	mockStorageEventCount.AssertNotCalled(t, "Append", mock.Anything)
	mockStorageEventTopics.AssertNotCalled(t, "Append", mock.Anything, mock.Anything)
}

func Test_Module_incrementProviders_RefStatusExisted(t *testing.T) {
	accountInfo := &primitives.AccountInfo{
		Sufficients: 1,
	}
	expectedResult := primitives.IncRefStatusExisted
	target := setupModule()

	result := target.incrementProviders(targetAccountId, accountInfo)

	assert.Equal(t, expectedResult, result)
	assert.Equal(t, sc.U32(1), accountInfo.Providers)

	mockStorageBlockNumber.AssertNotCalled(t, "Get")
	mockStorageExecutionPhase.AssertNotCalled(t, "Get")
	mockStorageEventCount.AssertNotCalled(t, "Get")
	mockStorageEventCount.AssertNotCalled(t, "Put", mock.Anything)
	mockStorageEventCount.AssertNotCalled(t, "Append", mock.Anything)
	mockStorageEventTopics.AssertNotCalled(t, "Append", mock.Anything, mock.Anything)
}

func Test_Module_DepositEvent_Success(t *testing.T) {
	firstHash := [32]sc.U8{}
	firstHash[0] = 1
	secondHash := [32]sc.U8{}
	secondHash[0] = 2
	event := newEventCodeUpdated(moduleId)
	expectEventRecord := primitives.EventRecord{
		Phase:  primitives.NewExtrinsicPhaseInitialization(),
		Event:  event,
		Topics: []primitives.H256{},
	}
	blockNum := sc.U64(1)
	eventCount := sc.U32(2)
	target := setupModule()

	mockStorageBlockNumber.On("Get").Return(blockNum, nil)
	mockStorageExecutionPhase.On("Get").Return(primitives.NewExtrinsicPhaseInitialization(), nil)
	mockStorageEventCount.On("Get").Return(eventCount, nil)
	mockStorageEventCount.On("Put", eventCount+1).Return()
	mockStorageEvents.On("Append", expectEventRecord).Return()

	target.DepositEvent(event)

	mockStorageBlockNumber.AssertCalled(t, "Get")
	mockStorageExecutionPhase.AssertCalled(t, "Get")
	mockStorageEventCount.AssertCalled(t, "Get")
	mockStorageEventCount.AssertCalled(t, "Put", eventCount+1)
	mockStorageEvents.AssertCalled(t, "Append", expectEventRecord)
}

func Test_Module_depositEventIndexed_Success(t *testing.T) {
	firstHash := [32]sc.U8{}
	firstHash[0] = 1
	secondHash := [32]sc.U8{}
	secondHash[0] = 2
	topics := []primitives.H256{
		{
			firstHash[:],
		},
		{
			secondHash[:],
		},
	}
	event := newEventCodeUpdated(moduleId)
	expectEventRecord := primitives.EventRecord{
		Phase:  primitives.NewExtrinsicPhaseInitialization(),
		Event:  event,
		Topics: topics,
	}
	blockNum := sc.U64(1)
	eventCount := sc.U32(2)
	topicValue := sc.NewVaryingData(blockNum, eventCount)
	target := setupModule()

	mockStorageBlockNumber.On("Get").Return(blockNum, nil)
	mockStorageExecutionPhase.On("Get").Return(primitives.NewExtrinsicPhaseInitialization(), nil)
	mockStorageEventCount.On("Get").Return(eventCount, nil)
	mockStorageEventCount.On("Put", eventCount+1).Return()
	mockStorageEvents.On("Append", expectEventRecord).Return()
	mockStorageEventTopics.On("Append", topics[0], topicValue).Once()
	mockStorageEventTopics.On("Append", topics[1], topicValue).Once()

	target.depositEventIndexed(topics, event)

	mockStorageBlockNumber.AssertCalled(t, "Get")
	mockStorageExecutionPhase.AssertCalled(t, "Get")
	mockStorageEventCount.AssertCalled(t, "Get")
	mockStorageEventCount.AssertCalled(t, "Put", eventCount+1)
	mockStorageEvents.AssertCalled(t, "Append", expectEventRecord)
	mockStorageEventTopics.AssertNumberOfCalls(t, "Append", 2)
	mockStorageEventTopics.AssertCalled(t, "Append", topics[0], topicValue)
	mockStorageEventTopics.AssertCalled(t, "Append", topics[0], topicValue)
}

func Test_Module_DepositEvent_Overflow(t *testing.T) {
	target := setupModule()
	mockStorageBlockNumber.On("Get").Return(sc.U64(1), nil)
	mockStorageExecutionPhase.On("Get").Return(primitives.NewExtrinsicPhaseInitialization(), nil)
	mockStorageEventCount.On("Get").Return(sc.U32(math.MaxUint32), nil)

	target.DepositEvent(newEventCodeUpdated(moduleId))

	mockStorageBlockNumber.AssertCalled(t, "Get")
	mockStorageExecutionPhase.AssertCalled(t, "Get")
	mockStorageEventCount.AssertCalled(t, "Get")
	mockStorageEventCount.AssertNotCalled(t, "Put", mock.Anything)
	mockStorageEventCount.AssertNotCalled(t, "Append", mock.Anything)
	mockStorageEventTopics.AssertNotCalled(t, "Append", mock.Anything, mock.Anything)
}

func Test_Module_DepositEvent_ZeroBlockNumber(t *testing.T) {
	target := setupModule()
	mockStorageBlockNumber.On("Get").Return(sc.U64(0), nil)

	target.DepositEvent(newEventCodeUpdated(moduleId))

	mockStorageBlockNumber.AssertCalled(t, "Get")
	mockStorageExecutionPhase.AssertNotCalled(t, "Get")
	mockStorageEventCount.AssertNotCalled(t, "Get")
	mockStorageEventCount.AssertNotCalled(t, "Put", mock.Anything)
	mockStorageEventCount.AssertNotCalled(t, "Append", mock.Anything)
	mockStorageEventTopics.AssertNotCalled(t, "Append", mock.Anything, mock.Anything)
}

func Test_Module_decrementProviders_HasAccount_NoProvidersLeft(t *testing.T) {
	target := setupModule()
	maybeAccount := sc.NewOption[primitives.AccountInfo](primitives.AccountInfo{})
	expectedResult := primitives.DecRefStatusReaped

	mockStorageBlockNumber.On("Get").Return(sc.U64(0), nil)

	result, err := target.decrementProviders(targetAccountId, &maybeAccount)

	assert.NoError(t, err)
	assert.Equal(t, expectedResult, result)
	assert.Equal(t, sc.U32(1), maybeAccount.Value.Providers)

	mockStorageBlockNumber.AssertCalled(t, "Get")
	mockStorageExecutionPhase.AssertNotCalled(t, "Get")
	mockStorageEventCount.AssertNotCalled(t, "Get")
}

func Test_Module_decrementProviders_HasAccount_ConsumerRemaining(t *testing.T) {
	target := setupModule()
	accountInfo := primitives.AccountInfo{
		Consumers: 1,
		Data:      primitives.AccountData{},
	}
	maybeAccount := sc.NewOption[primitives.AccountInfo](accountInfo)
	expectedErr := primitives.NewDispatchErrorConsumerRemaining()

	_, err := target.decrementProviders(targetAccountId, &maybeAccount)

	assert.Equal(t, expectedErr, err)
	assert.Equal(t, sc.U32(1), maybeAccount.Value.Providers)
	assert.Equal(t, sc.U32(1), maybeAccount.Value.Consumers)
}

func Test_Module_decrementProviders_HasAccount_ContinueExist(t *testing.T) {
	target := setupModule()
	accountInfo := primitives.AccountInfo{
		Sufficients: 1,
		Data:        primitives.AccountData{},
	}
	maybeAccount := sc.NewOption[primitives.AccountInfo](accountInfo)
	expectedResult := primitives.DecRefStatusExists

	result, err := target.decrementProviders(targetAccountId, &maybeAccount)

	assert.NoError(t, err)
	assert.Equal(t, expectedResult, result)
	assert.Equal(t, sc.U32(0), maybeAccount.Value.Providers)
	assert.Equal(t, sc.U32(1), maybeAccount.Value.Sufficients)
}

func Test_Module_decrementProviders_NoAccount(t *testing.T) {
	target := setupModule()
	maybeAccount := sc.NewOption[primitives.AccountInfo](nil)
	expectedResult := primitives.DecRefStatusReaped

	result, err := target.decrementProviders(targetAccountId, &maybeAccount)

	assert.NoError(t, err)
	assert.Equal(t, expectedResult, result)
}

func Test_Module_mutateAccount(t *testing.T) {
	accountInfo := &primitives.AccountInfo{
		Nonce:       1,
		Consumers:   2,
		Providers:   3,
		Sufficients: 4,
		Data: primitives.AccountData{
			Free:     sc.NewU128(1),
			Reserved: sc.NewU128(2),
			Frozen:   sc.NewU128(4),
		},
	}
	accountData := primitives.AccountData{
		Free:     sc.NewU128(5),
		Reserved: sc.NewU128(6),
		Frozen:   sc.NewU128(8),
	}
	expectAccountInfo := &primitives.AccountInfo{
		Nonce:       1,
		Consumers:   2,
		Providers:   3,
		Sufficients: 4,
		Data:        accountData,
	}

	mutateAccount(accountInfo, &accountData)

	assert.Equal(t, expectAccountInfo, accountInfo)
}

func Test_Module_ErrorsDefinition(t *testing.T) {
	target := setupModule()

	assert.Equal(t, &expectErrorsDefinition, target.errorsDefinition())
}

func Test_Module_mutateAccount_NilData(t *testing.T) {
	accountInfo := &primitives.AccountInfo{
		Nonce:       1,
		Consumers:   2,
		Providers:   3,
		Sufficients: 4,
		Data: primitives.AccountData{
			Free:     sc.NewU128(1),
			Reserved: sc.NewU128(2),
			Frozen:   sc.NewU128(4),
		},
	}
	expectAccountInfo := &primitives.AccountInfo{
		Nonce:       1,
		Consumers:   2,
		Providers:   3,
		Sufficients: 4,
		Data:        primitives.DefaultAccountData(),
	}

	mutateAccount(accountInfo, nil)

	assert.Equal(t, expectAccountInfo, accountInfo)
}

func Test_Module_Metadata(t *testing.T) {
	target := setupModule()

	expectedSystemCallId := mdGenerator.GetLastAvailableIndex() + 1
	expectedSystemErrorsId := expectedSystemCallId + 1
	expectedTypesPhaseId := expectedSystemErrorsId + 1
	expectedTypesBlockId := expectedTypesPhaseId + 1
	expectedTypesWeightPerClassId := expectedTypesBlockId + 1
	expectedTypesOptionWeightId := expectedTypesWeightPerClassId + 1
	expectedPerDispatchClassWeightId := expectedTypesOptionWeightId + 1
	expectedPerDispatchClassWeightPerClassId := expectedPerDispatchClassWeightId + 1
	expectedTypesBlockWeightsId := expectedPerDispatchClassWeightPerClassId + 1
	expectedTypesDbWeightId := expectedTypesBlockWeightsId + 1
	expectedTypesValidTransactionId := expectedTypesDbWeightId + 1
	expectedLastRuntimeUpgradeInfoId := expectedTypesValidTransactionId + 1
	expectedCompactU32Id := expectedLastRuntimeUpgradeInfoId + 1
	expectedTransactionSourceId := expectedCompactU32Id + 1
	expectedTypeInvalidTransactionId := expectedTransactionSourceId + 1
	expectedTypeUnknownTransactionId := expectedTypeInvalidTransactionId + 1
	expectedTransactionValidityErrorTypeId := expectedTypeUnknownTransactionId + 1
	expectedTransactionValidityResultId := expectedTransactionValidityErrorTypeId + 1
	expectedPerDispatchClassU32Id := expectedTransactionValidityResultId + 1
	expectedTypesBlockLengthId := expectedPerDispatchClassU32Id + 1

	expectMetadataTypes := sc.Sequence[primitives.MetadataType]{
		primitives.NewMetadataTypeWithParam(expectedSystemCallId,
			"System calls",
			sc.Sequence[sc.Str]{"pallet_system", "pallet", "Call"},
			primitives.NewMetadataTypeDefinitionVariant(

				sc.Sequence[primitives.MetadataDefinitionVariant]{
					primitives.NewMetadataDefinitionVariant(
						"remark",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionField(metadata.TypesSequenceU8),
						},
						functionRemarkIndex,
						"Make some on-chain remark.",
					),

					primitives.NewMetadataDefinitionVariant(
						"set_heap_pages",
						sc.Sequence[types.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionField(metadata.PrimitiveTypesU64),
						},
						functionSetHeapPagesIndex,
						"Set the number of pages in the WebAssembly environment's heap.",
					),

					primitives.NewMetadataDefinitionVariant(
						"set_code",
						sc.Sequence[types.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionField(metadata.TypesSequenceU8),
						},
						functionSetCodeIndex,
						"Set the new runtime code.",
					),

					primitives.NewMetadataDefinitionVariant(
						"set_code_without_checks",
						sc.Sequence[types.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionField(metadata.TypesSequenceU8),
						},
						functionSetCodeWithoutChecksIndex,
						"Set the new runtime code without any checks.",
					),

					primitives.NewMetadataDefinitionVariant(
						"set_storage",
						sc.Sequence[types.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionField(metadata.TypesSequenceKeyValue),
						},
						functionSetStorageIndex,
						"Set some items of storage.",
					),

					primitives.NewMetadataDefinitionVariant(
						"kill_storage",
						sc.Sequence[types.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionField(metadata.TypesSequenceSequenceU8),
						},
						functionKillStorageIndex,
						"Kill some items from storage.",
					),

					primitives.NewMetadataDefinitionVariant(
						"kill_prefix",
						sc.Sequence[types.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionField(metadata.TypesSequenceU8),
							primitives.NewMetadataTypeDefinitionField(metadata.PrimitiveTypesU32),
						},
						functionKillPrefixIndex,
						"Kill all storage items with a key that starts with the given prefix.",
					),

					primitives.NewMetadataDefinitionVariant(
						"remark_with_event",
						sc.Sequence[types.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionField(metadata.TypesSequenceU8),
						},
						functionRemarkWithEventIndex,
						"Make some on-chain remark and emit an event.",
					),

					primitives.NewMetadataDefinitionVariant(
						"authorize_upgrade",
						sc.Sequence[types.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionField(metadata.TypesH256),
						},
						functionAuthorizeUpgradeIndex,
						"Authorize new runtime code.",
					),

					primitives.NewMetadataDefinitionVariant(
						"authorize_upgrade_without_checks",
						sc.Sequence[types.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionField(metadata.TypesH256),
						},
						functionAuthorizeUpgradeWithoutChecksIndex,
						"Authorize new runtime code and an upgrade sans verification.",
					),

					primitives.NewMetadataDefinitionVariant(
						"apply_authorized_upgrade",
						sc.Sequence[types.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionField(metadata.TypesSequenceU8),
						},
						functionApplyAuthorizedUpgradeIndex,
						"Provide new, already-authorized runtime code.",
					),
				}),
			primitives.NewMetadataEmptyTypeParameter("T")),
		primitives.NewMetadataTypeWithPath(
			expectedSystemErrorsId,
			"frame_system pallet Error",
			sc.Sequence[sc.Str]{"frame_system", "pallet", "Error"},
			expectErrorsDefinition,
		),
		primitives.NewMetadataTypeWithPath(expectedTypesPhaseId,
			"ExtrinsicPhase",
			sc.Sequence[sc.Str]{"frame_system", "Phase"},
			primitives.NewMetadataTypeDefinitionVariant(
				sc.Sequence[primitives.MetadataDefinitionVariant]{
					primitives.NewMetadataDefinitionVariant(
						"ApplyExtrinsic",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionField(metadata.PrimitiveTypesU32),
						},
						primitives.PhaseApplyExtrinsic,
						"Phase.ApplyExtrinsic"),
					primitives.NewMetadataDefinitionVariant(
						"Finalization",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{},
						primitives.PhaseFinalization,
						"Phase.Finalization"),
					primitives.NewMetadataDefinitionVariant(
						"Initialization",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{},
						primitives.PhaseInitialization,
						"Phase.Initialization"),
				})),
		primitives.NewMetadataTypeWithParams(expectedTypesBlockId, "block",
			sc.Sequence[sc.Str]{"sp_runtime", "generic", "block", "Block"},
			primitives.NewMetadataTypeDefinitionComposite(
				sc.Sequence[primitives.MetadataTypeDefinitionField]{
					primitives.NewMetadataTypeDefinitionFieldWithName(metadata.Header, "header"),
					primitives.NewMetadataTypeDefinitionFieldWithName(metadata.TypesSequenceUncheckedExtrinsics, "Vec<extrinsics>"),
				}),
			sc.Sequence[primitives.MetadataTypeParameter]{
				primitives.NewMetadataTypeParameter(metadata.Header, "Header"),
				primitives.NewMetadataTypeParameter(metadata.UncheckedExtrinsic, "Extrinsic"),
			},
		),
		primitives.NewMetadataTypeWithParam(expectedTypesOptionWeightId, "Option<Weight>", sc.Sequence[sc.Str]{"Option"}, primitives.NewMetadataTypeDefinitionVariant(
			sc.Sequence[primitives.MetadataDefinitionVariant]{
				primitives.NewMetadataDefinitionVariant(
					"None",
					sc.Sequence[primitives.MetadataTypeDefinitionField]{},
					0,
					"Option<Weight>(nil)"),
				primitives.NewMetadataDefinitionVariant(
					"Some",
					sc.Sequence[primitives.MetadataTypeDefinitionField]{
						primitives.NewMetadataTypeDefinitionField(metadata.TypesWeight),
					},
					1,
					"Option<Weight>(value)"),
			}),
			primitives.NewMetadataTypeParameter(metadata.TypesWeight, "T"),
		),
		primitives.NewMetadataTypeWithPath(expectedTypesWeightPerClassId, "WeightsPerClass", sc.Sequence[sc.Str]{"frame_system", "limits", "WeightsPerClass"}, primitives.NewMetadataTypeDefinitionComposite(
			sc.Sequence[primitives.MetadataTypeDefinitionField]{
				primitives.NewMetadataTypeDefinitionFieldWithName(metadata.TypesWeight, "BaseExtrinsic"),
				primitives.NewMetadataTypeDefinitionFieldWithName(expectedTypesOptionWeightId, "MaxExtrinsic"),
				primitives.NewMetadataTypeDefinitionFieldWithName(expectedTypesOptionWeightId, "MaxTotal"),
				primitives.NewMetadataTypeDefinitionFieldWithName(expectedTypesOptionWeightId, "Reserved"),
			})),
		primitives.NewMetadataTypeWithPath(expectedPerDispatchClassWeightId, "PerDispatchClassWeight", sc.Sequence[sc.Str]{"frame_support", "dispatch", "PerDispatchClass"}, primitives.NewMetadataTypeDefinitionComposite(
			sc.Sequence[primitives.MetadataTypeDefinitionField]{
				primitives.NewMetadataTypeDefinitionFieldWithName(metadata.TypesWeight, "Normal"),
				primitives.NewMetadataTypeDefinitionFieldWithName(metadata.TypesWeight, "Operational"),
				primitives.NewMetadataTypeDefinitionFieldWithName(metadata.TypesWeight, "Mandatory"),
			})),
		primitives.NewMetadataTypeWithPath(expectedPerDispatchClassWeightPerClassId, "PerDispatchClassWeightsPerClass", sc.Sequence[sc.Str]{"frame_support", "dispatch", "PerDispatchClass"}, primitives.NewMetadataTypeDefinitionComposite(
			sc.Sequence[primitives.MetadataTypeDefinitionField]{
				primitives.NewMetadataTypeDefinitionFieldWithName(expectedTypesWeightPerClassId, "Normal"),
				primitives.NewMetadataTypeDefinitionFieldWithName(expectedTypesWeightPerClassId, "Operational"),
				primitives.NewMetadataTypeDefinitionFieldWithName(expectedTypesWeightPerClassId, "Mandatory"),
			})),
		primitives.NewMetadataTypeWithPath(expectedTypesBlockWeightsId,
			"BlockWeights",
			sc.Sequence[sc.Str]{"frame_system", "limits", "BlockWeights"}, primitives.NewMetadataTypeDefinitionComposite(
				sc.Sequence[primitives.MetadataTypeDefinitionField]{
					primitives.NewMetadataTypeDefinitionFieldWithName(metadata.TypesWeight, "BaseBlock"),
					primitives.NewMetadataTypeDefinitionFieldWithName(metadata.TypesWeight, "MaxBlock"),
					primitives.NewMetadataTypeDefinitionFieldWithName(expectedPerDispatchClassWeightPerClassId, "PerClass"),
				})),
		primitives.NewMetadataTypeWithPath(expectedTypesDbWeightId, "RuntimeDbWeight", sc.Sequence[sc.Str]{"sp_weights", "RuntimeDbWeight"}, primitives.NewMetadataTypeDefinitionComposite(
			sc.Sequence[primitives.MetadataTypeDefinitionField]{
				primitives.NewMetadataTypeDefinitionFieldWithName(metadata.PrimitiveTypesU64, "Read"),
				primitives.NewMetadataTypeDefinitionFieldWithName(metadata.PrimitiveTypesU64, "Write"),
			})),

		primitives.NewMetadataTypeWithPath(expectedTypesValidTransactionId, "ValidTransaction", sc.Sequence[sc.Str]{"sp_runtime", "transaction_validity", "ValidTransaction"},
			primitives.NewMetadataTypeDefinitionComposite(
				sc.Sequence[primitives.MetadataTypeDefinitionField]{
					primitives.NewMetadataTypeDefinitionFieldWithName(metadata.PrimitiveTypesU64, "Priority"),
					primitives.NewMetadataTypeDefinitionFieldWithName(metadata.TypesSequenceSequenceU8, "Vec<Requires>"),
					primitives.NewMetadataTypeDefinitionFieldWithName(metadata.TypesSequenceSequenceU8, "Vec<Provides>"),
					primitives.NewMetadataTypeDefinitionFieldWithName(metadata.PrimitiveTypesU64, "Longevity"),
					primitives.NewMetadataTypeDefinitionFieldWithName(metadata.PrimitiveTypesBool, "Propagate"),
				},
			)),
		primitives.NewMetadataType(expectedCompactU32Id, "CompactU32", primitives.NewMetadataTypeDefinitionCompact(sc.ToCompact(metadata.PrimitiveTypesU32))),
		primitives.NewMetadataTypeWithPath(expectedLastRuntimeUpgradeInfoId,
			"LastRuntimeUpgradeInfo",
			sc.Sequence[sc.Str]{"frame_system", "LastRuntimeUpgradeInfo"}, primitives.NewMetadataTypeDefinitionComposite(
				sc.Sequence[primitives.MetadataTypeDefinitionField]{
					primitives.NewMetadataTypeDefinitionFieldWithName(expectedCompactU32Id, "SpecVersion"),
					primitives.NewMetadataTypeDefinitionFieldWithName(metadata.PrimitiveTypesString, "SpecName"),
				})),
		primitives.NewMetadataTypeWithPath(expectedTransactionSourceId, "TransactionSource", sc.Sequence[sc.Str]{"sp_runtime", "transaction_validity", "TransactionSource"},
			primitives.NewMetadataTypeDefinitionVariant(
				sc.Sequence[primitives.MetadataDefinitionVariant]{
					primitives.NewMetadataDefinitionVariant(
						"InBlock",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{},
						primitives.TransactionSourceInBlock,
						"TransactionSourceInBlock"),
					primitives.NewMetadataDefinitionVariant(
						"Local",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{},
						primitives.TransactionSourceLocal,
						"TransactionSourceLocal"),
					primitives.NewMetadataDefinitionVariant(
						"External",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{},
						primitives.TransactionSourceExternal,
						"TransactionSourceExternal"),
				})),
		// type 871
		primitives.NewMetadataTypeWithPath(expectedTypeInvalidTransactionId, "InvalidTransaction", sc.Sequence[sc.Str]{"sp_runtime", "transaction_validity", "InvalidTransaction"},
			primitives.NewMetadataTypeDefinitionVariant(
				sc.Sequence[primitives.MetadataDefinitionVariant]{
					primitives.NewMetadataDefinitionVariant(
						"Call",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{},
						primitives.InvalidTransactionCall,
						""),
					primitives.NewMetadataDefinitionVariant(
						"Payment",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{},
						primitives.InvalidTransactionPayment,
						""),
					primitives.NewMetadataDefinitionVariant(
						"Future",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{},
						primitives.InvalidTransactionFuture,
						""),
					primitives.NewMetadataDefinitionVariant(
						"Stale",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{},
						primitives.InvalidTransactionStale,
						""),
					primitives.NewMetadataDefinitionVariant(
						"BadProof",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{},
						primitives.InvalidTransactionBadProof,
						""),
					primitives.NewMetadataDefinitionVariant(
						"AncientBirthBlock",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{},
						primitives.InvalidTransactionAncientBirthBlock,
						""),
					primitives.NewMetadataDefinitionVariant(
						"ExhaustsResources",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{},
						primitives.InvalidTransactionExhaustsResources,
						""),
					primitives.NewMetadataDefinitionVariant(
						"Custom",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionField(metadata.PrimitiveTypesU8),
						},
						primitives.InvalidTransactionCustom,
						""),
					primitives.NewMetadataDefinitionVariant(
						"BadMandatory",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{},
						primitives.InvalidTransactionBadMandatory,
						""),
					primitives.NewMetadataDefinitionVariant(
						"MandatoryValidation",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{},
						primitives.InvalidTransactionMandatoryValidation,
						""),
					primitives.NewMetadataDefinitionVariant(
						"BadSigner",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{},
						primitives.InvalidTransactionBadSigner,
						""),
				},
			)),
		// type 872
		primitives.NewMetadataTypeWithPath(expectedTypeUnknownTransactionId, "UnknownTransaction", sc.Sequence[sc.Str]{"sp_runtime", "transaction_validity", "UnknownTransaction"},
			primitives.NewMetadataTypeDefinitionVariant(
				sc.Sequence[primitives.MetadataDefinitionVariant]{
					primitives.NewMetadataDefinitionVariant(
						"CannotLookup",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{},
						primitives.UnknownTransactionCannotLookup,
						""),
					primitives.NewMetadataDefinitionVariant(
						"NoUnsignedValidator",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{},
						primitives.UnknownTransactionNoUnsignedValidator,
						""),
					primitives.NewMetadataDefinitionVariant(
						"Custom",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionField(metadata.PrimitiveTypesU8),
						},
						primitives.UnknownTransactionCustomUnknownTransaction,
						""),
				},
			)),
		// type 870
		primitives.NewMetadataTypeWithPath(expectedTransactionValidityErrorTypeId, "TransactionValidityError", sc.Sequence[sc.Str]{"sp_runtime", "transaction_validity", "TransactionValidityError"},
			primitives.NewMetadataTypeDefinitionVariant(
				sc.Sequence[primitives.MetadataDefinitionVariant]{
					primitives.NewMetadataDefinitionVariant(
						"Invalid",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionField(expectedTypeInvalidTransactionId),
						},
						primitives.TransactionValidityErrorInvalidTransaction,
						""),
					primitives.NewMetadataDefinitionVariant(
						"Unknown",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionField(expectedTypeUnknownTransactionId),
						},
						primitives.TransactionValidityErrorUnknownTransaction,
						""),
				},
			)),
		primitives.NewMetadataTypeWithPath(expectedTransactionValidityResultId, "TransactionValidityResult", sc.Sequence[sc.Str]{"Result"},
			primitives.NewMetadataTypeDefinitionVariant(
				sc.Sequence[primitives.MetadataDefinitionVariant]{
					primitives.NewMetadataDefinitionVariant(
						"Ok",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionField(expectedTypesValidTransactionId),
						},
						primitives.TransactionValidityResultValid,
						""),
					primitives.NewMetadataDefinitionVariant(
						"Err",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionField(expectedTransactionValidityErrorTypeId),
						},
						primitives.TransactionValidityResultError,
						""),
				})),
		primitives.NewMetadataTypeWithPath(expectedPerDispatchClassU32Id,
			"PerDispatchClassU32",
			sc.Sequence[sc.Str]{"frame_support", "dispatch", "PerDispatchClass"},
			primitives.NewMetadataTypeDefinitionComposite(
				sc.Sequence[primitives.MetadataTypeDefinitionField]{
					primitives.NewMetadataTypeDefinitionFieldWithName(metadata.PrimitiveTypesU32, "Normal"),
					primitives.NewMetadataTypeDefinitionFieldWithName(metadata.PrimitiveTypesU32, "Operational"),
					primitives.NewMetadataTypeDefinitionFieldWithName(metadata.PrimitiveTypesU32, "Mandatory"),
				},
			),
		),
		primitives.NewMetadataTypeWithPath(expectedTypesBlockLengthId,
			"BlockLength",
			sc.Sequence[sc.Str]{"frame_system", "limits", "BlockLength"},
			primitives.NewMetadataTypeDefinitionComposite(sc.Sequence[primitives.MetadataTypeDefinitionField]{
				primitives.NewMetadataTypeDefinitionFieldWithName(expectedPerDispatchClassU32Id, "Max"),
			})),

		primitives.NewMetadataType(metadata.TypesSystemEventStorage,
			"Vec<Box<EventRecord<T::RuntimeEvent, T::Hash>>>",
			primitives.NewMetadataTypeDefinitionSequence(sc.ToCompact(metadata.TypesEventRecord))),

		primitives.NewMetadataType(metadata.TypesVecBlockNumEventIndex, "Vec<BlockNumber, EventIndex>",
			primitives.NewMetadataTypeDefinitionSequence(sc.ToCompact(metadata.TypesTupleU32U32))),

		primitives.NewMetadataTypeWithParams(metadata.TypesEventRecord,
			"frame_system EventRecord",
			sc.Sequence[sc.Str]{"frame_system", "EventRecord"},
			primitives.NewMetadataTypeDefinitionComposite(sc.Sequence[primitives.MetadataTypeDefinitionField]{
				primitives.NewMetadataTypeDefinitionFieldWithNames(expectedTypesPhaseId, "phase", "Phase"),
				primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesRuntimeEvent, "event", "E"),
				primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesVecTopics, "topics", "Vec<T>"),
			}),
			sc.Sequence[primitives.MetadataTypeParameter]{
				primitives.NewMetadataTypeParameter(metadata.TypesRuntimeEvent, "E"),
				primitives.NewMetadataTypeParameter(metadata.TypesH256, "T"),
			}),
		primitives.NewMetadataTypeWithPath(metadata.TypesSystemEvent,
			"frame_system pallet Event",
			sc.Sequence[sc.Str]{"frame_system", "pallet", "Event"}, primitives.NewMetadataTypeDefinitionVariant(
				sc.Sequence[primitives.MetadataDefinitionVariant]{
					primitives.NewMetadataDefinitionVariant(
						"ExtrinsicSuccess",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesDispatchInfo, "dispatch_info", "DispatchInfo"),
						},
						EventExtrinsicSuccess,
						"Event.ExtrinsicSuccess"),
					primitives.NewMetadataDefinitionVariant(
						"ExtrinsicFailed",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesDispatchError, "dispatch_error", "DispatchError"),
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesDispatchInfo, "dispatch_info", "DispatchInfo"),
						},
						EventExtrinsicFailed,
						"Events.ExtrinsicFailed"),
					primitives.NewMetadataDefinitionVariant(
						"CodeUpdated",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{},
						EventCodeUpdated,
						"Events.CodeUpdated"),
					primitives.NewMetadataDefinitionVariant(
						"NewAccount",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesAddress32, "account", "T::AccountId"),
						},
						EventNewAccount,
						"Events.NewAccount"),
					primitives.NewMetadataDefinitionVariant(
						"KilledAccount",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesAddress32, "account", "T::AccountId"),
						},
						EventKilledAccount,
						"Events.KilledAccount"),
					primitives.NewMetadataDefinitionVariant(
						"Remarked",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesAddress32, "sender", "T::AccountId"),
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesH256, "hash", "T::Hash"),
						},
						EventRemarked,
						"Events.Remarked"),
					primitives.NewMetadataDefinitionVariant(
						"UpgradeAuthorized",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesH256, "code_hash", "T::Hash"),
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.PrimitiveTypesBool, "check_version", "bool"),
						},
						EventUpgradeAuthorized,
						"Events.UpgradeAuthorized"),
				})),

		primitives.NewMetadataTypeWithPath(metadata.TypesEra, "Era", sc.Sequence[sc.Str]{"sp_runtime", "generic", "era", "Era"}, primitives.NewMetadataTypeDefinitionVariant(primitives.EraTypeDefinition())),
	}

	moduleV14 := primitives.MetadataModuleV14{
		Name: name,
		Storage: sc.NewOption[primitives.MetadataModuleStorage](primitives.MetadataModuleStorage{
			Prefix: name,
			Items: sc.Sequence[primitives.MetadataModuleStorageEntry]{
				primitives.NewMetadataModuleStorageEntry(
					"Account",
					primitives.MetadataModuleStorageEntryModifierDefault,
					primitives.NewMetadataModuleStorageEntryDefinitionMap(
						sc.Sequence[primitives.MetadataModuleStorageHashFunc]{primitives.MetadataModuleStorageHashFuncMultiBlake128Concat},
						sc.ToCompact(metadata.TypesAddress32),
						sc.ToCompact(metadata.TypesAccountInfo)),
					"The full account information for a particular account ID."),
				primitives.NewMetadataModuleStorageEntry(
					"ExtrinsicCount",
					primitives.MetadataModuleStorageEntryModifierOptional,
					primitives.NewMetadataModuleStorageEntryDefinitionPlain(
						sc.ToCompact(metadata.PrimitiveTypesU32)),
					"Total extrinsics count for the current block."),
				primitives.NewMetadataModuleStorageEntry(
					"BlockWeight",
					primitives.MetadataModuleStorageEntryModifierDefault,
					primitives.NewMetadataModuleStorageEntryDefinitionPlain(
						sc.ToCompact(expectedPerDispatchClassWeightId)),
					"The current weight for the block."),
				primitives.NewMetadataModuleStorageEntry(
					"AllExtrinsicsLen",
					primitives.MetadataModuleStorageEntryModifierOptional,
					primitives.NewMetadataModuleStorageEntryDefinitionPlain(
						sc.ToCompact(metadata.PrimitiveTypesU32)),
					"Total length (in bytes) for all extrinsics put together, for the current block."),
				primitives.NewMetadataModuleStorageEntry(
					"BlockHash",
					primitives.MetadataModuleStorageEntryModifierDefault,
					primitives.NewMetadataModuleStorageEntryDefinitionMap(
						sc.Sequence[primitives.MetadataModuleStorageHashFunc]{primitives.MetadataModuleStorageHashFuncMultiXX64},
						sc.ToCompact(metadata.PrimitiveTypesU32),
						sc.ToCompact(metadata.TypesFixedSequence32U8)),
					"Map of block numbers to block hashes."),
				primitives.NewMetadataModuleStorageEntry(
					"ExtrinsicData",
					primitives.MetadataModuleStorageEntryModifierDefault,
					primitives.NewMetadataModuleStorageEntryDefinitionMap(
						sc.Sequence[primitives.MetadataModuleStorageHashFunc]{primitives.MetadataModuleStorageHashFuncMultiXX64},
						sc.ToCompact(metadata.PrimitiveTypesU32),
						sc.ToCompact(metadata.TypesSequenceU8)),
					"Extrinsics data for the current block (maps an extrinsic's index to its data)."),
				primitives.NewMetadataModuleStorageEntry(
					"Number",
					primitives.MetadataModuleStorageEntryModifierDefault,
					primitives.NewMetadataModuleStorageEntryDefinitionPlain(
						sc.ToCompact(metadata.PrimitiveTypesU32)),
					"The current block number being processed. Set by `execute_block`."),
				primitives.NewMetadataModuleStorageEntry(
					"ParentHash",
					primitives.MetadataModuleStorageEntryModifierDefault,
					primitives.NewMetadataModuleStorageEntryDefinitionPlain(
						sc.ToCompact(metadata.TypesFixedSequence32U8)),
					"Hash of the previous block."),
				primitives.NewMetadataModuleStorageEntry(
					"Digest",
					primitives.MetadataModuleStorageEntryModifierDefault,
					primitives.NewMetadataModuleStorageEntryDefinitionPlain(
						sc.ToCompact(metadata.TypesDigest)),
					"Digest of the current block, also part of the block header."),
				primitives.NewMetadataModuleStorageEntry(
					"Events",
					primitives.MetadataModuleStorageEntryModifierDefault,
					primitives.NewMetadataModuleStorageEntryDefinitionPlain(sc.ToCompact(metadata.TypesSystemEventStorage)),
					"Events deposited for the current block.   NOTE: The item is unbound and should therefore never be read on chain."),
				primitives.NewMetadataModuleStorageEntry(
					"EventTopics",
					primitives.MetadataModuleStorageEntryModifierDefault,
					primitives.NewMetadataModuleStorageEntryDefinitionMap(
						sc.Sequence[primitives.MetadataModuleStorageHashFunc]{primitives.MetadataModuleStorageHashFuncMultiBlake128Concat},
						sc.ToCompact(metadata.TypesH256),
						sc.ToCompact(metadata.TypesVecBlockNumEventIndex)), "Mapping between a topic (represented by T::Hash) and a vector of indexes  of events in the `<Events<T>>` list."),
				primitives.NewMetadataModuleStorageEntry(
					"EventCount",
					primitives.MetadataModuleStorageEntryModifierDefault,
					primitives.NewMetadataModuleStorageEntryDefinitionPlain(
						sc.ToCompact(metadata.PrimitiveTypesU32)),
					"The number of events in the `Events<T>` list."),
				primitives.NewMetadataModuleStorageEntry(
					"LastRuntimeUpgrade",
					primitives.MetadataModuleStorageEntryModifierOptional,
					primitives.NewMetadataModuleStorageEntryDefinitionPlain(sc.ToCompact(expectedLastRuntimeUpgradeInfoId)),
					"Stores the `spec_version` and `spec_name` of when the last runtime upgrade happened."),
				primitives.NewMetadataModuleStorageEntry(
					"ExecutionPhase",
					primitives.MetadataModuleStorageEntryModifierOptional,
					primitives.NewMetadataModuleStorageEntryDefinitionPlain(sc.ToCompact(expectedTypesPhaseId)),
					"The execution phase of the block.",
				),
				primitives.NewMetadataModuleStorageEntry(
					"AuthorizedUpgrade",
					primitives.MetadataModuleStorageEntryModifierOptional,
					primitives.NewMetadataModuleStorageEntryDefinitionPlain(sc.ToCompact(metadata.TypesCodeUpgradeAuthorization)),
					"Optional code upgrade authorization for the runtime.",
				),
			},
		}),
		Call: sc.NewOption[sc.Compact](sc.ToCompact(expectedSystemCallId)),
		CallDef: sc.NewOption[primitives.MetadataDefinitionVariant](
			primitives.NewMetadataDefinitionVariantStr(
				name,
				sc.Sequence[primitives.MetadataTypeDefinitionField]{
					primitives.NewMetadataTypeDefinitionFieldWithName(expectedSystemCallId, "self::sp_api_hidden_includes_construct_runtime::hidden_include::dispatch\n::CallableCallFor<System, Runtime>"),
				},
				moduleId,
				"Call.System"),
		),
		Event: sc.NewOption[sc.Compact](sc.ToCompact(metadata.TypesSystemEvent)),
		EventDef: sc.NewOption[primitives.MetadataDefinitionVariant](
			primitives.NewMetadataDefinitionVariantStr(
				name,
				sc.Sequence[primitives.MetadataTypeDefinitionField]{
					primitives.NewMetadataTypeDefinitionFieldWithName(metadata.TypesSystemEvent, "frame_system::Event<Runtime>"),
				},
				moduleId,
				"Events.System"),
		),
		Constants: sc.Sequence[primitives.MetadataModuleConstant]{
			primitives.NewMetadataModuleConstant(
				"BlockWeights",
				sc.ToCompact(expectedTypesBlockWeightsId),
				sc.BytesToSequenceU8(blockWeights.Bytes()),
				"Block & extrinsics weights: base values and limits.",
			),
			primitives.NewMetadataModuleConstant(
				"BlockLength",
				sc.ToCompact(expectedTypesBlockLengthId),
				sc.BytesToSequenceU8(blockLength.Bytes()),
				"The maximum length of a block (in bytes).",
			),
			primitives.NewMetadataModuleConstant(
				"BlockHashCount",
				sc.ToCompact(metadata.PrimitiveTypesU32),
				sc.BytesToSequenceU8(sc.U32(blockHashCount).Bytes()),
				"Maximum number of block number to block hash mappings to keep (oldest pruned first).",
			),
			primitives.NewMetadataModuleConstant(
				"DbWeight",
				sc.ToCompact(expectedTypesDbWeightId),
				sc.BytesToSequenceU8(dbWeight.Bytes()),
				"The weight of runtime database operations the runtime can invoke.",
			),
			primitives.NewMetadataModuleConstant(
				"Version",
				sc.ToCompact(metadata.TypesRuntimeVersion),
				sc.BytesToSequenceU8(version.Bytes()),
				"Get the chain's current version.",
			),
		},
		Error: sc.NewOption[sc.Compact](sc.ToCompact(expectedSystemErrorsId)),
		ErrorDef: sc.NewOption[primitives.MetadataDefinitionVariant](
			primitives.NewMetadataDefinitionVariantStr(
				name,
				sc.Sequence[primitives.MetadataTypeDefinitionField]{
					primitives.NewMetadataTypeDefinitionField(expectedSystemErrorsId),
				},
				moduleId,
				"Errors.System"),
		),
		Index: moduleId,
	}

	expectMetadataModule := primitives.MetadataModule{
		Version:   primitives.ModuleVersion14,
		ModuleV14: moduleV14,
	}

	resultMetadataModule := target.Metadata()
	resultTypes := mdGenerator.GetMetadataTypes()

	assert.Equal(t, expectMetadataTypes, resultTypes)
	assert.Equal(t, expectMetadataModule, resultMetadataModule)
}

func testDigest() primitives.Digest {
	items := sc.Sequence[primitives.DigestItem]{
		primitives.NewDigestItemSeal(
			sc.BytesToFixedSequenceU8([]byte{'t', 'e', 's', 't'}),
			sc.BytesToSequenceU8(sc.U64(5).Bytes()),
		),
	}
	return primitives.NewDigest(items)
}

func Test_Clear_Metadata(t *testing.T) {
	target := setupModule()

	target.Metadata()

	target.mdGenerator.ClearMetadata()

	types := target.mdGenerator.GetMetadataTypes()

	ids := target.mdGenerator.GetIdsMap()

	assert.Equal(t, sc.Sequence[primitives.MetadataType]{}, types)
	assert.Equal(t, ids, primitives.BuildMetadataTypesIdsMap())
}

func Test_Build_Module_Constants(t *testing.T) {
	target := setupModule()
	expectedBlockWeightsId := target.mdGenerator.GetLastAvailableIndex() + 1

	type testConsts struct {
		BlockWeights   primitives.BlockWeights
		RuntimeVersion primitives.RuntimeVersion
	}

	c := testConsts{
		BlockWeights:   blockWeights,
		RuntimeVersion: version,
	}

	expectedConstants := sc.Sequence[primitives.MetadataModuleConstant]{
		primitives.NewMetadataModuleConstant(
			"BlockWeights",
			sc.ToCompact(expectedBlockWeightsId),
			sc.BytesToSequenceU8(blockWeights.Bytes()),
			"Block & extrinsics weights: base values and limits.",
		),
		primitives.NewMetadataModuleConstant(
			"RuntimeVersion",
			sc.ToCompact(metadata.TypesRuntimeVersion),
			sc.BytesToSequenceU8(version.Bytes()),
			"Get the chain's current version.",
		),
	}
	result := target.mdGenerator.BuildModuleConstants(reflect.ValueOf(c))
	assert.Equal(t, expectedConstants, result)
}

func Test_CanSetCode_ErrorFailedToExtractRuntimeVersion(t *testing.T) {
	target := setupModule()

	runtimeVersionBytes := sc.NewOption[sc.Sequence[sc.U8]](sc.Sequence[sc.U8]{0}).Bytes()

	mockIoMisc.On("RuntimeVersion", sc.SequenceU8ToBytes(codeBlob)).Return(runtimeVersionBytes)

	err := target.CanSetCode(codeBlob)
	assert.Equal(t, NewDispatchErrorFailedToExtractRuntimeVersion(target.Index), err)
}

func Test_CanSetCode_ErrorInvalidSpecName(t *testing.T) {
	target := setupModule()

	v2 := version
	v2.SpecName = "test-spec-2"
	versionBytes := v2.Bytes()
	runtimeVersionBytes := sc.NewOption[sc.Sequence[sc.U8]](sc.BytesToSequenceU8(versionBytes)).Bytes()

	mockIoMisc.On("RuntimeVersion", sc.SequenceU8ToBytes(codeBlob)).Return(runtimeVersionBytes)

	err := target.CanSetCode(codeBlob)
	assert.Equal(t, NewDispatchErrorInvalidSpecName(target.Index), err)
}

func Test_CanSetCode_ErrorSpecVersionNeedsToIncrease(t *testing.T) {
	target := setupModule()

	versionBytes := version.Bytes()
	runtimeVersionBytes := sc.NewOption[sc.Sequence[sc.U8]](sc.BytesToSequenceU8(versionBytes)).Bytes()

	mockIoMisc.On("RuntimeVersion", sc.SequenceU8ToBytes(codeBlob)).Return(runtimeVersionBytes)

	err := target.CanSetCode(codeBlob)
	assert.Equal(t, NewDispatchErrorSpecVersionNeedsToIncrease(target.Index), err)
}

func Test_CanSetCode_Success(t *testing.T) {
	target := setupModule()

	v2 := version
	v2.SpecVersion += 1
	versionBytes := v2.Bytes()
	runtimeVersionBytes := sc.NewOption[sc.Sequence[sc.U8]](sc.BytesToSequenceU8(versionBytes)).Bytes()

	mockIoMisc.On("RuntimeVersion", sc.SequenceU8ToBytes(codeBlob)).Return(runtimeVersionBytes)

	err := target.CanSetCode(codeBlob)

	assert.NoError(t, err)
}

func Test_DoAuthorizeUpgrade_Success(t *testing.T) {
	target := setupModule()

	codeHash, err := primitives.NewH256(sc.BytesToFixedSequenceU8(hashBytes)...)
	assert.NoError(t, err)

	checkVersion := sc.Bool(true)

	expectEventRecord := primitives.EventRecord{
		Phase:  primitives.NewExtrinsicPhaseInitialization(),
		Event:  newEventUpgradeAuthorized(moduleId, codeHash, checkVersion),
		Topics: []primitives.H256{},
	}

	upgradeAuthorization := CodeUpgradeAuthorization{codeHash, checkVersion}

	mockStorageAuthorizedUpgrade.On("Put", upgradeAuthorization).Return()
	mockStorageBlockNumber.On("Get").Return(blockNum, nil)
	mockStorageExecutionPhase.On("Get").Return(primitives.NewExtrinsicPhaseInitialization(), nil)
	mockStorageEventCount.On("Get").Return(eventCount, nil)
	mockStorageEventCount.On("Put", eventCount+1).Return()
	mockStorageEvents.On("Append", expectEventRecord).Return()

	target.DoAuthorizeUpgrade(codeHash, checkVersion)

	mockStorageAuthorizedUpgrade.AssertCalled(t, "Put", upgradeAuthorization)
	mockStorageBlockNumber.AssertCalled(t, "Get")
	mockStorageExecutionPhase.AssertCalled(t, "Get")
	mockStorageEventCount.AssertCalled(t, "Get")
	mockStorageEventCount.AssertCalled(t, "Put", eventCount+1)
	mockStorageEvents.AssertCalled(t, "Append", expectEventRecord)
}

func Test_DoApplyAuthorizeUpgrade_Success(t *testing.T) {
	target := setupModule()

	v2 := version
	v2.SpecVersion += 1
	versionBytes := v2.Bytes()
	runtimeVersionBytes := sc.NewOption[sc.Sequence[sc.U8]](sc.BytesToSequenceU8(versionBytes)).Bytes()

	codeHash, err := primitives.NewH256(sc.BytesToFixedSequenceU8(hashBytes)...)
	assert.NoError(t, err)

	checkVersion := sc.Bool(true)
	upgradeAuthorization := CodeUpgradeAuthorization{codeHash, checkVersion}

	digestItem := primitives.NewDigestItemRuntimeEnvironmentUpgrade()

	phase := primitives.NewExtrinsicPhaseInitialization()

	eventRecord := primitives.EventRecord{
		Phase:  phase,
		Event:  newEventCodeUpdated(moduleId),
		Topics: []primitives.H256{},
	}

	mockStorageAuthorizedUpgrade.On("Get").Return(upgradeAuthorization, nil)

	mockIoHashing.On("Blake256", sc.SequenceU8ToBytes(codeBlob)).Return(hashBytes)
	mockIoMisc.On("RuntimeVersion", sc.SequenceU8ToBytes(codeBlob)).Return(runtimeVersionBytes)
	mockStorageCode.On("Put", codeBlob).Return()

	mockStorageDigest.On("AppendItem", digestItem).Return()

	mockStorageBlockNumber.On("Get").Return(blockNum, nil)
	mockStorageExecutionPhase.On("Get").Return(phase, nil)
	mockStorageEventCount.On("Get").Return(eventCount, nil)
	mockStorageEventCount.On("Put", eventCount+1).Return()
	mockStorageEvents.On("Append", eventRecord).Return()

	mockStorageAuthorizedUpgrade.On("Clear").Return()

	_, err = target.DoApplyAuthorizeUpgrade(codeBlob)

	assert.NoError(t, err)
	mockStorageAuthorizedUpgrade.AssertCalled(t, "Get")

	mockIoHashing.AssertCalled(t, "Blake256", sc.SequenceU8ToBytes(codeBlob))
	mockIoMisc.AssertCalled(t, "RuntimeVersion", sc.SequenceU8ToBytes(codeBlob))
	mockStorageCode.AssertCalled(t, "Put", codeBlob)
	mockStorageDigest.AssertCalled(t, "AppendItem", digestItem)

	mockStorageBlockNumber.AssertCalled(t, "Get")
	mockStorageExecutionPhase.AssertCalled(t, "Get")
	mockStorageEventCount.AssertCalled(t, "Get")
	mockStorageEventCount.AssertCalled(t, "Put", eventCount+1)
	mockStorageEvents.AssertCalled(t, "Append", eventRecord)

	mockStorageAuthorizedUpgrade.AssertCalled(t, "Clear")
}

func setupModule() module {
	config := NewConfig(primitives.BlockHashCount{U32: sc.U32(blockHashCount)}, blockWeights, blockLength, dbWeight, &version)

	target := New(moduleId, config, mdGenerator, log.NewLogger()).(module)

	initMockStorage()

	target.storage.Account = mockStorageAccount
	target.storage.BlockWeight = mockStorageBlockWeight
	target.storage.BlockHash = mockStorageBlockHash
	target.storage.BlockNumber = mockStorageBlockNumber
	target.storage.AllExtrinsicsLen = mockStorageAllExtrinsicsLen
	target.storage.ExtrinsicIndex = mockStorageExtrinsicIndex
	target.storage.ExtrinsicData = mockStorageExtrinsicData
	target.storage.ExtrinsicCount = mockStorageExtrinsicCount
	target.storage.ParentHash = mockStorageParentHash
	target.storage.Digest = mockStorageDigest
	target.storage.Events = mockStorageEvents
	target.storage.EventCount = mockStorageEventCount
	target.storage.EventTopics = mockStorageEventTopics
	target.storage.LastRuntimeUpgrade = mockStorageLastRuntimeUpgrade
	target.storage.ExecutionPhase = mockStorageExecutionPhase
	target.storage.HeapPages = mockStorageHeapPages
	target.storage.Code = mockStorageCode
	target.storage.AuthorizedUpgrade = mockStorageAuthorizedUpgrade

	target.ioStorage = mockIoStorage
	target.ioHashing = mockIoHashing
	target.ioMisc = mockIoMisc
	target.trie = mockIoTrie

	return target
}

func initMockStorage() {
	mockStorageAccount = new(mocks.StorageMap[primitives.AccountId, primitives.AccountInfo])
	mockStorageBlockWeight = new(mocks.StorageValue[primitives.ConsumedWeight])
	mockStorageBlockHash = new(mocks.StorageMap[sc.U64, primitives.Blake2bHash])
	mockStorageBlockNumber = new(mocks.StorageValue[sc.U64])
	mockStorageAllExtrinsicsLen = new(mocks.StorageValue[sc.U32])
	mockStorageExtrinsicIndex = new(mocks.StorageValue[sc.U32])
	mockStorageExtrinsicData = new(mocks.StorageMap[sc.U32, sc.Sequence[sc.U8]])
	mockStorageExtrinsicCount = new(mocks.StorageValue[sc.U32])
	mockStorageParentHash = new(mocks.StorageValue[primitives.Blake2bHash])
	mockStorageDigest = new(mocks.StorageValue[primitives.Digest])
	mockStorageEvents = new(mocks.StorageValue[primitives.EventRecord])
	mockStorageEventCount = new(mocks.StorageValue[sc.U32])
	mockStorageEventTopics = new(mocks.StorageMap[primitives.H256, sc.VaryingData])
	mockStorageLastRuntimeUpgrade = new(mocks.StorageValue[primitives.LastRuntimeUpgradeInfo])
	mockStorageExecutionPhase = new(mocks.StorageValue[primitives.ExtrinsicPhase])
	mockStorageHeapPages = new(mocks.StorageValue[sc.U64])
	mockStorageCode = new(mocks.RawStorageValue)
	mockStorageAuthorizedUpgrade = new(mocks.StorageValue[CodeUpgradeAuthorization])

	mockIoStorage = new(mocks.IoStorage)
	mockIoHashing = new(mocks.IoHashing)
	mockIoMisc = new(mocks.IoMisc)
	mockIoTrie = new(mocks.IoTrie)
}
