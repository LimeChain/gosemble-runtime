package support

import (
	"errors"
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	prefix   = []byte("test")
	name     = []byte("support")
	keyValue = sc.U64(100)

	prefixHash                 = []byte("test_prefix")
	nameHash                   = []byte("test_name")
	keyValueHash               = []byte("test_100")
	concatHashStorageMapKeyKey = append(
		append(prefixHash, nameHash...),
		append(keyValueHash, keyValue.Bytes()...)...)

	mockHashing *mocks.IoHashing
	errPanic    = errors.New("panic")
)

func Test_HashStorageMap_Get(t *testing.T) {
	target := setupHashStorageMap()

	mockHashing.On("Twox128", prefix).Return(prefixHash)
	mockHashing.On("Twox128", name).Return(nameHash)
	mockHashing.On("Twox64", keyValue.Bytes()).Return(keyValueHash)
	mockStorage.On("Get", concatHashStorageMapKeyKey).Return(sc.NewOption[sc.Sequence[sc.U8]](nil), nil)

	result, err := target.Get(keyValue)
	assert.NoError(t, err)

	assert.Equal(t, sc.U32(0), result)
	mockHashing.AssertNumberOfCalls(t, "Twox128", 2)
	mockHashing.AssertCalled(t, "Twox128", prefix)
	mockHashing.AssertCalled(t, "Twox128", name)
	mockHashing.AssertNumberOfCalls(t, "Twox64", 1)
	mockHashing.AssertCalled(t, "Twox64", keyValue.Bytes())
	mockStorage.AssertCalled(t, "Get", concatHashStorageMapKeyKey)
}

func Test_HashStorageMap_Exists(t *testing.T) {
	target := setupHashStorageMap()

	mockHashing.On("Twox128", prefix).Return(prefixHash)
	mockHashing.On("Twox128", name).Return(nameHash)
	mockHashing.On("Twox64", keyValue.Bytes()).Return(keyValueHash)
	mockStorage.On("Exists", concatHashStorageMapKeyKey).Return(true)

	result := target.Exists(keyValue)

	assert.True(t, result)
	mockHashing.AssertNumberOfCalls(t, "Twox128", 2)
	mockHashing.AssertCalled(t, "Twox128", prefix)
	mockHashing.AssertCalled(t, "Twox128", name)
	mockHashing.AssertNumberOfCalls(t, "Twox64", 1)
	mockStorage.AssertCalled(t, "Exists", concatHashStorageMapKeyKey)
}

func Test_HashStorageMap_Append(t *testing.T) {
	target := setupHashStorageMap()

	mockHashing.On("Twox128", prefix).Return(prefixHash)
	mockHashing.On("Twox128", name).Return(nameHash)
	mockHashing.On("Twox64", keyValue.Bytes()).Return(keyValueHash)
	mockStorage.On("Append", concatHashStorageMapKeyKey, storageValue.Bytes())

	target.Append(keyValue, storageValue)

	mockHashing.AssertNumberOfCalls(t, "Twox128", 2)
	mockHashing.AssertCalled(t, "Twox128", prefix)
	mockHashing.AssertCalled(t, "Twox128", name)
	mockHashing.AssertNumberOfCalls(t, "Twox64", 1)
	mockHashing.AssertCalled(t, "Twox64", keyValue.Bytes())
	mockStorage.AssertCalled(t, "Append", concatHashStorageMapKeyKey, storageValue.Bytes())
}

func Test_HashStorageMap_TakeBytes(t *testing.T) {
	target := setupHashStorageMap()

	mockHashing.On("Twox128", prefix).Return(prefixHash)
	mockHashing.On("Twox128", name).Return(nameHash)
	mockHashing.On("Twox64", keyValue.Bytes()).Return(keyValueHash)
	mockStorage.On("Clear", concatHashStorageMapKeyKey)
	mockStorage.On("Get", concatHashStorageMapKeyKey).Return(
		sc.NewOption[sc.Sequence[sc.U8]](
			sc.BytesToSequenceU8(storageValue.Bytes()),
		), nil)
	mockStorage.On("Clear", concatHashStorageMapKeyKey).Return()

	result, err := target.TakeBytes(keyValue)
	assert.NoError(t, err)

	assert.Equal(t, storageValue.Bytes(), result)
	mockHashing.AssertNumberOfCalls(t, "Twox128", 2)
	mockHashing.AssertCalled(t, "Twox128", prefix)
	mockHashing.AssertCalled(t, "Twox128", name)
	mockHashing.AssertNumberOfCalls(t, "Twox64", 1)
	mockHashing.AssertCalled(t, "Twox64", keyValue.Bytes())
	mockStorage.AssertCalled(t, "Get", concatHashStorageMapKeyKey)
	mockStorage.AssertCalled(t, "Clear", concatHashStorageMapKeyKey)
}

