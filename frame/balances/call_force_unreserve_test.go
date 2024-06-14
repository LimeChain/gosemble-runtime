package balances

import (
	"bytes"
	"errors"
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants"
	"github.com/LimeChain/gosemble/mocks"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	argsBytesCallForceFree = sc.NewVaryingData(primitives.MultiAddress{}, sc.U128{}).Bytes()
	errPanic               = errors.New("panic")
)

var (
	mockBalances *MockModule
)

func setupCallForceUnreserve() primitives.Call {
	mockBalances = new(MockModule)
	mockStoredMap = new(mocks.StoredMap)

	return newCallForceUnreserve(moduleId, functionForceUnreserve, mockBalances)
}

func Test_Call_ForceFree_new(t *testing.T) {
	target := setupCallForceUnreserve()

	expected := callForceUnreserve{
		Callable: primitives.Callable{
			ModuleId:   moduleId,
			FunctionId: functionForceUnreserve,
			Arguments:  sc.NewVaryingData(primitives.MultiAddress{}, sc.U128{}),
		},
		module: mockBalances,
	}

	assert.Equal(t, expected, target)
}

func Test_Call_ForceFree_DecodeArgs(t *testing.T) {
	amount := sc.NewU128(5)
	buf := bytes.NewBuffer(append(targetMultiAddress.Bytes(), amount.Bytes()...))

	target := setupCallForceUnreserve()

	call, err := target.DecodeArgs(buf)
	assert.Nil(t, err)

	assert.Equal(t, sc.NewVaryingData(targetMultiAddress, amount), call.Args())
}

func Test_Call_ForceFree_Encode(t *testing.T) {
	target := setupCallForceUnreserve()

	expectedBuffer := bytes.NewBuffer(append([]byte{byte(moduleId), byte(functionForceUnreserve)}, argsBytesCallForceFree...))
	buf := &bytes.Buffer{}

	err := target.Encode(buf)

	assert.NoError(t, err)
	assert.Equal(t, expectedBuffer, buf)
}

func Test_Call_ForceFree_Bytes(t *testing.T) {
	expected := append([]byte{byte(moduleId), byte(functionForceUnreserve)}, argsBytesCallForceFree...)

	target := setupCallForceUnreserve()

	assert.Equal(t, expected, target.Bytes())
}

func Test_Call_ForceFree_ModuleIndex(t *testing.T) {
	target := setupCallForceUnreserve()

	assert.Equal(t, sc.U8(moduleId), target.ModuleIndex())
}

func TestCall_ForceFree_FunctionIndex(t *testing.T) {
	target := setupCallForceUnreserve()

	assert.Equal(t, sc.U8(functionForceUnreserve), target.FunctionIndex())
}

func Test_Call_ForceFree_EncodeWithArgs(t *testing.T) {
	expectedBuffer := bytes.NewBuffer([]byte{byte(moduleId), byte(functionForceUnreserve)})
	bArgs := append(targetMultiAddress.Bytes(), targetValue.Bytes()...)
	expectedBuffer.Write(bArgs)

	buf := bytes.NewBuffer(bArgs)

	target := setupCallForceUnreserve()
	call, err := target.DecodeArgs(buf)
	assert.Nil(t, err)

	buf.Reset()
	call.Encode(buf)

	assert.Equal(t, expectedBuffer, buf)
}

func Test_Call_ForceFree_BaseWeight(t *testing.T) {
	target := setupCallForceUnreserve()

	mockBalances.On("DbWeight").Return(dbWeight)

	result := target.BaseWeight()

	assert.Equal(t, callForceUnreserveWeight(dbWeight), result)
	mockBalances.AssertCalled(t, "DbWeight")
}

func Test_Call_ForceFree_WeighData(t *testing.T) {
	target := setupCallForceUnreserve()

	assert.Equal(t, primitives.WeightFromParts(124, 0), target.WeighData(baseWeight))
}

func Test_Call_ForceFree_ClassifyDispatch(t *testing.T) {
	target := setupCallForceUnreserve()

	assert.Equal(t, primitives.NewDispatchClassNormal(), target.ClassifyDispatch(baseWeight))
}

func Test_Call_ForceFree_PaysFee(t *testing.T) {
	target := setupCallForceUnreserve()

	assert.Equal(t, primitives.PaysYes, target.PaysFee(baseWeight))
}

func Test_Call_ForceFree_Dispatch_Success(t *testing.T) {
	target := setupCallForceUnreserve()

	// TODO: add test in module.go
	// actual := sc.NewU128(1)
	// event := newEventUnreserved(moduleId, targetAddressAccId, actual)
	// mockStoredMap.On("Get", targetAddressAccId).Return(accountInfo, nil)
	// mockStoredMap.On("DepositEvent", event)

	mockBalances.On("Unreserve", targetAddress, targetValue).Return(targetValue, nil)

	_, err := target.Dispatch(primitives.NewRawOriginRoot(), sc.NewVaryingData(targetMultiAddress, targetValue))

	assert.Nil(t, err)
	mockBalances.AssertCalled(t, "Unreserve", targetAddress, targetValue)
}

func Test_Call_ForceFree_Dispatch_InvalidOrigin(t *testing.T) {
	target := setupCallForceUnreserve()

	_, dispatchErr := target.Dispatch(primitives.NewRawOriginNone(), sc.NewVaryingData(targetAddress, targetValue))

	assert.Equal(t, primitives.NewDispatchErrorBadOrigin(), dispatchErr)
	mockBalances.AssertNotCalled(t, "Unreserve", mock.Anything, mock.Anything)
}

func Test_Call_ForceFree_Dispatch_InvalidArgs(t *testing.T) {
	target := setupCallForceUnreserve()

	_, dispatchErr := target.Dispatch(primitives.NewRawOriginRoot(), sc.NewVaryingData(targetMultiAddress, sc.NewU64(0)))

	assert.Equal(t, primitives.NewDispatchErrorOther("invalid u128 value in callForceUnreserve"), dispatchErr)
	mockBalances.AssertNotCalled(t, "Unreserve", mock.Anything, mock.Anything)
}

func Test_Call_ForceFree_Dispatch_InvalidLookup(t *testing.T) {
	target := setupCallForceUnreserve()

	_, dispatchErr := target.Dispatch(primitives.NewRawOriginRoot(), sc.NewVaryingData(primitives.NewMultiAddress20(primitives.Address20{}), targetValue))

	assert.Equal(t, primitives.NewDispatchErrorCannotLookup(), dispatchErr)
	mockBalances.AssertNotCalled(t, "Unreserve", mock.Anything, mock.Anything)
}

func Test_Call_ForceFree_Dispatch_ZeroBalance(t *testing.T) {
	target := setupCallForceUnreserve()

	mockBalances.On("Unreserve", targetAddress, constants.Zero).Return(targetValue, nil)

	_, dispatchErr := target.Dispatch(primitives.NewRawOriginRoot(), sc.NewVaryingData(targetMultiAddress, constants.Zero))

	assert.Nil(t, dispatchErr)
	mockBalances.AssertCalled(t, "Unreserve", targetAddress, constants.Zero)
}

func Test_removeReserveAndFree(t *testing.T) {
	value := sc.NewU128(4)
	accountData := &primitives.AccountData{
		Free:     sc.NewU128(1),
		Reserved: sc.NewU128(10),
	}
	expectedResult := value

	result := removeReserveAndFree(accountData, value)

	assert.Equal(t, expectedResult, result)
	assert.Equal(t, sc.NewU128(6), accountData.Reserved)
	assert.Equal(t, sc.NewU128(5), accountData.Free)
}
