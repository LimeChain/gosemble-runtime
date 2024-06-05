package system

import (
	"bytes"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/frame/support"
	"github.com/LimeChain/gosemble/primitives/io"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

// Kill all storage items with a key that starts with the given prefix.
type callKillPrefix struct {
	primitives.Callable
	dbWeight  primitives.RuntimeDbWeight
	ioStorage io.Storage
}

func newCallKillPrefix(moduleId sc.U8, functionId sc.U8, dbWeight primitives.RuntimeDbWeight, ioStorage io.Storage) primitives.Call {
	call := callKillPrefix{
		Callable: primitives.Callable{
			ModuleId:   moduleId,
			FunctionId: functionId,
			Arguments:  sc.NewVaryingData(sc.Sequence[sc.U8]{}, sc.U32(0)),
		},
		dbWeight:  dbWeight,
		ioStorage: ioStorage,
	}

	return call
}

func (c callKillPrefix) DecodeArgs(buffer *bytes.Buffer) (primitives.Call, error) {
	prefix, err := sc.DecodeSequence[sc.U8](buffer)
	if err != nil {
		return nil, err
	}
	subkeys, err := sc.DecodeU32(buffer)
	if err != nil {
		return nil, err
	}
	c.Arguments = sc.NewVaryingData(prefix, subkeys)
	return c, nil
}

func (c callKillPrefix) Encode(buffer *bytes.Buffer) error {
	return c.Callable.Encode(buffer)
}

func (c callKillPrefix) Bytes() []byte {
	return c.Callable.Bytes()
}

func (c callKillPrefix) ModuleIndex() sc.U8 {
	return c.Callable.ModuleIndex()
}

func (c callKillPrefix) FunctionIndex() sc.U8 {
	return c.Callable.FunctionIndex()
}

func (c callKillPrefix) Args() sc.VaryingData {
	return c.Callable.Args()
}

func (c callKillPrefix) BaseWeight() primitives.Weight {
	subkeys := c.Arguments[1].(sc.U32)
	return callKillPrefixWeight(c.dbWeight, sc.U64(subkeys))
}

func (_ callKillPrefix) WeighData(baseWeight primitives.Weight) primitives.Weight {
	return primitives.WeightFromParts(baseWeight.RefTime, 0)
}

func (_ callKillPrefix) ClassifyDispatch(baseWeight primitives.Weight) primitives.DispatchClass {
	return primitives.NewDispatchClassOperational()
}

func (_ callKillPrefix) PaysFee(baseWeight primitives.Weight) primitives.Pays {
	return primitives.PaysNo
}

func (c callKillPrefix) Dispatch(origin primitives.RuntimeOrigin, args sc.VaryingData) (primitives.PostDispatchInfo, error) {
	// TODO: enable once 'sudo' module is implemented
	//
	// err := EnsureRoot(origin)
	// if err != nil {
	// 	return primitives.PostDispatchInfo{}, err
	// }

	prefix := args[0].(sc.Sequence[sc.U8])
	subkeys := args[1].(sc.U32)

	rsv := support.NewRawStorageValueFrom(c.ioStorage, sc.SequenceU8ToBytes(prefix))
	rsv.ClearPrefix(subkeys)

	return primitives.PostDispatchInfo{}, nil
}

func (_ callKillPrefix) Docs() string {
	return "Kill all storage items with a key that starts with the given prefix."
}
