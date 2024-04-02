package aura

import (
	"errors"
	"github.com/LimeChain/gosemble/primitives/log"
	"io"
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants/metadata"
	"github.com/LimeChain/gosemble/mocks"
	"github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	moduleId                          = 13
	weightRefTimePerNanos      sc.U64 = 1_000
	timestampMinimumPeriod            = 2_000
	maxAuthorities                    = 10
	allowMultipleBlocksPerSlot        = false
	keyType                           = types.PublicKeySr25519
	blockNumber                sc.U64 = 0
	currentSlot                       = sc.U32(4)
)

var (
	digestsPreRuntime = sc.Sequence[types.DigestPreRuntime]{
		{
			ConsensusEngineId: sc.BytesToFixedSequenceU8(EngineId[:]),
			Message:           sc.BytesToSequenceU8(sc.U64(currentSlot).Bytes()),
		},
	}
	mdGenerator = types.NewMetadataTypeGenerator()
)

var (
	unknownTransactionNoUnsignedValidator = types.NewTransactionValidityError(types.NewUnknownTransactionNoUnsignedValidator())
)

var (
	dbWeight = types.RuntimeDbWeight{
		Read:  3 * weightRefTimePerNanos,
		Write: 7 * weightRefTimePerNanos,
	}
	module                 Module
	mockLogDepositor       *mocks.LogDepositor
	mockStorageDigest      *mocks.StorageValue[types.Digest]
	mockStorageCurrentSlot *mocks.StorageValue[sc.U64]
	mockStorageAuthorities *mocks.StorageValue[sc.Sequence[types.Sr25519PublicKey]]
)

var (
	expectedMetadataTypes = sc.Sequence[types.MetadataType]{
		types.NewMetadataTypeWithParams(
			metadata.TypesAuraStorageAuthorities,
			"BoundedVec<T::AuthorityId, T::MaxAuthorities>",
			sc.Sequence[sc.Str]{"bounded_collection", "bounded_vec", "BoundedVec"},
			types.NewMetadataTypeDefinitionComposite(
				sc.Sequence[types.MetadataTypeDefinitionField]{
					types.NewMetadataTypeDefinitionField(metadata.TypesSequencePubKeys),
				}), sc.Sequence[types.MetadataTypeParameter]{
				types.NewMetadataTypeParameter(metadata.TypesAuthorityId, "T"),
				types.NewMetadataEmptyTypeParameter("S"),
			}),

		types.NewMetadataTypeWithPath(metadata.TypesAuthorityId,
			"sp_consensus_aura sr25519 app_sr25519 Public",
			sc.Sequence[sc.Str]{"sp_consensus_aura", "sr25519", "app_sr25519", "Public"},
			types.NewMetadataTypeDefinitionComposite(
				sc.Sequence[types.MetadataTypeDefinitionField]{types.NewMetadataTypeDefinitionField(metadata.TypesSr25519PubKey)})),

		types.NewMetadataTypeWithPath(metadata.TypesSr25519PubKey,
			"sp_core sr25519 Public",
			sc.Sequence[sc.Str]{"sp_core", "sr25519", "Public"},
			types.NewMetadataTypeDefinitionComposite(
				sc.Sequence[types.MetadataTypeDefinitionField]{types.NewMetadataTypeDefinitionField(metadata.TypesFixedSequence32U8)})),

		types.NewMetadataType(metadata.TypesSequencePubKeys,
			"[]PublicKey",
			types.NewMetadataTypeDefinitionSequence(sc.ToCompact(metadata.TypesAuthorityId))),

		types.NewMetadataTypeWithPath(metadata.TypesAuraSlot,
			"sp_consensus_slots Slot",
			sc.Sequence[sc.Str]{"sp_consensus_slots", "Slot"},
			types.NewMetadataTypeDefinitionComposite(
				sc.Sequence[types.MetadataTypeDefinitionField]{
					types.NewMetadataTypeDefinitionField(metadata.PrimitiveTypesU64),
				})),

		// type 924
		types.NewMetadataType(metadata.TypesTupleSequenceU8KeyTypeId, "(Seq<U8>, KeyTypeId)",
			types.NewMetadataTypeDefinitionTuple(sc.Sequence[sc.Compact]{sc.ToCompact(metadata.TypesSequenceU8), sc.ToCompact(metadata.TypesKeyTypeId)})),

		// type 923
		types.NewMetadataType(metadata.TypesSequenceTupleSequenceU8KeyTypeId, "[]byte TupleSequenceU8KeyTypeId", types.NewMetadataTypeDefinitionSequence(sc.ToCompact(metadata.TypesTupleSequenceU8KeyTypeId))),

		// type 922
		types.NewMetadataTypeWithParam(metadata.TypesOptionTupleSequenceU8KeyTypeId, "Option<TupleSequenceU8KeyTypeId>", sc.Sequence[sc.Str]{"Option"}, types.NewMetadataTypeDefinitionVariant(
			sc.Sequence[types.MetadataDefinitionVariant]{
				types.NewMetadataDefinitionVariant(
					"None",
					sc.Sequence[types.MetadataTypeDefinitionField]{},
					0,
					""),
				types.NewMetadataDefinitionVariant(
					"Some",
					sc.Sequence[types.MetadataTypeDefinitionField]{
						types.NewMetadataTypeDefinitionField(metadata.TypesSequenceTupleSequenceU8KeyTypeId),
					},
					1,
					""),
			}),
			types.NewMetadataTypeParameter(metadata.TypesSequenceTupleSequenceU8KeyTypeId, "T")),
	}

	moduleV14 = types.MetadataModuleV14{
		Name: "Aura",
		Storage: sc.NewOption[types.MetadataModuleStorage](types.MetadataModuleStorage{
			Prefix: "Aura",
			Items: sc.Sequence[types.MetadataModuleStorageEntry]{
				types.NewMetadataModuleStorageEntry(
					"Authorities",
					types.MetadataModuleStorageEntryModifierDefault,
					types.NewMetadataModuleStorageEntryDefinitionPlain(sc.ToCompact(metadata.TypesAuraStorageAuthorities)),
					"The current authority set."),
				types.NewMetadataModuleStorageEntry(
					"CurrentSlot",
					types.MetadataModuleStorageEntryModifierDefault,
					types.NewMetadataModuleStorageEntryDefinitionPlain(sc.ToCompact(metadata.TypesAuraSlot)),
					"The current slot of this block.   This will be set in `on_initialize`."),
			},
		}),
		Call:      sc.NewOption[sc.Compact](nil),
		CallDef:   sc.NewOption[types.MetadataDefinitionVariant](nil),
		Event:     sc.NewOption[sc.Compact](nil),
		EventDef:  sc.NewOption[types.MetadataDefinitionVariant](nil),
		Constants: sc.Sequence[types.MetadataModuleConstant]{},
		Error:     sc.NewOption[sc.Compact](nil),
		ErrorDef:  sc.NewOption[types.MetadataDefinitionVariant](nil),
		Index:     sc.U8(13),
	}

	expectedMetadataModule = types.MetadataModule{
		Version:   types.ModuleVersion14,
		ModuleV14: moduleV14,
	}
)

