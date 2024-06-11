package main

import (
	"os"
	"testing"

	"github.com/ChainSafe/gossamer/lib/common"
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/benchmarking"
	"github.com/LimeChain/gosemble/frame/system"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/LimeChain/gosemble/testhelpers"
	"github.com/stretchr/testify/assert"
)

func BenchmarkSystemApplyAuthorizedUpgrade(b *testing.B) {
	benchmarking.RunDispatchCall(b, "../../../frame/system/call_apply_authorized_upgrade_weight.go", func(i *benchmarking.Instance) {
		codeSpecVersion101, err := os.ReadFile(testhelpers.RuntimeWasmSpecVersion101)
		assert.NoError(b, err)
		codeHash := common.MustBlake2bHash(codeSpecVersion101)
		hash, err := primitives.NewH256(sc.BytesToFixedSequenceU8(codeHash.ToBytes())...)
		assert.NoError(b, err)

		upgradeAuthorization := system.CodeUpgradeAuthorization{
			CodeHash:     hash,
			CheckVersion: true,
		}

		err = (*i.Storage()).Put(append(testhelpers.KeySystemHash, testhelpers.KeyAuthorizedUpgradeHash...), upgradeAuthorization.Bytes())
		assert.NoError(b, err)

		err = i.ExecuteExtrinsic(
			"System.apply_authorized_upgrade",
			primitives.NewRawOriginRoot(),
			codeSpecVersion101,
		)

		assert.NoError(b, err)
		upgradeAuthorizationBytes := (*i.Storage()).Get(append(testhelpers.KeySystemHash, testhelpers.KeyAuthorizedUpgradeHash...))
		assert.Equal(b, []byte(nil), upgradeAuthorizationBytes)
	})
}
