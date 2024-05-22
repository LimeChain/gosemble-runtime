package support

import (
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/mocks"
	"github.com/stretchr/testify/assert"
)

var (
	defaultValue = sc.U32(1)
	storageValue = sc.U32(5)
	decodeFunc   = sc.DecodeU32
	key          = []byte("key")

	mockStorage *mocks.IoStorage
)

func Test_SimpleStorageValue_Get(t *testing.T) {
	target := setupSimpleStorageValue()

	mockStorage.On("Get", key).Return(sc.NewOption[sc.Sequence[sc.U8]](sc.BytesToSequenceU8(storageValue.Bytes())), nil)

	result, err := target.Get()
	assert.NoError(t, err)

	assert.Equal(t, storageValue, result)
	mockStorage.AssertCalled(t, "Get", key)
}

func Test_SimpleStorageValue_Get_Nil(t *testing.T) {
	target := setupSimpleStorageValue()

	mockStorage.On("Get", key).Return(sc.NewOption[sc.Sequence[sc.U8]](nil), nil)

	result, err := target.Get()
	assert.NoError(t, err)

	assert.Equal(t, sc.U32(0), result)
	mockStorage.AssertCalled(t, "Get", key)
}

func Test_SimpleStorageValue_Get_OnEmpty(t *testing.T) {
	target := setupSimpleStorageValue()
	target.defaultValue = &defaultValue

	mockStorage.On("Get", key).Return(sc.NewOption[sc.Sequence[sc.U8]](nil), nil)

	result, err := target.Get()
	assert.NoError(t, err)

	assert.Equal(t, defaultValue, result)
	mockStorage.AssertCalled(t, "Get", key)
}

func Test_SimpleStorageValue_Get_Default_HasStorageValue(t *testing.T) {
	target := setupSimpleStorageValue()
	target.defaultValue = &defaultValue

	mockStorage.On("Get", key).Return(
		sc.NewOption[sc.Sequence[sc.U8]](
			sc.BytesToSequenceU8(storageValue.Bytes()),
		), nil)

	result, err := target.Get()
	assert.NoError(t, err)

	assert.Equal(t, storageValue, result)
	mockStorage.AssertCalled(t, "Get", key)
}

func Test_SimpleStorageValue_GetBytes(t *testing.T) {
	target := setupSimpleStorageValue()
	expect := sc.NewOption[sc.Sequence[sc.U8]](nil)

	mockStorage.On("Get", key).Return(expect, nil)

	result, err := target.GetBytes()
	assert.NoError(t, err)

	assert.Equal(t, expect, result)
	mockStorage.AssertCalled(t, "Get", key)
}

func Test_SimpleStorageValue_Exists(t *testing.T) {
	target := setupSimpleStorageValue()

	mockStorage.On("Exists", key).Return(true)

	result := target.Exists()

	assert.True(t, result)
	mockStorage.AssertCalled(t, "Exists", key)
}

func Test_SimpleStorageValue_Append(t *testing.T) {
	target := setupSimpleStorageValue()

	mockStorage.On("Append", key, storageValue.Bytes())

	target.Append(storageValue)

	mockStorage.AssertCalled(t, "Append", key, storageValue.Bytes())
}

func Test_SimpleStorageValue_Put(t *testing.T) {
	target := setupSimpleStorageValue()

	mockStorage.On("Set", key, storageValue.Bytes())

	target.Put(storageValue)

	mockStorage.AssertCalled(t, "Set", key, storageValue.Bytes())
}

func Test_SimpleStorageValue_Clear(t *testing.T) {
	target := setupSimpleStorageValue()

	mockStorage.On("Clear", key)

	target.Clear()

	mockStorage.AssertCalled(t, "Clear", key)
}

func Test_SimpleStorageValue_Take(t *testing.T) {
	target := setupSimpleStorageValue()

	mockStorage.On("Get", key).Return(
		sc.NewOption[sc.Sequence[sc.U8]](
			sc.BytesToSequenceU8(storageValue.Bytes()),
		), nil)
	mockStorage.On("Clear", key).Return()

	result, err := target.Take()
	assert.NoError(t, err)

	assert.Equal(t, storageValue, result)
	mockStorage.AssertCalled(t, "Get", key)
	mockStorage.AssertCalled(t, "Clear", key)
}

func Test_SimpleStorageValue_Take_Nil(t *testing.T) {
	target := setupSimpleStorageValue()

	mockStorage.On("Get", key).Return(sc.NewOption[sc.Sequence[sc.U8]](nil), nil)

	result, err := target.Take()
	assert.NoError(t, err)

	assert.Equal(t, sc.U32(0), result)
	mockStorage.AssertCalled(t, "Get", key)
}

func Test_SimpleStorageValue_TakeBytes(t *testing.T) {
	target := setupSimpleStorageValue()

	mockStorage.On("Get", key).Return(
		sc.NewOption[sc.Sequence[sc.U8]](
			sc.BytesToSequenceU8(storageValue.Bytes()),
		), nil)
	mockStorage.On("Clear", key).Return()

	result, err := target.TakeBytes()
	assert.NoError(t, err)

	assert.Equal(t, storageValue.Bytes(), result)
	mockStorage.AssertCalled(t, "Get", key)
	mockStorage.AssertCalled(t, "Clear", key)
}

func Test_SimpleStorageValue_TakeBytes_Nil(t *testing.T) {
	target := setupSimpleStorageValue()

	mockStorage.On("Get", key).Return(sc.NewOption[sc.Sequence[sc.U8]](nil), nil)

	result, err := target.TakeBytes()
	assert.NoError(t, err)

	assert.Equal(t, []byte(nil), result)
	mockStorage.AssertCalled(t, "Get", key)
}

func Test_SimpleStorageValue_DecodeLen(t *testing.T) {
	target := setupSimpleStorageValue()
	compactBytes := [5]byte{}
	offset := int32(0)

	mockStorage.On("Read", key, compactBytes[:], offset).Return(sc.NewOption[sc.U32](sc.U32(4)), nil)

	result, err := target.DecodeLen()
	assert.NoError(t, err)

	assert.Equal(t, sc.NewOption[sc.U64](sc.U64(0)), result)
	mockStorage.AssertCalled(t, "Read", key, compactBytes[:], offset)
}

func Test_SimpleStorageValue_DecodeLen_Nil(t *testing.T) {
	target := setupSimpleStorageValue()
	compactBytes := [5]byte{}
	offset := int32(0)

	mockStorage.On("Read", key, compactBytes[:], offset).Return(sc.NewOption[sc.U32](nil), nil)

	result, err := target.DecodeLen()
	assert.NoError(t, err)

	assert.Equal(t, sc.NewOption[sc.U64](nil), result)
	mockStorage.AssertCalled(t, "Read", key, compactBytes[:], offset)
}

func Test_SimpleStorageValue_key(t *testing.T) {
	target := setupSimpleStorageValue()

	assert.Equal(t, key, target.key)
}

func setupSimpleStorageValue() SimpleStorageValue[sc.U32] {
	mockStorage = new(mocks.IoStorage)

	target := NewSimpleStorageValue(mockStorage, key, decodeFunc).(SimpleStorageValue[sc.U32])

	return target
}
