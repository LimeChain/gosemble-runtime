package main

import (
	"testing"

	"github.com/LimeChain/gosemble/benchmarking"
	"github.com/LimeChain/gosemble/primitives/types"
	"github.com/LimeChain/gosemble/testhelpers"
	ctypes "github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/stretchr/testify/assert"
)

func BenchmarkSudoSudo(b *testing.B) {
	benchmarking.RunDispatchCall(b, "../../../frame/sudo/call_sudo_weight.go", func(i *benchmarking.Instance) {
		err := (*i.Storage()).Put(append(testhelpers.KeySudoHash, testhelpers.KeyKeyHash...), aliceAddress.AsID.ToBytes())
		assert.NoError(b, err)

		call, err := ctypes.NewCall(i.Metadata(), "System.remark", []byte{})
		assert.NoError(b, err)

		err = i.ExecuteExtrinsic(
			"Sudo.sudo",
			types.NewRawOriginSigned(aliceAccountId),
			call,
		)
		assert.NoError(b, err)
	})
}
