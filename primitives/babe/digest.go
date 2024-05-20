package babe

import (
	"bytes"
	"errors"

	sc "github.com/LimeChain/goscale"
	// babetypes "github.com/LimeChain/gosemble/primitives/babe"
	"github.com/LimeChain/gosemble/primitives/types"
)

var (
	errInvalidPreDigest = errors.New("invalid 'PreDigest' type")
)

// A BABE pre-runtime digest. This contains all data required to validate a
// block and for the BABE runtime module. Slots can be assigned to a primary
// (VRF based) and to a secondary (slot number based).
const (
	_ sc.U8 = iota

	// A primary VRF-based slot assignment.
	Primary
	// A secondary deterministic slot assignment.
	SecondaryPlain
	// A secondary deterministic slot assignment with VRF outputs.
	SecondaryVRF
)

type PreDigest struct {
	sc.VaryingData
}

func NewPrimaryPreDigest(authorityIndex AuthorityIndex, slot Slot, vrfSignature types.VrfSignature) PreDigest {
	return PreDigest{sc.NewVaryingData(Primary, PrimaryPreDigest{
		AuthorityIndex: authorityIndex,
		Slot:           slot,
		VrfSignature:   vrfSignature,
	})}
}

func NewSecondaryPlainPreDigest(authorityIndex AuthorityIndex, slot Slot) PreDigest {
	return PreDigest{sc.NewVaryingData(SecondaryPlain, SecondaryPlainPreDigest{
		AuthorityIndex: authorityIndex,
		Slot:           slot,
	})}
}

func NewSecondaryVRFPreDigest(authorityIndex AuthorityIndex, slot Slot, vrfSignature types.VrfSignature) PreDigest {
	return PreDigest{sc.NewVaryingData(SecondaryVRF, SecondaryVRFPreDigest{
		AuthorityIndex: authorityIndex,
		Slot:           slot,
		VrfSignature:   vrfSignature,
	})}
}

func DecodePreDigest(buffer *bytes.Buffer) (PreDigest, error) {
	b, err := sc.DecodeU8(buffer)
	if err != nil {
		return PreDigest{}, err
	}

	switch b {
	case Primary:
		primaryPreDigest, err := DecodePrimaryPreDigest(buffer)
		if err != nil {
			return PreDigest{}, err
		}
		return PreDigest{sc.NewVaryingData(b, primaryPreDigest)}, nil
	case SecondaryPlain:
		secondaryPlainPreDigest, err := DecodeSecondaryPlainPreDigest(buffer)
		if err != nil {
			return PreDigest{}, err
		}
		return PreDigest{sc.NewVaryingData(b, secondaryPlainPreDigest)}, nil
	case SecondaryVRF:
		secondaryVRFPreDigest, err := DecodeSecondaryVRFPreDigest(buffer)
		if err != nil {
			return PreDigest{}, err
		}
		return PreDigest{sc.NewVaryingData(b, secondaryVRFPreDigest)}, nil
	default:
		return PreDigest{}, errInvalidPreDigest
	}
}

// Raw BABE primary slot assignment pre-digest.
type PrimaryPreDigest struct {
	AuthorityIndex AuthorityIndex
	Slot           Slot
	VrfSignature   types.VrfSignature
}

func (d PrimaryPreDigest) Encode(buffer *bytes.Buffer) error {
	return sc.EncodeEach(buffer,
		d.AuthorityIndex,
		d.Slot,
		d.VrfSignature,
	)
}

func (d PrimaryPreDigest) Bytes() []byte {
	return sc.EncodedBytes(d)
}

func DecodePrimaryPreDigest(buffer *bytes.Buffer) (PrimaryPreDigest, error) {
	authorityIndex, err := sc.DecodeU32(buffer)
	if err != nil {
		return PrimaryPreDigest{}, err
	}

	slot, err := sc.DecodeU64(buffer)
	if err != nil {
		return PrimaryPreDigest{}, err
	}

	vrfSignature, err := types.DecodeVrfSignature(buffer)
	if err != nil {
		return PrimaryPreDigest{}, err
	}

	return PrimaryPreDigest{
		AuthorityIndex: authorityIndex,
		Slot:           slot,
		VrfSignature:   vrfSignature,
	}, nil
}

