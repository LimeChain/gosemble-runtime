package balances

import (
	"bytes"
	"errors"
	"testing"

	sc "github.com/LimeChain/goscale"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	transferKeepAliveArgsBytes = sc.NewVaryingData(primitives.MultiAddress{}, sc.Compact{Number: sc.U128{}}).Bytes()
)

func Test_Call_TransferKeepAlive_new(t *testing.T) {
	target := setupCallTransferKeepAlive()
	expected := callTransferKeepAlive{
		Callable: primitives.Callable{
			ModuleId:   target.module.Index,
			FunctionId: functionTransferKeepAliveIndex,
			Arguments:  sc.NewVaryingData(primitives.MultiAddress{}, sc.Compact{Number: sc.U128{}}),
		},
		module: target.module,
	}

	assert.Equal(t, expected, target)
}

func Test_Call_TransferKeepAlive_DecodeArgs(t *testing.T) {
	amount := sc.ToCompact(sc.NewU128(1))
	buf := bytes.NewBuffer(append(targetAddress.Bytes(), amount.Bytes()...))

	target := setupCallTransferKeepAlive()
	call, err := target.DecodeArgs(buf)
	assert.Nil(t, err)

	assert.Equal(t, sc.NewVaryingData(targetAddress, amount), call.Args())
}

func Test_Call_TransferKeepAlive_Encode(t *testing.T) {
	target := setupCallTransferKeepAlive()
	expectedBuffer := bytes.NewBuffer(append([]byte{moduleId, functionTransferKeepAliveIndex}, transferKeepAliveArgsBytes...))
	buf := &bytes.Buffer{}

	err := target.Encode(buf)

	assert.NoError(t, err)
	assert.Equal(t, expectedBuffer, buf)
}

func Test_Call_TransferKeepAlive_Bytes(t *testing.T) {
	expected := append([]byte{moduleId, functionTransferKeepAliveIndex}, transferKeepAliveArgsBytes...)

	target := setupCallTransferKeepAlive()

	assert.Equal(t, expected, target.Bytes())
}

func Test_Call_TransferKeepAlive_ModuleIndex(t *testing.T) {
	target := setupCallTransferKeepAlive()

	assert.Equal(t, sc.U8(moduleId), target.ModuleIndex())
}

func Test_Call_TransferKeepAlive_FunctionIndex(t *testing.T) {
	target := setupCallTransferKeepAlive()

	assert.Equal(t, sc.U8(functionTransferKeepAliveIndex), target.FunctionIndex())
}

func Test_Call_TransferKeepAlive_BaseWeight(t *testing.T) {
	target := setupCallTransferKeepAlive()

	assert.Equal(t, callTransferKeepAliveWeight(target.module.dbWeight()), target.BaseWeight())
}

func Test_Call_TransferKeepAlive_WeighData(t *testing.T) {
	target := setupCallTransferKeepAlive()
	assert.Equal(t, primitives.WeightFromParts(124, 0), target.WeighData(baseWeight))
}

func Test_Call_TransferKeepAlive_ClassifyDispatch(t *testing.T) {
	target := setupCallTransferKeepAlive()

	assert.Equal(t, primitives.NewDispatchClassNormal(), target.ClassifyDispatch(baseWeight))
}

func Test_Call_TransferKeepAlive_PaysFee(t *testing.T) {
	target := setupCallTransferKeepAlive()

	assert.Equal(t, primitives.PaysYes, target.PaysFee(baseWeight))
}

