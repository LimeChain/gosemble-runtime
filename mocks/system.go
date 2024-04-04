package mocks

import (
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/primitives/types"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/mock"
)

type SystemModule struct {
	mock.Mock
}

func (m *SystemModule) CreateInherent(inherent primitives.InherentData) (sc.Option[primitives.Call], error) {
	args := m.Called(inherent)
	if args.Get(1) == nil {
		return args.Get(0).(sc.Option[primitives.Call]), nil
	}
	return args.Get(0).(sc.Option[primitives.Call]), args.Get(1).(error)
}

func (m *SystemModule) CheckInherent(call primitives.Call, data primitives.InherentData) error {
	args := m.Called(call, data)
	return args.Get(0).(error)
}

func (m *SystemModule) InherentIdentifier() [8]byte {
	args := m.Called()
	return args.Get(0).([8]byte)
}

func (m *SystemModule) IsInherent(call primitives.Call) bool {
	args := m.Called(call)
	return args.Get(0).(bool)
}

func (m *SystemModule) OnInitialize(n sc.U64) (primitives.Weight, error) {
	args := m.Called(n)
	if args.Get(1) == nil {
		return args.Get(0).(primitives.Weight), nil
	}
	return args.Get(0).(primitives.Weight), args.Get(1).(error)
}

func (m *SystemModule) OnRuntimeUpgrade() primitives.Weight {
	args := m.Called()
	return args.Get(0).(primitives.Weight)
}

func (m *SystemModule) OnFinalize(n sc.U64) error {
	m.Called(n)
	return nil
}

func (m *SystemModule) OnIdle(n sc.U64, remainingWeight primitives.Weight) primitives.Weight {
	args := m.Called(n, remainingWeight)
	return args.Get(0).(primitives.Weight)
}

func (m *SystemModule) OffchainWorker(n sc.U64) {
	m.Called(n)
}

func (m *SystemModule) GetIndex() sc.U8 {
	args := m.Called()
	return args.Get(0).(sc.U8)
}

func (m *SystemModule) Functions() map[sc.U8]primitives.Call {
	args := m.Called()
	return args.Get(0).(map[sc.U8]primitives.Call)
}

func (m *SystemModule) PreDispatch(call primitives.Call) (sc.Empty, error) {
	args := m.Called(call)
	return args.Get(0).(sc.Empty), args.Get(1).(error)
}

func (m *SystemModule) ValidateUnsigned(txSource primitives.TransactionSource, call primitives.Call) (primitives.ValidTransaction, error) {
	args := m.Called(txSource, call)
	return args.Get(0).(primitives.ValidTransaction), args.Get(1).(error)
}

func (m *SystemModule) Initialize(blockNumber sc.U64, parentHash primitives.Blake2bHash, digest primitives.Digest) {
	m.Called(blockNumber, parentHash, digest)
}

func (m *SystemModule) RegisterExtraWeightUnchecked(weight primitives.Weight, class primitives.DispatchClass) error {
	args := m.Called(weight, class)
	if args.Get(0) == nil {
		return nil
	}

	return args.Get(0).(error)
}

func (m *SystemModule) NoteFinishedInitialize() {
	m.Called()
}

func (m *SystemModule) NoteExtrinsic(encodedExt []byte) error {
	m.Called(encodedExt)
	return nil
}

func (m *SystemModule) NoteAppliedExtrinsic(postInfo primitives.PostDispatchInfo, postDispatchErr error, info primitives.DispatchInfo) error {
	args := m.Called(postInfo, postDispatchErr, info)

	if args.Get(0) == nil {
		return nil
	}

	return args.Get(0).(error)
}

func (m *SystemModule) Finalize() (primitives.Header, error) {
	args := m.Called()
	if args.Get(1) == nil {
		return args.Get(0).(primitives.Header), nil
	}
	return args.Get(0).(primitives.Header), args.Get(1).(error)
}

func (m *SystemModule) NoteFinishedExtrinsics() error {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}

	return args.Get(0).(error)
}

func (m *SystemModule) ResetEvents() {
	m.Called()
}

