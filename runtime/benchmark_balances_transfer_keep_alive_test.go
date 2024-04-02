package main

import (
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/benchmarking"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	ctypes "github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/stretchr/testify/assert"
)

// Benchmark `transfer_keep_alive` with the worst possible condition:
// * The recipient account is created.
func BenchmarkBalancesTransferKeepAlive(b *testing.B) {
	benchmarking.RunDispatchCall(b, "../frame/balances/call_transfer_keep_alive_weight.go", func(i *benchmarking.Instance) {
		// arrange
		transferAmount := BalancesExistentialDeposit.Mul(sc.NewU128(10))

		accountInfo := primitives.AccountInfo{
			Providers: 1,
			Data: primitives.AccountData{
				Free: sc.MaxU128(),
			},
		}
		err := i.SetAccountInfoNew(aliceAccountIdBytes, accountInfo)
		assert.NoError(b, err)

		// act
		err = i.ExecuteExtrinsic(
			"Balances.transfer_keep_alive",
			primitives.NewRawOriginSigned(aliceAccountId),
			bobAddress,
			ctypes.NewUCompact(transferAmount.ToBigInt()),
		)

		// assert
		assert.NoError(b, err)

		senderAccInfo, err := i.GetAccountInfoNew(aliceAccountIdBytes)
		assert.NoError(b, err)
		assert.Equal(b, sc.MaxU128().Sub(transferAmount), senderAccInfo.Data.Free)

		recipientAccInfo, err := i.GetAccountInfoNew(bobAccountIdBytes)
		assert.NoError(b, err)
		assert.Equal(b, transferAmount, recipientAccInfo.Data.Free)
	})
}
