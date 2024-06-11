package executive

import (
	"errors"
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/execution/extrinsic"
	"github.com/LimeChain/gosemble/execution/types"
	"github.com/LimeChain/gosemble/mocks"
	"github.com/LimeChain/gosemble/primitives/log"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	runtimeVersion = &primitives.RuntimeVersion{
		SpecName:           "new-version",
		ImplName:           "new-version",
		AuthoringVersion:   1,
		SpecVersion:        100,
		ImplVersion:        1,
		TransactionVersion: 1,
		StateVersion:       1,
	}

	oldUpgradeInfo = primitives.LastRuntimeUpgradeInfo{
		SpecVersion: sc.Compact{Number: sc.U32(99)},
		SpecName:    "old-version",
	}

	currentUpgradeInfo = primitives.LastRuntimeUpgradeInfo{
		SpecVersion: sc.Compact{Number: sc.U32(100)},
		SpecName:    "new-version",
	}

	blockWeights = primitives.BlockWeights{
		BaseBlock: primitives.WeightFromParts(1, 1),
		MaxBlock:  primitives.WeightFromParts(7, 7),
	}

	consumedWeight = primitives.ConsumedWeight{
		Normal:      primitives.WeightFromParts(1, 1),
		Operational: primitives.WeightFromParts(2, 2),
		Mandatory:   primitives.WeightFromParts(3, 3),
	}

	baseWeight = primitives.WeightFromParts(1, 1)

	dispatchClassNormal    = primitives.NewDispatchClassNormal()
	dispatchClassMandatory = primitives.NewDispatchClassMandatory()

	dispatchInfo = primitives.DispatchInfo{
		Weight:  primitives.WeightFromParts(2, 2),
		Class:   dispatchClassNormal,
		PaysFee: primitives.PaysYes,
	}

	dispatchErr = primitives.NewDispatchErrorBadOrigin()

	unsignedValidator primitives.UnsignedValidator

	txSource = primitives.NewTransactionSourceExternal()

	defaultDigest = primitives.Digest{}

	blockNumber = sc.U64(1)

	blake256Hash = []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31}
	blockHash, _ = primitives.NewBlake2bHash(sc.BytesToSequenceU8(blake256Hash)...)

	header = primitives.Header{
		Number:     blockNumber,
		ParentHash: blockHash,
		Digest:     testDigest(),
	}

	block = types.NewBlock(header, sc.Sequence[primitives.UncheckedExtrinsic]{})

	encodedExtrinsic    = []byte{0, 1, 2, 3, 4, 5}
	encodedExtrinsicLen = sc.ToCompact(len(encodedExtrinsic))

	signer = sc.Option[primitives.Address32]{}

	errPanic = errors.New("panic")
)

var (
	unknownTransactionCannotLookupError = primitives.NewTransactionValidityError(
		primitives.NewUnknownTransactionCannotLookup(),
	)

	invalidTransactionExhaustsResourcesError = primitives.NewTransactionValidityError(
		primitives.NewInvalidTransactionExhaustsResources(),
	)

	invalidTransactionBadMandatory = primitives.NewTransactionValidityError(
		primitives.NewInvalidTransactionBadMandatory(),
	)

	invalidTransactionMandatoryValidation = primitives.NewTransactionValidityError(
		primitives.NewInvalidTransactionMandatoryValidation(),
	)

	defaultDispatchOutcome  = primitives.DispatchOutcome{}
	defaultValidTransaction = primitives.ValidTransaction{}
)

var (
	target module

	mockSystemModule                  *mocks.SystemModule
	mockRuntimeExtrinsic              *mocks.RuntimeExtrinsic
	mockOnRuntimeUpgradeHook          *mocks.DefaultOnRuntimeUpgrade
	mockUncheckedExtrinsic            *mocks.UncheckedExtrinsic
	mockSignedExtra                   *mocks.SignedExtra
	mockCheckedExtrinsic              *mocks.CheckedExtrinsic
	mockCall                          *mocks.Call
	mockStorageLastRuntimeUpgradeInfo *mocks.StorageValue[primitives.LastRuntimeUpgradeInfo]
	mockStorageBlockHash              *mocks.StorageMap[sc.U64, primitives.Blake2bHash]
	mockStorageBlockNumber            *mocks.StorageValue[sc.U64]
	mockStorageBlockWeight            *mocks.StorageValue[primitives.ConsumedWeight]
	mockIoHashing                     *mocks.IoHashing
)

func setup() {
	mockSystemModule = new(mocks.SystemModule)
	mockRuntimeExtrinsic = new(mocks.RuntimeExtrinsic)
	mockOnRuntimeUpgradeHook = new(mocks.DefaultOnRuntimeUpgrade)
	mockUncheckedExtrinsic = new(mocks.UncheckedExtrinsic)
	mockSignedExtra = new(mocks.SignedExtra)
	mockCheckedExtrinsic = new(mocks.CheckedExtrinsic)
	mockCall = new(mocks.Call)
	mockStorageLastRuntimeUpgradeInfo = new(mocks.StorageValue[primitives.LastRuntimeUpgradeInfo])
	mockStorageBlockHash = new(mocks.StorageMap[sc.U64, primitives.Blake2bHash])
	mockStorageBlockNumber = new(mocks.StorageValue[sc.U64])
	mockStorageBlockWeight = new(mocks.StorageValue[primitives.ConsumedWeight])
	mockIoHashing = new(mocks.IoHashing)
	logger := log.NewLogger()

	target = New(
		mockSystemModule,
		mockRuntimeExtrinsic,
		mockOnRuntimeUpgradeHook,
		logger,
	).(module)
	target.hashing = mockIoHashing

	unsignedValidator = extrinsic.NewUnsignedValidatorForChecked(mockRuntimeExtrinsic)
}

func testDigest() primitives.Digest {
	items := sc.Sequence[primitives.DigestItem]{
		primitives.NewDigestItemPreRuntime(
			sc.BytesToFixedSequenceU8([]byte{'a', 'u', 'r', 'a'}),
			sc.BytesToSequenceU8(sc.U64(0).Bytes()),
		),
	}
	return primitives.NewDigest(items)
}

