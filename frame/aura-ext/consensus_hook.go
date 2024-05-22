package aura_ext

import (
	"errors"
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/primitives/log"
	"github.com/LimeChain/gosemble/primitives/parachain"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

type FixedVelocityConsensusHook struct {
	RelayChainSlotDurationMillis sc.U32
	BlockProcessingVelocity      sc.U32
	NotIncludedSegmentCapacity   sc.U32
	dbWeight                     primitives.RuntimeDbWeight
	module                       Module
	logger                       log.Logger
}

func NewFixedVelocityConsensusHook(relayChainSlotDurationMillis, blockProcessingVelocity, notIncludedSegmentCapacity sc.U32, DbWeight primitives.RuntimeDbWeight, module Module, logger log.Logger) FixedVelocityConsensusHook {
	return FixedVelocityConsensusHook{
		relayChainSlotDurationMillis,
		blockProcessingVelocity,
		notIncludedSegmentCapacity,
		DbWeight,
		module,
		logger,
	}
}

func (fvch FixedVelocityConsensusHook) OnStateProof(stateProof parachain.RelayChainStateProof) (primitives.Weight, parachain.UnincludedSegmentCapacity, error) {
	// TODO: add sc.MaxU32
	velocity := sc.Max64(sc.U64(fvch.BlockProcessingVelocity), 1)

	currentSlot, err := stateProof.ReadSlot()
	if err != nil {
		return primitives.WeightZero(), parachain.UnincludedSegmentCapacity{}, err
	}

	slotInfo, err := fvch.module.storage.SlotInfo.Get()
	if err != nil {
		return primitives.WeightZero(), parachain.UnincludedSegmentCapacity{}, err
	}

	relayChainTimestamp := sc.SaturatingMulU64(sc.U64(fvch.RelayChainSlotDurationMillis), currentSlot)

	paraSlotDuration := fvch.module.auraModule.SlotDuration()
	paraSlotFromRelay := relayChainTimestamp / paraSlotDuration

	if slotInfo.Slot != paraSlotFromRelay {
		return primitives.WeightZero(), parachain.UnincludedSegmentCapacity{}, errors.New("slot number mismatch")
	}
	if slotInfo.Authored > sc.U32(velocity)+1 {
		fvch.logger.Critical("authored blocks limit is reached for current slot")
	}

	weight := fvch.dbWeight.Reads(1)

	return weight, parachain.NewUnincludedSegmentCapacityValue(1), nil
}
