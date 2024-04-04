package types

import (
	"bytes"

	sc "github.com/LimeChain/goscale"
)

type AccountId Address32

func NewAccountId(values ...sc.U8) (AccountId, error) {
	address, err := NewAddress32(values...)
	if err != nil {
		return AccountId{}, err
	}

	return AccountId(address), nil
}

func NewAccountIdFromAddress32(address Address32) AccountId {
	return AccountId(address)
}

func (a AccountId) Encode(buffer *bytes.Buffer) error {
	return a.FixedSequence.Encode(buffer)
}

func DecodeAccountId(buffer *bytes.Buffer) (AccountId, error) {
	address, err := DecodeAddress32(buffer)
	if err != nil {
		return AccountId{}, err
	}

	return AccountId(address), nil
}

func (a AccountId) Bytes() []byte {
	return sc.EncodedBytes(a.FixedSequence)
}

func DecodeSequenceAccountId(buffer *bytes.Buffer) (sc.Sequence[AccountId], error) {
	return sc.DecodeSequenceWith[AccountId](buffer, DecodeAccountId)
}
