package babe

import (
	"bytes"

	sc "github.com/LimeChain/goscale"
)

type EpochStartBlocks struct {
	Previous sc.U64
	Current  sc.U64
}

func (e EpochStartBlocks) Encode(buffer *bytes.Buffer) error {
	return sc.EncodeEach(buffer,
		e.Previous,
		e.Current,
	)
}

func (e EpochStartBlocks) Bytes() []byte {
	return sc.EncodedBytes(e)
}

func DecodeEpochStartBlocks(buffer *bytes.Buffer) (EpochStartBlocks, error) {
	previous, err := sc.DecodeU64(buffer)
	if err != nil {
		return EpochStartBlocks{}, err
	}

	current, err := sc.DecodeU64(buffer)
	if err != nil {
		return EpochStartBlocks{}, err
	}

	return EpochStartBlocks{
		Previous: previous,
		Current:  current,
	}, nil
}
