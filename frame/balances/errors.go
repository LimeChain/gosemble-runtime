package balances

import sc "github.com/LimeChain/goscale"

// Balances module errors.
const (
	ErrorVestingBalance sc.U8 = iota
	ErrorLiquidityRestrictions
	ErrorInsufficientBalance
	ErrorExistentialDeposit
	ErrorExpendability
	ErrorExistingVestingSchedule
	ErrorDeadAccount
	ErrorTooManyReserves
	ErrorTooManyHolds
	ErrorTooManyFreezes
	ErrorIssuanceDeactivated
	ErrorDeltaZero
)
