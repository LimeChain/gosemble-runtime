package babe

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"reflect"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/frame/system"
	"github.com/LimeChain/gosemble/hooks"
	babetypes "github.com/LimeChain/gosemble/primitives/babe"
	"github.com/LimeChain/gosemble/primitives/io"
	"github.com/LimeChain/gosemble/primitives/log"
	sessiontypes "github.com/LimeChain/gosemble/primitives/session"

	primitives "github.com/LimeChain/gosemble/primitives/types"
)

const (
	SkippedEpochsBound                    = 100
	UnderConstructionSegmentLength sc.U32 = 256
)

// VRF context used for per-slot randomness generation.
var RandomnessVrfContext = []byte("BabeVRFInOutContext")

var (
	EngineId  = [4]byte{'B', 'A', 'B', 'E'}
	KeyTypeId = [4]byte{'b', 'a', 'b', 'e'}
)

var (
	errAuthoritiesListEmpty            = errors.New("Babe: Authorities cannot be empty!")
	errAuthoritiesAlreadyInitialized   = errors.New("Babe: Authorities are already initialized!")
	errAuthoritiesExceedMaxAuthorities = errors.New("Babe: Initial number of authorities should be lower than MaxAuthorities")
	errEpochConfigIsUninitialized      = errors.New("Babe: EpochConfig is initialized in genesis; we never `take` or `kill` it; qed")
	errEpochIndexOverflow              = errors.New("Babe: epoch index is u64; it is always only incremented by one; if u64 is not enough we should crash for safety; qed.")
	errTimestampMismatch               = errors.New("Babe: Timestamp slot must match `CurrentSlot`")
	errZeroSlotDuration                = errors.New("Babe: Slot duration cannot be zero.")
)

const (
	functionReportEquivocationIndex         = iota // TODO: implemented
	functionReportEquivocationUnsignedIndex        // TODO: implemented
	functionPlanConfigChangeIndex
)

type Module interface {
	primitives.Module
	hooks.OnTimestampSet[sc.U64]
	sessiontypes.OneSessionHandler

	FindAuthor(digests sc.Sequence[primitives.DigestPreRuntime]) (sc.Option[sc.U32], error)

	StorageAuthorities() (sc.Sequence[primitives.Authority], error)
	StorageRandomness() (babetypes.Randomness, error)
	StorageSegmentIndexSet(sc.U32)
	StorageEpochConfig() (babetypes.EpochConfiguration, error)
	StorageEpochConfigSet(value babetypes.EpochConfiguration)

	EnactEpochChange(authorities sc.Sequence[primitives.Authority], nextAuthorities sc.Sequence[primitives.Authority], sessionIndex sc.Option[sc.U32]) error
	ShouldEpochChange(now sc.U64) bool

	SlotDuration() sc.U64
	EpochDuration() sc.U64

	EpochConfig() babetypes.EpochConfiguration
	EpochStartSlot(epochIndex sc.U64, genesisSlot babetypes.Slot, epochDuration sc.U64) (babetypes.Slot, error)

	CurrentEpochStart() (babetypes.Slot, error)
	CurrentEpoch() (babetypes.Epoch, error)
	NextEpoch() (babetypes.Epoch, error)
}

type module struct {
	primitives.DefaultInherentProvider
	hooks.DefaultDispatchModule

	index              sc.U8
	config             *Config
	constants          *consts
	functions          map[sc.U8]primitives.Call
	storage            *storage
	logger             log.Logger
	systemModule       system.Module
	disabledValidators primitives.DisabledValidators
	epochChangeTrigger EpochChangeTrigger
	ioHashing          io.Hashing
	mdGenerator        *primitives.MetadataTypeGenerator
}

