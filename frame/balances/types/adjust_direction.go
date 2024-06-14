package types

import (
	"bytes"
	"errors"

	sc "github.com/LimeChain/goscale"
)

const (
	AdjustDirectionIncrease sc.U8 = iota
	AdjustDirectionDecrease
)

type AdjustDirection struct {
	sc.VaryingData
}

var (
	errInvalidAdjustDirectionType = errors.New("invalid adjust direction type")
)

func NewAdjustDirectionIncrease() AdjustDirection {
	return AdjustDirection{sc.NewVaryingData(AdjustDirectionIncrease)}
}

func NewAdjustDirectionDecrease() AdjustDirection {
	return AdjustDirection{sc.NewVaryingData(AdjustDirectionDecrease)}
}

func DecodeAdjustDirection(buffer *bytes.Buffer) (AdjustDirection, error) {
	value, err := sc.DecodeU8(buffer)
	if err != nil {
		return AdjustDirection{}, err
	}
	switch value {
	case AdjustDirectionIncrease:
		return NewAdjustDirectionIncrease(), nil
	case AdjustDirectionDecrease:
		return NewAdjustDirectionDecrease(), nil
	default:
		return AdjustDirection{}, errInvalidAdjustDirectionType
	}
}

func (ad AdjustDirection) IsIncrease() bool {
	return ad.VaryingData[0] == AdjustDirectionIncrease
}

func (ad AdjustDirection) IsDecrease() bool {
	return ad.VaryingData[0] == AdjustDirectionDecrease
}
