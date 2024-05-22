package transaction_payment

import (
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants/metadata"
	"github.com/LimeChain/gosemble/frame/transaction_payment/types"
	"github.com/LimeChain/gosemble/hooks"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

type Module interface {
	primitives.Module

	ComputeFee(len sc.U32, info primitives.DispatchInfo, tip primitives.Balance) (primitives.Balance, error)
	ComputeFeeDetails(len sc.U32, info primitives.DispatchInfo, tip primitives.Balance) (types.FeeDetails, error)
	ComputeActualFee(len sc.U32, info primitives.DispatchInfo, postInfo primitives.PostDispatchInfo, tip primitives.Balance) (primitives.Balance, error)
	OperationalFeeMultiplier() sc.U8
}

type module struct {
	primitives.DefaultInherentProvider
	hooks.DefaultDispatchModule
	index       sc.U8
	config      *Config
	constants   *consts
	storage     *storage
	mdGenerator *primitives.MetadataTypeGenerator
}

func New(index sc.U8, config *Config, mdGenerator *primitives.MetadataTypeGenerator) Module {
	return module{
		index:       index,
		config:      config,
		constants:   newConstants(config.OperationalFeeMultiplier),
		storage:     newStorage(config.Storage),
		mdGenerator: mdGenerator,
	}
}

func (m module) GetIndex() sc.U8 {
	return m.index
}

func (m module) name() sc.Str {
	return "TransactionPayment"
}

func (m module) Functions() map[sc.U8]primitives.Call {
	return map[sc.U8]primitives.Call{}
}

func (m module) PreDispatch(_ primitives.Call) (sc.Empty, error) {
	return sc.Empty{}, nil
}

func (m module) ValidateUnsigned(_ primitives.TransactionSource, _ primitives.Call) (primitives.ValidTransaction, error) {
	return primitives.ValidTransaction{}, primitives.NewTransactionValidityError(primitives.NewUnknownTransactionNoUnsignedValidator())
}

func (m module) Metadata() primitives.MetadataModule {
	dataV14 := primitives.MetadataModuleV14{
		Name:    m.name(),
		Storage: m.metadataStorage(),
		Call:    sc.NewOption[sc.Compact](nil),
		CallDef: sc.NewOption[primitives.MetadataDefinitionVariant](nil),
		Event:   sc.NewOption[sc.Compact](sc.ToCompact(metadata.TypesTransactionPaymentEvent)),
		EventDef: sc.NewOption[primitives.MetadataDefinitionVariant](
			primitives.NewMetadataDefinitionVariantStr(
				m.name(),
				sc.Sequence[primitives.MetadataTypeDefinitionField]{
					primitives.NewMetadataTypeDefinitionFieldWithName(metadata.TypesTransactionPaymentEvent, "pallet_transaction_payment::Event<Runtime>"),
				},
				m.index,
				"Events.TransactionPayment"),
		),
		Constants: sc.Sequence[primitives.MetadataModuleConstant]{
			primitives.NewMetadataModuleConstant(
				"OperationalFeeMultiplier",
				sc.ToCompact(metadata.PrimitiveTypesU8),
				sc.BytesToSequenceU8(m.constants.OperationalFeeMultiplier.Bytes()),
				"A fee multiplier for `Operational` extrinsics to compute \"virtual tip\" to boost their  `priority` ",
			),
		},
		Error:    sc.NewOption[sc.Compact](nil),
		ErrorDef: sc.NewOption[primitives.MetadataDefinitionVariant](nil),
		Index:    m.index,
	}

	m.mdGenerator.AppendMetadataTypes(m.metadataTypes())

	return primitives.MetadataModule{
		Version:   primitives.ModuleVersion14,
		ModuleV14: dataV14,
	}
}

func (m module) metadataTypes() sc.Sequence[primitives.MetadataType] {
	return sc.Sequence[primitives.MetadataType]{
		primitives.NewMetadataTypeWithPath(metadata.TypesTransactionPaymentReleases, "Releases", sc.Sequence[sc.Str]{"pallet_transaction_payment", "Releases"}, primitives.NewMetadataTypeDefinitionVariant(
			sc.Sequence[primitives.MetadataDefinitionVariant]{
				primitives.NewMetadataDefinitionVariant(
					"V1Ancient",
					sc.Sequence[primitives.MetadataTypeDefinitionField]{},
					0,
					"Original version of the pallet."),
				primitives.NewMetadataDefinitionVariant(
					"V2",
					sc.Sequence[primitives.MetadataTypeDefinitionField]{},
					1,
					"One that bumps the usage to FixedU128 from FixedI128."),
			})),

		primitives.NewMetadataTypeWithParam(metadata.TypesTransactionPaymentEvent, "pallet_transaction_payment pallet Event", sc.Sequence[sc.Str]{"pallet_transaction_payment", "pallet", "Event"}, primitives.NewMetadataTypeDefinitionVariant(
			sc.Sequence[primitives.MetadataDefinitionVariant]{
				primitives.NewMetadataDefinitionVariant(
					"TransactionFeePaid",
					sc.Sequence[primitives.MetadataTypeDefinitionField]{
						primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesAddress32, "who", "T::AccountId"),
						primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.PrimitiveTypesU128, "actual_fee", "BalanceOf<T>"),
						primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.PrimitiveTypesU128, "tip", "BalanceOf<T>"),
					},
					0,
					"Event.TransactionFeePaid"),
			}), primitives.NewMetadataEmptyTypeParameter("T")),

		primitives.NewMetadataTypeWithParams(metadata.TypesTransactionPaymentRuntimeDispatchInfo, "pallet_transaction_payment types RuntimeDispatchInfo", sc.Sequence[sc.Str]{"pallet_transaction_payment", "types", "RuntimeDispatchInfo"}, primitives.NewMetadataTypeDefinitionComposite(
			sc.Sequence[primitives.MetadataTypeDefinitionField]{
				primitives.NewMetadataTypeDefinitionFieldWithName(metadata.TypesWeight, "Weight"),
				primitives.NewMetadataTypeDefinitionFieldWithName(metadata.TypesDispatchClass, "Class"),
				primitives.NewMetadataTypeDefinitionFieldWithName(metadata.PrimitiveTypesU128, "Balance")}),
			sc.Sequence[primitives.MetadataTypeParameter]{
				primitives.NewMetadataTypeParameter(metadata.PrimitiveTypesU128, "Balance"),
				primitives.NewMetadataTypeParameter(metadata.TypesWeight, "Weight"),
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
}

func (m module) metadataStorage() sc.Option[primitives.MetadataModuleStorage] {
	return sc.NewOption[primitives.MetadataModuleStorage](primitives.MetadataModuleStorage{
		Prefix: m.name(),
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
	})
}

func (m module) ComputeFee(len sc.U32, info primitives.DispatchInfo, tip primitives.Balance) (primitives.Balance, error) {
	fee, err := m.ComputeFeeDetails(len, info, tip)
	return fee.FinalFee(), err
}

func (m module) ComputeFeeDetails(len sc.U32, info primitives.DispatchInfo, tip primitives.Balance) (types.FeeDetails, error) {
	return m.computeFeeRaw(len, info.Weight, tip, info.PaysFee, info.Class)
}

func (m module) ComputeActualFee(len sc.U32, info primitives.DispatchInfo, postInfo primitives.PostDispatchInfo, tip primitives.Balance) (primitives.Balance, error) {
	fee, err := m.computeActualFeeDetails(len, info, postInfo, tip)
	return fee.FinalFee(), err
}

func (m module) computeActualFeeDetails(len sc.U32, info primitives.DispatchInfo, postInfo primitives.PostDispatchInfo, tip primitives.Balance) (types.FeeDetails, error) {
	return m.computeFeeRaw(len, postInfo.CalcActualWeight(&info), tip, postInfo.Pays(&info), info.Class)
}

func (m module) computeFeeRaw(len sc.U32, weight primitives.Weight, tip primitives.Balance, paysFee primitives.Pays, class primitives.DispatchClass) (types.FeeDetails, error) {
	if paysFee == primitives.PaysYes {
		unadjustedWeightFee := m.weightToFee(weight)
		multiplier, err := m.storage.NextFeeMultiplier.Get()
		if err != nil {
			return types.FeeDetails{}, err
		}
		// Storage value is FixedU128, which is different from U128.
		// It implements a decimal fixed point number, which is `1 / VALUE`
		// Example: FixedU128, VALUE is 1_000_000_000_000_000_000.
		// FixedU64, VALUE is 1_000_000_000.
		fixedU128Div := sc.NewU128(uint64(1_000_000_000_000_000_000))
		bnAdjustedWeightFee := multiplier.Mul(unadjustedWeightFee)
		adjustedWeightFee := bnAdjustedWeightFee.Div(fixedU128Div) // TODO: Create FixedU128 type

		dispatchClass, err := m.config.BlockWeights.Get(class)
		if err != nil {
			return types.FeeDetails{}, err
		}

		baseFee := m.weightToFee(dispatchClass.BaseExtrinsic)
		lenFee := m.lengthToFee(len)
		inclusionFee := sc.NewOption[types.InclusionFee](types.NewInclusionFee(baseFee, lenFee, adjustedWeightFee))

		return types.FeeDetails{
			InclusionFee: inclusionFee,
			Tip:          tip,
		}, nil
	}

	return types.FeeDetails{
		InclusionFee: sc.NewOption[types.InclusionFee](nil),
		Tip:          tip,
	}, nil
}

func (m module) lengthToFee(length sc.U32) primitives.Balance {
	return m.config.LengthToFee.WeightToFee(primitives.WeightFromParts(sc.U64(length), 0))
}

func (m module) weightToFee(weight primitives.Weight) primitives.Balance {
	cappedWeight := weight.Min(m.config.BlockWeights.MaxBlock)

	return m.config.WeightToFee.WeightToFee(cappedWeight)
}

func (m module) OperationalFeeMultiplier() sc.U8 {
	return m.constants.OperationalFeeMultiplier
}
