package main

import (
	"github.com/LimeChain/gosemble/benchmarking"
	"github.com/LimeChain/gosemble/primitives/types"
	ctypes "github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func BenchmarkSudoSudoAs(b *testing.B) {
	benchmarking.RunDispatchCall(b, "../frame/sudo/call_sudo_as_weight.go", func(i *benchmarking.Instance) {
		err := (*i.Storage()).Put(append(keySudoHash, keyKeyHash...), aliceAddress.AsID.ToBytes())
		assert.NoError(b, err)

		call, err := ctypes.NewCall(i.Metadata(), "System.remark", []byte{})
		assert.NoError(b, err)

		err = i.ExecuteExtrinsic(
			"Sudo.sudo_as",
			types.NewRawOriginSigned(aliceAccountId),
			bobAddress,
			call,
		)
		assert.NoError(b, err)
	})
}
