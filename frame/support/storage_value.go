package support

import sc "github.com/LimeChain/goscale"

type StorageValue[T sc.Encodable] interface {
	Get() (T, error)
	GetBytes() (sc.Option[sc.Sequence[sc.U8]], error)
	Exists() bool
	Put(value T)
	Clear()
	Append(value T)
	// TODO:
	// support appending values with type different from T
	AppendItem(value sc.Encodable)
	TakeBytes() ([]byte, error)
	Take() (T, error)
	DecodeLen() (sc.Option[sc.U64], error)
	Mutate(f func(*T) (T, error)) (T, error)
}
