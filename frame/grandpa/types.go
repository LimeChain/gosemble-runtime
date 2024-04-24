package grandpa

import (
	"bytes"
	"errors"

	sc "github.com/LimeChain/goscale"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

// type BoundedList[T sc.Encodable] struct {
// 	List  sc.Sequence[T]
// 	Limit sc.U32
// }

type StaleNotifier interface {
	onStalled(furtherWait sc.U64, median sc.U64)
}

// A scheduled change of authority set.
type ScheduledChange struct {
	// The new authorities after the change, along with their respective weights.
	NextAuthorities sc.Sequence[primitives.Authority]
	// The number of blocks to delay.
	Delay sc.U64
}

func (s ScheduledChange) Encode(buffer *bytes.Buffer) error {
	return sc.EncodeEach(buffer,
		s.NextAuthorities,
		s.Delay,
	)
}

func (s ScheduledChange) Bytes() []byte {
	return sc.EncodedBytes(s)
}

func DecodeScheduledChange(buffer *bytes.Buffer) (ScheduledChange, error) {
	nextAuthorities, err := sc.DecodeSequenceWith(buffer, primitives.DecodeAuthority)
	if err != nil {
		return ScheduledChange{}, err
	}

	delay, err := sc.DecodeU64(buffer)
	if err != nil {
		return ScheduledChange{}, err
	}

	return ScheduledChange{
		NextAuthorities: nextAuthorities,
		Delay:           delay,
	}, nil
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

// #[codec(mel_bound(N: MaxEncodedLen, Limit: Get<u32>))]

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
		return StoredState{}, errors.New("invalid 'StoredState' index")
	}
}
