package parachain

import (
	"bytes"
	sc "github.com/LimeChain/goscale"
)

type AsyncBackingParams struct {
	MaxCandidateDepth  sc.U32
	AllowedAncestryLen sc.U32
}

func (abp AsyncBackingParams) Encode(buffer *bytes.Buffer) error {
	return sc.EncodeEach(buffer, abp.MaxCandidateDepth, abp.AllowedAncestryLen)
}

func DecodeAsyncBackingParams(buffer *bytes.Buffer) (AsyncBackingParams, error) {
	maxCandidateDepth, err := sc.DecodeU32(buffer)
	if err != nil {
		return AsyncBackingParams{}, err
	}

	allowedAncestryLen, err := sc.DecodeU32(buffer)
	if err != nil {
		return AsyncBackingParams{}, err
	}

	return AsyncBackingParams{
		MaxCandidateDepth:  maxCandidateDepth,
		AllowedAncestryLen: allowedAncestryLen,
	}, nil
}

func (abp AsyncBackingParams) Bytes() []byte {
	return sc.EncodedBytes(abp)
}
