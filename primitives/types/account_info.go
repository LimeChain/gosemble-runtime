package types

import (
	"bytes"

	sc "github.com/LimeChain/goscale"
)

type RefCount = sc.U32

type AccountInfo struct {
	Nonce       AccountIndex
	Consumers   RefCount
	Providers   RefCount
	Sufficients RefCount
	Data        AccountData
}

func (ai AccountInfo) Encode(buffer *bytes.Buffer) error {
	return sc.EncodeEach(buffer,
		ai.Nonce,
		ai.Consumers,
		ai.Providers,
		ai.Sufficients,
		ai.Data,
	)
}

func (ai AccountInfo) Bytes() []byte {
	return sc.EncodedBytes(ai)
}

func DecodeAccountInfo(buffer *bytes.Buffer) (AccountInfo, error) {
	nonce, err := sc.DecodeU32(buffer)
	if err != nil {
		return AccountInfo{}, err
	}
	consumers, err := sc.DecodeU32(buffer)
	if err != nil {
		return AccountInfo{}, err
	}
	providers, err := sc.DecodeU32(buffer)
	if err != nil {
		return AccountInfo{}, err
	}
	sufficients, err := sc.DecodeU32(buffer)
	if err != nil {
		return AccountInfo{}, err
	}
	data, err := DecodeAccountData(buffer)
	if err != nil {
		return AccountInfo{}, err
	}
	return AccountInfo{
		Nonce:       nonce,
		Consumers:   consumers,
		Providers:   providers,
		Sufficients: sufficients,
		Data:        data,
	}, nil
}

// func (ai AccountInfo) Frozen(reasons Reasons) sc.U128 {
// 	switch reasons {
// 	case ReasonsAll:
// 		if ai.Data.MiscFrozen.Gt(ai.Data.FeeFrozen) {
// 			return ai.Data.MiscFrozen
// 		}
// 		return ai.Data.FeeFrozen
// 	case ReasonsMisc:
// 		return ai.Data.MiscFrozen
// 	case ReasonsFee:
// 		return ai.Data.MiscFrozen
// 	}

// 	return sc.NewU128(0)
// }
