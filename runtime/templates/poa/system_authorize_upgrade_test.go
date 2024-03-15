package main

import (
	"bytes"
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/frame/system"
	cscale "github.com/centrifuge/go-substrate-rpc-client/v4/scale"
	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	ctypes "github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/stretchr/testify/assert"
)

func Test_AuthorizeUpgrade_DispatchOutcome(t *testing.T) {
	rt, storage := newTestRuntime(t)
	metadata := runtimeMetadata(t, rt)

	runtimeVersion, err := rt.Version()
	assert.NoError(t, err)

	initializeBlock(t, rt, parentHash, stateRoot, extrinsicsRoot, blockNumber)

	call, err := ctypes.NewCall(metadata, "System.authorize_upgrade", codeHash)
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

	res, err := rt.Exec("BlockBuilder_apply_extrinsic", extEnc.Bytes())
	assert.NoError(t, err)

	upgradeAuthorizationBytes := (*storage).Get(append(keySystemHash, keyAuthorizedUpgradeHash...))
	upgradeAuthorization, err := system.DecodeCodeUpgradeAuthorization(bytes.NewBuffer(upgradeAuthorizationBytes))
	assert.NoError(t, err)

	assert.Equal(t, codeHash.ToBytes(), sc.FixedSequenceU8ToBytes(upgradeAuthorization.CodeHash.FixedSequence))
	assert.Equal(t, sc.Bool(true), upgradeAuthorization.CheckVersion)

	// Event are emitted
	buffer := &bytes.Buffer{}
	buffer.Write((*storage).Get(append(keySystemHash, keyEventsHash...)))

	decodedCount, err := sc.DecodeCompact[sc.U32](buffer)
	assert.NoError(t, err)
	assert.Equal(t, sc.U32(3), decodedCount.Number)

	assertEmittedSystemEvent(t, system.EventUpgradeAuthorized, buffer)

	assert.Equal(t, applyExtrinsicResultOutcome.Bytes(), res)
}
