package system

import (
	"bytes"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/frame/support"
	"github.com/LimeChain/gosemble/primitives/io"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

// Kill some items from storage.
type callKillStorage struct {
	primitives.Callable
	dbWeight  primitives.RuntimeDbWeight
	ioStorage io.Storage
}

func newCallKillStorage(moduleId sc.U8, functionId sc.U8, dbWeight primitives.RuntimeDbWeight, ioStorage io.Storage) primitives.Call {
	call := callKillStorage{
		Callable: primitives.Callable{
			ModuleId:   moduleId,
			FunctionId: functionId,
			Arguments:  sc.NewVaryingData(sc.Sequence[sc.Sequence[sc.U8]]{}),
		},
		dbWeight:  dbWeight,
		ioStorage: ioStorage,
	}

	return call
}

func (c callKillStorage) DecodeArgs(buffer *bytes.Buffer) (primitives.Call, error) {
	args, err := sc.DecodeSequence[sc.Sequence[sc.U8]](buffer)
	if err != nil {
		return nil, err
	}
	c.Arguments = sc.NewVaryingData(args)
	return c, nil
}

func (c callKillStorage) Encode(buffer *bytes.Buffer) error {
	return c.Callable.Encode(buffer)
}

func (c callKillStorage) Bytes() []byte {
	return c.Callable.Bytes()
}

func (c callKillStorage) ModuleIndex() sc.U8 {
	return c.Callable.ModuleIndex()
}

func (c callKillStorage) FunctionIndex() sc.U8 {
	return c.Callable.FunctionIndex()
}

func (c callKillStorage) Args() sc.VaryingData {
	return c.Callable.Args()
}

func (c callKillStorage) BaseWeight() primitives.Weight {
	keys := c.Arguments[0].(sc.Sequence[sc.Sequence[sc.U8]])
	return callKillStorageWeight(c.dbWeight, sc.U64(len(keys)))
}

func (_ callKillStorage) WeighData(baseWeight primitives.Weight) primitives.Weight {
	return primitives.WeightFromParts(baseWeight.RefTime, 0)
}

func (_ callKillStorage) ClassifyDispatch(baseWeight primitives.Weight) primitives.DispatchClass {
	return primitives.NewDispatchClassOperational()
}

func (_ callKillStorage) PaysFee(baseWeight primitives.Weight) primitives.Pays {
	return primitives.PaysNo
}

func (c callKillStorage) Dispatch(origin primitives.RuntimeOrigin, args sc.VaryingData) (primitives.PostDispatchInfo, error) {
	// TODO: enable once 'sudo' module is implemented
	//
	// err := EnsureRoot(origin)
	// if err != nil {
	// 	return primitives.PostDispatchInfo{}, err
	// }

	keys := args[0].(sc.Sequence[sc.Sequence[sc.U8]])

	for _, key := range keys {
		rsv := support.NewRawStorageValueFrom(c.ioStorage, sc.SequenceU8ToBytes(key))
		rsv.Clear()
	}

	return primitives.PostDispatchInfo{}, nil
}

func (_ callKillStorage) Docs() string {
	return "Kill some items from storage."
}
