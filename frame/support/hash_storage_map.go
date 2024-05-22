package support

import (
	"bytes"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/primitives/io"
)

// HashStorageMap is a key-value storage map, which takes `prefix` and `name` that are hashed using hashing.Twox128 and appended before each key value.
type HashStorageMap[K, V sc.Encodable] struct {
	baseStorage[V]
	prefix      []byte
	name        []byte
	keyHashFunc func([]byte) []byte
	decodeFunc  func(buffer *bytes.Buffer) (V, error)
	hashing     io.Hashing
}

func NewHashStorageMap[K, V sc.Encodable](storage io.Storage, prefix []byte, name []byte, keyHashFunc func([]byte) []byte, decodeFunc func(buffer *bytes.Buffer) (V, error)) StorageMap[K, V] {
	return HashStorageMap[K, V]{
		newBaseStorage[V](storage, decodeFunc, nil),
		prefix,
		name,
		keyHashFunc,
		decodeFunc,
		io.NewHashing(),
	}
}

func (hsm HashStorageMap[K, V]) Get(k K) (V, error) {
	return hsm.baseStorage.getDecode(hsm.key(k))
}

func (hsm HashStorageMap[K, V]) Exists(k K) bool {
	return hsm.baseStorage.exists(hsm.key(k))
}

func (hsm HashStorageMap[K, V]) Put(k K, value V) {
	hsm.baseStorage.put(hsm.key(k), value)
}

func (hsm HashStorageMap[K, V]) Append(k K, value V) {
	hsm.baseStorage.append(hsm.key(k), value)
}

func (hsm HashStorageMap[K, V]) TakeBytes(k K) ([]byte, error) {
	return hsm.baseStorage.takeBytes(hsm.key(k))
}

func (hsm HashStorageMap[K, V]) Remove(k K) {
	hsm.baseStorage.clear(hsm.key(k))
}

func (hsm HashStorageMap[K, V]) Clear(limit sc.U32) {
	prefixHash := hsm.hashing.Twox128(hsm.prefix)
	nameHash := hsm.hashing.Twox128(hsm.name)

	hsm.baseStorage.storage.ClearPrefix(append(prefixHash, nameHash...), sc.NewOption[sc.U32](limit).Bytes())
}

func (hsm HashStorageMap[K, V]) Mutate(k K, f func(*V) (sc.Encodable, error)) (sc.Encodable, error) {
	v, err := hsm.Get(k)
	if err != nil {
		return nil, err
	}

	result, err := f(&v)
	if err == nil {
		hsm.Put(k, v)
	}

	return result, err
}

func (hsm HashStorageMap[K, V]) TryMutateExists(k K, f func(option *sc.Option[V]) (sc.Encodable, error)) (sc.Encodable, error) {
	// TODO: This should get the storage value and try to decode it. It should return an Option<value>
	// If it cannot decode it, return Empty Option.
	// If it can decode it, return Option with the value.
	v, err := hsm.Get(k)
	if err != nil {
		return nil, err
	}
	option := sc.NewOption[V](v)

	result, err := f(&option)
	if err == nil {
		if option.HasValue {
			hsm.Put(k, option.Value)
		}
		// }
		// else {
		// hsm.Remove(k)
		// }
	}

	return result, err
}

func (hsm HashStorageMap[K, V]) key(key K) []byte {
	prefixHash := hsm.hashing.Twox128(hsm.prefix)
	nameHash := hsm.hashing.Twox128(hsm.name)

	keyBytes := key.Bytes()
	keyHash := hsm.keyHashFunc(keyBytes)

	concatKey := append(prefixHash, nameHash...)
	concatKey = append(concatKey, keyHash...)
	concatKey = append(concatKey, keyBytes...)

	return concatKey
}
