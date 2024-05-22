package parachain_system

import "github.com/LimeChain/gosemble/primitives/parachain"

type CheckAssociatedRelayNumber interface {
	CheckAssociatedRelayNumber(current parachain.RelayChainBlockNumber, previous parachain.RelayChainBlockNumber)
}

type RelayNumberStrictlyIncreases struct {
}

func (rnsi RelayNumberStrictlyIncreases) CheckAssociatedRelayNumber(current parachain.RelayChainBlockNumber, previous parachain.RelayChainBlockNumber) {
	if current <= previous {
		panic("Relay chain block number needs to strictly increase between Parachain blocks!")
	}
}
