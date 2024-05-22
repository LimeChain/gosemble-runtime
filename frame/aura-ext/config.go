package aura_ext

import (
	"github.com/LimeChain/gosemble/primitives/io"
	"github.com/LimeChain/gosemble/primitives/types"
)

type Config struct {
	Storage  io.Storage
	DbWeight types.RuntimeDbWeight
}

func NewConfig(storage io.Storage, dbWeight types.RuntimeDbWeight) Config {
	return Config{
		storage,
		dbWeight,
	}
}
