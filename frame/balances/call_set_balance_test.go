package balances

import (
	"bytes"
	"errors"
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/mocks"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	newFree     = sc.NewU128(5)
	newReserved = sc.NewU128(6)
	oldFree     = sc.NewU128(4)
	oldReserved = sc.NewU128(3)

	callSetBalanceArgsBytes = sc.NewVaryingData(primitives.MultiAddress{}, sc.Compact{Number: sc.U128{}}, sc.Compact{Number: sc.U128{}}).Bytes()
)

func Test_Call_SetBalance_new(t *testing.T) {
	target := setupCallSetBalance()
	expected := callSetBalance{
		Callable: primitives.Callable{
			ModuleId:   moduleId,
			FunctionId: functionSetBalanceIndex,
			Arguments:  sc.NewVaryingData(primitives.MultiAddress{}, sc.Compact{Number: sc.U128{}}, sc.Compact{Number: sc.U128{}}),
		},
		constants:      testConstants,
		storedMap:      mockStoredMap,
		accountMutator: mockMutator,
		issuance:       mockStorageTotalIssuance,
	}

	assert.Equal(t, expected, target)
}

func Test_Call_SetBalance_DecodeArgs(t *testing.T) {
	freeAmount := sc.ToCompact(sc.NewU128(1))
	reserveAmount := sc.ToCompact(sc.NewU128(1))
	buf := &bytes.Buffer{}
	buf.Write(targetAddress.Bytes())
	buf.Write(freeAmount.Bytes())
	buf.Write(reserveAmount.Bytes())

	target := setupCallSetBalance()
	call, err := target.DecodeArgs(buf)
	assert.Nil(t, err)

	assert.Equal(t, sc.NewVaryingData(targetAddress, freeAmount, reserveAmount), call.Args())
}

func Test_Call_SetBalance_Encode(t *testing.T) {
	target := setupCallSetBalance()
	expectedBuffer := bytes.NewBuffer(append([]byte{moduleId, functionSetBalanceIndex}, callSetBalanceArgsBytes...))
	buf := &bytes.Buffer{}

	err := target.Encode(buf)

	assert.NoError(t, err)
	assert.Equal(t, expectedBuffer, buf)
}

func Test_Call_SetBalance_Bytes(t *testing.T) {
	expected := append([]byte{moduleId, functionSetBalanceIndex}, callSetBalanceArgsBytes...)

	target := setupCallSetBalance()

	assert.Equal(t, expected, target.Bytes())
}

func Test_Call_SetBalance_ModuleIndex(t *testing.T) {
	target := setupCallSetBalance()

	assert.Equal(t, sc.U8(moduleId), target.ModuleIndex())
}

func Test_Call_SetBalance_FunctionIndex(t *testing.T) {
	target := setupCallSetBalance()

	assert.Equal(t, sc.U8(functionSetBalanceIndex), target.FunctionIndex())
}

func Test_Call_SetBalance_BaseWeight(t *testing.T) {
	target := setupCallSetBalance()

	assert.Equal(t, callSetBalanceCreatingWeight(dbWeight).Max(callSetBalanceKillingWeight(dbWeight)), target.BaseWeight())
}

func Test_Call_SetBalance_IsInherent(t *testing.T) {
	target, ok := setupCallSetBalance().(callSetBalance)
	assert.True(t, ok)
	assert.Equal(t, false, target.IsInherent())
}

func Test_Call_SetBalance_WeighData(t *testing.T) {
	target := setupCallSetBalance()
	assert.Equal(t, primitives.WeightFromParts(124, 0), target.WeighData(baseWeight))
}

func Test_Call_SetBalance_ClassifyDispatch(t *testing.T) {
	target := setupCallSetBalance()

	assert.Equal(t, primitives.NewDispatchClassNormal(), target.ClassifyDispatch(baseWeight))
}

func Test_Call_SetBalance_PaysFee(t *testing.T) {
	target := setupCallSetBalance()

	assert.Equal(t, primitives.PaysYes, target.PaysFee(baseWeight))
}

