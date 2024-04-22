package babe

import (
	"bytes"

	sc "github.com/LimeChain/goscale"
)

type SkippedEpoch struct {
	sc.U64
	SessionIndex
}

func (se SkippedEpoch) Encode(buffer *bytes.Buffer) error {
	return sc.EncodeEach(buffer,
		se.U64,
		se.SessionIndex,
	)
}

func (se SkippedEpoch) Bytes() []byte {
	return sc.EncodedBytes(se)
}
