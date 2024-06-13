package sudo

import (
	"errors"
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants"
	"github.com/LimeChain/gosemble/constants/metadata"
	"github.com/LimeChain/gosemble/mocks"
	"github.com/LimeChain/gosemble/primitives/log"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
)

const (
	moduleId = 7
)

var (
	dbWeight = primitives.RuntimeDbWeight{
		Read:  1,
		Write: 2,
	}

	signedOrigin = primitives.NewRawOriginSigned(constants.OneAccountId)

	oldKey       = constants.TwoAccountId
	oldKeyOption = sc.NewOption[primitives.AccountId](oldKey)
	newKey       = constants.OneAccountId

	newMultiAddress = primitives.NewMultiAddressId(constants.OneAccountId)

	mdGenerator                           = primitives.NewMetadataTypeGenerator()
	unknownTransactionNoUnsignedValidator = primitives.NewTransactionValidityError(primitives.NewUnknownTransactionNoUnsignedValidator())
	eventFunc                             = func(u8 sc.U8, outcome primitives.DispatchOutcome) primitives.Event { return newEventSudid(u8, outcome) }
	dispatchErrOther                      = primitives.NewDispatchErrorOther("error")
	dispatchOutcomeEmpty, _               = primitives.NewDispatchOutcome(sc.Empty{})
	dispatchOutomeErr, _                  = primitives.NewDispatchOutcome(dispatchErrOther)
)
var (
	mockStorage        *mocks.IoStorage
	mockEventDepositor *mocks.EventDepositor
	mockStorageKey     *mocks.StorageValue[primitives.AccountId]
	mockCall           *mocks.Call
)

func Test_Module_GetIndex(t *testing.T) {
	target := setupModule()

	assert.Equal(t, sc.U8(moduleId), target.GetIndex())
}