func Test_Call_SetBalance_Dispatch_Success(t *testing.T) {
	newFree := sc.NewU128(0)
	newReserved := sc.NewU128(0)
	target := setupCallSetBalance()

	expectedResult := sc.NewVaryingData(sc.NewU128(0), sc.NewU128(0))

	targetAddressAccId, err := targetAddress.AsAccountId()
	assert.Nil(t, err)

	mockMutator.On(
		"tryMutateAccount",
		targetAddressAccId,
		mockTypeMutateAccountDataBool,
	).
		Return(expectedResult, nil)
	mockStoredMap.On("DepositEvent", newEventBalanceSet(moduleId, targetAddressAccId, newFree, newReserved))

	_, dispatchErr := target.Dispatch(primitives.NewRawOriginRoot(), sc.NewVaryingData(targetAddress, sc.ToCompact(newFree), sc.ToCompact(newReserved)))

	assert.NoError(t, dispatchErr)
	mockStorageTotalIssuance.AssertNotCalled(t, "Get")
	mockStorageTotalIssuance.AssertNotCalled(t, "Put", mock.Anything)
	mockMutator.AssertCalled(t,
		"tryMutateAccount",
		targetAddressAccId,
		mockTypeMutateAccountDataBool,
	)
	mockStoredMap.AssertCalled(t,
		"DepositEvent",
		newEventBalanceSet(moduleId, targetAddressAccId, newFree, newReserved),
	)
}

func Test_Call_SetBalance_Dispatch_BadOrigin(t *testing.T) {
	target := setupCallSetBalance()

	_, dispatchErr := target.Dispatch(
		primitives.NewRawOriginNone(),
		sc.NewVaryingData(targetAddress, sc.ToCompact(newFree), sc.ToCompact(newReserved)))

	assert.Equal(t, primitives.NewDispatchErrorBadOrigin(), dispatchErr)
	mockMutator.AssertNotCalled(t, "tryMutateAccount", mock.Anything, mock.Anything)
	mockStorageTotalIssuance.AssertNotCalled(t, "Get")
	mockStorageTotalIssuance.AssertNotCalled(t, "Put", mock.Anything)
	mockStoredMap.AssertNotCalled(t, "DepositEvent", mock.Anything)
}

func Test_Call_SetBalance_Dispatch_CannotLookup(t *testing.T) {
	target := setupCallSetBalance()

	_, dispatchErr := target.Dispatch(
		primitives.NewRawOriginRoot(),
		sc.NewVaryingData(primitives.NewMultiAddress20(primitives.Address20{}), sc.ToCompact(newFree), sc.ToCompact(newReserved)))

	assert.Equal(t, primitives.NewDispatchErrorCannotLookup(), dispatchErr)
	mockMutator.AssertNotCalled(t, "tryMutateAccount", mock.Anything, mock.Anything)
	mockStorageTotalIssuance.AssertNotCalled(t, "Get")
	mockStorageTotalIssuance.AssertNotCalled(t, "Put", mock.Anything)
	mockStoredMap.AssertNotCalled(t, "DepositEvent", mock.Anything)
}

func Test_Call_SetBalance_Dispatch_InvalidArg_Free_InvalidCompact(t *testing.T) {
	target := setupCallSetBalance()

	_, dispatchErr := target.Dispatch(
		primitives.NewRawOriginRoot(),
		sc.NewVaryingData(primitives.NewMultiAddress20(primitives.Address20{}), sc.NewU128(0), sc.ToCompact(newReserved)))

	assert.Equal(t, errors.New("invalid free compact value when dispatching balance call set"), dispatchErr)

	mockMutator.AssertNotCalled(t, "tryMutateAccount", mock.Anything, mock.Anything)
	mockStorageTotalIssuance.AssertNotCalled(t, "Get")
	mockStorageTotalIssuance.AssertNotCalled(t, "Put", mock.Anything)
	mockStoredMap.AssertNotCalled(t, "DepositEvent", mock.Anything)
}

func Test_Call_SetBalance_Dispatch_InvalidArg_Free_InvalidCompactNumber(t *testing.T) {
	target := setupCallSetBalance()
	_, dispatchErr := target.Dispatch(
		primitives.NewRawOriginRoot(),
		sc.NewVaryingData(primitives.NewMultiAddress20(primitives.Address20{}), sc.Compact{}, sc.ToCompact(newReserved)))

	assert.Equal(t, errors.New("invalid free compact number when dispatching balance call set"), dispatchErr)

	mockMutator.AssertNotCalled(t, "tryMutateAccount", mock.Anything, mock.Anything)
	mockStorageTotalIssuance.AssertNotCalled(t, "Get")
	mockStorageTotalIssuance.AssertNotCalled(t, "Put", mock.Anything)
	mockStoredMap.AssertNotCalled(t, "DepositEvent", mock.Anything)
}

