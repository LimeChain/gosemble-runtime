package main

import (
	"bytes"
	"os"
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/benchmarking"
	"github.com/LimeChain/gosemble/frame/system"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/LimeChain/gosemble/testhelpers"
	"github.com/stretchr/testify/assert"
)

func BenchmarkSystemSetCode(b *testing.B) {
	benchmarking.RunDispatchCall(b, "../../../frame/system/call_set_code_weight.go", func(i *benchmarking.Instance) {
		codeSpecVersion101, err := os.ReadFile(testhelpers.RuntimeWasmSpecVersion101)
		assert.NoError(b, err)

		err = i.ExecuteExtrinsic(
			"System.set_code",
			primitives.NewRawOriginRoot(),
			codeSpecVersion101,
		)

		assert.NoError(b, err)

		buffer := &bytes.Buffer{}

		testhelpers.AssertStorageSystemEventCount(b, i.Storage(), uint32(1))

		buffer.Write((*i.Storage()).Get(append(testhelpers.KeySystemHash, testhelpers.KeyEventsHash...)))
		decodedCount, err := sc.DecodeCompact[sc.U32](buffer)
		assert.NoError(b, err)
		assert.Equal(b, uint32(decodedCount.Number.(sc.U32)), uint32(1))

		testhelpers.AssertEmittedSystemEvent(b, system.EventCodeUpdated, buffer)
	})
}
