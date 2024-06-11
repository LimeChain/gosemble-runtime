package babe

import (
	"bytes"
	"errors"

	sc "github.com/LimeChain/goscale"
	babetypes "github.com/LimeChain/gosemble/primitives/babe"
)

const (
	NextConfigDescriptorV1 sc.U8 = 1
)

var (
	errInvalidNextConfigDescriptor = errors.New("invalid 'NextConfigDescriptor' type")
)

type NextConfigDescriptor struct {
	V1 babetypes.EpochConfiguration
}

func (d NextConfigDescriptor) Encode(buffer *bytes.Buffer) error {
	vd := sc.NewVaryingData(NextConfigDescriptorV1, d.V1)
	return vd.Encode(buffer)
}

func (d NextConfigDescriptor) Bytes() []byte {
	return sc.EncodedBytes(d)
}

func DecodeNextConfigDescriptor(buffer *bytes.Buffer) (NextConfigDescriptor, error) {
	index, err := sc.DecodeU8(buffer)
	if err != nil {
		return NextConfigDescriptor{}, err
	}

	if index != NextConfigDescriptorV1 {
		return NextConfigDescriptor{}, errInvalidNextConfigDescriptor
	}

	config, err := babetypes.DecodeEpochConfiguration(buffer)
	if err != nil {
		return NextConfigDescriptor{}, err
	}

	return NextConfigDescriptor{V1: config}, nil
}