func New(index sc.U8, config *Config, mdGenerator *primitives.MetadataTypeGenerator, logger log.Logger) Module {
	storage := newStorage()

	functions := map[sc.U8]primitives.Call{
		functionPlanConfigChangeIndex: newCallPlanConfigChange(index, functionPlanConfigChangeIndex, config.DbWeight, storage.PendingEpochConfigChange),
	}

	return module{
		index:              index,
		config:             config,
		constants:          newConstants(config.EpochDuration, config.MinimumPeriod, config.MaxAuthorities),
		storage:            storage,
		functions:          functions,
		logger:             logger,
		systemModule:       config.SystemModule,
		disabledValidators: config.SessionModule,
		epochChangeTrigger: config.EpochChangeTrigger,
		ioHashing:          io.NewHashing(),
		mdGenerator:        mdGenerator,
	}
}

func (m module) name() sc.Str {
	return "Babe"
}

func (m module) GetIndex() sc.U8 {
	return m.index
}

func (m module) Functions() map[sc.U8]primitives.Call {
	return m.functions
}

func (m module) PreDispatch(_ primitives.Call) (sc.Empty, error) {
	// TODO: Implement once the rest of the calls are implemented
	return sc.Empty{}, nil
}

func (m module) ValidateUnsigned(_ primitives.TransactionSource, _ primitives.Call) (primitives.ValidTransaction, error) {
	// TODO: Implement once the rest of the calls are implemented
	return primitives.ValidTransaction{}, primitives.NewTransactionValidityError(primitives.NewUnknownTransactionNoUnsignedValidator())
}

// Storage

// Current epoch authorities.
func (m module) StorageAuthorities() (sc.Sequence[primitives.Authority], error) {
	return m.storage.Authorities.Get()
}

func (m module) StorageRandomness() (babetypes.Randomness, error) {
	return m.storage.Randomness.Get()
}

// Randomness under construction.
//
// We make a trade-off between storage accesses and list length.
// We store the under-construction randomness in segments of up to
// `UNDER_CONSTRUCTION_SEGMENT_LENGTH`.
//
// Once a segment reaches this length, we begin the next one.
// We reset all segments and return to `0` at the beginning of every
// epoch.
func (m module) StorageSegmentIndexSet(value sc.U32) {
	m.storage.SegmentIndex.Put(value)
}

// The configuration for the current epoch. Should never be `None` as it is initialized in genesis.
func (m module) StorageEpochConfig() (babetypes.EpochConfiguration, error) {
	return m.storage.EpochConfig.Get()
}

func (m module) StorageEpochConfigSet(value babetypes.EpochConfiguration) {
	m.storage.EpochConfig.Put(value)
}

// FindAuthor interface

func (m module) FindAuthor(digests sc.Sequence[primitives.DigestPreRuntime]) (sc.Option[sc.U32], error) {
	for _, preRuntime := range digests {
		if reflect.DeepEqual(sc.FixedSequenceU8ToBytes(preRuntime.ConsensusEngineId), EngineId[:]) {
			buffer := bytes.NewBuffer(sc.SequenceU8ToBytes(preRuntime.Message))

			preDigest, err := babetypes.DecodePreDigest(buffer)
			if err != nil {
				return sc.NewOption[sc.U32](nil), err
			}

			authorIndex, err := preDigest.AuthorityIndex()
			if err != nil {
				return sc.NewOption[sc.U32](nil), err
			}

			return sc.NewOption[sc.U32](sc.U32(authorIndex)), nil
		}
	}

	return sc.NewOption[sc.U32](nil), nil
}

// ShouldEndSession interface

func (m module) ShouldEndSession(now sc.U64) bool {
	// it might be (and it is in current implementation) that session module is calling
	// `should_end_session` from it's own `on_initialize` handler, in which case it's
	// possible that babe's own `on_initialize` has not run yet, so let's ensure that we
	// have initialized the pallet and updated the current slot.
	m.initialize(now)
	return m.ShouldEpochChange(now)
}

// Config

// The amount of time, in slots, that each epoch should last.
func (m module) EpochDuration() sc.U64 {
	return m.config.EpochDuration
}

// Determine the BABE slot duration based on the Timestamp module configuration.
func (m module) SlotDuration() sc.U64 {
	// we double the minimum block-period so each author can always propose within
	// the majority of their slot.
	return m.config.MinimumPeriod * 2
}

