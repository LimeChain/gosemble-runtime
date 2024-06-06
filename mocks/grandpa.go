package mocks

import (
	sc "github.com/LimeChain/goscale"
	grandpatypes "github.com/LimeChain/gosemble/primitives/grandpa"
	"github.com/LimeChain/gosemble/primitives/types"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/mock"
)

type GrandpaModule struct {
	mock.Mock
}

func (m *GrandpaModule) GetIndex() sc.U8 {
	args := m.Called()
	return args.Get(0).(sc.U8)
}

func (m *GrandpaModule) Functions() map[sc.U8]primitives.Call {
	args := m.Called()
	return args.Get(0).(map[sc.U8]primitives.Call)
}

func (m *GrandpaModule) PreDispatch(call primitives.Call) (sc.Empty, error) {
	args := m.Called(call)
	return args.Get(0).(sc.Empty), args.Get(1).(error)
}

func (m *GrandpaModule) ValidateUnsigned(txSource primitives.TransactionSource, call primitives.Call) (primitives.ValidTransaction, error) {
	args := m.Called(txSource, call)
	return args.Get(0).(primitives.ValidTransaction), args.Get(1).(error)
}

func (m *GrandpaModule) KeyType() primitives.PublicKeyType {
	args := m.Called()
	return args.Get(0).(primitives.PublicKeyType)
}

func (m *GrandpaModule) KeyTypeId() [4]byte {
	args := m.Called()
	return args.Get(0).([4]byte)
}

func (m *GrandpaModule) OnInitialize(n sc.U64) (primitives.Weight, error) {
	args := m.Called(n)
	if args.Get(1) == nil {
		return args.Get(0).(primitives.Weight), nil
	}
	return args.Get(0).(primitives.Weight), args.Get(1).(error)
}

func (m *GrandpaModule) Metadata() primitives.MetadataModule {
	args := m.Called()
	return args.Get(0).(primitives.MetadataModule)
}

func (m *GrandpaModule) Authorities() (sc.Sequence[primitives.Authority], error) {
	args := m.Called()
	if args.Get(1) == nil {
		return args.Get(0).(sc.Sequence[primitives.Authority]), nil
	}
	return args.Get(0).(sc.Sequence[primitives.Authority]), args.Get(1).(error)
}

func (m *GrandpaModule) StorageSetId() (sc.U64, error) {
	args := m.Called()

	if args.Get(1) == nil {
		return args.Get(0).(sc.U64), nil
	}

	return args.Get(0).(sc.U64), args.Get(1).(error)
}

func (m *GrandpaModule) HistoricalKeyOwnershipProof(authorityId primitives.AccountId) sc.Option[grandpatypes.OpaqueKeyOwnershipProof] {
	args := m.Called(authorityId)
	return args.Get(0).(sc.Option[grandpatypes.OpaqueKeyOwnershipProof])
}

func (m *GrandpaModule) SubmitUnsignedEquivocationReport(equivocationProof grandpatypes.EquivocationProof, keyOwnerProof grandpatypes.KeyOwnerProof) error {
	args := m.Called(equivocationProof, keyOwnerProof)
	return args.Get(0).(error)
}

func (m *GrandpaModule) CreateInherent(inherent types.InherentData) (sc.Option[types.Call], error) {
	args := m.Called(inherent)
	if args.Get(1) == nil {
		return args.Get(0).(sc.Option[types.Call]), nil
	}
	return args.Get(0).(sc.Option[types.Call]), args.Get(1).(error)
}

func (m *GrandpaModule) CheckInherent(call types.Call, data types.InherentData) error {
	args := m.Called(call, data)
	return args.Get(0).(error)
}

func (m *GrandpaModule) InherentIdentifier() [8]byte {
	args := m.Called()
	return args.Get(0).([8]byte)
}

func (m *GrandpaModule) IsInherent(call types.Call) bool {
	args := m.Called(call)
	return args.Get(0).(bool)
}

func (m *GrandpaModule) OnRuntimeUpgrade() primitives.Weight {
	args := m.Called()
	return args.Get(0).(primitives.Weight)
}

func (m *GrandpaModule) OnFinalize(n sc.U64) error {
	args := m.Called(n)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(error)
}

func (m *GrandpaModule) OnIdle(n sc.U64, remainingWeight primitives.Weight) primitives.Weight {
	args := m.Called(n, remainingWeight)
	return args.Get(0).(primitives.Weight)
}

func (m *GrandpaModule) OffchainWorker(n sc.U64) {
	m.Called(n)
}

func (m *GrandpaModule) StorageSetIdSessionGet(key sc.U64) (sc.U32, error) {
	args := m.Called(key)

	if args.Get(1) == nil {
		return args.Get(0).(sc.U32), nil
	}

	return args.Get(0).(sc.U32), args.Get(1).(error)
}
