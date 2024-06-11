package balances

import (
	"bytes"

	sc "github.com/LimeChain/goscale"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

type callForceUnreserve struct {
	primitives.Callable
	module Module
}

func newCallForceUnreserve(moduleId sc.U8, functionId sc.U8, module Module) primitives.Call {
	return callForceUnreserve{
		Callable: primitives.Callable{
			ModuleId:   moduleId,
			FunctionId: functionId,
			Arguments:  sc.NewVaryingData(primitives.MultiAddress{}, sc.U128{}),
		},
		module: module,
	}
}

func (c callForceUnreserve) DecodeArgs(buffer *bytes.Buffer) (primitives.Call, error) {
	who, err := primitives.DecodeMultiAddress(buffer)
	if err != nil {
		return nil, err
	}
	value, err := sc.DecodeU128(buffer)
	if err != nil {
		return nil, err
	}

	c.Arguments = sc.NewVaryingData(
		who,
		value,
	)

	return c, nil
}

func (c callForceUnreserve) Encode(buffer *bytes.Buffer) error {
	return c.Callable.Encode(buffer)
}

func (c callForceUnreserve) Bytes() []byte {
	return c.Callable.Bytes()
}

func (c callForceUnreserve) ModuleIndex() sc.U8 {
	return c.Callable.ModuleIndex()
}

func (c callForceUnreserve) FunctionIndex() sc.U8 {
	return c.Callable.FunctionIndex()
}

func (c callForceUnreserve) Args() sc.VaryingData {
	return c.Callable.Args()
}

func (c callForceUnreserve) BaseWeight() primitives.Weight {
	return callForceUnreserveWeight(c.module.constants.DbWeight)
}

func (_ callForceUnreserve) WeighData(baseWeight primitives.Weight) primitives.Weight {
	return primitives.WeightFromParts(baseWeight.RefTime, 0)
}

func (_ callForceUnreserve) ClassifyDispatch(baseWeight primitives.Weight) primitives.DispatchClass {
	return primitives.NewDispatchClassNormal()
}

func (_ callForceUnreserve) PaysFee(baseWeight primitives.Weight) primitives.Pays {
	return primitives.PaysYes
}

func (_ callForceUnreserve) Docs() string {
	return "Unreserve some balance from a user by force."
}

func (c callForceUnreserve) Dispatch(origin primitives.RuntimeOrigin, args sc.VaryingData) (primitives.PostDispatchInfo, error) {
	if !origin.IsRootOrigin() {
		return primitives.PostDispatchInfo{}, primitives.NewDispatchErrorBadOrigin()
	}

	whoMultiAddress, ok := args[0].(primitives.MultiAddress)
	if !ok {
		return primitives.PostDispatchInfo{}, primitives.NewDispatchErrorOther("invalid destination value in callForceUnreserve")
	}

	who, err := primitives.Lookup(whoMultiAddress)
	if err != nil {
		return primitives.PostDispatchInfo{}, primitives.NewDispatchErrorCannotLookup()
	}

	value, ok := args[1].(sc.U128)
	if !ok {
		return primitives.PostDispatchInfo{}, primitives.NewDispatchErrorOther("invalid u128 value in callForceUnreserve")
	}

	_, err = c.module.unreserve(who, value)
	if err != nil {
		return primitives.PostDispatchInfo{}, err
	}

	return primitives.PostDispatchInfo{}, nil
}
