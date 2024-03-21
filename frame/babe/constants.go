package babe

import sc "github.com/LimeChain/goscale"

// VRF output length for per-slot randomness.
const RandomnessLength = 32 // usize

// Randomness type required by BABE operations.
type Randomness = sc.FixedSequence[sc.U8]

type consts struct{}

func newConstants() *consts {
	return &consts{}
}
