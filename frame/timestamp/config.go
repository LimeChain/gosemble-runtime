package timestamp

import (
	"github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/hooks"
	"github.com/LimeChain/gosemble/primitives/io"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

type Config struct {
	Storage        io.Storage
	OnTimestampSet hooks.OnTimestampSet[goscale.U64]
	DbWeight       primitives.RuntimeDbWeight
	MinimumPeriod  goscale.U64
}

func NewConfig(storage io.Storage, onTsSet hooks.OnTimestampSet[goscale.U64], dbWeight primitives.RuntimeDbWeight, minimumPeriod goscale.U64) *Config {
	return &Config{
		Storage:        storage,
		OnTimestampSet: onTsSet,
		DbWeight:       dbWeight,
		MinimumPeriod:  minimumPeriod,
	}
}