func Test_HashStorageMap_TakeBytes_Nil(t *testing.T) {
	target := setupHashStorageMap()

	mockHashing.On("Twox128", prefix).Return(prefixHash)
	mockHashing.On("Twox128", name).Return(nameHash)
	mockHashing.On("Twox64", keyValue.Bytes()).Return(keyValueHash)
	mockStorage.On("Clear", concatHashStorageMapKeyKey)
	mockStorage.On("Get", concatHashStorageMapKeyKey).Return(sc.NewOption[sc.Sequence[sc.U8]](nil), nil)

	result, err := target.TakeBytes(keyValue)
	assert.NoError(t, err)

	assert.Equal(t, []byte(nil), result)
	mockHashing.AssertNumberOfCalls(t, "Twox128", 2)
	mockHashing.AssertCalled(t, "Twox128", prefix)
	mockHashing.AssertCalled(t, "Twox128", name)
	mockHashing.AssertNumberOfCalls(t, "Twox64", 1)
	mockHashing.AssertCalled(t, "Twox64", keyValue.Bytes())
	mockStorage.AssertCalled(t, "Get", concatHashStorageMapKeyKey)
	mockStorage.AssertNotCalled(t, "Clear", mock.Anything)
}

func Test_HashStorageMap_Remove(t *testing.T) {
	target := setupHashStorageMap()

	mockHashing.On("Twox128", prefix).Return(prefixHash)
	mockHashing.On("Twox128", name).Return(nameHash)
	mockHashing.On("Twox64", keyValue.Bytes()).Return(keyValueHash)
	mockStorage.On("Clear", concatHashStorageMapKeyKey)

	target.Remove(keyValue)

	mockHashing.AssertNumberOfCalls(t, "Twox128", 2)
	mockHashing.AssertCalled(t, "Twox128", prefix)
	mockHashing.AssertCalled(t, "Twox128", name)
	mockHashing.AssertNumberOfCalls(t, "Twox64", 1)
	mockHashing.AssertCalled(t, "Twox64", keyValue.Bytes())
	mockStorage.AssertCalled(t, "Clear", concatHashStorageMapKeyKey)
}

func Test_HashStorageMap_Clear(t *testing.T) {
	target := setupHashStorageMap()
	limit := sc.U32(7)

	mockHashing.On("Twox128", prefix).Return(prefixHash)
	mockHashing.On("Twox128", name).Return(nameHash)
	mockStorage.On("ClearPrefix", append(prefixHash, nameHash...), sc.NewOption[sc.U32](limit).Bytes()).Return()

	target.Clear(limit)

	mockHashing.AssertNumberOfCalls(t, "Twox128", 2)
	mockHashing.AssertCalled(t, "Twox128", prefix)
	mockHashing.AssertCalled(t, "Twox128", name)
	mockStorage.AssertCalled(t, "ClearPrefix", append(prefixHash, nameHash...), sc.NewOption[sc.U32](limit).Bytes())
}

func Test_HashStorageMap_Mutate(t *testing.T) {
	target := setupHashStorageMap()
	expectedResult := sc.NewU128(3)

	mockHashing.On("Twox128", prefix).Return(prefixHash)
	mockHashing.On("Twox128", name).Return(nameHash)
	mockHashing.On("Twox64", keyValue.Bytes()).Return(keyValueHash)
	mockStorage.On("Get", concatHashStorageMapKeyKey).Return(
		sc.NewOption[sc.Sequence[sc.U8]](
			sc.BytesToSequenceU8(storageValue.Bytes())), nil)
	mockStorage.On("Set", concatHashStorageMapKeyKey, storageValue.Bytes()).Return()

	result, err := target.Mutate(keyValue, func(s *sc.U32) (sc.Encodable, error) {
		return expectedResult, nil
	})

	assert.NoError(t, err)
	assert.Equal(t, expectedResult, result)
	mockHashing.AssertNumberOfCalls(t, "Twox128", 4)
	mockHashing.AssertCalled(t, "Twox128", prefix)
	mockHashing.AssertCalled(t, "Twox128", name)
	mockHashing.AssertNumberOfCalls(t, "Twox64", 2)
	mockHashing.AssertCalled(t, "Twox64", keyValue.Bytes())
	mockStorage.AssertCalled(t, "Get", concatHashStorageMapKeyKey)
	mockStorage.AssertCalled(t, "Set", concatHashStorageMapKeyKey, storageValue.Bytes())
}

