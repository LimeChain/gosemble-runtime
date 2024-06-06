package mocks

import (
	sc "github.com/LimeChain/goscale"
	"github.com/stretchr/testify/mock"
)

type StorageValue[T sc.Encodable] struct {
	mock.Mock
}

func (m *StorageValue[T]) Get() (T, error) {
	args := m.Called()

	if args.Get(1) == nil {
		return args.Get(0).(T), nil
	}

	return args.Get(0).(T), args.Get(1).(error)
}

func (m *StorageValue[T]) GetBytes() (sc.Option[sc.Sequence[sc.U8]], error) {
	args := m.Called()

	if args.Get(1) == nil {
		return args.Get(0).(sc.Option[sc.Sequence[sc.U8]]), nil
	}

	return args.Get(0).(sc.Option[sc.Sequence[sc.U8]]), args.Get(1).(error)
}

func (m *StorageValue[T]) Exists() bool {
	args := m.Called()

	return args.Bool(0)
}

func (m *StorageValue[T]) Put(value T) {
	m.Called(value)
}

func (m *StorageValue[T]) Clear() {
	m.Called()
}

func (m *StorageValue[T]) Append(value T) {
	m.Called(value)
}

// TODO:
// support appending values with type different from T
func (m *StorageValue[T]) AppendItem(value sc.Encodable) {
	m.Called(value)
}

func (m *StorageValue[T]) Take() (T, error) {
	args := m.Called()

	if args.Get(1) == nil {
		return args.Get(0).(T), nil
	}

	return args.Get(0).(T), args.Get(1).(error)
}

func (m *StorageValue[T]) TakeBytes() ([]byte, error) {
	args := m.Called()

	if args.Error(1) == nil {
		return args.Get(0).([]byte), nil
	}

	return args.Get(0).([]byte), args.Get(1).(error)
}

func (m *StorageValue[T]) DecodeLen() (sc.Option[sc.U64], error) {
	args := m.Called()

	if args.Get(1) == nil {
		return args.Get(0).(sc.Option[sc.U64]), nil
	}

	return args.Get(0).(sc.Option[sc.U64]), args.Get(1).(error)
}

func (m *StorageValue[T]) Mutate(f func(*T) (T, error)) (T, error) {
	args := m.MethodCalled("Mutate", f)

	if args.Get(1) == nil {
		return args.Get(0).(T), nil
	}

	return args.Get(0).(T), args.Get(1).(error)
}
