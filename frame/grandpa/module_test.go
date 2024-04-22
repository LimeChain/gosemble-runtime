package grandpa

import (
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants"
	"github.com/LimeChain/gosemble/constants/metadata"
	"github.com/LimeChain/gosemble/hooks"
	"github.com/LimeChain/gosemble/mocks"
	"github.com/LimeChain/gosemble/primitives/log"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
)

const moduleId = sc.U8(3)

var (
	mdGenerator = primitives.NewMetadataTypeGenerator()
)

var (
	unknownTransactionNoUnsignedValidator = primitives.NewTransactionValidityError(primitives.NewUnknownTransactionNoUnsignedValidator())
)

var (
	mockStorageAuthorities *mocks.StorageValue[sc.Sequence[primitives.Authority]]
	target                 Module
	logger                 = log.NewLogger()
)

func Test_Module_New(t *testing.T) {
	setup()

	assert.Equal(t, Module{
		DefaultInherentProvider: primitives.DefaultInherentProvider{},
		DefaultDispatchModule:   hooks.DefaultDispatchModule{},
		Index:                   moduleId,
		storage: &storage{
			mockStorageAuthorities,
		},
		logger:      logger,
		mdGenerator: mdGenerator,
	}, target)
}

func Test_Module_KeyType(t *testing.T) {
	setup()
	assert.Equal(t, primitives.PublicKeyEd25519, target.KeyType())
}

func Test_Module_KeyTypeId(t *testing.T) {
	setup()
	assert.Equal(t, KeyTypeId, target.KeyTypeId())
}

func Test_Module_GetIndex(t *testing.T) {
	setup()
	assert.Equal(t, moduleId, target.GetIndex())
}

func Test_Module_Functions(t *testing.T) {
	setup()

	assert.Equal(t, map[sc.U8]primitives.Call{}, target.Functions())
}

func Test_Module_PreDispatch(t *testing.T) {
	setup()

	result, err := target.PreDispatch(new(mocks.Call))

	assert.Nil(t, err)
	assert.Equal(t, sc.Empty{}, result)
}

func Test_Module_ValidateUnsigned(t *testing.T) {
	setup()

	result, err := target.ValidateUnsigned(primitives.NewTransactionSourceLocal(), new(mocks.Call))

	assert.Equal(t, unknownTransactionNoUnsignedValidator, err)
	assert.Equal(t, primitives.ValidTransaction{}, result)
}

func Test_Module_Authorities_Success(t *testing.T) {
	setup()
	expectAuthorites := sc.Sequence[primitives.Authority]{
		{
			Id:     constants.ZeroAccountId,
			Weight: 5,
		},
	}

	mockStorageAuthorities.On("Get").Return(expectAuthorites, nil)

	result, err := target.Authorities()
	assert.Nil(t, err)

	assert.Equal(t, expectAuthorites, result)
	mockStorageAuthorities.AssertCalled(t, "Get")
}

