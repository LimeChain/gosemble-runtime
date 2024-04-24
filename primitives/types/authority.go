package types

import (
	"bytes"

	sc "github.com/LimeChain/goscale"
)

type Authority struct {
	Id     AccountId
	Weight sc.U64
}

func (a Authority) Encode(buffer *bytes.Buffer) error {
	return sc.EncodeEach(buffer,
		a.Id,
		a.Weight,
	)
}

func DecodeAuthority(buffer *bytes.Buffer) (Authority, error) {
	pk, err := DecodeAccountId(buffer)
	if err != nil {
		return Authority{}, err
	}
	weight, err := sc.DecodeU64(buffer)
	if err != nil {
		return Authority{}, err
	}
	return Authority{
		Id:     pk,
		Weight: weight,
	}, nil
}

func (a Authority) Bytes() []byte {
	return sc.EncodedBytes(a)
}

func DecodeAuthorityList(buffer *bytes.Buffer) (sc.Sequence[Authority], error) {
	return sc.DecodeSequenceWith(buffer, DecodeAuthority)
}

func AuthoritiesFrom(validators sc.Sequence[Validator]) sc.Sequence[Authority] {
	authorities := make(sc.Sequence[Authority], 0)
	for _, v := range validators {
		authorities = append(authorities, Authority{Id: AccountId(v.AuthorityId), Weight: 1}) // TODO check if weight is correct
	}
	return authorities
}
