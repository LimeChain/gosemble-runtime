package parachain

import (
	sc "github.com/LimeChain/goscale"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

var (
	ConsensusTypeId = [4]byte{'R', 'S', 'P', 'R'}
)

func NewDigestRelayParentStorageRoot(storageRoot primitives.H256, number RelayChainBlockNumber) primitives.DigestItem {
	msg := sc.BytesToSequenceU8(sc.NewVaryingData(storageRoot, number).Bytes())

	return primitives.NewDigestItemConsensusMessage(
		sc.BytesToFixedSequenceU8(ConsensusTypeId[:]),
		msg,
	)
}
