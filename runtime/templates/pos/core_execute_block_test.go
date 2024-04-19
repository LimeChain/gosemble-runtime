package main

import (
	"time"

	"bytes"
	"testing"

	"github.com/ChainSafe/gossamer/lib/common"

	gossamertypes "github.com/ChainSafe/gossamer/dot/types"
	"github.com/ChainSafe/gossamer/pkg/scale"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/LimeChain/gosemble/testhelpers"

	"github.com/stretchr/testify/assert"
)

var (
	authority1  = gossamertypes.AuthorityRaw{Key: [32]byte{0xd4, 0x35, 0x93, 0xc7, 0x15, 0xfd, 0xd3, 0x1c, 0x61, 0x14, 0x1a, 0xbd, 0x04, 0xa9, 0x9f, 0xd6, 0x82, 0x2c, 0x85, 0x58, 0x85, 0x4c, 0xcd, 0xe3, 0x9a, 0x56, 0x84, 0xe7, 0xa5, 0x6d, 0xa2, 0x7d}, Weight: 0}
	dateTime    = time.Date(2023, time.January, 2, 3, 4, 5, 6, time.UTC)
	storageRoot = common.MustHexToHash("0x48048c3e427fa73c56be0f8859aa0471c836c9fb874f3277f1b19e68a59539c1") // Depends on date
)