func setup(minimumPeriod sc.U64) {
	mockLogDepositor = new(mocks.LogDepositor)
	mockStorageDigest = new(mocks.StorageValue[types.Digest])
	mockStorageCurrentSlot = new(mocks.StorageValue[sc.U64])
	mockStorageAuthorities = new(mocks.StorageValue[sc.Sequence[types.Sr25519PublicKey]])

	config := NewConfig(
		keyType,
		dbWeight,
		minimumPeriod,
		maxAuthorities,
		allowMultipleBlocksPerSlot,
		mockStorageDigest.Get,
		mockLogDepositor,
		nil,
	)
	module = New(moduleId, config, mdGenerator, log.NewLogger())
	module.storage.CurrentSlot = mockStorageCurrentSlot
	module.storage.Authorities = mockStorageAuthorities
}

func newPreRuntimeDigest(n sc.U64) types.Digest {
	items := sc.Sequence[types.DigestItem]{
		types.NewDigestItemPreRuntime(
			sc.BytesToFixedSequenceU8(EngineId[:]),
			sc.BytesToSequenceU8(n.Bytes()),
		),
	}
	return types.NewDigest(items)
}

func invalidPreRuntimeDigest() types.Digest {
	items := sc.Sequence[types.DigestItem]{
		types.NewDigestItemPreRuntime(
			sc.BytesToFixedSequenceU8(EngineId[:]),
			sc.BytesToSequenceU8(sc.U8(1).Bytes()),
		),
	}
	return types.NewDigest(items)
}

func Test_Aura_GetIndex(t *testing.T) {
	setup(timestampMinimumPeriod)

	assert.Equal(t, sc.U8(13), module.GetIndex())
}

func Test_Aura_Functions(t *testing.T) {
	setup(timestampMinimumPeriod)

	assert.Equal(t, map[sc.U8]types.Call{}, module.Functions())
}

func Test_Module_PreDispatch(t *testing.T) {
	setup(timestampMinimumPeriod)

	result, err := module.PreDispatch(new(mocks.Call))

	assert.Nil(t, err)
	assert.Equal(t, sc.Empty{}, result)
}

func Test_Module_ValidateUnsigned(t *testing.T) {
	setup(timestampMinimumPeriod)

	result, err := module.ValidateUnsigned(types.NewTransactionSourceLocal(), new(mocks.Call))

	assert.Equal(t, unknownTransactionNoUnsignedValidator, err)
	assert.Equal(t, types.ValidTransaction{}, result)
}

