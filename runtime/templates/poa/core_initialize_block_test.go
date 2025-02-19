package main

import (
	"testing"

	gossamertypes "github.com/ChainSafe/gossamer/dot/types"
	"github.com/ChainSafe/gossamer/lib/common"
	"github.com/ChainSafe/gossamer/pkg/scale"
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants"
	"github.com/LimeChain/gosemble/primitives/types"
	"github.com/LimeChain/gosemble/testhelpers"
	"github.com/stretchr/testify/assert"
)

func Test_CoreInitializeBlock(t *testing.T) {
	preRuntimeDigest := gossamertypes.PreRuntimeDigest{
		ConsensusEngineID: gossamertypes.BabeEngineID,
		// bytes for PreRuntimeDigest that was created in setupHeaderFile function
		Data: []byte{1, 60, 0, 0, 0, 150, 89, 189, 15, 0, 0, 0, 0, 112, 237, 173, 28, 144, 100, 255,
			247, 140, 177, 132, 53, 34, 61, 138, 218, 245, 234, 4, 194, 75, 26, 135, 102, 227, 220, 1, 235, 3, 204,
			106, 12, 17, 183, 151, 147, 212, 227, 28, 192, 153, 8, 56, 34, 156, 68, 254, 209, 102, 154, 124, 124,
			121, 225, 230, 208, 169, 99, 116, 214, 73, 103, 40, 6, 157, 30, 247, 57, 226, 144, 73, 122, 14, 59, 114,
			143, 168, 143, 203, 221, 58, 85, 4, 224, 239, 222, 2, 66, 231, 168, 6, 221, 79, 169, 38, 12},
	}

	expectedStorageDigest := gossamertypes.NewDigest()

	digest := gossamertypes.NewDigest()

	preRuntimeDigestItem := gossamertypes.NewDigestItem()
	assert.NoError(t, preRuntimeDigestItem.SetValue(preRuntimeDigest))

	sealDigestItem := gossamertypes.NewDigestItem()
	assert.NoError(t, sealDigestItem.SetValue(testhelpers.SealDigest))

	prdi, err := preRuntimeDigestItem.Value()
	assert.NoError(t, err)
	assert.NoError(t, digest.Add(prdi))

	sdi, err := sealDigestItem.Value()
	assert.NoError(t, err)
	assert.NoError(t, digest.Add(sdi))
	assert.NoError(t, expectedStorageDigest.Add(prdi))

	header := gossamertypes.NewHeader(testhelpers.ParentHash, testhelpers.StateRoot, testhelpers.ExtrinsicsRoot, uint(testhelpers.BlockNumber), digest)
	encodedHeader, err := scale.Marshal(*header)
	assert.NoError(t, err)

	rt, storage := testhelpers.NewRuntimeInstance(t)

	_, err = rt.Exec("Core_initialize_block", encodedHeader)
	assert.NoError(t, err)

	lrui := types.LastRuntimeUpgradeInfo{
		SpecVersion: sc.Compact{Number: sc.U32(constants.SpecVersion)},
		SpecName:    constants.SpecName,
	}
	assert.Equal(t, lrui.Bytes(), (*storage).Get(append(testhelpers.KeySystemHash, testhelpers.KeyLastRuntimeHash...)))

	encExtrinsicIndex0, _ := scale.Marshal(uint32(0))
	assert.Equal(t, encExtrinsicIndex0, (*storage).Get(testhelpers.KeyExtrinsicIndex))

	expectedExecutionPhase := types.NewExtrinsicPhaseApply(sc.U32(0))
	assert.Equal(t, expectedExecutionPhase.Bytes(), (*storage).Get(append(testhelpers.KeySystemHash, testhelpers.KeyExecutionPhaseHash...)))

	encBlockNumber, err := scale.Marshal(testhelpers.BlockNumber)
	assert.NoError(t, err)
	assert.Equal(t, encBlockNumber, (*storage).Get(append(testhelpers.KeySystemHash, testhelpers.KeyNumberHash...)))

	encExpectedDigest, err := scale.Marshal(expectedStorageDigest)
	assert.NoError(t, err)
	assert.Equal(t, encExpectedDigest, (*storage).Get(append(testhelpers.KeySystemHash, testhelpers.KeyDigestHash...)))
	assert.Equal(t, testhelpers.ParentHash.ToBytes(), (*storage).Get(append(testhelpers.KeySystemHash, testhelpers.KeyParentHash...)))

	blockHashKey := append(testhelpers.KeySystemHash, testhelpers.KeyBlockHash...)
	encPrevBlock, err := scale.Marshal(testhelpers.BlockNumber - 1)
	assert.NoError(t, err)
	numHash, err := common.Twox64(encPrevBlock)
	assert.NoError(t, err)

	blockHashKey = append(blockHashKey, numHash...)
	blockHashKey = append(blockHashKey, encPrevBlock...)
	assert.Equal(t, testhelpers.ParentHash.ToBytes(), (*storage).Get(blockHashKey))

	allConsumedWeight := types.ConsumedWeight{
		Operational: types.Weight{RefTime: 0, ProofSize: 0},
		Normal:      types.Weight{RefTime: 0, ProofSize: 0},
		// initial weight 0 + on initialize aura weight + base ext weight + extra weight
		Mandatory: types.Weight{RefTime: 4_313_948_000, ProofSize: 0},
	}
	assert.Equal(t, allConsumedWeight.Bytes(), (*storage).Get(append(testhelpers.KeySystemHash, testhelpers.KeyBlockWeightHash...)))
}
