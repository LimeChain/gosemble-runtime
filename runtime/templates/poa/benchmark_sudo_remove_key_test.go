package main

import (
	"testing"

	"github.com/LimeChain/gosemble/benchmarking"
	"github.com/LimeChain/gosemble/primitives/types"
	"github.com/LimeChain/gosemble/testhelpers"
	"github.com/stretchr/testify/assert"
)

func BenchmarkSudoRemoveKey(b *testing.B) {
	benchmarking.RunDispatchCall(b, "../frame/sudo/call_remove_key_weight.go", func(i *benchmarking.Instance) {
		err := (*i.Storage()).Put(append(testhelpers.KeySudoHash, testhelpers.KeyKeyHash...), aliceAddress.AsID.ToBytes())
		assert.NoError(b, err)

		err = i.ExecuteExtrinsic(
			"Sudo.remove_key",
			types.NewRawOriginSigned(aliceAccountId),
		)
		assert.NoError(b, err)

		assert.Nil(b, (*i.Storage()).Get(append(testhelpers.KeySudoHash, testhelpers.KeyKeyHash...)))
	})
}
