package transaction_payment

import (
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants"
	"github.com/LimeChain/gosemble/constants/metadata"
	"github.com/LimeChain/gosemble/mocks"
	"github.com/LimeChain/gosemble/primitives/types"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
)

const (
	moduleId                 = sc.U8(5)
	operationalFeeMultiplier = sc.U8(3)
)

var (
	who = constants.ZeroAccountId

	weightToFee types.WeightToFee = types.IdentityFee{}
	lengthToFee types.WeightToFee = types.IdentityFee{}

	blockWeights = types.BlockWeights{
		BaseBlock: types.WeightFromParts(1, 2),
		MaxBlock:  types.WeightFromParts(1000, 0),
		PerClass: types.PerDispatchClassWeightsPerClass{
			Normal: types.WeightsPerClass{
				BaseExtrinsic: types.WeightFromParts(5, 6),
			},
			Operational: types.WeightsPerClass{
				BaseExtrinsic: types.WeightFromParts(100, 0),
			},
			Mandatory: types.WeightsPerClass{
				BaseExtrinsic: types.WeightFromParts(9, 10),
			},
		},
	}
)

var (
	mdGenerator = primitives.NewMetadataTypeGenerator()

	noUnsignedValidatorError = types.NewTransactionValidityError(
		types.NewUnknownTransactionNoUnsignedValidator(),
	)

	typesWeightId, _ = mdGenerator.GetId("Weight")

	expectedMetadataTypes = sc.Sequence[types.MetadataType]{
		types.NewMetadataTypeWithPath(metadata.TypesTransactionPaymentReleases, "Releases", sc.Sequence[sc.Str]{"pallet_transaction_payment", "Releases"}, types.NewMetadataTypeDefinitionVariant(
			sc.Sequence[types.MetadataDefinitionVariant]{
				types.NewMetadataDefinitionVariant(
					"V1Ancient",
					sc.Sequence[types.MetadataTypeDefinitionField]{},
					0,
					"Original version of the pallet."),
				types.NewMetadataDefinitionVariant(
					"V2",
					sc.Sequence[types.MetadataTypeDefinitionField]{},
					1,
					"One that bumps the usage to FixedU128 from FixedI128."),
			})),

		types.NewMetadataTypeWithParam(metadata.TypesTransactionPaymentEvent, "pallet_transaction_payment pallet Event", sc.Sequence[sc.Str]{"pallet_transaction_payment", "pallet", "Event"}, types.NewMetadataTypeDefinitionVariant(
			sc.Sequence[types.MetadataDefinitionVariant]{
				types.NewMetadataDefinitionVariant(
					"TransactionFeePaid",
					sc.Sequence[types.MetadataTypeDefinitionField]{
						types.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesAddress32, "who", "T::AccountId"),
						types.NewMetadataTypeDefinitionFieldWithNames(metadata.PrimitiveTypesU128, "actual_fee", "BalanceOf<T>"),
						types.NewMetadataTypeDefinitionFieldWithNames(metadata.PrimitiveTypesU128, "tip", "BalanceOf<T>"),
					},
					0,
					"Event.TransactionFeePaid"),
			}), types.NewMetadataEmptyTypeParameter("T")),

		primitives.NewMetadataTypeWithParams(metadata.TypesTransactionPaymentRuntimeDispatchInfo, "pallet_transaction_payment types RuntimeDispatchInfo", sc.Sequence[sc.Str]{"pallet_transaction_payment", "types", "RuntimeDispatchInfo"}, primitives.NewMetadataTypeDefinitionComposite(
			sc.Sequence[primitives.MetadataTypeDefinitionField]{
				primitives.NewMetadataTypeDefinitionFieldWithName(typesWeightId, "Weight"),
				primitives.NewMetadataTypeDefinitionFieldWithName(metadata.TypesDispatchClass, "Class"),
				primitives.NewMetadataTypeDefinitionFieldWithName(metadata.PrimitiveTypesU128, "Balance")}),
			sc.Sequence[primitives.MetadataTypeParameter]{
				primitives.NewMetadataTypeParameter(metadata.PrimitiveTypesU128, "Balance"),
				primitives.NewMetadataTypeParameter(typesWeightId, "Weight"),
			}),

		// type 910
		primitives.NewMetadataTypeWithParams(metadata.TypesTransactionPaymentInclusionFee, "pallet_transaction_payment types InclusionFee", sc.Sequence[sc.Str]{"pallet_transaction_payment", "types", "InclusionFee"}, primitives.NewMetadataTypeDefinitionComposite(
			sc.Sequence[primitives.MetadataTypeDefinitionField]{
				primitives.NewMetadataTypeDefinitionFieldWithName(metadata.PrimitiveTypesU128, "Balance"),
				primitives.NewMetadataTypeDefinitionFieldWithName(metadata.PrimitiveTypesU128, "Balance"),
				primitives.NewMetadataTypeDefinitionFieldWithName(metadata.PrimitiveTypesU128, "Balance")}),
			sc.Sequence[primitives.MetadataTypeParameter]{
				primitives.NewMetadataTypeParameter(metadata.PrimitiveTypesU128, "Balance"),
			}),

		primitives.NewMetadataTypeWithParam(metadata.TypeOptionInclusionFee, "Option<InclusionFee>", sc.Sequence[sc.Str]{"Option"}, primitives.NewMetadataTypeDefinitionVariant(
			sc.Sequence[primitives.MetadataDefinitionVariant]{
				primitives.NewMetadataDefinitionVariant(
					"None",
					sc.Sequence[primitives.MetadataTypeDefinitionField]{},
					0,
					""),
				primitives.NewMetadataDefinitionVariant(
					"Some",
					sc.Sequence[primitives.MetadataTypeDefinitionField]{
						primitives.NewMetadataTypeDefinitionField(metadata.TypesTransactionPaymentInclusionFee),
					},
					1,
					""),
			}),
			primitives.NewMetadataTypeParameter(metadata.TypesTransactionPaymentInclusionFee, "T")),

		primitives.NewMetadataTypeWithParam(metadata.TypesTransactionPaymentFeeDetails, "pallet_transaction_payment types FeeDetails", sc.Sequence[sc.Str]{"pallet_transaction_payment", "types", "FeeDetails"}, primitives.NewMetadataTypeDefinitionComposite(
			sc.Sequence[primitives.MetadataTypeDefinitionField]{
				primitives.NewMetadataTypeDefinitionFieldWithName(metadata.TypeOptionInclusionFee, "Option<InclusionFee<Balance>>"),
				primitives.NewMetadataTypeDefinitionFieldWithName(metadata.PrimitiveTypesU128, "Balance")}),
			primitives.NewMetadataTypeParameter(metadata.PrimitiveTypesU128, "Balance"),
		),
	}

	moduleV14 = types.MetadataModuleV14{
		Name: "TransactionPayment",
		Storage: sc.NewOption[primitives.MetadataModuleStorage](primitives.MetadataModuleStorage{
			Prefix: "TransactionPayment",
			Items: sc.Sequence[primitives.MetadataModuleStorageEntry]{
				primitives.NewMetadataModuleStorageEntry(
					"NextFeeMultiplier",
					primitives.MetadataModuleStorageEntryModifierDefault,
					primitives.NewMetadataModuleStorageEntryDefinitionPlain(sc.ToCompact(metadata.TypesFixedU128)),
					"NextFeeMultiplier"),
				primitives.NewMetadataModuleStorageEntry(
					"StorageVersion",
					primitives.MetadataModuleStorageEntryModifierDefault,
					primitives.NewMetadataModuleStorageEntryDefinitionPlain(sc.ToCompact(metadata.TypesTransactionPaymentReleases)),
					"StorageVersion"),
			},
		}),
		Call:    sc.NewOption[sc.Compact](nil),
		CallDef: sc.NewOption[types.MetadataDefinitionVariant](nil),
		Event:   sc.NewOption[sc.Compact](sc.ToCompact(metadata.TypesTransactionPaymentEvent)),
		EventDef: sc.NewOption[types.MetadataDefinitionVariant](
			types.NewMetadataDefinitionVariantStr(
				"TransactionPayment",
				sc.Sequence[types.MetadataTypeDefinitionField]{
					types.NewMetadataTypeDefinitionFieldWithName(metadata.TypesTransactionPaymentEvent, "pallet_transaction_payment::Event<Runtime>"),
				},
				moduleId,
				"Events.TransactionPayment"),
		),
		Constants: sc.Sequence[types.MetadataModuleConstant]{
			types.NewMetadataModuleConstant(
				"OperationalFeeMultiplier",
				sc.ToCompact(metadata.PrimitiveTypesU8),
				sc.BytesToSequenceU8(operationalFeeMultiplier.Bytes()),
				"A fee multiplier for `Operational` extrinsics to compute \"virtual tip\" to boost their  `priority` ",
			),
		},
		Error:    sc.NewOption[sc.Compact](nil),
		ErrorDef: sc.NewOption[primitives.MetadataDefinitionVariant](nil),
		Index:    moduleId,
	}

	expectedMetadataModule = types.MetadataModule{
		Version:   primitives.ModuleVersion14,
		ModuleV14: moduleV14,
	}
)

