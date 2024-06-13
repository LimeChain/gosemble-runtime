package timestamp

import (
	"errors"
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants/metadata"
	"github.com/LimeChain/gosemble/mocks"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
)

var (
	dbWeight = primitives.RuntimeDbWeight{
		Read:  1,
		Write: 2,
	}
	minimumPeriod = sc.U64(5)
	ts            = sc.U64(1000)
	mockIoStorage *mocks.IoStorage
	mockCall      *mocks.Call
)

var (
	errorCannotGetStorageValue = errors.New("cannot get storage value")
	mdGenerator                = primitives.NewMetadataTypeGenerator()
)

func Test_Module_GetIndex(t *testing.T) {
	target := setupModule()
	assert.Equal(t, sc.U8(moduleId), target.GetIndex())
}

func Test_Module_name(t *testing.T) {
	target := setupModule()
	assert.Equal(t, name, target.name())
}

func Test_Module_Functions(t *testing.T) {
	target := setupModule()

	assert.Equal(t, 1, len(target.Functions()))
	assert.NotNil(t, target.Functions()[functionSetIndex])
}

func Test_Module_InherentIdentifier(t *testing.T) {
	target := setupModule()

	assert.Equal(t, inherentIdentifier, target.InherentIdentifier())
}

func Test_Module_IsInherent(t *testing.T) {
	target := setupModule()

	mockCall.On("ModuleIndex").Return(sc.U8(moduleId))
	mockCall.On("FunctionIndex").Return(sc.U8(functionSetIndex))

	assert.Equal(t, true, target.IsInherent(mockCall))

	mockCall.AssertCalled(t, "ModuleIndex")
	mockCall.AssertCalled(t, "FunctionIndex")
}

func Test_Module_PreDispatch(t *testing.T) {
	target := setupModule()

	result, err := target.PreDispatch(mockCall)

	assert.Nil(t, err)
	assert.Equal(t, sc.Empty{}, result)
}

func Test_Module_ValidateUnsigned(t *testing.T) {
	target := setupModule()

	result, err := target.ValidateUnsigned(primitives.TransactionSource{}, mockCall)

	assert.Nil(t, err)
	assert.Equal(t, primitives.DefaultValidTransaction(), result)
}

func Test_Module_OnFinalize_Nil(t *testing.T) {
	target := setupModule()

	mockStorageDidUpdate.On("TakeBytes").Return([]byte(nil), nil)

	result := target.OnFinalize(ts)

	assert.Equal(t, errTimestampNotUpdated, result)
	mockStorageDidUpdate.AssertCalled(t, "TakeBytes")
}

func Test_Module_OnFinalize_CannotTakeStorageValue(t *testing.T) {
	target := setupModule()

	mockStorageDidUpdate.On("TakeBytes").Return([]byte(nil), errorCannotGetStorageValue)

	result := target.OnFinalize(ts)

	assert.Equal(t, errorCannotGetStorageValue, result)
	mockStorageDidUpdate.AssertCalled(t, "TakeBytes")
}

func Test_Module_OnFinalize(t *testing.T) {
	target := setupModule()

	mockStorageDidUpdate.On("TakeBytes").Return([]byte("test"), nil)

	result := target.OnFinalize(ts)

	assert.NoError(t, result)
	mockStorageDidUpdate.AssertCalled(t, "TakeBytes")
}

func Test_Module_CreateInherent(t *testing.T) {
	target := setupModule()

	data := primitives.NewInherentData()
	assert.NoError(t, data.Put(inherentIdentifier, ts))
	expect := sc.NewOption[primitives.Call](newCallSetWithArgs(moduleId, functionSetIndex, sc.NewVaryingData(sc.ToCompact(ts+minimumPeriod))))

	mockStorageNow.On("Get").Return(ts, nil)

	result, err := target.CreateInherent(*data)
	assert.Nil(t, err)

	assert.Equal(t, expect, result)

	mockStorageNow.AssertCalled(t, "Get")
}

