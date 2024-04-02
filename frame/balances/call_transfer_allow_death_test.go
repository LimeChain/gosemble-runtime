package balances

import (
	"bytes"
	"errors"
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	maxLocks           = sc.U32(5)
	maxReserves        = sc.U32(6)
	existentialDeposit = sc.NewU128(1)
	mockMutator        *mockAccountMutator // todo remove + remove the mock type
	testConstants      = newConstants(dbWeight, maxLocks, maxReserves, existentialDeposit)

	fromAccountData *primitives.AccountData
	toAccountData   *primitives.AccountData

	fromAddress = primitives.
			NewMultiAddressId(constants.OneAccountId)
	toAddress = primitives.
			NewMultiAddressId(constants.TwoAccountId)
	argsBytes = sc.NewVaryingData(primitives.MultiAddress{}, sc.Compact{Number: sc.U128{}}).Bytes()

	callTransferArgsBytes = sc.NewVaryingData(primitives.MultiAddress{}, sc.Compact{Number: sc.U128{}}).Bytes()
)

func Test_Call_TransferAllowDeath_New(t *testing.T) {
	target := setupCallTransferAllowDeath()
	expected := callTransferAllowDeath{
		Callable: primitives.Callable{
			ModuleId:   target.module.Index,
			FunctionId: functionTransferAllowDeathIndex,
			Arguments:  sc.NewVaryingData(primitives.MultiAddress{}, sc.Compact{Number: sc.U128{}}),
		},
		module: target.module,
	}

	assert.Equal(t, expected, target)
}

func Test_Call_TransferAllowDeath_DecodeArgs(t *testing.T) {
	amount := sc.ToCompact(sc.NewU128(5))
	buf := bytes.NewBuffer(append(targetAddress.Bytes(), amount.Bytes()...))

	target := setupCallTransferAllowDeath()
	call, err := target.DecodeArgs(buf)
	assert.Nil(t, err)

	assert.Equal(t, sc.NewVaryingData(targetAddress, amount), call.Args())
}

func Test_Call_TransferAllowDeath_Encode(t *testing.T) {
	target := setupCallTransferAllowDeath()
	expectedBuffer := bytes.NewBuffer(append([]byte{moduleId, functionTransferAllowDeathIndex}, callTransferArgsBytes...))
	buf := &bytes.Buffer{}

	err := target.Encode(buf)

	assert.NoError(t, err)
	assert.Equal(t, expectedBuffer, buf)
}

func Test_Call_TransferAllowDeath_Bytes(t *testing.T) {
	expected := append([]byte{moduleId, functionTransferAllowDeathIndex}, callTransferArgsBytes...)

	target := setupCallTransferAllowDeath()

	assert.Equal(t, expected, target.Bytes())
}

func Test_Call_TransferAllowDeath_ModuleIndex(t *testing.T) {
	target := setupCallTransferAllowDeath()

	assert.Equal(t, sc.U8(moduleId), target.ModuleIndex())
}

func Test_Call_TransferAllowDeath_FunctionIndex(t *testing.T) {
	target := setupCallTransferAllowDeath()

	assert.Equal(t, sc.U8(functionTransferAllowDeathIndex), target.FunctionIndex())
}

func Test_Call_TransferAllowDeath_BaseWeight(t *testing.T) {
	target := setupCallTransferAllowDeath()

	assert.Equal(t, callTransferAllowDeathWeight(target.module.dbWeight()), target.BaseWeight())
}

func Test_Call_TransferAllowDeath_WeighData(t *testing.T) {
	target := setupCallTransferAllowDeath()
	assert.Equal(t, primitives.WeightFromParts(124, 0), target.WeighData(baseWeight))
}

func Test_Call_TransferAllowDeath_ClassifyDispatch(t *testing.T) {
	target := setupCallTransferAllowDeath()

	assert.Equal(t, primitives.NewDispatchClassNormal(), target.ClassifyDispatch(baseWeight))
}

