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

var value = uint64(existentialMultiplier * existentialAmount)

// Coming from ROOT account. This always creates an account.
func BenchmarkBalancesSetBalanceCreating(b *testing.B) {
	benchmarkBalancesSetBalance(b, "../frame/balances/call_set_balance_creating_weight.go", value, value)
}

func BenchmarkBalancesSetBalanceKilling(b *testing.B) {
	benchmarkBalancesSetBalance(b, "../frame/balances/call_set_balance_killing_weight.go", value, 0)
}

func benchmarkBalancesSetBalance(b *testing.B, outputPath string, balance, amount uint64) {
	benchmarking.RunDispatchCall(b, outputPath, func(i *benchmarking.Instance) {
		// arrange
		accountInfo := gossamertypes.AccountInfo{
			Nonce:       0,
			Consumers:   0,
			Producers:   1,
			Sufficients: 0,
			Data: gossamertypes.AccountData{
				Free:       scale.MustNewUint128(big.NewInt(int64(balance))),
				Reserved:   scale.MustNewUint128(big.NewInt(0)),
				MiscFrozen: scale.MustNewUint128(big.NewInt(0)),
				FreeFrozen: scale.MustNewUint128(big.NewInt(0)),
			},
		}
		err := i.SetAccountInfo(aliceAccountIdBytes, accountInfo)
		assert.NoError(b, err)

		// act
		err = i.ExecuteExtrinsic(
			"Balances.set_balance",
			types.NewRawOriginRoot(),
			aliceAddress,
			ctypes.NewUCompactFromUInt(amount),
			ctypes.NewUCompactFromUInt(amount),
		)

		// assert
		assert.NoError(b, err)

		senderInfo, err := i.GetAccountInfo(aliceAccountIdBytes)
		assert.NoError(b, err)
		assert.Equal(b, scale.MustNewUint128(big.NewInt(int64(amount))), senderInfo.Data.Free)
	})
}
