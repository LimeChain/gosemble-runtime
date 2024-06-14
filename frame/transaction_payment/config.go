package transaction_payment

import (
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/primitives/io"
	"github.com/LimeChain/gosemble/primitives/types"
)

type Config struct {
	Storage                  io.Storage
	OperationalFeeMultiplier sc.U8
	WeightToFee              types.WeightToFee
	LengthToFee              types.WeightToFee
	BlockWeights             types.BlockWeights
}

func NewConfig(storage io.Storage, operationalFeeMultiplier sc.U8, weightToFee, lengthToFee types.WeightToFee, blockWeights types.BlockWeights) *Config {
	return &Config{
		storage,
		operationalFeeMultiplier,
		weightToFee,
		lengthToFee,
		blockWeights,
	}
}
