package balances

// func Test_AccountMutator_ensureCanWithdraw_Success(t *testing.T) {
// 	target := setupModule()

// 	fromAddressId, err := fromAddress.AsAccountId()
// 	assert.Nil(t, err)

// 	mockStoredMap.On("Get", fromAddressId).Return(accountInfo, nil)

// 	result := target.ensureCanWithdraw(fromAddressId, targetValue, primitives.ReasonsFee, sc.NewU128(5))

// 	assert.Nil(t, result)
// 	mockStoredMap.AssertCalled(t, "Get", fromAddressId)
// }

// func Test_AccountMutator_ensureCanWithdraw_ZeroAmount(t *testing.T) {
// 	target := setupModule()

// 	fromAddressId, err := fromAddress.AsAccountId()
// 	assert.Nil(t, err)

// 	result := target.ensureCanWithdraw(fromAddressId, sc.NewU128(0), primitives.ReasonsFee, sc.NewU128(5))

// 	assert.Nil(t, result)
// 	mockStoredMap.AssertNotCalled(t, "Get", fromAddressId)
// }

// func Test_AccountMutator_ensureCanWithdraw_LiquidityRestrictions(t *testing.T) {
// 	target := setupModule()
// 	expected := primitives.NewDispatchErrorModule(primitives.CustomModuleError{
// 		Index:   moduleId,
// 		Err:     sc.U32(ErrorLiquidityRestrictions),
// 		Message: sc.NewOption[sc.Str](nil),
// 	})
// 	frozenAccountInfo := primitives.AccountInfo{
// 		Data: primitives.AccountData{
// 			Frozen: sc.NewU128(11),
// 		},
// 	}

// 	fromAddressId, err := fromAddress.AsAccountId()
// 	assert.Nil(t, err)

// 	mockStoredMap.On("Get", fromAddressId).Return(frozenAccountInfo, nil)

// 	result := target.ensureCanWithdraw(fromAddressId, targetValue, primitives.ReasonsFee, sc.NewU128(5))

// 	assert.Equal(t, expected, result)
// 	mockStoredMap.AssertCalled(t, "Get", fromAddressId)
// }

// func Test_AccountMutator_tryMutateAccount_Success(t *testing.T) {
// 	target := setupModule()
// 	mockTotalIssuance := new(mocks.StorageValue[sc.U128])
// 	target.storage.TotalIssuance = mockTotalIssuance

// 	tryMutateResult := sc.NewVaryingData(sc.NewOption[sc.U128](nil), sc.NewOption[negativeImbalance](nil), sc.U128{})

// 	fromAddressId, err := fromAddress.AsAccountId()
// 	assert.Nil(t, err)

// 	mockStoredMap.On("TryMutateExists", fromAddressId, mockTypeMutateAccountData).Return(tryMutateResult, nil)

// 	_, err = target.tryMutateAccount(fromAddressId, func(who *primitives.AccountData, _ bool) (sc.Encodable, error) { return nil, nil })

// 	assert.NoError(t, err)
// 	mockStoredMap.AssertCalled(t, "TryMutateExists", fromAddressId, mockTypeMutateAccountData)
// }

// func Test_AccountMutator_tryMutateAccount_TryMutateAccountWithDust_Fails(t *testing.T) {
// 	target := setupModule()
// 	expectedErr := primitives.NewDispatchErrorCannotLookup()

// 	fromAddressId, err := fromAddress.AsAccountId()
// 	assert.Nil(t, err)

// 	mockStoredMap.On("TryMutateExists", fromAddressId, mockTypeMutateAccountData).Return(sc.NewU128(0), expectedErr)

// 	_, err = target.tryMutateAccount(fromAddressId, func(who *primitives.AccountData, _ bool) (sc.Encodable, error) { return nil, nil })

// 	assert.Equal(t, expectedErr, err)
// 	mockStoredMap.AssertCalled(t, "TryMutateExists", fromAddressId, mockTypeMutateAccountData)
// 	mockStoredMap.AssertNotCalled(t, "DepositEvent", mock.Anything)
// }

