package grandpa

import (
	"bytes"
	"errors"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/primitives/session"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

var (
	errInvalidStoredStateType = errors.New("invalid 'StoredState' index")
)

type StaleNotifier interface {
	OnStalled(furtherWait sc.U64, median sc.U64)
}

// A stored pending change. `Limit` is the bound for `next_authorities`
type StoredPendingChange struct {
	// The block number this was scheduled at.
	ScheduledAt sc.U64
	// The delay in blocks until it will be applied.
	Delay sc.U64
	// The next authority set, weakly bounded in size by `Limit`.
	NextAuthorities sc.Sequence[primitives.Authority]
	// If defined it means the change was forced and the given block number
	// indicates the median last finalized block when the change was signaled.
	Forced sc.Option[sc.U64]
}

func (s StoredPendingChange) Encode(buffer *bytes.Buffer) error {
	return sc.EncodeEach(buffer,
		s.ScheduledAt,
		s.Delay,
		s.NextAuthorities,
		s.Forced,
	)
}

func (s StoredPendingChange) Bytes() []byte {
	return sc.EncodedBytes(s)
}

func DecodeStoredPendingChange(buffer *bytes.Buffer) (StoredPendingChange, error) {
	scheduledAt, err := sc.DecodeU64(buffer)
	if err != nil {
		return StoredPendingChange{}, err
	}

	delay, err := sc.DecodeU64(buffer)
	if err != nil {
		return StoredPendingChange{}, err
	}

	nextAuthorities, err := sc.DecodeSequenceWith(buffer, primitives.DecodeAuthority)
	if err != nil {
		return StoredPendingChange{}, err
	}

	forced, err := sc.DecodeOptionWith(buffer, sc.DecodeU64)
	if err != nil {
		return StoredPendingChange{}, err
	}

	return StoredPendingChange{
		ScheduledAt:     scheduledAt,
		Delay:           delay,
		NextAuthorities: nextAuthorities,
		Forced:          forced,
	}, nil
}

type ScheduledAction struct {
	// Block at which the action was scheduled.
	ScheduledAt sc.U64
	// Number of blocks after which the change will be enacted.
	Delay sc.U64
}

func (s ScheduledAction) Encode(buffer *bytes.Buffer) error {
	return sc.EncodeEach(buffer,
		s.ScheduledAt,
		s.Delay,
	)
}

func (s ScheduledAction) Bytes() []byte {
	return sc.EncodedBytes(s)
}

func DecodeScheduledAction(buffer *bytes.Buffer) (ScheduledAction, error) {
	scheduled, err := sc.DecodeU64(buffer)
	if err != nil {
		return ScheduledAction{}, err
	}

	delay, err := sc.DecodeU64(buffer)
	if err != nil {
		return ScheduledAction{}, err
	}

	return ScheduledAction{
		ScheduledAt: scheduled,
		Delay:       delay,
	}, nil
}

const (
	// The current authority set is live, and GRANDPA is enabled.
	StoredStateLive sc.U8 = iota

	// There is a pending pause event which will be enacted at the given block
	// height.
	StoredStatePendingPause

	// The current GRANDPA authority set is paused.
	StoredStatePaused

	// There is a pending resume event which will be enacted at the given block
	// height.
	StoredStatePendingResume
)

// Current state of the GRANDPA authority set. State transitions must happen in
// the same order of states defined below, e.g. `Paused` implies a prior
// `PendingPause`.
type StoredState struct {
	sc.VaryingData
}

func NewStoredStateLive() StoredState {
	return StoredState{sc.NewVaryingData(StoredStateLive)}
}

func NewStoredStatePendingPause(scheduledAction ScheduledAction) StoredState {
	return StoredState{sc.NewVaryingData(StoredStatePendingPause, scheduledAction)}
}

func NewStoredStatePaused() StoredState {
	return StoredState{sc.NewVaryingData(StoredStatePaused)}
}

func NewStoredStatePendingResume(scheduledAction ScheduledAction) StoredState {
	return StoredState{sc.NewVaryingData(StoredStatePendingResume, scheduledAction)}
}

func DecodeStoredState(buffer *bytes.Buffer) (StoredState, error) {
	index, err := sc.DecodeU8(buffer)
	if err != nil {
		return StoredState{}, err
	}

	switch index {
	case StoredStateLive:
		return NewStoredStateLive(), nil
	case StoredStatePendingPause:
		scheduledAction, err := DecodeScheduledAction(buffer)
		if err != nil {
			return StoredState{}, err
		}
		return NewStoredStatePendingPause(scheduledAction), nil
	case StoredStatePaused:
		return NewStoredStatePaused(), nil
	case StoredStatePendingResume:
		scheduledAction, err := DecodeScheduledAction(buffer)
		if err != nil {
			return StoredState{}, err
		}
		return NewStoredStatePendingResume(scheduledAction), nil
	default:
		return StoredState{}, errInvalidStoredStateType
	}
}

// Something which can compute and check proofs of
// a historical key owner and return full identification data of that
// key owner.
type KeyOwnerProofSystem interface {
	// Prove membership of a key owner in the current block-state.
	//
	// This should typically only be called off-chain, since it may be
	// computationally heavy.
	//
	// Returns `Some` iff the key owner referred to by the given `key` is a
	// member of the current set.
	Prove(key [4]byte, authorityId primitives.AccountId) sc.Option[session.MembershipProof]

	// TODO
	// Check a proof of membership on-chain. Return `Some` iff the proof is
	// valid and recent enough to check.
	// CheckProof(key Key, proof session.MembershipProof) sc.Option[IdentificationTuple]
}

type DefaultKeyOwnerProofSystem struct{}

func (d DefaultKeyOwnerProofSystem) Prove(key [4]byte, authorityId primitives.AccountId) sc.Option[session.MembershipProof] {
	return sc.NewOption[session.MembershipProof](nil)
}
