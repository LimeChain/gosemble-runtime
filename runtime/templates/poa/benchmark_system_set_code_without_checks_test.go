package main

import (
	"testing"

	"github.com/LimeChain/gosemble/benchmarking"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
)

func BenchmarkSystemSetCodeWithoutChecks(b *testing.B) {
	benchmarking.RunDispatchCall(b, "../frame/system/call_set_code_without_checks_weight.go", func(i *benchmarking.Instance) {
		// It is possible to pass more than 5MB due to the fact that
		// we are not executing benchmarks through the normal flow
		code := make([]byte, 6*1024*1024)

		err := i.ExecuteExtrinsic(
			"System.set_code_without_checks",
			primitives.NewRawOriginRoot(),
			code,
		)

		assert.NoError(b, err)

		assertStorageSystemEventCount(b, i.Storage(), uint32(1))
	})
}
