package types

import (
	"bytes"

	sc "github.com/LimeChain/goscale"
)

type Validator struct {
	AccountId   AccountId
	AuthorityId Sr25519PublicKey
}

func (v Validator) Encode(buffer *bytes.Buffer) error {
	return sc.EncodeEach(buffer, v.AccountId, v.AuthorityId)
}

func DecodeValidator(buffer *bytes.Buffer) (Validator, error) {
	accountId, err := DecodeAccountId(buffer)
	if err != nil {
		return Validator{}, err
	}
	authorityId, err := DecodeSr25519PublicKey(buffer)
	if err != nil {
		return Validator{}, err
	}

	return Validator{
		AccountId:   accountId,
		AuthorityId: authorityId,
	}, nil
}

func (v Validator) Bytes() []byte {
	return sc.EncodedBytes(v)
}
