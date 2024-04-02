package types

import (
	"bytes"
	"math/big"

	sc "github.com/LimeChain/goscale"
)

var isNewLogic, _ = new(big.Int).SetString("80000000000000000000000000000000", 16)

type ExtraFlags struct {
	sc.U128
}

func DefaultExtraFlags() ExtraFlags {
	return ExtraFlags{sc.NewU128(isNewLogic)}
}

func (ef ExtraFlags) Encode(buffer *bytes.Buffer) error {
	return ef.U128.Encode(buffer)
}

func (ef ExtraFlags) Bytes() []byte {
	return ef.U128.Bytes()
}

func (ef ExtraFlags) OldLogic() ExtraFlags {
	return ef
}

func (ef ExtraFlags) SetNewLogic() ExtraFlags {
	currentEf := ef.ToBigInt()
	newEf := currentEf.Or(currentEf, isNewLogic)
	return ExtraFlags{sc.NewU128(newEf)}
}

func (ef ExtraFlags) IsNewLogic() bool {
	currentEf := ef.ToBigInt()
	currentEf = currentEf.And(currentEf, isNewLogic)
	return currentEf.Cmp(isNewLogic) == 0
}

func DecodeExtraFlags(buffer *bytes.Buffer) (ExtraFlags, error) {
	flags, err := sc.DecodeU128(buffer)
	if err != nil {
		return ExtraFlags{}, err
	}
	return ExtraFlags{flags}, nil
}

// todo delete - old impl

// type ExtraFlags sc.U128

// func DefaultExtraFlags() ExtraFlags {
// 	return ExtraFlags(sc.NewU128(isNewLogic))
// }

// func (ef ExtraFlags) Encode(buffer *bytes.Buffer) error {
// 	return sc.U128(ef).Encode(buffer)
// }

// func (ef ExtraFlags) Bytes() []byte {
// 	return sc.U128(ef).Bytes()
// }

// func (ef ExtraFlags) OldLogic() ExtraFlags {
// 	return ef
// }

// func (ef ExtraFlags) SetNewLogic() ExtraFlags {
// 	currentEf := sc.U128(ef).ToBigInt()
// 	newEf := currentEf.Or(currentEf, isNewLogic)
// 	return ExtraFlags(sc.NewU128(newEf))
// }

// func (ef ExtraFlags) IsNewLogic() bool {
// 	currentEf := sc.U128(ef).ToBigInt()
// 	currentEf = currentEf.And(currentEf, isNewLogic)
// 	return currentEf.Cmp(isNewLogic) == 0
// }

// var isNewLogic, _ = sc.NewU128FromString("170141183460469231731687303715884105728")
