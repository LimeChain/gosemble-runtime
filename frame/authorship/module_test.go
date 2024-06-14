package authorship

import (
	"errors"
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants"
	"github.com/LimeChain/gosemble/constants/metadata"
	"github.com/LimeChain/gosemble/mocks"
	"github.com/LimeChain/gosemble/primitives/log"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	moduleId sc.U8 = 1

	digest = primitives.NewDigest(sc.Sequence[primitives.DigestItem]{})

	unknownTransactionNoUnsignedValidator = primitives.NewTransactionValidityError(primitives.NewUnknownTransactionNoUnsignedValidator())
)

var (
	mockStorage                    *mocks.IoStorage
	mockSystemModule               *mocks.SystemModule
	mockEventHandler               *mocks.AuthorshipEventHandler
	mockFindAccountFromAuthorIndex *mocks.FindAccountFromAuthorIndex
	mockStorageAuthor              *mocks.StorageValue[primitives.AccountId]
	logger                         = log.NewLogger()
	mdGenerator                    = primitives.NewMetadataTypeGenerator()
)

var target module

func setup() {
	mockStorage = new(mocks.IoStorage)
	mockSystemModule = new(mocks.SystemModule)
	mockFindAccountFromAuthorIndex = new(mocks.FindAccountFromAuthorIndex)
	mockEventHandler = new(mocks.AuthorshipEventHandler)
	mockStorageAuthor = new(mocks.StorageValue[primitives.AccountId])

	config := NewConfig(
		mockStorage,
		mockFindAccountFromAuthorIndex,
		mockEventHandler,
		mockSystemModule,
	)
	target = New(moduleId, config, mdGenerator, logger).(module)

	target.storage.Author = mockStorageAuthor
}

func Test_Authorship_Module_GetIndex(t *testing.T) {
	setup()

	assert.Equal(t, moduleId, target.GetIndex())
}

func Test_Authorship_Module_Functions(t *testing.T) {
	setup()

	assert.Equal(t, 0, len(target.Functions()))
}

func Test_Authorship_Module_PreDispatch(t *testing.T) {
	setup()

	result, err := target.PreDispatch(new(mocks.Call))

	assert.Nil(t, err)
	assert.Equal(t, sc.Empty{}, result)
}

func Test_Authorship_Module_ValidateUnsigned(t *testing.T) {
	setup()

	result, err := target.ValidateUnsigned(primitives.NewTransactionSourceLocal(), new(mocks.Call))

	assert.Equal(t, unknownTransactionNoUnsignedValidator, err)
	assert.Equal(t, primitives.ValidTransaction{}, result)
}

func Test_Authorship_Module_Author_Fails_To_Get_Storage_Author(t *testing.T) {
	setup()

	mockStorageAuthor.On("GetBytes").Return(sc.NewOption[sc.Sequence[sc.U8]](nil), errors.New("get author error"))

	result, err := target.Author()

	assert.Error(t, err)
	assert.Equal(t, sc.NewOption[primitives.AccountId](nil), result)

	mockStorageAuthor.AssertCalled(t, "GetBytes")
}

func Test_Authorship_Module_No_Author(t *testing.T) {
	setup()

	mockStorageAuthor.On("GetBytes").Return(sc.NewOption[sc.Sequence[sc.U8]](nil), nil)
	mockSystemModule.On("StorageDigest").Return(digest, nil)
	mockFindAccountFromAuthorIndex.On("FindAuthor", sc.Sequence[primitives.DigestPreRuntime]{}).
		Return(sc.NewOption[primitives.AccountId](nil), errors.New("find author error"))

	result, err := target.Author()

	assert.Error(t, err)
	assert.Equal(t, sc.NewOption[primitives.AccountId](nil), result)

	mockStorageAuthor.AssertCalled(t, "GetBytes")
	mockSystemModule.AssertCalled(t, "StorageDigest")
	mockFindAccountFromAuthorIndex.AssertCalled(t, "FindAuthor", sc.Sequence[primitives.DigestPreRuntime]{})
	mockStorageAuthor.AssertNotCalled(t, "Put", mock.Anything)
}

