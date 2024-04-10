package sudo

import (
	"bytes"
	sc "github.com/LimeChain/goscale"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

// callSudoAs authenticates the sudo key and dispatches a function call with `Signed` origin from the given account.
type callSudoAs struct {
	primitives.Callable
	dbWeight primitives.RuntimeDbWeight
	module   Module
}

func newCallSudoAs(moduleId sc.U8, functionId sc.U8, dbWeight primitives.RuntimeDbWeight, module Module) primitives.Call {
	call := callSudoAs{
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

func (c callSudoAs) DecodeSudoArgs(buffer *bytes.Buffer, decodeCallFunc func(buffer *bytes.Buffer) (primitives.Call, error)) (primitives.Call, error) {
	who, err := primitives.DecodeMultiAddress(buffer)
	if err != nil {
		return nil, err
	}

	call, err := decodeCallFunc(buffer)
	if err != nil {
		return nil, err
	}
	c.Arguments = sc.NewVaryingData(who, call)

	return c, nil
}

func (c callSudoAs) DecodeArgs(_ *bytes.Buffer) (primitives.Call, error) {
	panic("not implemented")
}

func (c callSudoAs) Encode(buffer *bytes.Buffer) error {
	return c.Callable.Encode(buffer)
}

func (c callSudoAs) Bytes() []byte {
	return c.Callable.Bytes()
}

func (c callSudoAs) ModuleIndex() sc.U8 {
	return c.Callable.ModuleIndex()
}

func (c callSudoAs) FunctionIndex() sc.U8 {
	return c.Callable.FunctionIndex()
}

func (c callSudoAs) Args() sc.VaryingData {
	return c.Callable.Args()
}

func (c callSudoAs) BaseWeight() primitives.Weight {
	call := c.Args()[1].(primitives.Call)

	return callSudoAsWeight(c.dbWeight).
		Add(call.WeighData(call.BaseWeight()))
}

func (_ callSudoAs) WeighData(baseWeight primitives.Weight) primitives.Weight {
	return primitives.WeightFromParts(baseWeight.RefTime, 0)
}

func (c callSudoAs) ClassifyDispatch(baseWeight primitives.Weight) primitives.DispatchClass {
	call := c.Args()[1].(primitives.Call)

	return call.ClassifyDispatch(baseWeight)
}

func (_ callSudoAs) PaysFee(_ primitives.Weight) primitives.Pays {
	return primitives.PaysNo
}

func (c callSudoAs) Dispatch(origin primitives.RuntimeOrigin, args sc.VaryingData) (primitives.PostDispatchInfo, error) {
	err := c.module.ensureSudo(origin)
	if err != nil {
		return primitives.PostDispatchInfo{}, err
	}

	multiAddress := args[0].(primitives.MultiAddress)

	who, err := primitives.Lookup(multiAddress)
	if err != nil {
		c.module.logger.Debugf("Failed to lookup [%s]", multiAddress.Bytes())
		return primitives.PostDispatchInfo{}, primitives.NewDispatchErrorCannotLookup()
	}

	call := args[1].(primitives.Call)

	return c.module.executeCall(primitives.NewRawOriginSigned(who), call, newEventSudoAsDone)
}

func (_ callSudoAs) Docs() string {
	return "Authenticates the sudo key and dispatches a function call with `Signed` origin from a given account." +
		"The dispatch origin for this call must be `Signed`."
}
