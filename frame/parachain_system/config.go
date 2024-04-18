package parachain_system

import (
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

type Config struct {
	DbWeight primitives.RuntimeDbWeight
}

func NewConfig(dbWeight primitives.RuntimeDbWeight) Config {
	return Config{
		DbWeight: dbWeight,
	}
}
