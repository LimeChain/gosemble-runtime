package main

import (
	"bytes"
	"math/big"
	"testing"

	gossamertypes "github.com/ChainSafe/gossamer/dot/types"
	"github.com/ChainSafe/gossamer/pkg/scale"
	"github.com/LimeChain/gosemble/testhelpers"
	cscale "github.com/centrifuge/go-substrate-rpc-client/v4/scale"
	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	ctypes "github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/stretchr/testify/assert"
)

func Test_Balances_ForceTransfer_BadOrigin(t *testing.T) {
	rt, storage := testhelpers.NewRuntimeInstance(t)
	runtimeVersion, err := rt.Version()
	assert.NoError(t, err)

	metadata := testhelpers.RuntimeMetadata(t, rt)

	alice, err := ctypes.NewMultiAddressFromAccountID(signature.TestKeyringPairAlice.PublicKey)

	bob, err := ctypes.NewMultiAddressFromHexAccountID(
		"0x90b5ab205c6974c9ea841be688864633dc9ca8a357843eeacf2314649965fe22")
	assert.NoError(t, err)

	call, err := ctypes.NewCall(metadata, "Balances.force_transfer", alice, bob, ctypes.NewUCompactFromUInt(10000000000))
	assert.NoError(t, err)

	// Create the extrinsic
	ext := ctypes.NewExtrinsic(call)
	o := ctypes.SignatureOptions{
		BlockHash:          ctypes.Hash(testhelpers.ParentHash),
		Era:                ctypes.ExtrinsicEra{IsImmortalEra: true},
		GenesisHash:        ctypes.Hash(testhelpers.ParentHash),
		Nonce:              ctypes.NewUCompactFromUInt(0),
		SpecVersion:        ctypes.U32(runtimeVersion.SpecVersion),
		Tip:                ctypes.NewUCompactFromUInt(0),
		TransactionVersion: ctypes.U32(runtimeVersion.TransactionVersion),
	}

	// Set Account Info
	balance, ok := big.NewInt(0).SetString("500000000000000", 10)
	assert.True(t, ok)

	testhelpers.SetStorageAccountInfo(t, storage, signature.TestKeyringPairAlice.PublicKey, balance, 0)

	// Sign the transaction using Alice's default account
	err = ext.Sign(signature.TestKeyringPairAlice, o)
	assert.NoError(t, err)

	extEnc := bytes.Buffer{}
	encoder := cscale.NewEncoder(&extEnc)
	err = ext.Encode(*encoder)
	assert.NoError(t, err)

	header := gossamertypes.NewHeader(testhelpers.ParentHash, testhelpers.StateRoot, testhelpers.ExtrinsicsRoot, uint(testhelpers.BlockNumber), gossamertypes.NewDigest())
	encodedHeader, err := scale.Marshal(*header)
	assert.NoError(t, err)

	_, err = rt.Exec("Core_initialize_block", encodedHeader)
	assert.NoError(t, err)

	res, err := rt.Exec("BlockBuilder_apply_extrinsic", extEnc.Bytes())
	assert.NoError(t, err)

	assert.Equal(t, testhelpers.ApplyExtrinsicResultBadOriginErr.Bytes(), res)
}
