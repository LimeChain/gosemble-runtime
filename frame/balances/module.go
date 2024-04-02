package balances

import (
	"reflect"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants"
	"github.com/LimeChain/gosemble/constants/metadata"
	"github.com/LimeChain/gosemble/frame/balances/types"
	balancestypes "github.com/LimeChain/gosemble/frame/balances/types"
	"github.com/LimeChain/gosemble/hooks"
	"github.com/LimeChain/gosemble/primitives/log"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

const (
	functionTransferAllowDeathIndex = iota
	functionSetBalanceIndex
	functionForceTransferIndex
	functionTransferKeepAliveIndex
	functionTransferAllIndex
	functionForceUnreserveIndex
	functionUpgradeAccountsIndex
	functionForceAdjustTotalIssuanceIndex
)

const (
	name = sc.Str("Balances")
)

type Module struct {
	primitives.DefaultInherentProvider
	hooks.DefaultDispatchModule
	Index       sc.U8
	Config      *Config
	constants   *consts
	storage     *storage
	functions   map[sc.U8]primitives.Call
	mdGenerator *primitives.MetadataTypeGenerator
	logger      log.WarnLogger
}

func New(index sc.U8, config *Config, logger log.WarnLogger, mdGenerator *primitives.MetadataTypeGenerator) Module {
	constants := newConstants(config.DbWeight, config.MaxLocks, config.MaxReserves, config.ExistentialDeposit)
	storage := newStorage()

	module := Module{
		Index:       index,
		Config:      config,
		constants:   constants,
		storage:     storage,
		mdGenerator: mdGenerator,
		logger:      logger,
	}
	functions := make(map[sc.U8]primitives.Call)
	functions[functionTransferAllowDeathIndex] = newCallTransferAllowDeath(functionTransferAllowDeathIndex, module)
	functions[functionSetBalanceIndex] = newCallSetBalance(index, functionSetBalanceIndex, config.StoredMap, constants, module, storage.TotalIssuance)
	functions[functionForceTransferIndex] = newCallForceTransfer(functionForceTransferIndex, module)
	functions[functionTransferKeepAliveIndex] = newCallTransferKeepAlive(functionTransferKeepAliveIndex, module)
	functions[functionTransferAllIndex] = newCallTransferAll(functionTransferAllIndex, module)
	functions[functionForceUnreserveIndex] = newCallForceUnreserve(functionForceUnreserveIndex, module)
	functions[functionUpgradeAccountsIndex] = newCallUpgradeAccounts(functionUpgradeAccountsIndex, module)
	functions[functionForceAdjustTotalIssuanceIndex] = newCallForceAdjustTotalIssuance(functionForceAdjustTotalIssuanceIndex, module)

	module.functions = functions

	return module
}

func (m Module) GetIndex() sc.U8 {
	return m.Index
}

func (m Module) name() sc.Str {
	return name
}

func (m Module) Functions() map[sc.U8]primitives.Call {
	return m.functions
}

func (m Module) PreDispatch(_ primitives.Call) (sc.Empty, error) {
	return sc.Empty{}, nil
}

func (m Module) ValidateUnsigned(_ primitives.TransactionSource, _ primitives.Call) (primitives.ValidTransaction, error) {
	return primitives.ValidTransaction{}, primitives.NewTransactionValidityError(primitives.NewUnknownTransactionNoUnsignedValidator())
}

// DepositIntoExisting deposits `value` into the free balance of an existing target account `who`.
// If `value` is 0, it does nothing.
func (m Module) DepositIntoExisting(who primitives.AccountId, value sc.U128) (primitives.Balance, error) {
	if value.Eq(constants.Zero) {
		return sc.NewU128(0), nil
	}

	result, err := m.tryMutateAccount(
		who,
		func(account *primitives.AccountData, isNew bool) (sc.Encodable, error) {
			return m.deposit(who, account, isNew, value)
		},
	)

	return result.(primitives.Balance), err
}

// Withdraw withdraws `value` free balance from `who`, respecting existence requirements.
// Does not do anything if value is 0.
func (m Module) Withdraw(who primitives.AccountId, value sc.U128, reasons sc.U8, liveness primitives.ExistenceRequirement) (primitives.Balance, error) {
	if value.Eq(constants.Zero) {
		return sc.NewU128(0), nil
	}

	result, err := m.tryMutateAccount(who, func(account *primitives.AccountData, _ bool) (sc.Encodable, error) {
		var preservation balancestypes.Preservation
		if liveness == primitives.ExistenceRequirementAllowDeath {
			preservation = balancestypes.PreservationExpendable
		} else {
			preservation = balancestypes.PreservationPreserve
		}
		return m.withdraw(who, value, account, reasons, preservation, false)
	})

	return result.(primitives.Balance), err
}

// todo delete the following three (+ in mocks)
func (m Module) dbWeight() primitives.RuntimeDbWeight {
	return m.constants.DbWeight
}

func (m Module) moduleIndex() sc.U8 {
	return m.Index
}

func (m Module) totalIssuance() sc.U128 {
	ti, _ := m.storage.TotalIssuance.Get()
	return ti
}

func (m Module) Metadata() primitives.MetadataModule {
	metadataIdBalancesCalls := m.mdGenerator.BuildCallsMetadata("Balances", m.functions, &sc.Sequence[primitives.MetadataTypeParameter]{
		primitives.NewMetadataEmptyTypeParameter("T"),
		primitives.NewMetadataEmptyTypeParameter("I")},
	)

	mdConstants := metadataConstants{
		ExistentialDeposit: primitives.ExistentialDeposit{U128: m.constants.ExistentialDeposit},
		MaxLocks:           primitives.MaxLocks{U32: m.constants.MaxLocks},
		MaxReserves:        primitives.MaxReserves{U32: m.constants.MaxReserves},
	}

	moduleMdConstants := m.mdGenerator.BuildModuleConstants(reflect.ValueOf(mdConstants))

	dataV14 := primitives.MetadataModuleV14{
		Name:    m.name(),
		Storage: m.metadataStorage(),
		Call:    sc.NewOption[sc.Compact](sc.ToCompact(metadataIdBalancesCalls)),
		CallDef: sc.NewOption[primitives.MetadataDefinitionVariant](
			primitives.NewMetadataDefinitionVariantStr(
				m.name(),
				sc.Sequence[primitives.MetadataTypeDefinitionField]{
					primitives.NewMetadataTypeDefinitionFieldWithName(metadataIdBalancesCalls, "self::sp_api_hidden_includes_construct_runtime::hidden_include::dispatch\n::CallableCallFor<Balances, Runtime>"),
				},
				m.Index,
				"Call.Balances"),
		),
		Event: sc.NewOption[sc.Compact](sc.ToCompact(metadata.TypesBalancesEvent)),
		EventDef: sc.NewOption[primitives.MetadataDefinitionVariant](
			primitives.NewMetadataDefinitionVariantStr(
				m.name(),
				sc.Sequence[primitives.MetadataTypeDefinitionField]{
					primitives.NewMetadataTypeDefinitionFieldWithName(metadata.TypesBalancesEvent, "pallet_balances::Event<Runtime>"),
				},
				m.Index,
				"Events.Balances"),
		),
		Constants: moduleMdConstants,
		Error:     sc.NewOption[sc.Compact](sc.ToCompact(metadata.TypesBalancesErrors)),
		ErrorDef: sc.NewOption[primitives.MetadataDefinitionVariant](
			primitives.NewMetadataDefinitionVariantStr(
				m.name(),
				sc.Sequence[primitives.MetadataTypeDefinitionField]{
					primitives.NewMetadataTypeDefinitionField(metadata.TypesBalancesErrors),
				},
				m.Index,
				"Errors.Balances"),
		),
		Index: m.Index,
	}

	m.mdGenerator.AppendMetadataTypes(m.metadataTypes())

	return primitives.MetadataModule{
		Version:   primitives.ModuleVersion14,
		ModuleV14: dataV14,
	}
}

func (m Module) metadataTypes() sc.Sequence[primitives.MetadataType] {
	return sc.Sequence[primitives.MetadataType]{
		primitives.NewMetadataTypeWithPath(metadata.TypesBalancesEvent, "pallet_balances pallet Event", sc.Sequence[sc.Str]{"pallet_balances", "pallet", "Event"}, primitives.NewMetadataTypeDefinitionVariant(
			sc.Sequence[primitives.MetadataDefinitionVariant]{
				primitives.NewMetadataDefinitionVariant(
					"Endowed",
					sc.Sequence[primitives.MetadataTypeDefinitionField]{
						primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesAddress32, "account", "T::AccountId"),
						primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.PrimitiveTypesU128, "free_balance", "T::Balance"),
					},
					EventEndowed,
					"Event.Endowed"),
				primitives.NewMetadataDefinitionVariant(
					"DustLost",
					sc.Sequence[primitives.MetadataTypeDefinitionField]{
						primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesAddress32, "account", "T::AccountId"),
						primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.PrimitiveTypesU128, "amount", "T::Balance"),
					},
					EventDustLost,
					"Events.DustLost"),
				primitives.NewMetadataDefinitionVariant(
					"Transfer",
					sc.Sequence[primitives.MetadataTypeDefinitionField]{
						primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesAddress32, "from", "T::AccountId"),
						primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesAddress32, "to", "T::AccountId"),
						primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.PrimitiveTypesU128, "amount", "T::Balance"),
					},
					EventTransfer,
					"Events.Transfer"),
				primitives.NewMetadataDefinitionVariant(
					"BalanceSet",
					sc.Sequence[primitives.MetadataTypeDefinitionField]{
						primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesAddress32, "who", "T::AccountId"),
						primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.PrimitiveTypesU128, "free", "T::Balance"),
						primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.PrimitiveTypesU128, "reserved", "T::Balance"),
					},
					EventBalanceSet,
					"Events.BalanceSet"),
				primitives.NewMetadataDefinitionVariant(
					"Reserved",
					sc.Sequence[primitives.MetadataTypeDefinitionField]{
						primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesAddress32, "who", "T::AccountId"),
						primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.PrimitiveTypesU128, "amount", "T::Balance"),
					},
					EventReserved,
					"Events.Reserved"),
				primitives.NewMetadataDefinitionVariant(
					"Unreserved",
					sc.Sequence[primitives.MetadataTypeDefinitionField]{
						primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesAddress32, "who", "T::AccountId"),
						primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.PrimitiveTypesU128, "amount", "T::Balance"),
					},
					EventUnreserved,
					"Events.Unreserved"),
				primitives.NewMetadataDefinitionVariant(
					"ReserveRepatriated",
					sc.Sequence[primitives.MetadataTypeDefinitionField]{
						primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesAddress32, "from", "T::AccountId"),
						primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesAddress32, "to", "T::AccountId"),
						primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.PrimitiveTypesU128, "amount", "T::Balance"),
						primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesBalanceStatus, "destination_status", "Status"),
					},
					EventReserveRepatriated,
					"Events.ReserveRepatriated"),
				primitives.NewMetadataDefinitionVariant(
					"Deposit",
					sc.Sequence[primitives.MetadataTypeDefinitionField]{
						primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesAddress32, "who", "T::AccountId"),
						primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.PrimitiveTypesU128, "amount", "T::Balance"),
					},
					EventDeposit,
					"Event.Deposit"),
				primitives.NewMetadataDefinitionVariant(
					"Withdraw",
					sc.Sequence[primitives.MetadataTypeDefinitionField]{
						primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesAddress32, "who", "T::AccountId"),
						primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.PrimitiveTypesU128, "amount", "T::Balance"),
					},
					EventWithdraw,
					"Event.Withdraw"),
				primitives.NewMetadataDefinitionVariant(
					"Slashed",
					sc.Sequence[primitives.MetadataTypeDefinitionField]{
						primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesAddress32, "who", "T::AccountId"),
						primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.PrimitiveTypesU128, "amount", "T::Balance"),
					},
					EventSlashed,
					"Event.Slashed"),
				primitives.NewMetadataDefinitionVariant(
					"Upgraded",
					sc.Sequence[primitives.MetadataTypeDefinitionField]{
						primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesAddress32, "who", "T::AccountId"),
					},
					EventUpgraded,
					"Event.Upgraded"),
				primitives.NewMetadataDefinitionVariant(
					"TotalIssuanceForced",
					sc.Sequence[primitives.MetadataTypeDefinitionField]{
						primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.PrimitiveTypesU128, "old", "T::Balance"),
						primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.PrimitiveTypesU128, "new", "T::Balance"),
					},
					EventTotalIssuanceForced,
					"Event.TotalIssuanceForced"),
			},
		)),
		primitives.NewMetadataTypeWithPath(metadata.TypesBalanceStatus,
			"BalanceStatus",
			sc.Sequence[sc.Str]{"frame_support", "traits", "tokens", "misc", "BalanceStatus"}, primitives.NewMetadataTypeDefinitionVariant(
				sc.Sequence[primitives.MetadataDefinitionVariant]{
					primitives.NewMetadataDefinitionVariant(
						"Free",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{},
						types.BalanceStatusFree,
						"BalanceStatus.Free"),
					primitives.NewMetadataDefinitionVariant(
						"Reserved",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{},
						types.BalanceStatusReserved,
						"BalanceStatus.Reserved"),
				}),
		),
		// primitives.NewMetadataTypeWithPath(metadata.TypesAdjustmentDirection,
		// 	"AdjustmentDirection",
		// 	sc.Sequence[sc.Str]{"pallet_balances", "AdjustmentDirection"}, primitives.NewMetadataTypeDefinitionVariant(
		// 		sc.Sequence[primitives.MetadataDefinitionVariant]{
		// 			primitives.NewMetadataDefinitionVariant(
		// 				"Increase",
		// 				sc.Sequence[primitives.MetadataTypeDefinitionField]{},
		// 				types.AdjustmentDirectionIncrease,
		// 				"AdjustmentDirection.Increase"),
		// 			primitives.NewMetadataDefinitionVariant(
		// 				"Decrease",
		// 				sc.Sequence[primitives.MetadataTypeDefinitionField]{},
		// 				types.AdjustmentDirectionDecrease,
		// 				"AdjustmentDirection.Decrease"),
		// 		}),
		// ),
		primitives.NewMetadataTypeWithPath(metadata.TypesExtraFlags,
			"ExtraFlags",
			sc.Sequence[sc.Str]{"pallet_balances", "ExtraFlags"},
			primitives.NewMetadataTypeDefinitionComposite(
				sc.Sequence[primitives.MetadataTypeDefinitionField]{
					primitives.NewMetadataTypeDefinitionFieldWithName(metadata.PrimitiveTypesU128, "u128"),
				},
			),
		),
		primitives.NewMetadataTypeWithParams(metadata.TypesBalancesErrors,
			"pallet_balances pallet Error",
			sc.Sequence[sc.Str]{"pallet_balances", "pallet", "Error"},
			primitives.NewMetadataTypeDefinitionVariant(
				sc.Sequence[primitives.MetadataDefinitionVariant]{
					primitives.NewMetadataDefinitionVariant(
						"VestingBalance",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{},
						ErrorVestingBalance,
						"Vesting balance too high to send value"),
					primitives.NewMetadataDefinitionVariant(
						"LiquidityRestrictions",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{},
						ErrorLiquidityRestrictions,
						"Account liquidity restrictions prevent withdrawal"),
					primitives.NewMetadataDefinitionVariant(
						"InsufficientBalance",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{},
						ErrorInsufficientBalance,
						"Balance too low to send value."),
					primitives.NewMetadataDefinitionVariant(
						"ExistentialDeposit",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{},
						ErrorExistentialDeposit,
						"Value too low to create account due to existential deposit"),
					primitives.NewMetadataDefinitionVariant(
						"KeepAlive",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{},
						ErrorKeepAlive,
						"Transfer/payment would kill account"),
					primitives.NewMetadataDefinitionVariant(
						"ExistingVestingSchedule",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{},
						ErrorExistingVestingSchedule,
						"A vesting schedule already exists for this account"),
					primitives.NewMetadataDefinitionVariant(
						"DeadAccount",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{},
						ErrorDeadAccount,
						"Beneficiary account must pre-exist"),
					primitives.NewMetadataDefinitionVariant(
						"TooManyReserves",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{},
						ErrorTooManyReserves,
						"Number of named reserves exceed MaxReserves"),
					// todo new errors
				}),
			sc.Sequence[primitives.MetadataTypeParameter]{
				primitives.NewMetadataEmptyTypeParameter("T"),
				primitives.NewMetadataEmptyTypeParameter("I"),
			}),
	}
}

func (m Module) metadataStorage() sc.Option[primitives.MetadataModuleStorage] {
	return sc.NewOption[primitives.MetadataModuleStorage](primitives.MetadataModuleStorage{
		Prefix: m.name(),
		Items: sc.Sequence[primitives.MetadataModuleStorageEntry]{
			primitives.NewMetadataModuleStorageEntry(
				"TotalIssuance",
				primitives.MetadataModuleStorageEntryModifierDefault,
				primitives.NewMetadataModuleStorageEntryDefinitionPlain(sc.ToCompact(metadata.PrimitiveTypesU128)),
				"The total units issued in the system.",
			),
			primitives.NewMetadataModuleStorageEntry(
				"InactiveIssuance",
				primitives.MetadataModuleStorageEntryModifierDefault,
				primitives.NewMetadataModuleStorageEntryDefinitionPlain(sc.ToCompact(metadata.PrimitiveTypesU128)),
				"The total units of outstanding deactivated balance in the system.",
			),
		},
	})
}
