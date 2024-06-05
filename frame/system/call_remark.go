package system

import (
	"bytes"

	sc "github.com/LimeChain/goscale"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

// Makes an on-chain remark.
// Can be executed by any origin.
type callRemark struct {
	primitives.Callable
	dbWeight primitives.RuntimeDbWeight
}

func newCallRemark(moduleId sc.U8, functionId sc.U8, dbWeight primitives.RuntimeDbWeight) primitives.Call {
	call := callRemark{
		Callable: primitives.Callable{
			ModuleId:   moduleId,
			FunctionId: functionId,
			Arguments:  sc.NewVaryingData(sc.Sequence[sc.U8]{}),
		},
		dbWeight: dbWeight,
	}

	return call
}

func (c callRemark) DecodeArgs(buffer *bytes.Buffer) (primitives.Call, error) {
	args, err := sc.DecodeSequence[sc.U8](buffer)
	if err != nil {
		return nil, err
	}
	c.Arguments = sc.NewVaryingData(args)
	return c, nil
}

func (c callRemark) Encode(buffer *bytes.Buffer) error {
	return c.Callable.Encode(buffer)
}

func (c callRemark) Bytes() []byte {
	return c.Callable.Bytes()
}

func (c callRemark) ModuleIndex() sc.U8 {
	return c.Callable.ModuleIndex()
}

func (c callRemark) FunctionIndex() sc.U8 {
	return c.Callable.FunctionIndex()
}

func (c callRemark) Args() sc.VaryingData {
	return c.Callable.Args()
}

func (c callRemark) BaseWeight() primitives.Weight {
	message := c.Arguments[0].(sc.Sequence[sc.U8])
	return callRemarkWeight(c.dbWeight, sc.U64(len(message)))
}

func (_ callRemark) WeighData(baseWeight primitives.Weight) primitives.Weight {
	return primitives.WeightFromParts(baseWeight.RefTime, 0)
}

func (_ callRemark) ClassifyDispatch(baseWeight primitives.Weight) primitives.DispatchClass {
	return primitives.NewDispatchClassNormal()
}

func (_ callRemark) PaysFee(baseWeight primitives.Weight) primitives.Pays {
	return primitives.PaysYes
}

func (_ callRemark) Dispatch(origin primitives.RuntimeOrigin, _ sc.VaryingData) (primitives.PostDispatchInfo, error) {
	return primitives.PostDispatchInfo{}, nil
}

func (_ callRemark) Docs() string {
	return "Make some on-chain remark."
}
