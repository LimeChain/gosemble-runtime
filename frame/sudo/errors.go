package sudo

import (
	sc "github.com/LimeChain/goscale"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

// Sudo module errors.
const (
	ErrorRequireSudo sc.U8 = iota
)

func NewDispatchErrorRequireSudo(moduleId sc.U8) primitives.DispatchError {
	return primitives.NewDispatchErrorModule(primitives.CustomModuleError{
		Index:   moduleId,
		Err:     sc.U32(ErrorRequireSudo),
		Message: sc.NewOption[sc.Str](nil),
	})
}
