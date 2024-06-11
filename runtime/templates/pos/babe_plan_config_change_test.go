package main

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/ChainSafe/gossamer/pkg/scale"
	"github.com/LimeChain/gosemble/testhelpers"
	cscale "github.com/centrifuge/go-substrate-rpc-client/v4/scale"
	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	ctypes "github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/stretchr/testify/assert"
)

type NextConfigDataV1 struct {
	V1             byte
	C1             uint64
	C2             uint64
	SecondarySlots byte
}

func Test_Babe_Plan_Config_Change(t *testing.T) {
	rt, storage := testhelpers.NewRuntimeInstance(t)

	metadata := testhelpers.RuntimeMetadata(t, rt)

	// Set Account Info
	balance, e := big.NewInt(0).SetString("500000000000000", 10)
	assert.True(t, e)
	testhelpers.SetStorageAccountInfo(t, storage, signature.TestKeyringPairAlice.PublicKey, balance, 0)

	testhelpers.InitializeBlock(t, rt, testhelpers.ParentHash, testhelpers.StateRoot, testhelpers.ExtrinsicsRoot, testhelpers.BlockNumber)

	// Gossamer 'scale' codec correctly encodes the NextConfigDataV1 (01 0300000000000000 0500000000000000 02)
	// but 'go-substrate-rpc-client' does not (which is used to set up the call extrinsic below)
	// nextConfigData := types.NextConfigDataV1{
	// 	C1:             3,
	// 	C2:             5,
	// 	SecondarySlots: 2,
	// }
	// versionedNextConfigData := types.NewVersionedNextConfigData()
	// versionedNextConfigData.SetValue(nextConfigData)
	versionedNextConfigData := NextConfigDataV1{
		V1:             1,
		C1:             3,
		C2:             5,
		SecondarySlots: 2,
	}

	// Set Sudo Key
	err := (*storage).Put(append(testhelpers.KeySudoHash, testhelpers.KeyKeyHash...), signature.TestKeyringPairAlice.PublicKey)
	assert.NoError(t, err)

	callArg, err := ctypes.NewCall(metadata, "Babe.plan_config_change", versionedNextConfigData)

	call, err := ctypes.NewCall(metadata, "Sudo.sudo", callArg)
	assert.NoError(t, err)

	extrinsic := ctypes.NewExtrinsic(call)

	runtimeVersion, err := rt.Version()
	assert.NoError(t, err)

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

	configData := &NextConfigDataV1{}
	scale.Unmarshal((*storage).Get(append(testhelpers.KeyBabeHash, testhelpers.KeyPendingEpochConfigChangeHash...)), configData)
	assert.Equal(t, versionedNextConfigData, *configData)

	assert.Equal(t, testhelpers.ApplyExtrinsicResultOutcome.Bytes(), res)
}
