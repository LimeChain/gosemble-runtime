package main

import (
	"bytes"
	"testing"

	cscale "github.com/centrifuge/go-substrate-rpc-client/v4/scale"
	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	ctypes "github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/stretchr/testify/assert"
)

func Test_KillPrefix_DispatchOutcome(t *testing.T) {
	rt, storage := newTestRuntime(t)
	metadata := runtimeMetadata(t, rt)

	runtimeVersion, err := rt.Version()
	assert.NoError(t, err)

	initializeBlock(t, rt, parentHash, stateRoot, extrinsicsRoot, blockNumber)

	prefix := []byte("test")
	limit := uint32(2)

	call, err := ctypes.NewCall(metadata, "System.kill_prefix", prefix, limit)
	assert.NoError(t, err)

	extrinsic := ctypes.NewExtrinsic(call)

	o := ctypes.SignatureOptions{
		BlockHash:          ctypes.Hash(parentHash),
		Era:                ctypes.ExtrinsicEra{IsImmortalEra: true},
		GenesisHash:        ctypes.Hash(parentHash),
		Nonce:              ctypes.NewUCompactFromUInt(0),
		SpecVersion:        ctypes.U32(runtimeVersion.SpecVersion),
		Tip:                ctypes.NewUCompactFromUInt(0),
		TransactionVersion: ctypes.U32(runtimeVersion.TransactionVersion),
	}

	err = extrinsic.Sign(signature.TestKeyringPairAlice, o)
	assert.NoError(t, err)

	extEnc := bytes.Buffer{}
	encoder := cscale.NewEncoder(&extEnc)
	err = extrinsic.Encode(*encoder)
	assert.NoError(t, err)

	(*storage).Put([]byte("testkey111"), []byte("testvalue1"))
	(*storage).Put([]byte("testkey222"), []byte("testvalue2"))
	(*storage).Put([]byte("testkey333"), []byte("testvalue3"))
	(*storage).Put([]byte("key444"), []byte("testvalue4"))
	(*storage).Put([]byte("key555"), []byte("testvalue5"))

	assert.Equal(t, []byte("testvalue1"), (*storage).Get([]byte("testkey111")))
	assert.Equal(t, []byte("testvalue2"), (*storage).Get([]byte("testkey222")))
	assert.Equal(t, []byte("testvalue3"), (*storage).Get([]byte("testkey333")))
	assert.Equal(t, []byte("testvalue4"), (*storage).Get([]byte("key444")))
	assert.Equal(t, []byte("testvalue5"), (*storage).Get([]byte("key555")))

	res, err := rt.Exec("BlockBuilder_apply_extrinsic", extEnc.Bytes())
	assert.NoError(t, err)

	assert.Equal(t, []byte(nil), (*storage).Get([]byte("testkey111")))
	assert.Equal(t, []byte(nil), (*storage).Get([]byte("testkey222")))
	assert.Equal(t, []byte("testvalue3"), (*storage).Get([]byte("testkey333")))
	assert.Equal(t, []byte("testvalue4"), (*storage).Get([]byte("key444")))
	assert.Equal(t, []byte("testvalue5"), (*storage).Get([]byte("key555")))

	assert.Equal(t, applyExtrinsicResultOutcome.Bytes(), res)
}