func Test_Call_SetBalance_Dispatch_InvalidArg_Reserved_InvalidCompact(t *testing.T) {
	target := setupCallSetBalance()

	_, dispatchErr := target.Dispatch(
		primitives.NewRawOriginRoot(),
		sc.NewVaryingData(primitives.NewMultiAddress20(primitives.Address20{}), sc.ToCompact(newFree), sc.NewU128(0)))

	assert.Equal(t, errors.New("invalid reserved compact value when dispatching balance call set"), dispatchErr)

	mockMutator.AssertNotCalled(t, "tryMutateAccount", mock.Anything, mock.Anything)
	mockStorageTotalIssuance.AssertNotCalled(t, "Get")
	mockStorageTotalIssuance.AssertNotCalled(t, "Put", mock.Anything)
	mockStoredMap.AssertNotCalled(t, "DepositEvent", mock.Anything)
}

func Test_Call_SetBalance_Dispatch_InvalidArg_Reserved_InvalidCompactNumber(t *testing.T) {
	target := setupCallSetBalance()

	_, dispatchErr := target.Dispatch(
		primitives.NewRawOriginRoot(),
		sc.NewVaryingData(primitives.NewMultiAddress20(primitives.Address20{}), sc.ToCompact(newFree), sc.Compact{}))

	assert.Equal(t, errors.New("invalid reserved compact number when dispatching balance call set"), dispatchErr)

	mockMutator.AssertNotCalled(t, "tryMutateAccount", mock.Anything, mock.Anything)
	mockStorageTotalIssuance.AssertNotCalled(t, "Get")
	mockStorageTotalIssuance.AssertNotCalled(t, "Put", mock.Anything)
	mockStoredMap.AssertNotCalled(t, "DepositEvent", mock.Anything)
}

func Test_Call_SetBalance_setBalance_Success(t *testing.T) {
	target, ok := setupCallSetBalance().(callSetBalance)
	assert.True(t, ok)
	expectedResult := sc.NewVaryingData(oldFree, oldReserved)

	targetAddressAccId, err := targetAddress.AsAccountId()
	assert.Nil(t, err)

	mockMutator.On(
		"tryMutateAccount",
		targetAddressAccId,
		mockTypeMutateAccountDataBool,
	).Return(expectedResult, nil)
	mockStorageTotalIssuance.On("Get").Return(sc.NewU128(1), nil) // positive imbalance
	mockStorageTotalIssuance.On("Put", newFree.Sub(oldFree).Add(sc.NewU128(1))).
		Return().Once() // newFree positive imbalance
	mockStorageTotalIssuance.On("Put", newReserved.Sub(oldReserved).Add(sc.NewU128(1))).
		Return().Once() // newReserved positive imbalance
	mockStoredMap.On(
		"DepositEvent",
		newEventBalanceSet(moduleId, targetAddressAccId, newFree, newReserved))

	result := target.setBalance(primitives.NewRawOriginRoot(), targetAddress, newFree, newReserved)

	assert.Nil(t, result)
	mockMutator.AssertCalled(t,
		"tryMutateAccount",
		targetAddressAccId,
		mockTypeMutateAccountDataBool,
	)
	mockStorageTotalIssuance.AssertNumberOfCalls(t, "Get", 2)
	mockStorageTotalIssuance.AssertNumberOfCalls(t, "Put", 2)
	mockStorageTotalIssuance.AssertCalled(t, "Put", newFree.Sub(oldFree).Add(sc.NewU128(1)))
	mockStorageTotalIssuance.AssertCalled(t, "Put", newReserved.Sub(oldReserved).Add(sc.NewU128(1)))
	mockStoredMap.AssertCalled(t,
		"DepositEvent",
		newEventBalanceSet(moduleId, targetAddressAccId, newFree, newReserved),
	)
}

