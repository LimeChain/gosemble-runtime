package parachain

import (
	"errors"
	sc "github.com/LimeChain/goscale"
)

const (
	ErrorStateProofRootMismatch sc.U8 = iota
	ErrorStateProofReadEntry
	ErrorStateProofSlot
	ErrorStateProofUpgradeGoAhead
	ErrorStateProofUpgradeRestriction
	ErrorStateProofConfig
	ErrorStateProofDmqMqcHead
	ErrorStateProofRelayDispatchQueueRemainingCapacity
	ErrorStateProofHrmpIngressChannelIndex
	ErrorStateProofHrmpEgressChannelIndex
	ErrorStateProofHrmpChannel
	ErrorStateProofParaHead
)

type StateProofError struct {
	sc.VaryingData
}

func NewErrorStateProofRootMismatch() StateProofError {
	return StateProofError{sc.NewVaryingData(ErrorStateProofRootMismatch)}
}

func NewErrorStateProofReadEntry() StateProofError {
	return StateProofError{sc.NewVaryingData(ErrorStateProofReadEntry)}
}

func NewErrorStateProofSlot(entryError ReadEntryError) StateProofError {
	return StateProofError{sc.NewVaryingData(ErrorStateProofSlot, entryError)}
}

func NewErrorStateProofUpgradeGoAhead(entryError ReadEntryError) StateProofError {
	return StateProofError{sc.NewVaryingData(ErrorStateProofUpgradeGoAhead, entryError)}
}

func NewErrorStateProofUpgradeRestriction(entryError ReadEntryError) StateProofError {
	return StateProofError{sc.NewVaryingData(ErrorStateProofUpgradeRestriction, entryError)}
}

func NewErrorStateProofConfig(entryError ReadEntryError) StateProofError {
	return StateProofError{sc.NewVaryingData(ErrorStateProofConfig, entryError)}
}

func NewErrorStateProofDmqMqcHead(entryError ReadEntryError) StateProofError {
	return StateProofError{sc.NewVaryingData(ErrorStateProofDmqMqcHead, entryError)}
}

func NewErrorStateProofRelayDispatchQueueRemainingCapacity(entryError ReadEntryError) StateProofError {
	return StateProofError{sc.NewVaryingData(ErrorStateProofRelayDispatchQueueRemainingCapacity, entryError)}
}

func NewErrorStateProofHrmpIngressChannelIndex(entryError ReadEntryError) StateProofError {
	return StateProofError{sc.NewVaryingData(ErrorStateProofHrmpIngressChannelIndex, entryError)}
}

func NewErrorStateProofHrmpEgressChannelIndex(entryError ReadEntryError) StateProofError {
	return StateProofError{sc.NewVaryingData(ErrorStateProofHrmpEgressChannelIndex, entryError)}
}

func NewErrorStateProofHrmpChannel(from sc.U32, to sc.U32, entryError ReadEntryError) StateProofError {
	return StateProofError{sc.NewVaryingData(ErrorStateProofHrmpChannel, from, to, entryError)}
}

func NewErrorStateProofParaHead(entryError ReadEntryError) StateProofError {
	return StateProofError{sc.NewVaryingData(ErrorStateProofParaHead, entryError)}
}

func (err StateProofError) Error() string {
	switch err.VaryingData[0] {
	case ErrorStateProofRootMismatch:
		return "State Proof Root Mismatch"
	case ErrorStateProofReadEntry:
		return "State Proof Read Entry"
	case ErrorStateProofSlot:
		return "State Proof Slot"
	case ErrorStateProofUpgradeGoAhead:
		return "State Proof Upgrade GoAhead"
	case ErrorStateProofUpgradeRestriction:
		return "State Proof Upgrade Restriction"
	case ErrorStateProofConfig:
		return "State Proof Config"
	case ErrorStateProofDmqMqcHead:
		return "State Proof DmqMqcHead"
	case ErrorStateProofRelayDispatchQueueRemainingCapacity:
		return "State Proof RelayDispatchQueueRemainingCapacity"
	case ErrorStateProofHrmpIngressChannelIndex:
		return "State Proof HrmpIngressChannelIndex"
	case ErrorStateProofHrmpEgressChannelIndex:
		return "State Proof HrmpEgressChannelIndex"
	case ErrorStateProofHrmpChannel:
		return "State Proof HrmpChannel"
	case ErrorStateProofParaHead:
		return "State Proof ParaHead"
	default:
		return errors.New("invalid StateProofError").Error()
	}
}
