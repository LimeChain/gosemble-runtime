package balances

import (
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants"
	"github.com/LimeChain/gosemble/mocks"
	"github.com/LimeChain/gosemble/primitives/log"
	"github.com/LimeChain/gosemble/primitives/types"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	moduleId = sc.U8(3)
)

var (
	maxLocks           = sc.U32(5)
	maxReserves        = sc.U32(6)
	existentialDeposit = sc.NewU128(1)
	dbWeight           = primitives.RuntimeDbWeight{
		Read:  1,
		Write: 2,
	}
	baseWeight  = primitives.WeightFromParts(124, 123)
	mdGenerator = primitives.NewMetadataTypeGenerator()
	logger      = log.NewLogger()
	accountInfo = primitives.AccountInfo{
		Data: primitives.AccountData{
			Free:     sc.NewU128(4),
			Reserved: primitives.Balance{},
			Frozen:   primitives.Balance{},
			// Flags:    primitives.DefaultExtraFlags,
		},
	}
	fromAddress        = constants.OneAccountId
	toAddress          = constants.TwoAccountId
	targetAddress      = constants.ZeroAccountId
	targetMultiAddress = primitives.NewMultiAddressId(targetAddress)
	targetValue        = sc.NewU128(5)

	newFree     = sc.NewU128(5)
	newReserved = sc.NewU128(6)
	oldFree     = sc.NewU128(4)
	oldReserved = sc.NewU128(3)
)

var (
	expectedErr = primitives.NewDispatchErrorCannotLookup()
)

var (
	mockStorage                   *mocks.IoStorage
	mockStoredMap                 *mocks.StoredMap
	mockTotalIssuance             *mocks.StorageValue[sc.U128]
	mockCall                      = new(mocks.Call)
	mockTypeMutateAccountData     = mock.AnythingOfType("func(*types.AccountData) (goscale.Encodable, error)")
	mockTypeMutateAccountDataBool = mock.AnythingOfType("func(*types.AccountData, bool) (goscale.Encodable, error)")
)

var (
	target module
)

func setupModule() module {
	mockStorage = new(mocks.IoStorage)
	mockStoredMap = new(mocks.StoredMap)
	mockTotalIssuance = new(mocks.StorageValue[sc.U128])

	config := NewConfig(mockStorage, dbWeight, maxLocks, maxReserves, existentialDeposit, mockStoredMap)
	target = New(moduleId, config, mdGenerator, logger).(module)
	target.storage.TotalIssuance = mockTotalIssuance

	return target
}

func Test_Module_GetIndex(t *testing.T) {
	target = setupModule()

	assert.Equal(t, sc.U8(moduleId), target.GetIndex())
}

func Test_Module_Functions(t *testing.T) {
	target = setupModule()

	assert.Equal(t, 8, len(target.Functions()))
}

func Test_Module_PreDispatch(t *testing.T) {
	target = setupModule()

	result, err := target.PreDispatch(mockCall)

	assert.Nil(t, err)
	assert.Equal(t, sc.Empty{}, result)
}

func Test_Module_ValidateUnsigned(t *testing.T) {
	target = setupModule()

	result, err := target.ValidateUnsigned(primitives.TransactionSource{}, mockCall)

	assert.Equal(t, primitives.NewTransactionValidityError(primitives.NewUnknownTransactionNoUnsignedValidator()), err)
	assert.Equal(t, primitives.ValidTransaction{}, result)
}

func Test_Module_DepositIntoExisting_Success(t *testing.T) {
	target = setupModule()

	mockStoredMap.On("Get", fromAddress).Return(accountInfo, nil)
	tryMutateResult := sc.NewVaryingData(sc.NewOption[sc.U128](nil), sc.NewOption[sc.U128](nil), targetValue)
	mockStoredMap.On("TryMutateExists", fromAddress, mockTypeMutateAccountData).Return(tryMutateResult, nil)
	mockStoredMap.On("DepositEvent", newEventUpgraded(moduleId, fromAddress))

	result, err := target.DepositIntoExisting(fromAddress, targetValue)

	assert.Nil(t, err)
	assert.Equal(t, targetValue, result)

	mockStoredMap.AssertCalled(t, "Get", fromAddress)
	mockStoredMap.AssertCalled(t, "TryMutateExists", fromAddress, mockTypeMutateAccountData)
	mockStoredMap.AssertCalled(t, "DepositEvent", newEventUpgraded(moduleId, fromAddress))
}

