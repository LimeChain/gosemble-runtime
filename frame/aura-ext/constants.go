package aura_ext

import (
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/primitives/types"
)

type consts struct {
	DbWeight                     types.RuntimeDbWeight
	RelayChainSlotDurationMillis sc.U32
	BlockProcessingVelocity      sc.U32
	NotIncludedSegmentCapacity   sc.U32
}

func newConstants(dbWeight types.RuntimeDbWeight, relayChainSlotDurationMillis, blockProcessingVelocity, notIncludedSegmentCapacity sc.U32) consts {
	return consts{
		dbWeight,
		relayChainSlotDurationMillis,
		blockProcessingVelocity,
		notIncludedSegmentCapacity,
	}
}
