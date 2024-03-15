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

func BenchmarkBalancesForceFree(b *testing.B) {
	benchmarking.RunDispatchCall(b, "../frame/balances/call_force_free_weight.go", func(i *benchmarking.Instance) {
		// arrange
		accountInfo := gossamertypes.AccountInfo{
			Nonce:       0,
			Consumers:   0,
			Producers:   0,
			Sufficients: 0,
			Data: gossamertypes.AccountData{
				Free:       scale.MustNewUint128(big.NewInt(existentialAmount)),
				Reserved:   scale.MustNewUint128(big.NewInt(existentialAmount)),
				MiscFrozen: scale.MustNewUint128(big.NewInt(0)),
				FreeFrozen: scale.MustNewUint128(big.NewInt(0)),
			},
		}

		err := i.SetAccountInfo(aliceAccountIdBytes, accountInfo)
		assert.NoError(b, err)

		// act
		err = i.ExecuteExtrinsic(
			"Balances.force_free",
			types.NewRawOriginRoot(),
			aliceAddress,
			ctypes.NewU128(*big.NewInt(2 * existentialAmount)),
		)

		// assert
		assert.NoError(b, err)

		existingAccountInfo, err := i.GetAccountInfo(aliceAccountIdBytes)
		assert.NoError(b, err)
		assert.Equal(b, scale.MustNewUint128(big.NewInt(0)), existingAccountInfo.Data.Reserved)
		assert.Equal(b, scale.MustNewUint128(big.NewInt(2*existentialAmount)), existingAccountInfo.Data.Free)
	})
}
