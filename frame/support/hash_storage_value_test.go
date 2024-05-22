package support

import (
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/mocks"
	"github.com/stretchr/testify/assert"
)

var (
	concatHashStorageKey = append(prefixHash, nameHash...)
)

func Test_HashStorageValue_Get(t *testing.T) {
	target := setupHashStorageValue()

	mockHashing.On("Twox128", prefix).Return(prefixHash)
	mockHashing.On("Twox128", name).Return(nameHash)
	mockStorage.On("Get", concatHashStorageKey).Return(
		sc.NewOption[sc.Sequence[sc.U8]](
			sc.BytesToSequenceU8(storageValue.Bytes()),
		), nil)

	result, err := target.Get()
	assert.NoError(t, err)

	assert.Equal(t, storageValue, result)
	mockHashing.AssertNumberOfCalls(t, "Twox128", 2)
	mockHashing.AssertCalled(t, "Twox128", prefix)
	mockHashing.AssertCalled(t, "Twox128", name)
	mockStorage.AssertCalled(t, "Get", concatHashStorageKey)
}

func Test_HashStorageValue_Get_Nil(t *testing.T) {
	target := setupHashStorageValue()

	mockHashing.On("Twox128", prefix).Return(prefixHash)
	mockHashing.On("Twox128", name).Return(nameHash)
	mockStorage.On("Get", concatHashStorageKey).Return(sc.NewOption[sc.Sequence[sc.U8]](nil), nil)

	result, err := target.Get()
	assert.NoError(t, err)

	assert.Equal(t, sc.U32(0), result)
	mockHashing.AssertNumberOfCalls(t, "Twox128", 2)
	mockHashing.AssertCalled(t, "Twox128", prefix)
	mockHashing.AssertCalled(t, "Twox128", name)
	mockStorage.AssertCalled(t, "Get", concatHashStorageKey)
}

func Test_HashStorageValue_Get_OnEmpty(t *testing.T) {
	target := setupHashStorageValue()
	target.defaultValue = &defaultValue

	mockHashing.On("Twox128", prefix).Return(prefixHash)
	mockHashing.On("Twox128", name).Return(nameHash)
	mockStorage.On("Get", concatHashStorageKey).Return(sc.NewOption[sc.Sequence[sc.U8]](nil), nil)

	result, err := target.Get()
	assert.NoError(t, err)

	assert.Equal(t, defaultValue, result)
	mockHashing.AssertNumberOfCalls(t, "Twox128", 2)
	mockHashing.AssertCalled(t, "Twox128", prefix)
	mockHashing.AssertCalled(t, "Twox128", name)
	mockStorage.AssertCalled(t, "Get", concatHashStorageKey)
}

func Test_HashStorageValue_Get_Default_HasStorageValue(t *testing.T) {
	target := setupHashStorageValue()
	target.defaultValue = &defaultValue

	mockHashing.On("Twox128", prefix).Return(prefixHash)
	mockHashing.On("Twox128", name).Return(nameHash)
	mockStorage.On("Get", concatHashStorageKey).Return(
		sc.NewOption[sc.Sequence[sc.U8]](
			sc.BytesToSequenceU8(storageValue.Bytes()),
		), nil)

	result, err := target.Get()
	assert.NoError(t, err)

	assert.Equal(t, storageValue, result)
	mockHashing.AssertNumberOfCalls(t, "Twox128", 2)
	mockHashing.AssertCalled(t, "Twox128", prefix)
	mockHashing.AssertCalled(t, "Twox128", name)
	mockStorage.AssertCalled(t, "Get", concatHashStorageKey)
}

func Test_HashStorageValue_GetBytes(t *testing.T) {
	target := setupHashStorageValue()
	expect := sc.NewOption[sc.Sequence[sc.U8]](nil)

	mockHashing.On("Twox128", prefix).Return(prefixHash)
	mockHashing.On("Twox128", name).Return(nameHash)
	mockStorage.On("Get", concatHashStorageKey).Return(expect, nil)

	result, err := target.GetBytes()
	assert.NoError(t, err)

	assert.Equal(t, expect, result)
	mockHashing.AssertNumberOfCalls(t, "Twox128", 2)
	mockHashing.AssertCalled(t, "Twox128", prefix)
	mockHashing.AssertCalled(t, "Twox128", name)
	mockStorage.AssertCalled(t, "Get", concatHashStorageKey)
}

