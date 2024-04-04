package mocks

import (
	"bytes"
	sc "github.com/LimeChain/goscale"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/mock"
)

type OneSessionHandler struct {
	mock.Mock
}

func (m *OneSessionHandler) KeyType() primitives.PublicKeyType {
	args := m.Called()

	return args.Get(0).(primitives.PublicKeyType)
}

func (m *OneSessionHandler) KeyTypeId() [4]byte {
	args := m.Called()

	return args.Get(0).([4]byte)
}

func (m *OneSessionHandler) DecodeKey(buffer *bytes.Buffer) (primitives.Sr25519PublicKey, error) {
	args := m.Called(buffer)

	if args.Get(1) == nil {
		return args.Get(0).(primitives.Sr25519PublicKey), nil
	}

	return args.Get(0).(primitives.Sr25519PublicKey), args.Get(1).(error)
}

func (m *OneSessionHandler) OnGenesisSession(validators sc.Sequence[primitives.Validator]) error {
	args := m.Called(validators)

	if args.Get(0) == nil {
		return nil
	}

	return args.Get(0).(error)
}

func (m *OneSessionHandler) OnNewSession(changed bool, validators sc.Sequence[primitives.Validator], queuedValidators sc.Sequence[primitives.Validator]) error {
	args := m.Called(changed, validators, queuedValidators)

	if args.Get(0) == nil {
		return nil
	}

	return args.Get(0).(error)
}

func (m *OneSessionHandler) OnBeforeSessionEnding() {
	m.Called()
}

func (m *OneSessionHandler) OnDisabled(validatorIndex sc.U32) {
	m.Called(validatorIndex)
}
