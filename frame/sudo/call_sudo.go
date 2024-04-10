package sudo

import (
	"bytes"
	sc "github.com/LimeChain/goscale"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

// callSudo authenticates the sudo key and dispatches a call function with `Root` origin.
type callSudo struct {
	primitives.Callable
	dbWeight primitives.RuntimeDbWeight
	module   Module
}

func newCallSudo(moduleId sc.U8, functionId sc.U8, dbWeight primitives.RuntimeDbWeight, module Module) primitives.Call {
	call := callSudo{
		Callable: primitives.Callable{
			ModuleId:   moduleId,
			FunctionId: functionId,
			Arguments:  sc.NewVaryingData(),
		},
		dbWeight: dbWeight,
		module:   module,
	}

	return call
}

func (c callSudo) DecodeSudoArgs(buffer *bytes.Buffer, decodeCallFunc func(buffer *bytes.Buffer) (primitives.Call, error)) (primitives.Call, error) {
	call, err := decodeCallFunc(buffer)
	if err != nil {
		return nil, err
	}
	c.Arguments = sc.NewVaryingData(call)
	return c, nil
}

func (c callSudo) DecodeArgs(_ *bytes.Buffer) (primitives.Call, error) {
	panic("not implemented")
}

func (c callSudo) Encode(buffer *bytes.Buffer) error {
	return c.Callable.Encode(buffer)
}

func (c callSudo) Bytes() []byte {
	return c.Callable.Bytes()
}

func (c callSudo) ModuleIndex() sc.U8 { return c.Callable.ModuleIndex() }

func (c callSudo) FunctionIndex() sc.U8 { return c.Callable.FunctionIndex() }

func (c callSudo) Args() sc.VaryingData { return c.Callable.Args() }

func (c callSudo) BaseWeight() primitives.Weight {
	call := c.Args()[0].(primitives.Call)

	return callSudoWeight(c.dbWeight).
		Add(call.WeighData(call.BaseWeight()))
}

func (_ callSudo) WeighData(baseWeight primitives.Weight) primitives.Weight {
	return primitives.WeightFromParts(baseWeight.RefTime, 0)
}

func (c callSudo) ClassifyDispatch(baseWeight primitives.Weight) primitives.DispatchClass {
	call := c.Args()[0].(primitives.Call)

	return call.ClassifyDispatch(baseWeight)
}

func (_ callSudo) PaysFee(_ primitives.Weight) primitives.Pays { return primitives.PaysNo }

func (c callSudo) Dispatch(origin primitives.RuntimeOrigin, args sc.VaryingData) (primitives.PostDispatchInfo, error) {
	err := c.module.ensureSudo(origin)
	if err != nil {
		return primitives.PostDispatchInfo{}, err
	}

	call := args[0].(primitives.Call)

	return c.module.executeCall(primitives.NewRawOriginRoot(), call, newEventSudid)
}

func (_ callSudo) Docs() string {
	return "Authenticates the sudo key and dispatches a call function with `Root` origin."
}
