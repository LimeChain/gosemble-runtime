package main

import (
	"bytes"
	"testing"

	cscale "github.com/centrifuge/go-substrate-rpc-client/v4/scale"
	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	ctypes "github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/stretchr/testify/assert"
)

func Test_SetStorage_DispatchOutcome(t *testing.T) {
	rt, storage := newTestRuntime(t)
	metadata := runtimeMetadata(t, rt)

	runtimeVersion, err := rt.Version()
	assert.NoError(t, err)

	// Initialize block
	initializeBlock(t, rt, parentHash, stateRoot, extrinsicsRoot, blockNumber)

	items := []struct {
		Key   []byte
		Value []byte
	}{
		{
			Key:   []byte("testkey1"),
			Value: []byte("testvalue1"),
		},
		{
			Key:   []byte("testkey2"),
			Value: []byte("testvalue2"),
		},
	}

	call, err := ctypes.NewCall(metadata, "System.set_storage", items)
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

	// Sign the extrinsic
	err = extrinsic.Sign(signature.TestKeyringPairAlice, o)
	assert.NoError(t, err)

	extEnc := bytes.Buffer{}
	encoder := cscale.NewEncoder(&extEnc)
	err = extrinsic.Encode(*encoder)
	assert.NoError(t, err)

	assert.Equal(t, []byte(nil), (*storage).Get([]byte("testkey1")))
	assert.Equal(t, []byte(nil), (*storage).Get([]byte("testkey2")))

	res, err := rt.Exec("BlockBuilder_apply_extrinsic", extEnc.Bytes())
	assert.NoError(t, err)

	assert.Equal(t, []byte("testvalue1"), (*storage).Get([]byte("testkey1")))
	assert.Equal(t, []byte("testvalue2"), (*storage).Get([]byte("testkey2")))
	assert.Equal(t, []byte(nil), (*storage).Get([]byte("testkey3")))

	assert.Equal(t, applyExtrinsicResultOutcome.Bytes(), res)
}
