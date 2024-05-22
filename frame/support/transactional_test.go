package support

import (
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/mocks"
	"github.com/LimeChain/gosemble/primitives/log"
	"github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

var (
	transactionLevel      = sc.U32(10)
	mockStorageValue      *mocks.StorageValue[sc.U32]
	mockTransactionBroker *mocks.IoTransactionBroker
)

func Test_Transactional_GetTransactionLevel(t *testing.T) {
	target := setupTransactional()

	mockStorageValue.On("Get").Return(transactionLevel, nil)
	txLevel, err := target.GetTransactionLevel()
	assert.NoError(t, err)

	assert.Equal(t, transactionLevel, txLevel)
	mockStorageValue.AssertCalled(t, "Get")
}

func Test_Transactional_SetTransactionLevel(t *testing.T) {
	target := setupTransactional()

	mockStorageValue.On("Put", transactionLevel).Return()

	target.SetTransactionLevel(transactionLevel)

	mockStorageValue.AssertCalled(t, "Put", transactionLevel)
}

func Test_Transactional_KillTransactionLevel(t *testing.T) {
	target := setupTransactional()

	mockStorageValue.On("Clear").Return()

	target.KillTransactionLevel()

	mockStorageValue.AssertCalled(t, "Clear")
}

func Test_Transactional_IncTransactionLevel(t *testing.T) {
	target := setupTransactional()

	mockStorageValue.On("Get").Return(transactionLevel, nil)
	mockStorageValue.On("Put", transactionLevel+1).Return()

	err := target.IncTransactionLevel()

	assert.Nil(t, err)
	mockStorageValue.AssertCalled(t, "Get")
	mockStorageValue.AssertCalled(t, "Put", transactionLevel+1)
}

func Test_Transactional_IncTransactionLevel_MaxLimit(t *testing.T) {
	target := setupTransactional()

	mockStorageValue.On("Get").Return(TransactionalLimit, nil)

	result := target.IncTransactionLevel()

	assert.Equal(t, errTransactionalLimitReached, result)
	mockStorageValue.AssertCalled(t, "Get")
	mockStorageValue.AssertNotCalled(t, "Put", mock.Anything)
}

func Test_Transactional_DecTransactionLevel_Zero(t *testing.T) {
	target := setupTransactional()

	mockStorageValue.On("Get").Return(sc.U32(0), nil)

	target.DecTransactionLevel()

	mockStorageValue.AssertCalled(t, "Get")
	mockStorageValue.AssertNotCalled(t, "Clear")
	mockStorageValue.AssertNotCalled(t, "Put", mock.Anything)
}

func Test_Transactional_DecTransactionLevel_One(t *testing.T) {
	target := setupTransactional()

	mockStorageValue.On("Get").Return(sc.U32(1), nil)
	mockStorageValue.On("Clear").Return()

	target.DecTransactionLevel()

	mockStorageValue.AssertCalled(t, "Get")
	mockStorageValue.AssertCalled(t, "Clear")
	mockStorageValue.AssertNotCalled(t, "Put", mock.Anything)
}

func Test_Transactional_DecTransactionLevel(t *testing.T) {
	target := setupTransactional()

	mockStorageValue.On("Get").Return(transactionLevel, nil)
	mockStorageValue.On("Put", transactionLevel-1).Return()

	target.DecTransactionLevel()

	mockStorageValue.AssertCalled(t, "Get")
	mockStorageValue.AssertCalled(t, "Put", transactionLevel-1)
}

func Test_Transactional_WithTransaction_ErrorLimitReached(t *testing.T) {
	target := setupTransactional()
	expect := types.NewDispatchErrorTransactional(types.NewTransactionalErrorLimitReached())

	mockStorageValue.On("Get").Return(TransactionalLimit, nil)

	res, err := target.WithTransaction(func() types.TransactionOutcome {
		return types.NewTransactionOutcomeCommit(sc.U32(2))
	})

	assert.Equal(t, sc.U32(0), res)
	assert.Equal(t, expect, err)

	mockStorageValue.AssertCalled(t, "Get")
}

func Test_Transactional_WithTransaction_Commit(t *testing.T) {
	target := setupTransactional()
	expect := sc.U32(2)

	mockStorageValue.On("Get").Return(transactionLevel, nil).Once()
	mockStorageValue.On("Put", transactionLevel+1).Once()
	mockTransactionBroker.On("Start").Return()
	mockTransactionBroker.On("Commit").Return()
	mockStorageValue.On("Get").Return(transactionLevel+1, nil).Once()
	mockStorageValue.On("Put", transactionLevel).Once()

	res, err := target.WithTransaction(func() types.TransactionOutcome {
		return types.NewTransactionOutcomeCommit(expect)
	})

	assert.Equal(t, expect, res)
	assert.Nil(t, err)

	mockStorageValue.AssertNumberOfCalls(t, "Get", 2)
	mockStorageValue.AssertCalled(t, "Put", transactionLevel+1)
	mockTransactionBroker.AssertCalled(t, "Start")
	mockTransactionBroker.AssertCalled(t, "Commit")
	mockStorageValue.AssertCalled(t, "Put", transactionLevel)
}

func Test_Transactional_WithTransaction_Rollback(t *testing.T) {
	target := setupTransactional()
	expect := types.NewDispatchErrorCorruption()

	mockStorageValue.On("Get").Return(transactionLevel, nil).Once()
	mockStorageValue.On("Put", transactionLevel+1).Once()
	mockTransactionBroker.On("Start").Return()
	mockTransactionBroker.On("Rollback").Return()
	mockStorageValue.On("Get").Return(transactionLevel+1, nil).Once()
	mockStorageValue.On("Put", transactionLevel).Once()

	res, err := target.WithTransaction(func() types.TransactionOutcome {
		return types.NewTransactionOutcomeRollback(expect)
	})

	assert.Equal(t, sc.U32(0), res)
	assert.Equal(t, expect, err)

	mockStorageValue.AssertNumberOfCalls(t, "Get", 2)
	mockStorageValue.AssertNumberOfCalls(t, "Put", 2)
	mockStorageValue.AssertCalled(t, "Put", transactionLevel+1)
	mockTransactionBroker.AssertCalled(t, "Start")
	mockTransactionBroker.AssertCalled(t, "Rollback")
	mockStorageValue.AssertCalled(t, "Put", transactionLevel)
}

func Test_Transactional_WithTransaction_InvalidTransactionOutcome(t *testing.T) {
	target := setupTransactional()

	mockStorageValue.On("Get").Return(transactionLevel, nil).Once()
	mockStorageValue.On("Put", transactionLevel+1).Once()
	mockTransactionBroker.On("Start").Return()

	_, err := target.WithTransaction(func() types.TransactionOutcome {
		return types.TransactionOutcome(sc.NewVaryingData(sc.U32(3)))
	})
	assert.Equal(t, types.NewDispatchErrorOther(sc.Str(errInvalidTransactionOutcome.Error())), err)

	mockStorageValue.AssertCalled(t, "Get")
	mockStorageValue.AssertCalled(t, "Put", transactionLevel+1)
	mockTransactionBroker.AssertCalled(t, "Start")
	mockTransactionBroker.AssertNotCalled(t, "Rollback")
}

func Test_Transactional_WithStorageLayer_Commit(t *testing.T) {
	target := setupTransactional()
	expect := sc.U32(2)

	mockStorageValue.On("Get").Return(transactionLevel, nil).Once()
	mockStorageValue.On("Put", transactionLevel+1).Once()
	mockTransactionBroker.On("Start").Return()
	mockTransactionBroker.On("Commit").Return()
	mockStorageValue.On("Get").Return(transactionLevel+1, nil).Once()
	mockStorageValue.On("Put", transactionLevel).Once()

	res, err := target.WithStorageLayer(func() (sc.U32, error) {
		return expect, nil
	})

	assert.Equal(t, expect, res)
	assert.Nil(t, err)

	mockStorageValue.AssertNumberOfCalls(t, "Get", 2)
	mockStorageValue.AssertCalled(t, "Put", transactionLevel+1)
	mockTransactionBroker.AssertCalled(t, "Start")
	mockTransactionBroker.AssertCalled(t, "Commit")
	mockStorageValue.AssertCalled(t, "Put", transactionLevel)
}

func Test_Transactional_WithStorageLayer_Rollback(t *testing.T) {
	target := setupTransactional()
	expect := types.NewDispatchErrorCorruption()

	mockStorageValue.On("Get").Return(transactionLevel, nil).Once()
	mockStorageValue.On("Put", transactionLevel+1).Once()
	mockTransactionBroker.On("Start").Return()
	mockTransactionBroker.On("Rollback").Return()
	mockStorageValue.On("Get").Return(transactionLevel+1, nil).Once()
	mockStorageValue.On("Put", transactionLevel).Once()

	res, err := target.WithStorageLayer(func() (sc.U32, error) {
		return sc.U32(0), expect
	})

	assert.Equal(t, sc.U32(0), res)
	assert.Equal(t, expect, err)

	mockStorageValue.AssertNumberOfCalls(t, "Get", 2)
	mockStorageValue.AssertNumberOfCalls(t, "Put", 2)
	mockStorageValue.AssertCalled(t, "Put", transactionLevel+1)
	mockTransactionBroker.AssertCalled(t, "Start")
	mockTransactionBroker.AssertCalled(t, "Rollback")
	mockStorageValue.AssertCalled(t, "Put", transactionLevel)
}

func Test_Transactional_WithStorageLayer_Panics(t *testing.T) {
	target := setupTransactional()

	mockStorageValue.On("Get").Return(transactionLevel, nil).Once()
	mockStorageValue.On("Put", transactionLevel+1).Once()
	mockTransactionBroker.On("Start").Return()
	mockTransactionBroker.On("Rollback").Return()
	mockStorageValue.On("Get").Return(transactionLevel+1, nil).Once()
	mockStorageValue.On("Put", transactionLevel).Once()

	_, err := target.WithStorageLayer(func() (sc.U32, error) {
		return sc.U32(0), errPanic
	})

	assert.Equal(t, types.NewDispatchErrorOther(sc.Str(errInvalidTransactionOutcome.Error())), err)
	mockStorageValue.AssertCalled(t, "Get")
	mockStorageValue.AssertCalled(t, "Put", transactionLevel+1)
	mockTransactionBroker.AssertCalled(t, "Start")
}

func setupTransactional() transactional[sc.U32] {
	mockStorageValue = new(mocks.StorageValue[sc.U32])
	mockTransactionBroker = new(mocks.IoTransactionBroker)

	target := NewTransactional[sc.U32](new(mocks.IoStorage), mockTransactionBroker, log.NewLogger()).(transactional[sc.U32])
	target.storage = mockStorageValue
	target.transactionBroker = mockTransactionBroker

	return target
}
