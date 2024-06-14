package constants

import (
	sc "github.com/LimeChain/goscale"
)

const (
	MilliCents        = Cents / 1000
	Cents             = Dollar / 100
	Dollar            = Units
	Units      uint64 = 10_000_000_000
)

var (
	Zero       = sc.NewU128(0)
	One        = sc.NewU128(1)
	DefaultTip = Zero
)
