package sudo

import (
	"bytes"
	sc "github.com/LimeChain/goscale"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

const (
	docsCallSudoUncheckedWeight = "Authenticates the sudo key and dispatches a function call with `Root` origin." +
		"This function does not check the weight of the call, and instead allows the" +
		"Sudo user to specify the weight of the call." +
		"The dispatch origin for this call must be `Signed`."
)

// callSudoUncheckedWeight authenticates the sudo key and dispatches a function call with `Root` origin.
// This function does not check the weight of the call and instead allows the sudo user to specify the weight.
// The dispatch origin for this call must be `Signed`.
type callSudoUncheckedWeight struct {
	primitives.Callable
	dbWeight primitives.RuntimeDbWeight
	module   Module
}

func newCallSudoUncheckedWeight(moduleId sc.U8, functionId sc.U8, dbWeight primitives.RuntimeDbWeight, module Module) primitives.Call {
	call := callSudoUncheckedWeight{
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

func (c callSudoUncheckedWeight) DecodeSudoArgs(buffer *bytes.Buffer, decodeCallFunc func(buffer *bytes.Buffer) (primitives.Call, error)) (primitives.Call, error) {
	call, err := decodeCallFunc(buffer)
	if err != nil {
		return nil, err
	}
	weight, err := primitives.DecodeWeight(buffer)
	if err != nil {
		return nil, err
	}
	c.Arguments = sc.NewVaryingData(call, weight)

	return c, nil
}

func (c callSudoUncheckedWeight) DecodeArgs(buffer *bytes.Buffer) (primitives.Call, error) {
	c.module.logger.Critical("not implemented")
	return nil, nil
}

func (c callSudoUncheckedWeight) Encode(buffer *bytes.Buffer) error {
	return c.Callable.Encode(buffer)
}

func (c callSudoUncheckedWeight) Bytes() []byte {
	return c.Callable.Bytes()
}

func (c callSudoUncheckedWeight) ModuleIndex() sc.U8 { return c.Callable.ModuleIndex() }

func (c callSudoUncheckedWeight) FunctionIndex() sc.U8 { return c.Callable.FunctionIndex() }

func (c callSudoUncheckedWeight) Args() sc.VaryingData { return c.Callable.Args() }

func (c callSudoUncheckedWeight) BaseWeight() primitives.Weight {
	weight, ok := c.Arguments[1].(primitives.Weight)
	if !ok {
		c.module.logger.Critical("invalid [1] argument Weight")
	}

	return weight
}

func (_ callSudoUncheckedWeight) WeighData(baseWeight primitives.Weight) primitives.Weight {
	return primitives.WeightFromParts(baseWeight.RefTime, 0)
}

func (c callSudoUncheckedWeight) ClassifyDispatch(baseWeight primitives.Weight) primitives.DispatchClass {
	call := c.Args()[0].(primitives.Call)

	return call.ClassifyDispatch(baseWeight)
}

func (_ callSudoUncheckedWeight) PaysFee(_ primitives.Weight) primitives.Pays {
	return primitives.PaysNo
}

func (c callSudoUncheckedWeight) Dispatch(origin primitives.RuntimeOrigin, args sc.VaryingData) (primitives.PostDispatchInfo, error) {
	err := c.module.ensureSudo(origin)
	if err != nil {
		return primitives.PostDispatchInfo{}, err
	}

	call := args[0].(primitives.Call)

	return c.module.executeCall(primitives.NewRawOriginRoot(), call, newEventSudid)
}

func (_ callSudoUncheckedWeight) Docs() string {
	return docsCallSudoUncheckedWeight
}
