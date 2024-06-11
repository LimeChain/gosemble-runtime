package babe

import sc "github.com/LimeChain/goscale"

// VRF output length for per-slot randomness.
const RandomnessLength = 32

// Randomness type required by BABE operations.
type Randomness = sc.FixedSequence[sc.U8]

func NewRandomness() Randomness {
	return Randomness(
		sc.NewFixedSequence[sc.U8](RandomnessLength, make([]sc.U8, RandomnessLength)...),
	)
}
