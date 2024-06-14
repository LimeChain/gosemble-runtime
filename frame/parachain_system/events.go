package parachain_system

import (
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/primitives/parachain"
	"github.com/LimeChain/gosemble/primitives/types"
)

const (
	EventValidationFunctionStored sc.U8 = iota
	EventValidationFunctionApplied
	EventValidationFunctionDiscarded
	EventDownwardMessagesReceived
	EventDownwardMessagesProcessed
	EventUpwardMessageSent
)

func newEventValidationFunctionStored(moduleIndex sc.U8) types.Event {
	return types.NewEvent(moduleIndex, EventValidationFunctionStored)
}

func newEventValidationFunctionApplied(moduleIndex sc.U8, number parachain.RelayChainBlockNumber) types.Event {
	return types.NewEvent(moduleIndex, EventValidationFunctionApplied, number)
}

func newEventValidationFunctionDiscarded(moduleIndex sc.U8) types.Event {
	return types.NewEvent(moduleIndex, EventValidationFunctionDiscarded)
}

func newEventDownwardMessagesReceived(moduleIndex sc.U8, count sc.U32) types.Event {
	return types.NewEvent(moduleIndex, EventDownwardMessagesReceived, count)
}

func newEventDownwardMessagesProcessed(moduleIndex sc.U8, weight types.Weight, dmqHead types.H256) types.Event {
	return types.NewEvent(moduleIndex, EventDownwardMessagesProcessed, weight, dmqHead)
}

func newEventUpwardMessageSent(moduleIndex sc.U8, messageHash sc.Option[types.H256]) types.Event {
	return types.NewEvent(moduleIndex, EventUpwardMessageSent, messageHash)
}
