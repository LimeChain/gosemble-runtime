package main

import (
	"math/big"
	"testing"

	gossamertypes "github.com/ChainSafe/gossamer/dot/types"
	"github.com/ChainSafe/gossamer/pkg/scale"
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/benchmarking"
	"github.com/LimeChain/gosemble/primitives/types"
	"github.com/LimeChain/gosemble/testhelpers"
	ctypes "github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/stretchr/testify/assert"
)

// Benchmark `transfer_all` with the worst possible condition:
// * The recipient account is created
// * The sender is killed
func BenchmarkBalancesTransferAllAllowDeath(b *testing.B) {
	benchmarking.RunDispatchCall(b, "../../../frame/balances/call_transfer_all_weight.go", func(i *benchmarking.Instance) {
		balance := existentialMultiplier * existentialAmount

		accountInfo := gossamertypes.AccountInfo{
			Nonce:       0,
			Consumers:   0,
			Producers:   1,
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

		keyTotalIssuance := append(testhelpers.KeyBalancesHash, testhelpers.KeyTotalIssuanceHash...)
		err = (*i.Storage()).Put(keyTotalIssuance, sc.NewU128(balance).Bytes())
		assert.NoError(b, err)

		err = i.ExecuteExtrinsic(
			"Balances.transfer_all",
			types.NewRawOriginSigned(aliceAccountId),
			bobAddress,
			ctypes.NewBool(false),
		)

		assert.NoError(b, err)

		senderInfo, err := i.GetAccountInfo(aliceAccountIdBytes)
		assert.NoError(b, err)
		assert.Equal(b, scale.MustNewUint128(big.NewInt(int64(0))), senderInfo.Data.Free)

		recipientInfo, err := i.GetAccountInfo(bobAccountIdBytes)
		assert.NoError(b, err)
		assert.Equal(b, scale.MustNewUint128(big.NewInt(balance)), recipientInfo.Data.Free)
	})
}
