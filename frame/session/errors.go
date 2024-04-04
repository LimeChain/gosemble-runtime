package session

import (
	sc "github.com/LimeChain/goscale"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

// Session module errors.
const (
	ErrorInvalidProof sc.U8 = iota
	ErrorNoAssociatedValidatorId
	ErrorDuplicatedKey
	ErrorNoKeys
	ErrorNoAccount
)

func NewDispatchErrorInvalidProof(moduleId sc.U8) primitives.DispatchError {
	return primitives.NewDispatchErrorModule(primitives.CustomModuleError{
		Index:   moduleId,
		Err:     sc.U32(ErrorInvalidProof),
		Message: sc.NewOption[sc.Str](nil),
	})
}

func NewDispatchErrorNoAssociatedValidatorId(moduleId sc.U8) primitives.DispatchError {
	return primitives.NewDispatchErrorModule(primitives.CustomModuleError{
		Index:   moduleId,
		Err:     sc.U32(ErrorNoAssociatedValidatorId),
		Message: sc.NewOption[sc.Str](nil),
	})
}

func NewDispatchErrorDuplicatedKey(moduleId sc.U8) primitives.DispatchError {
	return primitives.NewDispatchErrorModule(primitives.CustomModuleError{
		Index:   moduleId,
		Err:     sc.U32(ErrorDuplicatedKey),
		Message: sc.NewOption[sc.Str](nil),
	})
}

func NewDispatchErrorNoKeys(moduleId sc.U8) primitives.DispatchError {
	return primitives.NewDispatchErrorModule(primitives.CustomModuleError{
		Index:   moduleId,
		Err:     sc.U32(ErrorNoKeys),
		Message: sc.NewOption[sc.Str](nil),
	})
}

func NewDispatchErrorNoAccount(moduleId sc.U8) primitives.DispatchError {
	return primitives.NewDispatchErrorModule(primitives.CustomModuleError{
		Index:   moduleId,
		Err:     sc.U32(ErrorNoAccount),
		Message: sc.NewOption[sc.Str](nil),
	})
}
