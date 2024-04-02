package balances

import (
	"errors"
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants/metadata"
	"github.com/LimeChain/gosemble/frame/balances/types"
	"github.com/LimeChain/gosemble/mocks"
	"github.com/LimeChain/gosemble/primitives/log"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	mdGenerator = primitives.NewMetadataTypeGenerator()
)

var (
	unknownTransactionNoUnsignedValidator = primitives.NewTransactionValidityError(primitives.NewUnknownTransactionNoUnsignedValidator())
	expectedErr                           = primitives.NewDispatchErrorOther(sc.Str(errors.New("Some unknown error occurred").Error()))
)

var (
	mockTypeMutateAccountData = mock.AnythingOfType("func(*types.AccountData) (goscale.Encodable, error)")
	logger                    = log.NewLogger()
)

func Test_Module_GetIndex(t *testing.T) {
	assert.Equal(t, sc.U8(moduleId), setupModule().GetIndex())
}

func Test_Module_Functions(t *testing.T) {
	target := setupModule()
	functions := target.Functions()

	assert.Equal(t, 8, len(functions))
}

func Test_Module_PreDispatch(t *testing.T) {
	target := setupModule()

	result, err := target.PreDispatch(setupCallTransferAllowDeath())

	assert.Nil(t, err)
	assert.Equal(t, sc.Empty{}, result)
}

func Test_Module_ValidateUnsigned(t *testing.T) {
	target := setupModule()

	result, err := target.ValidateUnsigned(primitives.TransactionSource{}, setupCallTransferAllowDeath())

	assert.Equal(t, unknownTransactionNoUnsignedValidator, err)
	assert.Equal(t, primitives.ValidTransaction{}, result)
}

func Test_Module_DepositIntoExisting_Success(t *testing.T) {
	target := setupModule()
	mockTotalIssuance := new(mocks.StorageValue[sc.U128])
	target.storage.TotalIssuance = mockTotalIssuance

	tryMutateResult := sc.NewVaryingData(sc.NewOption[sc.U128](nil), sc.NewOption[negativeImbalance](nil), targetValue)

	fromAddressId, err := fromAddress.AsAccountId()
	assert.Nil(t, err)

	mockStoredMap.On("TryMutateExists", fromAddressId, mockTypeMutateAccountData).Return(tryMutateResult, nil)

	result, errDeposit := target.DepositIntoExisting(fromAddressId, targetValue)
	assert.Nil(t, errDeposit)

	assert.Equal(t, targetValue, result)
	assert.Nil(t, err)
	mockStoredMap.AssertCalled(t, "TryMutateExists", fromAddressId, mockTypeMutateAccountData)
	mockTotalIssuance.AssertNotCalled(t, "Get")
	mockTotalIssuance.AssertNotCalled(t, "Put", mock.Anything)
}

func Test_Module_DepositIntoExisting_ZeroValue(t *testing.T) {
	target := setupModule()

	fromAddressId, err := fromAddress.AsAccountId()
	assert.Nil(t, err)

	result, errDeposit := target.DepositIntoExisting(fromAddressId, sc.NewU128(0))
	assert.Nil(t, errDeposit)

	assert.Equal(t, sc.NewU128(0), result)
	assert.Nil(t, err)
	mockStoredMap.AssertNotCalled(t, "TryMutateExists", mock.Anything, mock.Anything)
	mockStoredMap.AssertNotCalled(t, "DepositEvent", mock.Anything)
}

func Test_Module_DepositIntoExisting_TryMutateAccount_Fails(t *testing.T) {
	target := setupModule()

	expectedResult := sc.NewU128(1)
	expectedErr := primitives.NewDispatchErrorCannotLookup()

	fromAddressId, err := fromAddress.AsAccountId()
	assert.Nil(t, err)

	mockStoredMap.On("TryMutateExists", fromAddressId, mockTypeMutateAccountData).Return(expectedResult, expectedErr)

	_, errDeposit := target.DepositIntoExisting(fromAddressId, targetValue)

	assert.Equal(t, expectedErr, errDeposit)
	mockStoredMap.AssertCalled(t, "TryMutateExists", fromAddressId, mockTypeMutateAccountData)
	mockStoredMap.AssertNotCalled(t, "DepositEvent", mock.Anything)
}

func Test_Module_Withdraw_Success(t *testing.T) {
	target := setupModule()
	mockTotalIssuance := new(mocks.StorageValue[sc.U128])
	target.storage.TotalIssuance = mockTotalIssuance

	tryMutateResult := sc.NewVaryingData(sc.NewOption[sc.U128](nil), sc.NewOption[negativeImbalance](nil), targetValue)

	fromAddressId, err := fromAddress.AsAccountId()
	assert.Nil(t, err)

	mockStoredMap.On("TryMutateExists", fromAddressId, mockTypeMutateAccountData).Return(tryMutateResult, nil)

	result, errWithdraw := target.Withdraw(fromAddressId, targetValue, sc.U8(primitives.ReasonsFee), primitives.ExistenceRequirementKeepAlive)
	assert.Nil(t, errWithdraw)

	assert.Equal(t, targetValue, result)
	assert.Nil(t, err)
	mockStoredMap.AssertCalled(t, "TryMutateExists", fromAddressId, mockTypeMutateAccountData)
	mockTotalIssuance.AssertNotCalled(t, "Get")
	mockTotalIssuance.AssertNotCalled(t, "Put", mock.Anything)
}

