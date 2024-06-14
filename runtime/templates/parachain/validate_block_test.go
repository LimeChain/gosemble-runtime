package main

import (
	"bytes"
	"encoding/hex"
	"os"
	"testing"

	"github.com/ChainSafe/gossamer/pkg/trie/db"
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/execution/types"
	"github.com/LimeChain/gosemble/frame/aura"
	"github.com/LimeChain/gosemble/primitives/parachain"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/LimeChain/gosemble/testhelpers"
	"github.com/stretchr/testify/assert"
)

var (
	headData, _   = hex.DecodeString("8e183fe6c2678ab2adff521a132dbad70cdaf76f51711afdc195eefb79d5c79908813b8e255d23bd7780aa88279b3cdba7da24d1fc6f2ea92830ae3c616f737ea7b89301f9e4cf99453e46c5c9e371e026c44e4f9cb51b3f43135d659172856a120c0661757261202a883533000000000452535052905c75a1168797f55f2e58ec94e530ed1527152c3e9a1945c5e41f5470cf8689f83600000005617572610101320528c807bdc6abf7f1247e112c0314820331c73560bb825e921af62fd28b04da68eeb475e5c850f527a1883ab20865eda4a70417c6a18ed72f65aa9182618f")
	hrmpWatermark = sc.U32(54)
)

// Executes `validate_block` using a valid preconfigured `ValidationParams`.
func Test_ValidateBlock(t *testing.T) {
	hexValidationParams, err := os.ReadFile("./validation_params.txt")
	assert.NoError(t, err)

	bytesValidationParams, err := hex.DecodeString(string(hexValidationParams))
	assert.NoError(t, err)

	rt, _ := testhelpers.NewParachainRuntimeInstance(t)

	result, err := rt.Exec("validate_block", bytesValidationParams)
	assert.NoError(t, err)

	validationResult, err := parachain.DecodeValidationResult(bytes.NewBuffer(result))
	assert.NoError(t, err)

	expect := parachain.ValidationResult{
		HeadData:                  sc.BytesToSequenceU8(headData),
		NewValidationCode:         sc.Option[sc.Sequence[sc.U8]]{},
		UpwardMessages:            sc.Sequence[parachain.UpwardMessage]{},
		HorizontalMessages:        sc.Sequence[parachain.OutboundHrmpMessage]{},
		ProcessedDownwardMessages: 0,
		HrmpWatermark:             hrmpWatermark,
	}

	assert.Equal(t, expect, validationResult)
}

// Rebuilds state from validation params storage proof and executes `Execute_block`.
func Test_ExecuteBlockFromTrieState(t *testing.T) {
	hexValidationParams, err := os.ReadFile("./validation_params.txt")
	assert.NoError(t, err)

	bytesValidationParams, err := hex.DecodeString(string(hexValidationParams))
	assert.NoError(t, err)

	validationParams, err := parachain.DecodeValidationParams(bytes.NewBuffer(bytesValidationParams))
	assert.NoError(t, err)

	decoder := types.NewRuntimeDecoder(modules, extra, sc.U8(0), ioStorage, ioTransactionBroker, logger)
	blockData, err := decoder.DecodeParachainBlockData(validationParams.BlockData)
	assert.NoError(t, err)

	database, err := db.NewMemoryDBFromProof(blockData.CompactProof.ToBytes())
	assert.NoError(t, err)

	parentHeader, err := primitives.DecodeHeader(bytes.NewBuffer(sc.SequenceU8ToBytes(validationParams.ParentHead)))
	assert.NoError(t, err)

	trie, err := parachain.BuildTrie(parentHeader.StateRoot.Bytes(), database)
	assert.NoError(t, err)
	rt, _ := testhelpers.NewParachainRuntimeInstanceWithTrie(t, trie)

	digestItems := testhelpers.ExtractConsensusDigests(t, blockData.Block.Header().Digest.Sequence, aura.EngineId[:])

	header := blockData.Block.Header()
	header.Digest = primitives.NewDigest(digestItems)

	block := types.NewBlock(header, blockData.Block.Extrinsics())

	_, err = rt.Exec("Core_execute_block", block.Bytes())
	assert.NoError(t, err)
}