func Test_Authorship_Module_Author(t *testing.T) {
	setup()

	mockStorageAuthor.On("GetBytes").Return(sc.NewOption[sc.Sequence[sc.U8]](nil), nil)
	mockSystemModule.On("StorageDigest").Return(digest, nil)
	mockFindAccountFromAuthorIndex.On("FindAuthor", sc.Sequence[primitives.DigestPreRuntime]{}).
		Return(sc.NewOption[primitives.AccountId](constants.OneAccountId), nil)
	mockStorageAuthor.On("Put", constants.OneAccountId).Return(nil)

	result, err := target.Author()

	assert.NoError(t, err)
	assert.Equal(t, sc.NewOption[primitives.AccountId](constants.OneAccountId), result)

	mockStorageAuthor.AssertCalled(t, "GetBytes")
	mockSystemModule.AssertCalled(t, "StorageDigest")
	mockFindAccountFromAuthorIndex.AssertCalled(t, "FindAuthor", sc.Sequence[primitives.DigestPreRuntime]{})
	mockStorageAuthor.AssertNotCalled(t, "Put", sc.NewOption[primitives.AccountId](constants.OneAccountId))
}

func Test_Authorship_Module_OnInitialize(t *testing.T) {
	setup()

	accountIdBytes := sc.NewOption[sc.Sequence[sc.U8]](sc.BytesToSequenceU8(constants.OneAccountId.Bytes()))

	mockStorageAuthor.On("GetBytes").Return(accountIdBytes, nil)
	mockEventHandler.On("NoteAuthor", constants.OneAccountId).Return(nil)

	result, err := target.OnInitialize(0)

	assert.NoError(t, err)
	assert.Equal(t, primitives.WeightZero(), result)

	mockStorageAuthor.AssertCalled(t, "GetBytes")
	mockEventHandler.AssertCalled(t, "NoteAuthor", constants.OneAccountId)
}

func Test_Authorship_Module_OnFinalize(t *testing.T) {
	setup()

	mockStorageAuthor.On("Clear").Return(nil)

	err := target.OnFinalize(0)

	assert.NoError(t, err)
	mockStorageAuthor.AssertCalled(t, "Clear")
}

func Test_Authorship_Module_Metadata(t *testing.T) {
	setup()

	moduleV14 := primitives.MetadataModuleV14{
		Name: "Authorship",
		Storage: sc.NewOption[primitives.MetadataModuleStorage](primitives.MetadataModuleStorage{
			Prefix: "Authorship",
			Items: sc.Sequence[primitives.MetadataModuleStorageEntry]{
				primitives.NewMetadataModuleStorageEntry(
					"Author",
					primitives.MetadataModuleStorageEntryModifierOptional,
					primitives.NewMetadataModuleStorageEntryDefinitionPlain(sc.ToCompact(metadata.TypesAddress32)),
					"Author of current block.",
				),
			},
		}),
		Call:      sc.NewOption[sc.Compact](nil),
		CallDef:   sc.NewOption[primitives.MetadataDefinitionVariant](nil),
		Event:     sc.NewOption[sc.Compact](nil),
		EventDef:  sc.NewOption[primitives.MetadataDefinitionVariant](nil),
		Constants: sc.Sequence[primitives.MetadataModuleConstant]{},
		Error:     sc.NewOption[sc.Compact](nil),
		ErrorDef:  sc.NewOption[primitives.MetadataDefinitionVariant](nil),
		Index:     moduleId,
	}

	expectMetadataModule := primitives.MetadataModule{
		Version:   primitives.ModuleVersion14,
		ModuleV14: moduleV14,
	}

	metadataModule := target.Metadata()

	assert.Equal(t, expectMetadataModule, metadataModule)
}
