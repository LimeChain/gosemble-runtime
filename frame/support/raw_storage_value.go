package support

import (
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/primitives/io"
)

// The value is stored as raw bytes, without encoding,
// thus, it is always a sequence of bytes.
type RawStorageValue struct {
	storage io.Storage
	key     []byte
}

func NewRawStorageValue(storage io.Storage, key []byte) StorageRawValue {
	return RawStorageValue{
		storage: storage,
		key:     key,
	}
}

func NewRawStorageValueFrom(storage io.Storage, key []byte) StorageRawValue {
	return RawStorageValue{
		storage: storage,
		key:     key,
	}
}

// Get a sequence of bytes from storage.
func (rsv RawStorageValue) Get() (sc.Sequence[sc.U8], error) {
	option, err := rsv.storage.Get(rsv.key)
	if err != nil {
		return nil, err
	}

	if option.HasValue {
		return option.Value, nil
	}

	return sc.Sequence[sc.U8]{}, nil
}

// Put a raw byte slice into storage.
func (rsv RawStorageValue) Put(value sc.Sequence[sc.U8]) {
	rsv.storage.Set(rsv.key, sc.SequenceU8ToBytes(value))
}

func (rsv RawStorageValue) Clear() {
	rsv.storage.Clear(rsv.key)
}

func (rsv RawStorageValue) ClearPrefix(limit sc.U32) {
	rsv.storage.ClearPrefix(rsv.key, sc.NewOption[sc.U32](limit).Bytes())
}
