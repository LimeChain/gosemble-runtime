package types

import sc "github.com/LimeChain/goscale"

type Precision sc.U8

const (
	PrecisionExact Precision = iota
	PrecisionBestEffort
)
