package types

import sc "github.com/LimeChain/goscale"

type Preservation sc.U8

const (
	PreservationExpendable Preservation = iota
	PreservationProtect
	PreservationPreserve
)
