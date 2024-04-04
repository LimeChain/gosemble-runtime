package babe

import (
	"encoding/binary"

	sc "github.com/LimeChain/goscale"
	babetypes "github.com/LimeChain/gosemble/primitives/babe"
	"github.com/gtank/merlin"
)

// Make VRF input suitable for BABE's randomness generation.
func makeVrfTranscript(randomness babetypes.Randomness, slot babetypes.Slot, epoch sc.U64) *merlin.Transcript {
	t := merlin.NewTranscript(string(EngineId[:]))
	AppendUint64(t, []byte("slot number"), uint64(slot))
	AppendUint64(t, []byte("current epoch"), uint64(epoch))
	t.AppendMessage([]byte("chain randomness"), sc.FixedSequenceU8ToBytes(randomness[:]))
	return t
}

// AppendUint64 appends a uint64 to the given transcript using the given label
func AppendUint64(t *merlin.Transcript, label []byte, n uint64) {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, n)
	t.AppendMessage(label, buf)
}