func Test_Module_DepositIntoExisting_ZeroValue(t *testing.T) {
	target = setupModule()

	result, err := target.DepositIntoExisting(fromAddress, sc.NewU128(0))

	assert.Nil(t, err)
	assert.Equal(t, sc.NewU128(0), result)
	mockStoredMap.AssertNotCalled(t, "TryMutateExists", mock.Anything, mock.Anything)
	mockStoredMap.AssertNotCalled(t, "DepositEvent", mock.Anything)
}

func Test_Module_DepositIntoExisting_TryMutateAccount_Fails(t *testing.T) {
	target = setupModule()

	mockStoredMap.On("Get", fromAddress).Return(accountInfo, nil)
	tryMutateResult := sc.NewVaryingData(sc.NewOption[sc.U128](nil), sc.NewOption[sc.U128](nil), targetValue)
	mockStoredMap.On("TryMutateExists", fromAddress, mockTypeMutateAccountData).Return(tryMutateResult, expectedErr)

	_, errDeposit := target.DepositIntoExisting(fromAddress, targetValue)

	assert.Equal(t, expectedErr, errDeposit)
	mockStoredMap.AssertCalled(t, "TryMutateExists", fromAddress, mockTypeMutateAccountData)
	mockStoredMap.AssertNotCalled(t, "DepositEvent", mock.Anything)
}

func Test_Module_Withdraw_Success(t *testing.T) {
	target = setupModule()

	mockStoredMap.On("Get", fromAddress).Return(accountInfo, nil)
	tryMutateResult := sc.NewVaryingData(sc.NewOption[sc.U128](nil), sc.NewOption[sc.U128](nil), targetValue)
	mockStoredMap.On("TryMutateExists", fromAddress, mockTypeMutateAccountData).Return(tryMutateResult, nil)
	mockStoredMap.On("DepositEvent", newEventUpgraded(moduleId, fromAddress))

	result, err := target.Withdraw(fromAddress, targetValue, sc.U8(primitives.ReasonsFee), primitives.ExistenceRequirementKeepAlive)

	assert.Nil(t, err)
	assert.Equal(t, targetValue, result)

	mockStoredMap.AssertCalled(t, "Get", fromAddress)
	mockStoredMap.AssertCalled(t, "TryMutateExists", fromAddress, mockTypeMutateAccountData)
	mockStoredMap.AssertCalled(t, "DepositEvent", newEventUpgraded(moduleId, fromAddress))
	mockTotalIssuance.AssertNotCalled(t, "Get")
	mockTotalIssuance.AssertNotCalled(t, "Put", mock.Anything)
}

func Test_Module_Withdraw_ZeroValue(t *testing.T) {
	target = setupModule()

	result, err := target.Withdraw(fromAddress, sc.NewU128(0), sc.U8(primitives.ReasonsFee), primitives.ExistenceRequirementKeepAlive)

	assert.Nil(t, err)
	assert.Equal(t, sc.NewU128(0), result)
	mockStoredMap.AssertNotCalled(t, "TryMutateExists", mock.Anything, mock.Anything)
	mockStoredMap.AssertNotCalled(t, "DepositEvent", mock.Anything)
}

func Test_Module_Withdraw_TryMutateAccount_Fails(t *testing.T) {
	target = setupModule()

	mockStoredMap.On("Get", fromAddress).Return(accountInfo, nil)
	tryMutateResult := sc.NewVaryingData(sc.NewOption[sc.U128](nil), sc.NewOption[sc.U128](nil), targetValue)
	mockStoredMap.On("TryMutateExists", fromAddress, mockTypeMutateAccountData).Return(tryMutateResult, expectedErr)

	_, errWithdraw := target.Withdraw(fromAddress, targetValue, sc.U8(primitives.ReasonsFee), primitives.ExistenceRequirementKeepAlive)

	assert.Equal(t, expectedErr, errWithdraw)
	mockStoredMap.AssertCalled(t, "TryMutateExists", fromAddress, mockTypeMutateAccountData)
	mockStoredMap.AssertNotCalled(t, "DepositEvent", mock.Anything)
}

