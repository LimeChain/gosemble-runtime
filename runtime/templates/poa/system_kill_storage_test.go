package main

import (
	"bytes"
	"testing"

	cscale "github.com/centrifuge/go-substrate-rpc-client/v4/scale"
	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	ctypes "github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/stretchr/testify/assert"
)

func Test_KillStorage_DispatchOutcome(t *testing.T) {
	rt, storage := newTestRuntime(t)
	metadata := runtimeMetadata(t, rt)

	runtimeVersion, err := rt.Version()
	assert.NoError(t, err)

	initializeBlock(t, rt, parentHash, stateRoot, extrinsicsRoot, blockNumber)

	keys := [][]byte{
		[]byte("testkey1"),
		[]byte("testkey2"),
	}

	call, err := ctypes.NewCall(metadata, "System.kill_storage", keys)
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

	(*storage).Put([]byte("testkey1"), []byte("testvalue1"))
	(*storage).Put([]byte("testkey2"), []byte("testvalue2"))
	(*storage).Put([]byte("testkey3"), []byte("testvalue3"))

	assert.Equal(t, "testvalue1", string((*storage).Get([]byte("testkey1"))))
	assert.Equal(t, "testvalue2", string((*storage).Get([]byte("testkey2"))))

	res, err := rt.Exec("BlockBuilder_apply_extrinsic", extEnc.Bytes())
	assert.NoError(t, err)

	assert.Equal(t, []byte(nil), (*storage).Get([]byte("testkey1")))
	assert.Equal(t, []byte(nil), (*storage).Get([]byte("testkey2")))
	assert.Equal(t, "testvalue3", string((*storage).Get([]byte("testkey3"))))

	assert.Equal(t, applyExtrinsicResultOutcome.Bytes(), res)
}
