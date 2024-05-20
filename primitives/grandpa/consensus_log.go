package grandpa

import (
	sc "github.com/LimeChain/goscale"
)

// An consensus log item for GRANDPA.
const (
	_ sc.U8 = iota

	// Schedule an authority set change.
	//
	// The earliest digest of this type in a single block will be respected,
	// provided that there is no `ForcedChange` digest. If there is, then the
	// `ForcedChange` will take precedence.
	//
	// No change should be scheduled if one is already and the delay has not
	// passed completely.
	//
	// This should be a pure function: i.e. as long as the runtime can interpret
	// the digest type it should return the same result regardless of the current
	// state.
	ConsensusLogScheduledChange

	// Force an authority set change.
	//
	// Forced changes are applied after a delay of _imported_ blocks,
	// while pending changes are applied after a delay of _finalized_ blocks.
	//
	// The earliest digest of this type in a single block will be respected,
	// with others ignored.
	//
	// No change should be scheduled if one is already and the delay has not
	// passed completely.
	//
	// This should be a pure function: i.e. as long as the runtime can interpret
	// the digest type it should return the same result regardless of the current
	// state.
	ConsensusLogForcedChange

	// Note that the authority with given index is disabled until the next change.
	ConsensusLogOnDisabled

	// A signal to pause the current authority set after the given delay.
	// After finalizing the block at _delay_ the authorities should stop voting.
	ConsensusLogPause

	// A signal to resume the current authority set after the given delay.
	// After authoring the block at _delay_ the authorities should resume voting.
	ConsensusLogResume
)

type ConsensusLog struct {
	sc.VaryingData
}

func NewConsensusLogScheduledChange(scheduledChange ScheduledChange) ConsensusLog {
	return ConsensusLog{sc.NewVaryingData(ConsensusLogScheduledChange, scheduledChange)}
}

func NewConsensusLogForcedChange(median sc.U64, scheduledChange ScheduledChange) ConsensusLog {
	return ConsensusLog{sc.NewVaryingData(ConsensusLogForcedChange, median, scheduledChange)}
}

func NewConsensusLogOnDisabled(authorityIndex sc.U64) ConsensusLog {
	return ConsensusLog{sc.NewVaryingData(ConsensusLogOnDisabled, authorityIndex)}
}

func NewConsensusLogPause(_ sc.U64) ConsensusLog {
	return ConsensusLog{sc.NewVaryingData(ConsensusLogPause)}
}

func NewConsensusLogResume(_ sc.U64) ConsensusLog {
	return ConsensusLog{sc.NewVaryingData(ConsensusLogResume)}
}