// Determine whether an epoch change should take place at this block.
// Assumes that initialization has already taken place.
func (m module) ShouldEpochChange(now sc.U64) bool {
	// The epoch has technically ended during the passage of time
	// between this block and the last, but we have to "end" the epoch now,
	// since there is no earlier possible block we could have done it.
	//
	// The exception is for block 1: the genesis has slot 0, so we treat
	// epoch 0 as having started at the slot of block 1. We want to use
	// the same randomness and validator set as signalled in the genesis,
	// so we don't rotate the epoch.

	currentSlot, err := m.storage.CurrentSlot.Get()
	if err != nil {
		return false
	}

	currentEpochStart, err := m.CurrentEpochStart()
	if err != nil {
		return false
	}

	diff := sc.SaturatingSubU64(currentSlot, currentEpochStart)

	return now != 1 && (diff >= m.EpochDuration())
}

// DANGEROUS: Enact an epoch change. Should be done on every block where `should_epoch_change`
// has returned `true`, and the caller is the only caller of this function.
//
// Typically, this is not handled directly by the user, but by higher-level validator-set
// manager logic like `pallet-session`.
//
// This doesn't do anything if `authorities` is empty.
func (m module) EnactEpochChange(authorities sc.Sequence[primitives.Authority], nextAuthorities sc.Sequence[primitives.Authority], sessionIndex sc.Option[sc.U32]) error {
	// PRECONDITION: caller has done initialization and is guaranteed
	// by the session module to be called before this.
	// initializedBytes, err := m.storage.Initialized.GetBytes()
	// if err != nil {
	// 	return err
	// }
	// if initializedBytes.HasValue {
	// 	m.logger.Debug("")
	// }

	if len(authorities) == 0 {
		m.logger.Warn("Ignoring empty epoch change.")
		return nil
	}

	// Update epoch index.
	//
	// NOTE: we figure out the epoch index from the slot, which may not
	// necessarily be contiguous if the chain was offline for more than
	// `T::EpochDuration` slots. When skipping from epoch N to e.g. N+4, we
	// will be using the randomness and authorities for that epoch that had
	// been previously announced for epoch N+1, and the randomness collected
	// during the current epoch (N) will be used for epoch N+5.
	slot, err := m.storage.CurrentSlot.Get()
	if err != nil {
		return err
	}

	genesisSlot, err := m.storage.GenesisSlot.Get()
	if err != nil {
		return err
	}

	epochIndex := m.EpochIndex(slot, genesisSlot, m.config.EpochDuration)

	currentEpochIndex, err := m.storage.EpochIndex.Get()
	if err != nil {
		return err
	}

	if sc.SaturatingAddU64(currentEpochIndex, 1) != epochIndex {
		// we are skipping epochs therefore we need to update the mapping
		// of epochs to session
		if sessionIndex.HasValue {
			m.storage.SkippedEpochs.Mutate(
				func(skippedEpochs *sc.FixedSequence[babetypes.SkippedEpoch]) (sc.FixedSequence[babetypes.SkippedEpoch], error) {
					if epochIndex < sc.U64(sessionIndex.Value) {
						m.logger.Warn(fmt.Sprintf("Current epoch index %d is lower than session index %d.", epochIndex, sessionIndex.Value))
						return sc.FixedSequence[babetypes.SkippedEpoch]{}, nil
					}

					if len(*skippedEpochs) >= int(SkippedEpochsBound) {
						// NOTE: this is O(n) but we currently don't have a bounded `VecDeque`.
						// this vector is bounded to a small number of elements so performance
						// shouldn't be an issue.
						*skippedEpochs = (*skippedEpochs)[1:]
					}

					skipped := babetypes.SkippedEpoch{U64: epochIndex, SessionIndex: sc.U32(sessionIndex.Value)}
					(*skippedEpochs)[len(*skippedEpochs)] = skipped

					return *skippedEpochs, nil
				},
			)
		}
	}

	m.storage.EpochIndex.Put(epochIndex)
	m.storage.Authorities.Put(authorities)

	// Update epoch randomness.
	nextEpochIndex, err := sc.CheckedAddU64(epochIndex, 1)
	if err != nil {
		return errors.New("epoch indices will never reach 2^64 before the death of the universe; qed")
	}

	// Returns randomness for the current epoch and computes the *next*
	// epoch randomness.
	randomness, err := m.randomnessChangeEpoch(nextEpochIndex)
	if err != nil {
		return err
	}
	m.storage.Randomness.Put(randomness)

	// Update the next epoch authorities.
	m.storage.NextAuthorities.Put(nextAuthorities)

	// Update the start blocks of the previous and new current epoch.
	m.storage.EpochStart.Mutate(
		func(epochStartBlocks *babetypes.EpochStartBlocks) (babetypes.EpochStartBlocks, error) {
			blockNumber, err := m.config.SystemModule.StorageBlockNumber()
			if err != nil {
				return babetypes.EpochStartBlocks{}, err
			}

			(*epochStartBlocks).Previous = (*epochStartBlocks).Current
			(*epochStartBlocks).Current = 0
			(*epochStartBlocks).Current = blockNumber

			return *epochStartBlocks, nil
		},
	)

	// After we update the current epoch, we signal the *next* epoch change
	// so that nodes can track changes.
	nextRandomness, err := m.storage.NextRandomness.Get()
	if err != nil {
		return err
	}

	nextEpoch := NextEpochDescriptor{
		Authorities: nextAuthorities,
		Randomness:  nextRandomness,
	}
	m.depositConsensus(NewConsensusLogNextEpochData(nextEpoch))

	nextConfig, err := m.storage.NextEpochConfig.Get()
	if err != nil {
		return err
	}

	if !reflect.DeepEqual(nextConfig, babetypes.EpochConfiguration{}) {
		m.storage.EpochConfig.Put(nextConfig)
	}

	pendingEpochConfigChange, err := m.storage.PendingEpochConfigChange.Take()
	if err != nil {
		return err
	}

	if reflect.DeepEqual(pendingEpochConfigChange, NextConfigDescriptor{}) {
		nextEpochConfig := pendingEpochConfigChange.V1
		m.storage.NextEpochConfig.Put(nextEpochConfig)

		m.depositConsensus(NewConsensusLogNextConfigData(pendingEpochConfigChange))
	}

	return nil
}