func Test_Module_Withdraw_ZeroValue(t *testing.T) {
	target := setupModule()

	fromAddressId, err := fromAddress.AsAccountId()
	assert.Nil(t, err)

	result, errWithdraw := target.Withdraw(fromAddressId, sc.NewU128(0), sc.U8(primitives.ReasonsFee), primitives.ExistenceRequirementKeepAlive)
	assert.Nil(t, errWithdraw)

	assert.Equal(t, sc.NewU128(0), result)
	assert.Nil(t, err)
	mockStoredMap.AssertNotCalled(t, "TryMutateExists", mock.Anything, mock.Anything)
	mockStoredMap.AssertNotCalled(t, "DepositEvent", mock.Anything)
}

func Test_Module_Withdraw_TryMutateAccount_Fails(t *testing.T) {
	target := setupModule()

	expectedErr := primitives.NewDispatchErrorCannotLookup()

	fromAddressId, err := fromAddress.AsAccountId()
	assert.Nil(t, err)

	mockStoredMap.On("TryMutateExists", fromAddressId, mockTypeMutateAccountData).Return(sc.NewU128(1), expectedErr)

	_, errWithdraw := target.Withdraw(fromAddressId, targetValue, sc.U8(primitives.ReasonsFee), primitives.ExistenceRequirementKeepAlive)

	assert.Equal(t, expectedErr, errWithdraw)
	mockStoredMap.AssertCalled(t, "TryMutateExists", fromAddressId, mockTypeMutateAccountData)
	mockStoredMap.AssertNotCalled(t, "DepositEvent", mock.Anything)
}

