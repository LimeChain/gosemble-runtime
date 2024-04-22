package mocks

import (
	"bytes"

	sc "github.com/LimeChain/goscale"
	babetypes "github.com/LimeChain/gosemble/primitives/babe"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/mock"
)

type BabeModule struct {
	mock.Mock
}

func (m *BabeModule) GetIndex() sc.U8 {
	args := m.Called()
	return args.Get(0).(sc.U8)
}

func (m *BabeModule) Functions() map[sc.U8]primitives.Call {
	args := m.Called()
	return args.Get(0).(map[sc.U8]primitives.Call)
}

func (m *BabeModule) PreDispatch(call primitives.Call) (sc.Empty, error) {
	args := m.Called(call)
	return args.Get(0).(sc.Empty), args.Get(1).(error)
}

func (m *BabeModule) ValidateUnsigned(txSource primitives.TransactionSource, call primitives.Call) (primitives.ValidTransaction, error) {
	args := m.Called(txSource, call)
	return args.Get(0).(primitives.ValidTransaction), args.Get(1).(error)
}

func (m *BabeModule) KeyType() primitives.PublicKeyType {
	args := m.Called()
	return args.Get(0).(primitives.PublicKeyType)
}

func (m *BabeModule) KeyTypeId() [4]byte {
	args := m.Called()
	return args.Get(0).([4]byte)
}

func (m *BabeModule) DecodeKey(buffer *bytes.Buffer) (primitives.Sr25519PublicKey, error) {
	args := m.Called(buffer)

	if args.Error(1) == nil {
		return args.Get(0).(primitives.Sr25519PublicKey), nil
	}

	return args.Get(0).(primitives.Sr25519PublicKey), args.Error(1)
}

func (m *BabeModule) OnGenesisSession(validators sc.Sequence[primitives.Validator]) error {
	args := m.Called(validators)
	return args.Error(0)
}

func (m *BabeModule) OnNewSession(changed bool, validators sc.Sequence[primitives.Validator], queuedValidators sc.Sequence[primitives.Validator]) error {
	args := m.Called(changed, validators, queuedValidators)
	return args.Error(0)
}

func (m *BabeModule) OnBeforeSessionEnding() {
	m.Called()
}

func (m *BabeModule) OnDisabled(validatorIndex sc.U32) {
	m.Called(validatorIndex)
}

func (m *BabeModule) OnInitialize(n sc.U64) (primitives.Weight, error) {
	args := m.Called(n)

	if args.Get(1) == nil {
		return args.Get(0).(primitives.Weight), nil
	}

	return args.Get(0).(primitives.Weight), args.Get(1).(error)
}

// func (m *BabeModule) FindAuthor(digests sc.Sequence[primitives.DigestPreRuntime]) (sc.Option[sc.U32], error) {
// 	args := m.Called(digests)

// 	if args.Error(1) == nil {
// 		return args.Get(0).(sc.Option[sc.U32]), nil
// 	}

// 	return args.Get(0).(sc.Option[sc.U32]), args.Error(1)
// }

func (m *BabeModule) OnTimestampSet(now sc.U64) error {
	args := m.Called(now)

	if args.Error(0) == nil {
		return nil
	}

	return args.Error(0)
}

func (m *BabeModule) Metadata() primitives.MetadataModule {
	args := m.Called()
	return args.Get(0).(primitives.MetadataModule)
}

func (m *BabeModule) StorageCurrentSlot() (sc.U64, error) {
	args := m.Called()

	if args.Error(1) == nil {
		return args.Get(0).(sc.U64), nil
	}

	return args.Get(0).(sc.U64), args.Error(1)
}

func (m *BabeModule) StorageAuthoritiesBytes() (sc.Option[sc.Sequence[sc.U8]], error) {
	args := m.Called()

	if args.Get(1) == nil {
		return args.Get(0).(sc.Option[sc.Sequence[sc.U8]]), nil
	}

	return args.Get(0).(sc.Option[sc.Sequence[sc.U8]]), args.Error(1)
}

func (m *BabeModule) CreateInherent(inherent primitives.InherentData) (sc.Option[primitives.Call], error) {
	args := m.Called(inherent)

	if args.Get(1) == nil {
		return args.Get(0).(sc.Option[primitives.Call]), nil
	}

	return args.Get(0).(sc.Option[primitives.Call]), args.Error(1)
}