func Test_HashStorageMap_Mutate_Error(t *testing.T) {
	target := setupHashStorageMap()
	expectedResult := sc.NewU128(5)
	expectedErr := errPanic

	mockHashing.On("Twox128", prefix).Return(prefixHash)
	mockHashing.On("Twox128", name).Return(nameHash)
	mockHashing.On("Twox64", keyValue.Bytes()).Return(keyValueHash)
	mockStorage.On("Get", concatHashStorageMapKeyKey).Return(sc.NewOption[sc.Sequence[sc.U8]](nil), nil)

	result, err := target.Mutate(keyValue, func(s *sc.U32) (sc.Encodable, error) {
		return expectedResult, errPanic
	})
	assert.Equal(t, expectedErr, err)

	assert.Equal(t, expectedResult, result)
	mockHashing.AssertNumberOfCalls(t, "Twox128", 2)
	mockHashing.AssertCalled(t, "Twox128", prefix)
	mockHashing.AssertCalled(t, "Twox128", name)
	mockHashing.AssertNumberOfCalls(t, "Twox64", 1)
	mockHashing.AssertCalled(t, "Twox64", keyValue.Bytes())
	mockStorage.AssertCalled(t, "Get", concatHashStorageMapKeyKey)
}

func Test_HashStorageMap_Mutate_Get_Error(t *testing.T) {
	target := setupHashStorageMap()

	expectedErr := errPanic

	mockHashing.On("Twox128", prefix).Return(prefixHash)
	mockHashing.On("Twox128", name).Return(nameHash)
	mockHashing.On("Twox64", keyValue.Bytes()).Return(keyValueHash)
	mockStorage.On("Get", concatHashStorageMapKeyKey).Return(sc.NewOption[sc.Sequence[sc.U8]](nil), expectedErr)

	_, err := target.Mutate(keyValue, func(s *sc.U32) (sc.Encodable, error) {
		return nil, nil
	})

	assert.Equal(t, expectedErr, err)
	mockHashing.AssertNumberOfCalls(t, "Twox128", 2)
	mockHashing.AssertCalled(t, "Twox128", prefix)
	mockHashing.AssertCalled(t, "Twox128", name)
	mockHashing.AssertNumberOfCalls(t, "Twox64", 1)
	mockHashing.AssertCalled(t, "Twox64", keyValue.Bytes())
	mockStorage.AssertCalled(t, "Get", concatHashStorageMapKeyKey)
}

func Test_HashStorageMap_TryMutateExists(t *testing.T) {
	target := setupHashStorageMap()
	expectOption := sc.NewOption[sc.U32](storageValue)
	expectedResult := sc.NewU128(3)

	mockHashing.On("Twox128", prefix).Return(prefixHash)
	mockHashing.On("Twox128", name).Return(nameHash)
	mockHashing.On("Twox64", keyValue.Bytes()).Return(keyValueHash)
	mockStorage.On("Get", concatHashStorageMapKeyKey).Return(
		sc.NewOption[sc.Sequence[sc.U8]](
			sc.BytesToSequenceU8(storageValue.Bytes())), nil)
	mockStorage.On("Set", concatHashStorageMapKeyKey, storageValue.Bytes()).Return()

	result, err := target.TryMutateExists(keyValue, func(option *sc.Option[sc.U32]) (sc.Encodable, error) {
		assert.Equal(t, &expectOption, option)
		return expectedResult, nil
	})
	assert.NoError(t, err)

	assert.Equal(t, expectedResult, result)
	mockHashing.AssertNumberOfCalls(t, "Twox128", 4)
	mockHashing.AssertCalled(t, "Twox128", prefix)
	mockHashing.AssertCalled(t, "Twox128", name)
	mockHashing.AssertNumberOfCalls(t, "Twox64", 2)
	mockHashing.AssertCalled(t, "Twox64", keyValue.Bytes())
	mockStorage.AssertCalled(t, "Get", concatHashStorageMapKeyKey)
	mockStorage.AssertCalled(t, "Set", concatHashStorageMapKeyKey, storageValue.Bytes())
}

