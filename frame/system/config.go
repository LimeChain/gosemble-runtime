package system

import (
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/primitives/io"
	"github.com/LimeChain/gosemble/primitives/types"
)

type Config struct {
	Storage        io.Storage
	BlockHashCount types.BlockHashCount
	BlockWeights   types.BlockWeights
	BlockLength    types.BlockLength
	DbWeight       types.RuntimeDbWeight
	Version        *types.RuntimeVersion
	MaxConsumers   sc.U32
}

func NewConfig(
	storage io.Storage,
	blockHashCount types.BlockHashCount,
	blockWeights types.BlockWeights,
	blockLength types.BlockLength,
	dbWeight types.RuntimeDbWeight,
	version *types.RuntimeVersion,
	maxConsumers sc.U32,
) *Config {
	return &Config{
		storage,
		blockHashCount,
		blockWeights,
		blockLength,
		dbWeight,
		version,
		maxConsumers,
	}
}
