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
	callForceTransferArgsBytes = sc.NewVaryingData(primitives.MultiAddress{}, primitives.MultiAddress{}, sc.Compact{Number: sc.U128{}}).Bytes()
)

func Test_Call_ForceTransfer_new(t *testing.T) {
	target := setupCallForceTransfer()
	expected := callForceTransfer{
		Callable: primitives.Callable{
			ModuleId:   target.module.Index,
			FunctionId: functionForceTransferIndex,
			Arguments:  sc.NewVaryingData(primitives.MultiAddress{}, primitives.MultiAddress{}, sc.Compact{Number: sc.U128{}}),
		},
		module: target.module,
	}

	assert.Equal(t, expected, target)
}

func Test_Call_ForceTransfer_DecodeArgs(t *testing.T) {
	amount := sc.ToCompact(sc.NewU128(1))
	buf := &bytes.Buffer{}
	buf.Write(fromAddress.Bytes())
	buf.Write(toAddress.Bytes())
	buf.Write(amount.Bytes())

	target := setupCallForceTransfer()
	call, err := target.DecodeArgs(buf)
	assert.Nil(t, err)

	assert.Equal(t, sc.NewVaryingData(fromAddress, toAddress, amount), call.Args())
}

func Test_Call_ForceTransfer_Encode(t *testing.T) {
	target := setupCallForceTransfer()
	expectedBuffer := bytes.NewBuffer(append([]byte{moduleId, functionForceTransferIndex}, callForceTransferArgsBytes...))
	buf := &bytes.Buffer{}

	err := target.Encode(buf)

	assert.NoError(t, err)
	assert.Equal(t, expectedBuffer, buf)
}

func Test_Call_ForceTransfer_Bytes(t *testing.T) {
	expected := append([]byte{moduleId, functionForceTransferIndex}, callForceTransferArgsBytes...)

	target := setupCallForceTransfer()

	assert.Equal(t, expected, target.Bytes())
}

func Test_Call_ForceTransfer_ModuleIndex(t *testing.T) {
	target := setupCallForceTransfer()

	assert.Equal(t, sc.U8(moduleId), target.ModuleIndex())
}

func Test_Call_ForceTransfer_FunctionIndex(t *testing.T) {
	target := setupCallForceTransfer()

	assert.Equal(t, sc.U8(functionForceTransferIndex), target.FunctionIndex())
}

func Test_Call_ForceTransfer_BaseWeight(t *testing.T) {
	target := setupCallForceTransfer()

	assert.Equal(t, callForceTransferWeight(target.module.dbWeight()), target.BaseWeight())
}

func Test_Call_ForceTransfer_WeighData(t *testing.T) {
	target := setupCallForceTransfer()
	assert.Equal(t, primitives.WeightFromParts(124, 0), target.WeighData(baseWeight))
}

func Test_Call_ForceTransfer_ClassifyDispatch(t *testing.T) {
	target := setupCallForceTransfer()

	assert.Equal(t, primitives.NewDispatchClassNormal(), target.ClassifyDispatch(baseWeight))
}

func Test_Call_ForceTransfer_PaysFee(t *testing.T) {
	target := setupCallForceTransfer()

	assert.Equal(t, primitives.PaysYes, target.PaysFee(baseWeight))
}

