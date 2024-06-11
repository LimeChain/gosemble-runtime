package balances

import (
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants/metadata"
	"github.com/LimeChain/gosemble/frame/balances/types"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"reflect"
)

func (m Module) Metadata() primitives.MetadataModule {
	mdConstants := metadataConstants{
		ExistentialDeposit: primitives.ExistentialDeposit{U128: m.constants.ExistentialDeposit},
		MaxLocks:           primitives.MaxLocks{U32: m.constants.MaxLocks},
		MaxReserves:        primitives.MaxReserves{U32: m.constants.MaxReserves},
	}

	moduleMdConstants := m.mdGenerator.BuildModuleConstants(reflect.ValueOf(mdConstants))

	dataV14 := primitives.MetadataModuleV14{
		Name:    m.name(),
		Storage: m.metadataStorage(),
		Call:    sc.NewOption[sc.Compact](sc.ToCompact(metadata.BalancesCalls)),
		CallDef: sc.NewOption[primitives.MetadataDefinitionVariant](
			primitives.NewMetadataDefinitionVariantStr(
				m.name(),
				sc.Sequence[primitives.MetadataTypeDefinitionField]{
					primitives.NewMetadataTypeDefinitionFieldWithName(metadata.BalancesCalls, "self::sp_api_hidden_includes_construct_runtime::hidden_include::dispatch\n::CallableCallFor<Balances, Runtime>"),
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

func (m Module) metadataStorage() sc.Option[primitives.MetadataModuleStorage] {
	return sc.NewOption[primitives.MetadataModuleStorage](primitives.MetadataModuleStorage{
		Prefix: m.name(),
		Items: sc.Sequence[primitives.MetadataModuleStorageEntry]{
			primitives.NewMetadataModuleStorageEntry(
				"TotalIssuance",
				primitives.MetadataModuleStorageEntryModifierDefault,
				primitives.NewMetadataModuleStorageEntryDefinitionPlain(sc.ToCompact(metadata.PrimitiveTypesU128)),
				"The total units issued in the system."),
			primitives.NewMetadataModuleStorageEntry(
				"InactiveIssuance",
				primitives.MetadataModuleStorageEntryModifierDefault,
				primitives.NewMetadataModuleStorageEntryDefinitionPlain(sc.ToCompact(metadata.PrimitiveTypesU128)),
				"The total units of outstanding deactivated balance in the system."),
		},
	})
}

func (m Module) metadataTypes() sc.Sequence[primitives.MetadataType] {
	return sc.Sequence[primitives.MetadataType]{
		primitives.NewMetadataTypeWithPath(
			metadata.TypesBalancesEvent,
			"pallet_balances pallet Event",
			sc.Sequence[sc.Str]{"pallet_balances", "pallet", "Event"},
			primitives.NewMetadataTypeDefinitionVariant(
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
						"Minted",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesAddress32, "who", "T::AccountId"),
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.PrimitiveTypesU128, "amount", "T::Balance"),
						},
						EventMinted,
						"Event.Minted"),
					primitives.NewMetadataDefinitionVariant(
						"Burned",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesAddress32, "who", "T::AccountId"),
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.PrimitiveTypesU128, "amount", "T::Balance"),
						},
						EventBurned,
						"Event.Burned"),
					primitives.NewMetadataDefinitionVariant(
						"Suspended",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesAddress32, "who", "T::AccountId"),
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.PrimitiveTypesU128, "amount", "T::Balance"),
						},
						EventBurned,
						"Event.Suspended"),
					primitives.NewMetadataDefinitionVariant(
						"Restored",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesAddress32, "who", "T::AccountId"),
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.PrimitiveTypesU128, "amount", "T::Balance"),
						},
						EventRestored,
						"Event.Restored"),
					primitives.NewMetadataDefinitionVariant(
						"Upgraded",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesAddress32, "who", "T::AccountId"),
						},
						EventUpgraded,
						"Event.Upgraded"),
					primitives.NewMetadataDefinitionVariant(
						"Issued",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesAddress32, "who", "T::AccountId"),
						},
						EventIssued,
						"Event.Issued"),
					primitives.NewMetadataDefinitionVariant(
						"Rescinded",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesAddress32, "who", "T::AccountId"),
						},
						EventRescinded,
						"Event.Rescinded"),
					primitives.NewMetadataDefinitionVariant(
						"Locked",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesAddress32, "who", "T::AccountId"),
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.PrimitiveTypesU128, "amount", "T::Balance"),
						},
						EventLocked,
						"Event.Locked"),
					primitives.NewMetadataDefinitionVariant(
						"Unlocked",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesAddress32, "who", "T::AccountId"),
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.PrimitiveTypesU128, "amount", "T::Balance"),
						},
						EventUnlocked,
						"Event.Unlocked"),
					primitives.NewMetadataDefinitionVariant(
						"Frozen",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesAddress32, "who", "T::AccountId"),
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.PrimitiveTypesU128, "amount", "T::Balance"),
						},
						EventFrozen,
						"Event.Frozen"),
					primitives.NewMetadataDefinitionVariant(
						"Thawed",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesAddress32, "who", "T::AccountId"),
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.PrimitiveTypesU128, "amount", "T::Balance"),
						},
						EventThawed,
						"Event.Thawed"),
					primitives.NewMetadataDefinitionVariant(
						"TotalIssuanceForced",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.PrimitiveTypesU128, "old", "T::Balance"),
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.PrimitiveTypesU128, "new", "T::Balance"),
						},
						EventThawed,
						"Event.TotalIssuanceForced"),
				},
			),
		),
		primitives.NewMetadataTypeWithPath(metadata.TypesBalanceStatus,
			"BalanceStatus",
			sc.Sequence[sc.Str]{"frame_support", "traits", "tokens", "misc", "BalanceStatus"},
			primitives.NewMetadataTypeDefinitionVariant(
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
				},
			),
		),

		primitives.NewMetadataTypeWithPath(metadata.TypesBalancesAdjustDirection,
			"AdjustDirection",
			sc.Sequence[sc.Str]{"frame_support", "traits", "tokens", "misc", "AdjustDirection"}, primitives.NewMetadataTypeDefinitionVariant(
				sc.Sequence[primitives.MetadataDefinitionVariant]{
					primitives.NewMetadataDefinitionVariant(
						"Increase",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{},
						types.AdjustDirectionIncrease,
						"AdjustDirection.Increase"),
					primitives.NewMetadataDefinitionVariant(
						"Decrease",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{},
						types.AdjustDirectionDecrease,
						"AdjustDirection.Decrease"),
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
						"Expendability",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{},
						ErrorExpendability,
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
					primitives.NewMetadataDefinitionVariant(
						"TooManyHolds",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{},
						ErrorTooManyHolds,
						"Number of holds exceed `VariantCountOf<T::RuntimeHoldReason>`."),
					primitives.NewMetadataDefinitionVariant(
						"TooManyFreezes",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{},
						ErrorTooManyFreezes,
						"Number of freezes exceed `MaxFreezes`."),
					primitives.NewMetadataDefinitionVariant(
						"IssuanceDeactivated",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{},
						ErrorIssuanceDeactivated,
						"The issuance cannot be modified since it is already deactivated."),
					primitives.NewMetadataDefinitionVariant(
						"DeltaZero",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{},
						ErrorDeltaZero,
						"The delta cannot be zero."),
				}),
			sc.Sequence[primitives.MetadataTypeParameter]{
				primitives.NewMetadataEmptyTypeParameter("T"),
				primitives.NewMetadataEmptyTypeParameter("I"),
			}),

		primitives.NewMetadataTypeWithParam(metadata.BalancesCalls,
			"Balance calls",
			sc.Sequence[sc.Str]{"pallet_balances", "pallet", "Call"},
			primitives.NewMetadataTypeDefinitionVariant(
				sc.Sequence[primitives.MetadataDefinitionVariant]{
					primitives.NewMetadataDefinitionVariant(
						"transfer_allow_death",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesMultiAddress, "dest", "MultiAddress"),
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesCompactU128, "value", "T::Balance"),
						},
						functionTransferAllowDeath,
						"Transfer some liquid free balance to another account."),
					primitives.NewMetadataDefinitionVariant(
						"force_transfer",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesMultiAddress, "source", "MultiAddress"),
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesMultiAddress, "dest", "MultiAddress"),
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesCompactU128, "value", "T::Balance"),
						},
						functionForceTransfer,
						"Exactly as `transfer_allow_death`, except the origin must be root and the source account may be specified."),
					primitives.NewMetadataDefinitionVariant(
						"transfer_keep_alive",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesMultiAddress, "dest", "MultiAddress"),
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesCompactU128, "value", "T::Balance"),
						},
						functionTransferKeepAlive,
						"Same as the [`transfer_allow_death`] call, but with a check that the transfer will not kill the origin account."),
					primitives.NewMetadataDefinitionVariant(
						"transfer_all",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesMultiAddress, "dest", "MultiAddress"),
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.PrimitiveTypesBool, "keep_alive", "bool"),
						},
						functionTransferAll,
						"Transfer the entire transferable balance from the caller account."),
					primitives.NewMetadataDefinitionVariant(
						"force_unreserve",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesMultiAddress, "who", "MultiAddress"),
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.PrimitiveTypesU128, "amount", "T::Balance"),
						},
						functionForceUnreserve,
						"Unreserve some balance from a user by force."),
					primitives.NewMetadataDefinitionVariant(
						"upgrade_accounts",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesSequenceAddress32, "who", "Vec<AccountId>"),
						},
						functionForceUpgradeAccounts,
						"Upgrade a specified account."),
					primitives.NewMetadataDefinitionVariant(
						"force_set_balance",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesMultiAddress, "who", "MultiAddress"),
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesCompactU128, "new_free", "T::Balance"),
						},
						functionForceSetBalance,
						"Set the regular balance of a given account."),
					primitives.NewMetadataDefinitionVariant(
						"force_adjust_total_issuance",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesBalancesAdjustDirection, "direction", "AdjustmentDireciton"),
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesCompactU128, "delta", "T::Balance"),
						},
						functionForceAdjustTotalIssuance,
						"Adjust the total issuance in a saturating way."),
				}),
			primitives.NewMetadataEmptyTypeParameter("T")),
	}
}