func Test_Call_TransferAllowDeath_PaysFee(t *testing.T) {
	target := setupCallTransferAllowDeath()

	assert.Equal(t, primitives.PaysYes, target.PaysFee(baseWeight))
}

func Test_Call_TransferAllowDeath_Dispatch_Success(t *testing.T) {
	target := setupCallTransferAllowDeath()

	mockTotalIssuance.On("Get").Return(sc.MaxU128(), nil)
	mockStoredMap.On("Get", mock.Anything).Return(primitives.AccountInfo{Data: primitives.AccountData{Free: sc.NewU128(10), Flags: primitives.DefaultExtraFlags()}}, nil)
	mockStoredMap.On("TryMutateExistsNoClosure", mock.Anything, mock.Anything).Return(nil)
	mockStoredMap.On("CanDecProviders", mock.Anything).Return(true, nil)
	mockStoredMap.On("IncProviders", mock.Anything).Return(primitives.IncRefStatus(0), nil)
	mockStoredMap.On("DepositEvent", mock.Anything)

	fromAddressId, err := fromAddress.AsAccountId()
	assert.Nil(t, err)

	_, dispatchErr := target.
		Dispatch(primitives.NewRawOriginSigned(fromAddressId), sc.NewVaryingData(fromAddress, sc.ToCompact(targetValue)))

	assert.Nil(t, dispatchErr)
}

func Test_Call_TransferAllowDeath_Dispatch_BadOrigin(t *testing.T) {
	target := setupCallTransferAllowDeath()

	_, dispatchErr := target.Dispatch(primitives.NewRawOriginNone(), sc.NewVaryingData(toAddress, sc.ToCompact(targetValue)))

	assert.Equal(t, primitives.NewDispatchErrorBadOrigin(), dispatchErr)
}

func Test_Call_TransferAllowDeath_Dispatch_InvalidArg_InvalidCompactAmount(t *testing.T) {
	target := setupCallTransferAllowDeath()

	_, dispatchErr := target.Dispatch(primitives.NewRawOriginSigned(accId), sc.NewVaryingData(toAddress, sc.NewU64(0)))

	assert.Equal(t, errors.New("invalid compact value when dispatching call transfer_allow_death"), dispatchErr)
}

func Test_Call_TransferAllowDeath_Dispatch_InvalidArg_InvalidCompactNumber(t *testing.T) {
	target := setupCallTransferAllowDeath()

	_, dispatchErr := target.Dispatch(primitives.NewRawOriginSigned(accId), sc.NewVaryingData(toAddress, sc.Compact{}))

	assert.Equal(t, errors.New("invalid compact number field when dispatching call transfer_allow_death"), dispatchErr)
}

func Test_Call_TransferAllowDeath_Dispatch_CannotLookup(t *testing.T) {
	target := setupCallTransferAllowDeath()

	fromAddressId, err := fromAddress.AsAccountId()
	assert.Nil(t, err)

	_, dispatchErr := target.Dispatch(
		primitives.NewRawOriginSigned(fromAddressId),
		sc.NewVaryingData(primitives.NewMultiAddress20(primitives.Address20{}), sc.ToCompact(targetValue)),
	)

	assert.Equal(t, primitives.NewDispatchErrorCannotLookup(), dispatchErr)
}

func setupCallTransferAllowDeath() callTransferAllowDeath {
	fromAccountData = &primitives.AccountData{
		Free: sc.NewU128(5),
	}

	toAccountData = &primitives.AccountData{
		Free: sc.NewU128(1),
	}

	return newCallTransferAllowDeath(functionTransferAllowDeathIndex, setupModule())
}

// func setupTransfer() transfer {
// 	mockStoredMap = new(mocks.StoredMap)
// 	mockMutator = new(mockAccountMutator)

// 	fromAccountData = &primitives.AccountData{
// 		Free: sc.NewU128(5),
// 	}

// 	toAccountData = &primitives.AccountData{
// 		Free: sc.NewU128(1),
// 	}

// 	return newTransfer(moduleId, mockStoredMap, testConstants, mockMutator)
// }