func Test_Module_Metadata(t *testing.T) {
	target := setupModule()

	expectedBalancesCallsMetadataId := mdGenerator.GetLastAvailableIndex() + 1

	expectedCompactU128TypeId := expectedBalancesCallsMetadataId + 1

	expectMetadataTypes := sc.Sequence[primitives.MetadataType]{
		primitives.NewMetadataType(expectedCompactU128TypeId, "CompactU128", primitives.NewMetadataTypeDefinitionCompact(sc.ToCompact(metadata.PrimitiveTypesU128))),
		primitives.NewMetadataTypeWithParams(
			expectedBalancesCallsMetadataId,
			"Balances calls",
			sc.Sequence[sc.Str]{"pallet_balances", "pallet", "Call"},
			primitives.NewMetadataTypeDefinitionVariant(
				sc.Sequence[primitives.MetadataDefinitionVariant]{
					primitives.NewMetadataDefinitionVariant(
						"transfer_allow_death",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionField(metadata.TypesMultiAddress),
							primitives.NewMetadataTypeDefinitionField(expectedCompactU128TypeId),
						},
						functionTransferAllowDeathIndex,
						"Transfer some liquid free balance to another account.",
					),
					primitives.NewMetadataDefinitionVariant(
						"set_balance",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionField(metadata.TypesMultiAddress),
							primitives.NewMetadataTypeDefinitionField(expectedCompactU128TypeId),
							primitives.NewMetadataTypeDefinitionField(expectedCompactU128TypeId),
						},
						functionForceSetBalanceIndex,
						"Set the balances of a given account.",
					),
					primitives.NewMetadataDefinitionVariant(
						"force_transfer", // todo
						sc.Sequence[primitives.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionField(metadata.TypesMultiAddress),
							primitives.NewMetadataTypeDefinitionField(metadata.TypesMultiAddress),
							primitives.NewMetadataTypeDefinitionField(expectedCompactU128TypeId),
						},
						functionForceTransferIndex,
						"Exactly as `transfer_allow_death`, except the origin must be root and the source account may be specified.",
					),
					primitives.NewMetadataDefinitionVariant(
						"transfer_keep_alive",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionField(metadata.TypesMultiAddress),
							primitives.NewMetadataTypeDefinitionField(expectedCompactU128TypeId),
						},
						functionTransferKeepAliveIndex,
						"Same as the [`transfer_allow_death`] call, but with a check that the transfer will not kill the origin account.",
					),
					primitives.NewMetadataDefinitionVariant(
						"transfer_all",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionField(metadata.TypesMultiAddress),
							primitives.NewMetadataTypeDefinitionField(metadata.PrimitiveTypesBool),
						},
						functionTransferAllIndex,
						"Transfer the entire transferable balance from the caller account.",
					),
					primitives.NewMetadataDefinitionVariant(
						"force_unreserve",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionField(metadata.TypesMultiAddress),
							primitives.NewMetadataTypeDefinitionField(metadata.PrimitiveTypesU128),
						},
						functionForceUnreserveIndex,
						"Unreserve some balance from a user by force.",
					),

					primitives.NewMetadataDefinitionVariant(
						"upgrade_accounts",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionField(metadata.TypesSequenceSequenceU8),
						},
						functionUpgradeAccountsIndex,
						"Upgrade a specified account.",
					),
					primitives.NewMetadataDefinitionVariant(
						"force_adjust_total_issuance",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionField(metadata.PrimitiveTypesU8),
							primitives.NewMetadataTypeDefinitionField(expectedCompactU128TypeId),
						},
						functionForceAdjustTotalIssuanceIndex,
						"Adjust the total issuance in a saturating way.",
					),
				},
			),
			sc.Sequence[primitives.MetadataTypeParameter]{
				primitives.NewMetadataEmptyTypeParameter("T"),
				primitives.NewMetadataEmptyTypeParameter("I"),
			},
		),
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
				})),
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
				}),
			sc.Sequence[primitives.MetadataTypeParameter]{
				primitives.NewMetadataEmptyTypeParameter("T"),
				primitives.NewMetadataEmptyTypeParameter("I"),
			}),
	}
	moduleV14 := primitives.MetadataModuleV14{
		Name: name,
		Storage: sc.NewOption[primitives.MetadataModuleStorage](primitives.MetadataModuleStorage{
			Prefix: name,
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
		}),
		Call: sc.NewOption[sc.Compact](sc.ToCompact(expectedBalancesCallsMetadataId)),
		CallDef: sc.NewOption[primitives.MetadataDefinitionVariant](
			primitives.NewMetadataDefinitionVariantStr(
				name,
				sc.Sequence[primitives.MetadataTypeDefinitionField]{
					primitives.NewMetadataTypeDefinitionFieldWithName(expectedBalancesCallsMetadataId, "self::sp_api_hidden_includes_construct_runtime::hidden_include::dispatch\n::CallableCallFor<Balances, Runtime>"),
				},
				moduleId,
				"Call.Balances"),
		),
		Event: sc.NewOption[sc.Compact](sc.ToCompact(metadata.TypesBalancesEvent)),
		EventDef: sc.NewOption[primitives.MetadataDefinitionVariant](
			primitives.NewMetadataDefinitionVariantStr(
				name,
				sc.Sequence[primitives.MetadataTypeDefinitionField]{
					primitives.NewMetadataTypeDefinitionFieldWithName(metadata.TypesBalancesEvent, "pallet_balances::Event<Runtime>"),
				},
				moduleId,
				"Events.Balances"),
		),
		Constants: sc.Sequence[primitives.MetadataModuleConstant]{
			primitives.NewMetadataModuleConstant(
				"ExistentialDeposit",
				sc.ToCompact(metadata.PrimitiveTypesU128),
				sc.BytesToSequenceU8(existentialDeposit.Bytes()),
				"The minimum amount required to keep an account open. MUST BE GREATER THAN ZERO!",
			),
			primitives.NewMetadataModuleConstant(
				"MaxLocks",
				sc.ToCompact(metadata.PrimitiveTypesU32),
				sc.BytesToSequenceU8(maxLocks.Bytes()),
				"The maximum number of locks that should exist on an account.  Not strictly enforced, but used for weight estimation.",
			),
			primitives.NewMetadataModuleConstant(
				"MaxReserves",
				sc.ToCompact(metadata.PrimitiveTypesU32),
				sc.BytesToSequenceU8(maxReserves.Bytes()),
				"The maximum number of named reserves that can exist on an account.",
			),
		},
		Error: sc.NewOption[sc.Compact](sc.ToCompact(metadata.TypesBalancesErrors)),
		ErrorDef: sc.NewOption[primitives.MetadataDefinitionVariant](
			primitives.NewMetadataDefinitionVariantStr(
				name,
				sc.Sequence[primitives.MetadataTypeDefinitionField]{
					primitives.NewMetadataTypeDefinitionField(metadata.TypesBalancesErrors),
				},
				moduleId,
				"Errors.Balances"),
		),
		Index: moduleId,
	}

	expectMetadataModule := primitives.MetadataModule{
		Version:   primitives.ModuleVersion14,
		ModuleV14: moduleV14,
	}

	resultMetadataModule := target.Metadata()
	resultTypes := mdGenerator.GetMetadataTypes()

	assert.Equal(t, expectMetadataTypes, resultTypes)
	assert.Equal(t, expectMetadataModule, resultMetadataModule)
}

func setupModule() Module {
	mockStoredMap = new(mocks.StoredMap)
	mockTotalIssuance = new(mocks.StorageValue[sc.U128])
	mockInactiveIssuance = new(mocks.StorageValue[sc.U128])
	config := NewConfig(dbWeight, maxLocks, maxReserves, existentialDeposit, mockStoredMap)
	target := New(moduleId, config, logger, mdGenerator)
	target.storage.TotalIssuance = mockTotalIssuance
	target.storage.InactiveIssuance = mockInactiveIssuance

	fromAccountData = &primitives.AccountData{
		Free: sc.NewU128(5),
	}

	toAccountData = &primitives.AccountData{
		Free: sc.NewU128(1),
	}

	return target
}