func Test_Module_ensureCanWithdraw_Success(t *testing.T) {
	target = setupModule()

	mockStoredMap.On("Get", fromAddress).Return(accountInfo, nil)

	result := target.ensureCanWithdraw(fromAddress, targetValue, primitives.ReasonsFee, sc.NewU128(5))

	assert.Nil(t, result)
	mockStoredMap.AssertCalled(t, "Get", fromAddress)
}

func Test_Module_ensureCanWithdraw_ZeroAmount(t *testing.T) {
	target = setupModule()

	result := target.ensureCanWithdraw(fromAddress, sc.NewU128(0), primitives.ReasonsFee, sc.NewU128(5))

	assert.Nil(t, result)
	mockStoredMap.AssertNotCalled(t, "Get", fromAddress)
}

func Test_Module_ensureCanWithdraw_LiquidityRestrictions(t *testing.T) {
	target = setupModule()

	expected := primitives.NewDispatchErrorModule(primitives.CustomModuleError{
		Index:   moduleId,
		Err:     sc.U32(ErrorLiquidityRestrictions),
		Message: sc.NewOption[sc.Str](nil),
	})
	frozenAccountInfo := primitives.AccountInfo{
		Data: primitives.AccountData{
			Frozen: sc.NewU128(10),
		},
	}

	mockStoredMap.On("Get", fromAddress).Return(frozenAccountInfo, nil)

	result := target.ensureCanWithdraw(fromAddress, targetValue, primitives.ReasonsFee, sc.NewU128(5))

	assert.Equal(t, expected, result)
	mockStoredMap.AssertCalled(t, "Get", fromAddress)
}

func Test_Module_tryMutateAccount_Success(t *testing.T) {
	target = setupModule()

	mockStoredMap.On("Get", fromAddress).Return(accountInfo, nil)
	tryMutateResult := sc.NewVaryingData(sc.NewOption[sc.U128](nil), sc.NewOption[sc.U128](nil), sc.U128{})
	mockStoredMap.On("TryMutateExists", fromAddress, mockTypeMutateAccountData).Return(tryMutateResult, nil)
	mockStoredMap.On("DepositEvent", newEventUpgraded(moduleId, fromAddress))

	_, err := target.tryMutateAccount(fromAddress, func(who *primitives.AccountData, _ bool) (sc.Encodable, error) { return nil, nil })

	assert.NoError(t, err)
	mockStoredMap.AssertCalled(t, "Get", fromAddress)
	mockStoredMap.AssertCalled(t, "TryMutateExists", fromAddress, mockTypeMutateAccountData)
	mockStoredMap.AssertCalled(t, "DepositEvent", newEventUpgraded(moduleId, fromAddress))
}

func Test_Module_tryMutateAccount_TryMutateAccountWithDust_Fails(t *testing.T) {
	target = setupModule()

	mockStoredMap.On("Get", fromAddress).Return(accountInfo, nil)
	tryMutateResult := sc.NewVaryingData(sc.NewOption[sc.U128](nil), sc.NewOption[sc.U128](nil), targetValue)
	mockStoredMap.On("TryMutateExists", fromAddress, mockTypeMutateAccountData).Return(tryMutateResult, expectedErr)

	_, err := target.tryMutateAccount(fromAddress, func(who *primitives.AccountData, _ bool) (sc.Encodable, error) { return nil, nil })

	assert.Equal(t, expectedErr, err)
	mockStoredMap.AssertCalled(t, "TryMutateExists", fromAddress, mockTypeMutateAccountData)
	mockStoredMap.AssertNotCalled(t, "DepositEvent", mock.Anything)
}

// func Test_Module_tryMutateAccountHandlingDust_Success(t *testing.T) {
// 	target = setupModule()

