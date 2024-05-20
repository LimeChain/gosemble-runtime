package system

import (
	"bytes"
	"math"
	"reflect"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants"
	"github.com/LimeChain/gosemble/constants/metadata"
	execTypes "github.com/LimeChain/gosemble/execution/types"
	"github.com/LimeChain/gosemble/hooks"
	"github.com/LimeChain/gosemble/primitives/io"
	"github.com/LimeChain/gosemble/primitives/log"
	"github.com/LimeChain/gosemble/primitives/types"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

const (
	functionRemarkIndex = iota
	functionSetHeapPagesIndex
	functionSetCodeIndex
	functionSetCodeWithoutChecksIndex
	functionSetStorageIndex
	functionKillStorageIndex
	functionKillPrefixIndex
	functionRemarkWithEventIndex
	functionDoTaskIndex
	functionAuthorizeUpgradeIndex
	functionAuthorizeUpgradeWithoutChecksIndex
	functionApplyAuthorizedUpgradeIndex
)

const (
	name = sc.Str("System")
)

type Module interface {
	primitives.Module

	CodeUpgrader
	LogDepositor
	primitives.EventDepositor

	Initialize(blockNumber sc.U64, parentHash primitives.Blake2bHash, digest primitives.Digest)
	RegisterExtraWeightUnchecked(weight primitives.Weight, class primitives.DispatchClass) error
	NoteFinishedInitialize()
	NoteExtrinsic(encodedExt []byte) error
	NoteAppliedExtrinsic(postInfo primitives.PostDispatchInfo, postDispatchErr error, info primitives.DispatchInfo) error
	Finalize() (primitives.Header, error)
	NoteFinishedExtrinsics() error
	ResetEvents()
	Get(key primitives.AccountId) (primitives.AccountInfo, error)
	CanDecProviders(who primitives.AccountId) (bool, error)
	CanIncConsumer(who primitives.AccountId) (bool, error)
	DecConsumers(who primitives.AccountId) error
	IncConsumers(who primitives.AccountId) error
	IncConsumersWithoutLimit(who primitives.AccountId) error
	IncProviders(who primitives.AccountId) (primitives.IncRefStatus, error)

	TryMutateExists(who primitives.AccountId, f func(who *primitives.AccountData) (sc.Encodable, error)) (sc.Encodable, error)
	Metadata() primitives.MetadataModule

	BlockHashCount() types.BlockHashCount
	BlockLength() types.BlockLength
	BlockWeights() types.BlockWeights
	DbWeight() types.RuntimeDbWeight
	Version() types.RuntimeVersion

	StorageDigest() (types.Digest, error)

	StorageBlockWeight() (primitives.ConsumedWeight, error)
	StorageBlockWeightSet(weight primitives.ConsumedWeight)

	StorageBlockHash(key sc.U64) (types.Blake2bHash, error)
	StorageBlockHashSet(key sc.U64, value types.Blake2bHash)
	StorageBlockHashExists(key sc.U64) bool

	StorageBlockNumber() (sc.U64, error)
	StorageBlockNumberSet(sc.U64)

	StorageLastRuntimeUpgrade() (types.LastRuntimeUpgradeInfo, error)
	StorageLastRuntimeUpgradeSet(lrui types.LastRuntimeUpgradeInfo)

	StorageAccount(key types.AccountId) (types.AccountInfo, error)
	StorageAccountSet(key types.AccountId, value types.AccountInfo)

	StorageAllExtrinsicsLen() (sc.U32, error)
	StorageAllExtrinsicsLenSet(value sc.U32)

	StorageCodeSet(codeBlob sc.Sequence[sc.U8])
}

type module struct {
	primitives.DefaultInherentProvider
	hooks.DefaultDispatchModule
	OnSetCode hooks.OnSetCode

	Index       sc.U8
	Config      *Config
	storage     *storage
	constants   *consts
	functions   map[sc.U8]primitives.Call
	trie        io.Trie
	ioStorage   io.Storage
	ioMisc      io.Misc
	ioHashing   io.Hashing
	logger      log.Logger
	mdGenerator *primitives.MetadataTypeGenerator
}

func New(index sc.U8, config *Config, mdGenerator *primitives.MetadataTypeGenerator, logger log.Logger) Module {
	functions := make(map[sc.U8]primitives.Call)
	storage := newStorage()
	constants := newConstants(config.BlockHashCount, config.BlockWeights, config.BlockLength, config.DbWeight, *config.Version)
	ioStorage := io.NewStorage()
	ioHashing := io.NewHashing()

	moduleInstance := module{
		Index:       index,
		Config:      config,
		storage:     storage,
		constants:   constants,
		functions:   functions,
		trie:        io.NewTrie(),
		ioStorage:   ioStorage,
		ioHashing:   ioHashing,
		ioMisc:      io.NewMisc(),
		mdGenerator: mdGenerator,
		logger:      logger,
	}

	// TODO: pass it from the constructor
	defaultOnSetCode := NewDefaultOnSetCode(moduleInstance)
	moduleInstance.OnSetCode = defaultOnSetCode

	functions[functionRemarkIndex] = newCallRemark(index, functionRemarkIndex)
	functions[functionSetHeapPagesIndex] = newCallSetHeapPages(index, functionSetHeapPagesIndex, storage.HeapPages, moduleInstance)
	functions[functionSetCodeIndex] = newCallSetCode(index, functionSetCodeIndex, *constants, defaultOnSetCode, moduleInstance)
	functions[functionSetCodeWithoutChecksIndex] = newCallSetCodeWithoutChecks(index, functionSetCodeWithoutChecksIndex, *constants, defaultOnSetCode)
	functions[functionSetStorageIndex] = newCallSetStorage(index, functionSetStorageIndex, ioStorage)
	functions[functionKillStorageIndex] = newCallKillStorage(index, functionKillStorageIndex, ioStorage)
	functions[functionKillPrefixIndex] = newCallKillPrefix(index, functionKillPrefixIndex, ioStorage)
	functions[functionRemarkWithEventIndex] = newCallRemarkWithEvent(index, functionRemarkWithEventIndex, ioHashing, moduleInstance)
	functions[functionAuthorizeUpgradeIndex] = newCallAuthorizeUpgrade(index, functionAuthorizeUpgradeIndex, moduleInstance)
	functions[functionAuthorizeUpgradeWithoutChecksIndex] = newCallAuthorizeUpgradeWithoutChecks(index, functionAuthorizeUpgradeWithoutChecksIndex, moduleInstance)
	functions[functionApplyAuthorizedUpgradeIndex] = newCallApplyAuthorizedUpgrade(index, functionApplyAuthorizedUpgradeIndex, moduleInstance)

	moduleInstance.functions = functions

	return moduleInstance
}

func (m module) name() sc.Str {
	return name
}

func (m module) GetIndex() sc.U8 {
	return m.Index
}

func (m module) Functions() map[sc.U8]primitives.Call {
	return m.functions
}

func (m module) PreDispatch(_ primitives.Call) (sc.Empty, error) {
	return sc.Empty{}, nil
}

func (m module) ValidateUnsigned(_ primitives.TransactionSource, call primitives.Call) (primitives.ValidTransaction, error) {
	switch call := call.(type) {
	case callApplyAuthorizedUpgrade:
		code := call.Args()[0].(sc.Sequence[sc.U8])

		hash, err := m.validateAuthorizedUpgrade(code)
		if err != nil {
			return primitives.ValidTransaction{}, primitives.NewTransactionValidityError(primitives.NewInvalidTransactionCall())
		}

		return primitives.ValidTransaction{
			Priority:  100,
			Requires:  sc.Sequence[primitives.TransactionTag]{},
			Provides:  sc.Sequence[primitives.TransactionTag]{sc.BytesToSequenceU8(sc.FixedSequenceU8ToBytes(hash.FixedSequence))},
			Longevity: primitives.TransactionLongevity(math.MaxUint64),
			Propagate: true,
		}, nil
	default:
		return primitives.ValidTransaction{}, primitives.NewTransactionValidityError(primitives.NewUnknownTransactionNoUnsignedValidator())
	}
}

func (m module) BlockHashCount() types.BlockHashCount {
	return m.constants.BlockHashCount
}

func (m module) BlockLength() types.BlockLength {
	return m.constants.BlockLength
}

func (m module) BlockWeights() types.BlockWeights {
	return m.constants.BlockWeights
}

func (m module) DbWeight() types.RuntimeDbWeight {
	return m.constants.DbWeight
}

func (m module) Version() types.RuntimeVersion {
	return m.constants.Version
}

func (m module) StorageDigest() (types.Digest, error) {
	return m.storage.Digest.Get()
}

func (m module) StorageBlockWeight() (primitives.ConsumedWeight, error) {
	return m.storage.BlockWeight.Get()
}

func (m module) StorageBlockWeightSet(weight primitives.ConsumedWeight) {
	m.storage.BlockWeight.Put(weight)
}

func (m module) StorageBlockHash(key sc.U64) (types.Blake2bHash, error) {
	return m.storage.BlockHash.Get(key)
}

func (m module) StorageBlockHashSet(key sc.U64, value types.Blake2bHash) {
	m.storage.BlockHash.Put(key, value)
}

func (m module) StorageBlockHashExists(key sc.U64) bool {
	return m.storage.BlockHash.Exists(key)
}

func (m module) StorageBlockNumber() (sc.U64, error) {
	return m.storage.BlockNumber.Get()
}

func (m module) StorageBlockNumberSet(blockNumber sc.U64) {
	m.storage.BlockNumber.Put(blockNumber)
}

func (m module) StorageLastRuntimeUpgrade() (types.LastRuntimeUpgradeInfo, error) {
	return m.storage.LastRuntimeUpgrade.Get()
}

func (m module) StorageLastRuntimeUpgradeSet(lrui types.LastRuntimeUpgradeInfo) {
	m.storage.LastRuntimeUpgrade.Put(lrui)
}

func (m module) StorageAccount(key types.AccountId) (types.AccountInfo, error) {
	return m.storage.Account.Get(key)
}

func (m module) StorageAccountSet(key types.AccountId, value types.AccountInfo) {
	m.storage.Account.Put(key, value)
}

func (m module) StorageAllExtrinsicsLen() (sc.U32, error) {
	return m.storage.AllExtrinsicsLen.Get()
}

func (m module) StorageAllExtrinsicsLenSet(value sc.U32) {
	m.storage.AllExtrinsicsLen.Put(value)
}

func (m module) StorageCodeSet(codeBlob sc.Sequence[sc.U8]) {
	m.storage.Code.Put(codeBlob)
}

func (m module) Initialize(blockNumber sc.U64, parentHash primitives.Blake2bHash, digest primitives.Digest) {
	m.storage.ExecutionPhase.Put(primitives.NewExtrinsicPhaseInitialization())
	m.storage.ExtrinsicIndex.Put(sc.U32(0))
	m.storage.BlockNumber.Put(blockNumber)
	m.storage.Digest.Put(digest)
	m.storage.ParentHash.Put(parentHash)
	m.storage.BlockHash.Put(blockNumber-1, parentHash)
	m.storage.BlockWeight.Clear()
}

// RegisterExtraWeightUnchecked - Inform the system pallet of some additional weight that should be accounted for, in the
// current block.
//
// NOTE: use with extra care; this function is made public only be used for certain pallets
// that need it. A runtime that does not have dynamic calls should never need this and should
// stick to static weights. A typical use case for this is inner calls or smart contract calls.
// Furthermore, it only makes sense to use this when it is presumably  _cheap_ to provide the
// argument `weight`; In other words, if this function is to be used to account for some
// unknown, user provided call's weight, it would only make sense to use it if you are sure you
// can rapidly compute the weight of the inner call.
//
// Even more dangerous is to note that this function does NOT take any action, if the new sum
// of block weight is more than the block weight limit. This is what the _unchecked_.
//
// Another potential use-case could be for the `on_initialize` and `on_finalize` hooks.
func (m module) RegisterExtraWeightUnchecked(weight primitives.Weight, class primitives.DispatchClass) error {
	currentWeight, err := m.storage.BlockWeight.Get()
	if err != nil {
		return err
	}
	err = currentWeight.Accrue(weight, class)
	if err != nil {
		return err
	}
	m.storage.BlockWeight.Put(currentWeight)
	return nil
}

func (m module) NoteFinishedInitialize() {
	m.storage.ExecutionPhase.Put(primitives.NewExtrinsicPhaseApply(sc.U32(0)))
}

// NoteExtrinsic - what the extrinsic data of the current extrinsic index is.
//
// This is required to be called before applying an extrinsic. The data will used
// in [`finalize`] to calculate the correct extrinsics root.
func (m module) NoteExtrinsic(encodedExt []byte) error {
	extrinsicIndex, err := m.storage.ExtrinsicIndex.Get()
	if err != nil {
		return err
	}
	m.storage.ExtrinsicData.Put(extrinsicIndex, sc.BytesToSequenceU8(encodedExt))
	return nil
}

// NoteAppliedExtrinsic - To be called immediately after an extrinsic has been applied.
//
// Emits an `ExtrinsicSuccess` or `ExtrinsicFailed` event depending on the outcome.
// The emitted event contains the post-dispatch corrected weight including
// the base-weight for its dispatch class.
func (m module) NoteAppliedExtrinsic(postInfo primitives.PostDispatchInfo, postDispatchErr error, info primitives.DispatchInfo) error {
	dispatchClass, err := m.BlockWeights().Get(info.Class)
	if err != nil {
		return err
	}

	baseWeight := dispatchClass.BaseExtrinsic
	info.Weight = postInfo.CalcActualWeight(&info).SaturatingAdd(baseWeight)
	info.PaysFee = postInfo.Pays(&info)

	if dispatchErr, ok := postDispatchErr.(primitives.DispatchError); ok {
		blockNum, err := m.StorageBlockNumber()
		m.logger.Tracef("Extrinsic failed at block(%d): {%v}", blockNum, dispatchErr)
		if err != nil {
			return err
		}
		m.logger.Tracef("Extrinsic failed at block(%d): {}", blockNum)

		m.DepositEvent(newEventExtrinsicFailed(m.Index, dispatchErr, info))
	} else {
		m.DepositEvent(newEventExtrinsicSuccess(m.Index, info))
	}

	nextExtrinsicIndex, err := m.storage.ExtrinsicIndex.Get()
	if err != nil {
		return err
	}
	nextExtrinsicIndex = nextExtrinsicIndex + 1
	m.storage.ExtrinsicIndex.Put(nextExtrinsicIndex)
	m.storage.ExecutionPhase.Put(primitives.NewExtrinsicPhaseApply(nextExtrinsicIndex))
	return nil
}

func (m module) Finalize() (primitives.Header, error) {
	m.storage.ExecutionPhase.Clear()
	m.storage.AllExtrinsicsLen.Clear()

	blockNumber, err := m.StorageBlockNumber()
	if err != nil {
		return primitives.Header{}, err
	}
	parentHash, err := m.storage.ParentHash.Get()
	if err != nil {
		return primitives.Header{}, err
	}
	digest, err := m.StorageDigest()
	if err != nil {
		return primitives.Header{}, err
	}
	extrinsicCount, err := m.storage.ExtrinsicCount.Take()
	if err != nil {
		return primitives.Header{}, err
	}

	var extrinsics []byte

	for i := 0; i < int(extrinsicCount); i++ {
		sci := sc.U32(i)

		extrinsic, err := m.storage.ExtrinsicData.TakeBytes(sci)
		if err != nil {
			return primitives.Header{}, err
		}
		extrinsics = append(extrinsics, extrinsic...)
	}

	buf := &bytes.Buffer{}
	extrinsicsRootBytes := m.trie.Blake2256OrderedRoot(
		append(sc.ToCompact(uint64(extrinsicCount)).Bytes(), extrinsics...),
		constants.StorageVersion)
	buf.Write(extrinsicsRootBytes)
	extrinsicsRoot, err := primitives.DecodeH256(buf)
	if err != nil {
		return primitives.Header{}, err
	}
	buf.Reset()

	toRemove := sc.SaturatingSubU64(blockNumber, sc.U64(m.constants.BlockHashCount.U32))
	toRemove = sc.SaturatingSubU64(toRemove, 1)
	if toRemove != 0 {
		m.storage.BlockHash.Remove(toRemove)
	}

	storageRootBytes := m.ioStorage.Root(int32(m.constants.Version.StateVersion))
	buf.Write(storageRootBytes)
	storageRoot, err := primitives.DecodeH256(buf)
	if err != nil {
		return primitives.Header{}, err
	}
	buf.Reset()

	return primitives.Header{
		ExtrinsicsRoot: extrinsicsRoot,
		StateRoot:      storageRoot,
		ParentHash:     parentHash,
		Number:         blockNumber,
		Digest:         digest,
	}, nil
}

func (m module) NoteFinishedExtrinsics() error {
	extrinsicIndex, err := m.storage.ExtrinsicIndex.Take()
	if err != nil {
		return err
	}
	m.storage.ExtrinsicCount.Put(extrinsicIndex)
	m.storage.ExecutionPhase.Put(primitives.NewExtrinsicPhaseFinalization())
	return nil
}

func (m module) ResetEvents() {
	m.storage.Events.Clear()
	m.storage.EventCount.Clear()
	m.storage.EventTopics.Clear(sc.U32(math.MaxUint32))
}

func (m module) Get(key primitives.AccountId) (primitives.AccountInfo, error) {
	return m.storage.Account.Get(key)
}

func (m module) CanDecProviders(who primitives.AccountId) (bool, error) {
	acc, err := m.Get(who)
	if err != nil {
		return false, err
	}

	return acc.Consumers == 0 || acc.Providers > 1, nil
}

// DepositEvent deposits an event into block's event record.
func (m module) DepositEvent(event primitives.Event) {
	m.depositEventIndexed([]primitives.H256{}, event)
}

// DepositLog deposits a log and ensures it matches the block's log data.
func (m module) DepositLog(item primitives.DigestItem) {
	m.storage.Digest.AppendItem(item)
}

func (m module) TryMutateExists(who primitives.AccountId, f func(*primitives.AccountData) (sc.Encodable, error)) (sc.Encodable, error) {
	account, err := m.Get(who)
	if err != nil {
		return nil, err
	}
	wasProviding := false
	if !reflect.DeepEqual(account.Data, primitives.AccountData{}) {
		wasProviding = true
	}

	someData := &primitives.AccountData{}
	if wasProviding {
		someData = &account.Data
	}

	result, err := f(someData)
	if err != nil {
		return result, err
	}

	isProviding := !reflect.DeepEqual(*someData, primitives.AccountData{})

	if !wasProviding && isProviding {
		_, err := m.IncProviders(who)
		if err != nil {
			return nil, err
		}
	} else if wasProviding && !isProviding {
		status, err := m.decProviders(who)
		if err != nil {
			return nil, err
		}
		if status == primitives.DecRefStatusExists {
			return result, nil
		}
	} else if !wasProviding && !isProviding {
		return result, nil
	}

	_, err = m.storage.Account.Mutate(who, func(a *primitives.AccountInfo) (sc.Encodable, error) {
		mutateAccount(a, someData)
		return nil, nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (m module) CanIncConsumer(who primitives.AccountId) (bool, error) {
	acc, err := m.Get(who)
	if err != nil {
		return false, err
	}

	value, err := sc.CheckedAddU32(acc.Consumers, 1)
	if err != nil {
		return false, err
	}

	return value > 0 && value <= m.Config.MaxConsumers, nil
}

func (m module) DecConsumers(who primitives.AccountId) error {
	_, err := m.storage.Account.Mutate(who, func(account *primitives.AccountInfo) (sc.Encodable, error) {
		if account.Consumers > 0 {
			account.Consumers -= 1
			return nil, nil
		}
		m.logger.Critical("Unexpected underflow in reducing consumer.")
		return nil, nil
	})

	return err
}

func (m module) IncConsumers(who primitives.AccountId) error {
	_, err := m.storage.Account.Mutate(who, func(account *primitives.AccountInfo) (sc.Encodable, error) {
		if account.Providers > 0 {
			if account.Consumers < m.Config.MaxConsumers {
				account.Consumers = sc.SaturatingAddU32(account.Consumers, 1)
				return nil, nil
			}

			return nil, primitives.NewDispatchErrorTooManyConsumers()
		}

		return nil, primitives.NewDispatchErrorNoProviders()
	})

	return err
}

func (m module) IncConsumersWithoutLimit(who primitives.AccountId) error {
	_, err := m.storage.Account.Mutate(who, func(account *primitives.AccountInfo) (sc.Encodable, error) {
		if account.Providers > 0 {
			account.Consumers = sc.SaturatingAddU32(account.Consumers, 1)
			return nil, nil
		}

		return nil, primitives.NewDispatchErrorNoProviders()
	})

	return err
}

func (m module) IncProviders(who primitives.AccountId) (primitives.IncRefStatus, error) {
	result, err := m.storage.Account.Mutate(who, func(account *primitives.AccountInfo) (sc.Encodable, error) {
		return m.incrementProviders(who, account), nil
	})

	return result.(primitives.IncRefStatus), err
}

func (m module) decrementProviders(who primitives.AccountId, maybeAccount *sc.Option[primitives.AccountInfo]) (sc.Encodable, error) {
	if maybeAccount.HasValue {
		account := &maybeAccount.Value

		if account.Providers == 0 {
			m.logger.Warn("Logic error: Unexpected underflow in reducing provider")
			account.Providers = 1
		}

		if account.Providers == 1 && account.Consumers == 0 && account.Sufficients == 0 {
			m.onKilledAccount(who)
			// No providers left (and no consumers) and no sufficients. Account dead.
			return primitives.DecRefStatusReaped, nil
		}
		if account.Providers == 1 && account.Consumers > 0 {
			// Cannot remove last provider if there are consumers.
			return nil, primitives.NewDispatchErrorConsumerRemaining()
		}
		// Account will continue to exist as there is either > 1 provider or
		// > 0 sufficients.
		account.Providers = account.Providers - 1
		return primitives.DecRefStatusExists, nil
	} else {
		m.logger.Warn("Logic error: Account already dead when reducing provider")
		return primitives.DecRefStatusReaped, nil
	}
}

func (m module) incrementProviders(who primitives.AccountId, account *primitives.AccountInfo) primitives.IncRefStatus {
	if account.Providers == 0 && account.Sufficients == 0 {
		account.Providers = 1
		m.onCreatedAccount(who)

		return primitives.IncRefStatusCreated
	} else {
		account.Providers = sc.SaturatingAddU32(account.Providers, 1)

		return primitives.IncRefStatusExisted
	}
}

func (m module) decProviders(who primitives.AccountId) (primitives.DecRefStatus, error) {
	result, err := m.storage.Account.TryMutateExists(who, func(maybeAccount *sc.Option[primitives.AccountInfo]) (sc.Encodable, error) {
		return m.decrementProviders(who, maybeAccount)
	})

	if err != nil {
		return primitives.DecRefStatus(0), err
	}

	return result.(primitives.DecRefStatus), nil
}

// depositEventIndexed Deposits an event into this block's event record adding this event
// to the corresponding topic indexes.
//
// This will update storage entries that correspond to the specified topics.
// It is expected that light-clients could subscribe to this topics.
//
// NOTE: Events not registered at the genesis block and quietly omitted.
func (m module) depositEventIndexed(topics []primitives.H256, event primitives.Event) error {
	blockNumber, err := m.StorageBlockNumber()
	if err != nil {
		return err
	}

	if blockNumber == 0 {
		return nil
	}

	phase, err := m.storage.ExecutionPhase.Get()
	if err != nil {
		return err
	}

	eventRecord := primitives.EventRecord{
		Phase:  phase,
		Event:  event,
		Topics: topics,
	}

	oldEventCount, err := m.storage.EventCount.Get()
	if err != nil {
		return err
	}

	newEventCount, err := sc.CheckedAddU32(oldEventCount, 1)
	if err != nil {
		return err
	}

	m.storage.EventCount.Put(newEventCount)
	m.storage.Events.Append(eventRecord)

	topicValue := sc.NewVaryingData(blockNumber, oldEventCount)
	for _, topic := range topics {
		m.storage.EventTopics.Append(topic, topicValue)
	}
	return nil
}

func (m module) onCreatedAccount(who primitives.AccountId) {
	// hook on creating new account, currently not used in Substrate
	//T::OnNewAccount::on_new_account(&who);
	m.DepositEvent(newEventNewAccount(m.Index, who))
}

func (m module) onKilledAccount(who primitives.AccountId) {
	m.DepositEvent(newEventKilledAccount(m.Index, who))
}

func (m module) Metadata() primitives.MetadataModule {
	// build calls metadata
	metadataIdSystemCalls := m.mdGenerator.BuildCallsMetadata(
		"System",
		m.functions,
		&sc.Sequence[primitives.MetadataTypeParameter]{
			primitives.NewMetadataEmptyTypeParameter("T"),
		},
	)

	// build errors metadata
	errorsMetadataId := m.mdGenerator.BuildErrorsMetadata(
		"System",
		m.errorsDefinition(),
	)

	m.mdGenerator.BuildMetadataTypeRecursively(reflect.ValueOf(primitives.ExtrinsicPhase{}), &sc.Sequence[sc.Str]{"frame_system", "Phase"}, new(primitives.ExtrinsicPhase).MetadataDefinition(), nil)

	m.mdGenerator.BuildMetadataTypeRecursively(
		reflect.ValueOf(execTypes.NewBlock(primitives.Header{}, sc.Sequence[primitives.UncheckedExtrinsic]{})),
		&sc.Sequence[sc.Str]{"sp_runtime", "generic", "block", "Block"}, nil, &sc.Sequence[primitives.MetadataTypeParameter]{
			primitives.NewMetadataTypeParameter(metadata.Header, "Header"),
			primitives.NewMetadataTypeParameter(metadata.UncheckedExtrinsic, "Extrinsic"),
		},
	)

	m.mdGenerator.BuildMetadataTypeRecursively(
		reflect.ValueOf(primitives.WeightsPerClass{}),
		&sc.Sequence[sc.Str]{"frame_system", "limits", "WeightsPerClass"}, nil, nil,
	)

	m.mdGenerator.BuildMetadataTypeRecursively(reflect.ValueOf(primitives.PerDispatchClassWeight{}), &sc.Sequence[sc.Str]{"frame_support", "dispatch", "PerDispatchClass"}, nil, nil)

	m.mdGenerator.BuildMetadataTypeRecursively(reflect.ValueOf(primitives.PerDispatchClassWeightsPerClass{}), &sc.Sequence[sc.Str]{"frame_support", "dispatch", "PerDispatchClass"}, nil, nil)

	m.mdGenerator.BuildMetadataTypeRecursively(reflect.ValueOf(primitives.BlockWeights{}), &sc.Sequence[sc.Str]{"frame_system", "limits", "BlockWeights"}, nil, nil)

	m.mdGenerator.BuildMetadataTypeRecursively(reflect.ValueOf(primitives.RuntimeDbWeight{}), &sc.Sequence[sc.Str]{"sp_weights", "RuntimeDbWeight"}, nil, nil)

	validTransactionMdId := m.mdGenerator.BuildMetadataTypeRecursively(reflect.ValueOf(primitives.ValidTransaction{}), &sc.Sequence[sc.Str]{"sp_runtime", "transaction_validity", "ValidTransaction"}, nil, nil)

	m.mdGenerator.BuildMetadataTypeRecursively(reflect.ValueOf(primitives.LastRuntimeUpgradeInfo{SpecVersion: sc.Compact{Number: sc.U32(0)}, SpecName: ""}), &sc.Sequence[sc.Str]{"frame_system", "LastRuntimeUpgradeInfo"}, nil, nil)

	m.mdGenerator.BuildMetadataTypeRecursively(reflect.ValueOf(primitives.TransactionSource{}), &sc.Sequence[sc.Str]{"sp_runtime", "transaction_validity", "TransactionSource"}, new(primitives.TransactionSource).MetadataDefinition(), nil)

	// type 871
	invalidTxMdId := m.mdGenerator.BuildMetadataTypeRecursively(reflect.ValueOf(primitives.InvalidTransaction{}), &sc.Sequence[sc.Str]{"sp_runtime", "transaction_validity", "InvalidTransaction"}, new(primitives.InvalidTransaction).MetadataDefinition(), nil)

	// type872
	unknownTxMdId := m.mdGenerator.BuildMetadataTypeRecursively(reflect.ValueOf(primitives.UnknownTransaction{}), &sc.Sequence[sc.Str]{"sp_runtime", "transaction_validity", "UnknownTransaction"}, new(primitives.UnknownTransaction).MetadataDefinition(), nil)

	// type 870
	validityErrorMdId := m.mdGenerator.BuildMetadataTypeRecursively(reflect.ValueOf(primitives.TransactionValidityError{}), &sc.Sequence[sc.Str]{"sp_runtime", "transaction_validity", "TransactionValidityError"}, new(primitives.TransactionValidityError).MetadataDefinition(invalidTxMdId, unknownTxMdId), nil)

	m.mdGenerator.BuildMetadataTypeRecursively(reflect.ValueOf(primitives.TransactionValidityResult{}), &sc.Sequence[sc.Str]{"Result"}, new(primitives.TransactionValidityResult).MetadataDefinition(validTransactionMdId, validityErrorMdId), nil)

	m.mdGenerator.BuildMetadataTypeRecursively(reflect.ValueOf(primitives.PerDispatchClassU32{}), &sc.Sequence[sc.Str]{"frame_support", "dispatch", "PerDispatchClass"}, nil, nil)

	m.mdGenerator.BuildMetadataTypeRecursively(reflect.ValueOf(primitives.BlockLength{}), &sc.Sequence[sc.Str]{"frame_system", "limits", "BlockLength"}, nil, nil)

	constants := m.mdGenerator.BuildModuleConstants(reflect.ValueOf(*m.constants))

	dataV14 := primitives.MetadataModuleV14{
		Name:    m.name(),
		Storage: m.metadataStorage(),
		Call:    sc.NewOption[sc.Compact](sc.ToCompact(metadataIdSystemCalls)),
		CallDef: sc.NewOption[primitives.MetadataDefinitionVariant](
			primitives.NewMetadataDefinitionVariantStr(
				m.name(),
				sc.Sequence[primitives.MetadataTypeDefinitionField]{
					primitives.NewMetadataTypeDefinitionFieldWithName(metadataIdSystemCalls, "self::sp_api_hidden_includes_construct_runtime::hidden_include::dispatch\n::CallableCallFor<System, Runtime>"),
				},
				m.Index,
				"Call.System"),
		),
		Event: sc.NewOption[sc.Compact](sc.ToCompact(metadata.TypesSystemEvent)),
		EventDef: sc.NewOption[primitives.MetadataDefinitionVariant](
			primitives.NewMetadataDefinitionVariantStr(
				m.name(),
				sc.Sequence[primitives.MetadataTypeDefinitionField]{
					primitives.NewMetadataTypeDefinitionFieldWithName(metadata.TypesSystemEvent, "frame_system::Event<Runtime>"),
				},
				m.Index,
				"Events.System"),
		),
		Constants: constants,
		Error:     sc.NewOption[sc.Compact](sc.ToCompact(errorsMetadataId)),
		ErrorDef: sc.NewOption[primitives.MetadataDefinitionVariant](
			primitives.NewMetadataDefinitionVariantStr(
				m.name(),
				sc.Sequence[primitives.MetadataTypeDefinitionField]{
					primitives.NewMetadataTypeDefinitionField(errorsMetadataId),
				},
				m.Index,
				"Errors.System"),
		),
		Index: m.Index,
	}

	m.mdGenerator.AppendMetadataTypes(m.metadataTypes())

	return primitives.MetadataModule{
		Version:   primitives.ModuleVersion14,
		ModuleV14: dataV14,
	}
}

func (m module) errorsDefinition() *primitives.MetadataTypeDefinition {
	def := primitives.NewMetadataTypeDefinitionVariant(
		sc.Sequence[primitives.MetadataDefinitionVariant]{
			primitives.NewMetadataDefinitionVariant(
				"InvalidSpecName",
				sc.Sequence[primitives.MetadataTypeDefinitionField]{},
				ErrorInvalidSpecName,
				"The name of specification does not match between the current runtime and the new runtime.",
			),
			primitives.NewMetadataDefinitionVariant(
				"SpecVersionNeedsToIncrease",
				sc.Sequence[primitives.MetadataTypeDefinitionField]{},
				ErrorSpecVersionNeedsToIncrease,
				"The specification version is not allowed to decrease between the current runtime and the new runtime.",
			),
			primitives.NewMetadataDefinitionVariant(
				"FailedToExtractRuntimeVersion",
				sc.Sequence[primitives.MetadataTypeDefinitionField]{},
				ErrorFailedToExtractRuntimeVersion,
				"Failed to extract the runtime version from the new runtime.  Either calling `Core_version` or decoding `RuntimeVersion` failed.",
			),
			primitives.NewMetadataDefinitionVariant(
				"NonDefaultComposite",
				sc.Sequence[primitives.MetadataTypeDefinitionField]{},
				ErrorNonDefaultComposite,
				"Suicide called when the account has non-default composite data.",
			),
			primitives.NewMetadataDefinitionVariant(
				"NonZeroRefCount",
				sc.Sequence[primitives.MetadataTypeDefinitionField]{},
				ErrorNonZeroRefCount,
				"There is a non-zero reference count preventing the account from being purged.",
			),
			primitives.NewMetadataDefinitionVariant(
				"CallFiltered",
				sc.Sequence[primitives.MetadataTypeDefinitionField]{},
				ErrorCallFiltered,
				"The origin filter prevent the call to be dispatched.",
			),
			primitives.NewMetadataDefinitionVariant(
				"NothingAuthorized",
				sc.Sequence[primitives.MetadataTypeDefinitionField]{},
				ErrorNothingAuthorized,
				"No upgrade authorized.",
			),
			primitives.NewMetadataDefinitionVariant(
				"Unauthorized",
				sc.Sequence[primitives.MetadataTypeDefinitionField]{},
				ErrorUnauthorized,
				"The submitted code is not authorized.",
			),
		})
	return &def
}

func (m module) metadataTypes() sc.Sequence[primitives.MetadataType] {
	typesPhaseId, _ := m.mdGenerator.GetId("ExtrinsicPhase")

	return sc.Sequence[primitives.MetadataType]{
		primitives.NewMetadataType(
			metadata.TypesSystemEventStorage,
			"Vec<Box<EventRecord<T::RuntimeEvent, T::Hash>>>",
			primitives.NewMetadataTypeDefinitionSequence(sc.ToCompact(metadata.TypesEventRecord))),

		primitives.NewMetadataType(
			metadata.TypesVecBlockNumEventIndex,
			"Vec<BlockNumber, EventIndex>",
			primitives.NewMetadataTypeDefinitionSequence(sc.ToCompact(metadata.TypesTupleU32U32))),

		primitives.NewMetadataTypeWithParams(metadata.TypesEventRecord,
			"frame_system EventRecord",
			sc.Sequence[sc.Str]{"frame_system", "EventRecord"},
			primitives.NewMetadataTypeDefinitionComposite(sc.Sequence[primitives.MetadataTypeDefinitionField]{
				primitives.NewMetadataTypeDefinitionFieldWithNames(typesPhaseId, "phase", "Phase"),
				primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesRuntimeEvent, "event", "E"),
				primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesVecTopics, "topics", "Vec<T>"),
			}),
			sc.Sequence[primitives.MetadataTypeParameter]{
				primitives.NewMetadataTypeParameter(metadata.TypesRuntimeEvent, "E"),
				primitives.NewMetadataTypeParameter(metadata.TypesH256, "T"),
			}),

		primitives.NewMetadataTypeWithPath(metadata.TypesSystemEvent,
			"frame_system pallet Event",
			sc.Sequence[sc.Str]{"frame_system", "pallet", "Event"},
			primitives.NewMetadataTypeDefinitionVariant(
				sc.Sequence[primitives.MetadataDefinitionVariant]{
					primitives.NewMetadataDefinitionVariant(
						"ExtrinsicSuccess",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesDispatchInfo, "dispatch_info", "DispatchInfo"),
						},
						EventExtrinsicSuccess,
						"Event.ExtrinsicSuccess"),
					primitives.NewMetadataDefinitionVariant(
						"ExtrinsicFailed",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesDispatchError, "dispatch_error", "DispatchError"),
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesDispatchInfo, "dispatch_info", "DispatchInfo"),
						},
						EventExtrinsicFailed,
						"Events.ExtrinsicFailed"),
					primitives.NewMetadataDefinitionVariant(
						"CodeUpdated",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{},
						EventCodeUpdated,
						"Events.CodeUpdated"),
					primitives.NewMetadataDefinitionVariant(
						"NewAccount",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesAddress32, "account", "T::AccountId"),
						},
						EventNewAccount,
						"Events.NewAccount"),
					primitives.NewMetadataDefinitionVariant(
						"KilledAccount",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesAddress32, "account", "T::AccountId"),
						},
						EventKilledAccount,
						"Events.KilledAccount"),
					primitives.NewMetadataDefinitionVariant(
						"Remarked",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesAddress32, "sender", "T::AccountId"),
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesH256, "hash", "T::Hash"),
						},
						EventRemarked,
						"Events.Remarked"),
					primitives.NewMetadataDefinitionVariant(
						"UpgradeAuthorized",
						sc.Sequence[primitives.MetadataTypeDefinitionField]{
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.TypesH256, "code_hash", "T::Hash"),
							primitives.NewMetadataTypeDefinitionFieldWithNames(metadata.PrimitiveTypesBool, "check_version", "bool"),
						},
						EventUpgradeAuthorized,
						"Events.UpgradeAuthorized"),
				})),

		primitives.NewMetadataTypeWithPath(metadata.TypesEra, "Era", sc.Sequence[sc.Str]{"sp_runtime", "generic", "era", "Era"}, primitives.NewMetadataTypeDefinitionVariant(primitives.EraTypeDefinition())),
	}
}

