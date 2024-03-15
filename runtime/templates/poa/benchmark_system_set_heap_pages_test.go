package main

import (
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/benchmarking"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
)

func BenchmarkSystemSetHeapPages(b *testing.B) {
	benchmarking.RunDispatchCall(b, "../frame/system/call_set_heap_pages_weight.go", func(i *benchmarking.Instance) {
		pages := sc.U64(0)

		err := i.ExecuteExtrinsic(
			"System.set_heap_pages",
			primitives.NewRawOriginRoot(),
			pages,
		)

		assert.NoError(b, err)
	})
}
