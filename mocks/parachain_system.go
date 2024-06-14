package mocks

import (
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/primitives/parachain"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/mock"
)

type ParachainSystemModule struct {
	mock.Mock
}

func (m *ParachainSystemModule) GetIndex() sc.U8 {
	args := m.Called()
	return args.Get(0).(sc.U8)
}

func (m *ParachainSystemModule) Functions() map[sc.U8]primitives.Call {
	args := m.Called()
	return args.Get(0).(map[sc.U8]primitives.Call)
}

func (m *ParachainSystemModule) PreDispatch(call primitives.Call) (sc.Empty, error) {
	args := m.Called(call)
	return args.Get(0).(sc.Empty), args.Get(1).(error)
}

func (m *ParachainSystemModule) ValidateUnsigned(txSource primitives.TransactionSource, call primitives.Call) (primitives.ValidTransaction, error) {
	args := m.Called(txSource, call)
	return args.Get(0).(primitives.ValidTransaction), args.Get(1).(error)
}

func (m *ParachainSystemModule) CreateInherent(inherent primitives.InherentData) (sc.Option[primitives.Call], error) {
	args := m.Called(inherent)
	if args.Get(1) == nil {
		return args.Get(0).(sc.Option[primitives.Call]), nil
	}
	return args.Get(0).(sc.Option[primitives.Call]), args.Get(1).(error)
}

func (m *ParachainSystemModule) CheckInherent(call primitives.Call, inherent primitives.InherentData) error {
	args := m.Called(call, inherent)
	return args.Get(0).(error)
}

func (m *ParachainSystemModule) InherentIdentifier() [8]byte {
	args := m.Called()
	return args.Get(0).([8]byte)
}

func (m *ParachainSystemModule) IsInherent(call primitives.Call) bool {
	args := m.Called(call)
	return args.Get(0).(bool)
}

func (m *ParachainSystemModule) OnInitialize(n sc.U64) (primitives.Weight, error) {
	args := m.Called(n)
	if args.Get(1) == nil {
		return args.Get(0).(primitives.Weight), nil
	}
	return args.Get(0).(primitives.Weight), args.Get(1).(error)
}

func (m *ParachainSystemModule) OnFinalize(n sc.U64) error {
	m.Called(n)
	return nil
}

func (m *ParachainSystemModule) OnIdle(n sc.U64, remainingWeight primitives.Weight) primitives.Weight {
	args := m.Called(n, remainingWeight)
	return args.Get(0).(primitives.Weight)
}

func (m *ParachainSystemModule) OnRuntimeUpgrade() primitives.Weight {
	args := m.Called()
	return args.Get(0).(primitives.Weight)
}

func (m *ParachainSystemModule) OffchainWorker(n sc.U64) {
	m.Called(n)
}

func (m *ParachainSystemModule) StorageNewValidationCodeBytes() (sc.Option[sc.Sequence[sc.U8]], error) {
	args := m.Called()
	if args.Get(1) == nil {
		return args.Get(0).(sc.Option[sc.Sequence[sc.U8]]), nil
	}

	return args.Get(0).(sc.Option[sc.Sequence[sc.U8]]), args.Get(1).(error)
}

// ScheduleCodeUpgrade contains logic for parachain upgrade functionality.
func (m *ParachainSystemModule) ScheduleCodeUpgrade(code sc.Sequence[sc.U8]) error {
	args := m.Called(code)
	if args.Get(0) == nil {
		return nil
	}

	return args.Get(0).(error)
}

func (m *ParachainSystemModule) CollectCollationInfo(header primitives.Header) (parachain.CollationInfo, error) {
	args := m.Called(header)
	if args.Get(1) == nil {
		return args.Get(0).(parachain.CollationInfo), nil
	}

	return args.Get(0).(parachain.CollationInfo), args.Get(1).(error)
}

func (m *ParachainSystemModule) Metadata() primitives.MetadataModule {
	args := m.Called()
	return args.Get(0).(primitives.MetadataModule)
}
