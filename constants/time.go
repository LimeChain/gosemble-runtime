package constants

import (
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/primitives/types"
)

// Since BABE is probabilistic this is the average expected block time that
// we are targeting. Blocks will be produced at a minimum duration defined
// by `SlotDuration`, but some slots will not be allocated to any
// authority and hence no block will be produced. We expect to have this
// block time on average following the defined slot duration and the value
// of `c` configured for BABE (where `1 - c` represents the probability of
// a slot being empty).
// This value is only used indirectly to define the unit constants below
// that are expressed in blocks. The rest of the code should use
// `SlotDuration` instead (like the Timestamp pallet for calculating the
// minimum period).
//
// If using BABE with secondary slots (default) then all of the slots will
// always be assigned, in which case `MillisecsPerBlock` and
// `SlotDuration` should have the same value.

type BlockNumber = sc.U32

const (
	MillisecsPerBlock sc.U64 = 3000
	SecsPerBlock      sc.U64 = MillisecsPerBlock / 1000

	// Currently it is not possible to change the slot duration after the chain has started.
	// Attempting to do so will brick block production.
	SlotDuration = MillisecsPerBlock

	// These time units are defined in number of blocks.
	Minutes BlockNumber = 60 / BlockNumber(SecsPerBlock)
	Hours   BlockNumber = Minutes * 60
	Days    BlockNumber = Hours * 24

	// Currently it is not possible to change the epoch duration after the chain has started.
	// Attempting to do so will brick block production.
	EpochDurationInBlocks BlockNumber = 10 * Minutes
	SlotFillRate                      = float64(MillisecsPerBlock) / float64(SlotDuration)
	EpochDurationInSlots              = sc.U64(float64(EpochDurationInBlocks) * SlotFillRate)
)

// 1 in 4 blocks (on average, not counting collisions) will be primary BABE blocks.
var PrimaryProbability = types.RationalValue{Numerator: 1, Denominator: 4}
