package session

import (
	"bytes"
	"errors"
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants"
	"github.com/LimeChain/gosemble/mocks"
	"github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	keyTypeId  = [4]byte{'t', 'e', 's', 't'}
	queuedKeys = sc.Sequence[queuedKey]{
		queuedKey{
			Validator: constants.OneAccountId,
			Keys: sc.Sequence[types.SessionKey]{
				{
					Key:    sc.BytesToSequenceU8(constants.OneAccountId.Bytes()),
					TypeId: sc.BytesToFixedSequenceU8(keyTypeId[:]),
				},
			},
		},
	}
	validators = sc.Sequence[types.Validator]{
		{
			AccountId:   constants.OneAccountId,
			AuthorityId: types.Sr25519PublicKey{FixedSequence: constants.OneAccountId.FixedSequence},
		},
	}
	keys = sc.Sequence[types.AccountId]{
		constants.OneAccountId,
	}
)

var (
	mockOneSessionHandler *mocks.OneSessionHandler
)

func Test_Handler_KeyTypeIds(t *testing.T) {
	target := setupHandler()
	mockOneSessionHandler.On("KeyTypeId").Return(keyTypeId)

	result := target.KeyTypeIds()

	assert.Equal(t, sc.Sequence[sc.FixedSequence[sc.U8]]{
		sc.BytesToFixedSequenceU8(keyTypeId[:]),
	}, result)

	mockOneSessionHandler.AssertCalled(t, "KeyTypeId")
}

func Test_Handler_DecodeKeys(t *testing.T) {
	key := types.Sr25519PublicKey{FixedSequence: constants.OneAddress.FixedSequence}
	target := setupHandler()
	buffer := bytes.NewBuffer(key.Bytes())

	mockOneSessionHandler.On("DecodeKey", buffer).Return(key, nil)

	res, err := target.DecodeKeys(buffer)
	assert.Nil(t, err)

	assert.Equal(t, sc.FixedSequence[types.Sr25519PublicKey]{key}, res)
	mockOneSessionHandler.AssertCalled(t, "DecodeKey", buffer)
}

func Test_Handler_DecodeKeys_Err(t *testing.T) {
	expectErr := errors.New("expect")
	target := setupHandler()
	buffer := &bytes.Buffer{}

	mockOneSessionHandler.On("DecodeKey", buffer).Return(types.Sr25519PublicKey{}, expectErr)

	res, err := target.DecodeKeys(buffer)
	assert.Equal(t, expectErr, err)
	assert.Equal(t, sc.FixedSequence[types.Sr25519PublicKey](nil), res)

	mockOneSessionHandler.AssertCalled(t, "DecodeKey", buffer)
}

func Test_Handler_OnGenesisSession(t *testing.T) {
	target := setupHandler()

	mockOneSessionHandler.On("KeyTypeId").Return(keyTypeId)
	mockOneSessionHandler.On("OnGenesisSession", validators).Return(nil)

	err := target.OnGenesisSession(queuedKeys)
	assert.Nil(t, err)

	mockOneSessionHandler.AssertCalled(t, "KeyTypeId")
	mockOneSessionHandler.AssertCalled(t, "OnGenesisSession", validators)
}

func Test_Handler_OnNewSession(t *testing.T) {
	target := setupHandler()

	mockOneSessionHandler.On("KeyTypeId").Return(keyTypeId)
	mockOneSessionHandler.On("OnNewSession", true, validators, validators).Return(nil)

	err := target.OnNewSession(true, queuedKeys, queuedKeys)
	assert.Nil(t, err)

	mockOneSessionHandler.AssertCalled(t, "KeyTypeId")
	mockOneSessionHandler.AssertNumberOfCalls(t, "KeyTypeId", 2)
	mockOneSessionHandler.AssertCalled(t, "OnNewSession", true, validators, validators)
}

func Test_Handler_OnBeforeSessionEnding(t *testing.T) {
	target := setupHandler()

	mockOneSessionHandler.On("OnBeforeSessionEnding").Return()

	target.OnBeforeSessionEnding()

	mockOneSessionHandler.AssertCalled(t, "OnBeforeSessionEnding")
}

func Test_Handler_OnDisabled(t *testing.T) {
	validatorIndex := sc.U32(1)
	target := setupHandler()

	mockOneSessionHandler.On("OnDisabled", validatorIndex).Return()

	target.OnDisabled(validatorIndex)

	mockOneSessionHandler.AssertCalled(t, "OnDisabled", validatorIndex)
}

func setupHandler() Handler {
	mockOneSessionHandler = new(mocks.OneSessionHandler)

	return NewHandler([]OneSessionHandler{mockOneSessionHandler})
}

type MockSessionHandler struct {
	mock.Mock
}

func (m *MockSessionHandler) KeyTypeIds() sc.Sequence[sc.FixedSequence[sc.U8]] {
	args := m.Called()

	return args.Get(0).(sc.Sequence[sc.FixedSequence[sc.U8]])
}

func (m *MockSessionHandler) DecodeKeys(buffer *bytes.Buffer) (sc.FixedSequence[types.Sr25519PublicKey], error) {
	args := m.Called(buffer)

	if args.Get(1) == nil {
		return args.Get(0).(sc.FixedSequence[types.Sr25519PublicKey]), nil
	}

	return args.Get(0).(sc.FixedSequence[types.Sr25519PublicKey]), args.Get(1).(error)
}

func (m *MockSessionHandler) OnGenesisSession(validators sc.Sequence[queuedKey]) error {
	args := m.Called(validators)

	if args.Get(0) == nil {
		return nil
	}

	return args.Get(0).(error)
}

func (m *MockSessionHandler) OnNewSession(changed bool, validators sc.Sequence[queuedKey], queuedValidators sc.Sequence[queuedKey]) error {
	args := m.Called(changed, validators, queuedValidators)

	if args.Get(0) == nil {
		return nil
	}

	return args.Get(0).(error)
}

func (m *MockSessionHandler) OnBeforeSessionEnding() {
	m.Called()
}

func (m *MockSessionHandler) OnDisabled(validatorIndex sc.U32) {
	m.Called(validatorIndex)
}

func (m *MockSessionHandler) AppendHandlers(module OneSessionHandler) {
	m.Called(module)
}
