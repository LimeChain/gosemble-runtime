package parachain_system

import (
	"github.com/LimeChain/gosemble/primitives/log"
	"github.com/LimeChain/gosemble/primitives/parachain"
)

type CheckAssociatedRelayNumber interface {
	CheckAssociatedRelayNumber(current parachain.RelayChainBlockNumber, previous parachain.RelayChainBlockNumber)
}

type RelayNumberStrictlyIncreases struct {
	logger log.Logger
}

func NewRelayNumberStrictlyIncreases(logger log.Logger) RelayNumberStrictlyIncreases {
	return RelayNumberStrictlyIncreases{
		logger: logger,
	}
}

func (rnsi RelayNumberStrictlyIncreases) CheckAssociatedRelayNumber(current parachain.RelayChainBlockNumber, previous parachain.RelayChainBlockNumber) {
	if current <= previous {
		rnsi.logger.Critical("relay chain block number needs to strictly increase between Parachain blocks!")
	}
}
