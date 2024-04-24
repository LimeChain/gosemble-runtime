package grandpa

import (
	"bytes"
	"errors"
	"math"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/frame/session"
	"github.com/LimeChain/gosemble/frame/system"
	"github.com/LimeChain/gosemble/hooks"
	"github.com/LimeChain/gosemble/primitives/log"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

const (
	name = sc.Str("Grandpa")
)

var (
	errAuthoritiesAlreadyInitialized   = errors.New("Grandpa: Authorities are already initialized!")
	errAuthoritiesExceedMaxAuthorities = errors.New("Grandpa: The number of authorities given should be lower than MaxAuthorities")
)

var (
	EngineId  = [4]byte{'F', 'R', 'N', 'K'}
	KeyTypeId = [4]byte{'g', 'r', 'a', 'n'}
)

const (
	functionReportEquivocationIndex = iota
	functionReportEquivocationUnsignedIndex
	functionNoteStalledIndex
)

type Module interface {
	primitives.Module

	KeyType() primitives.PublicKeyType
	KeyTypeId() [4]byte
	Authorities() (sc.Sequence[primitives.Authority], error)
	CurrentSetId() (sc.U64, error)
}

type module struct {
	primitives.DefaultInherentProvider
	hooks.DefaultDispatchModule

	index         sc.U8
	config        *Config
	constants     *consts
	storage       *storage
	functions     map[sc.U8]primitives.Call
	systemModule  system.Module
	sessionModule session.Module
	mdGenerator   *primitives.MetadataTypeGenerator
	logger        log.Logger
}

func New(index sc.U8, config *Config, logger log.Logger, mdGenerator *primitives.MetadataTypeGenerator) Module {
	functions := map[sc.U8]primitives.Call{}

	moduleInstance := module{
		index:         index,
		config:        config,
		constants:     newConstants(config.MaxAuthorities, config.MaxNominators, config.MaxSetIdSessionEntries),
		storage:       newStorage(),
		functions:     functions,
		mdGenerator:   mdGenerator,
		logger:        logger,
		systemModule:  config.SystemModule,
		sessionModule: config.SessionModule,
	}

	functions[functionReportEquivocationIndex] = newCallReportEquivocation(index, functionReportEquivocationIndex)
	functions[functionReportEquivocationUnsignedIndex] = newCallReportEquivocationUnsigned(index, functionReportEquivocationUnsignedIndex)
	functions[functionNoteStalledIndex] = newCallNoteStalled(index, functionNoteStalledIndex, moduleInstance)

	moduleInstance.functions = functions

	return moduleInstance
}

func (m module) GetIndex() sc.U8 {
	return m.index
}

func (m module) name() sc.Str {
	return name
}

func (m module) Functions() map[sc.U8]primitives.Call {
	return m.functions
}

func (m module) PreDispatch(_ primitives.Call) (sc.Empty, error) {
	return sc.Empty{}, nil
}

func (m module) ValidateUnsigned(_ primitives.TransactionSource, _ primitives.Call) (primitives.ValidTransaction, error) {
	return primitives.ValidTransaction{}, primitives.NewTransactionValidityError(primitives.NewUnknownTransactionNoUnsignedValidator())
}

// OneSessionHandler interface implementation

func (m module) KeyType() primitives.PublicKeyType {
	return m.config.KeyType
}

func (m module) KeyTypeId() [4]byte {
	return KeyTypeId
}

func (m module) DecodeKey(buffer *bytes.Buffer) (primitives.Ed25519PublicKey, error) {
	key, err := primitives.DecodeEd25519PublicKey(buffer)
	if err != nil {
		return primitives.Ed25519PublicKey{}, err
	}
	return key, nil
}

func (m module) OnGenesisSession(validators sc.Sequence[primitives.Validator]) error {
	authorities := primitives.AuthoritiesFrom(validators)
	return m.initialize(authorities)
}

func (m module) OnNewSession(changed bool, validators sc.Sequence[primitives.Validator], _queuedValidators sc.Sequence[primitives.Validator]) error {
	// Always issue a change if `session` says that the validators have changed.
	// Even if their session keys are the same as before, the underlying economic
	// identities have changed.
	var currentSetId sc.U64

	if changed || m.storage.Stalled.Exists() {
		nextAuthorities := primitives.AuthoritiesFrom(validators)

		stalledBytes, err := m.storage.Stalled.TakeBytes()
		if err != nil {
			return err
		}

		var scheduleChangeErr error
		if len(stalledBytes) != 0 {
			stalled, err := primitives.DecodeTuple2U64(bytes.NewBuffer(stalledBytes))
			if err != nil {
				return err
			}
			furtherWait, median := stalled.First, stalled.Second
			scheduleChangeErr = m.scheduleChange(nextAuthorities, furtherWait, sc.NewOption[sc.U64](median))
		} else {
			scheduleChangeErr = m.scheduleChange(nextAuthorities, 0, sc.NewOption[sc.U64](nil))
		}

		if scheduleChangeErr == nil {
			currentSetId, err = m.storage.CurrentSetId.Mutate(
				func(s *sc.U64) (sc.U64, error) {
					*s += 1
					return *s, nil
				},
			)
			if err != nil {
				return err
			}

			maxSetIdSessionEntries := sc.U64(math.Max(float64(m.constants.MaxSetIdSessionEntries), 1))

			if currentSetId >= maxSetIdSessionEntries {
				m.storage.SetIdSession.Remove(currentSetId - maxSetIdSessionEntries)
			}
		} else {
			// either the session module signalled that the validators have changed
			// or the set was stalled. but since we didn't successfully schedule
			// an authority set change we do not increment the set id.
			currentSetId, err = m.storage.CurrentSetId.Get()
			if err != nil {
				return err
			}
		}
	} else {
		// nothing's changed, neither economic conditions nor session keys. update the pointer
		// of the current set.
		var err error
		currentSetId, err = m.storage.CurrentSetId.Get()
		if err != nil {
			return err
		}
	}

	// update the mapping to note that the current set corresponds to the
	// latest equivalent session (i.e. now).
	sessionIndex, err := m.sessionModule.CurrentIndex()
	if err != nil {
		return err
	}

	m.storage.SetIdSession.Put(currentSetId, sessionIndex)
	return nil
}

func (m module) OnBeforeSessionEnding() {}

func (m module) OnDisabled(validatorIndex sc.U32) {
	m.depositLog(NewConsensusLogOnDisabled(sc.U64(validatorIndex)))
}

// Module hooks implementation

func (m module) OnFinalize(blockNumber sc.U64) error {
	// check for scheduled pending authority set changes
	pendingChangeBytes, err := m.storage.PendingChange.GetBytes()
	if err != nil {
		return err
	}

	if pendingChangeBytes.HasValue {
		buffer := bytes.NewBuffer(sc.SequenceU8ToBytes(pendingChangeBytes.Value))
		pendingChange, err := DecodeStoredPendingChange(buffer)
		if err != nil {
			return err
		}

		// emit signal if we're at the block that scheduled the change
		if blockNumber == pendingChange.ScheduledAt {
			nextAuthorities := pendingChange.NextAuthorities

			scheduledChange := ScheduledChange{
				NextAuthorities: nextAuthorities,
				Delay:           pendingChange.Delay,
			}

			if median := pendingChange.Forced.Value; pendingChange.Forced.HasValue {
				m.depositLog(NewConsensusLogForcedChange(median, scheduledChange))
			} else {
				m.depositLog(NewConsensusLogScheduledChange(scheduledChange))
			}
		}

		// enact the change if we've reached the enacting block
		if blockNumber == pendingChange.ScheduledAt+pendingChange.Delay {
			m.storage.Authorities.Put(pendingChange.NextAuthorities)
			m.systemModule.DepositEvent(newEventNewAuthorities(m.index, pendingChange.NextAuthorities))
			m.storage.PendingChange.Clear()
		}
	}

	// check for scheduled pending state changes
	state, err := m.storage.State.Get()
	if err != nil {
		return err
	}

	switch state.VaryingData[0].(sc.U8) {
	case StoredStatePendingPause:
		// signal change to pause
		action := state.VaryingData[1].(ScheduledAction)
		if blockNumber == action.ScheduledAt {
			m.depositLog(NewConsensusLogPause(action.Delay))
		}

		// enact change to paused state
		if blockNumber == action.ScheduledAt+action.Delay {
			m.storage.State.Put(NewStoredStatePaused())
			m.systemModule.DepositEvent(newEventPaused(m.index))
		}
	case StoredStatePendingResume:
		// signal change to resume
		action := state.VaryingData[1].(ScheduledAction)
		if blockNumber == action.ScheduledAt {
			m.depositLog(NewConsensusLogResume(action.Delay))
		}

		// enact change to live state
		if blockNumber == action.ScheduledAt+action.Delay {
			m.storage.State.Put(NewStoredStateLive())
			m.systemModule.DepositEvent(newEventResumed(m.index))
		}
	default:
	}

	return nil
}

// Get the current GRANDPA authorities and weights. This should not change except
// for when changes are scheduled and the corresponding delay has passed.
//
// When called at block B, it will return the set of authorities that should be
// used to finalize descendants of this block (B+1, B+2, ...). The block B itself
// is finalized by the authorities from block B-1.
func (m module) Authorities() (sc.Sequence[primitives.Authority], error) {
	return m.storage.Authorities.Get()
}

// Get current GRANDPA authority set id.
func (m module) CurrentSetId() (sc.U64, error) {
	return m.storage.CurrentSetId.Get()
}

// Schedule a change in the authorities.
//
// The change will be applied at the end of execution of the block
// `in_blocks` after the current block. This value may be 0, in which
// case the change is applied at the end of the current block.
//
// If the `forced` parameter is defined, this indicates that the current
// set has been synchronously determined to be offline and that after
// `in_blocks` the given change should be applied. The given block number
// indicates the median last finalized block number and it should be used
// as the canon block when starting the new grandpa voter.
//
// No change should be signaled while any change is pending. Returns
// an error if a change is already pending.
func (m module) scheduleChange(nextAuthorities sc.Sequence[primitives.Authority], inBlocks sc.U64, forced sc.Option[sc.U64]) error {
	if !m.storage.PendingChange.Exists() {
		scheduledAt, err := m.systemModule.StorageBlockNumber()
		if err != nil {
			return err
		}

		if forced.HasValue {
			next, err := m.storage.NextForced.Get()
			if err != nil || next > scheduledAt {
				return NewDispatchErrorTooSoon(m.index)
			}

			// only allow the next forced change when twice the window has passed since
			// this one.
			m.storage.NextForced.Put(scheduledAt + inBlocks*2)
		}

		if len(nextAuthorities) > int(m.config.MaxAuthorities) {
			return errAuthoritiesExceedMaxAuthorities
		}

		m.storage.PendingChange.Put(
			StoredPendingChange{
				Delay:           inBlocks,
				ScheduledAt:     scheduledAt,
				NextAuthorities: nextAuthorities,
				Forced:          forced,
			},
		)
		return nil
	} else {
		return NewDispatchErrorChangePending(m.index)
	}
}

// Deposit one of this module's logs.
func (m module) depositLog(log ConsensusLog) {
	m.systemModule.DepositLog(
		primitives.NewDigestItemConsensusMessage(
			sc.BytesToFixedSequenceU8(EngineId[:]),
			sc.BytesToSequenceU8(log.Bytes()),
		),
	)
}

// Perform module initialization, abstracted so that it can be called either through genesis
// config builder or through `on_genesis_session`.
func (m module) initialize(authorities sc.Sequence[primitives.Authority]) error {
	if len(authorities) != 0 {
		storageAuthorities, err := m.storage.Authorities.Get()
		if err != nil {
			return err
		}

		if len(storageAuthorities) > 0 {
			return errAuthoritiesAlreadyInitialized
		}

		if len(authorities) > int(m.config.MaxAuthorities) {
			return errAuthoritiesExceedMaxAuthorities
		}

		m.storage.Authorities.Put(authorities)
	}

	// NOTE: initialize first session of first set. this is necessary for
	// the genesis set and session since we only update the set -> session
	// mapping whenever a new session starts, i.e. through `on_new_session`.
	m.storage.SetIdSession.Put(0, 0)
	return nil
}

func (m module) onStalled(furtherWait sc.U64, median sc.U64) {
	// when we record old authority sets we could try to figure out _who_
	// failed. until then, we can't meaningfully guard against
	// `next == last` the way that normal session changes do.
	m.storage.Stalled.Put(primitives.Tuple2U64{First: furtherWait, Second: median})
}
