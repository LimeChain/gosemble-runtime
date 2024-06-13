package system

import (
	"testing"

	"github.com/LimeChain/gosemble/hooks"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/mocks"
	"github.com/stretchr/testify/assert"
)

var (
	onSetCode hooks.OnSetCode
)

var (
	systemModule *mocks.SystemModule
)

func setupDefaultOnSetCode() {
	systemModule = new(mocks.SystemModule)
	onSetCode = NewDefaultOnSetCode(systemModule)
}

func Test_DefaultOnSetCode_New(t *testing.T) {
	setupDefaultOnSetCode()

	expected := defaultOnSetCode{module: systemModule}

	assert.Equal(t, expected, onSetCode)
}

func Test_DefaultOnSetCode_SetCode(t *testing.T) {
	setupDefaultOnSetCode()

	codeBlob := sc.BytesToSequenceU8([]byte{1, 2, 3})

	systemModule.On("UpdateCodeInStorage", codeBlob).Return()

	err := onSetCode.SetCode(codeBlob)

	assert.Nil(t, err)
	systemModule.AssertCalled(t, "UpdateCodeInStorage", codeBlob)
}
