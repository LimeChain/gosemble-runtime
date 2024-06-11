package system

import (
	"bytes"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/frame/support"
	"github.com/LimeChain/gosemble/primitives/io"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

// Set some items of storage.
type callSetStorage struct {
	primitives.Callable
	dbWeight  primitives.RuntimeDbWeight
	ioStorage io.Storage
}

func newCallSetStorage(moduleId sc.U8, functionId sc.U8, dbWeight primitives.RuntimeDbWeight, ioStorage io.Storage) primitives.Call {
	call := callSetStorage{
		Callable: primitives.Callable{
			ModuleId:   moduleId,
			FunctionId: functionId,
			Arguments:  sc.NewVaryingData(sc.Sequence[KeyValue]{}),
		},
		dbWeight:  dbWeight,
		ioStorage: ioStorage,
	}

	return call
}

func (c callSetStorage) DecodeArgs(buffer *bytes.Buffer) (primitives.Call, error) {
	args, err := sc.DecodeSequenceWith(buffer, DecodeKeyValue)
	if err != nil {
		return nil, err
	}
	c.Arguments = sc.NewVaryingData(args)
	return c, nil
}

func (c callSetStorage) Encode(buffer *bytes.Buffer) error {
	return c.Callable.Encode(buffer)
}

func (c callSetStorage) Bytes() []byte {
	return c.Callable.Bytes()
}

func (c callSetStorage) ModuleIndex() sc.U8 {
	return c.Callable.ModuleIndex()
}

func (c callSetStorage) FunctionIndex() sc.U8 {
	return c.Callable.FunctionIndex()
}

func (c callSetStorage) Args() sc.VaryingData {
	return c.Callable.Args()
}

func (c callSetStorage) BaseWeight() primitives.Weight {
	items := c.Arguments[0].(sc.Sequence[KeyValue])
	return callSetStorageWeight(c.dbWeight, sc.U64(len(items)))
}

func (_ callSetStorage) WeighData(baseWeight primitives.Weight) primitives.Weight {
	return primitives.WeightFromParts(baseWeight.RefTime, 0)
}

func (_ callSetStorage) ClassifyDispatch(baseWeight primitives.Weight) primitives.DispatchClass {
	return primitives.NewDispatchClassOperational()
}

func (_ callSetStorage) PaysFee(baseWeight primitives.Weight) primitives.Pays {
	return primitives.PaysNo
}

func (c callSetStorage) Dispatch(origin primitives.RuntimeOrigin, args sc.VaryingData) (primitives.PostDispatchInfo, error) {
	err := EnsureRoot(origin)
	if err != nil {
		return primitives.PostDispatchInfo{}, err
	}

	items := args[0].(sc.Sequence[KeyValue])

	for _, item := range items {
		rsv := support.NewRawStorageValueFrom(c.ioStorage, sc.SequenceU8ToBytes(item.Key))
		rsv.Put(item.Value)
	}

	return primitives.PostDispatchInfo{}, nil
}

func (_ callSetStorage) Docs() string {
	return "Set some items of storage."
}
