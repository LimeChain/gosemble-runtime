package types

import (
	"bytes"
	sc "github.com/LimeChain/goscale"
	"math/big"
)

var (
	FlagsNewLogic, _  = new(big.Int).SetString("80000000000000000000000000000000", 16)
	DefaultExtraFlags = ExtraFlags{sc.NewU128(FlagsNewLogic)}
)

type ExtraFlags struct {
	sc.U128
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
	newEf := currentEf.Or(currentEf, FlagsNewLogic)
	return ExtraFlags{sc.NewU128(newEf)}
}

func (ef ExtraFlags) IsNewLogic() bool {
	currentEf := ef.ToBigInt()
	currentEf = currentEf.And(currentEf, FlagsNewLogic)
	return currentEf.Cmp(FlagsNewLogic) == 0
}
