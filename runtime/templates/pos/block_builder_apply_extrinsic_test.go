package main

import (
	"bytes"
	"math/big"
	"testing"
	"time"

	gossamertypes "github.com/ChainSafe/gossamer/dot/types"
	"github.com/ChainSafe/gossamer/pkg/scale"
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants"
	"github.com/LimeChain/gosemble/frame/babe"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/LimeChain/gosemble/testhelpers"
	cscale "github.com/centrifuge/go-substrate-rpc-client/v4/scale"
	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	ctypes "github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/stretchr/testify/assert"
)

func Test_ApplyExtrinsic_Timestamp(t *testing.T) {
	rt, storage := testhelpers.NewRuntimeInstance(t)

	idata := gossamertypes.NewInherentData()
	time := time.Now().UnixMilli()

	babeConfigurationBytes, err := rt.Exec("BabeApi_configuration", []byte{})
	assert.NoError(t, err)

	buffer := bytes.NewBuffer(babeConfigurationBytes)

	babeConfiguration, err := babe.DecodeBabeConfiguration(buffer)
	assert.NoError(t, err)

	slot := sc.U64(time) / babeConfiguration.SlotDuration

	// preRuntimeDigest := gossamertypes.PreRuntimeDigest{
	// 	ConsensusEngineID: babe.EngineId,
	// 	Data:              slot.Bytes(),
	// }
	// assert.NoError(t, digest.Add(preRuntimeDigest))

	babeHeader := gossamertypes.NewBabeDigest()
	err = babeHeader.SetValue(*gossamertypes.NewBabePrimaryPreDigest(0, uint64(slot), [32]byte{}, [64]byte{}))
	assert.NoError(t, err)
	data, err := scale.Marshal(babeHeader)
	assert.NoError(t, err)
	preDigest := gossamertypes.NewBABEPreRuntimeDigest(data)

	digest := gossamertypes.NewDigest()
	err = digest.Add(*preDigest)
	assert.NoError(t, err)

	header := gossamertypes.NewHeader(testhelpers.ParentHash, testhelpers.StateRoot, testhelpers.ExtrinsicsRoot, uint(testhelpers.BlockNumber), digest)
	encodedHeader, err := scale.Marshal(*header)
	assert.NoError(t, err)

	_, err = rt.Exec("Core_initialize_block", encodedHeader)
	assert.NoError(t, err)

	err = idata.SetInherent(gossamertypes.Timstap0, uint64(time))
	assert.NoError(t, err)

	ienc, err := idata.Encode()
	assert.NoError(t, err)
	inherentExt, err := rt.Exec("BlockBuilder_inherent_extrinsics", ienc)
	assert.NoError(t, err)

	applyResult, err := rt.Exec("BlockBuilder_apply_extrinsic", inherentExt[1:])
	assert.NoError(t, err)

	assert.Equal(t,
		testhelpers.ApplyExtrinsicResultOutcome.Bytes(),
		applyResult,
	)

	assert.Equal(t, []byte{1}, (*storage).Get(append(testhelpers.KeyTimestampHash, testhelpers.KeyTimestampDidUpdateHash...)))
	assert.Equal(t, sc.U64(time).Bytes(), (*storage).Get(append(testhelpers.KeyTimestampHash, testhelpers.KeyTimestampNowHash...)))

	assert.Equal(t, slot.Bytes(), (*storage).Get(append(testhelpers.KeyBabeHash, testhelpers.KeyCurrentSlotHash...)))
}

func Test_ApplyExtrinsic_DispatchError_BadProofError(t *testing.T) {
	rt, storage := testhelpers.NewRuntimeInstance(t)
	runtimeVersion, err := rt.Version()
	assert.NoError(t, err)

	metadata := testhelpers.RuntimeMetadata(t, rt)

	digest := gossamertypes.NewDigest()

	header := gossamertypes.NewHeader(testhelpers.ParentHash, testhelpers.StateRoot, testhelpers.ExtrinsicsRoot, uint(testhelpers.BlockNumber), digest)
	encodedHeader, err := scale.Marshal(*header)
	assert.NoError(t, err)

	_, err = rt.Exec("Core_initialize_block", encodedHeader)
	assert.NoError(t, err)

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

	// Switch nonce
	extrinsic.Signature.Nonce = ctypes.NewUCompactFromUInt(1)

	extEnc := bytes.Buffer{}
	encoder := cscale.NewEncoder(&extEnc)
	err = extrinsic.Encode(*encoder)
	assert.NoError(t, err)

	res, err := rt.Exec("BlockBuilder_apply_extrinsic", extEnc.Bytes())
	assert.NoError(t, err)

	extrinsicIndex := sc.U32(0)
	extrinsicIndexValue := (*storage).Get(append(testhelpers.KeySystemHash, sc.NewOption[sc.U32](extrinsicIndex).Bytes()...))
	assert.Equal(t, []byte(nil), extrinsicIndexValue)

	assert.Equal(t, testhelpers.ApplyExtrinsicResultBadProofErr.Bytes(), res)
}