// Finds the start slot of the current epoch.
//
// Only guaranteed to give correct results after `initialize` of the first
// block in the chain (as its result is based off of `GenesisSlot`).
func (m module) CurrentEpochStart() (babetypes.Slot, error) {
	epochIndex, err := m.storage.EpochIndex.Get()
	if err != nil {
		return 0, err
	}

	genesisSlot, err := m.storage.GenesisSlot.Get()
	if err != nil {
		return 0, err
	}

	return m.EpochStartSlot(epochIndex, genesisSlot, m.EpochDuration())
}

// Produces information about the current epoch.
func (m module) CurrentEpoch() (babetypes.Epoch, error) {
	epochIndex, err := m.storage.EpochIndex.Get()
	if err != nil {
		return babetypes.Epoch{}, err
	}

	epochStart, err := m.CurrentEpochStart()
	if err != nil {
		return babetypes.Epoch{}, err
	}

	authorities, err := m.storage.Authorities.Get()
	if err != nil {
		return babetypes.Epoch{}, err
	}

	randomness, err := m.storage.Randomness.Get()
	if err != nil {
		return babetypes.Epoch{}, err
	}

	epochConfig, err := m.storage.EpochConfig.Get()
	if err != nil {
		return babetypes.Epoch{}, err
	}
	if reflect.DeepEqual(epochConfig, babetypes.EpochConfiguration{}) {
		return babetypes.Epoch{}, errEpochConfigIsUninitialized
	}

	return babetypes.Epoch{
		EpochIndex:  epochIndex,
		StartSlot:   epochStart,
		Duration:    m.EpochDuration(),
		Authorities: authorities,
		Randomness:  randomness,
		Config:      epochConfig,
	}, nil
}

