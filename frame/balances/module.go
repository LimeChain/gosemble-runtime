package balances

import (
	"reflect"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants"
	"github.com/LimeChain/gosemble/constants/metadata"
	"github.com/LimeChain/gosemble/frame/balances/types"
	"github.com/LimeChain/gosemble/hooks"
	"github.com/LimeChain/gosemble/primitives/log"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

const (
	functionTransferIndex = iota
	functionSetBalanceIndex
	functionForceTransferIndex
	functionTransferKeepAliveIndex
	functionTransferAllIndex
	functionForceFreeIndex
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
}

func New(index sc.U8, config *Config, logger log.RuntimeLogger, mdGenerator *primitives.MetadataTypeGenerator) Module {
	constants := newConstants(config.DbWeight, config.MaxLocks, config.MaxReserves, config.ExistentialDeposit)
	storage := newStorage()

	module := Module{
		Index:       index,
		Config:      config,
		constants:   constants,
		storage:     storage,
		mdGenerator: mdGenerator,
	}
	functions := make(map[sc.U8]primitives.Call)
	functions[functionTransferIndex] = newCallTransfer(index, functionTransferIndex, config.StoredMap, constants, module)
	functions[functionSetBalanceIndex] = newCallSetBalance(index, functionSetBalanceIndex, config.StoredMap, constants, module, storage.TotalIssuance)
	functions[functionForceTransferIndex] = newCallForceTransfer(index, functionForceTransferIndex, config.StoredMap, constants, module)
	functions[functionTransferKeepAliveIndex] = newCallTransferKeepAlive(index, functionTransferKeepAliveIndex, config.StoredMap, constants, module)
	functions[functionTransferAllIndex] = newCallTransferAll(index, functionTransferAllIndex, config.StoredMap, constants, module, logger)
	functions[functionForceFreeIndex] = newCallForceFree(index, functionForceFreeIndex, config.StoredMap, constants, module, logger)

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
		return m.withdraw(who, value, account, reasons, liveness)
	})

	return result.(primitives.Balance), err
}

// ensureCanWithdraw checks that an account can withdraw from their balance given any existing withdraw restrictions.
func (m Module) ensureCanWithdraw(who primitives.AccountId, amount sc.U128, reasons primitives.Reasons, newBalance sc.U128) error {
	if amount.Eq(constants.Zero) {
		return nil
	}

	accountInfo, err := m.Config.StoredMap.Get(who)
	if err != nil {
		return primitives.NewDispatchErrorOther(sc.Str(err.Error()))
	}

	minBalance := accountInfo.Frozen(reasons)
	if minBalance.Gt(newBalance) {
		return primitives.NewDispatchErrorModule(primitives.CustomModuleError{
			Index:   m.Index,
			Err:     sc.U32(ErrorLiquidityRestrictions),
			Message: sc.NewOption[sc.Str](nil),
		})
	}

	return nil
}

// tryMutateAccount mutates an account based on argument `f`. Does not change total issuance.
// Does not do anything if `f` returns an error.
func (m Module) tryMutateAccount(who primitives.AccountId, f func(who *primitives.AccountData, bool bool) (sc.Encodable, error)) (sc.Encodable, error) {
	result, err := m.tryMutateAccountWithDust(who, f)
	if err != nil {
		return result, err
	}

	r := result.(sc.VaryingData)

	dustCleaner := r[1].(dustCleaner)
	dustCleaner.Drop()

	return r[0].(sc.Encodable), nil
}

func (m Module) tryMutateAccountWithDust(who primitives.AccountId, f func(who *primitives.AccountData, _ bool) (sc.Encodable, error)) (sc.Encodable, error) {
	result, err := m.Config.StoredMap.TryMutateExists(
		who,
		func(maybeAccount *primitives.AccountData) (sc.Encodable, error) {
			return m.mutateAccount(maybeAccount, f)
		},
	)
	if err != nil {
		return result, err
	}

	resultValue := result.(sc.VaryingData)
	maybeEndowed := resultValue[0].(sc.Option[primitives.Balance])
	if maybeEndowed.HasValue {
		m.Config.StoredMap.DepositEvent(newEventEndowed(m.Index, who, maybeEndowed.Value))
	}

	maybeDust := resultValue[1].(sc.Option[negativeImbalance])
	dustCleaner := newDustCleaner(m.Index, who, maybeDust, m.Config.StoredMap)

	r := sc.NewVaryingData(resultValue[2], dustCleaner)
	return r, nil
}

