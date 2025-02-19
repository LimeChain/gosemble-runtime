package main

import (
	"bytes"
	"testing"

	"github.com/ChainSafe/gossamer/lib/common"
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/frame/aura"
	"github.com/LimeChain/gosemble/frame/grandpa"
	"github.com/LimeChain/gosemble/primitives/types"
	"github.com/LimeChain/gosemble/testhelpers"
	"github.com/stretchr/testify/assert"
)

func Test_SessionKeys_Generate_Session_Keys(t *testing.T) {
	rt, _ := testhelpers.NewRuntimeInstance(t)
	// TODO: not working with seed, Gossamer fails
	//seed := []byte("//Alice")
	option := sc.NewOption[sc.U8](nil)

	assert.Equal(t, 0, rt.Keystore().Aura.Size())
	assert.Equal(t, 0, rt.Keystore().Gran.Size())

	result, err := rt.Exec("SessionKeys_generate_session_keys", option.Bytes())
	assert.NoError(t, err)

	assert.Equal(t, 1, rt.Keystore().Aura.Size())
	assert.Equal(t, 1, rt.Keystore().Gran.Size())

	buffer := bytes.NewBuffer(result)

	seq, err := sc.DecodeSequence[sc.U8](buffer)
	assert.Nil(t, err)
	buffer.Reset()
	buffer.Write(sc.SequenceU8ToBytes(seq))

	auraKey, err := types.DecodeAccountId(buffer)
	assert.Nil(t, err)
	grandpaKey, err := types.DecodeAccountId(buffer)
	assert.Nil(t, err)

	assert.Equal(t, rt.Keystore().Aura.PublicKeys()[0].Encode(), auraKey.Bytes())
	assert.Equal(t, rt.Keystore().Gran.PublicKeys()[0].Encode(), grandpaKey.Bytes())
}

func Test_SessionKeys_Decode_Session_Keys(t *testing.T) {
	rt, _ := testhelpers.NewRuntimeInstance(t)

	auraKey := common.MustHexToBytes("0x88dc3417d5058ec4b4503e0c12ea1a0a89be200fe98922423d4334014fa6b0ee")
	grandpaKey := common.MustHexToBytes("0x88dc3417d5058ec4b4503e0c12ea1a0a89be200fe98922423d4334014fa6b0ef")

	sessionKeys := sc.Sequence[types.SessionKey]{
		types.NewSessionKey(auraKey, aura.KeyTypeId),
		types.NewSessionKey(grandpaKey, grandpa.KeyTypeId),
	}
	expectedResult := sc.NewOption[sc.Sequence[types.SessionKey]](sessionKeys)

	encodedKeys := sc.BytesToSequenceU8(append(auraKey, grandpaKey...)).Bytes()

	result, err := rt.Exec("SessionKeys_decode_session_keys", encodedKeys)
	assert.NoError(t, err)

	assert.Equal(t, expectedResult.Bytes(), result)
}
