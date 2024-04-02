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

const (
	moduleId = 5
)

var (
	accountInfo = primitives.AccountInfo{
		Data: primitives.AccountData{
			Free:     sc.NewU128(4),
			Reserved: primitives.Balance{},
			Frozen:   primitives.Balance{},
		},
	}
	dbWeight = primitives.RuntimeDbWeight{
		Read:  1,
		Write: 2,
	}
	baseWeight                    = primitives.WeightFromParts(124, 123)
	targetAddress                 = primitives.NewMultiAddressId(constants.ZeroAccountId)
	targetValue                   = sc.NewU128(5)
	mockTypeMutateAccountDataBool = mock.AnythingOfType("func(*types.AccountData, bool) (goscale.Encodable, error)")
	argsBytesCallForceUnreserve   = sc.NewVaryingData(primitives.MultiAddress{}, sc.U128{}).Bytes()
	mockStoredMap                 *mocks.StoredMap
	mockTotalIssuance             *mocks.StorageValue[sc.U128]
	mockInactiveIssuance          *mocks.StorageValue[sc.U128]
	errPanic                      = errors.New("panic")
)

func Test_Call_ForceUnreserve_new(t *testing.T) {
	target := setupCallForceUnreserve()
	expected := callForceUnreserve{
		Callable: primitives.Callable{
			ModuleId:   moduleId,
			FunctionId: functionForceUnreserveIndex,
			Arguments:  sc.NewVaryingData(primitives.MultiAddress{}, sc.U128{}),
		},
		module: target.module,
	}

	assert.Equal(t, expected, target)
}

func Test_Call_ForceUnreserve_DecodeArgs(t *testing.T) {
	amount := sc.NewU128(5)
	buf := bytes.NewBuffer(append(targetAddress.Bytes(), amount.Bytes()...))

	target := setupCallForceUnreserve()
	call, err := target.DecodeArgs(buf)
	assert.Nil(t, err)

	assert.Equal(t, sc.NewVaryingData(targetAddress, amount), call.Args())
}

func Test_Call_ForceUnreserve_Encode(t *testing.T) {
	target := setupCallForceUnreserve()
	expectedBuffer := bytes.NewBuffer(append([]byte{moduleId, functionForceUnreserveIndex}, argsBytesCallForceUnreserve...))
	buf := &bytes.Buffer{}

	err := target.Encode(buf)

	assert.NoError(t, err)
	assert.Equal(t, expectedBuffer, buf)
}

func Test_Call_ForceUnreserve_Bytes(t *testing.T) {
	expected := append([]byte{moduleId, functionForceUnreserveIndex}, argsBytesCallForceUnreserve...)

	target := setupCallForceUnreserve()

	assert.Equal(t, expected, target.Bytes())
}

func Test_Call_ForceUnreserve_ModuleIndex(t *testing.T) {
	target := setupCallForceUnreserve()

	assert.Equal(t, sc.U8(moduleId), target.ModuleIndex())
}

func TestCall_ForceUnreserve_FunctionIndex(t *testing.T) {
	target := setupCallForceUnreserve()

	assert.Equal(t, sc.U8(functionForceUnreserveIndex), target.FunctionIndex())
}

func Test_Call_ForceUnreserve_EncodeWithArgs(t *testing.T) {
	expectedBuffer := bytes.NewBuffer([]byte{moduleId, functionForceUnreserveIndex})
	bArgs := append(targetAddress.Bytes(), targetValue.Bytes()...)
	expectedBuffer.Write(bArgs)

	buf := bytes.NewBuffer(bArgs)

	target := setupCallForceUnreserve()
	call, err := target.DecodeArgs(buf)
	assert.Nil(t, err)

	buf.Reset()
	call.Encode(buf)

	assert.Equal(t, expectedBuffer, buf)
}

func Test_Call_ForceUnreserve_BaseWeight(t *testing.T) {
	target := setupCallForceUnreserve()

	assert.Equal(t, callForceUnreserveWeight(dbWeight), target.BaseWeight())
}

func Test_Call_ForceUnreserve_WeighData(t *testing.T) {
	target := setupCallForceUnreserve()
	assert.Equal(t, primitives.WeightFromParts(124, 0), target.WeighData(baseWeight))
}

func Test_Call_ForceUnreserve_ClassifyDispatch(t *testing.T) {
	target := setupCallForceUnreserve()

	assert.Equal(t, primitives.NewDispatchClassNormal(), target.ClassifyDispatch(baseWeight))
}

func Test_Call_ForceUnreserve_PaysFee(t *testing.T) {
	target := setupCallForceUnreserve()

	assert.Equal(t, primitives.PaysYes, target.PaysFee(baseWeight))
}

func Test_Call_ForceUnreserve_Dispatch_Success(t *testing.T) {
	target := setupCallForceUnreserve()
	// actual := sc.NewU128(1)
	// mutateResult := actual
	targetAddressAccId, err := targetAddress.AsAccountId()
	assert.Nil(t, err)
	// event := newEventUnreserved(moduleId, targetAddressAccId, actual)

	mockStoredMap.On("Get", targetAddressAccId).Return(accountInfo, nil)
	mockStoredMap.On("TryMutateExistsNoClosure", targetAddressAccId, mock.Anything).Return(nil)
	mockStoredMap.On("DepositEvent", mock.Anything)

	_, err = target.Dispatch(primitives.NewRawOriginRoot(), sc.NewVaryingData(targetAddress, targetValue))
	assert.Nil(t, err)
	// mockStoredMap.AssertCalled(t, "Get", targetAddressAccId)
	mockStoredMap.AssertCalled(t,
		"TryMutateExistsNoClosure",
		targetAddressAccId,
		mock.Anything,
	)
	mockStoredMap.AssertCalled(t, "DepositEvent", mock.Anything)
}

