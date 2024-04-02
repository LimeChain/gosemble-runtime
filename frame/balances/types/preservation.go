package types

import (
	"bytes"
	"errors"

	sc "github.com/LimeChain/goscale"
)

type Preservation = sc.U8

const (
	/// We don't care if the account gets killed by this operation.
	PreservationExpendable Preservation = iota
	/// The account may not be killed, but we don't care if the balance gets dusted.
	PreservationProtect
	/// The account may not be killed and our provider reference must remain (in the context of
	/// tokens, this means that the account may not be dusted).
	PreservationPreserve
)

var (
	errInvalidPreservationType = errors.New("invalid adjustment direction type")
)

// todo test
func DecodePreservation(buffer *bytes.Buffer) (Preservation, error) {
	value, err := sc.DecodeU8(buffer)
	if err != nil {
		return sc.U8(0), err
	}

	switch value {
	case PreservationExpendable, PreservationProtect, PreservationPreserve:
		return value, nil
	default:
		return sc.U8(0), errInvalidPreservationType
	}
}
