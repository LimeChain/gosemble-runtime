package main

import (
	"bytes"
	"os"
	"testing"

	gossamertypes "github.com/ChainSafe/gossamer/dot/types"
	"github.com/ChainSafe/gossamer/pkg/scale"
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/testhelpers"
	cscale "github.com/centrifuge/go-substrate-rpc-client/v4/scale"
	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	ctypes "github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/stretchr/testify/assert"
)

func Test_Block_Execution_After_Code_Upgrade(t *testing.T) {
	rt, storage := testhelpers.NewRuntimeInstance(t)
	metadata := testhelpers.RuntimeMetadata(t, rt)

	runtimeVersion, err := rt.Version()
	assert.NoError(t, err)

	time := dateTime.UnixMilli()

	slot := testhelpers.GetBabeSlot(t, rt, uint64(time))
	digest := testhelpers.NewBabeDigest(t, slot)

	header := gossamertypes.NewHeader(testhelpers.ParentHash, storageRoot, testhelpers.ExtrinsicsRoot, uint(testhelpers.BlockNumber), digest)

	encodedHeader, err := scale.Marshal(*header)
	assert.NoError(t, err)

	_, err = rt.Exec("Core_initialize_block", encodedHeader)
	assert.NoError(t, err)

	idata := gossamertypes.NewInherentData()
	err = idata.SetInherent(gossamertypes.Timstap0, uint64(dateTime.UnixMilli()))
	assert.NoError(t, err)

	ienc, err := idata.Encode()
	assert.NoError(t, err)

	inherentExt, err := rt.Exec("BlockBuilder_inherent_extrinsics", ienc)
	assert.NoError(t, err)
	assert.NotNil(t, inherentExt)

	buffer := bytes.NewBuffer([]byte{})
	buffer.Write([]byte{inherentExt[0]})

	totalInherents, err := sc.DecodeCompact[sc.U128](buffer)
	assert.Nil(t, err)
	assert.Equal(t, int64(1), totalInherents.ToBigInt().Int64())
	buffer.Reset()

	applyResult, err := rt.Exec("BlockBuilder_apply_extrinsic", inherentExt[1:])
	assert.NoError(t, err)

	assert.Equal(t, testhelpers.ApplyExtrinsicResultOutcome.Bytes(), applyResult)

	codeSpecVersion101, err := os.ReadFile(testhelpers.RuntimeWasmSpecVersion101)
	assert.NoError(t, err)

	// Set Sudo Key
	err = (*storage).Put(append(testhelpers.KeySudoHash, testhelpers.KeyKeyHash...), signature.TestKeyringPairAlice.PublicKey)
	assert.NoError(t, err)

	callArg, err := ctypes.NewCall(metadata, "System.set_code_without_checks", codeSpecVersion101)
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

	// Code is written to storage
	assert.Equal(t, codeSpecVersion101, (*storage).LoadCode())

	// Runtime version is updated
	rt, storage = testhelpers.NewRuntimeInstanceFromCode(t, rt, (*storage).LoadCode())

	runtimeVersion, err = rt.Version()
	assert.NoError(t, err)
	assert.Equal(t, runtimeVersion.SpecVersion, uint32(101))

	assert.Equal(t, testhelpers.ApplyExtrinsicResultOutcome.Bytes(), res)

	bytesResult, err := rt.Exec("BlockBuilder_finalize_block", []byte{})
	assert.NoError(t, err)

	resultHeader := gossamertypes.NewEmptyHeader()
	assert.NoError(t, scale.Unmarshal(bytesResult, resultHeader))
}
