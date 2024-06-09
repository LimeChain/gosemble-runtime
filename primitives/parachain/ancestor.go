package parachain

import (
	"bytes"
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/primitives/types"
)

// / Ancestor of the block being currently executed, not yet included
// / into the relay chain.
type Ancestor struct {
	UsedBandwidth         UsedBandwidth
	ParaHeadHash          sc.Option[types.H256]
	ConsumedGoAheadSignal sc.Option[sc.U8]
}

func (a Ancestor) Encode(buffer *bytes.Buffer) error {
	return sc.EncodeEach(buffer,
		a.UsedBandwidth,
		a.ParaHeadHash,
		a.ConsumedGoAheadSignal,
	)
}

func DecodeAncestor(buffer *bytes.Buffer) (Ancestor, error) {
	ub, err := DecodeUsedBandwidth(buffer)
	if err != nil {
		return Ancestor{}, err
	}

	paraHeadHash, err := sc.DecodeOptionWith(buffer, types.DecodeH256)
	if err != nil {
		return Ancestor{}, err
	}

	consumedGoAheadSignal, err := sc.DecodeOptionWith(buffer, sc.DecodeU8)
	if err != nil {
		return Ancestor{}, err
	}

	return Ancestor{
		UsedBandwidth:         ub,
		ParaHeadHash:          paraHeadHash,
		ConsumedGoAheadSignal: consumedGoAheadSignal,
	}, nil
}

func (a Ancestor) Bytes() []byte {
	return sc.EncodedBytes(a)
}
