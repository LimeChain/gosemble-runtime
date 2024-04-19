package main

import (
	"testing"
	"time"

	gossamertypes "github.com/ChainSafe/gossamer/dot/types"
	"github.com/ChainSafe/gossamer/pkg/scale"
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/testhelpers"
	"github.com/stretchr/testify/assert"
)

func Test_ApplyExtrinsic_Timestamp(t *testing.T) {
	rt, storage := testhelpers.NewRuntimeInstance(t)

	idata := gossamertypes.NewInherentData()
	time := time.Now().UnixMilli()

	slot := testhelpers.GetBabeSlot(t, rt, uint64(time))
	digest := testhelpers.NewBabeDigest(t, slot)

	header := gossamertypes.NewHeader(testhelpers.ParentHash, storageRoot, testhelpers.ExtrinsicsRoot, uint(testhelpers.BlockNumber), digest)

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
	assert.Equal(t, sc.U64(slot).Bytes(), (*storage).Get(append(testhelpers.KeyBabeHash, testhelpers.KeyCurrentSlotHash...)))
}