// func Test_AccountMutator_tryMutateAccountWithDust_Success(t *testing.T) {
// 	target := setupModule()
// 	mockTotalIssuance := new(mocks.StorageValue[sc.U128])
// 	target.storage.TotalIssuance = mockTotalIssuance

// 	tryMutateResult := sc.NewVaryingData(sc.NewOption[sc.U128](nil), sc.NewOption[negativeImbalance](nil), sc.NewOption[sc.U128](nil))

// 	fromAddressId, err := fromAddress.AsAccountId()
// 	assert.Nil(t, err)

// 	expectedResult := sc.NewVaryingData(sc.NewOption[sc.U128](nil), newDustCleaner(moduleId, fromAddressId, sc.NewOption[negativeImbalance](nil), mockStoredMap))

// 	mockStoredMap.On("TryMutateExists", fromAddressId, mockTypeMutateAccountData).Return(tryMutateResult, nil)

// 	result, err := target.tryMutateAccountWithDust(fromAddressId, func(who *primitives.AccountData, _ bool) (sc.Encodable, error) { return nil, nil })

// 	assert.NoError(t, err)
// 	assert.Equal(t, expectedResult, result)
// 	mockStoredMap.AssertCalled(t, "TryMutateExists", fromAddressId, mockTypeMutateAccountData)
// }

// func Test_AccountMutator_tryMutateAccountWithDust_Success_Endowed(t *testing.T) {
// 	target := setupModule()
// 	mockTotalIssuance := new(mocks.StorageValue[sc.U128])
// 	target.storage.TotalIssuance = mockTotalIssuance

// 	tryMutateResult := sc.NewVaryingData(sc.NewOption[sc.U128](targetValue), sc.NewOption[negativeImbalance](nil), sc.NewOption[sc.U128](targetValue))

// 	fromAddressId, err := fromAddress.AsAccountId()
// 	assert.Nil(t, err)

// 	expectedResult := sc.NewVaryingData(sc.NewOption[sc.U128](targetValue), newDustCleaner(moduleId, fromAddressId, sc.NewOption[negativeImbalance](nil), mockStoredMap))

// 	mockStoredMap.On("TryMutateExists", fromAddressId, mockTypeMutateAccountData).Return(tryMutateResult, nil)
// 	mockStoredMap.On("DepositEvent", newEventEndowed(moduleId, fromAddressId, targetValue))

// 	result, err := target.tryMutateAccountWithDust(fromAddressId, func(who *primitives.AccountData, _ bool) (sc.Encodable, error) { return nil, nil })

// 	assert.NoError(t, err)
// 	assert.Equal(t, expectedResult, result)
// 	mockStoredMap.AssertCalled(t, "TryMutateExists", fromAddressId, mockTypeMutateAccountData)
// 	mockStoredMap.AssertCalled(t, "DepositEvent", newEventEndowed(moduleId, fromAddressId, targetValue))
// }

// func Test_AccountMutator_tryMutateAccountWithDust_TryMutateExists_Fail(t *testing.T) {
// 	target := setupModule()
// 	expectedErr := primitives.NewDispatchErrorCannotLookup()

// 	fromAddressId, err := fromAddress.AsAccountId()
// 	assert.Nil(t, err)

// 	mockStoredMap.On("TryMutateExists", fromAddressId, mockTypeMutateAccountData).Return(sc.NewU128(1), expectedErr)

// 	_, err = target.tryMutateAccountWithDust(fromAddressId, func(who *primitives.AccountData, _ bool) (sc.Encodable, error) { return nil, nil })

// 	assert.Equal(t, expectedErr, err)
// 	mockStoredMap.AssertCalled(t, "TryMutateExists", fromAddressId, mockTypeMutateAccountData)
// 	mockStoredMap.AssertNotCalled(t, "DepositEvent", mock.Anything)
// }

