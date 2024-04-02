package main

import (
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/benchmarking"
	"github.com/LimeChain/gosemble/primitives/types"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	ctypes "github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/stretchr/testify/assert"
)

// Benchmark extrinsic with the worst possible conditions:
// * Transfer will kill the sender account.
// * Transfer will create the recipient account.
func BenchmarkBalancesForceTransfer(b *testing.B) {
	benchmarking.RunDispatchCall(b, "../frame/balances/call_force_transfer_weight.go", func(i *benchmarking.Instance) {
		// arrange
		balance := BalancesExistentialDeposit.Mul(sc.NewU128(10))
		transferAmount := BalancesExistentialDeposit.Mul(sc.NewU128(9)).Add(sc.NewU128(1))

		accountInfo := primitives.AccountInfo{
			Data: primitives.AccountData{
				Free: balance,
			},
		}
		err := i.SetAccountInfoNew(aliceAccountIdBytes, accountInfo)
		assert.NoError(b, err)

		// act
		err = i.ExecuteExtrinsic(
			"Balances.force_transfer",
			types.NewRawOriginRoot(),
			aliceAddress,
			bobAddress,
			ctypes.NewUCompact(transferAmount.ToBigInt()),
		)

		// assert
		assert.NoError(b, err)

		senderInfo, err := i.GetAccountInfoNew(aliceAccountIdBytes)
		assert.NoError(b, err)
		assert.Zero(b, senderInfo.Data.Free)

		recipientInfo, err := i.GetAccountInfoNew(bobAccountIdBytes)
		assert.NoError(b, err)
		assert.Equal(b, transferAmount, recipientInfo.Data.Free)
	})
}
