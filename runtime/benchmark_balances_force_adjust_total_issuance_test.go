package main

import (
	"testing"

	// "github.com/ChainSafe/gossamer/lib/crypto/sr25519"

	// "github.com/ChainSafe/gossamer/lib/common"
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/benchmarking"
	balancestypes "github.com/LimeChain/gosemble/frame/balances/types"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	ctypes "github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/stretchr/testify/assert"
)

// var (
// 	keyStorageTotalIssuance = keyStorageTotalIssuance
// )

func BenchmarkBalancesForceAdjustTotalIssuance(b *testing.B) {
	benchmarking.RunDispatchCall(b, "../frame/balances/call_force_adjust_total_issuance_weight.go", func(i *benchmarking.Instance) {
		totalIssuance := (*i.Storage()).Get(keyStorageTotalIssuance)
		assert.Zero(b, totalIssuance)

		err := i.ExecuteExtrinsic(
			"Balances.force_adjust_total_issuance",
			primitives.NewRawOriginRoot(),
			balancestypes.AdjustmentDirectionIncrease,
			ctypes.NewUCompactFromUInt(123),
		)
		assert.NoError(b, err)

		totalIssuance = (*i.Storage()).Get(keyStorageTotalIssuance)
		assert.Equal(b, sc.NewU128(123).Bytes(), totalIssuance)
	})
}