// func Test_AccountMutator_mutateAccount_Success(t *testing.T) {
// 	target := setupModule()
// 	target.storage.TotalIssuance = new(mocks.StorageValue[sc.U128])
// 	maybeAccount := &primitives.AccountData{}
// 	expectedResult := sc.NewVaryingData(sc.NewOption[sc.U128](nil), sc.NewOption[negativeImbalance](nil), sc.U128{})

// 	result, err := target.
// 		mutateAccount(
// 			maybeAccount,
// 			func(who *primitives.AccountData, _ bool) (sc.Encodable, error) {
// 				return sc.U128{}, nil
// 			},
// 		)

// 	assert.NoError(t, err)
// 	assert.Equal(t, expectedResult, result)
// }

// func Test_AccountMutator_mutateAccount_f_result(t *testing.T) {
// 	target := setupModule()
// 	target.storage.TotalIssuance = new(mocks.StorageValue[sc.U128])
// 	maybeAccount := &primitives.AccountData{
// 		Free: sc.NewU128(2),
// 	}
// 	expectedErr := primitives.NewDispatchErrorBadOrigin()

// 	_, err := target.
// 		mutateAccount(
// 			maybeAccount,
// 			func(who *primitives.AccountData, _ bool) (sc.Encodable, error) {
// 				return nil, expectedErr
// 			},
// 		)

// 	assert.Equal(t, expectedErr, err)
// }

// func Test_AccountMutator_mutateAccount_Success_NotNewAccount(t *testing.T) {
// 	target := setupModule()
// 	target.storage.TotalIssuance = new(mocks.StorageValue[sc.U128])
// 	maybeAccount := &primitives.AccountData{
// 		Free: sc.NewU128(2),
// 	}
// 	expectedResult := sc.NewVaryingData(sc.NewOption[sc.U128](nil), sc.NewOption[negativeImbalance](nil), sc.U128{})

// 	result, err := target.
// 		mutateAccount(
// 			maybeAccount,
// 			func(who *primitives.AccountData, _ bool) (sc.Encodable, error) {
// 				return sc.U128{}, nil
// 			},
// 		)

// 	assert.NoError(t, err)
// 	assert.Equal(t, expectedResult, result)
// }

// func Test_AccountMutator_postMutation_Success(t *testing.T) {
// 	target := setupModule()

// 	accOption, imbalance := target.postMutation(*fromAccountData)

// 	assert.Equal(t, sc.NewOption[primitives.AccountData](*fromAccountData), accOption)
// 	assert.Equal(t, sc.NewOption[negativeImbalance](nil), imbalance)
// }

// func Test_AccountMutator_postMutation_ZeroTotal(t *testing.T) {
// 	target := setupModule()

// 	fromAccountData.Free = sc.NewU128(0)

// 	accOption, imbalance := target.postMutation(*fromAccountData)

// 	assert.Equal(t, sc.NewOption[primitives.AccountData](nil), accOption)
// 	assert.Equal(t, sc.NewOption[negativeImbalance](nil), imbalance)
// }

// func Test_AccountMutator_postMutation_LessExistentialDeposit(t *testing.T) {
// 	target := setupModule()
// 	mockTotalIssuance := new(mocks.StorageValue[sc.U128])
// 	target.storage.TotalIssuance = mockTotalIssuance
// 	target.constants.ExistentialDeposit = sc.NewU128(6)

// 	accOption, imbalance := target.postMutation(*fromAccountData)

// 	assert.Equal(t, sc.NewOption[primitives.AccountData](nil), accOption)
// 	assert.Equal(t, sc.NewOption[negativeImbalance](newNegativeImbalance(fromAccountData.Total(), target.storage.TotalIssuance)), imbalance)
// }

// func Test_AccountMutator_withdraw_Success(t *testing.T) {
// 	target := setupModule()
// 	value := sc.NewU128(3)

// 	fromAddressId, err := fromAddress.AsAccountId()
// 	assert.Nil(t, err)

