package main

import (
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/primitives/types"
	ctypes "github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

// TODO: switch to Gosemble types

var (
	aliceAddress, _     = ctypes.NewMultiAddressFromHexAccountID("0xd43593c715fdd31c61141abd04a99fd6822c8558854ccde39a5684e7a56da27d")
	aliceAccountIdBytes = aliceAddress.AsID.ToBytes()
	aliceAccountId, _   = types.NewAccountId(sc.BytesToSequenceU8(aliceAccountIdBytes)...)

	bobAddress, _     = ctypes.NewMultiAddressFromHexAccountID("0x90b5ab205c6974c9ea841be688864633dc9ca8a357843eeacf2314649965fe22")
	bobAccountIdBytes = bobAddress.AsID.ToBytes()
)

var (
	existentialAmount     = int64(BalancesExistentialDeposit.ToBigInt().Int64())
	existentialMultiplier = int64(10)
)
