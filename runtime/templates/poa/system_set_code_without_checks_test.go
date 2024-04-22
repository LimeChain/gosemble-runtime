package main

import (
	"bytes"
	"os"
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/frame/system"
	"github.com/LimeChain/gosemble/frame/transaction_payment"
	"github.com/LimeChain/gosemble/primitives/types"
	"github.com/LimeChain/gosemble/testhelpers"
	cscale "github.com/centrifuge/go-substrate-rpc-client/v4/scale"
	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	ctypes "github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/stretchr/testify/assert"
)

func Test_SetCodeWithoutChecks_DispatchOutcome(t *testing.T) {
	rt, storage := testhelpers.NewRuntimeInstance(t)
	metadata := testhelpers.RuntimeMetadata(t, rt)

	runtimeVersion, err := rt.Version()
	assert.NoError(t, err)

	testhelpers.InitializeBlock(t, rt, testhelpers.ParentHash, testhelpers.StateRoot, testhelpers.ExtrinsicsRoot, testhelpers.BlockNumber)

	codeSpecVersion101, err := os.ReadFile(testhelpers.RuntimeWasmSpecVersion101)
	assert.NoError(t, err)

	call, err := ctypes.NewCall(metadata, "System.set_code_without_checks", codeSpecVersion101)
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

	// Code is written to storage
	assert.Equal(t, codeSpecVersion101, (*storage).LoadCode())

	// Runtime environment upgraded digest item is logged
	testhelpers.AssertStorageDigestItem(t, storage, types.DigestItemRuntimeEnvironmentUpgraded)

	// Events are emitted
	buffer := &bytes.Buffer{}

	testhelpers.AssertStorageSystemEventCount(t, storage, uint32(3))

	buffer.Reset()
	buffer.Write((*storage).Get(append(testhelpers.KeySystemHash, testhelpers.KeyEventsHash...)))

	decodedCount, err := sc.DecodeCompact[sc.U32](buffer)
	assert.NoError(t, err)
	assert.Equal(t, uint32(decodedCount.Number.(sc.U32)), uint32(3))

	// Event system code updated
	testhelpers.AssertEmittedSystemEvent(t, system.EventCodeUpdated, buffer)

	// Event txpayment transaction fee paid
	testhelpers.AssertEmittedTransactionPaymentEvent(t, transaction_payment.EventTransactionFeePaid, buffer)

	// Event system extrinsic success
	testhelpers.AssertEmittedSystemEvent(t, system.EventExtrinsicSuccess, buffer)

	// Runtime version is updated
	rt, storage = testhelpers.NewRuntimeInstanceFromCode(t, rt, (*storage).LoadCode())

	runtimeVersion, err = rt.Version()
	assert.NoError(t, err)
	assert.Equal(t, runtimeVersion.SpecVersion, uint32(101))

	assert.Equal(t, testhelpers.ApplyExtrinsicResultOutcome.Bytes(), res)
}
