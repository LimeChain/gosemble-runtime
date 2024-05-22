package parachain

import (
	"bytes"
	sc "github.com/LimeChain/goscale"
)

type UnincludedSegment struct {
	Ancestors sc.Sequence[Ancestor]
}

func (us UnincludedSegment) Encode(buffer *bytes.Buffer) error {
	return sc.EncodeEach(buffer, us.Ancestors)
}

func DecodeUnincludedSegment(buffer *bytes.Buffer) (UnincludedSegment, error) {
	ancestors, err := sc.DecodeSequenceWith(buffer, DecodeAncestor)
	if err != nil {
		return UnincludedSegment{}, err
	}

	return UnincludedSegment{Ancestors: ancestors}, nil
}

func (us UnincludedSegment) Bytes() []byte {
	return sc.EncodedBytes(us)
}
