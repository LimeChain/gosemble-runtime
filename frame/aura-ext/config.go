package aura_ext

import (
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/primitives/types"
)

type Config struct {
	DbWeight                     types.RuntimeDbWeight
	RelayChainSlotDurationMillis sc.U32
	BlockProcessingVelocity      sc.U32
	NotIncludedSegmentCapacity   sc.U32
}

func NewConfig(dbWeight types.RuntimeDbWeight, relayChainSlotDurationMillis sc.U32, blockProcessingVelocity sc.U32, notIncludedSegmentCapacity sc.U32) Config {
	return Config{
		dbWeight,
		relayChainSlotDurationMillis,
		blockProcessingVelocity,
		notIncludedSegmentCapacity,
	}
}