func Test_HashStorageValue_Exists(t *testing.T) {
	target := setupHashStorageValue()

	mockHashing.On("Twox128", prefix).Return(prefixHash)
	mockHashing.On("Twox128", name).Return(nameHash)
	mockStorage.On("Exists", concatHashStorageKey).Return(true)

	result := target.Exists()

	assert.True(t, result)
	mockHashing.AssertNumberOfCalls(t, "Twox128", 2)
	mockHashing.AssertCalled(t, "Twox128", prefix)
	mockHashing.AssertCalled(t, "Twox128", name)
	mockStorage.AssertCalled(t, "Exists", concatHashStorageKey)
}

func Test_HashStorageValue_Append(t *testing.T) {
	target := setupHashStorageValue()

	mockHashing.On("Twox128", prefix).Return(prefixHash)
	mockHashing.On("Twox128", name).Return(nameHash)
	mockStorage.On("Append", concatHashStorageKey, storageValue.Bytes())

	target.Append(storageValue)

	mockHashing.AssertNumberOfCalls(t, "Twox128", 2)
	mockHashing.AssertCalled(t, "Twox128", prefix)
	mockHashing.AssertCalled(t, "Twox128", name)
	mockStorage.AssertCalled(t, "Append", concatHashStorageKey, storageValue.Bytes())
}

func Test_HashStorageValue_Put(t *testing.T) {
	target := setupHashStorageValue()

	mockHashing.On("Twox128", prefix).Return(prefixHash)
	mockHashing.On("Twox128", name).Return(nameHash)
	mockStorage.On("Set", concatHashStorageKey, storageValue.Bytes())

	target.Put(storageValue)

	mockHashing.AssertNumberOfCalls(t, "Twox128", 2)
	mockHashing.AssertCalled(t, "Twox128", prefix)
	mockHashing.AssertCalled(t, "Twox128", name)
	mockStorage.AssertCalled(t, "Set", concatHashStorageKey, storageValue.Bytes())
}

func Test_HashStorageValue_Clear(t *testing.T) {
	target := setupHashStorageValue()

	mockHashing.On("Twox128", prefix).Return(prefixHash)
	mockHashing.On("Twox128", name).Return(nameHash)
	mockStorage.On("Clear", concatHashStorageKey)

	target.Clear()

	mockHashing.AssertNumberOfCalls(t, "Twox128", 2)
	mockHashing.AssertCalled(t, "Twox128", prefix)
	mockHashing.AssertCalled(t, "Twox128", name)
	mockStorage.AssertCalled(t, "Clear", concatHashStorageKey)
}

func Test_HashStorageValue_Take(t *testing.T) {
	target := setupHashStorageValue()

	mockHashing.On("Twox128", prefix).Return(prefixHash)
	mockHashing.On("Twox128", name).Return(nameHash)
	mockHashing.On("Twox64", keyValue.Bytes()).Return(keyValueHash)
	mockStorage.On("Get", concatHashStorageKey).Return(
		sc.NewOption[sc.Sequence[sc.U8]](
			sc.BytesToSequenceU8(storageValue.Bytes()),
		), nil)
	mockStorage.On("Clear", concatHashStorageKey).Return()

	result, err := target.Take()
	assert.NoError(t, err)

	assert.Equal(t, storageValue, result)
	mockHashing.AssertNumberOfCalls(t, "Twox128", 2)
	mockHashing.AssertCalled(t, "Twox128", prefix)
	mockHashing.AssertCalled(t, "Twox128", name)
	mockStorage.AssertCalled(t, "Get", concatHashStorageKey)
	mockStorage.AssertCalled(t, "Clear", concatHashStorageKey)
}

func Test_HashStorageValue_Take_Nil(t *testing.T) {
	target := setupHashStorageValue()

	mockHashing.On("Twox128", prefix).Return(prefixHash)
	mockHashing.On("Twox128", name).Return(nameHash)
	mockHashing.On("Twox64", keyValue.Bytes()).Return(keyValueHash)
	mockStorage.On("Get", concatHashStorageKey).Return(sc.NewOption[sc.Sequence[sc.U8]](nil), nil)

	result, err := target.Take()
	assert.NoError(t, err)

	assert.Equal(t, sc.U32(0), result)
	mockHashing.AssertNumberOfCalls(t, "Twox128", 2)
	mockHashing.AssertCalled(t, "Twox128", prefix)
	mockHashing.AssertCalled(t, "Twox128", name)
	mockStorage.AssertCalled(t, "Get", concatHashStorageKey)
}

