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
	callSetBalanceArgsBytes = sc.NewVaryingData(primitives.MultiAddress{}, sc.Compact{Number: sc.U128{}}).Bytes()
)

func setupCallForceSetBalance() primitives.Call {
	mockBalances = new(MockModule)
	mockStoredMap = new(mocks.StoredMap)
	mockTotalIssuance = new(mocks.StorageValue[sc.U128])

	return newCallForceSetBalance(moduleId, functionForceSetBalance, mockBalances)
}

func Test_Call_SetBalance_new(t *testing.T) {
	target := setupCallForceSetBalance()

	expected := callForceSetBalance{
		Callable: primitives.Callable{
			ModuleId:   moduleId,
			FunctionId: functionForceSetBalance,
			Arguments:  sc.NewVaryingData(primitives.MultiAddress{}, sc.Compact{Number: sc.U128{}}),
		},
		module: mockBalances,
	}

	assert.Equal(t, expected, target)
}

func Test_Call_SetBalance_DecodeArgs(t *testing.T) {
	freeAmount := sc.ToCompact(sc.NewU128(1))

	buf := &bytes.Buffer{}
	buf.Write(targetMultiAddress.Bytes())
	buf.Write(freeAmount.Bytes())

	target := setupCallForceSetBalance()

	call, err := target.DecodeArgs(buf)

	assert.Nil(t, err)
	assert.Equal(t, sc.NewVaryingData(targetMultiAddress, freeAmount), call.Args())
}

func Test_Call_SetBalance_Encode(t *testing.T) {
	target := setupCallForceSetBalance()

	expectedBuffer := bytes.NewBuffer(append([]byte{byte(moduleId), byte(functionForceSetBalance)}, callSetBalanceArgsBytes...))
	buf := &bytes.Buffer{}

	err := target.Encode(buf)

	assert.NoError(t, err)
	assert.Equal(t, expectedBuffer, buf)
}

func Test_Call_SetBalance_Bytes(t *testing.T) {
	expected := append([]byte{byte(moduleId), byte(functionForceSetBalance)}, callSetBalanceArgsBytes...)

	target := setupCallForceSetBalance()

	assert.Equal(t, expected, target.Bytes())
}

func Test_Call_SetBalance_ModuleIndex(t *testing.T) {
	target := setupCallForceSetBalance()

	assert.Equal(t, sc.U8(moduleId), target.ModuleIndex())
}

func Test_Call_SetBalance_FunctionIndex(t *testing.T) {
	target := setupCallForceSetBalance()

	assert.Equal(t, sc.U8(functionForceSetBalance), target.FunctionIndex())
}

func Test_Call_SetBalance_BaseWeight(t *testing.T) {
	target := setupCallForceSetBalance()

	mockBalances.On("DbWeight").Return(dbWeight)

	result := target.BaseWeight()

	assert.Equal(t, callForceSetBalanceCreatingWeight(dbWeight).Max(callForceSetBalanceKillingWeight(dbWeight)), result)

	mockBalances.AssertCalled(t, "DbWeight")
}

func Test_Call_SetBalance_WeighData(t *testing.T) {
	target := setupCallForceSetBalance()

	assert.Equal(t, primitives.WeightFromParts(124, 0), target.WeighData(baseWeight))
}

func Test_Call_SetBalance_ClassifyDispatch(t *testing.T) {
	target := setupCallForceSetBalance()

	assert.Equal(t, primitives.NewDispatchClassNormal(), target.ClassifyDispatch(baseWeight))
}

func Test_Call_SetBalance_PaysFee(t *testing.T) {
	target := setupCallForceSetBalance()

	assert.Equal(t, primitives.PaysYes, target.PaysFee(baseWeight))
}