func Test_Module_CreateInherent_MoreThanStorageTimestamp(t *testing.T) {
	target := setupModule()

	data := primitives.NewInherentData()
	assert.NoError(t, data.Put(inherentIdentifier, ts+10))
	expect := sc.NewOption[primitives.Call](newCallSetWithArgs(moduleId, functionSetIndex, sc.NewVaryingData(sc.ToCompact(ts+10))))

	mockStorageNow.On("Get").Return(ts, nil)

	result, err := target.CreateInherent(*data)
	assert.Nil(t, err)

	assert.Equal(t, expect, result)

	mockStorageNow.AssertCalled(t, "Get")
}

func Test_Module_CreateInherent_NotProvided(t *testing.T) {
	data := primitives.NewInherentData()
	target := setupModule()

	result, err := target.CreateInherent(*data)
	assert.Equal(t, sc.NewOption[primitives.Call](nil), result)
	assert.Equal(t, errTimestampInherentNotProvided, err)
}

func Test_Module_CreateInherent_CannotGetStorageValue(t *testing.T) {
	target := setupModule()

	data := primitives.NewInherentData()
	assert.NoError(t, data.Put(inherentIdentifier, ts))

	mockStorageNow.On("Get").Return(ts, errorCannotGetStorageValue)

	result, err := target.CreateInherent(*data)

	assert.Equal(t, sc.NewOption[primitives.Call](nil), result)
	assert.Equal(t, errorCannotGetStorageValue, err)
	mockStorageNow.AssertCalled(t, "Get")
}

func Test_Module_CreateInherent_InherentDataNotCorrectlyEncoded(t *testing.T) {
	target := setupModule()

	invalid := sc.U32(1)
	inherentData := primitives.NewInherentData()
	assert.NoError(t, inherentData.Put(inherentIdentifier, invalid))

	result, err := target.CreateInherent(*inherentData)
	assert.Equal(t, sc.NewOption[primitives.Call](nil), result)
	assert.Equal(t, errTimestampInherentDataNotCorrectlyEncoded, err)
}

func Test_Module_CheckInherent(t *testing.T) {
	target := setupModule()
	inherentData := primitives.NewInherentData()
	inherentData.Put(inherentIdentifier, ts)

	validTs := sc.U64(2_000)

	mockCall.On("ModuleIndex").Return(sc.U8(moduleId))
	mockCall.On("FunctionIndex").Return(sc.U8(functionSetIndex))
	mockCall.On("Args").Return(sc.NewVaryingData(sc.ToCompact(validTs)))
	mockStorageNow.On("Get").Return(ts, nil)

	result := target.CheckInherent(mockCall, *inherentData)

	assert.Nil(t, result)

	mockCall.AssertCalled(t, "ModuleIndex")
	mockCall.AssertCalled(t, "FunctionIndex")
	mockCall.AssertCalled(t, "Args")
	mockStorageNow.AssertCalled(t, "Get")
}

func Test_Module_CheckInherent_NotInherent(t *testing.T) {
	target := setupModule()
	inherentData := primitives.NewInherentData()

	mockCall.On("ModuleIndex").Return(sc.U8(3))

	result := target.CheckInherent(mockCall, *inherentData)

	assert.Equal(t, primitives.NewTimestampErrorInvalid(), result)

	mockCall.AssertCalled(t, "ModuleIndex")
}

func Test_Module_CheckInherent_InherentNotProvided(t *testing.T) {
	target := setupModule()
	inherentData := primitives.NewInherentData()

	mockCall.On("ModuleIndex").Return(sc.U8(moduleId))
	mockCall.On("FunctionIndex").Return(sc.U8(functionSetIndex))
	mockCall.On("Args").Return(sc.NewVaryingData(sc.ToCompact(ts)))

	err := target.CheckInherent(mockCall, *inherentData)
	assert.Equal(t, errTimestampInherentNotProvided, err)

	mockCall.AssertCalled(t, "ModuleIndex")
	mockCall.AssertCalled(t, "FunctionIndex")
	mockCall.AssertCalled(t, "Args")
}

