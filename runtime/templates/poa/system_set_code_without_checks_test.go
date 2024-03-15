package main

import (
	"bytes"
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/frame/system"
	"github.com/LimeChain/gosemble/frame/transaction_payment"
	"github.com/LimeChain/gosemble/primitives/types"
	cscale "github.com/centrifuge/go-substrate-rpc-client/v4/scale"
	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	ctypes "github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/stretchr/testify/assert"
)

func Test_SetCodeWithoutChecks_DispatchOutcome(t *testing.T) {
	rt, storage := newTestRuntime(t)
	metadata := runtimeMetadata(t, rt)

	runtimeVersion, err := rt.Version()
	assert.NoError(t, err)

	initializeBlock(t, rt, parentHash, stateRoot, extrinsicsRoot, blockNumber)

	call, err := ctypes.NewCall(metadata, "System.set_code_without_checks", codeSpecVersion101)
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

	// Code is written to storage
	assert.Equal(t, codeSpecVersion101, (*storage).LoadCode())

	// Runtime environment upgraded digest item is logged
	assertStorageDigestItem(t, storage, types.DigestItemRuntimeEnvironmentUpgraded)

	// Events are emitted
	buffer := &bytes.Buffer{}

	assertStorageSystemEventCount(t, storage, uint32(3))

	buffer.Reset()
	buffer.Write((*storage).Get(append(keySystemHash, keyEventsHash...)))

	decodedCount, err := sc.DecodeCompact[sc.U32](buffer)
	assert.NoError(t, err)
	assert.Equal(t, uint32(decodedCount.Number.(sc.U32)), uint32(3))

	// Event system code updated
	assertEmittedSystemEvent(t, system.EventCodeUpdated, buffer)

	// Event txpayment transaction fee paid
	assertEmittedTransactionPaymentEvent(t, transaction_payment.EventTransactionFeePaid, buffer)

	// Event system extrinsic success
	assertEmittedSystemEvent(t, system.EventExtrinsicSuccess, buffer)

	// Runtime version is updated
	rt, storage = newTestRuntimeFromCode(t, rt, (*storage).LoadCode())

	runtimeVersion, err = rt.Version()
	assert.NoError(t, err)
	assert.Equal(t, runtimeVersion.SpecVersion, uint32(101))

	assert.Equal(t, applyExtrinsicResultOutcome.Bytes(), res)
}
