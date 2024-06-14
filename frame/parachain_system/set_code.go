package parachain_system

import (
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/hooks"
)

type SetCode struct {
	module Module
}

func NewSetCode(m Module) hooks.OnSetCode {
	return SetCode{m}
}

func (s SetCode) SetCode(code sc.Sequence[sc.U8]) error {
	return s.module.ScheduleCodeUpgrade(code)
}
