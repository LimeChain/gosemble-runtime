package sudo

import (
	sc "github.com/LimeChain/goscale"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_NewDispatchErrorRequireSudo(t *testing.T) {
	expect := primitives.NewDispatchErrorModule(primitives.CustomModuleError{
		Index:   moduleId,
		Err:     sc.U32(ErrorRequireSudo),
		Message: sc.NewOption[sc.Str](nil),
	})

	assert.Equal(t, expect, NewDispatchErrorRequireSudo(moduleId))
}
