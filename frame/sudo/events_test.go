package sudo

import (
	"bytes"
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_DecodeEvent_Sudid(t *testing.T) {
	buffer := &bytes.Buffer{}
	buffer.WriteByte(moduleId)
	buffer.Write(EventSudid.Bytes())
	buffer.Write(dispatchOutcomeEmpty.Bytes())

	result, err := DecodeEvent(moduleId, buffer)
	assert.Nil(t, err)

	assert.Equal(t,
		types.Event{VaryingData: sc.NewVaryingData(sc.U8(moduleId), EventSudid, dispatchOutcomeEmpty)},
		result,
	)
}

func Test_DecodeEvent_KeyChanged(t *testing.T) {
	buffer := &bytes.Buffer{}
	buffer.WriteByte(moduleId)
	buffer.Write(EventKeyChanged.Bytes())
	buffer.Write(oldKeyOption.Bytes())
	buffer.Write(newKey.Bytes())

	event, err := DecodeEvent(moduleId, buffer)
	assert.Nil(t, err)

	assert.Equal(t,
		types.Event{VaryingData: sc.NewVaryingData(sc.U8(moduleId), EventKeyChanged, oldKeyOption, newKey)},
		event,
	)
}

func Test_DecodeEvent_KeyRemoved(t *testing.T) {
	buffer := &bytes.Buffer{}
	buffer.WriteByte(moduleId)
	buffer.Write(EventKeyRemoved.Bytes())

	event, err := DecodeEvent(moduleId, buffer)
	assert.Nil(t, err)

	assert.Equal(t,
		types.Event{VaryingData: sc.NewVaryingData(sc.U8(moduleId), EventKeyRemoved)},
		event,
	)
}

func Test_DecodeEvent_SudoAsDone(t *testing.T) {
	buffer := &bytes.Buffer{}
	buffer.WriteByte(moduleId)
	buffer.Write(EventSudoAsDone.Bytes())
	buffer.Write(dispatchOutcomeEmpty.Bytes())

	result, err := DecodeEvent(moduleId, buffer)
	assert.Nil(t, err)

	assert.Equal(t,
		types.Event{VaryingData: sc.NewVaryingData(sc.U8(moduleId), EventSudoAsDone, dispatchOutcomeEmpty)},
		result,
	)
}

func Test_DecodeEvent_InvalidModule(t *testing.T) {
	buffer := &bytes.Buffer{}
	buffer.WriteByte(5)

	_, err := DecodeEvent(moduleId, buffer)
	assert.Equal(t, errInvalidEventModule, err)
}

func Test_DecodeEvent_InvalidType(t *testing.T) {
	buffer := &bytes.Buffer{}
	buffer.WriteByte(moduleId)
	buffer.WriteByte(255)

	_, err := DecodeEvent(moduleId, buffer)
	assert.Equal(t, errInvalidEventType, err)
}