func (m module) metadataStorage() sc.Option[primitives.MetadataModuleStorage] {
	typesPhaseId, _ := m.mdGenerator.GetId("ExtrinsicPhase")
	perDispatchClassWeightId, _ := m.mdGenerator.GetId("PerDispatchClassWeight")
	lastRuntimeUpgradeInfoId, _ := m.mdGenerator.GetId("LastRuntimeUpgradeInfo")

	return sc.NewOption[primitives.MetadataModuleStorage](
		primitives.MetadataModuleStorage{
			Prefix: m.name(),
			Items: sc.Sequence[primitives.MetadataModuleStorageEntry]{
				primitives.NewMetadataModuleStorageEntry(
					"Account",
					primitives.MetadataModuleStorageEntryModifierDefault,
					primitives.NewMetadataModuleStorageEntryDefinitionMap(
						sc.Sequence[primitives.MetadataModuleStorageHashFunc]{primitives.MetadataModuleStorageHashFuncMultiBlake128Concat},
						sc.ToCompact(metadata.TypesAddress32),
						sc.ToCompact(metadata.TypesAccountInfo),
					),
					"The full account information for a particular account ID.",
				),
				primitives.NewMetadataModuleStorageEntry(
					"ExtrinsicCount",
					primitives.MetadataModuleStorageEntryModifierOptional,
					primitives.NewMetadataModuleStorageEntryDefinitionPlain(
						sc.ToCompact(metadata.PrimitiveTypesU32),
					),
					"Total extrinsics count for the current block.",
				),
				primitives.NewMetadataModuleStorageEntry(
					"BlockWeight",
					primitives.MetadataModuleStorageEntryModifierDefault,
					primitives.NewMetadataModuleStorageEntryDefinitionPlain(
						sc.ToCompact(perDispatchClassWeightId),
					),
					"The current weight for the block.",
				),
				primitives.NewMetadataModuleStorageEntry(
					"AllExtrinsicsLen",
					primitives.MetadataModuleStorageEntryModifierOptional,
					primitives.NewMetadataModuleStorageEntryDefinitionPlain(
						sc.ToCompact(metadata.PrimitiveTypesU32),
					),
					"Total length (in bytes) for all extrinsics put together, for the current block.",
				),
				primitives.NewMetadataModuleStorageEntry(
					"BlockHash",
					primitives.MetadataModuleStorageEntryModifierDefault,
					primitives.NewMetadataModuleStorageEntryDefinitionMap(
						sc.Sequence[primitives.MetadataModuleStorageHashFunc]{primitives.MetadataModuleStorageHashFuncMultiXX64},
						sc.ToCompact(metadata.PrimitiveTypesU32),
						sc.ToCompact(metadata.TypesFixedSequence32U8),
					),
					"Map of block numbers to block hashes.",
				),
				primitives.NewMetadataModuleStorageEntry(
					"ExtrinsicData",
					primitives.MetadataModuleStorageEntryModifierDefault,
					primitives.NewMetadataModuleStorageEntryDefinitionMap(
						sc.Sequence[primitives.MetadataModuleStorageHashFunc]{primitives.MetadataModuleStorageHashFuncMultiXX64},
						sc.ToCompact(metadata.PrimitiveTypesU32),
						sc.ToCompact(metadata.TypesSequenceU8),
					),
					"Extrinsics data for the current block (maps an extrinsic's index to its data).",
				),
				primitives.NewMetadataModuleStorageEntry(
					"Number",
					primitives.MetadataModuleStorageEntryModifierDefault,
					primitives.NewMetadataModuleStorageEntryDefinitionPlain(
						sc.ToCompact(metadata.PrimitiveTypesU32),
					),
					"The current block number being processed. Set by `execute_block`.",
				),
				primitives.NewMetadataModuleStorageEntry(
					"ParentHash",
					primitives.MetadataModuleStorageEntryModifierDefault,
					primitives.NewMetadataModuleStorageEntryDefinitionPlain(
						sc.ToCompact(metadata.TypesFixedSequence32U8),
					),
					"Hash of the previous block.",
				),
				primitives.NewMetadataModuleStorageEntry(
					"Digest",
					primitives.MetadataModuleStorageEntryModifierDefault,
					primitives.NewMetadataModuleStorageEntryDefinitionPlain(
						sc.ToCompact(metadata.TypesDigest),
					),
					"Digest of the current block, also part of the block header.",
				),
				primitives.NewMetadataModuleStorageEntry(
					"Events",
					primitives.MetadataModuleStorageEntryModifierDefault,
					primitives.NewMetadataModuleStorageEntryDefinitionPlain(sc.ToCompact(metadata.TypesSystemEventStorage)),
					"Events deposited for the current block.   NOTE: The item is unbound and should therefore never be read on chain.",
				),
				primitives.NewMetadataModuleStorageEntry(
					"EventTopics",
					primitives.MetadataModuleStorageEntryModifierDefault,
					primitives.NewMetadataModuleStorageEntryDefinitionMap(
						sc.Sequence[primitives.MetadataModuleStorageHashFunc]{primitives.MetadataModuleStorageHashFuncMultiBlake128Concat},
						sc.ToCompact(metadata.TypesH256),
						sc.ToCompact(metadata.TypesVecBlockNumEventIndex),
					),
					"Mapping between a topic (represented by T::Hash) and a vector of indexes  of events in the `<Events<T>>` list.",
				),
				primitives.NewMetadataModuleStorageEntry(
					"EventCount",
					primitives.MetadataModuleStorageEntryModifierDefault,
					primitives.NewMetadataModuleStorageEntryDefinitionPlain(
						sc.ToCompact(metadata.PrimitiveTypesU32),
					),
					"The number of events in the `Events<T>` list.",
				),
				primitives.NewMetadataModuleStorageEntry(
					"LastRuntimeUpgrade",
					primitives.MetadataModuleStorageEntryModifierOptional,
					primitives.NewMetadataModuleStorageEntryDefinitionPlain(sc.ToCompact(lastRuntimeUpgradeInfoId)),
					"Stores the `spec_version` and `spec_name` of when the last runtime upgrade happened.",
				),
				primitives.NewMetadataModuleStorageEntry(
					"ExecutionPhase",
					primitives.MetadataModuleStorageEntryModifierOptional,
					primitives.NewMetadataModuleStorageEntryDefinitionPlain(sc.ToCompact(typesPhaseId)),
					"The execution phase of the block.",
				),
				primitives.NewMetadataModuleStorageEntry(
					"AuthorizedUpgrade",
					primitives.MetadataModuleStorageEntryModifierOptional,
					primitives.NewMetadataModuleStorageEntryDefinitionPlain(sc.ToCompact(metadata.TypesCodeUpgradeAuthorization)),
					"Optional code upgrade authorization for the runtime.",
				),
			},
		})
}