func Test_Call_SetBalance_Dispatch_Success(t *testing.T) {
	target := setupCallForceSetBalance()

	newFree := sc.NewU128(0)

	mockBalances.On("ExistentialDeposit").Return(sc.NewU128(1))
	mockBalances.On("MutateAccountHandlingDust", targetAddress, mockTypeMutateAccountDataBool).Return(primitives.Balance{}, nil)
	mockBalances.On("DepositEvent", newEventBalanceSet(moduleId, targetAddress, newFree)).Return()

	_, err := target.Dispatch(primitives.NewRawOriginRoot(), sc.NewVaryingData(targetMultiAddress, sc.ToCompact(newFree)))

	assert.NoError(t, err)

	mockBalances.AssertCalled(t, "ExistentialDeposit")
	mockBalances.AssertCalled(t, "MutateAccountHandlingDust", targetAddress, mockTypeMutateAccountDataBool)
	mockTotalIssuance.AssertNotCalled(t, "Get")
	mockTotalIssuance.AssertNotCalled(t, "Put", mock.Anything)
	mockBalances.AssertCalled(t, "DepositEvent", newEventBalanceSet(moduleId, targetAddress, newFree))
}

func Test_Call_SetBalance_Dispatch_BadOrigin(t *testing.T) {
	target := setupCallForceSetBalance()

	_, err := target.Dispatch(primitives.NewRawOriginNone(), sc.NewVaryingData(targetMultiAddress, sc.ToCompact(newFree)))

	assert.Equal(t, primitives.NewDispatchErrorBadOrigin(), err)

	mockBalances.AssertNotCalled(t, "ExistentialDeposit")
	mockBalances.AssertNotCalled(t, "MutateAccountHandlingDust", targetAddress, mockTypeMutateAccountDataBool)
	mockTotalIssuance.AssertNotCalled(t, "Get")
	mockTotalIssuance.AssertNotCalled(t, "Put", mock.Anything)
	mockBalances.AssertNotCalled(t, "DepositEvent", newEventBalanceSet(moduleId, targetAddress, newFree))
}

func Test_Call_SetBalance_Dispatch_CannotLookup(t *testing.T) {
	target := setupCallForceSetBalance()

	_, err := target.Dispatch(
		primitives.NewRawOriginRoot(),
		sc.NewVaryingData(primitives.NewMultiAddress20(primitives.Address20{}), sc.ToCompact(newFree)),
	)

	assert.Equal(t, primitives.NewDispatchErrorCannotLookup(), err)
	mockBalances.AssertNotCalled(t, "ExistentialDeposit")
	mockBalances.AssertNotCalled(t, "MutateAccountHandlingDust", targetAddress, mockTypeMutateAccountDataBool)
	mockTotalIssuance.AssertNotCalled(t, "Get")
	mockTotalIssuance.AssertNotCalled(t, "Put", mock.Anything)
	mockBalances.AssertNotCalled(t, "DepositEvent", newEventBalanceSet(moduleId, targetAddress, newFree))
}

func Test_Call_SetBalance_Dispatch_InvalidArg_Free_InvalidCompact(t *testing.T) {
	target := setupCallForceSetBalance()

	_, err := target.Dispatch(
		primitives.NewRawOriginRoot(),
		sc.NewVaryingData(targetMultiAddress, sc.NewU128(0)),
	)

	assert.Equal(t, primitives.NewDispatchErrorOther("invalid compact value in callForceSetBalance"), err)

	mockBalances.AssertNotCalled(t, "ExistentialDeposit")
	mockBalances.AssertNotCalled(t, "MutateAccountHandlingDust", targetAddress, mockTypeMutateAccountDataBool)
	mockTotalIssuance.AssertNotCalled(t, "Get")
	mockTotalIssuance.AssertNotCalled(t, "Put", mock.Anything)
	mockBalances.AssertNotCalled(t, "DepositEvent", newEventBalanceSet(moduleId, targetAddress, newFree))
}

func Test_Call_SetBalance_Dispatch_InvalidArg_Free_InvalidCompactNumber(t *testing.T) {
	target := setupCallForceSetBalance()

	_, err := target.Dispatch(
		primitives.NewRawOriginRoot(),
		sc.NewVaryingData(targetMultiAddress, sc.Compact{}),
	)

	assert.Equal(t, primitives.NewDispatchErrorOther("invalid U128 value in callForceSetBalance"), err)

	mockBalances.AssertNotCalled(t, "ExistentialDeposit")
	mockBalances.AssertNotCalled(t, "MutateAccountHandlingDust", targetAddress, mockTypeMutateAccountDataBool)
	mockTotalIssuance.AssertNotCalled(t, "Get")
	mockTotalIssuance.AssertNotCalled(t, "Put", mock.Anything)
	mockBalances.AssertNotCalled(t, "DepositEvent", newEventBalanceSet(moduleId, targetAddress, newFree))
}

