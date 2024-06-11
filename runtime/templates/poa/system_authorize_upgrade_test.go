package main

import (
	"bytes"
	"os"
	"testing"

	"github.com/ChainSafe/gossamer/lib/common"
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/frame/system"
	"github.com/LimeChain/gosemble/testhelpers"
	cscale "github.com/centrifuge/go-substrate-rpc-client/v4/scale"
	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	ctypes "github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/stretchr/testify/assert"
)

func Test_AuthorizeUpgrade_DispatchOutcome(t *testing.T) {
	rt, storage := testhelpers.NewRuntimeInstance(t)
	metadata := testhelpers.RuntimeMetadata(t, rt)

	runtimeVersion, err := rt.Version()
	assert.NoError(t, err)

	// Set Sudo Key
	err = (*storage).Put(append(testhelpers.KeySudoHash, testhelpers.KeyKeyHash...), signature.TestKeyringPairAlice.PublicKey)
	assert.NoError(t, err)

	testhelpers.InitializeBlock(t, rt, testhelpers.ParentHash, testhelpers.StateRoot, testhelpers.ExtrinsicsRoot, testhelpers.BlockNumber)

	codeSpecVersion101, err := os.ReadFile(testhelpers.RuntimeWasmSpecVersion101)
	assert.NoError(t, err)
	codeHash := common.MustBlake2bHash(codeSpecVersion101)

	callArg, err := ctypes.NewCall(metadata, "System.authorize_upgrade", codeHash)
	assert.NoError(t, err)

	call, err := ctypes.NewCall(metadata, "Sudo.sudo", callArg)
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

	upgradeAuthorizationBytes := (*storage).Get(append(testhelpers.KeySystemHash, testhelpers.KeyAuthorizedUpgradeHash...))
	upgradeAuthorization, err := system.DecodeCodeUpgradeAuthorization(bytes.NewBuffer(upgradeAuthorizationBytes))
	assert.NoError(t, err)

	assert.Equal(t, codeHash.ToBytes(), sc.FixedSequenceU8ToBytes(upgradeAuthorization.CodeHash.FixedSequence))
	assert.Equal(t, sc.Bool(true), upgradeAuthorization.CheckVersion)

	// Event are emitted
	buffer := &bytes.Buffer{}
	buffer.Write((*storage).Get(append(testhelpers.KeySystemHash, testhelpers.KeyEventsHash...)))

	decodedCount, err := sc.DecodeCompact[sc.U32](buffer)
	assert.NoError(t, err)
	assert.Equal(t, sc.U32(4), decodedCount.Number)

	testhelpers.AssertEmittedSystemEvent(t, system.EventUpgradeAuthorized, buffer)

	assert.Equal(t, testhelpers.ApplyExtrinsicResultOutcome.Bytes(), res)
}
