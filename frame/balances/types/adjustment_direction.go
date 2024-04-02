package types

import (
	"bytes"
	"errors"

	sc "github.com/LimeChain/goscale"
)

type AdjustmentDirection = sc.U8

const (
	AdjustmentDirectionIncrease AdjustmentDirection = iota
	AdjustmentDirectionDecrease
)

var (
	errInvalidAdjustmentDirectionType = errors.New("invalid adjustment direction type")
)

func DecodeAdjustmentDirection(buffer *bytes.Buffer) (sc.U8, error) {
	value, err := sc.DecodeU8(buffer)
	if err != nil {
		return sc.U8(0), err
	}

	switch value {
	case AdjustmentDirectionIncrease, AdjustmentDirectionDecrease:
		return value, nil
	default:
		return sc.U8(0), errInvalidAdjustmentDirectionType
	}
}
