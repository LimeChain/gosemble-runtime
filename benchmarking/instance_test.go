package benchmarking

import (
	"bytes"
	"errors"
	"path/filepath"
	"testing"

	gossamertypes "github.com/ChainSafe/gossamer/dot/types"
	"github.com/ChainSafe/gossamer/lib/common"
	wazero_runtime "github.com/ChainSafe/gossamer/lib/runtime/wazero"
	inmemory_trie "github.com/ChainSafe/gossamer/pkg/trie/inmemory"
	sc "github.com/LimeChain/goscale"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	ctypes "github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/stretchr/testify/assert"
)

var (
	keyTimestampHash, _    = common.Twox128Hash([]byte("Timestamp"))
	keyTimestampNowHash, _ = common.Twox128Hash([]byte("Now"))
	aliceAddress, _        = ctypes.NewMultiAddressFromHexAccountID("0xd43593c715fdd31c61141abd04a99fd6822c8558854ccde39a5684e7a56da27d")
	aliceAccountIdBytes    = aliceAddress.AsID.ToBytes()
	testAccStorageKey      = []byte{0x26, 0xaa, 0x39, 0x4e, 0xea, 0x56, 0x30, 0xe0, 0x7c, 0x48, 0xae, 0xc, 0x95, 0x58, 0xce, 0xf7, 0xb9, 0x9d, 0x88, 0xe, 0xc6, 0x81, 0x79, 0x9c, 0xc, 0xf3, 0xe, 0x88, 0x86, 0x37, 0x1d, 0xa9, 0x44, 0xa8, 0x99, 0x5d, 0xd5, 0xb, 0x66, 0x57, 0xa0, 0x37, 0xa7, 0x83, 0x93, 0x4, 0x53, 0x5b, 0x74, 0x65, 0x73, 0x74}
)

func TestInstance(t *testing.T) {
	testing.Benchmark(func(b *testing.B) {
		t.Skip()
		runtime := wazero_runtime.NewBenchInstanceWithTrie(b, filepath.Join("../build/", RuntimeWasmBinary), inmemory_trie.NewEmptyTrie())
		defer runtime.Stop()

		instance, err := newBenchmarkingInstance(runtime, 1)
		assert.NoError(t, err)
		assert.Equal(t, &runtime.Context.Storage, instance.Storage())

		err = instance.InitializeBlock(blockNumber, dateTime)
		assert.NoError(t, err)

		err = instance.ExecuteExtrinsic(
			"Timestamp.set",
			primitives.NewRawOriginNone(),
			ctypes.NewUCompactFromUInt(dateTime),
		)
		assert.NoError(t, err)

		br := instance.benchmarkResult
		assert.NotNil(t, br)
		assert.Positive(t, br.Time.ToBigInt().Uint64())
		assert.Positive(t, br.Reads.ToBigInt().Uint64())
		assert.Positive(t, br.Writes.ToBigInt().Uint64())

		timestampStorageValue, err := sc.DecodeU64(bytes.NewBuffer((*instance.Storage()).Get(append(keyTimestampHash, keyTimestampNowHash...))))
		assert.NoError(t, err)
		assert.Equal(t, sc.U64(dateTime), timestampStorageValue)

		err = instance.SetAccountInfo(aliceAccountIdBytes, gossamertypes.AccountInfo{})
		assert.Equal(t, errors.New("failed to marshal account info: uint128 in nil"), err)

		err = instance.ExecuteExtrinsic(
			"Timestamp.set",
			primitives.NewRawOriginNone(),
			ctypes.NewUCompactFromUInt(dateTime),
		)
		assert.Equal(t, errOnlyOneCall, err)

		instance.benchmarkResult = nil

		err = instance.ExecuteExtrinsic(
			"Timestamp.set",
			primitives.RawOrigin{},
			ctypes.NewUCompactFromUInt(dateTime),
		)
		assert.Error(t, err)

		err = instance.ExecuteExtrinsic("Invalid.invalid", primitives.NewRawOriginNone())
		assert.Equal(t, errors.New("failed to create new call: module Invalid not found in metadata for call Invalid.invalid"), err)

		assert.Equal(t, testAccStorageKey, accountStorageKey([]byte("test")))
	})
}
