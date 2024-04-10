package sudo

import (
	"bytes"
	"errors"
	sc "github.com/LimeChain/goscale"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

// Sudo module events.
const (
	EventSudid sc.U8 = iota
	EventKeyChanged
	EventKeyRemoved
	EventSudoAsDone
)

var (
	errInvalidEventModule = errors.New("invalid sudo.Event module")
	errInvalidEventType   = errors.New("invalid sudo.Event type")
)

func newEventSudid(moduleIndex sc.U8, outcome primitives.DispatchOutcome) primitives.Event {
	return primitives.NewEvent(moduleIndex, EventSudid, outcome)
}

func newEventKeyChanged(moduleIndex sc.U8, old sc.Option[primitives.AccountId], new primitives.AccountId) primitives.Event {
	return primitives.NewEvent(moduleIndex, EventKeyChanged, old, new)
}

func newEventKeyRemoved(moduleIndex sc.U8) primitives.Event {
	return primitives.NewEvent(moduleIndex, EventKeyRemoved)
}

func newEventSudoAsDone(moduleIndex sc.U8, outcome primitives.DispatchOutcome) primitives.Event {
	return primitives.NewEvent(moduleIndex, EventSudoAsDone, outcome)
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
	case EventSudid:
		outcome, err := primitives.DecodeDispatchOutcome(buffer)
		if err != nil {
			return primitives.Event{}, err
		}
		return newEventSudid(moduleIndex, outcome), nil
	case EventKeyChanged:
		old, err := sc.DecodeOptionWith(buffer, primitives.DecodeAccountId)
		if err != nil {
			return primitives.Event{}, err
		}
		new, err := primitives.DecodeAccountId(buffer)
		if err != nil {
			return primitives.Event{}, err
		}

		return newEventKeyChanged(moduleIndex, old, new), nil
	case EventKeyRemoved:
		return newEventKeyRemoved(moduleIndex), nil
	case EventSudoAsDone:
		outcome, err := primitives.DecodeDispatchOutcome(buffer)
		if err != nil {
			return primitives.Event{}, err
		}
		return newEventSudoAsDone(moduleIndex, outcome), nil
	default:
		return primitives.Event{}, errInvalidEventType
	}
}
