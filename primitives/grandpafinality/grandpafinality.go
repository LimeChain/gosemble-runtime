package grandpafinality

import (
	"bytes"

	sc "github.com/LimeChain/goscale"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

type TargetHashAndNumber struct {
	// The target block's hash.
	TargetHash primitives.H256
	// The target block's number.
	TargetNumber sc.U64
}

func (c TargetHashAndNumber) Encode(buffer *bytes.Buffer) error {
	return sc.EncodeEach(buffer,
		c.TargetHash,
		c.TargetNumber,
	)
}

func (c TargetHashAndNumber) Bytes() []byte {
	return sc.EncodedBytes(c)
}

func DecodeTargetHashAndNumber(buffer *bytes.Buffer) (TargetHashAndNumber, error) {
	targetHash, err := primitives.DecodeH256(buffer)
	if err != nil {
		return TargetHashAndNumber{}, err
	}

	targetNumber, err := sc.DecodeU64(buffer)
	if err != nil {
		return TargetHashAndNumber{}, err
	}

	return TargetHashAndNumber{
		TargetHash:   targetHash,
		TargetNumber: targetNumber,
	}, nil
}

type Tuple2HashSignature struct {
	TargetHashAndNumber
	primitives.SignatureEd25519
}

func (t Tuple2HashSignature) Encode(buffer *bytes.Buffer) error {
	return sc.EncodeEach(buffer,
		t.TargetHashAndNumber,
		t.SignatureEd25519,
	)
}

func (t Tuple2HashSignature) Bytes() []byte {
	return sc.EncodedBytes(t)
}

func DecodeTuple2HashSignature(buffer *bytes.Buffer) (Tuple2HashSignature, error) {
	targetHashAndNumber, err := DecodeTargetHashAndNumber(buffer)
	if err != nil {
		return Tuple2HashSignature{}, err
	}

	signature, err := primitives.DecodeSignatureEd25519(buffer)
	if err != nil {
		return Tuple2HashSignature{}, err
	}

	return Tuple2HashSignature{
		TargetHashAndNumber: targetHashAndNumber,
		SignatureEd25519:    signature,
	}, nil
}

type Equivocation struct {
	// The round number equivocated in.
	RoundNumber sc.U64
	// The identity of the equivocator.
	Identity primitives.Ed25519PublicKey
	// The first vote in the equivocation.
	First Tuple2HashSignature
	// The second vote in the equivocation.
	Second Tuple2HashSignature
}

func (e Equivocation) Encode(buffer *bytes.Buffer) error {
	return sc.EncodeEach(buffer,
		e.RoundNumber,
		e.Identity,
		e.First,
		e.Second,
	)
}

func (e Equivocation) Bytes() []byte {
	return sc.EncodedBytes(e)
}

func DecodeEquivocation(buffer *bytes.Buffer) (Equivocation, error) {
	roundNumber, err := sc.DecodeU64(buffer)
	if err != nil {
		return Equivocation{}, err
	}

	identity, err := primitives.DecodeEd25519PublicKey(buffer)
	if err != nil {
		return Equivocation{}, err
	}

	first, err := DecodeTuple2HashSignature(buffer)
	if err != nil {
		return Equivocation{}, err
	}

	second, err := DecodeTuple2HashSignature(buffer)
	if err != nil {
		return Equivocation{}, err
	}

	return Equivocation{
		RoundNumber: roundNumber,
		Identity:    identity,
		First:       first,
		Second:      second,
	}, nil
}
