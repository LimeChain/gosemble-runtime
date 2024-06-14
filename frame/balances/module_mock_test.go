package balances

import (
	"github.com/LimeChain/goscale"
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/frame/support"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/mock"
)

type MockModule struct {
	mock.Mock
}

func (m *MockModule) GetIndex() sc.U8 {
	args := m.Called()
	return args.Get(0).(sc.U8)
}

func (m *MockModule) Functions() map[sc.U8]primitives.Call {
	args := m.Called()
	return args.Get(0).(map[sc.U8]primitives.Call)
}

func (m *MockModule) PreDispatch(call primitives.Call) (sc.Empty, error) {
	args := m.Called(call)
	return args.Get(0).(sc.Empty), args.Get(1).(error)
}

func (m *MockModule) ValidateUnsigned(source primitives.TransactionSource, call primitives.Call) (primitives.ValidTransaction, error) {
	args := m.Called(source, call)
	return args.Get(0).(primitives.ValidTransaction), args.Get(1).(error)
}

func (m *MockModule) Metadata() primitives.MetadataModule {
	args := m.Called()
	return args.Get(0).(primitives.MetadataModule)
}

func (m *MockModule) CreateInherent(inherent primitives.InherentData) (sc.Option[primitives.Call], error) {
	args := m.Called(inherent)

	if args.Get(1) == nil {
		return args.Get(0).(sc.Option[primitives.Call]), nil
	}

	return args.Get(0).(sc.Option[primitives.Call]), args.Get(1).(error)
}

func (m *MockModule) CheckInherent(call primitives.Call, data primitives.InherentData) error {
	args := m.Called(call, data)
	return args.Get(0).(error)
}

func (m *MockModule) InherentIdentifier() [8]byte {
	args := m.Called()
	return args.Get(0).([8]byte)
}

func (m *MockModule) IsInherent(call primitives.Call) bool {
	args := m.Called(call)
	return args.Get(0).(bool)
}

func (m *MockModule) OnInitialize(n sc.U64) (primitives.Weight, error) {
	args := m.Called(n)

	if args.Get(1) == nil {
		return args.Get(0).(primitives.Weight), nil
	}

	return args.Get(0).(primitives.Weight), args.Get(1).(error)
}

func (m *MockModule) OnRuntimeUpgrade() primitives.Weight {
	args := m.Called()
	return args.Get(0).(primitives.Weight)
}

func (m *MockModule) OnFinalize(n sc.U64) error {
	args := m.Called(n)
	return args.Get(0).(error)
}

func (m *MockModule) OnIdle(n sc.U64, remainingWeight primitives.Weight) primitives.Weight {
	args := m.Called(n, remainingWeight)
	return args.Get(0).(primitives.Weight)
}

func (m *MockModule) OffchainWorker(n sc.U64) {
	m.Called(n)
}

func (m *MockModule) Unreserve(who primitives.AccountId, value sc.U128) (sc.U128, error) {
	args := m.Called(who, value)

	if args.Get(1) == nil {
		return args.Get(0).(sc.U128), nil
	}

	return args.Get(0).(sc.U128), args.Get(1).(error)
}

func (m *MockModule) DbWeight() primitives.RuntimeDbWeight {
	args := m.Called()
	return args.Get(0).(primitives.RuntimeDbWeight)
}

func (m *MockModule) ExistentialDeposit() sc.U128 {
	args := m.Called()
	return args.Get(0).(sc.U128)
}

func (m *MockModule) MutateAccountHandlingDust(who primitives.AccountId, f func(who *primitives.AccountData, bool bool) (sc.Encodable, error)) (sc.Encodable, error) {
	args := m.Called(who, f)

	if args.Get(1) == nil {
		return args.Get(0).(sc.Encodable), nil
	}

	return args.Get(0).(sc.Encodable), args.Get(1).(error)
}

func (m *MockModule) TotalIssuance() support.StorageValue[goscale.U128] {
	args := m.Called()
	return args.Get(0).(support.StorageValue[goscale.U128])
}

func (m *MockModule) DepositEvent(event primitives.Event) {
	m.Called(event)
}

func (m *MockModule) DepositIntoExisting(who primitives.AccountId, value sc.U128) (primitives.Balance, error) {
	args := m.Called(who, value)

	if args.Get(1) == nil {
		return args.Get(0).(primitives.Balance), nil
	}

	return args.Get(0).(primitives.Balance), args.Get(1).(error)
}

func (m *MockModule) Withdraw(who primitives.AccountId, value sc.U128, reasons sc.U8, liveness primitives.ExistenceRequirement) (primitives.Balance, error) {
	args := m.Called(who, value, reasons, liveness)

	if args.Get(1) == nil {
		return args.Get(0).(primitives.Balance), nil
	}

	return args.Get(0).(primitives.Balance), args.Get(1).(error)
}
