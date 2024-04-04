package session

import (
	sc "github.com/LimeChain/goscale"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_NewDispatchErrorInvalidProof(t *testing.T) {
	expect := primitives.NewDispatchErrorModule(primitives.CustomModuleError{
		Index:   moduleId,
		Err:     sc.U32(ErrorInvalidProof),
		Message: sc.NewOption[sc.Str](nil),
	})

	assert.Equal(t, expect, NewDispatchErrorInvalidProof(moduleId))
}

func Test_NewDispatchErrorNoAssociatedValidatorId(t *testing.T) {
	expect := primitives.NewDispatchErrorModule(primitives.CustomModuleError{
		Index:   moduleId,
		Err:     sc.U32(ErrorNoAssociatedValidatorId),
		Message: sc.NewOption[sc.Str](nil),
	})

	assert.Equal(t, expect, NewDispatchErrorNoAssociatedValidatorId(moduleId))
}

func Test_NewDispatchErrorDuplicatedKey(t *testing.T) {
	expect := primitives.NewDispatchErrorModule(primitives.CustomModuleError{
		Index:   moduleId,
		Err:     sc.U32(ErrorDuplicatedKey),
		Message: sc.NewOption[sc.Str](nil),
	})

	assert.Equal(t, expect, NewDispatchErrorDuplicatedKey(moduleId))
}

func Test_NewDispatchErrorNoKeys(t *testing.T) {
	expect := primitives.NewDispatchErrorModule(primitives.CustomModuleError{
		Index:   moduleId,
		Err:     sc.U32(ErrorNoKeys),
		Message: sc.NewOption[sc.Str](nil),
	})

	assert.Equal(t, expect, NewDispatchErrorNoKeys(moduleId))
}

func Test_NewDispatchErrorNoAccount(t *testing.T) {
	expect := primitives.NewDispatchErrorModule(primitives.CustomModuleError{
		Index:   moduleId,
		Err:     sc.U32(ErrorNoAccount),
		Message: sc.NewOption[sc.Str](nil),
	})

	assert.Equal(t, expect, NewDispatchErrorNoAccount(moduleId))
}