func (m *SystemModule) Get(key primitives.AccountId) (primitives.AccountInfo, error) {
	args := m.Called(key)
	if args.Get(1) == nil {
		return args.Get(0).(primitives.AccountInfo), nil
	}
	return args.Get(0).(primitives.AccountInfo), args.Get(1).(error)
}

func (m *SystemModule) CanDecProviders(who primitives.AccountId) (bool, error) {
	args := m.Called(who)
	if args.Get(1) == nil {
		return args.Get(0).(bool), nil
	}
	return args.Get(0).(bool), args.Get(1).(error)
}

func (m *SystemModule) DepositEvent(event primitives.Event) {
	m.Called(event)
}

func (m *SystemModule) DepositLog(item primitives.DigestItem) {
	m.Called(item)
}

func (m *SystemModule) Mutate(who primitives.AccountId, f func(who *primitives.AccountInfo) (sc.Encodable, error)) (sc.Encodable, error) {
	args := m.Called(who, f)
	if args[1] == nil {
		return args[0].(sc.Encodable), nil
	}

	return args[0].(sc.Encodable), args[1].(error)
}

func (m *SystemModule) TryMutateExists(who primitives.AccountId, f func(who *primitives.AccountData) (sc.Encodable, error)) (sc.Encodable, error) {
	args := m.Called(who, f)
	if args.Get(1) == nil {
		return args.Get(0).(sc.Encodable), nil
	}
	return args.Get(0).(sc.Encodable), args.Get(1).(error)
}

func (m *SystemModule) AccountTryMutateExists(who primitives.AccountId, f func(who *primitives.AccountInfo) (sc.Encodable, error)) (sc.Encodable, error) {
	args := m.Called(who, f)
	if args[1] == nil {
		return args[0].(sc.Encodable), nil
	}

	return args[0].(sc.Encodable), args[1].(error)
}

func (m *SystemModule) CanIncConsumer(who primitives.AccountId) (bool, error) {
	args := m.Called(who)
	if args[1] == nil {
		return args[0].(bool), nil
	}

	return args[0].(bool), args[1].(error)
}

func (m *SystemModule) DecConsumers(who primitives.AccountId) error {
	args := m.Called(who)
	if args[0] == nil {
		return nil
	}

	return args[0].(error)
}

func (m *SystemModule) IncConsumers(who primitives.AccountId) error {
	args := m.Called(who)
	if args[0] == nil {
		return nil
	}

	return args[0].(error)
}

func (m *SystemModule) IncConsumersWithoutLimit(who primitives.AccountId) error {
	args := m.Called(who)
	if args[0] == nil {
		return nil
	}

	return args[0].(error)
}

func (m *SystemModule) IncProviders(who primitives.AccountId) (primitives.IncRefStatus, error) {
	args := m.Called(who)
	if args[1] == nil {
		return args[0].(primitives.IncRefStatus), nil
	}

	return args[0].(primitives.IncRefStatus), args[1].(error)
}

func (m *SystemModule) Metadata() primitives.MetadataModule {
	args := m.Called()
	return args.Get(0).(primitives.MetadataModule)
}

func (m *SystemModule) CanSetCode(codeBlob sc.Sequence[sc.U8]) error {
	args := m.Called(codeBlob)

	if args.Get(0) == nil {
		return nil
	}

	return args.Get(0).(error)
}

func (m *SystemModule) DoAuthorizeUpgrade(codeHash primitives.H256, checkVersion sc.Bool) {
	m.Called(codeHash, checkVersion)
}

func (m *SystemModule) DoApplyAuthorizeUpgrade(code sc.Sequence[sc.U8]) (primitives.PostDispatchInfo, error) {
	args := m.Called(code)

	if args.Get(1) == nil {
		return args.Get(0).(primitives.PostDispatchInfo), nil
	}

	return args.Get(0).(primitives.PostDispatchInfo), args.Get(1).(error)
}

func (m *SystemModule) errorsDefinition() *primitives.MetadataTypeDefinition {
	args := m.Called()
	return args.Get(0).(*primitives.MetadataTypeDefinition)
}

