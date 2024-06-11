package types

import (
	"bytes"

	sc "github.com/LimeChain/goscale"
)

type Tuple2U64 struct {
	First  sc.U64
	Second sc.U64
}

func (t Tuple2U64) Encode(buffer *bytes.Buffer) error {
	return sc.EncodeEach(buffer,
		t.First,
		t.Second,
	)
}

func (t Tuple2U64) Bytes() []byte {
	return sc.EncodedBytes(t)
}

func DecodeTuple2U64(buffer *bytes.Buffer) (Tuple2U64, error) {
	first, err := sc.DecodeU64(buffer)
	if err != nil {
		return Tuple2U64{}, err
	}

	second, err := sc.DecodeU64(buffer)
	if err != nil {
		return Tuple2U64{}, err
	}

	return Tuple2U64{
		First:  first,
		Second: second,
	}, nil
}