// 	tryMutateResult := sc.NewVaryingData(sc.NewOption[sc.U128](nil), sc.NewOption[negativeImbalance](nil), sc.NewOption[sc.U128](nil))
// 	expectedResult := sc.NewVaryingData(sc.NewOption[sc.U128](nil), newDustCleaner(moduleId, fromAddress, sc.NewOption[negativeImbalance](nil), mockStoredMap))

// 	mockStoredMap.On("TryMutateExists", fromAddress, mockTypeMutateAccountData).Return(tryMutateResult, nil)

// 	result, err := target.tryMutateAccountHandlingDust(fromAddress, func(who *primitives.AccountData, _ bool) (sc.Encodable, error) { return nil, nil })

// 	assert.NoError(t, err)
// 	assert.Equal(t, expectedResult, result)
// 	mockStoredMap.AssertCalled(t, "TryMutateExists", fromAddress, mockTypeMutateAccountData)
// }

// func Test_Module_tryMutateAccountHandlingDust_Success_Endowed(t *testing.T) {
// 	target = setupModule()

// 	tryMutateResult := sc.NewVaryingData(sc.NewOption[sc.U128](targetValue), sc.NewOption[negativeImbalance](nil), sc.NewOption[sc.U128](targetValue))
// 	expectedResult := sc.NewVaryingData(sc.NewOption[sc.U128](targetValue), newDustCleaner(moduleId, fromAddress, sc.NewOption[negativeImbalance](nil), mockStoredMap))

// 	mockStoredMap.On("TryMutateExists", fromAddress, mockTypeMutateAccountData).Return(tryMutateResult, nil)
// 	mockStoredMap.On("DepositEvent", newEventEndowed(moduleId, fromAddress, targetValue))

// 	result, err := target.tryMutateAccountHandlingDust(fromAddress, func(who *primitives.AccountData, _ bool) (sc.Encodable, error) { return nil, nil })

// 	assert.NoError(t, err)
// 	assert.Equal(t, expectedResult, result)
// 	mockStoredMap.AssertCalled(t, "TryMutateExists", fromAddress, mockTypeMutateAccountData)
// 	mockStoredMap.AssertCalled(t, "DepositEvent", newEventEndowed(moduleId, fromAddress, targetValue))
// }

func Test_Module_tryMutateAccountHandlingDust_TryMutateExists_Fail(t *testing.T) {
	target = setupModule()

	mockStoredMap.On("Get", fromAddress).Return(accountInfo, nil)
	tryMutateResult := sc.NewVaryingData(sc.NewOption[sc.U128](nil), sc.NewOption[sc.U128](nil), targetValue)
	mockStoredMap.On("TryMutateExists", fromAddress, mockTypeMutateAccountData).Return(tryMutateResult, expectedErr)

	_, err := target.tryMutateAccountHandlingDust(fromAddress, func(who *primitives.AccountData, _ bool) (sc.Encodable, error) { return nil, nil })

	assert.Equal(t, expectedErr, err)
	mockStoredMap.AssertCalled(t, "TryMutateExists", fromAddress, mockTypeMutateAccountData)
	mockStoredMap.AssertNotCalled(t, "DepositEvent", mock.Anything)
}

func Test_Module_mutateAccount_Success(t *testing.T) {
	target = setupModule()

	maybeAccount := &primitives.AccountData{}
	expectedResult := sc.NewVaryingData(sc.NewOption[sc.U128](nil), sc.NewOption[sc.U128](nil), sc.U128{})

	mockStoredMap.On("Get", fromAddress).Return(accountInfo, nil)

	result, err := target.mutateAccount(
		fromAddress,
		maybeAccount,
		func(who *primitives.AccountData, _ bool) (sc.Encodable, error) {
			return sc.U128{}, nil
		},
	)

	assert.NoError(t, err)
	assert.Equal(t, expectedResult, result)
	mockStoredMap.AssertCalled(t, "Get", fromAddress)
}

