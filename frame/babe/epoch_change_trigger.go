package babe

import (
	sc "github.com/LimeChain/goscale"
)

type EpochChangeTrigger interface {
	Trigger(now sc.U64)
}

// A type signifying to BABE that an external trigger
// for epoch changes (e.g. pallet-session) is used.
type ExternalTrigger struct{}

// Trigger an epoch change, if any should take place. This should be called
// during every block, after initialization is done.
func (t ExternalTrigger) Trigger(_ sc.U64) {
	// nothing - trigger is external.
}

// A type signifying to BABE that it should perform epoch changes
// with an internal trigger, recycling the same authorities forever.
type SameAuthoritiesForever struct {
	Babe Module
}

func (t SameAuthoritiesForever) Trigger(now sc.U64) {
	babeModule := t.Babe

	if babeModule.ShouldEpochChange(now) {
		authorities, err := babeModule.StorageAuthorities()
		if err != nil {
			return
		}
		nextAuthorities := authorities

		babeModule.EnactEpochChange(authorities, nextAuthorities, sc.NewOption[sc.U32](nil))
	}
}
