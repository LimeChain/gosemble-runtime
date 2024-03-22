package main

import (
	"bytes"
	"testing"
	"time"

	gossamertypes "github.com/ChainSafe/gossamer/dot/types"
	"github.com/ChainSafe/gossamer/lib/common"
	"github.com/ChainSafe/gossamer/pkg/scale"
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/frame/babe"
	"github.com/LimeChain/gosemble/primitives/types"
	"github.com/LimeChain/gosemble/testhelpers"
	"github.com/stretchr/testify/assert"
)

func Test_Offchain_Worker(t *testing.T) {
	rt, storage := testhelpers.NewRuntimeInstance(t)

	time := time.Date(2023, time.January, 2, 3, 4, 5, 6, time.UTC)

	babeConfigurationBytes, err := rt.Exec("BabeApi_configuration", []byte{})
	assert.NoError(t, err)

	buffer := bytes.NewBuffer(babeConfigurationBytes)

	babeConfiguration, err := babe.DecodeBabeConfiguration(buffer)
	assert.NoError(t, err)

	slot := sc.U64(time.UnixMilli()) / babeConfiguration.SlotDuration

	// preRuntimeDigest := gossamertypes.PreRuntimeDigest{
	// 	ConsensusEngineID: aura.EngineId,
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

	expectedStorageDigest, err := scale.Marshal(digest)
	assert.NoError(t, err)

	encBlockNumber, err := scale.Marshal(testhelpers.BlockNumber)
	assert.NoError(t, err)

	encodedHeader, err := scale.Marshal(*header)
	assert.NoError(t, err)

	blockHashKey := append(testhelpers.KeySystemHash, testhelpers.KeyBlockHash...)
	encPrevBlock, err := scale.Marshal(testhelpers.BlockNumber - 1)
	assert.NoError(t, err)
	prevBlockNumHash, err := common.Twox64(encPrevBlock)
	assert.NoError(t, err)

	prevBlockHashKey := append(blockHashKey, prevBlockNumHash...)
	prevBlockHashKey = append(prevBlockHashKey, encPrevBlock...)

	expectedBlockHash, err := common.Blake2bHash(encodedHeader)
	assert.NoError(t, err)
	blockNumHash, err := common.Twox64(encBlockNumber)
	assert.NoError(t, err)
	blockHashKey = append(blockHashKey, blockNumHash...)
	blockHashKey = append(blockHashKey, encBlockNumber...)

	_, err = rt.Exec("OffchainWorkerApi_offchain_worker", encodedHeader)
	assert.NoError(t, err)

	assert.Equal(t, types.PhaseInitialization.Bytes(), (*storage).Get(append(testhelpers.KeySystemHash, testhelpers.KeyExecutionPhaseHash...)))
	assert.Equal(t, sc.U32(0).Bytes(), (*storage).Get(testhelpers.KeyExtrinsicIndex))
	assert.Equal(t, encBlockNumber, (*storage).Get(append(testhelpers.KeySystemHash, testhelpers.KeyNumberHash...)))
	assert.Equal(t, expectedStorageDigest, (*storage).Get(append(testhelpers.KeySystemHash, testhelpers.KeyDigestHash...)))
	assert.Equal(t, testhelpers.ParentHash.ToBytes(), (*storage).Get(append(testhelpers.KeySystemHash, testhelpers.KeyParentHash...)))
	assert.Equal(t, testhelpers.ParentHash.ToBytes(), (*storage).Get(prevBlockHashKey))
	assert.Equal(t, []byte(nil), (*storage).Get(append(testhelpers.KeySystemHash, testhelpers.KeyBlockWeightHash...)))

	assert.Equal(t, expectedBlockHash.ToBytes(), (*storage).Get(blockHashKey))
}
