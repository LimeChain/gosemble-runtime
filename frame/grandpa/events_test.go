package grandpa

import (
	"bytes"
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
)

func Test_newEventNewAuthorities(t *testing.T) {
	assert.Equal(t,
		types.Event{VaryingData: sc.NewVaryingData(moduleId, EventNewAuthorities, authorities)},
		newEventNewAuthorities(moduleId, authorities),
	)
}

func Test_newEventPaused(t *testing.T) {
	assert.Equal(t,
		types.Event{VaryingData: sc.NewVaryingData(moduleId, EventPaused)},
		newEventPaused(moduleId),
	)
}

func Test_newEventResumed(t *testing.T) {
	assert.Equal(t,
		types.Event{VaryingData: sc.NewVaryingData(moduleId, EventResumed)},
		newEventResumed(moduleId),
	)
}

func Test_DecodeEvent_EventNewAuthorities(t *testing.T) {
	buffer := &bytes.Buffer{}
	buffer.Write(moduleId.Bytes())
	buffer.Write(EventNewAuthorities.Bytes())
	buffer.Write(authorities.Bytes())

	event, err := DecodeEvent(moduleId, buffer)

	assert.Nil(t, err)
	assert.Equal(t,
		types.Event{VaryingData: sc.NewVaryingData(sc.U8(moduleId), EventNewAuthorities, authorities)},
		event,
	)
}

func Test_DecodeEvent_EventPaused(t *testing.T) {
	buffer := &bytes.Buffer{}
	buffer.Write(moduleId.Bytes())
	buffer.Write(EventPaused.Bytes())

	event, err := DecodeEvent(moduleId, buffer)

	assert.Nil(t, err)
	assert.Equal(t,
		types.Event{VaryingData: sc.NewVaryingData(sc.U8(moduleId), EventPaused)},
		event,
	)
}

func Test_DecodeEvent_EventResumed(t *testing.T) {
	buffer := &bytes.Buffer{}
	buffer.Write(moduleId.Bytes())
	buffer.Write(EventResumed.Bytes())

	event, err := DecodeEvent(moduleId, buffer)

	assert.Nil(t, err)
	assert.Equal(t,
		types.Event{VaryingData: sc.NewVaryingData(sc.U8(moduleId), EventResumed)},
		event,
	)
}

func Test_DecodeEvent_Fails_With_Event_Module_Error(t *testing.T) {
	buffer := &bytes.Buffer{}
	buffer.Write(sc.U8(100).Bytes())

	_, err := DecodeEvent(moduleId, buffer)

	assert.Equal(t, errInvalidEventModule, err)
}

func Test_DecodeEvent_Fails_With_Event_Type_Error(t *testing.T) {
	buffer := &bytes.Buffer{}
	buffer.Write(moduleId.Bytes())
	buffer.Write(sc.U8(3).Bytes())

	_, err := DecodeEvent(moduleId, buffer)

	assert.Equal(t, errInvalidEventType, err)
}