func Test_Call_SetBalance_setBalance_Success_LessThanExistentialDeposit(t *testing.T) {
	newFree := sc.NewU128(0)
	newReserved := sc.NewU128(0)
	target, ok := setupCallSetBalance().(callSetBalance)
	assert.True(t, ok)
	expectedResult := sc.NewVaryingData(sc.NewU128(0), sc.NewU128(0))

	targetAddressAccId, err := targetAddress.AsAccountId()
	assert.Nil(t, err)

	mockMutator.On(
		"tryMutateAccount",
		targetAddressAccId,
		mockTypeMutateAccountDataBool,
	).Return(expectedResult, nil)
	mockStoredMap.On(
		"DepositEvent",
		newEventBalanceSet(moduleId, targetAddressAccId, newFree, newReserved))

	result := target.setBalance(primitives.NewRawOriginRoot(), targetAddress, newFree, newReserved)

	assert.Nil(t, result)
	mockStorageTotalIssuance.AssertNotCalled(t, "Get")
	mockStorageTotalIssuance.AssertNotCalled(t, "Put", mock.Anything)
	mockMutator.AssertCalled(t,
		"tryMutateAccount",
		targetAddressAccId,
		mockTypeMutateAccountDataBool,
	)
	mockStoredMap.AssertCalled(t,
		"DepositEvent",
		newEventBalanceSet(moduleId, targetAddressAccId, newFree, newReserved),
	)
}

func Test_Call_SetBalance_setBalance_Success_NegativeImbalance(t *testing.T) {
	newFree := sc.NewU128(1)
	newReserved := sc.NewU128(1)
	target, ok := setupCallSetBalance().(callSetBalance)
	assert.True(t, ok)
	expectedResult := sc.NewVaryingData(oldFree, oldReserved)

	targetAddressAccId, err := targetAddress.AsAccountId()
	assert.Nil(t, err)

	mockMutator.On("tryMutateAccount",
		targetAddressAccId,
		mockTypeMutateAccountDataBool,
	).Return(expectedResult, nil)
	mockStorageTotalIssuance.On("Get").Return(oldReserved.Add(oldFree), nil).Once() // newFree negative imbalance
	mockStorageTotalIssuance.On("Put", oldFree).Return().Once()                     // newFree negative imbalance
	mockStorageTotalIssuance.On("Get").Return(sc.NewU128(4), nil).Once()            // newReserved negative imbalance
	mockStorageTotalIssuance.On("Put", sc.NewU128(2)).Return().Once()               // newReserved negative imbalance
	mockStoredMap.On("DepositEvent", newEventBalanceSet(moduleId, targetAddressAccId, newFree, newReserved))

	result := target.setBalance(primitives.NewRawOriginRoot(), targetAddress, newFree, newReserved)

	assert.Nil(t, result)
	mockMutator.AssertCalled(t,
		"tryMutateAccount",
		targetAddressAccId,
		mockTypeMutateAccountDataBool,
	)
	mockStorageTotalIssuance.AssertNumberOfCalls(t, "Get", 2)
	mockStorageTotalIssuance.AssertNumberOfCalls(t, "Put", 2)
	mockStorageTotalIssuance.AssertCalled(t, "Put", sc.NewU128(4))
	mockStorageTotalIssuance.AssertCalled(t, "Put", sc.NewU128(2))
	mockStoredMap.AssertCalled(t,
		"DepositEvent",
		newEventBalanceSet(moduleId, targetAddressAccId, newFree, newReserved),
	)
}

func Test_Call_SetBalance_setBalance_InvalidOrigin(t *testing.T) {
	target, ok := setupCallSetBalance().(callSetBalance)
	assert.True(t, ok)

	result := target.setBalance(primitives.NewRawOriginNone(), targetAddress, targetValue, targetValue)

	assert.Equal(t, primitives.NewDispatchErrorBadOrigin(), result)
	mockMutator.AssertNotCalled(t, "tryMutateAccount", mock.Anything, mock.Anything)
	mockStorageTotalIssuance.AssertNotCalled(t, "Get")
	mockStorageTotalIssuance.AssertNotCalled(t, "Put", mock.Anything)
	mockStoredMap.AssertNotCalled(t, "DepositEvent", mock.Anything)
}

func Test_Call_SetBalance_setBalance_Lookup(t *testing.T) {
	target, ok := setupCallSetBalance().(callSetBalance)
	assert.True(t, ok)

	result := target.setBalance(primitives.NewRawOriginRoot(), primitives.NewMultiAddress20(primitives.Address20{}), targetValue, targetValue)

	assert.Equal(t, primitives.NewDispatchErrorCannotLookup(), result)
	mockMutator.AssertNotCalled(t, "tryMutateAccount", mock.Anything, mock.Anything)
	mockStorageTotalIssuance.AssertNotCalled(t, "Get")
	mockStorageTotalIssuance.AssertNotCalled(t, "Put", mock.Anything)
	mockStoredMap.AssertNotCalled(t, "DepositEvent", mock.Anything)
}