// Produces information about the next epoch (which was already previously announced).
func (m module) NextEpoch() (babetypes.Epoch, error) {
	epochIndex, err := m.storage.EpochIndex.Get()
	if err != nil {
		return babetypes.Epoch{}, err
	}

	nextEpochIndex, err := sc.CheckedAddU64(epochIndex, 1)
	if err != nil {
		return babetypes.Epoch{}, errEpochIndexOverflow
	}

	genesisSlot, err := m.storage.GenesisSlot.Get()
	if err != nil {
		return babetypes.Epoch{}, err
	}

	epochDuration := m.EpochDuration()

	startSlot, err := m.EpochStartSlot(
		nextEpochIndex,
		genesisSlot,
		epochDuration,
	)
	if err != nil {
		return babetypes.Epoch{}, err
	}

	nextAuthorities, err := m.storage.NextAuthorities.Get()
	if err != nil {
		return babetypes.Epoch{}, err
	}

	nextRandomness, err := m.storage.NextRandomness.Get()
	if err != nil {
		return babetypes.Epoch{}, err
	}

	// The configuration for the next epoch, `None` if the config will not change
	// (you can fallback to `EpochConfig` instead in that case).
	nextEpochConfig, err := m.storage.NextEpochConfig.Get()
	if err != nil {
		return babetypes.Epoch{}, err
	}

	if reflect.DeepEqual(nextEpochConfig, babetypes.EpochConfiguration{}) {
		epochConfig, err := m.storage.EpochConfig.Get()
		if err != nil {
			return babetypes.Epoch{}, err
		}
		// expect("EpochConfig is initialized in genesis; we never `take` or `kill` it; qed")
		if reflect.DeepEqual(epochConfig, babetypes.EpochConfiguration{}) {
			nextEpochConfig = epochConfig
		}
	}

	return babetypes.Epoch{
		EpochIndex:  nextEpochIndex,
		StartSlot:   startSlot,
		Duration:    epochDuration,
		Authorities: nextAuthorities,
		Randomness:  nextRandomness,
		Config:      nextEpochConfig,
	}, nil
}

func (m module) depositConsensus(new ConsensusLog) {
	log := primitives.NewDigestItemConsensusMessage(
		sc.BytesToFixedSequenceU8(EngineId[:]),
		sc.BytesToSequenceU8(new.Bytes()),
	)
	m.systemModule.DepositLog(log)
}

func (m module) depositRandomness(randomness babetypes.Randomness) error {
	segmentIdx, err := m.storage.SegmentIndex.Get()
	if err != nil {
		return err
	}

	segment, err := m.storage.UnderConstruction.Get(segmentIdx)
	if err != nil {
		return err
	}

	// segment.try_push(randomness)
	if len(segment)+len(randomness) <= int(UnderConstructionSegmentLength) {
		// push onto current segment: not full.
		m.storage.UnderConstruction.Put(segmentIdx, segment)
	} else {
		// move onto the next segment and update the index.
		segmentIdx += 1
		// boundedRandomness := BoundedVec::<_, ConstU32<UNDER_CONSTRUCTION_SEGMENT_LENGTH>>::try_from(vec![
		// 			*randomness,
		// 		])
		// .expect("UNDER_CONSTRUCTION_SEGMENT_LENGTH >= 1");
		// m.storage.UnderConstruction.Put(segmentIdx, boundedRandomness)
		m.storage.SegmentIndex.Put(segmentIdx)
	}

	return nil
}

func (m module) initializeGenesisAuthorities(authorities sc.Sequence[primitives.Authority]) error {
	if len(authorities) != 0 {
		totalAuthorities, err := m.storage.Authorities.DecodeLen()
		if err != nil {
			return err
		}

		if totalAuthorities.HasValue {
			return errAuthoritiesAlreadyInitialized
		}

		if len(authorities) > int(m.config.MaxAuthorities) {
			return errAuthoritiesExceedMaxAuthorities
		}

		m.storage.Authorities.Put(authorities)
		m.storage.NextAuthorities.Put(authorities)
	}

	return nil
}

