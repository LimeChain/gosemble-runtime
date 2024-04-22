package main

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/LimeChain/gosemble/testhelpers"
	cscale "github.com/centrifuge/go-substrate-rpc-client/v4/scale"
	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	ctypes "github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/stretchr/testify/assert"
)

func Test_Sudo_SetKey_Success(t *testing.T) {
	rt, storage := testhelpers.NewRuntimeInstance(t)
	metadata := testhelpers.RuntimeMetadata(t, rt)

	runtimeVersion, err := rt.Version()
	assert.NoError(t, err)

	// Set account info
	balance, e := big.NewInt(0).SetString("50000000000000000", 10)
	assert.True(t, e)

	testhelpers.SetStorageAccountInfo(t, storage, signature.TestKeyringPairAlice.PublicKey, balance, 0)
	testhelpers.InitializeBlock(t, rt, testhelpers.ParentHash, testhelpers.StateRoot, testhelpers.ExtrinsicsRoot, testhelpers.BlockNumber)

	bob, err := ctypes.NewMultiAddressFromHexAccountID(
		"0x90b5ab205c6974c9ea841be688864633dc9ca8a357843eeacf2314649965fe22")
	assert.NoError(t, err)

	// Set Sudo Key
	err = (*storage).Put(append(testhelpers.KeySudoHash, testhelpers.KeyKeyHash...), signature.TestKeyringPairAlice.PublicKey)
	assert.NoError(t, err)

	call, err := ctypes.NewCall(metadata, "Sudo.set_key", bob)
	assert.NoError(t, err)

	extrinsic := ctypes.NewExtrinsic(call)

	o := ctypes.SignatureOptions{
		BlockHash:          ctypes.Hash(testhelpers.ParentHash),
		Era:                ctypes.ExtrinsicEra{IsImmortalEra: true},
		GenesisHash:        ctypes.Hash(testhelpers.ParentHash),
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

	res, err := rt.Exec("BlockBuilder_apply_extrinsic", extEnc.Bytes())
	assert.NoError(t, err)

	assert.Equal(t, testhelpers.ApplyExtrinsicResultOutcome.Bytes(), res)
	assert.Equal(t, bob.AsID.ToBytes(), (*storage).Get(append(testhelpers.KeySudoHash, testhelpers.KeyKeyHash...)))
}

func Test_Sudo_SetKey_RequireSudo(t *testing.T) {
	rt, storage := testhelpers.NewRuntimeInstance(t)
	metadata := testhelpers.RuntimeMetadata(t, rt)

	runtimeVersion, err := rt.Version()
	assert.NoError(t, err)

	// Set account info
	balance, e := big.NewInt(0).SetString("50000000000000000", 10)
	assert.True(t, e)

	testhelpers.SetStorageAccountInfo(t, storage, signature.TestKeyringPairAlice.PublicKey, balance, 0)
	testhelpers.InitializeBlock(t, rt, testhelpers.ParentHash, testhelpers.StateRoot, testhelpers.ExtrinsicsRoot, testhelpers.BlockNumber)

	bob, err := ctypes.NewMultiAddressFromHexAccountID(
		"0x90b5ab205c6974c9ea841be688864633dc9ca8a357843eeacf2314649965fe22")
	assert.NoError(t, err)

	// Set bob Sudo Key
	err = (*storage).Put(append(testhelpers.KeySudoHash, testhelpers.KeyKeyHash...), bob.AsID.ToBytes())
	assert.NoError(t, err)

	call, err := ctypes.NewCall(metadata, "Sudo.set_key", bob)
	assert.NoError(t, err)

	extrinsic := ctypes.NewExtrinsic(call)

	o := ctypes.SignatureOptions{
		BlockHash:          ctypes.Hash(testhelpers.ParentHash),
		Era:                ctypes.ExtrinsicEra{IsImmortalEra: true},
		GenesisHash:        ctypes.Hash(testhelpers.ParentHash),
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

	res, err := rt.Exec("BlockBuilder_apply_extrinsic", extEnc.Bytes())
	assert.NoError(t, err)

	assert.Equal(t, testhelpers.ApplyExtrinsicResultSudoRequireSudoErr.Bytes(), res)
	assert.Equal(t, bob.AsID.ToBytes(), (*storage).Get(append(testhelpers.KeySudoHash, testhelpers.KeyKeyHash...)))
}