func Test_Call_SetBalance_setBalance_Success(t *testing.T) {
	target, ok := setupCallForceSetBalance().(callForceSetBalance)
	assert.True(t, ok)

	mockBalances.On("ExistentialDeposit").Return(sc.NewU128(1))
	mockBalances.On("MutateAccountHandlingDust", targetAddress, mockTypeMutateAccountDataBool).Return(primitives.Balance{}, nil)
	mockBalances.On("TotalIssuance").Return(mockTotalIssuance)
	mockTotalIssuance.On("Get").Return(sc.NewU128(1), nil)
	mockTotalIssuance.On("Put", sc.NewU128(6)).Return().Once()
	mockBalances.On("DepositEvent", newEventBalanceSet(moduleId, targetAddress, newFree)).Return()

	result := target.setBalance(targetAddress, newFree)

	assert.Nil(t, result)
	mockBalances.AssertCalled(t, "ExistentialDeposit")
	mockBalances.AssertCalled(t, "MutateAccountHandlingDust", targetAddress, mockTypeMutateAccountDataBool)
	mockBalances.AssertCalled(t, "TotalIssuance")
	mockTotalIssuance.AssertCalled(t, "Get")
	mockTotalIssuance.AssertCalled(t, "Put", sc.NewU128(6))
	mockBalances.AssertCalled(t, "DepositEvent", newEventBalanceSet(moduleId, targetAddress, newFree))
}

func Test_Call_SetBalance_setBalance_Success_LessThanExistentialDeposit(t *testing.T) {
	target, ok := setupCallForceSetBalance().(callForceSetBalance)
	assert.True(t, ok)

	newFree := sc.NewU128(0)
	mockBalances.On("ExistentialDeposit").Return(sc.NewU128(1))
	mockBalances.On("MutateAccountHandlingDust", targetAddress, mockTypeMutateAccountDataBool).Return(primitives.Balance(sc.NewU128(1)), nil)
	mockBalances.On("TotalIssuance").Return(mockTotalIssuance)
	mockTotalIssuance.On("Get").Return(sc.NewU128(1), nil)
	mockTotalIssuance.On("Put", sc.NewU128(0)).Return().Once()
	mockBalances.On("DepositEvent", newEventBalanceSet(moduleId, targetAddress, newFree)).Return()

	result := target.setBalance(targetAddress, newFree)

	assert.Nil(t, result)
	mockBalances.AssertCalled(t, "ExistentialDeposit")
	mockBalances.AssertCalled(t, "MutateAccountHandlingDust", targetAddress, mockTypeMutateAccountDataBool)
	mockBalances.AssertCalled(t, "TotalIssuance")
	mockTotalIssuance.AssertCalled(t, "Get")
	mockTotalIssuance.AssertCalled(t, "Put", sc.NewU128(0))
	mockBalances.AssertCalled(t, "DepositEvent", newEventBalanceSet(moduleId, targetAddress, newFree))
}

func Test_Call_SetBalance_setBalance_tryMutateAccount_Fails(t *testing.T) {
	target, ok := setupCallForceSetBalance().(callForceSetBalance)
	assert.True(t, ok)

	expectedErr := errors.New("some MutateAccountHandlingDust error")
	mockBalances.On("ExistentialDeposit").Return(sc.NewU128(1))
	mockBalances.On("MutateAccountHandlingDust", targetAddress, mockTypeMutateAccountDataBool).Return(primitives.Balance{}, expectedErr)

	result := target.setBalance(targetAddress, targetValue)

	assert.Equal(t, expectedErr, result)
	mockBalances.AssertCalled(t, "ExistentialDeposit")
	mockBalances.AssertCalled(t, "MutateAccountHandlingDust", targetAddress, mockTypeMutateAccountDataBool)
	mockBalances.AssertNotCalled(t, "TotalIssuance")
	mockTotalIssuance.AssertNotCalled(t, "Get")
	mockTotalIssuance.AssertNotCalled(t, "Put", mock.Anything)
	mockBalances.AssertNotCalled(t, "DepositEvent", mock.Anything)
}