var (
	target                module
	mockStorage           *mocks.IoStorage
	mockCall              *mocks.Call
	mockNextFeeMultiplier *mocks.StorageValue[sc.U128]
)

func setup() {
	mockStorage = new(mocks.IoStorage)
	mockNextFeeMultiplier = new(mocks.StorageValue[sc.U128])

	config := NewConfig(mockStorage, operationalFeeMultiplier, weightToFee, lengthToFee, blockWeights)
	target = New(moduleId, config, mdGenerator).(module)
	target.storage.NextFeeMultiplier = mockNextFeeMultiplier
}

func Test_GetIndex(t *testing.T) {
	setup()

	assert.Equal(t, moduleId, target.GetIndex())
}

func Test_Functions(t *testing.T) {
	setup()

	result := target.Functions()

	assert.Equal(t, 0, len(result))
}

func Test_PreDispatch(t *testing.T) {
	setup()

	result, err := target.PreDispatch(mockCall)

	assert.Nil(t, err)
	assert.Equal(t, sc.Empty{}, result)
}

func Test_ValidateUnsigned(t *testing.T) {
	setup()

	result, err := target.ValidateUnsigned(types.TransactionSource{}, mockCall)

	assert.Equal(t, noUnsignedValidatorError, err)
	assert.Equal(t, types.ValidTransaction{}, result)
}

