package aura_ext

import (
	"errors"
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants/metadata"
	"github.com/LimeChain/gosemble/mocks"
	"github.com/LimeChain/gosemble/primitives/log"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
)

var (
	moduleId sc.U8 = 3
	dbWeight       = primitives.RuntimeDbWeight{
		Read:  1,
		Write: 2,
	}

	currentSlot = sc.U64(5)
	slotInfo    = SlotInfo{
		Slot:     currentSlot,
		Authored: 1,
	}
	authoredSlotInfo = SlotInfo{
		Slot:     slotInfo.Slot,
		Authored: slotInfo.Authored + 1,
	}
	optionSlotInfo = sc.NewOption[sc.Sequence[sc.U8]](sc.BytesToSequenceU8(slotInfo.Bytes()))

	logger = log.NewLogger()
)

var (
	unknownTransactionNoUnsignedValidator = primitives.NewTransactionValidityError(primitives.NewUnknownTransactionNoUnsignedValidator())
)

var (
	mockStorage     *mocks.IoStorage
	mockAuthorities *mocks.StorageValue[sc.Sequence[primitives.Sr25519PublicKey]]
	mockSlotInfo    *mocks.StorageValue[SlotInfo]
	mockAuraModule  *mocks.AuraModule
)

func Test_Module_GetIndex(t *testing.T) {
	assert.Equal(t, sc.U8(moduleId), setupModule().GetIndex())
}

func Test_Module_Functions(t *testing.T) {
	target := setupModule()
	functions := target.Functions()

	assert.Equal(t, 0, len(functions))
}

func Test_Module_PreDispatch(t *testing.T) {
	target := setupModule()

	result, err := target.PreDispatch(nil)

	assert.Nil(t, err)
	assert.Equal(t, sc.Empty{}, result)
}

func Test_Module_ValidateUnsigned(t *testing.T) {
	target := setupModule()

	result, err := target.ValidateUnsigned(primitives.TransactionSource{}, nil)

	assert.Equal(t, unknownTransactionNoUnsignedValidator, err)
	assert.Equal(t, primitives.ValidTransaction{}, result)
}

func Test_Module_OnInitialize(t *testing.T) {
	target := setupModule()

	mockAuthorities.On("Get").Return(sc.Sequence[primitives.Sr25519PublicKey]{}, nil)
	mockAuraModule.On("StorageCurrentSlot").Return(currentSlot, nil)
	mockSlotInfo.On("GetBytes").Return(optionSlotInfo, nil)
	mockSlotInfo.On("Put", authoredSlotInfo).Return()

	result, err := target.OnInitialize(1)
	assert.Nil(t, err)

	assert.Equal(t, dbWeight.ReadsWrites(2, 1), result)
	mockAuthorities.AssertCalled(t, "Get")
	mockAuraModule.AssertCalled(t, "StorageCurrentSlot")
	mockSlotInfo.AssertCalled(t, "GetBytes")
	mockSlotInfo.AssertCalled(t, "Put", authoredSlotInfo)
}

func Test_Module_OnInitialize_NotFound(t *testing.T) {
	target := setupModule()

	mockAuthorities.On("Get").Return(sc.Sequence[primitives.Sr25519PublicKey]{}, nil)
	mockAuraModule.On("StorageCurrentSlot").Return(currentSlot, nil)
	mockSlotInfo.On("GetBytes").Return(sc.NewOption[sc.Sequence[sc.U8]](nil), nil)
	mockSlotInfo.On("Put", slotInfo).Return()

	result, err := target.OnInitialize(1)
	assert.Nil(t, err)

	assert.Equal(t, dbWeight.ReadsWrites(2, 1), result)

	mockAuthorities.AssertCalled(t, "Get")
	mockAuraModule.AssertCalled(t, "StorageCurrentSlot")
	mockSlotInfo.AssertCalled(t, "GetBytes")
	mockSlotInfo.AssertCalled(t, "Put", slotInfo)
}

func Test_Module_OnFinalize(t *testing.T) {
	target := setupModule()
	expectAuthorities := sc.Sequence[primitives.Sr25519PublicKey]{}

	mockAuraModule.On("StorageAuthorities").Return(expectAuthorities, nil)
	mockAuthorities.On("Put", expectAuthorities).Return()

	err := target.OnFinalize(1)
	assert.Nil(t, err)

	mockAuraModule.AssertCalled(t, "StorageAuthorities")
	mockAuthorities.AssertCalled(t, "Put", expectAuthorities)
}

func Test_Module_OnFinalize_err(t *testing.T) {
	target := setupModule()
	expectError := errors.New("err")

	mockAuraModule.On("StorageAuthorities").Return(sc.Sequence[primitives.Sr25519PublicKey]{}, expectError)

	err := target.OnFinalize(1)
	assert.Equal(t, expectError, err)

	mockAuraModule.AssertCalled(t, "StorageAuthorities")
}

func Test_Module_Metadata(t *testing.T) {
	expect := primitives.MetadataModule{
		Version: primitives.ModuleVersion14,
		ModuleV14: primitives.MetadataModuleV14{
			Name: "AuraExt",
			Storage: sc.NewOption[primitives.MetadataModuleStorage](primitives.MetadataModuleStorage{
				Prefix: "AuraExt",
				Items: sc.Sequence[primitives.MetadataModuleStorageEntry]{
					primitives.NewMetadataModuleStorageEntry(
						"Authorities",
						primitives.MetadataModuleStorageEntryModifierDefault,
						primitives.NewMetadataModuleStorageEntryDefinitionPlain(sc.ToCompact(metadata.TypesAuraStorageAuthorities)),
						"The current authority set."),
					primitives.NewMetadataModuleStorageEntry(
						"CurrentSlot",
						primitives.MetadataModuleStorageEntryModifierDefault,
						primitives.NewMetadataModuleStorageEntryDefinitionPlain(sc.ToCompact(metadata.TypesAuraSlot)),
						"The current slot of this block.   This will be set in `on_initialize`."),
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
		},
	}

	target := setupModule()

	assert.Equal(t, expect, target.Metadata())
}

func setupModule() Module {
	mockStorage = new(mocks.IoStorage)
	mockAuthorities = new(mocks.StorageValue[sc.Sequence[primitives.Sr25519PublicKey]])
	mockSlotInfo = new(mocks.StorageValue[SlotInfo])
	mockAuraModule = new(mocks.AuraModule)

	config := NewConfig(mockStorage, dbWeight)

	target := New(moduleId, config, mockAuraModule, logger)
	target.storage.Authorities = mockAuthorities
	target.storage.SlotInfo = mockSlotInfo

	return target
}
