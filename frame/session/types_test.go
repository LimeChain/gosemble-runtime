package session

import (
	"bytes"
	"encoding/hex"
	"fmt"
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants"
	"github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io"
	"testing"
)

var (
	expectBytesQueuedKey, _ = hex.DecodeString("00000000000000000000000000000000000000000000000000000000000000010414010203040501020304")
)

var (
	targetQueuedKey = queuedKey{
		Validator: constants.OneAccountId,
		Keys: sc.Sequence[types.SessionKey]{
			{
				Key:    sc.Sequence[sc.U8]{1, 2, 3, 4, 5},
				TypeId: sc.FixedSequence[sc.U8]{1, 2, 3, 4},
			},
		},
	}
)

func Test_queuedKey_Encode(t *testing.T) {
	buffer := &bytes.Buffer{}

	err := targetQueuedKey.Encode(buffer)

	fmt.Println(hex.EncodeToString(buffer.Bytes()))

	assert.NoError(t, err)
	assert.Equal(t, expectBytesQueuedKey, buffer.Bytes())
}

func Test_DecodeQueuedKey(t *testing.T) {
	buffer := bytes.NewBuffer(expectBytesQueuedKey)

	result, err := DecodeQueuedKey(buffer)
	assert.NoError(t, err)

	assert.Equal(t, targetQueuedKey, result)
}

func Test_DecodeQueuedKey_EOF(t *testing.T) {
	result, err := DecodeQueuedKey(&bytes.Buffer{})
	assert.Error(t, io.EOF, err)

	assert.Equal(t, queuedKey{}, result)
}

func Test_DecodeQueuedKey_SessionKey_EOF(t *testing.T) {
	buffer := bytes.NewBuffer(constants.OneAccountId.Bytes())
	result, err := DecodeQueuedKey(buffer)
	assert.Error(t, io.EOF, err)

	assert.Equal(t, queuedKey{}, result)
}

func Test_QueuedKey_Bytes(t *testing.T) {
	assert.Equal(t, expectBytesQueuedKey, targetQueuedKey.Bytes())
}

func Test_DecodeSequenceAccountId(t *testing.T) {

	expect := sc.Sequence[queuedKey]{queuedKey{
		Validator: constants.ZeroAccountId,
		Keys:      sc.Sequence[types.SessionKey]{},
	}}

	buffer := &bytes.Buffer{}
	buffer.WriteByte(4)
	buffer.Write(constants.ZeroAccountId.Bytes())
	buffer.WriteByte(0)

	result, err := DecodeQueuedKeys(buffer)
	assert.Nil(t, err)
	assert.Equal(t, expect, result)
}

type MockKeyManager struct {
	mock.Mock
}

func (m *MockKeyManager) DoSetKeys(who types.AccountId, sessionKeys sc.Sequence[types.SessionKey]) error {
	args := m.Called(who, sessionKeys)

	if args.Get(0) == nil {
		return nil
	}

	return args.Get(0).(error)
}

func (m *MockKeyManager) DoPurgeKeys(who types.AccountId) error {
	args := m.Called(who)

	if args.Get(0) == nil {
		return nil
	}

	return args.Get(0).(error)
}
