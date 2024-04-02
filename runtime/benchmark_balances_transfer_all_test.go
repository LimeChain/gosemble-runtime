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

// Benchmark `transfer_all` with the worst possible condition:
// * The recipient account is created
// * The sender is killed
func BenchmarkBalancesTransferAllAllowDeath(b *testing.B) {
	benchmarking.RunDispatchCall(b, "../frame/balances/call_transfer_all_weight.go", func(i *benchmarking.Instance) {
		// arrange
		balance := BalancesExistentialDeposit.Mul(sc.NewU128(10))

		accountInfo := primitives.AccountInfo{
			Data: primitives.AccountData{
				Free: balance,
			},
		}

		err := i.SetAccountInfoNew(aliceAccountIdBytes, accountInfo)
		assert.NoError(b, err)

		// act
		err = i.ExecuteExtrinsic(
			"Balances.transfer_all",
			types.NewRawOriginSigned(aliceAccountId),
			bobAddress,
			ctypes.NewBool(false),
		)

		// assert
		assert.NoError(b, err)

		senderInfo, err := i.GetAccountInfoNew(aliceAccountIdBytes)
		assert.NoError(b, err)
		assert.Equal(b, sc.NewU128(0), senderInfo.Data.Free)

		recipientInfo, err := i.GetAccountInfoNew(bobAccountIdBytes)
		assert.NoError(b, err)
		assert.Equal(b, balance, recipientInfo.Data.Free)
	})
}