func Test_HashStorageMap_TryMutateExists_Error(t *testing.T) {
	target := setupHashStorageMap()
	expectOption := sc.NewOption[sc.U32](sc.U32(0))
	expectResult := sc.NewU128(5)
	expectedErr := errPanic

	mockHashing.On("Twox128", prefix).Return(prefixHash)
	mockHashing.On("Twox128", name).Return(nameHash)
	mockHashing.On("Twox64", keyValue.Bytes()).Return(keyValueHash)
	mockStorage.On("Get", concatHashStorageMapKeyKey).Return(sc.NewOption[sc.Sequence[sc.U8]](nil), nil)

	result, err := target.TryMutateExists(keyValue, func(option *sc.Option[sc.U32]) (sc.Encodable, error) {
		assert.Equal(t, &expectOption, option)
		return expectResult, expectedErr
	})

	assert.Equal(t, expectedErr, err)
	assert.Equal(t, expectResult, result)
	mockHashing.AssertNumberOfCalls(t, "Twox128", 2)
	mockHashing.AssertCalled(t, "Twox128", prefix)
	mockHashing.AssertCalled(t, "Twox128", name)
	mockHashing.AssertNumberOfCalls(t, "Twox64", 1)
	mockHashing.AssertCalled(t, "Twox64", keyValue.Bytes())
	mockStorage.AssertCalled(t, "Get", concatHashStorageMapKeyKey)
}

func Test_HashStorageMap_TryMutateExists_GetError(t *testing.T) {
	target := setupHashStorageMap()
	expectedErr := errPanic

	mockHashing.On("Twox128", prefix).Return(prefixHash)
	mockHashing.On("Twox128", name).Return(nameHash)
	mockHashing.On("Twox64", keyValue.Bytes()).Return(keyValueHash)
	mockStorage.On("Get", concatHashStorageMapKeyKey).Return(sc.NewOption[sc.Sequence[sc.U8]](nil), expectedErr)

	_, err := target.TryMutateExists(keyValue, func(option *sc.Option[sc.U32]) (sc.Encodable, error) {
		return nil, nil
	})

	assert.Equal(t, expectedErr, err)
	mockHashing.AssertNumberOfCalls(t, "Twox128", 2)
	mockHashing.AssertCalled(t, "Twox128", prefix)
	mockHashing.AssertCalled(t, "Twox128", name)
	mockHashing.AssertNumberOfCalls(t, "Twox64", 1)
	mockHashing.AssertCalled(t, "Twox64", keyValue.Bytes())
	mockStorage.AssertCalled(t, "Get", concatHashStorageMapKeyKey)
}

func Test_HashStorageMap_key(t *testing.T) {
	target := setupHashStorageMap()

	mockHashing.On("Twox128", prefix).Return(prefixHash)
	mockHashing.On("Twox128", name).Return(nameHash)
	mockHashing.On("Twox64", keyValue.Bytes()).Return(keyValueHash)

	result := target.key(keyValue)

	assert.Equal(t, concatHashStorageMapKeyKey, result)
	mockHashing.AssertNumberOfCalls(t, "Twox128", 2)
	mockHashing.AssertCalled(t, "Twox128", prefix)
	mockHashing.AssertCalled(t, "Twox128", name)
	mockHashing.AssertNumberOfCalls(t, "Twox64", 1)
	mockHashing.AssertCalled(t, "Twox64", keyValue.Bytes())
}

func setupHashStorageMap() HashStorageMap[sc.U64, sc.U32] {
	mockHashing = new(mocks.IoHashing)
	mockStorage = new(mocks.IoStorage)

	target := NewHashStorageMap[sc.U64, sc.U32](mockStorage, prefix, name, mockHashing.Twox64, decodeFunc).(HashStorageMap[sc.U64, sc.U32])
	target.hashing = mockHashing

	return target
}
