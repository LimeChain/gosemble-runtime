package balances

import (
	"bytes"
	"errors"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/frame/balances/types"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

// Balances module events.
const (
	EventEndowed sc.U8 = iota
	EventDustLost
	EventTransfer
	EventBalanceSet
	EventReserved
	EventUnreserved
	EventReserveRepatriated
	EventDeposit
	EventWithdraw
	EventSlashed
	EventUpgraded
	EventTotalIssuanceForced
)

var (
	errInvalidEventModule = errors.New("invalid balances.Event module")
	errInvalidEventType   = errors.New("invalid balances.Event type")
)

func newEventEndowed(moduleIndex sc.U8, account primitives.AccountId, freeBalance primitives.Balance) primitives.Event {
	return primitives.NewEvent(moduleIndex, EventEndowed, account, freeBalance)
}

func newEventDustLost(moduleIndex sc.U8, account primitives.AccountId, amount primitives.Balance) primitives.Event {
	return primitives.NewEvent(moduleIndex, EventDustLost, account, amount)
}

func newEventTransfer(moduleIndex sc.U8, from primitives.AccountId, to primitives.AccountId, amount primitives.Balance) primitives.Event {
	return primitives.NewEvent(moduleIndex, EventTransfer, from, to, amount)
}

func newEventBalanceSet(moduleIndex sc.U8, account primitives.AccountId, free primitives.Balance, reserved primitives.Balance) primitives.Event {
	return primitives.NewEvent(moduleIndex, EventBalanceSet, account, free, reserved)
}

func newEventReserved(moduleIndex sc.U8, account primitives.AccountId, amount primitives.Balance) primitives.Event {
	return primitives.NewEvent(moduleIndex, EventReserved, account, amount)
}

func newEventUnreserved(moduleIndex sc.U8, account primitives.AccountId, amount primitives.Balance) primitives.Event {
	return primitives.NewEvent(moduleIndex, EventUnreserved, account, amount)
}

func newEventReserveRepatriated(moduleIndex sc.U8, from primitives.AccountId, to primitives.AccountId, amount primitives.Balance, destinationStatus types.BalanceStatus) primitives.Event {
	return primitives.NewEvent(moduleIndex, EventReserveRepatriated, from, to, amount, destinationStatus)
}

func newEventDeposit(moduleIndex sc.U8, account primitives.AccountId, amount primitives.Balance) primitives.Event {
	return primitives.NewEvent(moduleIndex, EventDeposit, account, amount)
}

func newEventWithdraw(moduleIndex sc.U8, account primitives.AccountId, amount primitives.Balance) primitives.Event {
	return primitives.NewEvent(moduleIndex, EventWithdraw, account, amount)
}

func newEventSlashed(moduleIndex sc.U8, account primitives.AccountId, amount primitives.Balance) primitives.Event {
	return primitives.NewEvent(moduleIndex, EventSlashed, account, amount)
}

func newEventUpgraded(moduleIndex sc.U8, account primitives.AccountId) primitives.Event {
	return primitives.NewEvent(moduleIndex, EventUpgraded, account)
}

func newEventTotalIssuanceForced(moduleIndex sc.U8, old, new primitives.Balance) primitives.Event {
	return primitives.NewEvent(moduleIndex, EventTotalIssuanceForced, old, new)
}

func DecodeEvent(moduleIndex sc.U8, buffer *bytes.Buffer) (primitives.Event, error) {
	decodedModuleIndex, err := sc.DecodeU8(buffer)
	if err != nil {
		return primitives.Event{}, err
	}
	if decodedModuleIndex != moduleIndex {
		return primitives.Event{}, errInvalidEventModule
	}

	b, err := sc.DecodeU8(buffer)
	if err != nil {
		return primitives.Event{}, err
	}

	switch b {
	case EventEndowed:
		account, err := primitives.DecodeAccountId(buffer)
		if err != nil {
			return primitives.Event{}, err
		}
		freeBalance, err := sc.DecodeU128(buffer)
		if err != nil {
			return primitives.Event{}, err
		}
		return newEventEndowed(moduleIndex, account, freeBalance), nil
	case EventDustLost:
		account, err := primitives.DecodeAccountId(buffer)
		if err != nil {
			return primitives.Event{}, err
		}
		amount, err := sc.DecodeU128(buffer)
		if err != nil {
			return primitives.Event{}, err
		}
		return newEventDustLost(moduleIndex, account, amount), nil
	case EventTransfer:
		from, err := primitives.DecodeAccountId(buffer)
		if err != nil {
			return primitives.Event{}, err
		}
		to, err := primitives.DecodeAccountId(buffer)
		if err != nil {
			return primitives.Event{}, err
		}
		amount, err := sc.DecodeU128(buffer)
		if err != nil {
			return primitives.Event{}, err
		}
		return newEventTransfer(moduleIndex, from, to, amount), nil
	case EventBalanceSet:
		account, err := primitives.DecodeAccountId(buffer)
		if err != nil {
			return primitives.Event{}, err
		}
		free, err := sc.DecodeU128(buffer)
		if err != nil {
			return primitives.Event{}, err
		}
		reserved, err := sc.DecodeU128(buffer)
		if err != nil {
			return primitives.Event{}, err
		}
		return newEventBalanceSet(moduleIndex, account, free, reserved), nil
	case EventReserved:
		account, err := primitives.DecodeAccountId(buffer)
		if err != nil {
			return primitives.Event{}, err
		}
		amount, err := sc.DecodeU128(buffer)
		if err != nil {
			return primitives.Event{}, err
		}
		return newEventReserved(moduleIndex, account, amount), nil
	case EventUnreserved:
		account, err := primitives.DecodeAccountId(buffer)
		if err != nil {
			return primitives.Event{}, err
		}
		amount, err := sc.DecodeU128(buffer)
		if err != nil {
			return primitives.Event{}, err
		}
		return newEventUnreserved(moduleIndex, account, amount), nil
	case EventReserveRepatriated:
		from, err := primitives.DecodeAccountId(buffer)
		if err != nil {
			return primitives.Event{}, err
		}
		to, err := primitives.DecodeAccountId(buffer)
		if err != nil {
			return primitives.Event{}, err
		}
		amount, err := sc.DecodeU128(buffer)
		if err != nil {
			return primitives.Event{}, err
		}
		destinationStatus, err := types.DecodeBalanceStatus(buffer)
		if err != nil {
			return primitives.Event{}, err
		}
		return newEventReserveRepatriated(moduleIndex, from, to, amount, destinationStatus), nil
	case EventDeposit:
		account, err := primitives.DecodeAccountId(buffer)
		if err != nil {
			return primitives.Event{}, err
		}
		amount, err := sc.DecodeU128(buffer)
		if err != nil {
			return primitives.Event{}, err
		}
		return newEventDeposit(moduleIndex, account, amount), nil
	case EventWithdraw:
		account, err := primitives.DecodeAccountId(buffer)
		if err != nil {
			return primitives.Event{}, err
		}
		amount, err := sc.DecodeU128(buffer)
		if err != nil {
			return primitives.Event{}, err
		}
		return newEventWithdraw(moduleIndex, account, amount), nil
	case EventSlashed:
		account, err := primitives.DecodeAccountId(buffer)
		if err != nil {
			return primitives.Event{}, err
		}
		amount, err := sc.DecodeU128(buffer)
		if err != nil {
			return primitives.Event{}, err
		}
		return newEventSlashed(moduleIndex, account, amount), nil
	case EventUpgraded:
		account, err := primitives.DecodeAccountId(buffer)
		if err != nil {
			return primitives.Event{}, err
		}
		return newEventUpgraded(moduleIndex, account), nil
	case EventTotalIssuanceForced:
		old, err := sc.DecodeU128(buffer)
		if err != nil {
			return primitives.Event{}, err
		}
		new, err := sc.DecodeU128(buffer)
		if err != nil {
			return primitives.Event{}, err
		}
		return newEventTotalIssuanceForced(moduleIndex, old, new), nil
	default:
		return primitives.Event{}, errInvalidEventType
	}
}
