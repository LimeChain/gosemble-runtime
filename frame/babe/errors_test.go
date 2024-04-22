package babe

import (
	"testing"

	sc "github.com/LimeChain/goscale"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
)

func Test_NewDispatchErrorInvalidEquivocationProof(t *testing.T) {
	assert.Equal(t,
		primitives.NewDispatchErrorModule(
			primitives.CustomModuleError{
				Index:   moduleId,
				Err:     sc.U32(0),
				Message: sc.NewOption[sc.Str](nil),
			},
		),
		NewDispatchErrorInvalidEquivocationProof(moduleId),
	)
}

func Test_NewDispatchErrorInvalidKeyOwnershipProof(t *testing.T) {
	assert.Equal(t,
		primitives.NewDispatchErrorModule(
			primitives.CustomModuleError{
				Index:   moduleId,
				Err:     sc.U32(1),
				Message: sc.NewOption[sc.Str](nil),
			},
		),
		NewDispatchErrorInvalidKeyOwnershipProof(moduleId),
	)
}

func Test_NewDispatchErrorDuplicateOffenceReport(t *testing.T) {
	assert.Equal(t,
		primitives.NewDispatchErrorModule(
			primitives.CustomModuleError{
				Index:   moduleId,
				Err:     sc.U32(2),
				Message: sc.NewOption[sc.Str](nil),
			},
		),
		NewDispatchErrorDuplicateOffenceReport(moduleId),
	)
}

func Test_NewDispatchErrorInvalidConfiguration(t *testing.T) {
	assert.Equal(t,
		primitives.NewDispatchErrorModule(
			primitives.CustomModuleError{
				Index:   moduleId,
				Err:     sc.U32(3),
				Message: sc.NewOption[sc.Str](nil),
			},
		),
		NewDispatchErrorInvalidConfiguration(moduleId),
	)
}
