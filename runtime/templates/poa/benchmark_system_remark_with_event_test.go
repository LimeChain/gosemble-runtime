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

func BenchmarkSystemRemarkWithEvent(b *testing.B) {
	size, err := benchmarking.NewLinear("size", 0, uint32(blockLength.Max.Normal))
	assert.NoError(b, err)

	benchmarking.RunDispatchCall(b, "../frame/system/call_remark_with_event_weight.go", func(i *benchmarking.Instance) {
		message := make([]byte, sc.U32(size.Value()))

		err := i.ExecuteExtrinsic(
			"System.remark_with_event",
			primitives.NewRawOriginSigned(aliceAccountId),
			message,
		)

		assert.NoError(b, err)

		buffer := &bytes.Buffer{}
		buffer.Write((*i.Storage()).Get(append(keySystemHash, keyEventCountHash...)))
		storageEventCount, err := sc.DecodeU32(buffer)
		assert.NoError(b, err)
		assert.Equal(b, sc.U32(1), storageEventCount)

		buffer.Reset()
		buffer.Write((*i.Storage()).Get(append(keySystemHash, keyEventsHash...)))

		decodedCount, err := sc.DecodeCompact[sc.U32](buffer)
		assert.NoError(b, err)
		assert.Equal(b, decodedCount.Number, storageEventCount)

		assertEmittedSystemEvent(b, system.EventRemarked, buffer)
	}, size)
}
