package types

import (
	"fmt"
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/mocks"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
)

var (
	apiModuleOneName = "apiModuleOne"
	apiModuleTwoName = "apiModuleTwo"

	apiItemOne = primitives.ApiItem{
		Name:    sc.BytesToFixedSequenceU8(sc.Str(apiModuleOneName).Bytes()),
		Version: 4,
	}
	apiItemTwo = primitives.ApiItem{
		Name:    sc.BytesToFixedSequenceU8(sc.Str(apiModuleTwoName).Bytes()),
		Version: 2,
	}
)

var (
	mockApiModuleOne *mocks.ApiModule
	mockApiModuleTwo *mocks.ApiModule
)

func Test_RuntimeApi_New(t *testing.T) {
	target := setupRuntimeApi()

	expect := RuntimeApi{
		apis: []primitives.ApiModule{
			mockApiModuleOne,
			mockApiModuleTwo,
		},
		logger: logger,
	}

	assert.Equal(t, expect, target)
}

func Test_RuntimeApi_Items(t *testing.T) {
	target := setupRuntimeApi()
	expect := sc.Sequence[primitives.ApiItem]{
		apiItemOne,
		apiItemTwo,
	}

	mockApiModuleOne.On("Item").Return(apiItemOne)
	mockApiModuleTwo.On("Item").Return(apiItemTwo)

	result := target.Items()

	assert.Equal(t, expect, result)
	mockApiModuleOne.AssertCalled(t, "Item")
	mockApiModuleTwo.AssertCalled(t, "Item")
}

func Test_RuntimeApi_Module(t *testing.T) {
	target := setupRuntimeApi()

	mockApiModuleOne.On("Name").Return(apiModuleOneName)

	result := target.Module(apiModuleOneName)

	assert.Equal(t, mockApiModuleOne, result)
	mockApiModuleOne.AssertCalled(t, "Name")
}

func Test_RuntimeApi_Module_AllElements(t *testing.T) {
	target := setupRuntimeApi()

	mockApiModuleOne.On("Name").Return(apiModuleOneName)
	mockApiModuleTwo.On("Name").Return(apiModuleTwoName)

	result := target.Module(apiModuleTwoName)

	assert.Equal(t, mockApiModuleTwo, result)
	mockApiModuleOne.AssertCalled(t, "Name")
	mockApiModuleTwo.AssertCalled(t, "Name")
}

func Test_RuntimeApi_Module_Panics(t *testing.T) {
	target := setupRuntimeApi()
	name := "test"

	mockApiModuleOne.On("Name").Return(apiModuleOneName)
	mockApiModuleTwo.On("Name").Return(apiModuleTwoName)

	assert.PanicsWithValue(t,
		fmt.Sprintf("runtime module [%s] not found.", name),
		func() {
			target.Module("test")
		},
	)
	mockApiModuleOne.AssertCalled(t, "Name")
	mockApiModuleTwo.AssertCalled(t, "Name")
}

func setupRuntimeApi() RuntimeApi {
	mockApiModuleOne = new(mocks.ApiModule)
	mockApiModuleTwo = new(mocks.ApiModule)

	apis := []primitives.ApiModule{mockApiModuleOne, mockApiModuleTwo}

	return NewRuntimeApi(apis, logger)
}
