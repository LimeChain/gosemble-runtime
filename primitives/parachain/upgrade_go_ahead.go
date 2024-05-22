package parachain

import (
	"bytes"
	"errors"
	sc "github.com/LimeChain/goscale"
)

const (
	UpgradeGoAheadAbort sc.U8 = iota
	UpgradeGoAheadGoAhead
)

func DecodeUpgradeGoAhead(buffer *bytes.Buffer) (sc.U8, error) {
	value, err := sc.DecodeU8(buffer)
	if err != nil {
		return 0, err
	}

	switch value {
	case UpgradeGoAheadAbort:
		return UpgradeGoAheadAbort, nil
	case UpgradeGoAheadGoAhead:
		return UpgradeGoAheadGoAhead, nil
	default:
		return 0, errors.New("invalid UpgradeGoAhead")
	}
}

func DecodeOptionUpgradeGoAhead(buffer *bytes.Buffer) (sc.Option[sc.U8], error) {
	return sc.DecodeOptionWith(buffer, DecodeUpgradeGoAhead)
}
