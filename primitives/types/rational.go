package types

import (
	"bytes"

	sc "github.com/LimeChain/goscale"
)

// The rational represents a value between 0 and 1.
type RationalValue struct {
	Numerator   sc.U64
	Denominator sc.U64
}

func (r RationalValue) Encode(buffer *bytes.Buffer) error {
	return sc.EncodeEach(buffer,
		r.Numerator,
		r.Denominator,
	)
}

func (r RationalValue) Bytes() []byte {
	return sc.EncodedBytes(r)
}

func DecodeRationalValue(buffer *bytes.Buffer) (RationalValue, error) {
	Numerator, err := sc.DecodeU64(buffer)
	if err != nil {
		return RationalValue{}, err
	}

	Denominator, err := sc.DecodeU64(buffer)
	if err != nil {
		return RationalValue{}, err
	}

	return RationalValue{
		Numerator:   Numerator,
		Denominator: Denominator,
	}, nil
}
