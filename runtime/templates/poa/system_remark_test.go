package main

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/ChainSafe/gossamer/lib/common"
	"github.com/ChainSafe/gossamer/pkg/scale"
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/testhelpers"
	cscale "github.com/centrifuge/go-substrate-rpc-client/v4/scale"
	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	ctypes "github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/stretchr/testify/assert"
)

func Test_Remark_Signed_DispatchOutcome(t *testing.T) {
	rt, storage := testhelpers.NewRuntimeInstance(t)
	runtimeVersion, err := rt.Version()
	assert.NoError(t, err)

	metadata := testhelpers.RuntimeMetadata(t, rt)

	// Set Account Info
	balance, e := big.NewInt(0).SetString("500000000000000", 10)
	assert.True(t, e)
	testhelpers.SetStorageAccountInfo(t, storage, signature.TestKeyringPairAlice.PublicKey, balance, 0)

	testhelpers.InitializeBlock(t, rt, testhelpers.ParentHash, testhelpers.StateRoot, testhelpers.ExtrinsicsRoot, testhelpers.BlockNumber)

	call, err := ctypes.NewCall(metadata, "System.remark", []byte{})
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
	// Sign the transaction using Alice's default account
	err = extrinsic.Sign(signature.TestKeyringPairAlice, o)
	assert.NoError(t, err)

	extEnc := bytes.Buffer{}
	encoder := cscale.NewEncoder(&extEnc)
	err = extrinsic.Encode(*encoder)
	assert.NoError(t, err)

	res, err := rt.Exec("BlockBuilder_apply_extrinsic", extEnc.Bytes())
	assert.NoError(t, err)

	currentExtrinsicIndex := sc.U32(1)
	extrinsicIndexValue := (*storage).Get(testhelpers.KeyExtrinsicIndex)
	assert.Equal(t, currentExtrinsicIndex.Bytes(), extrinsicIndexValue)

	keyExtrinsicDataPrefixHash := append(testhelpers.KeySystemHash, testhelpers.KeyExtrinsicDataHash...)

	prevExtrinsic := currentExtrinsicIndex - 1
	hashIndex, err := common.Twox64(prevExtrinsic.Bytes())
	assert.NoError(t, err)

	keyExtrinsic := append(keyExtrinsicDataPrefixHash, hashIndex...)
	storageUxt := (*storage).Get(append(keyExtrinsic, prevExtrinsic.Bytes()...))

	expectedExtrinsicDataStorage, err := scale.Marshal(extEnc.Bytes())
	assert.NoError(t, err)

	assert.Equal(t, expectedExtrinsicDataStorage, storageUxt)
	assert.Equal(t, testhelpers.ApplyExtrinsicResultOutcome.Bytes(), res)
}

func Test_Remark_Unsigned_DispatchOutcome(t *testing.T) {
	rt, _ := testhelpers.NewRuntimeInstance(t)
	metadata := testhelpers.RuntimeMetadata(t, rt)

	call, err := ctypes.NewCall(metadata, "System.remark", []byte{})
	assert.NoError(t, err)

	extrinsic := ctypes.NewExtrinsic(call)

	extEnc := bytes.Buffer{}
	encoder := cscale.NewEncoder(&extEnc)
	err = extrinsic.Encode(*encoder)
	assert.NoError(t, err)

	res, err := rt.Exec("BlockBuilder_apply_extrinsic", extEnc.Bytes())
	assert.NoError(t, err)

	assert.Equal(t, testhelpers.ApplyExtrinsicResultOutcome.Bytes(), res)
}
