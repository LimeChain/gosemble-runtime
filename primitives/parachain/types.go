package parachain

import (
	sc "github.com/LimeChain/goscale"
)

type RelayChainBlockNumber = sc.U32

type HeadData = sc.Sequence[sc.U8]
