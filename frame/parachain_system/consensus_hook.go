package parachain_system

import (
	"github.com/LimeChain/gosemble/primitives/parachain"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

type ConsensusHook interface {
	OnStateProof(proof parachain.RelayChainStateProof) (primitives.Weight, parachain.UnincludedSegmentCapacity, error)
}