func Test_Module_CheckInherent_InherentDataNotCorrectlyEncoded(t *testing.T) {
	target := setupModule()

	invalid := sc.U32(1)
	inherentData := primitives.NewInherentData()
	assert.NoError(t, inherentData.Put(inherentIdentifier, invalid))

	mockCall.On("ModuleIndex").Return(sc.U8(moduleId))
	mockCall.On("FunctionIndex").Return(sc.U8(functionSetIndex))
	mockCall.On("Args").Return(sc.NewVaryingData(sc.ToCompact(ts)))

	err := target.CheckInherent(mockCall, *inherentData)
	assert.Equal(t, errTimestampInherentDataNotCorrectlyEncoded, err)

	mockCall.AssertCalled(t, "ModuleIndex")
	mockCall.AssertCalled(t, "FunctionIndex")
	mockCall.AssertCalled(t, "Args")
}

func Test_Module_CheckInherent_CannotGetStorageValue(t *testing.T) {
	target := setupModule()
	inherentData := primitives.NewInherentData()
	assert.NoError(t, inherentData.Put(inherentIdentifier, ts))

	mockCall.On("ModuleIndex").Return(sc.U8(moduleId))
	mockCall.On("FunctionIndex").Return(sc.U8(functionSetIndex))
	mockCall.On("Args").Return(sc.NewVaryingData(sc.ToCompact(sc.U64(2))))
	mockStorageNow.On("Get").Return(ts, errorCannotGetStorageValue)

	err := target.CheckInherent(mockCall, *inherentData)
	assert.Equal(t, errorCannotGetStorageValue, err)

	mockCall.AssertCalled(t, "ModuleIndex")
	mockCall.AssertCalled(t, "FunctionIndex")
	mockCall.AssertCalled(t, "Args")
	mockStorageNow.AssertCalled(t, "Get")
}

func Test_Module_CheckInherent_TooFarInFuture(t *testing.T) {
	target := setupModule()
	inherentData := primitives.NewInherentData()
	inherentData.Put(inherentIdentifier, ts)
	tsTooFar := sc.U64(50_000) // 50 seconds

	expect := primitives.NewTimestampErrorTooFarInFuture()

	mockCall.On("ModuleIndex").Return(sc.U8(moduleId))
	mockCall.On("FunctionIndex").Return(sc.U8(functionSetIndex))
	mockCall.On("Args").Return(sc.NewVaryingData(sc.ToCompact(tsTooFar)))
	mockStorageNow.On("Get").Return(ts, nil)

	result := target.CheckInherent(mockCall, *inherentData)

	assert.Equal(t, expect, result)

	mockCall.AssertCalled(t, "ModuleIndex")
	mockCall.AssertCalled(t, "FunctionIndex")
	mockCall.AssertCalled(t, "Args")
	mockStorageNow.AssertCalled(t, "Get")
}

func Test_Module_CheckInherent_TooEarly(t *testing.T) {
	target := setupModule()
	inherentData := primitives.NewInherentData()
	inherentData.Put(inherentIdentifier, ts)
	expect := primitives.NewTimestampErrorTooEarly()

	mockCall.On("ModuleIndex").Return(sc.U8(moduleId))
	mockCall.On("FunctionIndex").Return(sc.U8(functionSetIndex))
	mockCall.On("Args").Return(sc.NewVaryingData(sc.ToCompact(ts)))
	mockStorageNow.On("Get").Return(ts, nil)

	result := target.CheckInherent(mockCall, *inherentData)

	assert.Equal(t, expect, result)

	mockCall.AssertCalled(t, "ModuleIndex")
	mockCall.AssertCalled(t, "FunctionIndex")
	mockCall.AssertCalled(t, "Args")
	mockStorageNow.AssertCalled(t, "Get")
}

