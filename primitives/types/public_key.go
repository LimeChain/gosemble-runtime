package types

import (
	"errors"
	"fmt"

	sr25519 "github.com/LimeChain/go-schnorrkel"
	sc "github.com/LimeChain/goscale"
	"github.com/gtank/merlin"
)

type PublicKey interface {
	sc.Encodable
	SignatureType() sc.U8
}

type PubKey struct {
	key *sr25519.PublicKey
}

func NewPublicKey(in []byte) (*PubKey, error) {
	if len(in) != sr25519PublicKeyLength {
		return nil, errors.New("cannot create public key: input is not 32 bytes")
	}

	buf := [sr25519PublicKeyLength]byte{}
	copy(buf[:], in)

	sr25519Key, err := sr25519.NewPublicKey(buf)
	if err != nil {
		return nil, fmt.Errorf("creating sr25519 public key: %w", err)
	}

	return &PubKey{key: sr25519Key}, nil
}

func AttachInput(output [sr25519VRFOutputLength]byte, pub *PubKey, t *merlin.Transcript) (vrfInOut *sr25519.VrfInOut, err error) {
	out, err := sr25519.NewOutput(output)
	if err != nil {
		return nil, fmt.Errorf("creating sr25519 output: %w", err)
	}

	vrfInOut, err = out.AttachInput(pub.key, t)
	if err != nil {
		return nil, fmt.Errorf("attaching input: %w", err)
	}

	return vrfInOut, nil
}