func mutateAccount(account *primitives.AccountInfo, data *primitives.AccountData) {
	if data != nil {
		account.Data = *data
	} else {
		account.Data = primitives.AccountData{}
	}
}

// CanSetCode determines whether it is possible to update the code.
//
// Checks the given code if it is a valid runtime wasm blob by instantianting
// it and extracting the runtime version of it. It checks that the runtime version
// of the old and new runtime has the same spec name and that the spec version is increasing.
func (m module) CanSetCode(codeBlob sc.Sequence[sc.U8]) error {
	currentVersion := *m.Config.Version

	runtimeVersionBytes := m.ioMisc.RuntimeVersion(sc.SequenceU8ToBytes(codeBlob))
	buffer := bytes.NewBuffer(runtimeVersionBytes)
	sc.DecodeBool(buffer)            // option
	sc.DecodeCompact[sc.U32](buffer) // length

	newVersion, err := primitives.DecodeRuntimeVersion(buffer)
	if err != nil {
		return NewDispatchErrorFailedToExtractRuntimeVersion(m.Index)
	}

	if newVersion.SpecName != currentVersion.SpecName {
		return NewDispatchErrorInvalidSpecName(m.Index)
	}

	if newVersion.SpecVersion <= currentVersion.SpecVersion {
		return NewDispatchErrorSpecVersionNeedsToIncrease(m.Index)
	}

	return nil
}