func Test_Module_Metadata(t *testing.T) {
	expectedTimestampCallsMetadataId := mdGenerator.GetLastAvailableIndex() + 1
	expectedCompactU64TypeId := expectedTimestampCallsMetadataId + 1

	expectMetadataTypes := sc.Sequence[primitives.MetadataType]{
		primitives.NewMetadataType(expectedCompactU64TypeId, "CompactU64", primitives.NewMetadataTypeDefinitionCompact(sc.ToCompact(metadata.PrimitiveTypesU64))),
		primitives.NewMetadataTypeWithParam(expectedTimestampCallsMetadataId, "Timestamp calls", sc.Sequence[sc.Str]{"pallet_timestamp", "pallet", "Call"}, primitives.NewMetadataTypeDefinitionVariant(
			sc.Sequence[primitives.MetadataDefinitionVariant]{
				primitives.NewMetadataDefinitionVariant(
					"set",
					sc.Sequence[primitives.MetadataTypeDefinitionField]{
						primitives.NewMetadataTypeDefinitionField(expectedCompactU64TypeId),
					},
					functionSetIndex,
					"Set the current time."),
			}), primitives.NewMetadataEmptyTypeParameter("T")),
	}
	moduleV14 := primitives.MetadataModuleV14{
		Name: name,
		Storage: sc.NewOption[primitives.MetadataModuleStorage](primitives.MetadataModuleStorage{
			Prefix: name,
			Items: sc.Sequence[primitives.MetadataModuleStorageEntry]{
				primitives.NewMetadataModuleStorageEntry(
					"Now",
					primitives.MetadataModuleStorageEntryModifierDefault,
					primitives.NewMetadataModuleStorageEntryDefinitionPlain(sc.ToCompact(metadata.PrimitiveTypesU64)),
					"Current time for the current block."),
				primitives.NewMetadataModuleStorageEntry(
					"DidUpdate",
					primitives.MetadataModuleStorageEntryModifierDefault,
					primitives.NewMetadataModuleStorageEntryDefinitionPlain(sc.ToCompact(metadata.PrimitiveTypesBool)),
					"Did the timestamp get updated in this block?"),
			},
		}),
		Call: sc.NewOption[sc.Compact](sc.ToCompact(expectedTimestampCallsMetadataId)),
		CallDef: sc.NewOption[primitives.MetadataDefinitionVariant](
			primitives.NewMetadataDefinitionVariantStr(
				name,
				sc.Sequence[primitives.MetadataTypeDefinitionField]{
					primitives.NewMetadataTypeDefinitionFieldWithName(expectedTimestampCallsMetadataId, "self::sp_api_hidden_includes_construct_runtime::hidden_include::dispatch\n::CallableCallFor<Timestamp, Runtime>"),
				},
				moduleId,
				"Call.Timestamp"),
		),
		Event:    sc.NewOption[sc.Compact](nil),
		EventDef: sc.NewOption[primitives.MetadataDefinitionVariant](nil),
		Constants: sc.Sequence[primitives.MetadataModuleConstant]{
			primitives.NewMetadataModuleConstant(
				"MinimumPeriod",
				sc.ToCompact(metadata.PrimitiveTypesU64),
				sc.BytesToSequenceU8(minimumPeriod.Bytes()),
				"The minimum period between blocks. Beware that this is different to the *expected*  period that the block production apparatus provides.",
			),
		},
		Error:    sc.NewOption[sc.Compact](nil),
		ErrorDef: sc.NewOption[primitives.MetadataDefinitionVariant](nil),
		Index:    moduleId,
	}

	expectMetadataModule := primitives.MetadataModule{
		Version:   primitives.ModuleVersion14,
		ModuleV14: moduleV14,
	}

	target := setupModule()

	resultMetadataModule := target.Metadata()
	resultTypes := mdGenerator.GetMetadataTypes()

	assert.Equal(t, expectMetadataTypes, resultTypes)
	assert.Equal(t, expectMetadataModule, resultMetadataModule)
}

func setupModule() Module {
	mockIoStorage = new(mocks.IoStorage)
	mockOnTimestampSet = new(mocks.OnTimestampSet)
	mockStorageNow = new(mocks.StorageValue[sc.U64])
	mockStorageDidUpdate = new(mocks.StorageValue[sc.Bool])
	mockCall = new(mocks.Call)

	config := NewConfig(mockIoStorage, mockOnTimestampSet, dbWeight, minimumPeriod)

	target := New(moduleId, config, mdGenerator)
	target.storage.DidUpdate = mockStorageDidUpdate
	target.storage.Now = mockStorageNow

	return target
}