func Test_Module_mutateAccount_f_result(t *testing.T) {
	target = setupModule()

	maybeAccount := &primitives.AccountData{
		Free: sc.NewU128(2),
	}
	expectedErr := primitives.NewDispatchErrorBadOrigin()

	mockStoredMap.On("Get", fromAddress).Return(accountInfo, nil)

	result, err := target.mutateAccount(
		fromAddress,
		maybeAccount,
		func(who *primitives.AccountData, _ bool) (sc.Encodable, error) {
			return nil, expectedErr
		},
	)

	assert.Equal(t, expectedErr, err)
	assert.Equal(t, nil, result)
	mockStoredMap.AssertCalled(t, "Get", fromAddress)
}

func Test_Module_mutateAccount_Success_NotNewAccount(t *testing.T) {
	target = setupModule()

	maybeAccount := &primitives.AccountData{
		Free: sc.NewU128(2),
	}
	expectedResult := sc.NewVaryingData(sc.NewOption[sc.U128](nil), sc.NewOption[sc.U128](nil), sc.U128{})

	mockStoredMap.On("Get", fromAddress).Return(accountInfo, nil)
	mockStoredMap.On("IncProviders", fromAddress).Return(types.IncRefStatus(0), nil)

	result, err := target.mutateAccount(
		fromAddress,
		maybeAccount,
		func(who *primitives.AccountData, _ bool) (sc.Encodable, error) {
			return sc.U128{}, nil
		},
	)

	assert.NoError(t, err)
	assert.Equal(t, expectedResult, result)
	mockStoredMap.AssertCalled(t, "Get", fromAddress)
	mockStoredMap.AssertCalled(t, "IncProviders", fromAddress)
}

func Test_Module_withdraw_Success(t *testing.T) {
	target = setupModule()

	fromAccountData := &primitives.AccountData{
		Free: sc.NewU128(5),
	}

	value := sc.NewU128(3)

	mockStoredMap.On("Get", fromAddress).Return(accountInfo, nil)
	mockStoredMap.On("DepositEvent", newEventWithdraw(moduleId, fromAddress, value))

	result, err := target.withdraw(fromAddress, value, fromAccountData, sc.U8(primitives.ReasonsFee), primitives.ExistenceRequirementKeepAlive)

	assert.NoError(t, err)
	assert.Equal(t, value, result)
	assert.Equal(t, sc.NewU128(2), fromAccountData.Free)

	mockStoredMap.AssertCalled(t, "Get", fromAddress)
	mockStoredMap.AssertCalled(t, "DepositEvent", newEventWithdraw(moduleId, fromAddress, value))
}

func Test_Module_withdraw_InsufficientBalance(t *testing.T) {
	target = setupModule()

	fromAccountData := &primitives.AccountData{
		Free: sc.NewU128(5),
	}
	expectedErr := primitives.NewDispatchErrorModule(primitives.CustomModuleError{
		Index:   moduleId,
		Err:     sc.U32(ErrorInsufficientBalance),
		Message: sc.NewOption[sc.Str](nil),
	})

	_, err := target.withdraw(fromAddress, sc.NewU128(10), fromAccountData, sc.U8(primitives.ReasonsFee), primitives.ExistenceRequirementKeepAlive)

	assert.Equal(t, expectedErr, err)
	assert.Equal(t, sc.NewU128(5), fromAccountData.Free)

	mockStoredMap.AssertNotCalled(t, "Get", mock.Anything)
	mockStoredMap.AssertNotCalled(t, "DepositEvent", mock.Anything)
}

func Test_Module_withdraw_KeepAlive(t *testing.T) {
	target = setupModule()

	fromAccountData := &primitives.AccountData{
		Free: sc.NewU128(5),
	}
	expectedErr := primitives.NewDispatchErrorModule(primitives.CustomModuleError{
		Index:   moduleId,
		Err:     sc.U32(ErrorExpendability),
		Message: sc.NewOption[sc.Str](nil),
	})

	_, err := target.withdraw(fromAddress, targetValue, fromAccountData, sc.U8(primitives.ReasonsFee), primitives.ExistenceRequirementKeepAlive)

	assert.Equal(t, expectedErr, err)
	assert.Equal(t, sc.NewU128(5), fromAccountData.Free)
	mockStoredMap.AssertNotCalled(t, "Get", mock.Anything)
	mockStoredMap.AssertNotCalled(t, "DepositEvent", mock.Anything)
}