// To be called after any origin/privilege checks. Put the code upgrade authorization into
// storage and emit an event.
func (m module) DoAuthorizeUpgrade(codeHash primitives.H256, checkVersion sc.Bool) {
	value := CodeUpgradeAuthorization{codeHash, checkVersion}
	m.storage.AuthorizedUpgrade.Put(value)
	m.DepositEvent(newEventUpgradeAuthorized(m.Index, codeHash, checkVersion))
}

// DoApplyAuthorizeUpgrade applies an authorized upgrade, performing any validation checks,
// and removing the authorization. Whether or not the code is set directly depends on the
// `OnSetCode` configuration of the runtime.
func (m module) DoApplyAuthorizeUpgrade(codeBlob sc.Sequence[sc.U8]) (primitives.PostDispatchInfo, error) {
	_, err := m.validateAuthorizedUpgrade(codeBlob)
	if err != nil {
		return primitives.PostDispatchInfo{}, err
	}

	err = m.OnSetCode.SetCode(codeBlob)
	if err != nil {
		return primitives.PostDispatchInfo{}, err
	}

	m.storage.AuthorizedUpgrade.Clear()

	post := primitives.PostDispatchInfo{
		// consume the rest of the block to prevent further transactions
		ActualWeight: sc.NewOption[primitives.Weight](m.constants.BlockWeights.MaxBlock),
		// no fee for valid upgrade
		PaysFee: primitives.PaysNo,
	}

	return post, nil
}

