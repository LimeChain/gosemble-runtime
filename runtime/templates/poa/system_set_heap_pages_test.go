package main

import (
	"bytes"
	"testing"

	gossamertypes "github.com/ChainSafe/gossamer/dot/types"
	"github.com/ChainSafe/gossamer/pkg/scale"
	"github.com/LimeChain/gosemble/testhelpers"
	cscale "github.com/centrifuge/go-substrate-rpc-client/v4/scale"
	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	ctypes "github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/stretchr/testify/assert"
)

var (
	pages = uint64(5)
)

var (
	expectedHeapPagesStorage = uint64(0)
)

func Test_SetHeapPages_DispatchOutcome(t *testing.T) {
	rt, storage := testhelpers.NewRuntimeInstance(t)
	metadata := testhelpers.RuntimeMetadata(t, rt)

	runtimeVersion, err := rt.Version()
	assert.NoError(t, err)

	// Set Sudo Key
	err = (*storage).Put(append(testhelpers.KeySudoHash, testhelpers.KeyKeyHash...), signature.TestKeyringPairAlice.PublicKey)
	assert.NoError(t, err)

	testhelpers.InitializeBlock(t, rt, testhelpers.ParentHash, testhelpers.StateRoot, testhelpers.ExtrinsicsRoot, testhelpers.BlockNumber)

	callArg, err := ctypes.NewCall(metadata, "System.set_heap_pages", pages)
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

	heapPagesStorage := uint64(0)
	scale.Unmarshal((*storage).Get(testhelpers.KeyHeapPages), &heapPagesStorage)
	assert.Equal(t, expectedHeapPagesStorage, heapPagesStorage)

	digestStorage := gossamertypes.NewDigest()
	scale.Unmarshal((*storage).Get(append(testhelpers.KeySystemHash, testhelpers.KeyDigestHash...)[:]), &digestStorage)
	assert.Equal(t, gossamertypes.Digest(nil), digestStorage)

	res, err := rt.Exec("BlockBuilder_apply_extrinsic", extEnc.Bytes())
	assert.NoError(t, err)

	heapPagesStorage = uint64(0)
	scale.Unmarshal((*storage).Get(testhelpers.KeyHeapPages), &heapPagesStorage)
	expectedHeapPagesStorage = pages
	assert.Equal(t, expectedHeapPagesStorage, heapPagesStorage)

	digestStorage = gossamertypes.NewDigest()
	scale.Unmarshal((*storage).Get(append(testhelpers.KeySystemHash, testhelpers.KeyDigestHash...)[:]), &digestStorage)
	expectedDigestStorage := gossamertypes.Digest(nil)
	expectedDigestStorage.Add(gossamertypes.RuntimeEnvironmentUpdated{})
	assert.Equal(t, expectedDigestStorage, digestStorage)

	assert.Equal(t, testhelpers.ApplyExtrinsicResultOutcome.Bytes(), res)
}