func Test_Call_TransferKeepAlive_Dispatch_Success(t *testing.T) {
	target := setupCallTransferKeepAlive()

	fromAddressId, err := fromAddress.AsAccountId()
	assert.Nil(t, err)

	toAddressId, err := toAddress.AsAccountId()
	assert.Nil(t, err)

	mockTotalIssuance.On("Get").Return(sc.MaxU128(), nil)
	mockStoredMap.On("Get", mock.Anything).Return(primitives.AccountInfo{Data: primitives.AccountData{Free: sc.NewU128(10), Flags: primitives.DefaultExtraFlags()}}, nil)
	mockStoredMap.On("TryMutateExistsNoClosure", mock.Anything, mock.Anything).Return(nil)
	mockStoredMap.On("CanDecProviders", mock.Anything).Return(true, nil)
	mockStoredMap.On("IncProviders", mock.Anything).Return(primitives.IncRefStatus(0), nil)
	mockStoredMap.On(
		"DepositEvent",
		newEventTransfer(
			moduleId,
			fromAddressId,
			toAddressId,
			targetValue,
		),
	).Return()

	_, dispatchErr := target.Dispatch(primitives.NewRawOriginSigned(fromAddressId), sc.NewVaryingData(toAddress, sc.ToCompact(targetValue)))

	assert.Nil(t, dispatchErr)
	mockStoredMap.AssertCalled(t,
		"TryMutateExistsNoClosure",
		toAddressId,
		mock.Anything,
	)
	mockStoredMap.AssertCalled(t,
		"DepositEvent",
		newEventTransfer(
			moduleId,
			fromAddressId,
			toAddressId,
			targetValue,
		),
	)
}

func Test_Call_TransferKeepAlive_Dispatch_BadOrigin(t *testing.T) {
	target := setupCallTransferKeepAlive()

	_, dispatchErr := target.Dispatch(
		primitives.NewRawOriginNone(),
		sc.NewVaryingData(fromAddress, sc.ToCompact(targetValue)),
	)

	assert.Equal(t, primitives.NewDispatchErrorBadOrigin(), dispatchErr)
	mockStoredMap.AssertNotCalled(t, "TryMutateExistsNoClosure", mock.Anything, mock.Anything)
	mockStoredMap.AssertNotCalled(t, "DepositEvent", mock.Anything)
}

func Test_Call_TransferKeepAlive_Dispatch_InvalidArgs_InvalidCompact(t *testing.T) {
	target := setupCallTransferKeepAlive()

	_, dispatchErr := target.Dispatch(
		primitives.NewRawOriginSigned(accId),
		sc.NewVaryingData(fromAddress, sc.NewU64(0)),
	)

	assert.Equal(t, errors.New("invalid compact value when dispatching call transfer_keep_alive"), dispatchErr)

	mockStoredMap.AssertNotCalled(t, "TryMutateExistsNoClosure", mock.Anything, mock.Anything)
	mockStoredMap.AssertNotCalled(t, "DepositEvent", mock.Anything)
}

func Test_Call_TransferKeepAlive_Dispatch_InvalidArgs_InvalidCompactNumber(t *testing.T) {
	target := setupCallTransferKeepAlive()

	_, dispatchErr := target.Dispatch(
		primitives.NewRawOriginSigned(accId),
		sc.NewVaryingData(fromAddress, sc.Compact{}),
	)

	assert.Equal(t, errors.New("invalid compact number field when dispatching call transfer_keep_alive"), dispatchErr)

	mockStoredMap.AssertNotCalled(t, "TryMutateExistsNoClosure", mock.Anything, mock.Anything)
	mockStoredMap.AssertNotCalled(t, "DepositEvent", mock.Anything)
}

func Test_Call_TransferKeepAlive_Dispatch_CannotLookup(t *testing.T) {
	target := setupCallTransferKeepAlive()

	fromAddressId, err := fromAddress.AsAccountId()
	assert.Nil(t, err)

	_, dispatchErr := target.
		Dispatch(
			primitives.NewRawOriginSigned(fromAddressId),
			sc.NewVaryingData(primitives.NewMultiAddress20(primitives.Address20{}), sc.ToCompact(targetValue)),
		)

	assert.Equal(t, primitives.NewDispatchErrorCannotLookup(), dispatchErr)
	mockStoredMap.AssertNotCalled(t, "TryMutateExistsNoClosure", mock.Anything, mock.Anything)
	mockStoredMap.AssertNotCalled(t, "DepositEvent", mock.Anything)
}

func setupCallTransferKeepAlive() callTransferKeepAlive {
	return newCallTransferKeepAlive(functionTransferKeepAliveIndex, setupModule())
}
