package babe

import (
	sc "github.com/LimeChain/goscale"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

const (
	// An equivocation proof provided as part of an equivocation report is invalid.
	ErrorInvalidEquivocationProof sc.U8 = iota
	// A key ownership proof provided as part of an equivocation report is invalid.
	ErrorInvalidKeyOwnershipProof
	// A given equivocation report is valid but already previously reported.
	ErrorDuplicateOffenceReport
	// Submitted configuration is invalid.
	ErrorInvalidConfiguration
)

func NewDispatchErrorInvalidEquivocationProof(moduleId sc.U8) primitives.DispatchError {
	return primitives.NewDispatchErrorModule(primitives.CustomModuleError{
		Index:   moduleId,
		Err:     sc.U32(ErrorInvalidEquivocationProof),
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

func NewDispatchErrorDuplicateOffenceReport(moduleId sc.U8) primitives.DispatchError {
	return primitives.NewDispatchErrorModule(primitives.CustomModuleError{
		Index:   moduleId,
		Err:     sc.U32(ErrorDuplicateOffenceReport),
		Message: sc.NewOption[sc.Str](nil),
	})
}

func NewDispatchErrorInvalidConfiguration(moduleId sc.U8) primitives.DispatchError {
	return primitives.NewDispatchErrorModule(primitives.CustomModuleError{
		Index:   moduleId,
		Err:     sc.U32(ErrorInvalidConfiguration),
		Message: sc.NewOption[sc.Str](nil),
	})
}
