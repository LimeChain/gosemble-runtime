package grandpa

import (
	sc "github.com/LimeChain/goscale"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

const (
	ErrorPauseFailed sc.U8 = iota
	ErrorResumeFailed
	ErrorChangePending
	ErrorTooSoon
	ErrorInvalidKeyOwnershipProof
	ErrorInvalidEquivocationProof
	ErrorDuplicateOffenceReport
)

func NewDispatchErrorPauseFailed(moduleId sc.U8) primitives.DispatchError {
	return primitives.NewDispatchErrorModule(primitives.CustomModuleError{
		Index:   moduleId,
		Err:     sc.U32(ErrorPauseFailed),
		Message: sc.NewOption[sc.Str](nil),
	})
}

func NewDispatchErrorResumeFailed(moduleId sc.U8) primitives.DispatchError {
	return primitives.NewDispatchErrorModule(primitives.CustomModuleError{
		Index:   moduleId,
		Err:     sc.U32(ErrorResumeFailed),
		Message: sc.NewOption[sc.Str](nil),
	})
}

func NewDispatchErrorChangePending(moduleId sc.U8) primitives.DispatchError {
	return primitives.NewDispatchErrorModule(primitives.CustomModuleError{
		Index:   moduleId,
		Err:     sc.U32(ErrorChangePending),
		Message: sc.NewOption[sc.Str](nil),
	})
}

func NewDispatchErrorTooSoon(moduleId sc.U8) primitives.DispatchError {
	return primitives.NewDispatchErrorModule(primitives.CustomModuleError{
		Index:   moduleId,
		Err:     sc.U32(ErrorTooSoon),
		Message: sc.NewOption[sc.Str](nil),
	})
}

func NewDispatchErrorInvalidKeyOwnershipProof(moduleId sc.U8) primitives.DispatchError {
	return primitives.NewDispatchErrorModule(primitives.CustomModuleError{
		Index:   moduleId,
		Err:     sc.U32(ErrorInvalidKeyOwnershipProof),
		Message: sc.NewOption[sc.Str](nil),
	})
}

func NewDispatchErrorInvalidEquivocationProof(moduleId sc.U8) primitives.DispatchError {
	return primitives.NewDispatchErrorModule(primitives.CustomModuleError{
		Index:   moduleId,
		Err:     sc.U32(ErrorInvalidEquivocationProof),
		Message: sc.NewOption[sc.Str](nil),
	})
}

func NewDispatchErrorDuplicateOffenceReport(moduleId sc.U8) primitives.DispatchError {
	return primitives.NewDispatchErrorModule(primitives.CustomModuleError{
		Index:   moduleId,
		Err:     sc.U32(ErrorDuplicateOffenceReport),
		Message: sc.NewOption[sc.Str](nil),
	})
}
