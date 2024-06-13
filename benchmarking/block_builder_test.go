package benchmarking

import (
	"encoding/hex"
	"path/filepath"
	"testing"
	"time"

	inmemory_trie "github.com/ChainSafe/gossamer/pkg/trie/inmemory"

	wazero_runtime "github.com/ChainSafe/gossamer/lib/runtime/wazero"
	sc "github.com/LimeChain/goscale"
	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	ctypes "github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/stretchr/testify/assert"
)

func Test_BlockBuilder(t *testing.T) {
	t.Skip()
	dateTime = uint64(time.Date(2024, time.January, 2, 3, 4, 5, 6, time.UTC).UnixMilli())
	blockNumber = uint(1)
	inherentData, err := timestampInherentData(dateTime)
	assert.NoError(t, err)

	expectTimestampExtrinsic, err := hex.DecodeString("0401000b88d820c88c01")
	assert.Nil(t, err)

	testing.Benchmark(func(b *testing.B) {
		runtime := wazero_runtime.NewBenchInstanceWithTrie(b, filepath.Join("../build/", RuntimeWasmBinary), inmemory_trie.NewEmptyTrie())
		defer runtime.Stop()

		instance, err := newBenchmarkingInstance(runtime, 1)
		assert.NoError(t, err)
		assert.Equal(t, &runtime.Context.Storage, instance.Storage())

		err = instance.BuildGenesisConfig()
		assert.Nil(t, err)

		signetureOptions := ctypes.SignatureOptions{
			BlockHash:          ctypes.Hash(parentHash),
			Era:                ctypes.ExtrinsicEra{IsImmortalEra: true},
			GenesisHash:        ctypes.Hash(parentHash),
			Nonce:              ctypes.NewUCompactFromUInt(uint64(0)),
			SpecVersion:        ctypes.U32(instance.version.SpecVersion),
			Tip:                ctypes.NewUCompactFromUInt(0),
			TransactionVersion: ctypes.U32(instance.version.TransactionVersion),
		}

		extrinsic, err := instance.newSignedExtrinsic(signature.TestKeyringPairAlice, signetureOptions, "System.remark", []byte{})

		target := NewBlockBuilder(instance, inherentData)
		assert.Equal(t, inherentData, target.inherentData)
		assert.Nil(t, target.extrinsics)

		err = target.StartSimulation(blockNumber)
		assert.Nil(t, err)
		assert.Equal(t, sc.U64(blockNumber).Bytes(), (*instance.storage).Get(append(keySystemHash, keyNumberHash...)))

		err = target.ApplyInherentExtrinsics()
		assert.Nil(t, err)
		assert.Equal(t, [][]byte{expectTimestampExtrinsic}, target.extrinsics)

		hasReachedLimit, err := target.AddExtrinsic(extrinsic)
		assert.NoError(t, err)
		assert.False(t, hasReachedLimit)
		assert.Equal(t, 2, len(target.extrinsics))

		block, err := target.FinishSimulation()
		assert.Nil(t, err)
		assert.Equal(t, block.Header.Number, blockNumber)
		assert.Equal(t, 2, len(block.Body))

		assert.Equal(t, []byte(nil), (*instance.storage).Get(append(keySystemHash, keyNumberHash...)))
	})
}
