package types

import (
	"bytes"

	sc "github.com/LimeChain/goscale"
)

const (
	// PublicKeyLength is the expected public key length for sr25519.
	sr25519PublicKeyLength = 32
	// SeedLength is the expected seed length for sr25519.
	// sr25519SeedLength = 32
	// PrivateKeyLength is the expected private key length for sr25519.
	// sr25519PrivateKeyLength = 32
	// SignatureLength is the expected signature length for sr25519.
	sr25519SignatureLength = 64
	// VRFOutputLength is the expected VFR output length for sr25519.
	sr25519VRFOutputLength = 32
	// VRFProofLength is the expected VFR proof length for sr25519.
	sr25519VRFProofLength = 64
)

type SignatureSr25519 struct {
	sc.FixedSequence[sc.U8] // size 64
}

func NewSignatureSr25519(values ...sc.U8) SignatureSr25519 {
	return SignatureSr25519{sc.NewFixedSequence(sr25519SignatureLength, values...)}
}

func (s SignatureSr25519) Encode(buffer *bytes.Buffer) error {
	return s.FixedSequence.Encode(buffer)
}

func DecodeSignatureSr25519(buffer *bytes.Buffer) (SignatureSr25519, error) {
	s := SignatureSr25519{}
	seq, err := sc.DecodeFixedSequence[sc.U8](sr25519SignatureLength, buffer)
	if err != nil {
		return SignatureSr25519{}, nil
	}
	s.FixedSequence = seq
	return s, nil
}

func (s SignatureSr25519) Bytes() []byte {
	return sc.EncodedBytes(s)
}
