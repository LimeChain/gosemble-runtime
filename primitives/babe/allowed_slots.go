package babe

import (
	"bytes"
	"errors"

	sc "github.com/LimeChain/goscale"
)

var (
	ErrInvalidAllowedSlots = errors.New("invalid 'AllowedSlots' type")
)

// Types of allowed slots.
const (
	// Only allow primary slots.
	PrimarySlots sc.U8 = iota
	// Allow primary and secondary plain slots.
	PrimaryAndSecondaryPlainSlots
	// Allow primary and secondary VRF slots.
	PrimaryAndSecondaryVRFSlots
)

type AllowedSlots struct {
	sc.VaryingData
}

func NewPrimarySlots() AllowedSlots {
	return AllowedSlots{sc.NewVaryingData(PrimarySlots)}
}

func NewPrimaryAndSecondaryPlainSlots() AllowedSlots {
	return AllowedSlots{sc.NewVaryingData(PrimaryAndSecondaryPlainSlots)}
}

func NewPrimaryAndSecondaryVRFSlots() AllowedSlots {
	return AllowedSlots{sc.NewVaryingData(PrimaryAndSecondaryVRFSlots)}
}

func DecodeAllowedSlots(buffer *bytes.Buffer) (AllowedSlots, error) {
	b, err := sc.DecodeU8(buffer)
	if err != nil {
		return AllowedSlots{}, err
	}

	switch b {
	case PrimarySlots:
		return NewPrimarySlots(), nil
	case PrimaryAndSecondaryPlainSlots:
		return NewPrimaryAndSecondaryPlainSlots(), nil
	case PrimaryAndSecondaryVRFSlots:
		return NewPrimaryAndSecondaryVRFSlots(), nil
	default:
		return AllowedSlots{}, ErrInvalidAllowedSlots
	}
}

func (a AllowedSlots) String() string {
	switch a.VaryingData[0] {
	case PrimarySlots:
		return "PrimarySlots"
	case PrimaryAndSecondaryPlainSlots:
		return "PrimaryAndSecondaryPlainSlots"
	case PrimaryAndSecondaryVRFSlots:
		return "PrimaryAndSecondaryVRFSlots"
	default:
		return "invalid representation of AllowedSlots"
	}
}