func Test_Metadata(t *testing.T) {
	setup()

	metadataModule := target.Metadata()
	metadataTypes := mdGenerator.GetMetadataTypes()

	assert.Equal(t, expectedMetadataTypes, metadataTypes)
	assert.Equal(t, expectedMetadataModule, metadataModule)
}

func Test_ComputeFee_TipOnlyNoFee(t *testing.T) {
	setup()

	info := primitives.DispatchInfo{
		Weight:  primitives.WeightFromParts(0, 0),
		Class:   primitives.NewDispatchClassOperational(),
		PaysFee: primitives.PaysNo,
	}

	fee, err := target.ComputeFee(0, info, sc.NewU128(15))
	assert.Nil(t, err)

	mockNextFeeMultiplier.AssertNotCalled(t, "Get")
	assert.Equal(t, sc.NewU128(15), fee)
}

func Test_ComputeFee_NoTipOnlyBaseFee(t *testing.T) {
	setup()

	info := primitives.DispatchInfo{
		Weight:  primitives.WeightFromParts(0, 0),
		Class:   primitives.NewDispatchClassOperational(),
		PaysFee: primitives.PaysYes,
	}

	mockNextFeeMultiplier.On("Get").Return(sc.NewU128(0), nil)

	fee, err := target.ComputeFee(0, info, sc.NewU128(0))
	assert.Nil(t, err)

	mockNextFeeMultiplier.AssertCalled(t, "Get")
	assert.Equal(t, sc.NewU128(100), fee)
}