func Test_Aura_KeyType(t *testing.T) {
	setup(timestampMinimumPeriod)

	assert.Equal(t, keyType, module.KeyType())
}

func Test_Aura_KeyTypeId(t *testing.T) {
	setup(timestampMinimumPeriod)

	assert.Equal(t, [4]byte{'a', 'u', 'r', 'a'}, module.KeyTypeId())
}

func Test_Aura_Metadata(t *testing.T) {
	setup(timestampMinimumPeriod)

	metadataModule := module.Metadata()
	metadataTypes := mdGenerator.GetMetadataTypes()

	assert.Equal(t, expectedMetadataTypes, metadataTypes)
	assert.Equal(t, expectedMetadataModule, metadataModule)
}

func Test_Aura_OnInitialize_EmptySlot(t *testing.T) {
	setup(timestampMinimumPeriod)
	mockStorageDigest.On("Get").Return(types.Digest{}, nil)

	onInit, err := module.OnInitialize(blockNumber)
	assert.Nil(t, err)

	assert.Equal(t, types.WeightFromParts(3000, 0), onInit)
	mockStorageDigest.AssertCalled(t, "Get")
	mockStorageCurrentSlot.AssertNotCalled(t, "Put", mock.Anything)
}

func Test_Aura_OnInitialize_CurrentSlotMustIncrease(t *testing.T) {
	setup(timestampMinimumPeriod)
	mockStorageDigest.On("Get").Return(newPreRuntimeDigest(sc.U64(1)), nil)
	mockStorageCurrentSlot.On("Get").Return(sc.U64(2), nil)

	_, err := module.OnInitialize(blockNumber)
	assert.Equal(t, errSlotMustIncrease, err)

	mockStorageDigest.AssertCalled(t, "Get")
	mockStorageCurrentSlot.AssertNotCalled(t, "Put", mock.Anything)
}

func Test_Aura_OnInitialize_Fails_StorageDigestNotFound(t *testing.T) {
	setup(timestampMinimumPeriod)
	expectError := errors.New("not found")

	mockStorageDigest.On("Get").Return(types.Digest{}, expectError)

	result, err := module.OnInitialize(blockNumber)

	assert.Equal(t, expectError, err)
	assert.Equal(t, types.Weight{}, result)

	mockStorageDigest.AssertCalled(t, "Get")
	mockStorageCurrentSlot.AssertNotCalled(t, "Get")
}

func Test_Aura_OnInitialize_Fails_CannotDecodeDigestPreRuntimeMessage(t *testing.T) {
	setup(timestampMinimumPeriod)
	expectError := errors.New("can not read the required number of bytes 8, only 1 available")

	mockStorageDigest.On("Get").Return(invalidPreRuntimeDigest(), nil)
	mockStorageCurrentSlot.On("Get").Return(sc.U64(2), nil)

	result, err := module.OnInitialize(blockNumber)

	assert.Equal(t, expectError, err)
	assert.Equal(t, types.Weight{}, result)

	mockStorageDigest.AssertCalled(t, "Get")
	mockStorageCurrentSlot.AssertNotCalled(t, "Get")
}

func Test_Aura_OnInitialize_CurrentSlotUpdate(t *testing.T) {
	setup(timestampMinimumPeriod)
	mockStorageDigest.On("Get").Return(newPreRuntimeDigest(sc.U64(1)), nil)
	mockStorageCurrentSlot.On("Get").Return(sc.U64(0), nil)
	mockStorageCurrentSlot.On("Put", sc.U64(1)).Return()
	mockStorageAuthorities.On("DecodeLen").Return(sc.NewOption[sc.U64](sc.U64(3)), nil)

	onInit, err := module.OnInitialize(blockNumber)
	assert.Nil(t, err)

	assert.Equal(t, types.WeightFromParts(13_000, 0), onInit)
	mockStorageDigest.AssertCalled(t, "Get")
	mockStorageCurrentSlot.AssertCalled(t, "Put", sc.U64(1))
}

func Test_Aura_OnTimestampSet_DurationCannotBeZero(t *testing.T) {
	setup(0)
	mockStorageCurrentSlot.On("Get").Return(0, nil)

	err := module.OnTimestampSet(1)
	assert.Equal(t, errSlotDurationZero, err)
}

func Test_Aura_OnTimestampSet_TimestampSlotMismatch(t *testing.T) {
	setup(timestampMinimumPeriod)
	mockStorageCurrentSlot.On("Get").Return(sc.U64(2), nil)

	err := module.OnTimestampSet(sc.U64(4_000))
	assert.Equal(t, errTimestampSlotMismatch, err)

	mockStorageCurrentSlot.AssertCalled(t, "Get")
}