func Test_Executive_InitializeBlock_VersionUpgraded(t *testing.T) {
	setup()

	mockSystemModule.On("ResetEvents").Return()
	mockSystemModule.On("StorageLastRuntimeUpgrade").Return(oldUpgradeInfo, nil)
	mockSystemModule.On("Version").Return(*runtimeVersion)
	mockSystemModule.On("StorageLastRuntimeUpgradeSet", currentUpgradeInfo)
	mockOnRuntimeUpgradeHook.On("OnRuntimeUpgrade").Return(primitives.WeightFromParts(1, 1))
	mockRuntimeExtrinsic.On("OnRuntimeUpgrade").Return(primitives.WeightFromParts(2, 2))
	mockSystemModule.On("Initialize", header.Number, header.ParentHash, header.Digest)
	mockRuntimeExtrinsic.On("OnInitialize", header.Number).Return(primitives.WeightFromParts(3, 3), nil)
	mockSystemModule.On("BlockWeights").Return(blockWeights)
	mockSystemModule.On("RegisterExtraWeightUnchecked", primitives.WeightFromParts(7, 7), dispatchClassMandatory).Return(nil)
	mockSystemModule.On("NoteFinishedInitialize")

	target.InitializeBlock(header)

	mockSystemModule.AssertCalled(t, "ResetEvents")
	mockSystemModule.AssertCalled(t, "StorageLastRuntimeUpgradeSet", currentUpgradeInfo)
	mockOnRuntimeUpgradeHook.AssertCalled(t, "OnRuntimeUpgrade")
	mockRuntimeExtrinsic.AssertCalled(t, "OnRuntimeUpgrade")
	mockSystemModule.AssertCalled(t, "Initialize", header.Number, header.ParentHash, header.Digest)
	mockRuntimeExtrinsic.AssertCalled(t, "OnInitialize", header.Number)
	mockSystemModule.AssertCalled(t, "RegisterExtraWeightUnchecked", primitives.WeightFromParts(7, 7), dispatchClassMandatory)
	mockSystemModule.AssertCalled(t, "NoteFinishedInitialize")
}

func Test_Executive_InitializeBlock_VersionNotUpgraded(t *testing.T) {
	setup()

	mockSystemModule.On("ResetEvents").Return()
	mockSystemModule.On("StorageLastRuntimeUpgrade").Return(currentUpgradeInfo, nil)
	mockSystemModule.On("Version").Return(*runtimeVersion)
	mockSystemModule.On("Initialize", header.Number, header.ParentHash, header.Digest)
	mockRuntimeExtrinsic.On("OnInitialize", header.Number).Return(primitives.WeightFromParts(3, 3), nil)
	mockSystemModule.On("BlockWeights").Return(blockWeights)
	mockSystemModule.On("RegisterExtraWeightUnchecked", primitives.WeightFromParts(4, 4), dispatchClassMandatory).Return(nil)
	mockSystemModule.On("NoteFinishedInitialize")

	target.InitializeBlock(header)

	mockSystemModule.AssertCalled(t, "ResetEvents")
	mockStorageLastRuntimeUpgradeInfo.AssertNotCalled(t, "Put", currentUpgradeInfo)
	mockOnRuntimeUpgradeHook.AssertNotCalled(t, "OnRuntimeUpgrade")
	mockRuntimeExtrinsic.AssertNotCalled(t, "OnRuntimeUpgrade")
	mockSystemModule.AssertCalled(t, "Initialize", header.Number, header.ParentHash, header.Digest)
	mockRuntimeExtrinsic.AssertCalled(t, "OnInitialize", header.Number)
	mockSystemModule.AssertCalled(t, "RegisterExtraWeightUnchecked", primitives.WeightFromParts(4, 4), dispatchClassMandatory)
	mockSystemModule.AssertCalled(t, "NoteFinishedInitialize")
}

func Test_Executive_InitializeBlock_RegisterExtraWeightUnchecked_Error(t *testing.T) {
	setup()

	mockSystemModule.On("ResetEvents").Return()
	mockSystemModule.On("StorageLastRuntimeUpgrade").Return(currentUpgradeInfo, nil)
	mockSystemModule.On("Version").Return(*runtimeVersion)
	mockSystemModule.On("Initialize", header.Number, header.ParentHash, header.Digest)
	mockRuntimeExtrinsic.On("OnInitialize", header.Number).Return(primitives.WeightFromParts(3, 3), nil)
	mockSystemModule.On("BlockWeights").Return(blockWeights)
	mockSystemModule.On("RegisterExtraWeightUnchecked", primitives.WeightFromParts(4, 4), dispatchClassMandatory).Return(errPanic)

	err := target.InitializeBlock(header)

	assert.Equal(t, errPanic, err)
	mockSystemModule.AssertCalled(t, "ResetEvents")
	mockStorageLastRuntimeUpgradeInfo.AssertNotCalled(t, "Put", currentUpgradeInfo)
	mockOnRuntimeUpgradeHook.AssertNotCalled(t, "OnRuntimeUpgrade")
	mockRuntimeExtrinsic.AssertNotCalled(t, "OnRuntimeUpgrade")
	mockSystemModule.AssertCalled(t, "Initialize", header.Number, header.ParentHash, header.Digest)
	mockRuntimeExtrinsic.AssertCalled(t, "OnInitialize", header.Number)
	mockSystemModule.AssertCalled(t, "RegisterExtraWeightUnchecked", primitives.WeightFromParts(4, 4), dispatchClassMandatory)
}

func Test_Executive_InitializeBlock_OnInitialize_Error(t *testing.T) {
	setup()

	mockSystemModule.On("ResetEvents").Return()
	mockSystemModule.On("StorageLastRuntimeUpgrade").Return(currentUpgradeInfo, nil)
	mockSystemModule.On("Version").Return(*runtimeVersion)
	mockSystemModule.On("Initialize", header.Number, header.ParentHash, header.Digest)
	mockRuntimeExtrinsic.On("OnInitialize", header.Number).Return(primitives.WeightFromParts(3, 3), errPanic)

	err := target.InitializeBlock(header)

	assert.Equal(t, errPanic, err)
	mockSystemModule.AssertCalled(t, "ResetEvents")
	mockStorageLastRuntimeUpgradeInfo.AssertNotCalled(t, "Put", currentUpgradeInfo)
	mockOnRuntimeUpgradeHook.AssertNotCalled(t, "OnRuntimeUpgrade")
	mockRuntimeExtrinsic.AssertNotCalled(t, "OnRuntimeUpgrade")
	mockSystemModule.AssertCalled(t, "Initialize", header.Number, header.ParentHash, header.Digest)
	mockRuntimeExtrinsic.AssertCalled(t, "OnInitialize", header.Number)
}

func Test_Executive_ExecuteBlock_InvalidParentHash(t *testing.T) {
	setup()

	mockSystemModule.On("ResetEvents").Return()
	mockSystemModule.On("StorageLastRuntimeUpgrade").Return(currentUpgradeInfo, nil)
	mockSystemModule.On("Version").Return(*runtimeVersion)
	mockSystemModule.On("Initialize", header.Number, header.ParentHash, header.Digest)
	mockRuntimeExtrinsic.On("OnInitialize", header.Number).Return(primitives.WeightFromParts(3, 3), nil)
	mockSystemModule.On("BlockWeights").Return(blockWeights)
	mockSystemModule.On("RegisterExtraWeightUnchecked", primitives.WeightFromParts(4, 4), dispatchClassMandatory).Return(nil)
	mockSystemModule.On("NoteFinishedInitialize")

	invalidParentHash, _ := primitives.NewBlake2bHash(sc.BytesToSequenceU8([]byte("abcdefghijklmnopqrstuvwxyz123450"))...)
	mockSystemModule.On("StorageBlockHash", header.Number-1).Return(invalidParentHash, nil)

	err := target.ExecuteBlock(block)

	assert.Equal(t, errInvalidParentHash, err)
}