func Test_Call_ForceUnreserve_Dispatch_InvalidOrigin(t *testing.T) {
	target := setupCallForceUnreserve()

	_, dispatchErr := target.Dispatch(primitives.NewRawOriginNone(), sc.NewVaryingData(targetAddress, targetValue))

	assert.Equal(t, primitives.NewDispatchErrorBadOrigin(), dispatchErr)
	mockStoredMap.AssertNotCalled(t, "Get", mock.Anything)
	mockStoredMap.AssertNotCalled(t, "DepositEvent", mock.Anything)
}

func Test_Call_ForceUnreserve_Dispatch_InvalidArgs(t *testing.T) {
	target := setupCallForceUnreserve()

	_, dispatchErr := target.Dispatch(primitives.NewRawOriginRoot(), sc.NewVaryingData(targetAddress, sc.NewU64(0)))

	assert.Equal(t, errors.New("invalid amount value when dispatching call force unreserve").Error(), dispatchErr.Error())
	mockStoredMap.AssertNotCalled(t, "Get", mock.Anything)
	mockStoredMap.AssertNotCalled(t, "DepositEvent", mock.Anything)
}

func Test_Call_ForceUnreserve_Dispatch_InvalidLookup(t *testing.T) {
	target := setupCallForceUnreserve()

	_, dispatchErr := target.Dispatch(primitives.NewRawOriginRoot(), sc.NewVaryingData(primitives.NewMultiAddress20(primitives.Address20{}), targetValue))

	assert.Equal(t, primitives.NewDispatchErrorCannotLookup(), dispatchErr)
	targetAddressAccId, err := targetAddress.AsAccountId()
	assert.Nil(t, err)
	mockStoredMap.AssertNotCalled(t, "Get", targetAddressAccId)
	mockStoredMap.AssertNotCalled(t, "DepositEvent", mock.Anything)
}

func Test_Call_ForceUnreserve_Dispatch_ZeroBalance(t *testing.T) {
	target := setupCallForceUnreserve()

	_, dispatchErr := target.Dispatch(primitives.NewRawOriginRoot(), sc.NewVaryingData(targetAddress, constants.Zero))

	assert.Nil(t, dispatchErr)
	targetAddressAccId, err := targetAddress.AsAccountId()
	assert.Nil(t, err)
	mockStoredMap.AssertNotCalled(t, "Get", targetAddressAccId)
	mockStoredMap.AssertNotCalled(t, "DepositEvent", mock.Anything)
}

func Test_Call_ForceUnreserve_Dispatch_ZeroTotalStorageBalance(t *testing.T) {
	target := setupCallForceUnreserve()
	accountInfo := primitives.AccountInfo{Data: primitives.AccountData{}}

	targetAddressAccId, err := targetAddress.AsAccountId()
	assert.Nil(t, err)
	mockStoredMap.On("Get", targetAddressAccId).Return(accountInfo, nil)

	_, dispatchErr := target.Dispatch(primitives.NewRawOriginRoot(), sc.NewVaryingData(targetAddress, targetValue))

	assert.Nil(t, dispatchErr)
	mockStoredMap.AssertCalled(t, "Get", targetAddressAccId)
	mockStoredMap.AssertNotCalled(t, "DepositEvent", mock.Anything)
}

func Test_Call_ForceUnreserve_Dispatch_Other(t *testing.T) {
	target := setupCallForceUnreserve()
	accountInfo := primitives.AccountInfo{Data: primitives.AccountData{}}

	targetAddressAccId, err := targetAddress.AsAccountId()
	assert.Nil(t, err)

	mockStoredMap.On("Get", targetAddressAccId).Return(accountInfo, errPanic)

	_, dispatchErr := target.Dispatch(primitives.NewRawOriginRoot(), sc.NewVaryingData(targetAddress, targetValue))

	assert.Equal(t, primitives.NewDispatchErrorOther(sc.Str(errPanic.Error())), dispatchErr)
	mockStoredMap.AssertCalled(t, "Get", targetAddressAccId)
	mockStoredMap.AssertNotCalled(t, "DepositEvent", mock.Anything)
}

func setupCallForceUnreserve() callForceUnreserve {
	call := newCallForceUnreserve(functionForceUnreserveIndex, setupModule())
	mockStoredMap = new(mocks.StoredMap)
	mockTotalIssuance.On("Get").Return(sc.MaxU128(), nil)
	mockStoredMap.On("TryMutateExistsNoClosure", mock.Anything, mock.Anything).Return(nil)
	mockStoredMap.On("CanDecProviders", mock.Anything).Return(true, nil)
	mockStoredMap.On("IncProviders", mock.Anything).Return(primitives.IncRefStatus(0), nil)
	mockStoredMap.On("DepositEvent", mock.Anything)
	call.module.Config.StoredMap = mockStoredMap
	return call
}
