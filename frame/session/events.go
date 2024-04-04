package session

import (
	"bytes"
	"errors"
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/primitives/types"
)

// Session module events.
const (
	EventNewSession sc.U8 = iota
)

var (
	errInvalidEventModule = errors.New("invalid session.Event module")
	errInvalidEventType   = errors.New("invalid session.Event type")
)

func newEventNewSession(moduleIndex sc.U8, sessionIndex sc.U32) types.Event {
	return types.NewEvent(moduleIndex, EventNewSession, sessionIndex)
}

func DecodeEvent(moduleIndex sc.U8, buffer *bytes.Buffer) (types.Event, error) {
	decodedModuleIndex, err := sc.DecodeU8(buffer)
	if err != nil {
		return types.Event{}, err
	}
	if decodedModuleIndex != moduleIndex {
		return types.Event{}, errInvalidEventModule
	}

	b, err := sc.DecodeU8(buffer)
	if err != nil {
		return types.Event{}, err
	}

	switch b {
	case EventNewSession:
		sessionIndex, err := sc.DecodeU32(buffer)
		if err != nil {
			return types.Event{}, err
		}
		return newEventNewSession(moduleIndex, sessionIndex), nil
	default:
		return types.Event{}, errInvalidEventType
	}
}