func Test_HashStorageValue_TakeBytes(t *testing.T) {
	target := setupHashStorageValue()

	mockHashing.On("Twox128", prefix).Return(prefixHash)
	mockHashing.On("Twox128", name).Return(nameHash)
	mockHashing.On("Twox64", keyValue.Bytes()).Return(keyValueHash)
	mockStorage.On("Get", concatHashStorageKey).Return(
		sc.NewOption[sc.Sequence[sc.U8]](
			sc.BytesToSequenceU8(storageValue.Bytes()),
		), nil)
	mockStorage.On("Clear", concatHashStorageKey).Return()

	result, err := target.TakeBytes()
	assert.NoError(t, err)

	assert.Equal(t, storageValue.Bytes(), result)
	mockHashing.AssertNumberOfCalls(t, "Twox128", 2)
	mockHashing.AssertCalled(t, "Twox128", prefix)
	mockHashing.AssertCalled(t, "Twox128", name)
	mockStorage.AssertCalled(t, "Get", concatHashStorageKey)
	mockStorage.AssertCalled(t, "Clear", concatHashStorageKey)
}

func Test_HashStorageValue_TakeBytes_Nil(t *testing.T) {
	target := setupHashStorageValue()

	mockHashing.On("Twox128", prefix).Return(prefixHash)
	mockHashing.On("Twox128", name).Return(nameHash)
	mockHashing.On("Twox64", keyValue.Bytes()).Return(keyValueHash)
	mockStorage.On("Get", concatHashStorageKey).Return(sc.NewOption[sc.Sequence[sc.U8]](nil), nil)

	result, err := target.TakeBytes()
	assert.NoError(t, err)

	assert.Equal(t, []byte(nil), result)
	mockHashing.AssertNumberOfCalls(t, "Twox128", 2)
	mockHashing.AssertCalled(t, "Twox128", prefix)
	mockHashing.AssertCalled(t, "Twox128", name)
	mockStorage.AssertCalled(t, "Get", concatHashStorageKey)
}

func Test_HashStorageValue_DecodeLen(t *testing.T) {
	target := setupHashStorageValue()
	compactBytes := [5]byte{}
	offset := int32(0)

	mockHashing.On("Twox128", prefix).Return(prefixHash)
	mockHashing.On("Twox128", name).Return(nameHash)
	mockHashing.On("Twox64", keyValue.Bytes()).Return(keyValueHash)
	mockStorage.On("Read", concatHashStorageKey, compactBytes[:], offset).Return(sc.NewOption[sc.U32](sc.U32(4)), nil)

	result, err := target.DecodeLen()
	assert.NoError(t, err)

	assert.Equal(t, sc.NewOption[sc.U64](sc.U64(0)), result)
	mockHashing.AssertNumberOfCalls(t, "Twox128", 2)
	mockHashing.AssertCalled(t, "Twox128", prefix)
	mockHashing.AssertCalled(t, "Twox128", name)
	mockStorage.AssertCalled(t, "Read", concatHashStorageKey, compactBytes[:], offset)
}

func Test_HashStorageValue_DecodeLen_Nil(t *testing.T) {
	target := setupHashStorageValue()
	compactBytes := [5]byte{}
	offset := int32(0)

	mockHashing.On("Twox128", prefix).Return(prefixHash)
	mockHashing.On("Twox128", name).Return(nameHash)
	mockHashing.On("Twox64", keyValue.Bytes()).Return(keyValueHash)
	mockStorage.On("Read", concatHashStorageKey, compactBytes[:], offset).Return(sc.NewOption[sc.U32](nil), nil)

	result, err := target.DecodeLen()
	assert.NoError(t, err)

	assert.Equal(t, sc.NewOption[sc.U64](nil), result)
	mockHashing.AssertNumberOfCalls(t, "Twox128", 2)
	mockHashing.AssertCalled(t, "Twox128", prefix)
	mockHashing.AssertCalled(t, "Twox128", name)
	mockStorage.AssertCalled(t, "Read", concatHashStorageKey, compactBytes[:], offset)
}

func Test_HashStorageValue_key(t *testing.T) {
	target := setupHashStorageValue()

	mockHashing.On("Twox128", prefix).Return(prefixHash)
	mockHashing.On("Twox128", name).Return(nameHash)
	mockHashing.On("Twox64", keyValue.Bytes()).Return(keyValueHash)

	result := target.key()

	assert.Equal(t, concatHashStorageKey, result)
	mockHashing.AssertNumberOfCalls(t, "Twox128", 2)
	mockHashing.AssertCalled(t, "Twox128", prefix)
	mockHashing.AssertCalled(t, "Twox128", name)
}

func setupHashStorageValue() HashStorageValue[sc.U32] {
	mockHashing = new(mocks.IoHashing)
	mockStorage = new(mocks.IoStorage)

	target := NewHashStorageValue(mockStorage, prefix, name, decodeFunc).(HashStorageValue[sc.U32])
	target.hashing = mockHashing

	return target
}