func Test_Executive_ExecuteBlock_InvalidInherentPosition(t *testing.T) {
	setup()

	header := primitives.Header{
		Number:     sc.U64(0),
		ParentHash: blockHash,
		Digest:     testDigest(),
	}

	block := types.NewBlock(header, sc.Sequence[primitives.UncheckedExtrinsic]{})

	mockSystemModule.On("ResetEvents").Return()
	mockSystemModule.On("StorageLastRuntimeUpgrade").Return(currentUpgradeInfo, nil)
	mockSystemModule.On("Version").Return(*runtimeVersion)
	mockSystemModule.On("Initialize", header.Number, header.ParentHash, header.Digest)
	mockRuntimeExtrinsic.On("OnInitialize", header.Number).Return(primitives.WeightFromParts(3, 3), nil)
	mockSystemModule.On("BlockWeights").Return(blockWeights)
	mockSystemModule.On("RegisterExtraWeightUnchecked", primitives.WeightFromParts(4, 4), dispatchClassMandatory).Return(nil)
	mockSystemModule.On("NoteFinishedInitialize")
	mockRuntimeExtrinsic.On("EnsureInherentsAreFirst", block).Return(0)

	err := target.ExecuteBlock(block)

	assert.Equal(t, "invalid inherent position for extrinsic at index [0]", err.Error())
}

func Test_Executive_ExecuteBlock_StorageBlockHash_Error(t *testing.T) {
	setup()

	header := primitives.Header{
		Number:     sc.U64(1),
		ParentHash: blockHash,
		Digest:     testDigest(),
	}

	block := types.NewBlock(header, sc.Sequence[primitives.UncheckedExtrinsic]{})

	mockSystemModule.On("ResetEvents").Return()
	mockSystemModule.On("StorageLastRuntimeUpgrade").Return(currentUpgradeInfo, nil)
	mockSystemModule.On("Version").Return(*runtimeVersion)
	mockSystemModule.On("Initialize", header.Number, header.ParentHash, header.Digest)
	mockRuntimeExtrinsic.On("OnInitialize", header.Number).Return(primitives.WeightFromParts(3, 3), nil)
	mockSystemModule.On("BlockWeights").Return(blockWeights)
	mockSystemModule.On("RegisterExtraWeightUnchecked", primitives.WeightFromParts(4, 4), dispatchClassMandatory).Return(nil)
	mockSystemModule.On("NoteFinishedInitialize")
	mockSystemModule.On("StorageBlockHash", header.Number-1).Return(blockHash, errPanic)

	err := target.ExecuteBlock(block)

	assert.Equal(t, errPanic, err)
	mockSystemModule.AssertCalled(t, "StorageBlockHash", header.Number-1)
}

func Test_Executive_ExecuteBlock_InitializeBlock_Error(t *testing.T) {
	setup()

	header := primitives.Header{
		Number:     sc.U64(0),
		ParentHash: blockHash,
		Digest:     testDigest(),
	}

	block := types.NewBlock(header, sc.Sequence[primitives.UncheckedExtrinsic]{})

	mockSystemModule.On("ResetEvents").Return()
	mockSystemModule.On("StorageLastRuntimeUpgrade").Return(currentUpgradeInfo, errPanic)

	err := target.ExecuteBlock(block)

	assert.Equal(t, errPanic, err)
	mockSystemModule.AssertCalled(t, "StorageLastRuntimeUpgrade")
}

