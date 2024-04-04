package aura

import (
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/primitives/types"
)

const (
	ConsensusLogAuthoritiesChange sc.I8 = iota
	ConsensusLogOnDisabled
)

type ConsensusLog struct {
	sc.VaryingData
}

func NewConsensusLogAuthoritiesChange(authorities sc.Sequence[types.Sr25519PublicKey]) ConsensusLog {
	return ConsensusLog{sc.NewVaryingData(ConsensusLogAuthoritiesChange, authorities)}
}

func NewConsensusLogOnDisabled(authorityIndex sc.U32) ConsensusLog {
	return ConsensusLog{sc.NewVaryingData(ConsensusLogOnDisabled, authorityIndex)}
}
