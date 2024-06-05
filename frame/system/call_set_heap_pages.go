package system

import (
	"bytes"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/frame/support"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

// Set the number of pages in the WebAssembly environment's heap.
type callSetHeapPages struct {
	primitives.Callable
	dbWeight     primitives.RuntimeDbWeight
	logDepositor LogDepositor
	heapPages    support.StorageValue[sc.U64]
}

func newCallSetHeapPages(
	moduleId sc.U8,
	functionId sc.U8,
	dbWeight primitives.RuntimeDbWeight,
	heapPages support.StorageValue[sc.U64],
	logDepositor LogDepositor,
) primitives.Call {
	call := callSetHeapPages{
		Callable: primitives.Callable{
			ModuleId:   moduleId,
			FunctionId: functionId,
			Arguments:  sc.NewVaryingData(sc.U64(0)),
		},
		dbWeight:     dbWeight,
		logDepositor: logDepositor,
		heapPages:    heapPages,
	}

	return call
}

func (c callSetHeapPages) DecodeArgs(buffer *bytes.Buffer) (primitives.Call, error) {
	heapPages, err := sc.DecodeU64(buffer)
	if err != nil {
		return nil, err
	}
	c.Arguments = sc.NewVaryingData(heapPages)
	return c, nil
}

func (c callSetHeapPages) Encode(buffer *bytes.Buffer) error {
	return c.Callable.Encode(buffer)
}

func (c callSetHeapPages) Bytes() []byte {
	return c.Callable.Bytes()
}

func (c callSetHeapPages) ModuleIndex() sc.U8 {
	return c.Callable.ModuleIndex()
}

func (c callSetHeapPages) FunctionIndex() sc.U8 {
	return c.Callable.FunctionIndex()
}

func (c callSetHeapPages) Args() sc.VaryingData {
	return c.Callable.Args()
}

func (c callSetHeapPages) BaseWeight() primitives.Weight {
	return callSetHeapPagesWeight(c.dbWeight)
}

func (_ callSetHeapPages) WeighData(baseWeight primitives.Weight) primitives.Weight {
	return primitives.WeightFromParts(baseWeight.RefTime, 0)
}

func (_ callSetHeapPages) ClassifyDispatch(baseWeight primitives.Weight) primitives.DispatchClass {
	return primitives.NewDispatchClassOperational()
}

func (_ callSetHeapPages) PaysFee(baseWeight primitives.Weight) primitives.Pays {
	return primitives.PaysNo
}

func (c callSetHeapPages) Dispatch(origin primitives.RuntimeOrigin, args sc.VaryingData) (primitives.PostDispatchInfo, error) {
	// TODO: enable once 'sudo' module is implemented
	//
	// err := EnsureRoot(origin)
	// if err != nil {
	// 	return primitives.PostDispatchInfo{}, err
	// }

	pages := args[0].(sc.U64)

	c.heapPages.Put(pages)
	c.logDepositor.DepositLog(primitives.NewDigestItemRuntimeEnvironmentUpgrade())

	return primitives.PostDispatchInfo{}, nil
}

func (_ callSetHeapPages) Docs() string {
	return "Set the number of pages in the WebAssembly environment's heap."
}
