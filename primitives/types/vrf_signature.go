package types

import (
	"bytes"
	"errors"

	sc "github.com/LimeChain/goscale"
)

var (
	errInvalidOutputLength = errors.New("invalid output length")
	errInvalidProofLength  = errors.New("invalid proof length")
)

// VRF signature data
type VrfSignature struct {
	PreOutput sc.FixedSequence[sc.U8]
	Proof     sc.FixedSequence[sc.U8]
}

func NewVrfSignature(output, proof sc.FixedSequence[sc.U8]) (VrfSignature, error) {
	if len(output) != sr25519VRFOutputLength {
		return VrfSignature{}, errInvalidOutputLength
	}

	if len(proof) != sr25519VRFProofLength {
		return VrfSignature{}, errInvalidProofLength
	}

	return VrfSignature{
		PreOutput: output,
		Proof:     proof,
	}, nil
}

func (s VrfSignature) Encode(buffer *bytes.Buffer) error {
	return sc.EncodeEach(buffer,
		s.PreOutput,
		s.Proof,
	)
}

func (s VrfSignature) Bytes() []byte {
	return sc.EncodedBytes(s)
}

func DecodeVrfSignature(buffer *bytes.Buffer) (VrfSignature, error) {
	preOutput, err := sc.DecodeFixedSequence[sc.U8](sr25519VRFOutputLength, buffer)
	if err != nil {
		return VrfSignature{}, err
	}

	proof, err := sc.DecodeFixedSequence[sc.U8](sr25519VRFProofLength, buffer)
	if err != nil {
		return VrfSignature{}, err
	}

	return VrfSignature{
		PreOutput: preOutput,
		Proof:     proof,
	}, nil
}
