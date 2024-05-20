package mocks

import (
	"bytes"

	sc "github.com/LimeChain/goscale"
	sessiontypes "github.com/LimeChain/gosemble/primitives/session"
	"github.com/LimeChain/gosemble/primitives/types"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/mock"
)

type SessionModule struct {
	mock.Mock
}

func (m *SessionModule) CurrentIndex() (sc.U32, error) {
	args := m.Called()

	if args.Get(1) == nil {
		return args.Get(0).(sc.U32), nil
	}

	return args.Get(0).(sc.U32), args.Get(1).(error)
}

func (m *SessionModule) Validators() (sc.Sequence[primitives.AccountId], error) {
	args := m.Called()

	if args.Get(1) == nil {
		return args.Get(0).(sc.Sequence[primitives.AccountId]), nil
	}

	return args.Get(0).(sc.Sequence[primitives.AccountId]), args.Get(1).(error)

}

func (m *SessionModule) IsDisabled(index sc.U32) (bool, error) {
	args := m.Called(index)

	if args.Get(1) == nil {
		return args.Bool(0), nil
	}

	return args.Bool(0), args.Get(1).(error)
}

func (m *SessionModule) DecodeKeys(buffer *bytes.Buffer) (sc.FixedSequence[primitives.Sr25519PublicKey], error) {
	args := m.Called(buffer)

	if args.Get(1) == nil {
		return args.Get(0).(sc.FixedSequence[primitives.Sr25519PublicKey]), nil
	}

	return args.Get(0).(sc.FixedSequence[primitives.Sr25519PublicKey]), args.Get(1).(error)

}

func (m *SessionModule) GetIndex() sc.U8 {
	args := m.Called()
	return args.Get(0).(sc.U8)
}

func (m *SessionModule) Functions() map[sc.U8]types.Call {
	args := m.Called()
	return args.Get(0).(map[sc.U8]types.Call)
}

func (m *SessionModule) PreDispatch(call types.Call) (sc.Empty, error) {
	args := m.Called(call)

	if args.Get(1) == nil {
		return args.Get(0).(sc.Empty), nil
	}

	return args.Get(0).(sc.Empty), args.Get(1).(error)
}

func (m *SessionModule) ValidateUnsigned(txSource types.TransactionSource, call types.Call) (types.ValidTransaction, error) {
	args := m.Called(txSource, call)

	if args.Get(1) == nil {
		return args.Get(0).(types.ValidTransaction), nil
	}

	return args.Get(0).(primitives.ValidTransaction), args.Get(1).(error)
}

func (m *SessionModule) Metadata() types.MetadataModule {
	args := m.Called()
	return args.Get(0).(types.MetadataModule)
}

func (m *SessionModule) CreateInherent(inherent types.InherentData) (sc.Option[types.Call], error) {
	args := m.Called(inherent)

	if args.Get(1) == nil {
		return args.Get(0).(sc.Option[types.Call]), nil
	}

	return args.Get(0).(sc.Option[types.Call]), args.Get(1).(error)
}

func (m *SessionModule) CheckInherent(call types.Call, data types.InherentData) error {
	args := m.Called(call, data)
	return args.Get(0).(error)
}

func (m *SessionModule) InherentIdentifier() [8]byte {
	args := m.Called()
	return args.Get(0).([8]byte)
}

func (m *SessionModule) IsInherent(call types.Call) bool {
	args := m.Called(call)
	return args.Get(0).(bool)
}

func (m *SessionModule) OnInitialize(n sc.U64) (primitives.Weight, error) {
	args := m.Called(n)

	if args.Get(1) == nil {
		return args.Get(0).(primitives.Weight), nil
	}

	return args.Get(0).(primitives.Weight), args.Get(1).(error)
}

func (m *SessionModule) OnRuntimeUpgrade() primitives.Weight {
	args := m.Called()
	return args.Get(0).(primitives.Weight)
}

func (m *SessionModule) OnFinalize(n sc.U64) error {
	args := m.Called(n)
	return args.Get(0).(error)
}

func (m *SessionModule) OnIdle(n sc.U64, remainingWeight primitives.Weight) primitives.Weight {
	args := m.Called(n, remainingWeight)
	return args.Get(0).(primitives.Weight)
}

func (m *SessionModule) OffchainWorker(n sc.U64) {
	m.Called(n)
}

func (m *SessionModule) AppendHandlers(module sessiontypes.OneSessionHandler) {
	m.Called(module)
}

type FindAccountFromAuthorIndex struct {
	mock.Mock
}

func (m *FindAccountFromAuthorIndex) FindAuthor(digests sc.Sequence[primitives.DigestPreRuntime]) (sc.Option[primitives.AccountId], error) {
	args := m.Called(digests)

	if args.Get(1) == nil {
		return args.Get(0).(sc.Option[primitives.AccountId]), nil
	}

	return args.Get(0).(sc.Option[primitives.AccountId]), args.Get(1).(error)
}