func Test_Executive_ExecuteBlock_Success(t *testing.T) {
	setup()

	blockWeights := primitives.BlockWeights{
		BaseBlock: primitives.WeightFromParts(1, 1),
		MaxBlock:  primitives.WeightFromParts(6, 6),
	}
	header := primitives.Header{
		Number:     sc.U64(0),
		ParentHash: blockHash,
		Digest:     testDigest(),
	}

	block := types.NewBlock(header, sc.Sequence[primitives.UncheckedExtrinsic]{})

	mockSystemModule.On("ResetEvents").Return()
	mockSystemModule.On("StorageLastRuntimeUpgrade").Return(currentUpgradeInfo, nil)
	mockSystemModule.On("Version").Return(*runtimeVersion)
	mockSystemModule.On("Initialize", header.Number, header.ParentHash, header.Digest)
	mockRuntimeExtrinsic.On("OnInitialize", header.Number).Return(primitives.WeightFromParts(3, 3), nil)
	mockSystemModule.On("BlockWeights").Return(blockWeights)
	mockSystemModule.On("RegisterExtraWeightUnchecked", primitives.WeightFromParts(4, 4), dispatchClassMandatory).Return(nil)
	mockSystemModule.On("NoteFinishedInitialize")
	mockSystemModule.On("StorageBlockHash", header.Number-1).Return(blockHash, nil)
	mockRuntimeExtrinsic.On("EnsureInherentsAreFirst", block).Return(-1)
	mockSystemModule.On("NoteFinishedExtrinsics").Return(nil)
	mockSystemModule.On("StorageBlockWeight").Return(consumedWeight, nil)
	mockSystemModule.On("BlockWeights").Return(blockWeights)
	mockRuntimeExtrinsic.On("OnFinalize", blockNumber-1).Return(nil)
	mockSystemModule.On("Finalize").Return(header, nil)

	target.ExecuteBlock(block)

	mockSystemModule.AssertCalled(t, "ResetEvents")
	mockSystemModule.AssertNotCalled(t, "StorageLastRuntimeUpgradeSet", currentUpgradeInfo)
	mockOnRuntimeUpgradeHook.AssertNotCalled(t, "OnRuntimeUpgrade")
	mockRuntimeExtrinsic.AssertNotCalled(t, "OnRuntimeUpgrade")
	mockSystemModule.AssertCalled(t, "Initialize", header.Number, header.ParentHash, header.Digest)
	mockSystemModule.AssertCalled(t, "RegisterExtraWeightUnchecked", primitives.WeightFromParts(4, 4), dispatchClassMandatory)
	mockSystemModule.AssertCalled(t, "NoteFinishedInitialize")
	mockSystemModule.AssertNotCalled(t, "StorageBlockHash", header.Number-1)
	mockRuntimeExtrinsic.AssertCalled(t, "EnsureInherentsAreFirst", block)
	mockSystemModule.AssertCalled(t, "NoteFinishedExtrinsics")
	mockSystemModule.AssertCalled(t, "StorageBlockWeight")
	mockSystemModule.AssertCalled(t, "BlockWeights")
	mockRuntimeExtrinsic.AssertNotCalled(t, "OnIdle")
	mockSystemModule.AssertNotCalled(t, "RegisterExtraWeightUnchecked")
	mockRuntimeExtrinsic.AssertCalled(t, "OnFinalize", blockNumber-1)
	mockSystemModule.AssertCalled(t, "Finalize")
}
func Test_Executive_ExecuteBlock_Finalize_Error(t *testing.T) {
	setup()

	blockWeights := primitives.BlockWeights{
		BaseBlock: primitives.WeightFromParts(1, 1),
		MaxBlock:  primitives.WeightFromParts(6, 6),
	}
	header := primitives.Header{
		Number:     sc.U64(0),
		ParentHash: blockHash,
		Digest:     testDigest(),
	}

	block := types.NewBlock(header, sc.Sequence[primitives.UncheckedExtrinsic]{})

	mockSystemModule.On("ResetEvents").Return()
	mockSystemModule.On("StorageLastRuntimeUpgrade").Return(currentUpgradeInfo, nil)
	mockSystemModule.On("Version").Return(*runtimeVersion)
	mockSystemModule.On("Initialize", header.Number, header.ParentHash, header.Digest)
	mockRuntimeExtrinsic.On("OnInitialize", header.Number).Return(primitives.WeightFromParts(3, 3), nil)
	mockSystemModule.On("BlockWeights").Return(blockWeights)
	mockSystemModule.On("RegisterExtraWeightUnchecked", primitives.WeightFromParts(4, 4), dispatchClassMandatory).Return(nil)
	mockSystemModule.On("NoteFinishedInitialize")
	mockSystemModule.On("StorageBlockHash", header.Number-1).Return(blockHash, nil)
	mockRuntimeExtrinsic.On("EnsureInherentsAreFirst", block).Return(-1)
	mockSystemModule.On("NoteFinishedExtrinsics").Return(nil)
	mockSystemModule.On("StorageBlockWeight").Return(consumedWeight, nil)
	mockSystemModule.On("BlockWeights").Return(blockWeights)
	mockRuntimeExtrinsic.On("OnFinalize", blockNumber-1).Return(nil)
	mockSystemModule.On("Finalize").Return(header, errPanic)

	err := target.ExecuteBlock(block)

	assert.Equal(t, errPanic, err)
	mockSystemModule.AssertCalled(t, "ResetEvents")
	mockSystemModule.AssertNotCalled(t, "StorageLastRuntimeUpgradeSet", currentUpgradeInfo)
	mockOnRuntimeUpgradeHook.AssertNotCalled(t, "OnRuntimeUpgrade")
	mockRuntimeExtrinsic.AssertNotCalled(t, "OnRuntimeUpgrade")
	mockSystemModule.AssertCalled(t, "Initialize", header.Number, header.ParentHash, header.Digest)
	mockSystemModule.AssertCalled(t, "RegisterExtraWeightUnchecked", primitives.WeightFromParts(4, 4), dispatchClassMandatory)
	mockSystemModule.AssertCalled(t, "NoteFinishedInitialize")
	mockSystemModule.AssertNotCalled(t, "StorageBlockHash", header.Number-1)
	mockRuntimeExtrinsic.AssertCalled(t, "EnsureInherentsAreFirst", block)
	mockSystemModule.AssertCalled(t, "NoteFinishedExtrinsics")
	mockSystemModule.AssertCalled(t, "StorageBlockWeight")
	mockSystemModule.AssertCalled(t, "BlockWeights")
	mockRuntimeExtrinsic.AssertNotCalled(t, "OnIdle")
	mockSystemModule.AssertNotCalled(t, "RegisterExtraWeightUnchecked")
	mockRuntimeExtrinsic.AssertCalled(t, "OnFinalize", blockNumber-1)
	mockSystemModule.AssertCalled(t, "Finalize")
}

func Test_Executive_ApplyExtrinsic_UnknownTransactionCannotLookupError(t *testing.T) {
	setup()

	mockUncheckedExtrinsic.On("Bytes").Return(encodedExtrinsic)
	mockUncheckedExtrinsic.On("Check").Return(nil, unknownTransactionCannotLookupError)

	err := target.ApplyExtrinsic(mockUncheckedExtrinsic)
	assert.Equal(t, unknownTransactionCannotLookupError, err)
}

func Test_Executive_ApplyExtrinsic_IsMendatory_Error(t *testing.T) {
	setup()

	mockUncheckedExtrinsic.On("Bytes").Return(encodedExtrinsic)
	mockUncheckedExtrinsic.On("Check").Return(mockCheckedExtrinsic, nil)
	mockSystemModule.On("NoteExtrinsic", mockUncheckedExtrinsic.Bytes())
	mockCheckedExtrinsic.On("Function").Return(mockCall)
	mockCall.On("BaseWeight").Return(baseWeight)
	mockCall.On("WeighData", baseWeight).Return(dispatchInfo.Weight)
	mockCall.On("ClassifyDispatch", baseWeight).Return(primitives.DispatchClass{})
	mockCall.On("PaysFee", baseWeight).Return(dispatchInfo.PaysFee)
	mockCheckedExtrinsic.On("Apply", mock.Anything, mock.Anything, mock.Anything).
		Return(primitives.PostDispatchInfo{}, dispatchErr)

	err := target.ApplyExtrinsic(mockUncheckedExtrinsic)
	assert.Equal(t, "not a valid 'DispatchClass' type", err.Error())

	mockCheckedExtrinsic.AssertCalled(t, "Apply", mock.Anything, mock.Anything, mock.Anything)
}

func Test_Executive_ApplyExtrinsic_InvalidTransactionExhaustsResourcesError(t *testing.T) {
	setup()

	mockUncheckedExtrinsic.On("Bytes").Return(encodedExtrinsic)
	mockUncheckedExtrinsic.On("Check").Return(mockCheckedExtrinsic, nil)
	mockSystemModule.On("NoteExtrinsic", mockUncheckedExtrinsic.Bytes())
	mockCheckedExtrinsic.On("Function").Return(mockCall)
	mockCall.On("BaseWeight").Return(baseWeight)
	mockCall.On("WeighData", baseWeight).Return(dispatchInfo.Weight)
	mockCall.On("ClassifyDispatch", baseWeight).Return(dispatchInfo.Class)
	mockCall.On("PaysFee", baseWeight).Return(dispatchInfo.PaysFee)

	mockCheckedExtrinsic.On("Apply", unsignedValidator, &dispatchInfo, encodedExtrinsicLen).
		Return(primitives.PostDispatchInfo{}, invalidTransactionExhaustsResourcesError)

	err := target.ApplyExtrinsic(mockUncheckedExtrinsic)
	assert.Equal(t, invalidTransactionExhaustsResourcesError, err)

	mockSystemModule.AssertCalled(t, "NoteExtrinsic", mockUncheckedExtrinsic.Bytes())
	mockSystemModule.AssertNotCalled(t, "NoteAppliedExtrinsic", mock.Anything, mock.Anything)
}