func Test_Aura_FindAuthor(t *testing.T) {
	setup(timestampMinimumPeriod)

	mockStorageAuthorities.On("DecodeLen").Return(sc.NewOption[sc.U64](sc.U64(1)), nil)

	result, err := module.FindAuthor(digestsPreRuntime)

	assert.Nil(t, err)
	assert.Equal(t, sc.NewOption[sc.U32](sc.U32(0)), result)
	mockStorageAuthorities.AssertCalled(t, "DecodeLen")
}

func Test_Aura_FindAuthor_Empty(t *testing.T) {
	preRuntimes := sc.Sequence[types.DigestPreRuntime]{}
	setup(timestampMinimumPeriod)

	result, err := module.FindAuthor(preRuntimes)

	assert.Nil(t, err)
	assert.Equal(t, sc.NewOption[sc.U32](nil), result)
	mockStorageAuthorities.AssertNotCalled(t, "DecodeLen")
}

func Test_Aura_FindAuthor_InvalidMessage(t *testing.T) {
	preRuntimes := sc.Sequence[types.DigestPreRuntime]{
		{
			ConsensusEngineId: sc.BytesToFixedSequenceU8(EngineId[:]),
			Message:           sc.Sequence[sc.U8]{}, // empty sequence
		},
	}
	setup(timestampMinimumPeriod)

	result, err := module.FindAuthor(preRuntimes)

	assert.Equal(t, io.EOF, err)
	assert.Equal(t, sc.Option[sc.U32]{}, result)
	mockStorageAuthorities.AssertNotCalled(t, "DecodeLen")
}

func Test_Aura_FindAuthor_ErrDecodeLen(t *testing.T) {
	setup(timestampMinimumPeriod)
	expectError := errors.New("expect")

	mockStorageAuthorities.On("DecodeLen").Return(sc.NewOption[sc.U64](nil), expectError)

	result, err := module.FindAuthor(digestsPreRuntime)

	assert.Equal(t, expectError, err)
	assert.Equal(t, sc.Option[sc.U32]{}, result)
	mockStorageAuthorities.AssertCalled(t, "DecodeLen")
}

func Test_Aura_FindAuthor_ErrZeroAuthorities(t *testing.T) {
	setup(timestampMinimumPeriod)

	mockStorageAuthorities.On("DecodeLen").Return(sc.NewOption[sc.U64](sc.U64(0)), nil)

	result, err := module.FindAuthor(digestsPreRuntime)

	assert.Equal(t, errZeroAuthorities, err)
	assert.Equal(t, sc.Option[sc.U32]{}, result)
	mockStorageAuthorities.AssertCalled(t, "DecodeLen")
}

func Test_Aura_FindAuthor_ErrEmptyAuthorities(t *testing.T) {
	setup(timestampMinimumPeriod)

	mockStorageAuthorities.On("DecodeLen").Return(sc.NewOption[sc.U64](nil), nil)

	result, err := module.FindAuthor(digestsPreRuntime)

	assert.Equal(t, errEmptyAuthorities, err)
	assert.Equal(t, sc.Option[sc.U32]{}, result)
	mockStorageAuthorities.AssertCalled(t, "DecodeLen")
}

func Test_Aura_SlotDuration(t *testing.T) {
	setup(timestampMinimumPeriod)

	assert.Equal(t, sc.U64(4_000), module.SlotDuration())
}

func Test_Aura_StorageAuthoritiesBytes(t *testing.T) {
	bytesAuthorities := sc.Option[sc.Sequence[sc.U8]]{HasValue: true, Value: sc.BytesToSequenceU8([]byte{1, 2, 3})}
	setup(timestampMinimumPeriod)

	mockStorageAuthorities.On("GetBytes").Return(bytesAuthorities, nil)

	result, err := module.StorageAuthoritiesBytes()

	assert.Nil(t, err)
	assert.Equal(t, bytesAuthorities, result)
	mockStorageAuthorities.AssertCalled(t, "GetBytes")
}

func Test_Aura_StorageAuthorities(t *testing.T) {
	setup(timestampMinimumPeriod)

	mockStorageAuthorities.On("Get").Return(authorities, nil)

	result, err := module.StorageAuthorities()

	assert.Nil(t, err)
	assert.Equal(t, authorities, result)
	mockStorageAuthorities.AssertCalled(t, "Get")
}

func Test_Aura_StorageCurrentSlot(t *testing.T) {
	slot := sc.U64(5)
	setup(timestampMinimumPeriod)

	mockStorageCurrentSlot.On("Get").Return(slot, nil)

	result, err := module.StorageCurrentSlot()

	assert.Nil(t, err)
	assert.Equal(t, slot, result)
	mockStorageCurrentSlot.AssertCalled(t, "Get")
}
