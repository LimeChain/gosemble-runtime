package genesisbuilder

import (
	"errors"
	"testing"

	"github.com/ChainSafe/gossamer/lib/common"
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/mocks"
	"github.com/LimeChain/gosemble/primitives/log"
	"github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
)

var (
	genesis         = []byte("{\"module\":{\"field\":[]}}")
	genesisSequence = sc.BytesToSequenceU8(genesis).Bytes()
)

var (
	target          Module
	mockModule      *mocks.Module
	mockMemoryUtils *mocks.MemoryTranslator
)

func setup() {
	mockModule = new(mocks.Module)
	mockMemoryUtils = new(mocks.MemoryTranslator)

	target = New([]types.Module{mockModule}, log.NewLogger())
	target.memUtils = mockMemoryUtils
}

func Test_Module_Name(t *testing.T) {
	setup()
	result := target.Name()
	assert.Equal(t, ApiModuleName, result)
}

func Test_Module_Item(t *testing.T) {
	setup()
	hexName := common.MustBlake2b8([]byte(ApiModuleName))
	expect := types.NewApiItem(hexName, apiVersion)
	result := target.Item()
	assert.Equal(t, expect, result)
}

func Test_CreateDefaultConfig(t *testing.T) {
	setup()
	mockModule.On("CreateDefaultConfig").Return(genesis, nil)
	mockMemoryUtils.On("BytesToOffsetAndSize", genesisSequence).Return(int64(0))

	target.CreateDefaultConfig()

	mockModule.AssertCalled(t, "CreateDefaultConfig")
	mockMemoryUtils.AssertCalled(t, "BytesToOffsetAndSize", genesisSequence)
}

func Test_CreateDefaultConfig_Empty(t *testing.T) {
	genesisSequence := sc.BytesToSequenceU8([]byte("{}"))
	setup()
	mockModule.On("CreateDefaultConfig").Return([]byte{}, nil)
	mockMemoryUtils.On("BytesToOffsetAndSize", genesisSequence.Bytes()).Return(int64(0))

	target.CreateDefaultConfig()

	mockModule.AssertCalled(t, "CreateDefaultConfig")
	mockMemoryUtils.AssertCalled(t, "BytesToOffsetAndSize", genesisSequence.Bytes())
}

func Test_CreateDefaultConfig_Error(t *testing.T) {
	setup()
	mockModule.On("CreateDefaultConfig").Return(genesis, errors.New("err"))
	mockMemoryUtils.On("BytesToOffsetAndSize", genesisSequence).Return(int64(0))

	assert.PanicsWithValue(t,
		errors.New("err").Error(),
		func() { target.CreateDefaultConfig() },
	)
}

func Test_BuildConfig(t *testing.T) {
	setup()
	mockModule.On("BuildConfig", genesis).Return(nil)
	mockMemoryUtils.On("GetWasmMemorySlice", int32(0), int32(0)).Return(genesisSequence)
	mockMemoryUtils.On("BytesToOffsetAndSize", []byte{0}).Return(int64(0))

	target.BuildConfig(0, 0)

	mockMemoryUtils.AssertCalled(t, "GetWasmMemorySlice", int32(0), int32(0))
	mockModule.AssertCalled(t, "BuildConfig", genesis)
	mockMemoryUtils.AssertCalled(t, "BytesToOffsetAndSize", []byte{0})
}

func Test_BuildConfig_Error(t *testing.T) {
	setup()
	mockModule.On("BuildConfig", genesis).Return(errors.New("err"))
	mockMemoryUtils.On("GetWasmMemorySlice", int32(0), int32(0)).Return(genesisSequence)
	mockMemoryUtils.On("BytesToOffsetAndSize", []byte{0}).Return(int64(0))

	assert.PanicsWithValue(t,
		errors.New("err").Error(),
		func() { target.BuildConfig(0, 0) },
	)
}