func Test_Executive_ApplyExtrinsic_InvalidTransactionBadMandatoryError(t *testing.T) {
	setup()

	dispatchInfo := primitives.DispatchInfo{
		Weight:  primitives.WeightFromParts(2, 2),
		Class:   dispatchClassMandatory,
		PaysFee: primitives.PaysYes,
	}

	mockUncheckedExtrinsic.On("Bytes").Return(encodedExtrinsic)
	mockUncheckedExtrinsic.On("Check").Return(mockCheckedExtrinsic, nil)
	mockSystemModule.On("NoteExtrinsic", mockUncheckedExtrinsic.Bytes())
	mockCheckedExtrinsic.On("Function").Return(mockCall)
	mockCall.On("BaseWeight").Return(baseWeight)
	mockCall.On("WeighData", baseWeight).Return(dispatchInfo.Weight)
	mockCall.On("ClassifyDispatch", baseWeight).Return(dispatchInfo.Class)
	mockCall.On("PaysFee", baseWeight).Return(dispatchInfo.PaysFee)
	mockCheckedExtrinsic.On("Apply", unsignedValidator, &dispatchInfo, encodedExtrinsicLen).Return(primitives.PostDispatchInfo{}, dispatchErr)

	err := target.ApplyExtrinsic(mockUncheckedExtrinsic)
	assert.Equal(t, invalidTransactionBadMandatory, err)

	mockSystemModule.AssertCalled(t, "NoteExtrinsic", mockUncheckedExtrinsic.Bytes())
	mockSystemModule.AssertNotCalled(t, "NoteAppliedExtrinsic", mock.Anything, mock.Anything)
}

func Test_Executive_ApplyExtrinsic_NoteAppliedExtrinsic_Error(t *testing.T) {
	setup()

	dispatchInfo := primitives.DispatchInfo{
		Weight:  primitives.WeightFromParts(2, 2),
		Class:   dispatchClassMandatory,
		PaysFee: primitives.PaysYes,
	}

	mockUncheckedExtrinsic.On("Bytes").Return(encodedExtrinsic)
	mockUncheckedExtrinsic.On("Check").Return(mockCheckedExtrinsic, nil)
	mockSystemModule.On("NoteExtrinsic", mockUncheckedExtrinsic.Bytes())
	mockCheckedExtrinsic.On("Function").Return(mockCall)
	mockCall.On("BaseWeight").Return(baseWeight)
	mockCall.On("WeighData", baseWeight).Return(dispatchInfo.Weight)
	mockCall.On("ClassifyDispatch", baseWeight).Return(dispatchInfo.Class)
	mockCall.On("PaysFee", baseWeight).Return(dispatchInfo.PaysFee)
	mockCheckedExtrinsic.On("Apply", unsignedValidator, &dispatchInfo, encodedExtrinsicLen).Return(primitives.PostDispatchInfo{}, nil)
	mockSystemModule.On("NoteAppliedExtrinsic", mock.Anything, mock.Anything, mock.Anything).Return(errPanic)

	err := target.ApplyExtrinsic(mockUncheckedExtrinsic)
	assert.Equal(t, errPanic, err)

	mockSystemModule.AssertCalled(t, "NoteExtrinsic", mockUncheckedExtrinsic.Bytes())
	mockSystemModule.AssertCalled(t, "NoteAppliedExtrinsic", mock.Anything, mock.Anything, mock.Anything)
}

func Test_Executive_ApplyExtrinsic_Success_DispatchOutcomeErr(t *testing.T) {
	setup()

	postDispatchInfo := primitives.PostDispatchInfo{}

	mockUncheckedExtrinsic.On("Bytes").Return(encodedExtrinsic)
	mockUncheckedExtrinsic.On("Check").Return(mockCheckedExtrinsic, nil)
	mockSystemModule.On("NoteExtrinsic", mockUncheckedExtrinsic.Bytes())

	mockCheckedExtrinsic.On("Function").Return(mockCall)
	mockCall.On("BaseWeight").Return(baseWeight)
	mockCall.On("WeighData", baseWeight).Return(dispatchInfo.Weight)
	mockCall.On("ClassifyDispatch", baseWeight).Return(dispatchInfo.Class)
	mockCall.On("PaysFee", baseWeight).Return(dispatchInfo.PaysFee)
	mockCheckedExtrinsic.On("Apply", unsignedValidator, &dispatchInfo, encodedExtrinsicLen).
		Return(postDispatchInfo, dispatchErr)
	mockSystemModule.On("NoteAppliedExtrinsic", postDispatchInfo, dispatchErr, dispatchInfo).Return(nil)

	err := target.ApplyExtrinsic(mockUncheckedExtrinsic)
	assert.Equal(t, dispatchErr, err)

	mockSystemModule.AssertCalled(t, "NoteExtrinsic", mockUncheckedExtrinsic.Bytes())
	mockSystemModule.AssertCalled(t, "NoteAppliedExtrinsic", primitives.PostDispatchInfo{}, dispatchErr, dispatchInfo)
}

func Test_Executive_ApplyExtrinsic_Success_DispatchOutcomeNil(t *testing.T) {
	setup()

	postInfo := primitives.PostDispatchInfo{}

	mockUncheckedExtrinsic.On("Bytes").Return(encodedExtrinsic)
	mockUncheckedExtrinsic.On("Check").Return(mockCheckedExtrinsic, nil)
	mockSystemModule.On("NoteExtrinsic", mockUncheckedExtrinsic.Bytes())

	mockCheckedExtrinsic.On("Function").Return(mockCall)
	mockCall.On("BaseWeight").Return(baseWeight)
	mockCall.On("WeighData", baseWeight).Return(dispatchInfo.Weight)
	mockCall.On("ClassifyDispatch", baseWeight).Return(dispatchInfo.Class)
	mockCall.On("PaysFee", baseWeight).Return(dispatchInfo.PaysFee)
	mockCheckedExtrinsic.On("Apply", unsignedValidator, &dispatchInfo, encodedExtrinsicLen).
		Return(postInfo, nil)
	mockSystemModule.On("NoteAppliedExtrinsic", postInfo, nil, dispatchInfo).Return(nil)

	err := target.ApplyExtrinsic(mockUncheckedExtrinsic)
	assert.Nil(t, err)

	mockSystemModule.AssertCalled(t, "NoteAppliedExtrinsic", postInfo, nil, dispatchInfo)
}

