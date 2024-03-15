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

func BenchmarkBalancesTransferKeepAlive(b *testing.B) {
	benchmarking.RunDispatchCall(b, "../frame/balances/call_transfer_keep_alive_weight.go", func(i *benchmarking.Instance) {
		// arrange
		accountInfo := gossamertypes.AccountInfo{
			Nonce:       0,
			Consumers:   0,
			Producers:   0,
			Sufficients: 0,
			Data: gossamertypes.AccountData{
				Free:       scale.MaxUint128,
				Reserved:   scale.MustNewUint128(big.NewInt(existentialAmount)),
				MiscFrozen: scale.MustNewUint128(big.NewInt(0)),
				FreeFrozen: scale.MustNewUint128(big.NewInt(0)),
			},
		}

		err := i.SetAccountInfo(aliceAccountIdBytes, accountInfo)
		assert.NoError(b, err)

		transferAmount := existentialMultiplier * existentialAmount

		// act
		err = i.ExecuteExtrinsic(
			"Balances.transfer_keep_alive",
			types.NewRawOriginSigned(aliceAccountId),
			bobAddress,
			ctypes.NewUCompactFromUInt(uint64(transferAmount)),
		)

		// assert
		assert.NoError(b, err)

		expectedSenderBalance, ok := new(big.Int).SetString(scale.MaxUint128.String(), 10)
		assert.True(b, ok)
		expectedSenderBalance = expectedSenderBalance.Sub(big.NewInt(transferAmount), expectedSenderBalance)

		senderAccInfo, err := i.GetAccountInfo(aliceAccountIdBytes)
		assert.NoError(b, err)
		assert.Equal(b, scale.MustNewUint128(expectedSenderBalance), senderAccInfo.Data.Free)

		recipientAccInfo, err := i.GetAccountInfo(bobAccountIdBytes)
		assert.NoError(b, err)
		assert.Equal(b, scale.MustNewUint128(big.NewInt(transferAmount)), recipientAccInfo.Data.Free)
	})
}
