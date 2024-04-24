package main

import (
	"bytes"
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/benchmarking"
	"github.com/LimeChain/gosemble/primitives/types"
	"github.com/LimeChain/gosemble/testhelpers"
	"github.com/stretchr/testify/assert"
)

func BenchmarkGrandpaNoteStalled(b *testing.B) {
	benchmarking.RunDispatchCall(b, "../../../frame/grandpa/call_note_stalled_weight.go", func(i *benchmarking.Instance) {
		delayInBlocks := sc.U64(1000)
		bestFinalizedBlock := sc.U64(1)

		err := i.ExecuteExtrinsic(
			"Grandpa.note_stalled",
			types.NewRawOriginRoot(),
			delayInBlocks,
			bestFinalizedBlock,
		)

		assert.NoError(b, err)

		buffer := &bytes.Buffer{}
		buffer.Write((*i.Storage()).Get(append(testhelpers.KeyGrandpaHash, testhelpers.KeyStalledHash...)))
		stalled, err := types.DecodeTuple2U64(buffer)
		assert.NoError(b, err)
		assert.Equal(b, delayInBlocks, stalled.First)
		assert.Equal(b, bestFinalizedBlock, stalled.Second)
	})
}