func (m module) initializeGenesisEpoch(slot babetypes.Slot) error {
	m.storage.GenesisSlot.Put(slot)

	genesisSlot, err := m.storage.GenesisSlot.Get()
	if err != nil {
		return err
	}
	if genesisSlot != 0 {
		m.logger.Debug("genesis slot is set")
	}

	// deposit a log because this is the first block in epoch #0
	// we use the same values as genesis because we haven't collected any
	// randomness yet.
	authorities, err := m.storage.Authorities.Get()
	if err != nil {
		return err
	}

	randomness, err := m.storage.Randomness.Get()
	if err != nil {
		return err
	}

	next := NextEpochDescriptor{
		Authorities: authorities,
		Randomness:  randomness,
	}

	m.depositConsensus(NewConsensusLogNextEpochData(next))

	return nil
}

func (m module) initialize(now sc.U64) error {
	// since `initialize` can be called twice (e.g. if session module is present)
	// let's ensure that we only do the initialization once per block
	initialized, err := m.storage.Initialized.Get()
	if err != nil {
		return err
	}
	if initialized.HasValue {
		return nil
	}

	digest, err := m.config.SystemDigest()
	if err != nil {
		return err
	}

	preRuntimeDigests, err := digest.PreRuntimes()
	if err != nil {
		return err
	}

	preDigest := sc.NewOption[babetypes.PreDigest](nil)

	for _, digest := range preRuntimeDigests {
		if reflect.DeepEqual(sc.FixedSequenceU8ToBytes(digest.ConsensusEngineId), EngineId[:]) {
			buffer := bytes.NewBuffer(sc.SequenceU8ToBytes(digest.Message))

			pre, err := babetypes.DecodePreDigest(buffer)
			if err != nil {
				return err
			}

			preDigest = sc.NewOption[babetypes.PreDigest](pre)
			break
		}
	}

	if preDigest.HasValue {
		// the slot number of the current block being initialized
		slot, err := preDigest.Value.Slot()
		if err != nil {
			return err
		}

		// on the first non-zero block (i.e. block #1)
		// this is where the first epoch (epoch #0) actually starts.
		// we need to adjust internal storage accordingly.
		genesisSlot, err := m.storage.GenesisSlot.Get()
		if err != nil {
			return err
		}

		if genesisSlot == 0 {
			m.initializeGenesisEpoch(slot)
		}

		// how many slots were skipped between current and last block
		currentSlot, err := m.storage.CurrentSlot.Get()
		if err != nil {
			return err
		}

		lateness := sc.SaturatingSubU64(slot, currentSlot+1)

		m.storage.Lateness.Put(lateness)
		m.storage.CurrentSlot.Put(slot)
	}

	m.storage.Initialized.Put(preDigest)

	// enact epoch change, if necessary.
	m.epochChangeTrigger.Trigger(now)

	return nil
}

// Call this function exactly once when an epoch changes, to update the
// randomness. Returns the new randomness.
func (m module) randomnessChangeEpoch(nextEpochIndex sc.U64) (babetypes.Randomness, error) {
	thisRandomness, err := m.storage.NextRandomness.Get()
	if err != nil {
		return babetypes.Randomness{}, err
	}

	segmentIdx, err := m.storage.SegmentIndex.Mutate(func(s *sc.U32) (sc.U32, error) {
		*s = 0
		return *s, nil
	})

	// overestimate to the segment being full.
	rhoSize := sc.SaturatingAddU32(segmentIdx, 1) * UnderConstructionSegmentLength

	randomnessUnderConstruction := make(sc.Sequence[babetypes.Randomness], 0)

	for i := sc.U32(0); i < segmentIdx; i++ {
		b, err := m.storage.UnderConstruction.TakeBytes(i)
		if err != nil {
			return babetypes.Randomness{}, err
		}
		randomnessUnderConstruction = append(randomnessUnderConstruction, sc.BytesToFixedSequenceU8(b))
	}

	nextRandomness := m.computeRandomness(
		thisRandomness,
		nextEpochIndex,
		randomnessUnderConstruction,
		sc.NewOption[sc.U32](rhoSize),
	)

	m.storage.NextRandomness.Put(nextRandomness)

	return thisRandomness, nil
}

