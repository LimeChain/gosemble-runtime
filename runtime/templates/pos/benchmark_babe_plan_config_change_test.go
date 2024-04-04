package main

import (
	"testing"

	"github.com/ChainSafe/gossamer/pkg/scale"
	"github.com/LimeChain/gosemble/benchmarking"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/LimeChain/gosemble/testhelpers"
	"github.com/stretchr/testify/assert"
)

func BenchmarkBabePlanConfigChange(b *testing.B) {
	benchmarking.RunDispatchCall(b, "../../../frame/babe/call_plan_config_change_weight.go", func(i *benchmarking.Instance) {
		versionedNextConfigData := NextConfigDataV1{
			V1:             1,
			C1:             3,
			C2:             5,
			SecondarySlots: 2,
		}

		err := i.ExecuteExtrinsic(
			"Babe.plan_config_change",
			primitives.NewRawOriginNone(),
			versionedNextConfigData,
		)

		assert.NoError(b, err)

		configData := &NextConfigDataV1{}
		scale.Unmarshal((*i.Storage()).Get(append(testhelpers.KeyBabeHash, testhelpers.KeyPendingEpochConfigChangeHash...)), configData)
		assert.Equal(b, versionedNextConfigData, *configData)
	})
}
