package balances

import (
	sc "github.com/LimeChain/goscale"
	balancestypes "github.com/LimeChain/gosemble/frame/balances/types"

	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/mock"
)

type mockAccountMutator struct {
	mock.Mock
}

func (m *mockAccountMutator) ensureCanWithdraw(who primitives.AccountId, amount sc.U128, reasons primitives.Reasons, newBalance sc.U128) error {
	args := m.Called(who, amount, reasons, newBalance)

	if args[0] != nil {
		return args[0].(error)
	}

	return nil
}

func (m *mockAccountMutator) tryMutateAccountWithDust(who primitives.AccountId, f func(who *primitives.AccountData, bool bool) (sc.Encodable, error)) (sc.Encodable, error) {
	args := m.Called(who, f)

	if args[1] != nil {
		return args[0].(sc.Encodable), args[1].(error)
	}

	return args[0].(sc.Encodable), nil
}

func (m *mockAccountMutator) tryMutateAccount(who primitives.AccountId, f func(who *primitives.AccountData, bool bool) (sc.Encodable, error)) (sc.Encodable, error) {
	args := m.Called(who, f)

	if args[1] != nil {
		return args[0].(sc.Encodable), args[1].(error)
	}

	return args[0].(sc.Encodable), nil
}

func (m *mockAccountMutator) ensureUpgraded(who primitives.AccountId) (bool, error) {
	args := m.Called(who)

	if args[1] != nil {
		return args[0].(bool), args[1].(error)
	}

	return args[0].(bool), nil
}

func (m *mockAccountMutator) transfer(from primitives.AccountId, to primitives.AccountId, amount sc.U128, preservation balancestypes.Preservation) error {
	args := m.Called(from, to, amount)

	if args[0] != nil {
		return args[0].(error)
	}

	return nil
}