// 	mockStoredMap.On("Get", fromAddressId).Return(accountInfo, nil)
// 	mockStoredMap.On("DepositEvent", newEventWithdraw(moduleId, fromAddressId, value))

// 	result, err := target.withdraw(fromAddressId, value, fromAccountData, sc.U8(primitives.ReasonsFee), balancestypes.PreservationPreserve, false)

// 	assert.NoError(t, err)
// 	assert.Equal(t, value, result)
// 	mockStoredMap.AssertCalled(t, "Get", fromAddressId)
// 	assert.Equal(t, sc.NewU128(2), fromAccountData.Free)
// 	mockStoredMap.AssertCalled(t, "DepositEvent", newEventWithdraw(moduleId, fromAddressId, value))
// }

// func Test_AccountMutator_withdraw_InsufficientBalance(t *testing.T) {
// 	target := setupModule()
// 	expectedErr := primitives.NewDispatchErrorModule(primitives.CustomModuleError{
// 		Index:   moduleId,
// 		Err:     sc.U32(ErrorInsufficientBalance),
// 		Message: sc.NewOption[sc.Str](nil),
// 	})
// 	value := sc.NewU128(10)

// 	fromAddressId, err := fromAddress.AsAccountId()
// 	assert.Nil(t, err)

// 	_, err = target.withdraw(fromAddressId, value, fromAccountData, sc.U8(primitives.ReasonsFee), balancestypes.PreservationPreserve, false)

// 	assert.Equal(t, expectedErr, err)
// 	mockStoredMap.AssertNotCalled(t, "Get", mock.Anything)
// 	assert.Equal(t, sc.NewU128(5), fromAccountData.Free)
// 	mockStoredMap.AssertNotCalled(t, "DepositEvent", mock.Anything)
// }

// func Test_AccountMutator_withdraw_KeepAlive(t *testing.T) {
// 	target := setupModule()
// 	expectedErr := primitives.NewDispatchErrorModule(primitives.CustomModuleError{
// 		Index:   moduleId,
// 		Err:     sc.U32(ErrorKeepAlive),
// 		Message: sc.NewOption[sc.Str](nil),
// 	})

// 	fromAddressId, err := fromAddress.AsAccountId()
// 	assert.Nil(t, err)

// 	_, err = target.withdraw(fromAddressId, targetValue, fromAccountData, sc.U8(primitives.ReasonsFee), balancestypes.PreservationPreserve, false)

// 	assert.Equal(t, expectedErr, err)
// 	mockStoredMap.AssertNotCalled(t, "Get", mock.Anything)
// 	assert.Equal(t, sc.NewU128(5), fromAccountData.Free)
// 	mockStoredMap.AssertNotCalled(t, "DepositEvent", mock.Anything)
// }

// func Test_AccountMutator_withdraw_CannotWithdraw(t *testing.T) {
// 	target := setupModule()
// 	expectedErr := primitives.NewDispatchErrorModule(primitives.CustomModuleError{
// 		Index:   moduleId,
// 		Err:     sc.U32(ErrorLiquidityRestrictions),
// 		Message: sc.NewOption[sc.Str](nil),
// 	})
// 	value := sc.NewU128(3)

// 	frozenAccountInfo := primitives.AccountInfo{
// 		Data: primitives.AccountData{
// 			Frozen: sc.NewU128(11),
// 		},
// 	}

// 	fromAddressId, err := fromAddress.AsAccountId()
// 	assert.Nil(t, err)

// 	mockStoredMap.On("Get", fromAddressId).Return(frozenAccountInfo, nil)

// 	_, err = target.withdraw(fromAddressId, value, fromAccountData, sc.U8(primitives.ReasonsFee), balancestypes.PreservationPreserve, false)

// 	assert.Equal(t, expectedErr, err)
// 	mockStoredMap.AssertCalled(t, "Get", fromAddressId)
// 	assert.Equal(t, sc.NewU128(5), fromAccountData.Free)
// 	mockStoredMap.AssertNotCalled(t, "DepositEvent", mock.Anything)
// }

