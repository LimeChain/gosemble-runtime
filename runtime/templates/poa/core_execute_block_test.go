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
	"github.com/LimeChain/gosemble/frame/aura"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/LimeChain/gosemble/testhelpers"

	"github.com/stretchr/testify/assert"
)

var (
	dateTime    = time.Date(2023, time.January, 2, 3, 4, 5, 6, time.UTC)
	storageRoot = common.MustHexToHash("0xd940e147feef433028c8ca2db9ef6c7c51bf6e9538b81301fff3ff24950fa056") // Depends on date
)

func Test_BlockExecution(t *testing.T) {
	// core.InitializeBlock
	// blockBuilder.InherentExtrinsics
	// blockBuilder.ApplyExtrinsics
	// blockBuilder.FinalizeBlock

	expectedStorageDigest := gossamertypes.NewDigest()
	digest := gossamertypes.NewDigest()

	rt, storage := testhelpers.NewRuntimeInstance(t)
	metadata := testhelpers.RuntimeMetadata(t, rt)

	bytesSlotDuration, err := rt.Exec("AuraApi_slot_duration", []byte{})
	assert.NoError(t, err)

	buffer := &bytes.Buffer{}
	buffer.Write(bytesSlotDuration)

	slotDuration, err := sc.DecodeU64(buffer)
	assert.Nil(t, err)
	buffer.Reset()

	slot := sc.U64(dateTime.UnixMilli()) / slotDuration

	preRuntimeDigest := gossamertypes.PreRuntimeDigest{
		ConsensusEngineID: aura.EngineId,
		Data:              slot.Bytes(),
	}
	assert.NoError(t, digest.Add(preRuntimeDigest))
	assert.NoError(t, expectedStorageDigest.Add(preRuntimeDigest))

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

	idata := gossamertypes.NewInherentData()
	err = idata.SetInherent(gossamertypes.Timstap0, uint64(dateTime.UnixMilli()))
	assert.NoError(t, err)

	expectedExtrinsicBytes := testhelpers.TimestampExtrinsicBytes(t, metadata, uint64(dateTime.UnixMilli()))

	ienc, err := idata.Encode()
	assert.NoError(t, err)

	inherentExt, err := rt.Exec("BlockBuilder_inherent_extrinsics", ienc)
	assert.NoError(t, err)
	assert.NotNil(t, inherentExt)

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

	resultHeader := gossamertypes.NewEmptyHeader()
	assert.NoError(t, scale.Unmarshal(bytesResult, resultHeader))
	resultHeader.Hash() // Call this to be set, otherwise structs do not match...

	assert.Equal(t, header, resultHeader)

	assert.Equal(t, []byte(nil), (*storage).Get(append(testhelpers.KeyTimestampHash, testhelpers.KeyTimestampDidUpdateHash...)))
	assert.Equal(t, sc.U64(dateTime.UnixMilli()).Bytes(), (*storage).Get(append(testhelpers.KeyTimestampHash, testhelpers.KeyTimestampNowHash...)))

	assert.Equal(t, []byte(nil), (*storage).Get(testhelpers.KeyExtrinsicIndex))
	assert.Equal(t, []byte(nil), (*storage).Get(append(testhelpers.KeySystemHash, testhelpers.KeyExecutionPhaseHash...)))
	assert.Equal(t, []byte(nil), (*storage).Get(append(testhelpers.KeySystemHash, testhelpers.KeyAllExtrinsicsLenHash...)))
	assert.Equal(t, []byte(nil), (*storage).Get(append(testhelpers.KeySystemHash, testhelpers.KeyExtrinsicCountHash...)))

	assert.Equal(t, testhelpers.ParentHash.ToBytes(), (*storage).Get(append(testhelpers.KeySystemHash, testhelpers.KeyParentHash...)))
	assert.Equal(t, encExpectedDigest, (*storage).Get(append(testhelpers.KeySystemHash, testhelpers.KeyDigestHash...)))
	assert.Equal(t, encBlockNumber, (*storage).Get(append(testhelpers.KeySystemHash, testhelpers.KeyNumberHash...)))

	assert.Equal(t, slot.Bytes(), (*storage).Get(append(testhelpers.KeyAuraHash, testhelpers.KeyCurrentSlotHash...)))
}