// Check that provided `code` can be upgraded to. Namely, check that its hash matches an
// existing authorization and that it meets the specification requirements of `can_set_code`.
func (m module) validateAuthorizedUpgrade(codeBlob sc.Sequence[sc.U8]) (primitives.H256, error) {
	authorization, err := m.storage.AuthorizedUpgrade.Get()
	if err != nil {
		return primitives.H256{}, NewDispatchErrorNothingAuthorized(m.Index)
	}

	hash := m.ioHashing.Blake256(sc.SequenceU8ToBytes(codeBlob))
	actualHash, err := primitives.NewH256(sc.BytesToFixedSequenceU8(hash)...)
	if err != nil {
		return primitives.H256{}, err
	}

	if !reflect.DeepEqual(actualHash, authorization.CodeHash) {
		return primitives.H256{}, NewDispatchErrorUnauthorized(m.Index)
	}

	if authorization.CheckVersion {
		err := m.CanSetCode(codeBlob)
		if err != nil {
			return primitives.H256{}, err
		}
	}

	return actualHash, nil
}

// EnsureRoot ensures that the origin represents the root.
func EnsureRoot(origin primitives.RuntimeOrigin) error {
	if origin.IsRootOrigin() {
		return nil
	} else {
		return primitives.NewDispatchErrorBadOrigin()
	}
}

