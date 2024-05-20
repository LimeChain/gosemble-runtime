package grandpa

import (
	"testing"

	sc "github.com/LimeChain/goscale"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
)

func Test_NewDispatchErrorPauseFailed(t *testing.T) {
	expected := primitives.NewDispatchErrorModule(primitives.CustomModuleError{
		Index:   moduleId,
		Err:     sc.U32(ErrorPauseFailed),
		Message: sc.NewOption[sc.Str](nil),
	})

	assert.Equal(t, expected, NewDispatchErrorPauseFailed(moduleId))
}

func Test_NewDispatchErrorResumeFailed(t *testing.T) {
	expected := primitives.NewDispatchErrorModule(primitives.CustomModuleError{
		Index:   moduleId,
		Err:     sc.U32(ErrorResumeFailed),
		Message: sc.NewOption[sc.Str](nil),
	})

	assert.Equal(t, expected, NewDispatchErrorResumeFailed(moduleId))
}

func Test_NewDispatchErrorChangePending(t *testing.T) {
	expected := primitives.NewDispatchErrorModule(primitives.CustomModuleError{
		Index:   moduleId,
		Err:     sc.U32(ErrorChangePending),
		Message: sc.NewOption[sc.Str](nil),
	})

	assert.Equal(t, expected, NewDispatchErrorChangePending(moduleId))
}

func Test_NewDispatchErrorTooSoon(t *testing.T) {
	expected := primitives.NewDispatchErrorModule(primitives.CustomModuleError{
		Index:   moduleId,
		Err:     sc.U32(ErrorTooSoon),
		Message: sc.NewOption[sc.Str](nil),
	})

	assert.Equal(t, expected, NewDispatchErrorTooSoon(moduleId))
}

func Test_NewDispatchErrorInvalidKeyOwnershipProof(t *testing.T) {
	expected := primitives.NewDispatchErrorModule(primitives.CustomModuleError{
		Index:   moduleId,
		Err:     sc.U32(ErrorInvalidKeyOwnershipProof),
		Message: sc.NewOption[sc.Str](nil),
	})

	assert.Equal(t, expected, NewDispatchErrorInvalidKeyOwnershipProof(moduleId))
}

func Test_NewDispatchErrorInvalidEquivocationProof(t *testing.T) {
	expected := primitives.NewDispatchErrorModule(primitives.CustomModuleError{
		Index:   moduleId,
		Err:     sc.U32(ErrorInvalidEquivocationProof),
		Message: sc.NewOption[sc.Str](nil),
	})

	assert.Equal(t, expected, NewDispatchErrorInvalidEquivocationProof(moduleId))
}

func Test_NewDispatchErrorDuplicateOffenceReport(t *testing.T) {
	expected := primitives.NewDispatchErrorModule(primitives.CustomModuleError{
		Index:   moduleId,
		Err:     sc.U32(ErrorDuplicateOffenceReport),
		Message: sc.NewOption[sc.Str](nil),
	})

	assert.Equal(t, expected, NewDispatchErrorDuplicateOffenceReport(moduleId))
}