func (m *BabeModule) CheckInherent(call primitives.Call, data primitives.InherentData) error {
	args := m.Called(call, data)
	return args.Get(0).(error)
}

func (m *BabeModule) InherentIdentifier() [8]byte {
	args := m.Called()
	return args.Get(0).([8]byte)
}

func (m *BabeModule) IsInherent(call primitives.Call) bool {
	args := m.Called(call)
	return args.Get(0).(bool)
}

func (m *BabeModule) OnRuntimeUpgrade() primitives.Weight {
	args := m.Called()
	return args.Get(0).(primitives.Weight)
}

func (m *BabeModule) OnFinalize(n sc.U64) error {
	args := m.Called()

	if args.Get(0) == nil {
		return nil
	}

	return args.Get(0).(error)
}

func (m *BabeModule) OnIdle(n sc.U64, remainingWeight primitives.Weight) primitives.Weight {
	args := m.Called(n, remainingWeight)
	return args.Get(0).(primitives.Weight)
}

func (m *BabeModule) OffchainWorker(n sc.U64) {
	m.Called(n)
}

func (m *BabeModule) StorageAuthorities() (sc.Sequence[primitives.Authority], error) {
	args := m.Called()

	if args.Error(1) == nil {
		return args.Get(0).(sc.Sequence[primitives.Authority]), nil
	}

	return args.Get(0).(sc.Sequence[primitives.Authority]), args.Error(1)
}

func (m *BabeModule) StorageRandomness() (babetypes.Randomness, error) {
	args := m.Called()

	if args.Error(1) == nil {
		return args.Get(0).(babetypes.Randomness), nil
	}

	return args.Get(0).(babetypes.Randomness), args.Error(1)
}

func (m *BabeModule) StorageSegmentIndexSet(sc.U32) {
	m.Called()
}

func (m *BabeModule) StorageEpochConfig() (babetypes.EpochConfiguration, error) {
	args := m.Called()

	if args.Error(1) == nil {
		return args.Get(0).(babetypes.EpochConfiguration), nil
	}

	return args.Get(0).(babetypes.EpochConfiguration), args.Error(1)
}

func (m *BabeModule) StorageEpochConfigSet(value babetypes.EpochConfiguration) {
	m.Called(value)
}

func (m *BabeModule) EnactEpochChange(authorities sc.Sequence[primitives.Authority], nextAuthorities sc.Sequence[primitives.Authority], sessionIndex sc.Option[sc.U32]) error {
	args := m.Called(authorities, nextAuthorities, sessionIndex)
	return args.Error(0)
}

func (m *BabeModule) ShouldEpochChange(now sc.U64) bool {
	args := m.Called(now)
	return args.Get(0).(bool)
}

func (m *BabeModule) SlotDuration() sc.U64 {
	args := m.Called()
	return args.Get(0).(sc.U64)
}

func (m *BabeModule) EpochDuration() sc.U64 {
	args := m.Called()
	return args.Get(0).(sc.U64)
}

func (m *BabeModule) EpochConfig() babetypes.EpochConfiguration {
	args := m.Called()
	return args.Get(0).(babetypes.EpochConfiguration)
}

func (m *BabeModule) EpochStartSlot(epochIndex sc.U64, genesisSlot babetypes.Slot, epochDuration sc.U64) (babetypes.Slot, error) {
	args := m.Called(epochIndex, genesisSlot, epochDuration)

	if args.Error(1) == nil {
		return args.Get(0).(babetypes.Slot), nil
	}

	return args.Get(0).(babetypes.Slot), args.Error(1)
}

func (m *BabeModule) CurrentEpochStart() (babetypes.Slot, error) {
	args := m.Called()

	if args.Error(1) == nil {
		return args.Get(0).(babetypes.Slot), nil
	}

	return args.Get(0).(babetypes.Slot), args.Error(1)
}

func (m *BabeModule) CurrentEpoch() (babetypes.Epoch, error) {
	args := m.Called()

	if args.Error(1) == nil {
		return args.Get(0).(babetypes.Epoch), nil
	}

	return args.Get(0).(babetypes.Epoch), args.Error(1)
}

func (m *BabeModule) NextEpoch() (babetypes.Epoch, error) {
	args := m.Called()

	if args.Error(1) == nil {
		return args.Get(0).(babetypes.Epoch), nil
	}

	return args.Get(0).(babetypes.Epoch), args.Error(1)
}