// BABE secondary slot assignment pre-digest.
type SecondaryPlainPreDigest struct {
	// This is not strictly-speaking necessary, since the secondary slots
	// are assigned based on slot number and epoch randomness. But including
	// it makes things easier for higher-level users of the chain data to
	// be aware of the author of a secondary-slot block.
	AuthorityIndex AuthorityIndex
	Slot           Slot
}

func (d SecondaryPlainPreDigest) Encode(buffer *bytes.Buffer) error {
	return sc.EncodeEach(buffer,
		d.AuthorityIndex,
		d.Slot,
	)
}

func (d SecondaryPlainPreDigest) Bytes() []byte {
	return sc.EncodedBytes(d)
}

func DecodeSecondaryPlainPreDigest(buffer *bytes.Buffer) (SecondaryPlainPreDigest, error) {
	authorityIndex, err := sc.DecodeU32(buffer)
	if err != nil {
		return SecondaryPlainPreDigest{}, err
	}

	slot, err := sc.DecodeU64(buffer)
	if err != nil {
		return SecondaryPlainPreDigest{}, err
	}

	return SecondaryPlainPreDigest{
		AuthorityIndex: authorityIndex,
		Slot:           slot,
	}, nil
}

// BABE secondary deterministic slot assignment with VRF outputs.
type SecondaryVRFPreDigest struct {
	AuthorityIndex AuthorityIndex
	Slot           Slot
	VrfSignature   types.VrfSignature
}

func (d SecondaryVRFPreDigest) Encode(buffer *bytes.Buffer) error {
	return sc.EncodeEach(buffer,
		d.AuthorityIndex,
		d.Slot,
		d.VrfSignature,
	)
}

func (d SecondaryVRFPreDigest) Bytes() []byte {
	return sc.EncodedBytes(d)
}

func DecodeSecondaryVRFPreDigest(buffer *bytes.Buffer) (SecondaryVRFPreDigest, error) {
	authorityIndex, err := sc.DecodeU32(buffer)
	if err != nil {
		return SecondaryVRFPreDigest{}, err
	}

	slot, err := sc.DecodeU64(buffer)
	if err != nil {
		return SecondaryVRFPreDigest{}, err
	}

	vrfSignature, err := types.DecodeVrfSignature(buffer)
	if err != nil {
		return SecondaryVRFPreDigest{}, err
	}

	return SecondaryVRFPreDigest{
		AuthorityIndex: authorityIndex,
		Slot:           slot,
		VrfSignature:   vrfSignature,
	}, nil
}

// Returns the slot number of the pre digest.
func (d PreDigest) AuthorityIndex() (AuthorityIndex, error) {
	switch d.VaryingData[0] {
	case Primary:
		return d.VaryingData[1].(PrimaryPreDigest).AuthorityIndex, nil
	case SecondaryPlain:
		return d.VaryingData[1].(SecondaryPlainPreDigest).AuthorityIndex, nil
	case SecondaryVRF:
		return d.VaryingData[1].(SecondaryVRFPreDigest).AuthorityIndex, nil
	default:
		return 0, errInvalidPreDigest
	}
}

// Returns the slot of the pre digest.
func (d PreDigest) Slot() (Slot, error) {
	switch d.VaryingData[0] {
	case Primary:
		return d.VaryingData[1].(PrimaryPreDigest).Slot, nil
	case SecondaryPlain:
		return d.VaryingData[1].(SecondaryPlainPreDigest).Slot, nil
	case SecondaryVRF:
		return d.VaryingData[1].(SecondaryVRFPreDigest).Slot, nil
	default:
		return 0, errInvalidPreDigest
	}
}

// Returns true if this pre-digest is for a primary slot assignment.
func (d PreDigest) IsPrimary() bool {
	return d.VaryingData[0] == Primary
}

// Returns the VRF output and proof, if they exist.
func (d PreDigest) VrfSignature() (sc.Option[types.VrfSignature], error) {
	switch d.VaryingData[0] {
	case Primary:
		return sc.NewOption[types.VrfSignature](d.VaryingData[1].(PrimaryPreDigest).VrfSignature), nil
	case SecondaryPlain:
		return sc.NewOption[types.VrfSignature](nil), nil
	case SecondaryVRF:
		return sc.NewOption[types.VrfSignature](d.VaryingData[1].(SecondaryVRFPreDigest).VrfSignature), nil
	default:
		return sc.NewOption[types.VrfSignature](nil), errInvalidPreDigest
	}
}
