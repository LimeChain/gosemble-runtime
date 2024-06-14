package benchmarking

import (
	"fmt"
	"math"
	"testing"
	"time"

	gossamertypes "github.com/ChainSafe/gossamer/dot/types"
	wazero_runtime "github.com/ChainSafe/gossamer/lib/runtime/wazero"
	"github.com/ChainSafe/gossamer/pkg/scale"
	inmemory_trie "github.com/ChainSafe/gossamer/pkg/trie/inmemory"
	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	ctypes "github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

var (
	signer = signature.TestKeyringPairAlice
)

func BenchmarkOverheadBlockExecutionWeight(t *testing.B) {
	// todo set heapPages and dbCache when Gossamer starts supporting db caching
	runtime := wazero_runtime.NewBenchInstanceWithTrie(t, Config.WasmRuntime, inmemory_trie.NewEmptyTrie())
	defer runtime.Stop()

	instance, err := newBenchmarkingInstance(runtime, Config.Overhead.Repeat)
	if err != nil {
		t.Fatalf("failed to create benchmarking instance: [%v]", err)
	}

	err = instance.BuildGenesisConfig()
	if err != nil {
		t.Fatal(err)
	}

	stats := benchBlock(t, instance, Config.Overhead)
	fmt.Println(stats.String())

	if Config.GenerateWeightFiles {
		if err := generateOverheadWeightFile("../constants/block_execution_weight.go", stats); err != nil {
			t.Fatalf("failed to generate weight file: %v", err)
		}
	}
}

func BenchmarkOverheadBaseExtrinsicWeight(t *testing.B) {
	// todo set heapPages and dbCache when Gossamer starts supporting db caching
	runtime := wazero_runtime.NewBenchInstanceWithTrie(t, Config.WasmRuntime, inmemory_trie.NewEmptyTrie())
	defer runtime.Stop()

	instance, err := newBenchmarkingInstance(runtime, Config.Overhead.Repeat)
	if err != nil {
		t.Fatalf("failed to create benchmarking instance: [%v]", err)
	}

	err = instance.BuildGenesisConfig()
	if err != nil {
		t.Fatal(err)
	}

	stats := benchExtrinsic(t, instance, Config.Overhead)
	fmt.Println(stats.String())

	if Config.GenerateWeightFiles {
		if err := generateOverheadWeightFile("../constants/base_extrinsic_weight.go", stats); err != nil {
			t.Fatalf("failed to generate weight file: %v", err)
		}
	}
}

func benchBlock(b *testing.B, instance *Instance, config overheadConfig) StatsResult {
	// Build the block
	block, err := buildBlock(instance, false, config.MaxExtPerBlock)
	if err != nil {
		b.Fatalf("failed to build empty block: [%v]", err)
	}

	// Measure the block
	results, err := measureBlock(instance, config, block)
	if err != nil {
		b.Fatal(err)
	}

	// Create the stats
	stats, err := NewStatsResult(results)
	if err != nil {
		b.Fatal(err)
	}

	return stats
}

func benchExtrinsic(b *testing.B, instance *Instance, config overheadConfig) StatsResult {
	// Build an Empty block
	baseEmptyBlock, err := buildBlock(instance, false, config.MaxExtPerBlock)
	if err != nil {
		b.Fatalf("failed to build empty block: [%v]", err)
	}

	// Measure the Empty block
	baseResults, err := measureBlock(instance, config, baseEmptyBlock)
	if err != nil {
		b.Fatal(err)
	}

	// Create the Empty block stats
	baseStats, err := NewStatsResult(baseResults)
	if err != nil {
		b.Fatal(err)
	}

	// Build a Block with maximum possible Extrinsics
	block, err := buildBlock(instance, true, config.MaxExtPerBlock)
	if err != nil {
		b.Fatalf("failed to build block with extrinsics: [%v]", err)
	}
	totalExtrinsics := len(block.Body)

	// Measure the block
	extrinsicResults, err := measureBlock(instance, config, block)
	if err != nil {
		b.Fatal(err)
	}

	for i, extrinsicResult := range extrinsicResults {
		// Subtract the base stats
		extrinsicResult = math.Max(extrinsicResult-baseStats.Mean, 0)

		// Divide by the total extrinsics
		extrinsicResult = math.Ceil(extrinsicResult / float64(totalExtrinsics))

		extrinsicResults[i] = extrinsicResult
	}

	// Create the stats
	stats, err := NewStatsResult(extrinsicResults)
	if err != nil {
		b.Fatal(err)
	}
	return stats
}

func buildBlock(instance *Instance, hasExtrinsics bool, maxExtrinsicsPerBlock int) (gossamertypes.Block, error) {
	inherentData, err := timestampInherentData(dateTime)
	if err != nil {
		return gossamertypes.Block{}, fmt.Errorf("failed to create inherent data: [%v]", err)
	}

	blockBuilder := NewBlockBuilder(instance, inherentData)

	err = blockBuilder.StartSimulation(blockNumber)
	if err != nil {
		return gossamertypes.Block{}, fmt.Errorf("failed to start simulation: [%v]", err)
	}

	err = blockBuilder.ApplyInherentExtrinsics()
	if err != nil {
		return gossamertypes.Block{}, fmt.Errorf("failed to create inherent extrinsics: [%v]", err)
	}

	if hasExtrinsics {
		fmt.Println("Building block, this takes some time...")

		for i := 0; i < maxExtrinsicsPerBlock; i++ {
			signatureOptions := ctypes.SignatureOptions{
				BlockHash:          ctypes.Hash(parentHash),
				Era:                ctypes.ExtrinsicEra{IsImmortalEra: true},
				GenesisHash:        ctypes.Hash(parentHash),
				Nonce:              ctypes.NewUCompactFromUInt(uint64(i)),
				SpecVersion:        ctypes.U32(instance.version.SpecVersion),
				Tip:                ctypes.NewUCompactFromUInt(0),
				TransactionVersion: ctypes.U32(instance.version.TransactionVersion),
			}

			extrinsic, err := instance.newSignedExtrinsic(signer, signatureOptions, "System.remark", []byte{})
			if err != nil {
				return gossamertypes.Block{}, fmt.Errorf("failed to sign extrinsic: [%v]", err)
			}

			limitReached, err := blockBuilder.AddExtrinsic(extrinsic)
			if err != nil {
				return gossamertypes.Block{}, fmt.Errorf("failed to add extrinsic: [%v]", err)
			}

			if limitReached {
				break
			}
		}
	}

	block, err := blockBuilder.FinishSimulation()
	if err != nil {
		return gossamertypes.Block{}, fmt.Errorf("failed to finish simulation: [%v]", err)
	}

	return block, nil
}

func measureBlock(instance *Instance, config overheadConfig, block gossamertypes.Block) ([]float64, error) {
	encodedBlock, err := scale.Marshal(block)
	if err != nil {
		return nil, fmt.Errorf("failed to encode block: [%v]", err)
	}

	// Store the DB Snapshot with only Genesis Block
	(*instance.storage).DbStoreSnapshot()

	fmt.Println(fmt.Sprintf("Running [%d] warmups", config.Warmup))
	for i := 0; i < config.Warmup; i++ {
		(*instance.storage).DbRestoreSnapshot()
		_, err := instance.runtime.Exec("Core_execute_block", encodedBlock)
		if err != nil {
			return nil, fmt.Errorf("failed to warmup execute block: [%v]", err)
		}
	}

	var results []float64
	fmt.Println(fmt.Sprintf("Executing block [%d] times", config.Repeat))
	for i := 0; i < config.Repeat; i++ {
		(*instance.storage).DbRestoreSnapshot()

		start := time.Now().UnixNano()
		_, err := instance.runtime.Exec("Core_execute_block", encodedBlock)
		if err != nil {
			return nil, fmt.Errorf("failed to execute block: [%v]", err)
		}
		end := time.Now().UnixNano()
		results = append(results, float64(end-start))
	}

	// Restore the DB Snapshot
	(*instance.storage).DbRestoreSnapshot()

	return results, nil
}