// TimestampSet interface implementation

func (m module) OnTimestampSet(now sc.U64) error {
	slotDuration := m.SlotDuration()

	if slotDuration == 0 {
		return errZeroSlotDuration
	}

	timestampSlot := now / slotDuration
	if timestampSlot > math.MaxUint64 {
		timestampSlot = math.MaxUint64
	}

	// Current slot number.
	currentSlot, err := m.storage.CurrentSlot.Get()
	if err != nil {
		return err
	}

	if currentSlot != timestampSlot {
		m.logger.Warn(fmt.Sprintf("Timestamp slot %d does not match CurrentSlot %d", timestampSlot, currentSlot))
		return errTimestampMismatch
	}

	return nil
}

// OneSessionHandler interface implementation

func (m module) KeyType() primitives.PublicKeyType {
	return m.config.KeyType
}

func (m module) KeyTypeId() [4]byte {
	return KeyTypeId
}

func (m module) DecodeKey(buffer *bytes.Buffer) (primitives.Sr25519PublicKey, error) {
	key, err := primitives.DecodeSr25519PublicKey(buffer)
	if err != nil {
		return primitives.Sr25519PublicKey{}, err
	}

	return key, nil
}

func (m module) OnGenesisSession(validators sc.Sequence[primitives.Validator]) error {
	authorities := primitives.AuthoritiesFrom(validators)
	return m.initializeGenesisAuthorities(authorities)
}

// OnNewSession triggers once the session has changed.
func (m module) OnNewSession(changed bool, validators sc.Sequence[primitives.Validator], queuedValidators sc.Sequence[primitives.Validator]) error {
	authorities := primitives.AuthoritiesFrom(validators)
	if len(authorities) > int(m.config.MaxAuthorities) {
		m.logger.Warn("The session has more validators than expected. A runtime configuration adjustment may be needed.")
		authorities = authorities[:m.config.MaxAuthorities]
	}

	nextAuthorities := primitives.AuthoritiesFrom(queuedValidators)
	if len(nextAuthorities) > int(m.config.MaxAuthorities) {
		m.logger.Warn("The session has more queued validators than expected. A runtime configuration adjustment may be needed.")
		authorities = authorities[:m.config.MaxAuthorities]
	}

	sessionIndex, err := m.config.SessionModule.CurrentIndex()
	if err != nil {
		return err
	}

	return m.EnactEpochChange(authorities, nextAuthorities, sc.NewOption[sc.U32](sessionIndex))
}

// OnBeforeSessionEnding notifies for the end of the session. Triggered before the end of the session.
func (m module) OnBeforeSessionEnding() {}

// OnDisabled triggers hook for a disabled validator. Acts accordingly until a new session begins.
func (m module) OnDisabled(validatorIndex sc.U32) {
	m.depositConsensus(NewConsensusLogOnDisabled(validatorIndex))
}

// Module hooks implementation

// Initialization
func (m module) OnInitialize(now sc.U64) (primitives.Weight, error) {
	err := m.initialize(now)
	if err != nil {
		return primitives.Weight{}, err
	}
	return primitives.WeightZero(), nil
}

