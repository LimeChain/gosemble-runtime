package main

import (
	"bytes"
	"math/big"
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/frame/balances"
	"github.com/LimeChain/gosemble/frame/system"
	"github.com/LimeChain/gosemble/frame/transaction_payment"
	cscale "github.com/centrifuge/go-substrate-rpc-client/v4/scale"
	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	ctypes "github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/stretchr/testify/assert"
)

func Test_Remark_With_Event_Signed_DispatchOutcome(t *testing.T) {
	rt, storage := newTestRuntime(t)
	metadata := runtimeMetadata(t, rt)

	runtimeVersion, err := rt.Version()
	assert.NoError(t, err)

	// Set account info
	balance, e := big.NewInt(0).SetString("500000000000000", 10)
	assert.True(t, e)
	setStorageAccountInfo(t, storage, signature.TestKeyringPairAlice.PublicKey, balance, 0)

	initializeBlock(t, rt, parentHash, stateRoot, extrinsicsRoot, blockNumber)

	remarkMsg := sc.BytesToFixedSequenceU8([]byte("ngmi"))
	call, err := ctypes.NewCall(metadata, "System.remark_with_event", remarkMsg.Bytes())
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

	// Events are emitted
	buffer := &bytes.Buffer{}

	buffer.Write((*storage).Get(append(keySystemHash, keyEventCountHash...)))
	storageEventCount, err := sc.DecodeU32(buffer)
	assert.NoError(t, err)
	assert.Equal(t, sc.U32(4), storageEventCount)

	buffer.Reset()
	buffer.Write((*storage).Get(append(keySystemHash, keyEventsHash...)))

	decodedCount, err := sc.DecodeCompact[sc.U32](buffer)
	assert.NoError(t, err)
	assert.Equal(t, decodedCount.Number, storageEventCount)

	// balances withdraw event
	assertEmittedBalancesEvent(t, balances.EventWithdraw, buffer)

	// system remarked event
	assertEmittedSystemEvent(t, system.EventRemarked, buffer)

	// txpayment transaction fee paid event
	assertEmittedTransactionPaymentEvent(t, transaction_payment.EventTransactionFeePaid, buffer)

	// system extrinsic success event
	assertEmittedSystemEvent(t, system.EventExtrinsicSuccess, buffer)

	assert.Equal(t, applyExtrinsicResultOutcome.Bytes(), res)
}

func Test_Remark_With_Event_Unsigned_DispatchOutcome(t *testing.T) {
	rt, _ := newTestRuntime(t)
	metadata := runtimeMetadata(t, rt)

	call, err := ctypes.NewCall(metadata, "System.remark_with_event", []byte{})
	assert.NoError(t, err)

	extrinsic := ctypes.NewExtrinsic(call)

	extEnc := bytes.Buffer{}
	encoder := cscale.NewEncoder(&extEnc)
	err = extrinsic.Encode(*encoder)
	assert.NoError(t, err)

	res, err := rt.Exec("BlockBuilder_apply_extrinsic", extEnc.Bytes())
	assert.NoError(t, err)

	assert.Equal(t, applyExtrinsicResultBadOriginErr.Bytes(), res)
}
