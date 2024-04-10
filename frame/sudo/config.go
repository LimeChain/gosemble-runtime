package sudo

import (
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

type Config struct {
	DbWeight       primitives.RuntimeDbWeight
	EventDepositor primitives.EventDepositor
}

func NewConfig(dbWeight primitives.RuntimeDbWeight, eventDepositor primitives.EventDepositor) Config {
	return Config{
		dbWeight,
		eventDepositor,
	}
}
