package sudo

import (
	"github.com/LimeChain/gosemble/primitives/io"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

type Config struct {
	Storage        io.Storage
	DbWeight       primitives.RuntimeDbWeight
	EventDepositor primitives.EventDepositor
}

func NewConfig(storage io.Storage, dbWeight primitives.RuntimeDbWeight, eventDepositor primitives.EventDepositor) Config {
	return Config{
		storage,
		dbWeight,
		eventDepositor,
	}
}
