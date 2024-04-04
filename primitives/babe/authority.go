package babe

import (
	"bytes"

	sc "github.com/LimeChain/goscale"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

// A Babe authority identifier. Necessarily equivalent to the schnorrkel public key used in
// the main Babe module. If that ever changes, then this must, too.
type AuthorityId = primitives.Sr25519PublicKey

// The weight of an authority.
// We use a unique name for the weight to avoid conflicts with other
// `Weight` types, since the metadata isn't able to disambiguate.
type BabeAuthorityWeight = sc.U64

type Authority struct {
	Key    AuthorityId
	Weight BabeAuthorityWeight
}

func (a Authority) Encode(buffer *bytes.Buffer) error {
	return sc.EncodeEach(buffer,
		a.Key,
		a.Weight,
	)
}

func (a Authority) Bytes() []byte {
	return sc.EncodedBytes(a)
}

func DecodeAuthority(buffer *bytes.Buffer) (Authority, error) {
	accountId, err := primitives.DecodeSr25519PublicKey(buffer)
	if err != nil {
		return Authority{}, err
	}

	weight, err := sc.DecodeU64(buffer)
	if err != nil {
		return Authority{}, err
	}

	return Authority{
		Key:    accountId,
		Weight: weight,
	}, nil
}
