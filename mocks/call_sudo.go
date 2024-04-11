package mocks

import (
	"bytes"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

type CallSudo struct {
	Call
}

func (m *CallSudo) DecodeSudoArgs(buffer *bytes.Buffer, decodeCallFunc func(buffer *bytes.Buffer) (primitives.Call, error)) (primitives.Call, error) {
	args := m.Called(buffer, decodeCallFunc)

	if args.Get(1) == nil {
		return args.Get(0).(primitives.Call), nil
	}

	return args.Get(0).(primitives.Call), args.Get(1).(error)
}