func (m *SystemModule) RuntimeUpgrade() bool {
	args := m.Called()
	return args.Get(0).(bool)
}

func (m *SystemModule) BlockWeights() types.BlockWeights {
	args := m.Called()
	return args.Get(0).(types.BlockWeights)
}

func (m *SystemModule) BlockLength() types.BlockLength {
	args := m.Called()
	return args.Get(0).(types.BlockLength)
}

func (m *SystemModule) Version() types.RuntimeVersion {
	args := m.Called()
	return args.Get(0).(types.RuntimeVersion)
}

func (m *SystemModule) DbWeight() types.RuntimeDbWeight {
	args := m.Called()
	return args.Get(0).(types.RuntimeDbWeight)
}

func (m *SystemModule) BlockHashCount() types.BlockHashCount {
	args := m.Called()
	return args.Get(0).(types.BlockHashCount)
}

func (m *SystemModule) StorageDigest() (types.Digest, error) {
	args := m.Called()
	return args.Get(0).(types.Digest), nil
}

func (m *SystemModule) StorageBlockWeight() (primitives.ConsumedWeight, error) {
	args := m.Called()
	if args.Get(1) == nil {
		return args.Get(0).(primitives.ConsumedWeight), nil
	}
	return args.Get(0).(primitives.ConsumedWeight), args.Get(1).(error)
}

func (m *SystemModule) StorageBlockWeightSet(weight primitives.ConsumedWeight) {
	m.Called(weight)
}

func (m *SystemModule) StorageBlockHash(key sc.U64) (types.Blake2bHash, error) {
	args := m.Called(key)
	if args.Get(1) == nil {
		return args.Get(0).(types.Blake2bHash), nil
	}
	return args.Get(0).(types.Blake2bHash), args.Get(1).(error)
}

func (m *SystemModule) StorageBlockHashSet(key sc.U64, value types.Blake2bHash) {
	m.Called(key, value)
}

func (m *SystemModule) StorageBlockHashExists(key sc.U64) bool {
	args := m.Called(key)

	return args.Get(0).(bool)
}

func (m *SystemModule) StorageBlockNumber() (sc.U64, error) {
	args := m.Called()
	if args.Get(1) == nil {
		return args.Get(0).(sc.U64), nil
	}
	return args.Get(0).(sc.U64), args.Get(1).(error)
}

func (m *SystemModule) StorageBlockNumberSet(blockNumber sc.U64) {
	m.Called(blockNumber)
}

func (m *SystemModule) StorageLastRuntimeUpgrade() (types.LastRuntimeUpgradeInfo, error) {
	args := m.Called()
	if args.Get(1) == nil {
		return args.Get(0).(types.LastRuntimeUpgradeInfo), nil
	}
	return args.Get(0).(types.LastRuntimeUpgradeInfo), args.Get(1).(error)
}

func (m *SystemModule) StorageLastRuntimeUpgradeSet(lrui types.LastRuntimeUpgradeInfo) {
	m.Called(lrui)
}

func (m *SystemModule) StorageAccount(key types.AccountId) (types.AccountInfo, error) {
	args := m.Called(key)
	if args.Get(1) == nil {
		return args.Get(0).(types.AccountInfo), nil
	}
	return args.Get(0).(types.AccountInfo), args.Get(1).(error)
}

func (m *SystemModule) StorageAccountSet(key types.AccountId, value types.AccountInfo) {
	m.Called(key, value)
}

func (m *SystemModule) StorageAllExtrinsicsLen() (sc.U32, error) {
	args := m.Called()
	if args.Get(1) == nil {
		return args.Get(0).(sc.U32), nil
	}
	return args.Get(0).(sc.U32), args.Get(1).(error)
}

func (m *SystemModule) StorageAllExtrinsicsLenSet(value sc.U32) {
	m.Called(value)
}

func (m *SystemModule) StorageCodeSet(codeBlob sc.Sequence[sc.U8]) {
	m.Called(codeBlob)
}
