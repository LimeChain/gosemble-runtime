package main

import (
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/benchmarking"
	"github.com/LimeChain/gosemble/frame/system"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
)

func BenchmarkSystemApplyAuthorizedUpgrade(b *testing.B) {
	benchmarking.RunDispatchCall(b, "../frame/system/call_apply_authorized_upgrade_weight.go", func(i *benchmarking.Instance) {
		hash, err := primitives.NewH256(sc.BytesToFixedSequenceU8(codeHash.ToBytes())...)
		assert.NoError(b, err)

		upgradeAuthorization := system.CodeUpgradeAuthorization{
			CodeHash:     hash,
			CheckVersion: true,
		}

		err = (*i.Storage()).Put(append(keySystemHash, keyAuthorizedUpgradeHash...), upgradeAuthorization.Bytes())
		assert.NoError(b, err)

		err = i.ExecuteExtrinsic(
			"System.apply_authorized_upgrade",
			primitives.NewRawOriginRoot(),
			codeSpecVersion101,
		)

		assert.NoError(b, err)
		upgradeAuthorizationBytes := (*i.Storage()).Get(append(keySystemHash, keyAuthorizedUpgradeHash...))
		assert.Equal(b, []byte(nil), upgradeAuthorizationBytes)
	})
}