func (m Module) mutateAccount(maybeAccount *primitives.AccountData, f func(who *primitives.AccountData, _ bool) (sc.Encodable, error)) (sc.Encodable, error) {
	account := &primitives.AccountData{}
	isNew := true
	if !reflect.DeepEqual(*maybeAccount, primitives.AccountData{}) {
		account = maybeAccount
		isNew = false
	}

	result, err := f(account, isNew)
	if err != nil {
		return result, err
	}

	maybeEndowed := sc.NewOption[primitives.Balance](nil)
	if isNew {
		maybeEndowed = sc.NewOption[primitives.Balance](account.Free)
	}
	maybeAccountWithDust, imbalance := m.postMutation(*account)
	if !maybeAccountWithDust.HasValue {
		maybeAccount = &primitives.AccountData{}
	} else {
		maybeAccount.Free = maybeAccountWithDust.Value.Free
		maybeAccount.MiscFrozen = maybeAccountWithDust.Value.MiscFrozen
		maybeAccount.FeeFrozen = maybeAccountWithDust.Value.FeeFrozen
		maybeAccount.Reserved = maybeAccountWithDust.Value.Reserved
	}

	r := sc.NewVaryingData(maybeEndowed, imbalance, result)

	return r, nil
}

func (m Module) postMutation(new primitives.AccountData) (sc.Option[primitives.AccountData], sc.Option[negativeImbalance]) {
	total := new.Total()

	if total.Lt(m.constants.ExistentialDeposit) {
		if total.Eq(constants.Zero) {
			return sc.NewOption[primitives.AccountData](nil), sc.NewOption[negativeImbalance](nil)
		} else {
			return sc.NewOption[primitives.AccountData](nil), sc.NewOption[negativeImbalance](newNegativeImbalance(total, m.storage.TotalIssuance))
		}
	}

	return sc.NewOption[primitives.AccountData](new), sc.NewOption[negativeImbalance](nil)
}

func (m Module) withdraw(who primitives.AccountId, value sc.U128, account *primitives.AccountData, reasons sc.U8, liveness primitives.ExistenceRequirement) (sc.Encodable, error) {
	newFreeAccount, err := sc.CheckedSubU128(account.Free, value)
	if err != nil {
		return nil, primitives.NewDispatchErrorModule(primitives.CustomModuleError{
			Index:   m.Index,
			Err:     sc.U32(ErrorInsufficientBalance),
			Message: sc.NewOption[sc.Str](nil),
		})
	}

	existentialDeposit := m.constants.ExistentialDeposit

	wouldBeDead := (newFreeAccount.Add(account.Reserved)).Lt(existentialDeposit)
	wouldKill := wouldBeDead && ((account.Free.Add(account.Reserved)).Gte(existentialDeposit))

	if !(liveness == primitives.ExistenceRequirementAllowDeath || !wouldKill) {
		return nil, primitives.NewDispatchErrorModule(primitives.CustomModuleError{
			Index:   m.Index,
			Err:     sc.U32(ErrorKeepAlive),
			Message: sc.NewOption[sc.Str](nil),
		})
	}

	if err := m.ensureCanWithdraw(who, value, primitives.Reasons(reasons), newFreeAccount); err != nil {
		return nil, err
	}

	account.Free = newFreeAccount

	m.Config.StoredMap.DepositEvent(newEventWithdraw(m.Index, who, value))
	return value, nil
}

func (m Module) deposit(who primitives.AccountId, account *primitives.AccountData, isNew bool, value sc.U128) (sc.Encodable, error) {
	if isNew {
		return nil, primitives.NewDispatchErrorModule(primitives.CustomModuleError{
			Index:   m.Index,
			Err:     sc.U32(ErrorDeadAccount),
			Message: sc.NewOption[sc.Str](nil),
		})
	}

	free, err := sc.CheckedAddU128(account.Free, value)
	if err != nil {
		return nil, primitives.NewDispatchErrorArithmetic(primitives.NewArithmeticErrorOverflow())
	}
	account.Free = free

	m.Config.StoredMap.DepositEvent(newEventDeposit(m.Index, who, value))

	return value, nil
}

func (m Module) Metadata() primitives.MetadataModule {
	metadataIdBalancesCalls := m.mdGenerator.BuildCallsMetadata("Balances", m.functions, &sc.Sequence[primitives.MetadataTypeParameter]{
		primitives.NewMetadataEmptyTypeParameter("T"),
		primitives.NewMetadataEmptyTypeParameter("I")})

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
				})),

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
				"The total units issued in the system."),
		},
	})
}