func Test_ExecuteBlock(t *testing.T) {
	// blockBuilder.Inherent_Extrinsics
	// blockBuilder.ExecuteBlock

	rt, storage := testhelpers.NewRuntimeInstance(t)
	metadata := testhelpers.RuntimeMetadata(t, rt)

	bytesSlotDuration, err := rt.Exec("AuraApi_slot_duration", []byte{})
	assert.NoError(t, err)

	buffer := &bytes.Buffer{}
	buffer.Write(bytesSlotDuration)

	slotDuration, err := sc.DecodeU64(buffer)
	assert.Nil(t, err)
	buffer.Reset()

	slot := sc.U64(dateTime.UnixMilli()) / slotDuration

	preRuntimeDigest := gossamertypes.PreRuntimeDigest{
		ConsensusEngineID: aura.EngineId,
		Data:              slot.Bytes(),
	}

	idata := gossamertypes.NewInherentData()
	err = idata.SetInherent(gossamertypes.Timstap0, uint64(dateTime.UnixMilli()))

	assert.NoError(t, err)

	ienc, err := idata.Encode()
	assert.NoError(t, err)

	expectedExtrinsicBytes := testhelpers.TimestampExtrinsicBytes(t, metadata, uint64(dateTime.UnixMilli()))

	inherentExt, err := rt.Exec("BlockBuilder_inherent_extrinsics", ienc)
	assert.NoError(t, err)
	assert.NotNil(t, inherentExt)

	buffer.Write([]byte{inherentExt[0]})

	totalInherents, err := sc.DecodeCompact[sc.U128](buffer)
	assert.Nil(t, err)
	assert.Equal(t, int64(1), totalInherents.ToBigInt().Int64())
	buffer.Reset()

	actualExtrinsic := inherentExt[1:]
	assert.Equal(t, expectedExtrinsicBytes, actualExtrinsic)

	var exts [][]byte
	err = scale.Unmarshal(inherentExt, &exts)
	assert.Nil(t, err)

	digest := gossamertypes.NewDigest()

	assert.NoError(t, err)
	assert.NoError(t, digest.Add(preRuntimeDigest))

	expectedStorageDigest, err := scale.Marshal(digest)
	assert.NoError(t, err)
	encBlockNumber, err := scale.Marshal(testhelpers.BlockNumber)
	assert.NoError(t, err)

	header := gossamertypes.NewHeader(testhelpers.ParentHash, storageRoot, testhelpers.ExtrinsicsRoot, uint(testhelpers.BlockNumber), digest)

	block := gossamertypes.Block{
		Header: *header,
		Body:   gossamertypes.BytesArrayToExtrinsics(exts),
	}

	encodedBlock, err := scale.Marshal(block)
	assert.Nil(t, err)

	_, err = rt.Exec("Core_execute_block", encodedBlock)
	assert.NoError(t, err)

	assert.Equal(t, []byte(nil), (*storage).Get(append(testhelpers.KeyTimestampHash, testhelpers.KeyTimestampDidUpdateHash...)))
	assert.Equal(t, sc.U64(dateTime.UnixMilli()).Bytes(), (*storage).Get(append(testhelpers.KeyTimestampHash, testhelpers.KeyTimestampNowHash...)))

	assert.Equal(t, []byte(nil), (*storage).Get(testhelpers.KeyExtrinsicIndex))
	assert.Equal(t, []byte(nil), (*storage).Get(append(testhelpers.KeySystemHash, testhelpers.KeyExecutionPhaseHash...)))
	assert.Equal(t, []byte(nil), (*storage).Get(append(testhelpers.KeySystemHash, testhelpers.KeyAllExtrinsicsLenHash...)))
	assert.Equal(t, []byte(nil), (*storage).Get(append(testhelpers.KeySystemHash, testhelpers.KeyExtrinsicCountHash...)))

	assert.Equal(t, testhelpers.ParentHash.ToBytes(), (*storage).Get(append(testhelpers.KeySystemHash, testhelpers.KeyParentHash...)))
	assert.Equal(t, expectedStorageDigest, (*storage).Get(append(testhelpers.KeySystemHash, testhelpers.KeyDigestHash...)))
	assert.Equal(t, encBlockNumber, (*storage).Get(append(testhelpers.KeySystemHash, testhelpers.KeyNumberHash...)))

	assert.Equal(t, slot.Bytes(), (*storage).Get(append(testhelpers.KeyAuraHash, testhelpers.KeyCurrentSlotHash...)))
}