func Test_Call_SetBalance_setBalance_tryMutateAccount_Fails(t *testing.T) {
	target, ok := setupCallSetBalance().(callSetBalance)
	assert.True(t, ok)
	expectedErr := primitives.NewDispatchErrorBadOrigin()

	targetAddressAccId, errAccId := targetAddress.AsAccountId()
	assert.Nil(t, errAccId)

	mockMutator.On(
		"tryMutateAccount",
		targetAddressAccId,
		mockTypeMutateAccountDataBool,
	).Return(sc.Empty{}, expectedErr)

	result := target.setBalance(primitives.NewRawOriginRoot(), targetAddress, targetValue, targetValue)

	assert.Equal(t, expectedErr, result)
	mockMutator.AssertCalled(t,
		"tryMutateAccount",
		targetAddressAccId,
		mockTypeMutateAccountDataBool,
	)
	mockStorageTotalIssuance.AssertNotCalled(t, "Get")
	mockStorageTotalIssuance.AssertNotCalled(t, "Put", mock.Anything)
	mockStoredMap.AssertNotCalled(t, "DepositEvent", mock.Anything)
}

func Test_Call_SetBalance_setBalance_Drop(t *testing.T) {
	for _, tt := range []struct {
		name                                       string
		newFree, newReserved, oldFree, oldReserved sc.U128
	}{
		{
			name:    "newFree.Gt(oldFree)",
			newFree: sc.NewU128(1),
			oldFree: sc.NewU128(0),
		},
		{
			name:    "newFree.Lt(oldFree)",
			newFree: sc.NewU128(0),
			oldFree: sc.NewU128(1),
		},
		{
			name:        "newReserved.Gt(oldReserved)",
			newReserved: sc.NewU128(1),
			oldReserved: sc.NewU128(0),
		},
		{
			name:        "newReserved.Lt(oldReserved)",
			newReserved: sc.NewU128(0),
			oldReserved: sc.NewU128(1),
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			target, ok := setupCallSetBalance().(callSetBalance)
			assert.True(t, ok)

			mockMutator.On(
				"tryMutateAccount",
				mock.Anything,
				mock.Anything,
			).Return(sc.NewVaryingData(tt.oldFree, tt.oldReserved), nil)

			mockStorageTotalIssuance.On("Put", mock.Anything)
			mockStoredMap.On("DepositEvent", mock.Anything).Return()
			mockStorageTotalIssuance.On("Get").Return(sc.NewU128(0), errors.New("drop")).Once()

			result := target.setBalance(primitives.NewRawOriginRoot(), targetAddress, tt.newFree, tt.newReserved)

			assert.Equal(t, primitives.NewDispatchErrorOther("drop"), result)
		})
	}
}

func Test_Call_SetBalance_updateAccount(t *testing.T) {
	expectedOldFree := sc.NewU128(1)
	expectedOldReserved := sc.NewU128(2)
	newFree := sc.NewU128(5)
	newReserved := sc.NewU128(6)

	account := &primitives.AccountData{
		Free:     expectedOldFree,
		Reserved: expectedOldReserved,
		Frozen:   sc.NewU128(4),
	}
	expectAccount := &primitives.AccountData{
		Free:     newFree,
		Reserved: newReserved,
		Frozen:   sc.NewU128(4),
	}

	oldFree, oldReserved := updateAccount(account, newFree, newReserved)

	assert.Equal(t, expectedOldFree, oldFree)
	assert.Equal(t, expectedOldReserved, oldReserved)
	assert.Equal(t, expectAccount, account)
}

func setupCallSetBalance() primitives.Call {
	mockStoredMap = new(mocks.StoredMap)
	mockMutator = new(mockAccountMutator)
	mockStorageTotalIssuance = new(mocks.StorageValue[sc.U128])

	return newCallSetBalance(moduleId, functionSetBalanceIndex, mockStoredMap, testConstants, mockMutator, mockStorageTotalIssuance).(callSetBalance)
}
