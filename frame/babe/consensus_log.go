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
	NextEpochData
	// Disable the authority with given index.
	OnDisabled
	// The epoch has changed, and the epoch after the current one will
	// enact different epoch configurations.
	NextConfigData
)

type ConsensusLog struct {
	sc.VaryingData
}

func NewNextEpochDataConsensusLog(next NextEpochDescriptor) ConsensusLog {
	return ConsensusLog{sc.NewVaryingData(NextEpochData, next)}
}

func NewOnDisabledConsensusLog(index babetypes.AuthorityIndex) ConsensusLog {
	return ConsensusLog{sc.NewVaryingData(OnDisabled, index)}
}

func NewNextConfigDataConsensusLog(nextConfigData NextConfigDescriptor) ConsensusLog {
	return ConsensusLog{sc.NewVaryingData(NextConfigData, nextConfigData)}
}
