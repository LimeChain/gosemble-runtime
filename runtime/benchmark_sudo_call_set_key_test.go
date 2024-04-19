package main

import (
	"github.com/LimeChain/gosemble/benchmarking"
	"github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func BenchmarkSudoSetKey(b *testing.B) {
	benchmarking.RunDispatchCall(b, "../frame/sudo/call_set_key_weight.go", func(i *benchmarking.Instance) {
		err := (*i.Storage()).Put(append(keySudoHash, keyKeyHash...), aliceAddress.AsID.ToBytes())
		assert.NoError(b, err)

		err = i.ExecuteExtrinsic(
			"Sudo.set_key",
			types.NewRawOriginSigned(aliceAccountId),
			bobAddress,
		)
		assert.NoError(b, err)

		assert.Equal(b, bobAddress.AsID.ToBytes(), (*i.Storage()).Get(append(keySudoHash, keyKeyHash...)))
	})
}