func Test_Executive_FinalizeBlock(t *testing.T) {
	setup()

	blockNumber := sc.U64(3)
	header := primitives.Header{
		Number:     blockNumber,
		ParentHash: blockHash,
		Digest:     testDigest(),
	}

	mockSystemModule.On("NoteFinishedExtrinsics").Return(nil)
	mockSystemModule.On("StorageBlockNumber").Return(blockNumber, nil)
	mockSystemModule.On("StorageBlockWeight").Return(consumedWeight, nil)
	mockSystemModule.On("BlockWeights").Return(blockWeights)
	remainingWeight := primitives.WeightFromParts(1, 1)
	usedWeight := primitives.WeightFromParts(6, 6)
	mockRuntimeExtrinsic.On("OnIdle", blockNumber, remainingWeight).Return(usedWeight)
	mockSystemModule.On("RegisterExtraWeightUnchecked", usedWeight, dispatchClassMandatory).Return(nil)
	mockRuntimeExtrinsic.On("OnFinalize", blockNumber).Return(nil)
	mockSystemModule.On("Finalize").Return(header, nil)

	target.FinalizeBlock()

	mockSystemModule.AssertCalled(t, "NoteFinishedExtrinsics")
	mockSystemModule.AssertCalled(t, "StorageBlockNumber")
	mockSystemModule.AssertCalled(t, "StorageBlockWeight")
	mockSystemModule.AssertCalled(t, "BlockWeights")
	mockRuntimeExtrinsic.AssertCalled(t, "OnIdle", blockNumber, remainingWeight)
	mockSystemModule.AssertCalled(t, "RegisterExtraWeightUnchecked", usedWeight, dispatchClassMandatory)
	mockRuntimeExtrinsic.AssertCalled(t, "OnFinalize", blockNumber)
	mockSystemModule.AssertCalled(t, "Finalize")
}

func Test_Executive_FinalizeBlock_NoteFinishedExtrinsics_Error(t *testing.T) {
	setup()

	mockSystemModule.On("NoteFinishedExtrinsics").Return(errPanic)

	_, err := target.FinalizeBlock()
	assert.Equal(t, errPanic, err)
}

func Test_Executive_FinalizeBlock_StorageBlockNumber_Error(t *testing.T) {
	setup()

	mockSystemModule.On("NoteFinishedExtrinsics").Return(nil)
	mockSystemModule.On("StorageBlockNumber").Return(blockNumber, errPanic)

	_, err := target.FinalizeBlock()
	assert.Equal(t, errPanic, err)

	mockSystemModule.AssertCalled(t, "NoteFinishedExtrinsics")
	mockSystemModule.AssertCalled(t, "StorageBlockNumber")
}

func Test_Executive_FinalizeBlock_idleAndFinalizeHook_Error(t *testing.T) {
	setup()

	mockSystemModule.On("NoteFinishedExtrinsics").Return(nil)
	mockSystemModule.On("StorageBlockNumber").Return(blockNumber, nil)
	mockSystemModule.On("StorageBlockWeight").Return(consumedWeight, errPanic)

	_, err := target.FinalizeBlock()
	assert.Equal(t, errPanic, err)

	mockSystemModule.AssertCalled(t, "NoteFinishedExtrinsics")
	mockSystemModule.AssertCalled(t, "StorageBlockNumber")
	mockSystemModule.AssertCalled(t, "StorageBlockWeight")
}

func Test_Executive_ValidateTransaction_UnknownTransactionCannotLookupError(t *testing.T) {
	setup()

	mockSystemModule.On("StorageBlockNumber").Return(blockNumber, nil)
	mockSystemModule.On("Initialize", blockNumber+1, header.ParentHash, defaultDigest)
	mockUncheckedExtrinsic.On("Bytes").Return(encodedExtrinsic)
	mockUncheckedExtrinsic.On("Check").Return(nil, unknownTransactionCannotLookupError)

	outcome, err := target.ValidateTransaction(txSource, mockUncheckedExtrinsic, header.ParentHash)

	mockSystemModule.AssertCalled(t, "StorageBlockNumber")
	mockSystemModule.AssertCalled(t, "Initialize", blockNumber+1, header.ParentHash, defaultDigest)
	mockUncheckedExtrinsic.AssertCalled(t, "Bytes")
	mockUncheckedExtrinsic.AssertCalled(t, "Check")
	mockCheckedExtrinsic.AssertNotCalled(t, "Validate", unsignedValidator, txSource, &dispatchInfo, encodedExtrinsicLen)
	assert.Equal(t, defaultValidTransaction, outcome)
	assert.Equal(t, unknownTransactionCannotLookupError, err)
}

func Test_Executive_ValidateTransaction_InvalidTransactionMandatoryValidationError(t *testing.T) {
	setup()

	dispatchInfo := primitives.DispatchInfo{
		Weight:  primitives.WeightFromParts(2, 2),
		Class:   dispatchClassMandatory,
		PaysFee: primitives.PaysYes,
	}

	mockSystemModule.On("StorageBlockNumber").Return(blockNumber, nil)
	mockSystemModule.On("Initialize", blockNumber+1, header.ParentHash, defaultDigest)
	mockUncheckedExtrinsic.On("Bytes").Return(encodedExtrinsic)
	mockUncheckedExtrinsic.On("Check").Return(mockCheckedExtrinsic, nil)
	mockCheckedExtrinsic.On("Function").Return(mockCall)
	mockCall.On("BaseWeight").Return(baseWeight)
	mockCall.On("WeighData", baseWeight).Return(dispatchInfo.Weight)
	mockCall.On("ClassifyDispatch", baseWeight).Return(dispatchInfo.Class)
	mockCall.On("PaysFee", baseWeight).Return(dispatchInfo.PaysFee)

	outcome, err := target.ValidateTransaction(txSource, mockUncheckedExtrinsic, header.ParentHash)

	mockSystemModule.AssertCalled(t, "StorageBlockNumber")
	mockSystemModule.AssertCalled(t, "Initialize", blockNumber+1, header.ParentHash, defaultDigest)
	mockUncheckedExtrinsic.AssertCalled(t, "Bytes")
	mockUncheckedExtrinsic.AssertCalled(t, "Check")
	mockCheckedExtrinsic.AssertCalled(t, "Function")
	mockCall.AssertCalled(t, "BaseWeight")
	mockCall.AssertCalled(t, "WeighData", baseWeight)
	mockCall.AssertCalled(t, "ClassifyDispatch", baseWeight)
	mockCall.AssertCalled(t, "PaysFee", baseWeight)
	mockCheckedExtrinsic.AssertNotCalled(t, "Validate", unsignedValidator, txSource, &dispatchInfo, encodedExtrinsicLen)
	assert.Equal(t, defaultValidTransaction, outcome)
	assert.Equal(t, invalidTransactionMandatoryValidation, err)
}

