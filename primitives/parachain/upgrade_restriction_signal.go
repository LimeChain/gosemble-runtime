package parachain

import (
	"bytes"
	"errors"

	sc "github.com/LimeChain/goscale"
)

const (
	UpgradeRestrictionSignalPresent sc.U8 = iota
)

func DecodeUpgradeRestrictionSignal(buffer *bytes.Buffer) (sc.U8, error) {
	value, err := sc.DecodeU8(buffer)
	if err != nil {
		return 0, err
	}

	switch value {
	case UpgradeRestrictionSignalPresent:
		return UpgradeRestrictionSignalPresent, nil
	default:
		return 0, errors.New("invalid UpgradeRestrictionSignal")
	}
}

func DecodeOptionUpgradeRestrictionSignal(buffer *bytes.Buffer) (sc.Option[sc.U8], error) {
	return sc.DecodeOptionWith(buffer, DecodeUpgradeRestrictionSignal)
}
