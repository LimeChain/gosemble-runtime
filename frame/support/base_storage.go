package support

import (
	"bytes"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/primitives/io"
)

type baseStorage[T sc.Encodable] struct {
	storage      io.Storage
	decodeFunc   func(buffer *bytes.Buffer) (T, error)
	defaultValue *T
}

func newBaseStorage[T sc.Encodable](storage io.Storage, decodeFunc func(buffer *bytes.Buffer) (T, error), defaultValue *T) baseStorage[T] {
	return baseStorage[T]{
		storage:      storage,
		decodeFunc:   decodeFunc,
		defaultValue: defaultValue,
	}
}

func (bs baseStorage[T]) get(key []byte) (T, error) {
	if bs.defaultValue == nil {
		return bs.getDecode(key)
	}

	return bs.getDecodeOnEmpty(key)
}

func (bs baseStorage[T]) getBytes(key []byte) (sc.Option[sc.Sequence[sc.U8]], error) {
	return bs.storage.Get(key)
}

func (bs baseStorage[T]) exists(key []byte) bool {
	return bs.storage.Exists(key)
}

func (bs baseStorage[T]) put(key []byte, value T) {
	bs.storage.Set(key, value.Bytes())
}

func (bs baseStorage[T]) clear(key []byte) {
	bs.storage.Clear(key)
}

func (bs baseStorage[T]) append(key []byte, value T) {
	bs.storage.Append(key, value.Bytes())
}

func (bs baseStorage[T]) take(key []byte) (T, error) {
	return bs.takeDecode(key)
}

func (bs baseStorage[T]) decodeLen(key []byte) (sc.Option[sc.U64], error) {
	// `Compact<u32>` is 5 bytes in maximum.
	data := [5]byte{}
	option, err := bs.storage.Read(key, data[:], 0)
	if err != nil {
		return sc.Option[sc.U64]{}, err
	}

	if !option.HasValue {
		return sc.NewOption[sc.U64](nil), nil
	}

	length := sc.Min32(option.Value, sc.U32(len(data)))

	buffer := &bytes.Buffer{}
	buffer.Write(data[:length])

	compact, err := sc.DecodeCompact[sc.U64](buffer)
	if err != nil {
		return sc.Option[sc.U64]{}, err
	}
	toLen := sc.U64(compact.ToBigInt().Uint64())

	return sc.NewOption[sc.U64](toLen), nil
}

// getDecode gets the storage value and returns it decoded. The result from Get is Option<sc.Sequence[sc.U8]>.
// If the option is empty, it returns the default value T.
// If the option is not empty, it decodes it using decodeFunc and returns it.
func (bs baseStorage[T]) getDecode(key []byte) (T, error) {
	option, err := bs.storage.Get(key)
	if err != nil {
		return *new(T), err
	}

	if !option.HasValue {
		return *new(T), nil
	}

	buffer := &bytes.Buffer{}
	buffer.Write(sc.SequenceU8ToBytes(option.Value))

	f, err := bs.decodeFunc(buffer)
	if err != nil {
		return *new(T), err
	}

	return f, nil
}

func (bs baseStorage[T]) getDecodeOnEmpty(key []byte) (T, error) {
	option, err := bs.storage.Get(key)
	if err != nil {
		return *new(T), err
	}

	if !option.HasValue {
		return *bs.defaultValue, nil
	}

	buffer := &bytes.Buffer{}
	buffer.Write(sc.SequenceU8ToBytes(option.Value))

	f, err := bs.decodeFunc(buffer)
	if err != nil {
		return *new(T), err
	}

	return f, nil
}

// takeBytes gets the storage value. The result from Get is Option<sc.Sequence[sc.U8]>.
// If the option is empty, it returns nil.
// If the option is not empty, it clears it and returns the sequence as bytes.
func (bs baseStorage[T]) takeBytes(key []byte) ([]byte, error) {
	option, err := bs.storage.Get(key)
	if err != nil {
		return []byte{}, err
	}

	if !option.HasValue {
		return nil, nil
	}

	bs.storage.Clear(key)

	return sc.SequenceU8ToBytes(option.Value), nil
}

// TakeDecode gets the storage value and returns it decoded. The result from Get is Option<sc.Sequence[sc.U8]>.
// If the option is empty, it returns default value T.
// If the option is not empty, it clears it and returns decodeFunc(value).
func (bs baseStorage[T]) takeDecode(key []byte) (T, error) {
	option, err := bs.storage.Get(key)
	if err != nil {
		return *new(T), err
	}

	if !option.HasValue {
		return *new(T), nil
	}

	bs.storage.Clear(key)

	buffer := &bytes.Buffer{}
	buffer.Write(sc.SequenceU8ToBytes(option.Value))

	f, err := bs.decodeFunc(buffer)
	if err != nil {
		return *new(T), err
	}

	return f, nil
}