func Test_Executive_ValidateTransaction_IsMendatory_Error(t *testing.T) {
	setup()

	dispatchInfo := primitives.DispatchInfo{
		Weight:  primitives.WeightFromParts(2, 2),
		Class:   dispatchClassMandatory,
		PaysFee: primitives.PaysYes,
	}

	mockSystemModule.On("StorageBlockNumber").Return(blockNumber, nil)
	mockSystemModule.On("Initialize", blockNumber+1, header.ParentHash, defaultDigest)
	mockUncheckedExtrinsic.On("Bytes").Return(encodedExtrinsic)
	mockUncheckedExtrinsic.On("Check").Return(mockCheckedExtrinsic, nil)
	mockCheckedExtrinsic.On("Function").Return(mockCall)
	mockCall.On("BaseWeight").Return(baseWeight)
	mockCall.On("WeighData", baseWeight).Return(dispatchInfo.Weight)
	mockCall.On("ClassifyDispatch", baseWeight).Return(primitives.DispatchClass{})
	mockCall.On("PaysFee", baseWeight).Return(dispatchInfo.PaysFee)

	outcome, err := target.ValidateTransaction(txSource, mockUncheckedExtrinsic, header.ParentHash)

	mockSystemModule.AssertCalled(t, "StorageBlockNumber")
	mockSystemModule.AssertCalled(t, "Initialize", blockNumber+1, header.ParentHash, defaultDigest)
	mockUncheckedExtrinsic.AssertCalled(t, "Bytes")
	mockUncheckedExtrinsic.AssertCalled(t, "Check")
	mockCheckedExtrinsic.AssertCalled(t, "Function")
	mockCall.AssertCalled(t, "BaseWeight")
	mockCall.AssertCalled(t, "WeighData", baseWeight)
	mockCall.AssertCalled(t, "ClassifyDispatch", baseWeight)
	mockCall.AssertCalled(t, "PaysFee", baseWeight)
	mockCheckedExtrinsic.AssertNotCalled(t, "Validate", unsignedValidator, txSource, &dispatchInfo, encodedExtrinsicLen)
	assert.Equal(t, defaultValidTransaction, outcome)
	assert.Equal(t, "not a valid 'DispatchClass' type", err.Error())
}

func Test_Executive_ValidateTransaction_StorageBlockNumber_Error(t *testing.T) {
	setup()

	mockSystemModule.On("StorageBlockNumber").Return(blockNumber, errPanic)
	// mockSystemModule.On("Initialize", blockNumber+1, header.ParentHash, defaultDigest)
	// mockUncheckedExtrinsic.On("Bytes").Return(encodedExtrinsic)
	// mockUncheckedExtrinsic.On("Check").Return(mockCheckedExtrinsic, nil)
	// mockCheckedExtrinsic.On("Function").Return(mockCall)
	// mockCall.On("BaseWeight").Return(baseWeight)
	// mockCall.On("WeighData", baseWeight).Return(dispatchInfo.Weight)
	// mockCall.On("ClassifyDispatch", baseWeight).Return(primitives.DispatchClass{})
	// mockCall.On("PaysFee", baseWeight).Return(dispatchInfo.PaysFee)

	_, err := target.ValidateTransaction(txSource, mockUncheckedExtrinsic, header.ParentHash)

	assert.Equal(t, errPanic, err)
	mockSystemModule.AssertCalled(t, "StorageBlockNumber")
}

func Test_Executive_ValidateTransaction(t *testing.T) {
	setup()

	mockSystemModule.On("StorageBlockNumber").Return(blockNumber, nil)
	mockSystemModule.On("Initialize", blockNumber+1, header.ParentHash, defaultDigest)
	mockUncheckedExtrinsic.On("Bytes").Return(encodedExtrinsic)
	mockUncheckedExtrinsic.On("Check").Return(mockCheckedExtrinsic, nil)
	mockCheckedExtrinsic.On("Function").Return(mockCall)
	mockCall.On("BaseWeight").Return(baseWeight)
	mockCall.On("WeighData", baseWeight).Return(dispatchInfo.Weight)
	mockCall.On("ClassifyDispatch", baseWeight).Return(dispatchInfo.Class)
	mockCall.On("PaysFee", baseWeight).Return(dispatchInfo.PaysFee)
	mockCheckedExtrinsic.On("Validate", unsignedValidator, txSource, &dispatchInfo, encodedExtrinsicLen).
		Return(defaultValidTransaction, nil)

	outcome, err := target.ValidateTransaction(txSource, mockUncheckedExtrinsic, header.ParentHash)

	mockSystemModule.AssertCalled(t, "StorageBlockNumber")
	mockSystemModule.AssertCalled(t, "Initialize", blockNumber+1, header.ParentHash, defaultDigest)
	mockUncheckedExtrinsic.AssertCalled(t, "Bytes")
	mockUncheckedExtrinsic.AssertCalled(t, "Check")
	mockCheckedExtrinsic.AssertCalled(t, "Function")
	mockCall.AssertCalled(t, "BaseWeight")
	mockCall.AssertCalled(t, "WeighData", baseWeight)
	mockCall.AssertCalled(t, "ClassifyDispatch", baseWeight)
	mockCall.AssertCalled(t, "PaysFee", baseWeight)
	mockCheckedExtrinsic.AssertCalled(t, "Validate", unsignedValidator, txSource, &dispatchInfo, encodedExtrinsicLen)
	assert.Equal(t, defaultValidTransaction, outcome)
	assert.Nil(t, err)
}

func Test_Executive_OffchainWorker(t *testing.T) {
	setup()

	mockSystemModule.On("Initialize", header.Number, header.ParentHash, header.Digest)
	mockIoHashing.On("Blake256", header.Bytes()).Return(blake256Hash)
	mockSystemModule.On("StorageBlockHashSet", header.Number, blockHash)
	mockRuntimeExtrinsic.On("OffchainWorker", header.Number)

	target.OffchainWorker(header)

	mockSystemModule.AssertCalled(t, "Initialize", header.Number, header.ParentHash, header.Digest)
	mockSystemModule.AssertCalled(t, "StorageBlockHashSet", header.Number, blockHash)
	mockRuntimeExtrinsic.AssertCalled(t, "OffchainWorker", header.Number)
}

func Test_Executive_OffchainWorker_NewBlake2bHash_Error(t *testing.T) {
	setup()

	mockSystemModule.On("Initialize", header.Number, header.ParentHash, header.Digest)
	mockIoHashing.On("Blake256", header.Bytes()).Return([]byte{})

	err := target.OffchainWorker(header)
	assert.Equal(t, errors.New("Blake2bHash should be of size 32"), err)

	mockSystemModule.AssertCalled(t, "Initialize", header.Number, header.ParentHash, header.Digest)
	mockIoHashing.AssertCalled(t, "Blake256", header.Bytes())
}

func Test_Executive_idleAndFinalizeHook_RegisterExtraWeightUnchecked_Error(t *testing.T) {
	setup()

	mockSystemModule.On("StorageBlockWeight").Return(consumedWeight, nil)
	mockSystemModule.On("BlockWeights").Return(blockWeights)
	mockRuntimeExtrinsic.On("OnIdle", mock.Anything, mock.Anything).Return(baseWeight)
	mockSystemModule.On("RegisterExtraWeightUnchecked", mock.Anything, mock.Anything).Return(errPanic)

	err := target.idleAndFinalizeHook(blockNumber)
	assert.Equal(t, errPanic, err)

	mockSystemModule.AssertCalled(t, "StorageBlockWeight")
	mockSystemModule.AssertCalled(t, "BlockWeights")
	mockRuntimeExtrinsic.AssertCalled(t, "OnIdle", mock.Anything, mock.Anything)
	mockSystemModule.AssertCalled(t, "RegisterExtraWeightUnchecked", mock.Anything, mock.Anything)
}

