package grandpa

import (
	"bytes"
	"encoding/hex"
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants"
	"github.com/LimeChain/gosemble/primitives/session"
	"github.com/stretchr/testify/assert"
)

var (
	storedPendingChange = StoredPendingChange{
		ScheduledAt:     sc.U64(1),
		Delay:           sc.U64(2),
		NextAuthorities: authorities,
		Forced:          sc.NewOption[sc.U64](sc.U64(3)),
	}

	scheduledAction = ScheduledAction{
		ScheduledAt: sc.U64(1),
		Delay:       sc.U64(2),
	}
)

var (
	storedPendingChangeBytes, _ = hex.DecodeString("010000000000000002000000000000000400000000000000000000000000000000000000000000000000000000000000010100000000000000010300000000000000")
	scheduledActionBytes, _     = hex.DecodeString("010000000000000002000000000000000")
)

func Test_StoredPendingChange_Encode(t *testing.T) {
	buffer := &bytes.Buffer{}

	err := storedPendingChange.Encode(buffer)

	assert.NoError(t, err)
	assert.Equal(t, storedPendingChangeBytes, buffer.Bytes())
}

func Test_StoredPendingChange_Bytes(t *testing.T) {
	assert.Equal(t, storedPendingChangeBytes, storedPendingChange.Bytes())
}

func Test_DecodeStoredPendingChange_Fails_To_Decode_ScheduledAt(t *testing.T) {
	buffer := bytes.NewBuffer(storedPendingChangeBytes[:1])

	_, err := DecodeStoredPendingChange(buffer)

	assert.Error(t, err)
}

func Test_DecodeStoredPendingChange_Fails_To_Decode_Delay(t *testing.T) {
	buffer := bytes.NewBuffer(storedPendingChangeBytes[:8])

	_, err := DecodeStoredPendingChange(buffer)

	assert.Error(t, err)
}

func Test_DecodeStoredPendingChange_Fails_To_Decode_NextAuthorities(t *testing.T) {
	buffer := bytes.NewBuffer(storedPendingChangeBytes[:16])

	_, err := DecodeStoredPendingChange(buffer)

	assert.Error(t, err)
}

func Test_DecodeStoredPendingChange_Fails_To_Decode_Forced(t *testing.T) {
	buffer := bytes.NewBuffer(storedPendingChangeBytes[:len(storedPendingChangeBytes)-1])

	_, err := DecodeStoredPendingChange(buffer)

	assert.Error(t, err)
}

func Test_ScheduledAction_Encode(t *testing.T) {
	buffer := &bytes.Buffer{}

	err := scheduledAction.Encode(buffer)

	assert.NoError(t, err)
	assert.Equal(t, scheduledActionBytes, buffer.Bytes())
}

func Test_ScheduledAction_Bytes(t *testing.T) {
	assert.Equal(t, scheduledActionBytes, scheduledAction.Bytes())
}

func Test_DecodeScheduledAction_Fails_To_Decode_ScheduledAt(t *testing.T) {
	buffer := bytes.NewBuffer(scheduledActionBytes[:1])

	_, err := DecodeScheduledAction(buffer)

	assert.Error(t, err)
}

func Test_DecodeScheduledAction_Fails_To_Decode_Delay(t *testing.T) {
	buffer := bytes.NewBuffer(scheduledActionBytes[:8])

	_, err := DecodeScheduledAction(buffer)

	assert.Error(t, err)
}

func Test_NewStoredStateLive(t *testing.T) {
	assert.Equal(t, StoredState{sc.NewVaryingData(StoredStateLive)}, NewStoredStateLive())
}

func Test_NewStoredStatePendingPause(t *testing.T) {
	assert.Equal(t, StoredState{sc.NewVaryingData(StoredStatePendingPause, scheduledAction)}, NewStoredStatePendingPause(scheduledAction))
}
func Test_NewStoredStatePaused(t *testing.T) {
	assert.Equal(t, StoredState{sc.NewVaryingData(StoredStatePaused)}, NewStoredStatePaused())
}
func Test_NewStoredStatePendingResume(t *testing.T) {
	assert.Equal(t, StoredState{sc.NewVaryingData(StoredStatePendingResume, scheduledAction)}, NewStoredStatePendingResume(scheduledAction))
}

func Test_DecodeStoredState_StoredStateLive(t *testing.T) {
	buffer := bytes.NewBuffer(StoredStateLive.Bytes())

	result, err := DecodeStoredState(buffer)

	assert.NoError(t, err)
	assert.Equal(t, NewStoredStateLive(), result)
}

func Test_DecodeStoredState_StoredStatePendingPause(t *testing.T) {
	buffer := bytes.NewBuffer(StoredStatePendingPause.Bytes())
	buffer.Write(scheduledActionBytes)

	result, err := DecodeStoredState(buffer)

	assert.NoError(t, err)
	assert.Equal(t, NewStoredStatePendingPause(scheduledAction), result)
}

func Test_DecodeStoredState_StoredStatePaused(t *testing.T) {
	buffer := bytes.NewBuffer(StoredStatePaused.Bytes())

	result, err := DecodeStoredState(buffer)

	assert.NoError(t, err)
	assert.Equal(t, NewStoredStatePaused(), result)
}

func Test_DecodeStoredState_StoredStatePendingResume(t *testing.T) {
	buffer := bytes.NewBuffer(StoredStatePendingResume.Bytes())
	buffer.Write(scheduledActionBytes)

	result, err := DecodeStoredState(buffer)

	assert.NoError(t, err)
	assert.Equal(t, NewStoredStatePendingResume(scheduledAction), result)
}

func Test_DecodeStoredState_Fails_To_Decode(t *testing.T) {
	buffer := bytes.NewBuffer(sc.U8(4).Bytes())

	_, err := DecodeStoredState(buffer)

	assert.Equal(t, errInvalidStoredStateType, err)
}

func Test_DefaultKeyOwnerProofSystem_Prove(t *testing.T) {
	key := [4]byte{'t', 'e', 's', 't'}

	result := DefaultKeyOwnerProofSystem{}.Prove(key, constants.ZeroAccountId)

	assert.Equal(t, sc.NewOption[session.MembershipProof](nil), result)
}
