package types

import sc "github.com/LimeChain/goscale"

type DisabledValidators interface {
	IsDisabled(index sc.U32) bool
	Validators() sc.Sequence[sc.U32]
}
