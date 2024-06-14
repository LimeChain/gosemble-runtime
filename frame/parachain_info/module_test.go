package parachain_info

import (
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants/metadata"
	"github.com/LimeChain/gosemble/mocks"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
)

const (
	moduleId sc.U8 = 7
)

var (
	unknownTransactionNoUnsignedValidator = primitives.NewTransactionValidityError(primitives.NewUnknownTransactionNoUnsignedValidator())
)

var (
	mockStorage            *mocks.IoStorage
	mockStorageParachainId *mocks.StorageValue[sc.U32]
)

func Test_Module_GetIndex(t *testing.T) {
	target := setup()

	assert.Equal(t, moduleId, target.GetIndex())
}

func Test_Module_Functions(t *testing.T) {
	target := setup()

	assert.Equal(t, map[sc.U8]primitives.Call{}, target.Functions())
}

func Test_Module_PreDispatch(t *testing.T) {
	target := setup()

	result, err := target.PreDispatch(new(mocks.Call))

	assert.Nil(t, err)
	assert.Equal(t, sc.Empty{}, result)
}

func Test_Module_ValidateUnsigned(t *testing.T) {
	target := setup()

	result, err := target.ValidateUnsigned(primitives.NewTransactionSourceLocal(), new(mocks.Call))

	assert.Equal(t, unknownTransactionNoUnsignedValidator, err)
	assert.Equal(t, primitives.ValidTransaction{}, result)
}

func Test_Module_Metadata(t *testing.T) {
	target := setup()

	moduleV14 := primitives.MetadataModuleV14{
		Name: name,
		Storage: sc.NewOption[primitives.MetadataModuleStorage](primitives.MetadataModuleStorage{
			Prefix: name,
			Items: sc.Sequence[primitives.MetadataModuleStorageEntry]{
				primitives.NewMetadataModuleStorageEntry(
					"ParachainId",
					primitives.MetadataModuleStorageEntryModifierDefault,
					primitives.NewMetadataModuleStorageEntryDefinitionPlain(sc.ToCompact(metadata.PrimitiveTypesU32)),
					"The id of the parachain",
				),
			},
		}),
		Call:      sc.NewOption[sc.Compact](nil),
		CallDef:   sc.NewOption[primitives.MetadataDefinitionVariant](nil),
		Event:     sc.NewOption[sc.Compact](nil),
		EventDef:  sc.NewOption[primitives.MetadataDefinitionVariant](nil),
		Constants: sc.Sequence[primitives.MetadataModuleConstant]{},
		Error:     sc.NewOption[sc.Compact](nil),
		ErrorDef:  sc.NewOption[primitives.MetadataDefinitionVariant](nil),
		Index:     moduleId,
	}

	expectMetadata := primitives.MetadataModule{
		ModuleV14: moduleV14,
		Version:   primitives.ModuleVersion14,
	}

	result := target.Metadata()

	assert.Equal(t, result, expectMetadata)
}

func setup() Module {
	mockStorage = new(mocks.IoStorage)
	mockStorageParachainId = new(mocks.StorageValue[sc.U32])

	target := New(moduleId, mockStorage)
	target.storage.ParachainId = mockStorageParachainId

	return target
}