func Test_BlockExecution(t *testing.T) {
	// core.InitializeBlock
	// blockBuilder.InherentExtrinsics
	// blockBuilder.ApplyExtrinsics
	// blockBuilder.FinalizeBlock

	rt, storage := testhelpers.NewRuntimeInstance(t)
	metadata := testhelpers.RuntimeMetadata(t, rt)

	testhelpers.GenesisBuild(t, rt, testhelpers.GenesisConfigJson)

	time := dateTime.UnixMilli()

	slot := testhelpers.GetBabeSlot(t, rt, uint64(time))
	digest := testhelpers.NewBabeDigest(t, slot)

	header := gossamertypes.NewHeader(testhelpers.ParentHash, storageRoot, testhelpers.ExtrinsicsRoot, uint(testhelpers.BlockNumber), digest)

	encodedHeader, err := scale.Marshal(*header)
	assert.NoError(t, err)

	_, err = rt.Exec("Core_initialize_block", encodedHeader)
	assert.NoError(t, err)

	lrui := primitives.LastRuntimeUpgradeInfo{
		SpecVersion: sc.Compact{Number: sc.U32(constants.SpecVersion)},
		SpecName:    constants.SpecName,
	}
	assert.Equal(t, lrui.Bytes(), (*storage).Get(append(testhelpers.KeySystemHash, testhelpers.KeyLastRuntimeHash...)))

	encExtrinsicIndex0, _ := scale.Marshal(uint32(0))
	assert.Equal(t, encExtrinsicIndex0, (*storage).Get(testhelpers.KeyExtrinsicIndex))

	expectedExecutionPhase := primitives.NewExtrinsicPhaseApply(sc.U32(0))
	assert.Equal(t, expectedExecutionPhase.Bytes(), (*storage).Get(append(testhelpers.KeySystemHash, testhelpers.KeyExecutionPhaseHash...)))

	encBlockNumber, err := scale.Marshal(testhelpers.BlockNumber)
	assert.NoError(t, err)
	assert.Equal(t, encBlockNumber, (*storage).Get(append(testhelpers.KeySystemHash, testhelpers.KeyNumberHash...)))

	assert.Equal(t, testhelpers.ParentHash.ToBytes(), (*storage).Get(append(testhelpers.KeySystemHash, testhelpers.KeyParentHash...)))

	blockHashKey := append(testhelpers.KeySystemHash, testhelpers.KeyBlockHash...)
	encPrevBlock, err := scale.Marshal(testhelpers.BlockNumber - 1)
	assert.NoError(t, err)
	numHash, err := common.Twox64(encPrevBlock)
	assert.NoError(t, err)

	blockHashKey = append(blockHashKey, numHash...)
	blockHashKey = append(blockHashKey, encPrevBlock...)
	assert.Equal(t, testhelpers.ParentHash.ToBytes(), (*storage).Get(blockHashKey))

	idata := gossamertypes.NewInherentData()
	err = idata.SetInherent(gossamertypes.Timstap0, uint64(time))
	assert.NoError(t, err)

	expectedExtrinsicBytes := testhelpers.TimestampExtrinsicBytes(t, metadata, uint64(time))

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

	actualExtrinsic := inherentExt[1:]
	assert.Equal(t, expectedExtrinsicBytes, actualExtrinsic)

	applyResult, err := rt.Exec("BlockBuilder_apply_extrinsic", inherentExt[1:])
	assert.NoError(t, err)

	assert.Equal(t,
		testhelpers.ApplyExtrinsicResultOutcome.Bytes(),
		applyResult,
	)

	bytesResult, err := rt.Exec("BlockBuilder_finalize_block", []byte{})
	assert.NoError(t, err)

	expectedDigest := setupExpectedDigest(t, slot)
	resultHeader := gossamertypes.NewEmptyHeader()
	assert.NoError(t, scale.Unmarshal(bytesResult, resultHeader))
	resultHeader.Hash() // Call this to be set, otherwise structs do not match...
	expectedHeader := gossamertypes.NewHeader(testhelpers.ParentHash, storageRoot, testhelpers.ExtrinsicsRoot, uint(testhelpers.BlockNumber), expectedDigest)
	assert.Equal(t, expectedHeader, resultHeader)

	assert.Equal(t, []byte(nil), (*storage).Get(append(testhelpers.KeyTimestampHash, testhelpers.KeyTimestampDidUpdateHash...)))
	assert.Equal(t, sc.U64(time).Bytes(), (*storage).Get(append(testhelpers.KeyTimestampHash, testhelpers.KeyTimestampNowHash...)))

	assert.Equal(t, []byte(nil), (*storage).Get(testhelpers.KeyExtrinsicIndex))
	assert.Equal(t, []byte(nil), (*storage).Get(append(testhelpers.KeySystemHash, testhelpers.KeyExecutionPhaseHash...)))
	assert.Equal(t, []byte(nil), (*storage).Get(append(testhelpers.KeySystemHash, testhelpers.KeyAllExtrinsicsLenHash...)))
	assert.Equal(t, []byte(nil), (*storage).Get(append(testhelpers.KeySystemHash, testhelpers.KeyExtrinsicCountHash...)))

	expectedEncDigest, err := scale.Marshal(expectedDigest)
	assert.NoError(t, err)

	assert.Equal(t, testhelpers.ParentHash.ToBytes(), (*storage).Get(append(testhelpers.KeySystemHash, testhelpers.KeyParentHash...)))
	assert.Equal(t, expectedEncDigest, (*storage).Get(append(testhelpers.KeySystemHash, testhelpers.KeyDigestHash...)))
	assert.Equal(t, encBlockNumber, (*storage).Get(append(testhelpers.KeySystemHash, testhelpers.KeyNumberHash...)))

	assert.Equal(t, sc.U64(slot).Bytes(), (*storage).Get(append(testhelpers.KeyBabeHash, testhelpers.KeyCurrentSlotHash...)))
}

func setupExpectedDigest(t *testing.T, slot uint64) gossamertypes.Digest {
	nextEpoch := gossamertypes.NextEpochData{
		Authorities: []gossamertypes.AuthorityRaw{authority1},
		Randomness:  [32]byte{},
	}
	babeConsensusDigest := gossamertypes.NewBabeConsensusDigest()
	babeConsensusDigest.SetValue(nextEpoch)
	encConsensusDigest, err := scale.Marshal(babeConsensusDigest)
	assert.NoError(t, err)
	nextEpochConsensusDigest := gossamertypes.ConsensusDigest{
		ConsensusEngineID: gossamertypes.BabeEngineID,
		Data:              encConsensusDigest,
	}

	digest := testhelpers.NewBabeDigest(t, slot)
	assert.NoError(t, err)
	digest.Add(nextEpochConsensusDigest)

	return digest
}
