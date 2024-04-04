package session

import (
	"bytes"
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	sessionIndex sc.U32 = 5
)

func Test_DecodeEvent_NewSession(t *testing.T) {
	buffer := &bytes.Buffer{}
	buffer.WriteByte(moduleId)
	buffer.Write(EventNewSession.Bytes())
	buffer.Write(sessionIndex.Bytes())

	result, err := DecodeEvent(moduleId, buffer)
	assert.Nil(t, err)

	assert.Equal(t,
		types.Event{VaryingData: sc.NewVaryingData(sc.U8(moduleId), EventNewSession, sessionIndex)},
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