func Test_Call_ForceTransfer_Dispatch_Success(t *testing.T) {
	target := setupCallForceTransfer()

	fromAddressAccId, err := fromAddress.AsAccountId()
	assert.Nil(t, err)

	toAddressAccId, err := toAddress.AsAccountId()
	assert.Nil(t, err)

	mockTotalIssuance.On("Get").Return(sc.MaxU128(), nil)
	mockStoredMap.On("Get", mock.Anything).Return(primitives.AccountInfo{Data: primitives.AccountData{Free: sc.NewU128(10), Flags: primitives.DefaultExtraFlags()}}, nil)
	mockStoredMap.On("TryMutateExistsNoClosure", mock.Anything, mock.Anything).Return(nil)
	mockStoredMap.On("CanDecProviders", mock.Anything).Return(true, nil)
	mockStoredMap.On("IncProviders", mock.Anything).Return(primitives.IncRefStatus(0), nil)
	mockStoredMap.On("DepositEvent", mock.Anything)

	_, dispatchErr := target.Dispatch(primitives.NewRawOriginRoot(), sc.NewVaryingData(fromAddress, toAddress, sc.ToCompact(targetValue)))

	assert.Nil(t, dispatchErr)

	mockStoredMap.AssertCalled(t,
		"DepositEvent",
		newEventTransfer(moduleId, fromAddressAccId, toAddressAccId, targetValue),
	)
}

func Test_Call_ForceTransfer_Dispatch_InvalidBadOrigin(t *testing.T) {
	target := setupCallForceTransfer()

	_, dispatchErr := target.Dispatch(
		primitives.NewRawOriginNone(),
		sc.NewVaryingData(fromAddress, toAddress, sc.ToCompact(targetValue)))

	assert.Equal(t, primitives.NewDispatchErrorBadOrigin(), dispatchErr)
}

func Test_Call_ForceTransfer_Dispatch_InvalidArg_InvalidCompact(t *testing.T) {
	target := setupCallForceTransfer()

	_, dispatchErr := target.Dispatch(
		primitives.NewRawOriginRoot(),
		sc.NewVaryingData(fromAddress, toAddress, sc.NewU128(0)))

	assert.Equal(t, errors.New("invalid Compact value when dispatching call_force_transfer"), dispatchErr)
}

func Test_Call_ForceTransfer_Dispatch_InvalidArg_InvalidCompactNumber(t *testing.T) {
	target := setupCallForceTransfer()

	_, dispatchErr := target.Dispatch(
		primitives.NewRawOriginRoot(),
		sc.NewVaryingData(fromAddress, toAddress, sc.Compact{}))

	assert.Equal(t, errors.New("invalid Compact field number when dispatching call_force_transfer"), dispatchErr)

	mockMutator.AssertNotCalled(t, "tryMutateAccountWithDust", mock.Anything, mock.Anything)
	mockStoredMap.AssertNotCalled(t, "DepositEvent", mock.Anything)
}

func Test_Call_ForceTransfer_Dispatch_CannotLookup_Source(t *testing.T) {
	target := setupCallForceTransfer()

	_, dispatchErr := target.Dispatch(
		primitives.NewRawOriginRoot(),
		sc.NewVaryingData(primitives.NewMultiAddress20(primitives.Address20{}), toAddress, sc.ToCompact(targetValue)),
	)

	assert.Equal(t, primitives.NewDispatchErrorCannotLookup(), dispatchErr)
	mockMutator.AssertNotCalled(t, "tryMutateAccountWithDust", mock.Anything, mock.Anything)
	mockStoredMap.AssertNotCalled(t, "DepositEvent", mock.Anything)
}

func Test_Call_ForceTransfer_Dispatch_CannotLookup_Dest(t *testing.T) {
	target := setupCallForceTransfer()

	_, dispatchErr := target.Dispatch(
		primitives.NewRawOriginRoot(),
		sc.NewVaryingData(fromAddress, primitives.NewMultiAddress20(primitives.Address20{}), sc.ToCompact(targetValue)),
	)

	assert.Equal(t, primitives.NewDispatchErrorCannotLookup(), dispatchErr)
	mockMutator.AssertNotCalled(t, "tryMutateAccountWithDust", mock.Anything, mock.Anything)
	mockStoredMap.AssertNotCalled(t, "DepositEvent", mock.Anything)
}

func setupCallForceTransfer() callForceTransfer {
	mockStoredMap = new(mocks.StoredMap)
	mockMutator = new(mockAccountMutator)

	return newCallForceTransfer(functionForceTransferIndex, setupModule())
}
