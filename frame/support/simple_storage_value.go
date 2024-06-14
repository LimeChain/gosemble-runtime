package support

import (
	"bytes"
	"github.com/LimeChain/gosemble/primitives/io"

	sc "github.com/LimeChain/goscale"
)

// SimpleStorageValue takes a key upon initialisation and uses it in raw format to get storage values.
type SimpleStorageValue[T sc.Encodable] struct {
	baseStorage[T]
	key []byte
}

func NewSimpleStorageValue[T sc.Encodable](storage io.Storage, key []byte, decodeFunc func(buffer *bytes.Buffer) (T, error)) StorageValue[T] {
	return SimpleStorageValue[T]{
		baseStorage: newBaseStorage[T](storage, decodeFunc, nil),
		key:         key,
	}
}

func (ssv SimpleStorageValue[T]) Get() (T, error) {
	return ssv.baseStorage.get(ssv.key)
}

func (ssv SimpleStorageValue[T]) GetBytes() (sc.Option[sc.Sequence[sc.U8]], error) {
	return ssv.baseStorage.getBytes(ssv.key)
}

func (ssv SimpleStorageValue[T]) Exists() bool {
	return ssv.baseStorage.exists(ssv.key)
}

func (ssv SimpleStorageValue[T]) Put(value T) {
	ssv.baseStorage.put(ssv.key, value)
}

func (ssv SimpleStorageValue[T]) Clear() {
	ssv.baseStorage.clear(ssv.key)
}

func (ssv SimpleStorageValue[T]) Append(value T) {
	ssv.baseStorage.append(ssv.key, value)
}

// TODO:
// support appending values with type different from T
func (ssv SimpleStorageValue[T]) AppendItem(value sc.Encodable) {
	ssv.baseStorage.storage.Append(ssv.key, value.Bytes())
}

func (ssv SimpleStorageValue[T]) TakeBytes() ([]byte, error) {
	return ssv.baseStorage.takeBytes(ssv.key)
}

func (ssv SimpleStorageValue[T]) Take() (T, error) {
	return ssv.baseStorage.take(ssv.key)
}

func (ssv SimpleStorageValue[T]) DecodeLen() (sc.Option[sc.U64], error) {
	return ssv.baseStorage.decodeLen(ssv.key)
}

func (ssv SimpleStorageValue[T]) Mutate(f func(*T) (T, error)) (T, error) {
	v, err := ssv.Get()
	if err != nil {
		return *new(T), err
	}

	result, err := f(&v)
	if err == nil {
		ssv.Put(v)
	}

	return result, err
}
