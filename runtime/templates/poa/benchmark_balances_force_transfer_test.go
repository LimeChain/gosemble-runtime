package main

import (
	"math/big"
	"testing"

	gossamertypes "github.com/ChainSafe/gossamer/dot/types"
	"github.com/ChainSafe/gossamer/pkg/scale"
	"github.com/LimeChain/gosemble/benchmarking"
	"github.com/LimeChain/gosemble/primitives/types"
	ctypes "github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/stretchr/testify/assert"
)

// Benchmark extrinsic with the worst possible conditions:
// * Transfer will kill the sender account.
// * Transfer will create the recipient account.
func BenchmarkBalancesForceTransfer(b *testing.B) {
	benchmarking.RunDispatchCall(b, "../frame/balances/call_force_transfer_weight.go", func(i *benchmarking.Instance) {
		// arrange
		balance := existentialMultiplier * existentialAmount
		transferAmount := uint64(existentialAmount*(existentialMultiplier-1) + 1)

		accountInfo := gossamertypes.AccountInfo{
			Nonce:       0,
			Consumers:   0,
			Producers:   0,
			Sufficients: 0,
			Data: gossamertypes.AccountData{
				Free:       scale.MustNewUint128(big.NewInt(balance)),
				Reserved:   scale.MustNewUint128(big.NewInt(0)),
				MiscFrozen: scale.MustNewUint128(big.NewInt(0)),
				FreeFrozen: scale.MustNewUint128(big.NewInt(0)),
			},
		}

		err := i.SetAccountInfo(aliceAccountIdBytes, accountInfo)
		assert.NoError(b, err)

		// act
		err = i.ExecuteExtrinsic(
			"Balances.force_transfer",
			types.NewRawOriginRoot(),
			aliceAddress,
			bobAddress,
			ctypes.NewUCompactFromUInt(transferAmount),
		)

		// assert
		assert.NoError(b, err)

		senderInfo, err := i.GetAccountInfo(aliceAccountIdBytes)
		assert.NoError(b, err)
		assert.Equal(b, scale.MustNewUint128(big.NewInt(balance-int64(transferAmount))), senderInfo.Data.Free)

		recipientInfo, err := i.GetAccountInfo(bobAccountIdBytes)
		assert.NoError(b, err)
		assert.Equal(b, scale.MustNewUint128(big.NewInt(int64(transferAmount))), recipientInfo.Data.Free)
	})
}