func Test_ComputeFee_TipPlusBaseFee(t *testing.T) {
	setup()

	info := primitives.DispatchInfo{
		Weight:  primitives.WeightFromParts(0, 0),
		Class:   primitives.NewDispatchClassOperational(),
		PaysFee: primitives.PaysYes,
	}

	mockNextFeeMultiplier.On("Get").Return(sc.NewU128(2), nil)

	fee, err := target.ComputeFee(0, info, sc.NewU128(69))
	assert.Nil(t, err)

	mockNextFeeMultiplier.AssertCalled(t, "Get")
	assert.Equal(t, sc.NewU128(169), fee)
}

func Test_ComputeFee_ByteFeePlusBaseFee(t *testing.T) {
	setup()

	info := primitives.DispatchInfo{
		Weight:  primitives.WeightFromParts(0, 0),
		Class:   primitives.NewDispatchClassOperational(),
		PaysFee: primitives.PaysYes,
	}

	mockNextFeeMultiplier.On("Get").Return(sc.NewU128(0), nil)

	fee, err := target.ComputeFee(42, info, sc.NewU128(0))
	assert.Nil(t, err)

	mockNextFeeMultiplier.AssertCalled(t, "Get")
	assert.Equal(t, sc.NewU128(142), fee)
}

func Test_ComputeFee_WeightFeePlusBaseFee(t *testing.T) {
	setup()

	info := primitives.DispatchInfo{
		Weight:  primitives.WeightFromParts(1000, 0),
		Class:   primitives.NewDispatchClassOperational(),
		PaysFee: primitives.PaysYes,
	}

	mockNextFeeMultiplier.On("Get").Return(sc.NewU128(0), nil)

	fee, err := target.ComputeFee(0, info, sc.NewU128(0))
	assert.Nil(t, err)

	mockNextFeeMultiplier.AssertCalled(t, "Get")
	assert.Equal(t, sc.NewU128(100), fee)
}

func Test_ComputeFeeDetails(t *testing.T) {
	setup()

	info := primitives.DispatchInfo{
		Weight:  primitives.WeightFromParts(0, 0),
		Class:   primitives.NewDispatchClassOperational(),
		PaysFee: primitives.PaysYes,
	}

	mockNextFeeMultiplier.On("Get").Return(sc.NewU128(0), nil)

	result, err := target.ComputeFeeDetails(5, info, sc.NewU128(3))
	assert.NoError(t, err)

	mockNextFeeMultiplier.AssertCalled(t, "Get")
	assert.Equal(t, sc.NewU128(3), result.Tip)
	assert.Equal(t, sc.NewU128(0), result.InclusionFee.Value.AdjustedWeightFee)
	assert.Equal(t, sc.NewU128(100), result.InclusionFee.Value.BaseFee)
	assert.Equal(t, sc.NewU128(5), result.InclusionFee.Value.LenFee)
	assert.Equal(t, sc.NewU128(108), result.FinalFee())
}

func Test_ComputeActualFee(t *testing.T) {
	setup()

	info := primitives.DispatchInfo{
		Weight:  primitives.WeightFromParts(0, 0),
		Class:   primitives.NewDispatchClassOperational(),
		PaysFee: primitives.PaysYes,
	}
	postInfo := primitives.PostDispatchInfo{
		ActualWeight: sc.NewOption[types.Weight](primitives.WeightFromParts(0, 0)),
		PaysFee:      0,
	}
	mockNextFeeMultiplier.On("Get").Return(sc.NewU128(0), nil)

	result, err := target.ComputeActualFee(0, info, postInfo, sc.NewU128(0))
	assert.Nil(t, err)

	mockNextFeeMultiplier.AssertCalled(t, "Get")
	assert.Equal(t, sc.NewU128(100), result)
}

func Test_OperationalFeeMultiplier(t *testing.T) {
	setup()

	result := target.OperationalFeeMultiplier()

	assert.Equal(t, operationalFeeMultiplier, result)
}
