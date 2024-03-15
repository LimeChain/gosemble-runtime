package main

import (
	"testing"

	"github.com/LimeChain/gosemble/benchmarking"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
)

func BenchmarkSystemKillPrefix(b *testing.B) {
	limit, err := benchmarking.NewLinear("size", 0, uint32(1000))
	assert.NoError(b, err)

	benchmarking.RunDispatchCall(b, "../frame/system/call_kill_prefix_weight.go", func(i *benchmarking.Instance) {
		keys := make([][]byte, limit.Value())
		for j := range keys {
			keys[j] = buildBytes("testkey", j)
			(*i.Storage()).Put(keys[j], []byte("value"))
		}
		prefix := []byte("test")
		subkeys := uint32(limit.Value())

		err := i.ExecuteExtrinsic(
			"System.kill_prefix",
			primitives.NewRawOriginRoot(),
			prefix,
			subkeys,
		)

		assert.NoError(b, err)
		for j := range keys {
			assert.Equal(b, []byte(nil), (*i.Storage()).Get(keys[j]))
		}
	}, limit)
}
