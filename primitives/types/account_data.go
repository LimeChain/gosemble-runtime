package types

import (
	"bytes"

	sc "github.com/LimeChain/goscale"
)

type Balance = sc.U128

type AccountData struct {
	Free     Balance
	Reserved Balance
	Frozen   Balance
	Flags    ExtraFlags
}

func DefaultAccountData() AccountData {
	return AccountData{
		Flags: DefaultExtraFlags(),
	}
}

func (ad AccountData) Encode(buffer *bytes.Buffer) error {
	return sc.EncodeEach(buffer,
		ad.Free,
		ad.Reserved,
		ad.Frozen,
		ad.Flags.U128,
	)
}

func (ad AccountData) Bytes() []byte {
	return sc.EncodedBytes(ad)
}

func DecodeAccountData(buffer *bytes.Buffer) (AccountData, error) {
	free, err := sc.DecodeU128(buffer)
	if err != nil {
		return AccountData{}, err
	}
	reserved, err := sc.DecodeU128(buffer)
	if err != nil {
		return AccountData{}, err
	}
	frozen, err := sc.DecodeU128(buffer)
	if err != nil {
		return AccountData{}, err
	}
	flags, err := DecodeExtraFlags(buffer)
	if err != nil {
		return AccountData{}, err
	}
	return AccountData{
		Free:     free,
		Reserved: reserved,
		Frozen:   frozen,
		Flags:    flags,
	}, nil
}

func (ad AccountData) Total() sc.U128 {
	return sc.SaturatingAddU128(ad.Free, ad.Reserved)
}