func Test_Executive_idleAndFinalizeHook_OnFinalize_Error(t *testing.T) {
	setup()

	// blockNumber := sc.U64(1)
	mockSystemModule.On("StorageBlockWeight").Return(consumedWeight, nil)
	mockSystemModule.On("BlockWeights").Return(blockWeights)
	mockRuntimeExtrinsic.On("OnIdle", mock.Anything, mock.Anything).Return(baseWeight)
	mockSystemModule.On("RegisterExtraWeightUnchecked", mock.Anything, mock.Anything).Return(nil)
	mockRuntimeExtrinsic.On("OnFinalize", blockNumber).Return(errPanic)

	err := target.idleAndFinalizeHook(blockNumber)

	assert.Equal(t, errPanic, err)
	mockSystemModule.AssertCalled(t, "StorageBlockWeight")
	mockSystemModule.AssertCalled(t, "BlockWeights")
	mockRuntimeExtrinsic.AssertCalled(t, "OnIdle", mock.Anything, mock.Anything)
	mockSystemModule.AssertCalled(t, "RegisterExtraWeightUnchecked", mock.Anything, mock.Anything)
	mockRuntimeExtrinsic.AssertCalled(t, "OnFinalize", blockNumber)
}

func Test_executeExtrinsicsWithBookKeeping_ApplyExtrinsic_TransactionValidityError(t *testing.T) {
	invalidTransactionBadProofError := primitives.NewTransactionValidityError(primitives.NewInvalidTransactionBadProof())

	setup()

	block := types.NewBlock(header, sc.Sequence[primitives.UncheckedExtrinsic]{mockUncheckedExtrinsic})

	mockUncheckedExtrinsic.On("Bytes").Return(encodedExtrinsic)
	mockUncheckedExtrinsic.On("Check").Return(nil, invalidTransactionBadProofError)

	assert.PanicsWithValue(t, invalidTransactionBadProofError.Error(), func() {
		target.executeExtrinsicsWithBookKeeping(block)
	})

	mockUncheckedExtrinsic.AssertCalled(t, "Bytes")
	mockUncheckedExtrinsic.AssertCalled(t, "Check")
	mockSystemModule.AssertNotCalled(t, "NoteFinishedExtrinsics")
}

func Test_executeExtrinsicsWithBookKeepingNoteFinishedExtrinsics_Error(t *testing.T) {
	setup()
	postInfo := primitives.PostDispatchInfo{}
	block := types.NewBlock(header, sc.Sequence[primitives.UncheckedExtrinsic]{mockUncheckedExtrinsic})
	expectedErr := unknownTransactionCannotLookupError

	mockUncheckedExtrinsic.On("Bytes").Return(encodedExtrinsic)
	mockUncheckedExtrinsic.On("Check").Return(mockCheckedExtrinsic, nil)
	mockSystemModule.On("NoteExtrinsic", mockUncheckedExtrinsic.Bytes())

	mockCheckedExtrinsic.On("Function").Return(mockCall)
	mockCall.On("BaseWeight").Return(baseWeight)
	mockCall.On("WeighData", baseWeight).Return(dispatchInfo.Weight)
	mockCall.On("ClassifyDispatch", baseWeight).Return(dispatchInfo.Class)
	mockCall.On("PaysFee", baseWeight).Return(dispatchInfo.PaysFee)
	mockCheckedExtrinsic.On("Apply", unsignedValidator, &dispatchInfo, encodedExtrinsicLen).
		Return(postInfo, nil)
	mockSystemModule.On("NoteAppliedExtrinsic", postInfo, nil, dispatchInfo).Return(nil)
	mockSystemModule.On("NoteFinishedExtrinsics").Return(expectedErr)

	err := target.executeExtrinsicsWithBookKeeping(block)

	assert.Equal(t, expectedErr, err)
	mockUncheckedExtrinsic.AssertCalled(t, "Bytes")
	mockUncheckedExtrinsic.AssertCalled(t, "Check")
	mockSystemModule.AssertCalled(t, "NoteExtrinsic", mockUncheckedExtrinsic.Bytes())
	mockSystemModule.AssertCalled(t, "NoteAppliedExtrinsic", postInfo, nil, dispatchInfo)
	mockSystemModule.AssertCalled(t, "NoteFinishedExtrinsics")
}

func Test_Executive_finalChecks_Finalize_Error(t *testing.T) {
	setup()

	mockSystemModule.On("Finalize").Return(header, errPanic)

	err := target.finalChecks(&primitives.Header{})
	assert.Equal(t, errPanic, err)

	mockSystemModule.AssertCalled(t, "Finalize")
}

func Test_Executive_finalChecks_ErrorInvalidDigestNum(t *testing.T) {
	setup()

	mockSystemModule.On("Finalize").Return(header, nil)

	err := target.finalChecks(&primitives.Header{})
	assert.Equal(t, errInvalidDigestNum, err)

	mockSystemModule.AssertCalled(t, "Finalize")
}

func Test_Executive_finalChecks_ErrorInvalidDigestItem(t *testing.T) {
	setup()

	newHeader := primitives.Header{
		Number:     blockNumber,
		ParentHash: blockHash,
		Digest: primitives.NewDigest(sc.Sequence[primitives.DigestItem]{
			primitives.NewDigestItemPreRuntime(
				sc.BytesToFixedSequenceU8([]byte{'a', 'u', 'r', 'b'}),
				sc.BytesToSequenceU8(sc.U64(0).Bytes()),
			),
		}),
	}
	mockSystemModule.On("Finalize").Return(newHeader, nil)

	err := target.finalChecks(&header)
	assert.Equal(t, errInvalidDigestItem, err)

	mockSystemModule.AssertCalled(t, "Finalize")
}

func Test_Executive_finalChecks_ErrorInvalidStorageRoot(t *testing.T) {
	setup()

	newHeader := primitives.Header{
		Number:     blockNumber,
		ParentHash: blockHash,
		Digest:     testDigest(),
		StateRoot:  primitives.H256{FixedSequence: sc.NewFixedSequence[sc.U8](1, sc.U8(2))},
	}
	mockSystemModule.On("Finalize").Return(newHeader, nil)

	err := target.finalChecks(&header)
	assert.Equal(t, errInvalidStorageRoot, err)

	mockSystemModule.AssertCalled(t, "Finalize")
}

func Test_Executive_finalChecks_ErrorInvalidTxTrie(t *testing.T) {
	setup()

	newHeader := primitives.Header{
		Number:         blockNumber,
		ParentHash:     blockHash,
		Digest:         testDigest(),
		ExtrinsicsRoot: primitives.H256{FixedSequence: sc.NewFixedSequence[sc.U8](1, sc.U8(2))},
	}
	mockSystemModule.On("Finalize").Return(newHeader, nil)

	err := target.finalChecks(&header)
	assert.Equal(t, errInvalidTxTrie, err)

	mockSystemModule.AssertCalled(t, "Finalize")
}