// func Test_AccountMutator_deposit_Success(t *testing.T) {
// 	target := setupModule()

// 	expectedResult := targetValue
// 	toAddressId, err := toAddress.AsAccountId()
// 	assert.Nil(t, err)

// 	mockStoredMap.On("DepositEvent", newEventDeposit(moduleId, toAddressId, targetValue))

// 	result, err := target.deposit(toAddressId, toAccountData, false, targetValue)

// 	assert.NoError(t, err)
// 	assert.Equal(t, expectedResult, result)
// 	assert.Equal(t, sc.NewU128(6), toAccountData.Free)
// 	mockStoredMap.AssertCalled(t, "DepositEvent", newEventDeposit(moduleId, toAddressId, targetValue))
// }

// func Test_AccountMutator_deposit_DeadAccount(t *testing.T) {
// 	target := setupModule()
// 	expectedErr := primitives.NewDispatchErrorModule(primitives.CustomModuleError{
// 		Index:   moduleId,
// 		Err:     sc.U32(ErrorDeadAccount),
// 		Message: sc.NewOption[sc.Str](nil),
// 	})

// 	toAddressId, err := toAddress.AsAccountId()
// 	assert.Nil(t, err)

// 	_, err = target.deposit(toAddressId, toAccountData, true, targetValue)

// 	assert.Equal(t, expectedErr, err)
// 	assert.Equal(t, sc.NewU128(1), toAccountData.Free)
// 	mockStoredMap.AssertNotCalled(t, "DepositEvent", mock.Anything)
// }

// func Test_AccountMutator_deposit_ArithmeticOverflow(t *testing.T) {
// 	target := setupModule()
// 	expectedErr := primitives.NewDispatchErrorArithmetic(primitives.NewArithmeticErrorOverflow())
// 	toAccountData.Free = sc.MaxU128()

// 	toAddressId, err := toAddress.AsAccountId()
// 	assert.Nil(t, err)

// 	_, err = target.deposit(toAddressId, toAccountData, false, targetValue)

// 	assert.Equal(t, expectedErr, err)
// 	assert.Equal(t, sc.MaxU128(), toAccountData.Free)
// 	mockStoredMap.AssertNotCalled(t, "DepositEvent", mock.Anything)
// }

// func Test_transfer_sanityChecks_InsufficientBalance(t *testing.T) {
// 	target := setupModule()
// 	expectedErr := primitives.NewDispatchErrorModule(primitives.CustomModuleError{
// 		Index:   moduleId,
// 		Err:     sc.U32(ErrorInsufficientBalance),
// 		Message: sc.NewOption[sc.Str](nil),
// 	})

// 	targetAddressId, err := targetAddress.AsAccountId()
// 	assert.Nil(t, err)

// 	err = target.sanityChecks(targetAddressId, fromAccountData, toAccountData, sc.NewU128(6), balancestypes.PreservationPreserve)

// 	assert.Equal(t, expectedErr, err)
// 	assert.Equal(t, sc.NewU128(5), fromAccountData.Free)
// 	assert.Equal(t, sc.NewU128(1), toAccountData.Free)
// 	mockStoredMap.AssertNotCalled(t, "CanDecProviders", mock.Anything)
// }

// func Test_transfer_sanityChecks_ArithmeticOverflow(t *testing.T) {
// 	target := setupModule()
// 	expectedErr := primitives.NewDispatchErrorArithmetic(primitives.NewArithmeticErrorOverflow())
// 	toAccountData.Free = sc.MaxU128()

// 	targetAddressId, err := targetAddress.AsAccountId()
// 	assert.Nil(t, err)

// 	err = target.sanityChecks(targetAddressId, fromAccountData, toAccountData, sc.NewU128(1), balancestypes.PreservationPreserve)

// 	assert.Equal(t, expectedErr, err)
// 	assert.Equal(t, sc.NewU128(4), fromAccountData.Free)
// 	assert.Equal(t, sc.MaxU128(), toAccountData.Free)
// 	mockStoredMap.AssertNotCalled(t, "CanDecProviders", mock.Anything)
// }

