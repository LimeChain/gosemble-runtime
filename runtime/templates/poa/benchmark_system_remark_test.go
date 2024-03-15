package main

import (
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/benchmarking"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
)

func BenchmarkSystemRemark(b *testing.B) {
	size, err := benchmarking.NewLinear("size", 0, uint32(blockLength.Max.Normal))
	assert.NoError(b, err)

	benchmarking.RunDispatchCall(b, "../frame/system/call_remark_weight.go", func(i *benchmarking.Instance) {
		message := make([]byte, sc.U32(size.Value()))

		err := i.ExecuteExtrinsic(
			"System.remark",
			primitives.NewRawOriginSigned(aliceAccountId),
			message,
		)

		assert.NoError(b, err)
	}, size)
}