// Block finalization
func (m module) OnFinalize(_now sc.U64) error {
	// at the end of the block, we can safely include the new VRF output
	// from this block into the under-construction randomness. If we've determined
	// that this block was the first in a new epoch, the changeover logic has
	// already occurred at this point, so the under-construction randomness
	// will only contain outputs from the right epoch.
	preDigest, err := m.storage.Initialized.Take()
	if err != nil {
		return err
	}

	if preDigest.HasValue {
		authorityIndex, err := preDigest.Value.AuthorityIndex()
		if err != nil {
			return err
		}

		disabled, err := m.disabledValidators.IsDisabled(authorityIndex)
		if err != nil {
			return err
		}

		if disabled {
			m.logger.Critical(
				fmt.Sprintf("Validator with index %d is disabled and should not be attempting to author blocks.", authorityIndex),
			)
		}

		signature, err := preDigest.Value.VrfSignature()
		if err != nil {
			return err
		}

		var randomness babetypes.Randomness
		if signature.HasValue {
			authorities, err := m.storage.Authorities.Get()
			if err != nil {
				return err
			}

			if len(authorities) == 0 || authorityIndex > sc.U32(len(authorities)) {
				m.logger.Critical("Authority index is out of bounds.")
			}

			authority := authorities[authorityIndex]

			currentRandomness, err := m.storage.Randomness.Get()
			if err != nil {
				return err
			}

			currentSlot, err := m.storage.CurrentSlot.Get()
			if err != nil {
				return err
			}

			epochIndex, err := m.storage.EpochIndex.Get()
			if err != nil {
				return err
			}

			transcript := makeVrfTranscript(currentRandomness, currentSlot, epochIndex)

			// NOTE: this is verified by the client when importing the block, before
			// execution. We don't run the verification again here to avoid slowing
			// down the runtime.
			// debug_assert!({
			// 	use sp_core::crypto::VrfPublic;
			// 	public.vrf_verify(&transcript.clone().into_sign_data(), &signature)
			// });

			output := [32]byte{}
			for i := 0; i < 32; i++ {
				output[i] = byte(signature.Value.PreOutput[i])
			}

			public, err := primitives.NewPublicKey(sc.FixedSequenceU8ToBytes(authority.Id.FixedSequence))

			inout, err := primitives.AttachInput(output, public, transcript)
			if err != nil {
				return err
			}

			rand, err := inout.MakeBytes(16, RandomnessVrfContext)
			if err != nil {
				return err
			}

			randomness = sc.BytesToFixedSequenceU8(rand)
		}

		if preDigest.Value.IsPrimary() {
			m.depositRandomness(randomness)
		}

		m.storage.AuthorVrfRandomness.Put(sc.NewOption[babetypes.Randomness](randomness))
	}

	// remove temporary "environment" entry from storage
	m.storage.Lateness.Clear()

	return nil
}

// Returns the epoch index the given slot belongs to.
func (m module) EpochIndex(slot babetypes.Slot, genesisSlot babetypes.Slot, epochDuration sc.U64) sc.U64 {
	return sc.SaturatingSubU64(slot, genesisSlot) / epochDuration
}

// Returns the first slot at the given epoch index.
func (m module) EpochStartSlot(epochIndex sc.U64, genesisSlot babetypes.Slot, epochDuration sc.U64) (babetypes.Slot, error) {
	a, err := sc.CheckedMulU64(epochIndex, epochDuration)
	if err != nil {
		return 0, err
	}

	b, err := sc.CheckedAddU64(a, genesisSlot)
	if err != nil {
		return 0, errors.New("slot number is u64; it should relate in some way to wall clock time; if u64 is not enough we should crash for safety; qed.")
	}

	return b, nil // epochIndex * epochDuration + genesisSlot
}

func (m module) EpochConfig() babetypes.EpochConfiguration {
	return m.config.EpochConfig
}

// compute randomness for a new epoch. rho is the concatenation of all
// VRF outputs in the prior epoch.
//
// an optional size hint as to how many VRF outputs there were may be provided.
func (m module) computeRandomness(lastEpochRandomness babetypes.Randomness, epochIndex sc.U64, rho sc.Sequence[babetypes.Randomness], rhoSizeHint sc.Option[sc.U32]) babetypes.Randomness {
	// TODO: rhoSizeHint is usize
	var sizeHint int64
	if rhoSizeHint.HasValue {
		sizeHint = rhoSizeHint.Value.ToBigInt().Int64()
	} else {
		sizeHint = 0
	}

	s := make([]byte, 0, 40+sizeHint*babetypes.RandomnessLength)
	s = append(s, sc.FixedSequenceU8ToBytes(lastEpochRandomness)[:]...)

	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, uint64(epochIndex))
	s = append(s, buf...)

	for _, vrfOutput := range rho {
		s = append(s, sc.FixedSequenceU8ToBytes(vrfOutput)[:]...)
	}

	return sc.BytesToFixedSequenceU8(m.ioHashing.Blake256(s))
}
