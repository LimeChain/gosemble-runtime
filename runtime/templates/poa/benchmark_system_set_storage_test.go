package main

import (
	"strconv"
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/benchmarking"
	"github.com/LimeChain/gosemble/frame/system"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
)

func BenchmarkSystemSetStorage(b *testing.B) {
	size, err := benchmarking.NewLinear("size", 0, uint32(1000))
	assert.NoError(b, err)

	benchmarking.RunDispatchCall(b, "../frame/system/call_set_storage_weight.go", func(i *benchmarking.Instance) {
		items := make([]system.KeyValue, size.Value())
		for j := range items {
			items[j].Key = buildSequence("key", j)
			items[j].Value = buildSequence("value", j)
		}

		err := i.ExecuteExtrinsic(
			"System.set_storage",
			primitives.NewRawOriginRoot(),
			items,
		)

		assert.NoError(b, err)
		for j := range items {
			value := (*i.Storage()).Get(sc.SequenceU8ToBytes(items[j].Key))
			assert.Equal(b, buildBytes("value", j), value)
		}
	}, size)
}

func buildSequence(name string, i int) sc.Sequence[sc.U8] {
	return sc.BytesToSequenceU8(buildBytes(name, i))
}

func buildBytes(name string, i int) []byte {
	return []byte(name + strconv.Itoa(i))
}