func Test_ApplyExtrinsic_ExhaustsResourcesError(t *testing.T) {
	rt, storage := testhelpers.NewRuntimeInstance(t)
	runtimeVersion, err := rt.Version()
	assert.NoError(t, err)

	metadata := testhelpers.RuntimeMetadata(t, rt)

	digest := gossamertypes.NewDigest()

	header := gossamertypes.NewHeader(testhelpers.ParentHash, testhelpers.StateRoot, testhelpers.ExtrinsicsRoot, uint(testhelpers.BlockNumber), digest)
	encodedHeader, err := scale.Marshal(*header)
	assert.NoError(t, err)

	_, err = rt.Exec("Core_initialize_block", encodedHeader)
	assert.NoError(t, err)

	// Append long args
	args := make([]byte, constants.FiveMbPerBlockPerExtrinsic)

	call, err := ctypes.NewCall(metadata, "System.remark", args)
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

	extrinsicIndex := sc.U32(0)
	extrinsicIndexValue := (*storage).Get(append(testhelpers.KeySystemHash, sc.NewOption[sc.U32](extrinsicIndex).Bytes()...))
	assert.Equal(t, []byte(nil), extrinsicIndexValue)

	assert.Equal(t, testhelpers.ApplyExtrinsicResultExhaustsResourcesErr.Bytes(), res)
}

func Test_ApplyExtrinsic_FutureError_InvalidNonce(t *testing.T) {
	rt, storage := testhelpers.NewRuntimeInstance(t)
	runtimeVersion, err := rt.Version()
	assert.NoError(t, err)

	metadata := testhelpers.RuntimeMetadata(t, rt)

	// Set Balance & Nonce
	testhelpers.SetStorageAccountInfo(t, storage, signature.TestKeyringPairAlice.PublicKey, big.NewInt(5), 3)

	digest := gossamertypes.NewDigest()

	header := gossamertypes.NewHeader(testhelpers.ParentHash, testhelpers.StateRoot, testhelpers.ExtrinsicsRoot, uint(testhelpers.BlockNumber), digest)
	encodedHeader, err := scale.Marshal(*header)
	assert.NoError(t, err)

	_, err = rt.Exec("Core_initialize_block", encodedHeader)
	assert.NoError(t, err)

	call, err := ctypes.NewCall(metadata, "System.remark", []byte{})
	assert.NoError(t, err)

	extrinsic := ctypes.NewExtrinsic(call)
	o := ctypes.SignatureOptions{
		BlockHash:          ctypes.Hash(testhelpers.ParentHash),
		Era:                ctypes.ExtrinsicEra{IsImmortalEra: true},
		GenesisHash:        ctypes.Hash(testhelpers.ParentHash),
		Nonce:              ctypes.NewUCompactFromUInt(5), // Invalid nonce
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

	encTransactionValidityResult, err := rt.Exec("BlockBuilder_apply_extrinsic", extEnc.Bytes())
	assert.NoError(t, err)

	buffer := &bytes.Buffer{}
	buffer.Write(encTransactionValidityResult)
	transactionValidityResult, err := primitives.DecodeTransactionValidityResult(buffer)
	assert.Nil(t, err)

	assert.Equal(t, testhelpers.TransactionValidityResultFutureErr, transactionValidityResult)
}

func Test_ApplyExtrinsic_InvalidLengthPrefix(t *testing.T) {
	rt, _ := testhelpers.NewRuntimeInstance(t)
	runtimeVersion, err := rt.Version()
	assert.NoError(t, err)

	metadata := testhelpers.RuntimeMetadata(t, rt)

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

	// Increase extrinsic length by 1
	bytesExtrinsic := extEnc.Bytes()
	bytesExtrinsic[0] += 4

	_, err = rt.Exec("BlockBuilder_apply_extrinsic", bytesExtrinsic)
	assert.Error(t, err)
}
