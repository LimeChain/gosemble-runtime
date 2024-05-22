package testable

import (
	"bytes"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/frame/support"
	"github.com/LimeChain/gosemble/primitives/io"
	"github.com/LimeChain/gosemble/primitives/log"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

type callTest struct {
	primitives.Callable
	storage           io.Storage
	transactionBroker io.TransactionBroker
}

func newCallTest(moduleId, functionId sc.U8, storage io.Storage, transactionBroker io.TransactionBroker) primitives.Call {
	call := callTest{
		Callable: primitives.Callable{
			ModuleId:   moduleId,
			FunctionId: functionId,
			Arguments:  sc.NewVaryingData(sc.Sequence[sc.U8]{}),
		},
		storage:           storage,
		transactionBroker: transactionBroker,
	}

	return call
}

func (c callTest) DecodeArgs(buffer *bytes.Buffer) (primitives.Call, error) {
	args, err := sc.DecodeSequence[sc.U8](buffer)
	if err != nil {
		return nil, err
	}
	c.Arguments = sc.NewVaryingData(args)
	return c, nil
}

func (c callTest) Encode(buffer *bytes.Buffer) error {
	return c.Callable.Encode(buffer)
}

func (c callTest) Bytes() []byte {
	return c.Callable.Bytes()
}

func (c callTest) ModuleIndex() sc.U8 {
	return c.Callable.ModuleIndex()
}

func (c callTest) FunctionIndex() sc.U8 {
	return c.Callable.FunctionIndex()
}

func (c callTest) Args() sc.VaryingData {
	return c.Callable.Args()
}

func (_ callTest) BaseWeight() primitives.Weight {
	return primitives.WeightFromParts(1_000_000, 0)
}

func (_ callTest) WeighData(baseWeight primitives.Weight) primitives.Weight {
	return primitives.WeightFromParts(baseWeight.RefTime, 0)
}

func (_ callTest) ClassifyDispatch(baseWeight primitives.Weight) primitives.DispatchClass {
	return primitives.NewDispatchClassNormal()
}

func (_ callTest) PaysFee(baseWeight primitives.Weight) primitives.Pays {
	return primitives.PaysYes
}

func (c callTest) Dispatch(origin primitives.RuntimeOrigin, _ sc.VaryingData) (primitives.PostDispatchInfo, error) {
	storage := io.NewStorage()
	storage.Set([]byte("testvalue"), []byte{1})

	transactional := support.NewTransactional[primitives.PostDispatchInfo](c.storage, c.transactionBroker, log.NewLogger())
	// TODO: handle err
	transactional.WithStorageLayer(func() (primitives.PostDispatchInfo, error) {
		storage.Set([]byte("testvalue"), []byte{2})
		return primitives.PostDispatchInfo{}, primitives.NewDispatchErrorOther("revert")
	})

	return primitives.PostDispatchInfo{}, nil
}

func (_ callTest) Docs() string {
	return "Make test"
}
