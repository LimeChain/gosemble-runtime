package grandpa

import (
	"bytes"
	"errors"

	sc "github.com/LimeChain/goscale"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

var (
	errInvalidEventModule = errors.New("invalid grandpa.Event module")
	errInvalidEventType   = errors.New("invalid grandpa.Event type")
)

const (
	// New authority set has been applied.
	EventNewAuthorities sc.U8 = iota
	// Current authority set has been paused.
	EventPaused
	// Current authority set has been resumed.
	EventResumed
)

func newEventNewAuthorities(moduleIndex sc.U8, authoritySet sc.Sequence[primitives.Authority]) primitives.Event {
	return primitives.NewEvent(moduleIndex, EventNewAuthorities, authoritySet)
}

func newEventPaused(moduleIndex sc.U8) primitives.Event {
	return primitives.NewEvent(moduleIndex, EventPaused)
}

func newEventResumed(moduleIndex sc.U8) primitives.Event {
	return primitives.NewEvent(moduleIndex, EventResumed)
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
	case EventNewAuthorities:
		authorities, err := primitives.DecodeAuthorityList(buffer)
		if err != nil {
			return primitives.Event{}, err
		}
		return newEventNewAuthorities(moduleIndex, authorities), nil
	case EventPaused:
		return newEventPaused(moduleIndex), nil
	case EventResumed:
		return newEventResumed(moduleIndex), nil
	default:
		return primitives.Event{}, errInvalidEventType
	}
}
