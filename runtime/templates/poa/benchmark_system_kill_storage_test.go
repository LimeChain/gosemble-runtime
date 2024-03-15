package main

import (
	"testing"

	"github.com/LimeChain/gosemble/benchmarking"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
)

func BenchmarkSystemKillStorage(b *testing.B) {
	size, err := benchmarking.NewLinear("size", 0, uint32(1000))
	assert.NoError(b, err)

	benchmarking.RunDispatchCall(b, "../frame/system/call_kill_storage_weight.go", func(i *benchmarking.Instance) {
		keys := make([][]byte, size.Value())
		for j := range keys {
			keys[j] = buildBytes("key", j)
			(*i.Storage()).Put(keys[j], []byte("value"))
		}

		err := i.ExecuteExtrinsic(
			"System.kill_storage",
			primitives.NewRawOriginRoot(),
			keys,
		)

		assert.NoError(b, err)
		for j := range keys {
			assert.Equal(b, []byte(nil), (*i.Storage()).Get(keys[j]))
		}
	}, size)
}
