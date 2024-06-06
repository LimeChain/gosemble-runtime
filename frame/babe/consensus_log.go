package babe

import (
	sc "github.com/LimeChain/goscale"
	babetypes "github.com/LimeChain/gosemble/primitives/babe"
)

// An consensus log item for BABE.
const (
	_ sc.U8 = iota
	// The epoch has changed. This provides information about the _next_
	// epoch - information about the _current_ epoch (i.e. the one we've just
	// entered) should already be available earlier in the chain.
	ConsensusLogNextEpochData
	// Disable the authority with given index.
	ConsensusLogOnDisabled
	// The epoch has changed, and the epoch after the current one will
	// enact different epoch configurations.
	ConsensusLogNextConfigData
)

type ConsensusLog struct {
	sc.VaryingData
}

func NewConsensusLogNextEpochData(next NextEpochDescriptor) ConsensusLog {
	return ConsensusLog{sc.NewVaryingData(ConsensusLogNextEpochData, next)}
}

func NewConsensusLogOnDisabled(index babetypes.AuthorityIndex) ConsensusLog {
	return ConsensusLog{sc.NewVaryingData(ConsensusLogOnDisabled, index)}
}

func NewConsensusLogNextConfigData(nextConfigData NextConfigDescriptor) ConsensusLog {
	return ConsensusLog{sc.NewVaryingData(ConsensusLogNextConfigData, nextConfigData)}
}