func Test_Module_withdraw_CannotWithdraw(t *testing.T) {
	target = setupModule()

	fromAccountData := &primitives.AccountData{
		Free: sc.NewU128(5),
	}

	value := sc.NewU128(3)

	frozenAccountInfo := primitives.AccountInfo{
		Data: primitives.AccountData{
			Frozen: sc.NewU128(10),
		},
	}
	expectedErr := primitives.NewDispatchErrorModule(primitives.CustomModuleError{
		Index:   moduleId,
		Err:     sc.U32(ErrorLiquidityRestrictions),
		Message: sc.NewOption[sc.Str](nil),
	})

	mockStoredMap.On("Get", fromAddress).Return(frozenAccountInfo, nil)

	_, err := target.withdraw(fromAddress, value, fromAccountData, sc.U8(primitives.ReasonsFee), primitives.ExistenceRequirementKeepAlive)

	assert.Equal(t, expectedErr, err)
	assert.Equal(t, sc.NewU128(5), fromAccountData.Free)
	mockStoredMap.AssertCalled(t, "Get", fromAddress)
	mockStoredMap.AssertNotCalled(t, "DepositEvent", mock.Anything)
}

func Test_Module_deposit_Success(t *testing.T) {
	target = setupModule()

	toAccountData := &primitives.AccountData{
		Free: sc.NewU128(1),
	}
	expectedResult := targetValue

	mockStoredMap.On("DepositEvent", newEventDeposit(moduleId, toAddress, targetValue))

	result, err := target.deposit(toAddress, toAccountData, false, targetValue)

	assert.NoError(t, err)
	assert.Equal(t, expectedResult, result)
	assert.Equal(t, sc.NewU128(6), toAccountData.Free)
	mockStoredMap.AssertCalled(t, "DepositEvent", newEventDeposit(moduleId, toAddress, targetValue))
}

func Test_Module_deposit_DeadAccount(t *testing.T) {
	target = setupModule()

	toAccountData := &primitives.AccountData{
		Free: sc.NewU128(1),
	}
	expectedErr := primitives.NewDispatchErrorModule(primitives.CustomModuleError{
		Index:   moduleId,
		Err:     sc.U32(ErrorDeadAccount),
		Message: sc.NewOption[sc.Str](nil),
	})

	_, err := target.deposit(toAddress, toAccountData, true, targetValue)

	assert.Equal(t, expectedErr, err)
	assert.Equal(t, sc.NewU128(1), toAccountData.Free)
	mockStoredMap.AssertNotCalled(t, "DepositEvent", mock.Anything)
}

func Test_Module_deposit_ArithmeticOverflow(t *testing.T) {
	target = setupModule()

	toAccountData := &primitives.AccountData{
		Free: sc.MaxU128(),
	}
	expectedErr := primitives.NewDispatchErrorArithmetic(primitives.NewArithmeticErrorOverflow())

	_, err := target.deposit(toAddress, toAccountData, false, targetValue)

	assert.Equal(t, expectedErr, err)
	assert.Equal(t, sc.MaxU128(), toAccountData.Free)
	mockStoredMap.AssertNotCalled(t, "DepositEvent", mock.Anything)
}

// func Test_Module_updateAccount(t *testing.T) {
// 	expectedOldFree := sc.NewU128(1)
// 	expectedOldReserved := sc.NewU128(2)
// 	newFree := sc.NewU128(5)
// 	newReserved := sc.NewU128(6)

// 	account := &primitives.AccountData{
// 		Free:       expectedOldFree,
// 		Reserved:   expectedOldReserved,
// 		MiscFrozen: sc.NewU128(3),
// 		FeeFrozen:  sc.NewU128(4),
// 	}
// 	expectAccount := &primitives.AccountData{
// 		Free:       newFree,
// 		Reserved:   newReserved,
// 		MiscFrozen: sc.NewU128(3),
// 		FeeFrozen:  sc.NewU128(4),
// 	}

// 	updateAccount(account, newFree)

// 	assert.Equal(t, expectedOldFree, oldFree)
// 	assert.Equal(t, expectedOldReserved, oldReserved)
// 	assert.Equal(t, expectAccount, account)
// }
