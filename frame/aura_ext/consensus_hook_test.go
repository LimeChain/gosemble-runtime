package aura_ext

import (
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/mocks"
	"github.com/LimeChain/gosemble/primitives/parachain"
	"github.com/stretchr/testify/assert"
)

var (
	relayChainSlotDurationMillis sc.U32 = 6_000
	blockProcessingVelocity      sc.U32 = 1
	unincludedSegmentCapacity    sc.U32 = 1

	relayChainSlot    = sc.U64(5)
	consensusSlotInfo = SlotInfo{
		Slot:     15,
		Authored: 1,
	}
	parachainSlotDuration = sc.U64(2_000)
)

var (
	mockRelayChainStateProof *mocks.RelayChainStateProof
)

func Test_FixedVelocityConsensusHook_OnStateProof(t *testing.T) {
	target := setupFixedVelocityConsensusHook()

	mockRelayChainStateProof.On("ReadSlot").Return(relayChainSlot, nil)
	mockSlotInfo.On("Get").Return(consensusSlotInfo, nil)
	mockAuraModule.On("SlotDuration").Return(parachainSlotDuration)

	weight, unincludedSegment, err := target.OnStateProof(mockRelayChainStateProof)
	assert.Nil(t, err)

	assert.Equal(t, dbWeight.Reads(1), weight)
	assert.Equal(t, parachain.NewUnincludedSegmentCapacityValue(1), unincludedSegment)
}

func setupFixedVelocityConsensusHook() FixedVelocityConsensusHook {
	mockRelayChainStateProof = new(mocks.RelayChainStateProof)
	module := setupModule()

	return NewFixedVelocityConsensusHook(relayChainSlotDurationMillis, blockProcessingVelocity, unincludedSegmentCapacity, dbWeight, module, logger)
}
