package aura_ext

import (
	"testing"

	"github.com/LimeChain/gosemble/mocks"
)

const (
	relayChainSlotDurationMillis = 6_000
	blockProcessingVelocity      = 1
	unincludedSegmentCapacity    = 1
)

var (
	mockRelayChainStateProof *mocks.RelayChainStateProof
)

func Test_FixedVelocityConsensusHook_OnStateProof(t *testing.T) {
	//target := setupFixedVelocityConsensusHook()

	//target.OnStateProof(mockRelayChainStateProof)
}

func setupFixedVelocityConsensusHook() FixedVelocityConsensusHook {
	mockRelayChainStateProof = new(mocks.RelayChainStateProof)
	module := setupModule()

	return NewFixedVelocityConsensusHook(relayChainSlotDurationMillis, blockProcessingVelocity, unincludedSegmentCapacity, dbWeight, module, logger)
}
