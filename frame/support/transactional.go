package support

import (
	"errors"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/primitives/io"
	"github.com/LimeChain/gosemble/primitives/log"
	"github.com/LimeChain/gosemble/primitives/types"
)

// Layer is the type that is being used to store the current number of active layers.
type Layer = sc.U32

// TransactionalLimit returns the maximum number of nested layers.
const TransactionalLimit Layer = 255

var keyTransactionLevel = []byte(":transaction_level:")

var (
	errTransactionalLimitReached = errors.New("transactional error limit reached")
	errInvalidTransactionOutcome = errors.New("invalid transaction outcome")
)

type Transactional[T sc.Encodable] interface {
	WithStorageLayer(fn func() (T, error)) (T, error)
}

type transactional[T sc.Encodable] struct {
	storage           StorageValue[sc.U32]
	transactionBroker io.TransactionBroker
	logger            log.WarnLogger
}

func NewTransactional[T sc.Encodable](storage io.Storage, transactionBroker io.TransactionBroker, logger log.WarnLogger) Transactional[T] {
	storageVal := NewSimpleStorageValue(storage, keyTransactionLevel, sc.DecodeU32)
	return transactional[T]{
		storage:           storageVal,
		transactionBroker: transactionBroker,
		logger:            logger,
	}
}

// GetTransactionLevel returns the current number of nested transactional layers.
func (t transactional[T]) GetTransactionLevel() (Layer, error) {
	return t.storage.Get()
}

// SetTransactionLevel Set the current number of nested transactional layers.
func (t transactional[T]) SetTransactionLevel(level Layer) {
	t.storage.Put(level)
}

// KillTransactionLevel kill the transactional layers storage.
func (t transactional[T]) KillTransactionLevel() {
	t.storage.Clear()
}

// IncTransactionLevel increments the transaction level. Returns an error if levels go past the limit.
//
// Returns a guard that when dropped decrements the transaction level automatically.
func (t transactional[T]) IncTransactionLevel() error {
	existingLevels, err := t.GetTransactionLevel()
	if err != nil {
		return err
	}
	if existingLevels >= TransactionalLimit {
		return errTransactionalLimitReached
	}
	// Cannot overflow because of check above.
	t.SetTransactionLevel(existingLevels + 1)
	return nil
}

func (t transactional[T]) DecTransactionLevel() error {
	existingLevels, err := t.GetTransactionLevel()
	if err != nil {
		return err
	}
	if existingLevels == 0 {
		t.logger.Warn("We are underflowing with calculating transactional levels. Not great, but let's not panic...")
	} else if existingLevels == 1 {
		// Don't leave any trace of this storage item.
		t.KillTransactionLevel()
	} else {
		// Cannot underflow because of checks above.
		t.SetTransactionLevel(existingLevels - 1)
	}
	return nil
}

// WithTransaction executes the supplied function in a new storage transaction.
//
// All changes to storage performed by the supplied function are discarded if the returned
// outcome is `TransactionOutcome::Rollback`.
//
// Transactions can be nested up to `TRANSACTIONAL_LIMIT` times; more than that will result in an
// error.
//
// Commits happen to the parent transaction.
func (t transactional[T]) WithTransaction(fn func() types.TransactionOutcome) (ok T, err error) {
	// This needs to happen before `start_transaction` below.
	// Otherwise we may rollback the increase, then decrease as the guard goes out of scope
	// and then end in some bad state.
	e := t.IncTransactionLevel()
	if e != nil {
		return ok, types.NewDispatchErrorTransactional(types.NewTransactionalErrorLimitReached())
	}

	t.transactionBroker.Start()

	res := fn()

	switch res[0] {
	case types.TransactionOutcomeCommit:
		t.transactionBroker.Commit()
		t.DecTransactionLevel()
		return res[1].(T), nil
	case types.TransactionOutcomeRollback:
		t.transactionBroker.Rollback()
		t.DecTransactionLevel()
		return ok, res[1].(error)
	default:
		return ok, types.NewDispatchErrorOther(sc.Str(errInvalidTransactionOutcome.Error()))
	}
}

// WithStorageLayer executes the supplied function, adding a new storage layer.
//
// This is the same as `with_transaction`, but assuming that any function returning an `Err` should
// rollback, and any function returning `Ok` should commit. This provides a cleaner API to the
// developer who wants this behavior.
func (t transactional[T]) WithStorageLayer(fn func() (T, error)) (T, error) {
	return t.WithTransaction(
		func() types.TransactionOutcome {
			switch ok, err := fn(); typedErr := err.(type) {
			case types.DispatchError:
				return types.NewTransactionOutcomeRollback(typedErr)
			case nil:
				return types.NewTransactionOutcomeCommit(ok)
			default:
				return types.TransactionOutcome{sc.U8(99)}
			}
		},
	)
}
