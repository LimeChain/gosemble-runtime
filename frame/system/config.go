package system

import (
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/primitives/types"
)

type Config struct {
	BlockHashCount types.BlockHashCount
	BlockWeights   types.BlockWeights
	BlockLength    types.BlockLength
	DbWeight       types.RuntimeDbWeight
	Version        *types.RuntimeVersion
	MaxConsumers   sc.U32
}

func NewConfig(
	blockHashCount types.BlockHashCount,
	blockWeights types.BlockWeights,
	blockLength types.BlockLength,
	dbWeight types.RuntimeDbWeight,
	version *types.RuntimeVersion,
	maxConsumers sc.U32,
) *Config {
	return &Config{
		blockHashCount,
		blockWeights,
		blockLength,
		dbWeight,
		version,
		maxConsumers,
	}
}
