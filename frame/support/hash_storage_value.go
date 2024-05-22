package support

import (
	"bytes"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/primitives/io"
)

// HashStorageValue is a storage value, which takes `prefix` and `name` that are hashed using hashing.Twox128 and appended before each key value.
type HashStorageValue[T sc.Encodable] struct {
	baseStorage[T]
	prefix  []byte
	name    []byte
	hashing io.Hashing
}

func NewHashStorageValue[T sc.Encodable](storage io.Storage, prefix []byte, name []byte, decodeFunc func(buffer *bytes.Buffer) (T, error)) StorageValue[T] {
	return NewHashStorageValueWithDefault(storage, prefix, name, decodeFunc, nil)
}

func NewHashStorageValueWithDefault[T sc.Encodable](storage io.Storage, prefix []byte, name []byte, decodeFunc func(buffer *bytes.Buffer) (T, error), defaultValue *T) StorageValue[T] {
	return HashStorageValue[T]{
		baseStorage: newBaseStorage[T](storage, decodeFunc, defaultValue),
		prefix:      prefix,
		name:        name,
		hashing:     io.NewHashing(),
	}
}

func (hsv HashStorageValue[T]) Get() (T, error) {
	return hsv.baseStorage.get(hsv.key())
}

func (hsv HashStorageValue[T]) GetBytes() (sc.Option[sc.Sequence[sc.U8]], error) {
	return hsv.baseStorage.getBytes(hsv.key())
}

func (hsv HashStorageValue[T]) Exists() bool {
	return hsv.baseStorage.exists(hsv.key())
}

func (hsv HashStorageValue[T]) Put(value T) {
	hsv.baseStorage.put(hsv.key(), value)
}

func (hsv HashStorageValue[T]) Clear() {
	hsv.baseStorage.clear(hsv.key())
}

func (hsv HashStorageValue[T]) Append(value T) {
	hsv.baseStorage.append(hsv.key(), value)
}

// TODO:
// support appending values with type different from T
func (hsv HashStorageValue[T]) AppendItem(value sc.Encodable) {
	hsv.baseStorage.storage.Append(hsv.key(), value.Bytes())
}

func (hsv HashStorageValue[T]) Take() (T, error) {
	return hsv.baseStorage.take(hsv.key())
}

func (hsv HashStorageValue[T]) TakeBytes() ([]byte, error) {
	return hsv.baseStorage.takeBytes(hsv.key())
}

func (hsv HashStorageValue[T]) DecodeLen() (sc.Option[sc.U64], error) {
	return hsv.baseStorage.decodeLen(hsv.key())
}

func (hsv HashStorageValue[T]) key() []byte {
	prefixHash := hsv.hashing.Twox128(hsv.prefix)
	nameHash := hsv.hashing.Twox128(hsv.name)

	return append(prefixHash, nameHash...)
}