func Test_Module_Functions(t *testing.T) {
	target := setupModule()
	functions := target.Functions()

	assert.Equal(t, 5, len(functions))
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

func Test_Module_executeCall(t *testing.T) {
	target := setupModule()

	mockCall.On("Args").Return(sc.NewVaryingData())
	mockCall.On("Dispatch", signedOrigin, sc.NewVaryingData()).Return(primitives.PostDispatchInfo{}, nil)
	mockEventDepositor.On("DepositEvent", newEventSudid(moduleId, dispatchOutcomeEmpty)).Return()

	res, err := target.executeCall(signedOrigin, mockCall, eventFunc)
	assert.Nil(t, err)
	assert.Equal(t, primitives.PostDispatchInfo{
		PaysFee: primitives.PaysNo,
	}, res)

	mockCall.AssertCalled(t, "Args")
	mockCall.AssertCalled(t, "Dispatch", signedOrigin, sc.NewVaryingData())
	mockEventDepositor.AssertCalled(t, "DepositEvent", newEventSudid(moduleId, dispatchOutcomeEmpty))
}

func Test_Module_executeCall_CallErr(t *testing.T) {
	target := setupModule()

	mockCall.On("Args").Return(sc.NewVaryingData())
	mockCall.On("Dispatch", signedOrigin, sc.NewVaryingData()).Return(primitives.PostDispatchInfo{}, dispatchErrOther)
	mockEventDepositor.On("DepositEvent", newEventSudid(moduleId, dispatchOutomeErr)).Return()

	res, err := target.executeCall(signedOrigin, mockCall, eventFunc)
	assert.Nil(t, err)
	assert.Equal(t, primitives.PostDispatchInfo{
		PaysFee: primitives.PaysNo,
	}, res)

	mockCall.AssertCalled(t, "Args")
	mockCall.AssertCalled(t, "Dispatch", signedOrigin, sc.NewVaryingData())
	mockEventDepositor.AssertCalled(t, "DepositEvent", newEventSudid(moduleId, dispatchOutomeErr))
}

func Test_Module_executeCall_CallErr_Invalid(t *testing.T) {
	target := setupModule()

	mockCall.On("Args").Return(sc.NewVaryingData())
	mockCall.On("Dispatch", signedOrigin, sc.NewVaryingData()).Return(primitives.PostDispatchInfo{}, errors.New("error"))

	mockEventDepositor.On("DepositEvent", newEventSudid(moduleId, dispatchOutomeErr)).Return()

	res, err := target.executeCall(signedOrigin, mockCall, eventFunc)
	assert.Equal(t, dispatchErrOther, err)
	assert.Equal(t, primitives.PostDispatchInfo{}, res)

	mockCall.AssertCalled(t, "Args")
	mockCall.AssertCalled(t, "Dispatch", signedOrigin, sc.NewVaryingData())
	mockEventDepositor.AssertNotCalled(t, "DepositEvent", newEventSudid(moduleId, dispatchOutomeErr))
}

func Test_Module_ensureSudo_Root(t *testing.T) {
	target := setupModule()

	err := target.ensureSudo(primitives.NewRawOriginRoot())
	assert.Nil(t, err)
}

func Test_Module_ensureSudo_None(t *testing.T) {
	target := setupModule()

	err := target.ensureSudo(primitives.NewRawOriginNone())
	assert.Equal(t, primitives.NewDispatchErrorBadOrigin(), err)
}

func Test_Module_ensureSudo_Signed(t *testing.T) {
	target := setupModule()

	mockStorageKey.On("Get").Return(constants.OneAccountId, nil)

	err := target.ensureSudo(signedOrigin)
	assert.Nil(t, err)

	mockStorageKey.AssertCalled(t, "Get")
}

func Test_Module_ensureSudo_Signed_StorageErr(t *testing.T) {
	expect := errors.New("storage error")
	target := setupModule()

	mockStorageKey.On("Get").Return(primitives.AccountId{}, expect)

	err := target.ensureSudo(signedOrigin)
	assert.Equal(t, primitives.NewDispatchErrorOther(sc.Str(expect.Error())), err)

	mockStorageKey.AssertCalled(t, "Get")
}

func Test_Module_ensureSudo_Signed_MismatchErr(t *testing.T) {
	target := setupModule()

	mockStorageKey.On("Get").Return(constants.ZeroAccountId, nil)

	err := target.ensureSudo(signedOrigin)
	assert.Equal(t, NewDispatchErrorRequireSudo(moduleId), err)

	mockStorageKey.AssertCalled(t, "Get")
}

func Test_Module_Metadata(t *testing.T) {
	target := setupModule()

	dataV14 := primitives.MetadataModuleV14{
		Name: name,
		Storage: sc.NewOption[primitives.MetadataModuleStorage](primitives.MetadataModuleStorage{
			Prefix: name,
			Items: sc.Sequence[primitives.MetadataModuleStorageEntry]{
				primitives.NewMetadataModuleStorageEntry(
					"Key",
					primitives.MetadataModuleStorageEntryModifierOptional,
					primitives.NewMetadataModuleStorageEntryDefinitionPlain(sc.ToCompact(metadata.TypesAddress32)),
					"The `AccountId` of the sudo key.",
				),
			},
		}),
		Call: sc.NewOption[sc.Compact](sc.ToCompact(metadata.TypesSudoCalls)),
		CallDef: sc.NewOption[primitives.MetadataDefinitionVariant](
			primitives.NewMetadataDefinitionVariantStr(
				name,
				sc.Sequence[primitives.MetadataTypeDefinitionField]{
					primitives.NewMetadataTypeDefinitionFieldWithName(metadata.TypesSudoCalls, "self::sp_api_hidden_includes_construct_runtime::hidden_include::dispatch\n::CallableCallFor<Sudo, Runtime>"),
				},
				moduleId,
				"Call.Sudo")),
		Event: sc.NewOption[sc.Compact](sc.ToCompact(metadata.TypesSudoEvent)),
		EventDef: sc.NewOption[primitives.MetadataDefinitionVariant](
			primitives.NewMetadataDefinitionVariantStr(
				name,
				sc.Sequence[primitives.MetadataTypeDefinitionField]{
					primitives.NewMetadataTypeDefinitionFieldWithName(metadata.TypesSudoEvent, "frame_sudo::Event<Runtime>"),
				},
				moduleId,
				"Events.System"),
		),
		Constants: sc.Sequence[primitives.MetadataModuleConstant]{},
		Error:     sc.NewOption[sc.Compact](sc.ToCompact(metadata.TypesSudoErrors)),
		ErrorDef: sc.NewOption[primitives.MetadataDefinitionVariant](
			primitives.NewMetadataDefinitionVariantStr(
				name,
				sc.Sequence[primitives.MetadataTypeDefinitionField]{
					primitives.NewMetadataTypeDefinitionField(metadata.TypesSudoErrors),
				},
				moduleId,
				"Errors.Sudo"),
		),
		Index: moduleId,
	}

	expectMetadata := primitives.MetadataModule{
		Version:   primitives.ModuleVersion14,
		ModuleV14: dataV14,
	}

	expectMetadataTypes := sc.Sequence[primitives.MetadataType]{
		primitives.NewMetadataTypeWithPath(
			metadata.TypesSudoEvent,
			"pallet_sudo pallet Sudo",
			sc.Sequence[sc.Str]{"pallet_sudo", "pallet", "Event"},
			primitives.NewMetadataTypeDefinitionVariant(
				sc.Sequence[primitives.MetadataDefinitionVariant]{
					primitives.NewMetadataDefinitionVariant(
						"Sudid",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesDispatchOutcome, "sudo_result", "DispatchResult"),
						},
						EventSudid,
						"Events.Sudid"),
					primitives.NewMetadataDefinitionVariant(
						"KeyChanged",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesOptionAccountId, "old", "Option<T::AccountId>"),
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesAddress32, "new", "T::AccountId"),
						},
						EventKeyChanged,
						"Events.KeyChanged"),
					primitives.NewMetadataDefinitionVariant(
						"KeyRemoved",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{},
						EventKeyRemoved,
						"Events.KeyRemoved"),
					primitives.NewMetadataDefinitionVariant(
						"SudoAsDone",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesDispatchOutcome, "sudo_result", "DispatchResult"),
						},
						EventSudoAsDone,
						"Events.SudoAsDone"),
				})),

		primitives.NewMetadataTypeWithParams(metadata.TypesSudoErrors,
			"pallet_sudo pallet Error",
			sc.Sequence[sc.Str]{"pallet_sudo", "pallet", "Error"},
			primitives.NewMetadataTypeDefinitionVariant(
				sc.Sequence[primitives.MetadataDefinitionVariant]{
					primitives.NewMetadataDefinitionVariant(
						"RequireSudo",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{},
						ErrorRequireSudo,
						"Sender must be the sudo account.",
					),
				}),
			sc.Sequence[primitives.MetadataTypeParameter]{
				primitives.NewMetadataEmptyTypeParameter("T"),
				primitives.NewMetadataEmptyTypeParameter("I"),
			}),

		primitives.NewMetadataTypeWithParam(metadata.TypesSudoCalls,
			"Sudo calls",
			sc.Sequence[sc.Str]{"pallet_sudo", "pallet", "Call"},
			primitives.NewMetadataTypeDefinitionVariant(
				sc.Sequence[primitives.MetadataDefinitionVariant]{
					primitives.NewMetadataDefinitionVariant(
						"sudo",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.RuntimeCall, "call", "Box<<T as Config>::RuntimeCall>"),
						},
						functionSudo,
						"Authenticates the sudo key and dispatches a function call with `Root` origin."),
					primitives.NewMetadataDefinitionVariant(
						"sudo_unchecked_weight",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.RuntimeCall, "call", "Box<<T as Config>::RuntimeCall>"),
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesWeight, "weight", "Weight"),
						},
						functionSudoUncheckedWeight,
						"Authenticates the sudo key and dispatches a function call with `Root` origin."+
							"This function does not check the weight of the call, and instead allows the"+
							"Sudo user to specify the weight of the call."+
							"The dispatch origin for this call must be `Signed`.",
					),
					primitives.NewMetadataDefinitionVariant(
						"set_key",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesMultiAddress, "new", "AccountIdLookupOf<T>"),
						},
						functionSetKey,
						"Authenticates the current sudo key and sets the given `AccountId` as the new sudo."),
					primitives.NewMetadataDefinitionVariant(
						"sudo_as",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesMultiAddress, "who", "AccountIdLookupOf<T>"),
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.RuntimeCall, "call", "Box<<T as Config>::RuntimeCall>"),
						},
						functionSudoAs,
						"Authenticates the sudo key and dispatches a function call with `Signed` origin from a given account."+
							"The dispatch origin for this call must be `Signed`.",
					),
					primitives.NewMetadataDefinitionVariant(
						"remove_key",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{},
						functionRemoveKey,
						"Permanently removes the sudo key. This cannot be undone."),
				}),
			primitives.NewMetadataEmptyTypeParameter("T")),
	}

	metadataModule := target.Metadata()
	metadataTypes := mdGenerator.GetMetadataTypes()

	assert.Equal(t, expectMetadataTypes, metadataTypes)
	assert.Equal(t, expectMetadata, metadataModule)
}

func setupModule() Module {
	mockStorage = new(mocks.IoStorage)
	mockEventDepositor = new(mocks.EventDepositor)
	mockStorageKey = new(mocks.StorageValue[primitives.AccountId])
	mockCall = new(mocks.Call)

	config := NewConfig(mockStorage, dbWeight, mockEventDepositor)

	target := New(moduleId, config, mdGenerator, log.NewLogger())

	target.storage.Key = mockStorageKey

	return target
}