func Test_Module_Metadata(t *testing.T) {
	setup()

	expectedGrandpaCallsMetadataId := mdGenerator.GetLastAvailableIndex() + 1

	expectMetadataTypes := sc.Sequence[primitives.MetadataType]{
		primitives.NewMetadataTypeWithParams(expectedGrandpaCallsMetadataId, "Grandpa calls", sc.Sequence[sc.Str]{"pallet_grandpa", "pallet", "Call"}, primitives.NewMetadataTypeDefinitionVariant(
			sc.Sequence[primitives.MetadataDefinitionVariant]{}),
			sc.Sequence[primitives.MetadataTypeParameter]{
				primitives.NewMetadataEmptyTypeParameter("T"),
				primitives.NewMetadataEmptyTypeParameter("I"),
			}),
		primitives.NewMetadataTypeWithParams(metadata.TypesGrandpaErrors, "The `Error` enum of this pallet.", sc.Sequence[sc.Str]{"pallet_grandpa", "pallet", "Error"}, primitives.NewMetadataTypeDefinitionVariant(
			sc.Sequence[primitives.MetadataDefinitionVariant]{
				primitives.NewMetadataDefinitionVariant("PauseFailed", sc.Sequence[primitives.MetadataTypeDefinitionField]{}, PauseFailedError, ""),
				primitives.NewMetadataDefinitionVariant("ResumeFailed", sc.Sequence[primitives.MetadataTypeDefinitionField]{}, ResumeFailedError, ""),
				primitives.NewMetadataDefinitionVariant("ChangePending", sc.Sequence[primitives.MetadataTypeDefinitionField]{}, ChangePendingError, ""),
				primitives.NewMetadataDefinitionVariant("TooSoon", sc.Sequence[primitives.MetadataTypeDefinitionField]{}, TooSoonError, ""),
				primitives.NewMetadataDefinitionVariant("InvalidKeyOwnershipProof", sc.Sequence[primitives.MetadataTypeDefinitionField]{}, InvalidKeyOwnershipProofError, ""),
				primitives.NewMetadataDefinitionVariant("InvalidEquivocationProof", sc.Sequence[primitives.MetadataTypeDefinitionField]{}, InvalidEquivocationProofError, ""),
				primitives.NewMetadataDefinitionVariant("DuplicateOffenceReport", sc.Sequence[primitives.MetadataTypeDefinitionField]{}, DuplicateOffenceReportError, ""),
			}),
			sc.Sequence[primitives.MetadataTypeParameter]{
				primitives.NewMetadataEmptyTypeParameter("T"),
			}),
		primitives.NewMetadataTypeWithPath(metadata.TypesGrandpaAppPublic, "sp_consensus_grandpa app Public", sc.Sequence[sc.Str]{"sp_consensus_grandpa", "app", "Public"}, primitives.NewMetadataTypeDefinitionComposite(
			sc.Sequence[primitives.MetadataTypeDefinitionField]{
				primitives.NewMetadataTypeDefinitionField(metadata.TypesEd25519PubKey),
			})),
		primitives.NewMetadataType(metadata.TypesTupleGrandpaAppPublicU64, "(GrandpaAppPublic, U64)",
			primitives.NewMetadataTypeDefinitionTuple(sc.Sequence[sc.Compact]{sc.ToCompact(metadata.TypesGrandpaAppPublic), sc.ToCompact(metadata.PrimitiveTypesU64)})),
		primitives.NewMetadataType(metadata.TypesSequenceTupleGrandpaAppPublic, "[]byte (GrandpaAppPublic, U64)", primitives.NewMetadataTypeDefinitionSequence(sc.ToCompact(metadata.TypesTupleGrandpaAppPublicU64))),
	}
	moduleV14 := primitives.MetadataModuleV14{
		Name:      name,
		Storage:   sc.Option[primitives.MetadataModuleStorage]{},
		Call:      sc.NewOption[sc.Compact](nil),
		CallDef:   sc.NewOption[primitives.MetadataDefinitionVariant](nil),
		Event:     sc.NewOption[sc.Compact](nil),
		EventDef:  sc.NewOption[primitives.MetadataDefinitionVariant](nil),
		Constants: sc.Sequence[primitives.MetadataModuleConstant]{},
		Error:     sc.NewOption[sc.Compact](nil),
		ErrorDef: sc.NewOption[primitives.MetadataDefinitionVariant](
			primitives.NewMetadataDefinitionVariantStr(
				name,
				sc.Sequence[primitives.MetadataTypeDefinitionField]{
					primitives.NewMetadataTypeDefinitionField(metadata.TypesGrandpaErrors),
				},
				moduleId,
				"Errors.Grandpa"),
		),
		Index: moduleId,
	}

	expectMetadataModule := primitives.MetadataModule{
		Version:   primitives.ModuleVersion14,
		ModuleV14: moduleV14,
	}

	metadataModule := target.Metadata()
	metadataTypes := mdGenerator.GetMetadataTypes()

	assert.Equal(t, expectMetadataTypes, metadataTypes)
	assert.Equal(t, expectMetadataModule, metadataModule)
}

func setup() {
	mockStorageAuthorities = new(mocks.StorageValue[sc.Sequence[primitives.Authority]])
	target = New(moduleId, logger, mdGenerator)

	target.storage.Authorities = mockStorageAuthorities
}
