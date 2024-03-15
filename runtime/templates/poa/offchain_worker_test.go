package main

import (
	"bytes"
	"testing"
	"time"

	gossamertypes "github.com/ChainSafe/gossamer/dot/types"
	"github.com/ChainSafe/gossamer/lib/common"
	"github.com/ChainSafe/gossamer/pkg/scale"
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/frame/aura"
	"github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
)

func Test_Offchain_Worker(t *testing.T) {
	rt, storage := newTestRuntime(t)

	time := time.Date(2023, time.January, 2, 3, 4, 5, 6, time.UTC)

	digest := gossamertypes.NewDigest()

	bytesSlotDuration, err := rt.Exec("AuraApi_slot_duration", []byte{})
	assert.NoError(t, err)

	buffer := &bytes.Buffer{}
	buffer.Write(bytesSlotDuration)

	slotDuration, err := sc.DecodeU64(buffer)
	assert.Nil(t, err)
	buffer.Reset()

	slot := sc.U64(time.UnixMilli()) / slotDuration

	preRuntimeDigest := gossamertypes.PreRuntimeDigest{
		ConsensusEngineID: aura.EngineId,
		Data:              slot.Bytes(),
	}
	assert.NoError(t, digest.Add(preRuntimeDigest))

	header := gossamertypes.NewHeader(parentHash, stateRoot, extrinsicsRoot, uint(blockNumber), digest)

	expectedStorageDigest, err := scale.Marshal(digest)
	assert.NoError(t, err)

	encBlockNumber, err := scale.Marshal(blockNumber)
	assert.NoError(t, err)

	encodedHeader, err := scale.Marshal(*header)
	assert.NoError(t, err)

	blockHashKey := append(keySystemHash, keyBlockHash...)
	encPrevBlock, err := scale.Marshal(blockNumber - 1)
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

	assert.Equal(t, types.PhaseInitialization.Bytes(), (*storage).Get(append(keySystemHash, keyExecutionPhaseHash...)))
	assert.Equal(t, sc.U32(0).Bytes(), (*storage).Get(keyExtrinsicIndex))
	assert.Equal(t, encBlockNumber, (*storage).Get(append(keySystemHash, keyNumberHash...)))
	assert.Equal(t, expectedStorageDigest, (*storage).Get(append(keySystemHash, keyDigestHash...)))
	assert.Equal(t, parentHash.ToBytes(), (*storage).Get(append(keySystemHash, keyParentHash...)))
	assert.Equal(t, parentHash.ToBytes(), (*storage).Get(prevBlockHashKey))
	assert.Equal(t, []byte(nil), (*storage).Get(append(keySystemHash, keyBlockWeightHash...)))

	assert.Equal(t, expectedBlockHash.ToBytes(), (*storage).Get(blockHashKey))
}
