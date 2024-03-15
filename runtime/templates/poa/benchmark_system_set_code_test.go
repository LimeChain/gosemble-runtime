package main

import (
	"bytes"
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/benchmarking"
	"github.com/LimeChain/gosemble/frame/system"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
)

func BenchmarkSystemSetCode(b *testing.B) {
	benchmarking.RunDispatchCall(b, "../frame/system/call_set_code_weight.go", func(i *benchmarking.Instance) {
		err := i.ExecuteExtrinsic(
			"System.set_code",
			primitives.NewRawOriginRoot(),
			codeSpecVersion101,
		)

		assert.NoError(b, err)

		buffer := &bytes.Buffer{}

		assertStorageSystemEventCount(b, i.Storage(), uint32(1))

		buffer.Write((*i.Storage()).Get(append(keySystemHash, keyEventsHash...)))
		decodedCount, err := sc.DecodeCompact[sc.U32](buffer)
		assert.NoError(b, err)
		assert.Equal(b, uint32(decodedCount.Number.(sc.U32)), uint32(1))

		assertEmittedSystemEvent(b, system.EventCodeUpdated, buffer)
	})
}