// EnsureSigned ensures that the origin represents a signed extrinsic (i.e. transaction).
// Returns `Ok` with the account that signed the extrinsic or an `Err` otherwise.
func EnsureSigned(origin primitives.RawOrigin) (sc.Option[primitives.AccountId], error) {
	if origin.IsSignedOrigin() {
		return sc.NewOption[primitives.AccountId](origin.VaryingData[1]), nil
	}

	return sc.Option[primitives.AccountId]{}, primitives.NewDispatchErrorBadOrigin()
}

// EnsureSignedOrRoot ensures the origin represents either a signed extrinsic or the root.
// Returns an empty Option if the origin is `Root`.
// Returns an Option with the signer if the origin is signed.
// Returns a `BadOrigin` error if neither of the above.
func EnsureSignedOrRoot(origin primitives.RawOrigin) (sc.Option[primitives.AccountId], error) {
	if origin.IsRootOrigin() {
		return sc.NewOption[primitives.AccountId](nil), nil
	} else if origin.IsSignedOrigin() {
		return sc.NewOption[primitives.AccountId](origin.VaryingData[1]), nil
	}

	return sc.Option[primitives.AccountId]{}, primitives.NewDispatchErrorBadOrigin()
}

// Ensure that the origin `o` represents an unsigned extrinsic. Returns `Ok` or an `Err` otherwise.
func EnsureNone(origin primitives.RawOrigin) (sc.Option[primitives.AccountId], error) {
	switch origin.IsNoneOrigin() {
	case true:
		return sc.Option[primitives.AccountId]{}, nil
	default:
		return sc.Option[primitives.AccountId]{}, primitives.NewDispatchErrorBadOrigin()
	}
}