// func Test_transfer_sanityChecks_ExistentialDeposit(t *testing.T) {
// 	target := setupModule()
// 	expectedErr := primitives.NewDispatchErrorModule(primitives.CustomModuleError{
// 		Index:   moduleId,
// 		Err:     sc.U32(ErrorExistentialDeposit),
// 		Message: sc.NewOption[sc.Str](nil),
// 	})
// 	toAccountData.Free = sc.NewU128(0)

// 	targetAddressId, err := targetAddress.AsAccountId()
// 	assert.Nil(t, err)

// 	err = target.sanityChecks(targetAddressId, fromAccountData, toAccountData, sc.NewU128(0), balancestypes.PreservationPreserve)

// 	assert.Equal(t, expectedErr, err)
// 	assert.Equal(t, sc.NewU128(5), fromAccountData.Free)
// 	assert.Equal(t, sc.NewU128(0), toAccountData.Free)
// 	mockStoredMap.AssertNotCalled(t, "CanDecProviders", mock.Anything)
// }

// func Test_transfer_sanityChecks_CannotWithdraw(t *testing.T) {
// 	target := setupModule()
// 	expectedErr := primitives.NewDispatchErrorCannotLookup()

// 	targetAddressId, err := targetAddress.AsAccountId()
// 	assert.Nil(t, err)
// 	mockStoredMap.On("Get", targetAddressId).Return(primitives.AccountInfo{}, expectedErr)

// 	err = target.sanityChecks(targetAddressId, fromAccountData, toAccountData, targetValue, balancestypes.PreservationExpendable)

// 	assert.Error(t, err)
// 	assert.Equal(t, sc.NewU128(0), fromAccountData.Free)
// 	assert.Equal(t, sc.NewU128(6), toAccountData.Free)
// }

// func Test_transfer_sanityChecks_KeepAlive(t *testing.T) {
// 	target := setupModule()
// 	expectedErr := primitives.NewDispatchErrorModule(primitives.CustomModuleError{
// 		Index:   moduleId,
// 		Err:     sc.U32(ErrorKeepAlive),
// 		Message: sc.NewOption[sc.Str](nil),
// 	})

// 	targetAddressId, err := targetAddress.AsAccountId()
// 	assert.Nil(t, err)

// 	mockStoredMap.On("Get", targetAddressId).Return(accountInfo, nil)
// 	mockStoredMap.On("CanDecProviders", targetAddressId).Return(false, nil)

// 	err = target.sanityChecks(targetAddressId, fromAccountData, toAccountData, targetValue, balancestypes.PreservationExpendable)

// 	assert.Equal(t, expectedErr, err)
// 	assert.Equal(t, sc.NewU128(0), fromAccountData.Free)
// 	assert.Equal(t, sc.NewU128(6), toAccountData.Free)
// 	mockStoredMap.AssertCalled(t, "CanDecProviders", targetAddressId)
// }

// func Test_transfer_sanityChecks_CanDecProviders_Error(t *testing.T) {
// 	target := setupModule()

// 	targetAddressId, err := targetAddress.AsAccountId()
// 	assert.Nil(t, err)

// 	mockErr := errors.New("err")
// 	expectedErr := primitives.NewDispatchErrorOther(sc.Str(mockErr.Error()))

// 	mockStoredMap.On("Get", targetAddressId).Return(accountInfo, nil)
// 	// mockMutator.On("ensureCanWithdraw", targetAddressId, targetValue, primitives.ReasonsAll, sc.NewU128(0)).Return(nil)
// 	mockStoredMap.On("CanDecProviders", targetAddressId).Return(true, mockErr)

// 	err = target.sanityChecks(targetAddressId, fromAccountData, toAccountData, targetValue, balancestypes.PreservationExpendable)

// 	assert.Equal(t, expectedErr, err)
// 	mockStoredMap.AssertCalled(t, "CanDecProviders", targetAddressId)
// }
