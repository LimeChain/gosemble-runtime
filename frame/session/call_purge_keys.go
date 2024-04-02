package session

import (
	"bytes"
	sc "github.com/LimeChain/goscale"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

// callPurgeKeys removes any session key(s) of the function caller.
// This doesn't take effect until the next session.
type callPurgeKeys struct {
	primitives.Callable
	module Module
}

func newCallPurgeKeys(moduleId sc.U8, functionId sc.U8, module Module) primitives.Call {
	call := callPurgeKeys{
		Callable: primitives.Callable{
			ModuleId:   moduleId,
			FunctionId: functionId,
			Arguments:  sc.NewVaryingData(sc.Sequence[sc.U8]{}),
		},
		module: module,
	}

	return call
}

func (c callPurgeKeys) DecodeArgs(buffer *bytes.Buffer) (primitives.Call, error) {
	c.Arguments = sc.NewVaryingData()
	return c, nil
}

func (c callPurgeKeys) Encode(buffer *bytes.Buffer) error {
	return c.Callable.Encode(buffer)
}

func (c callPurgeKeys) Bytes() []byte {
	return c.Callable.Bytes()
}

func (c callPurgeKeys) ModuleIndex() sc.U8 {
	return c.Callable.ModuleIndex()
}

func (c callPurgeKeys) FunctionIndex() sc.U8 {
	return c.Callable.FunctionIndex()
}

func (c callPurgeKeys) Args() sc.VaryingData {
	return c.Callable.Args()
}

func (c callPurgeKeys) BaseWeight() primitives.Weight {
	return callPurgeKeysWeight(c.module.config.DbWeight)
}

func (_ callPurgeKeys) WeighData(baseWeight primitives.Weight) primitives.Weight {
	return primitives.WeightFromParts(baseWeight.RefTime, 0)
}

func (_ callPurgeKeys) ClassifyDispatch(baseWeight primitives.Weight) primitives.DispatchClass {
	return primitives.NewDispatchClassNormal()
}

func (_ callPurgeKeys) PaysFee(baseWeight primitives.Weight) primitives.Pays {
	return primitives.PaysYes
}

func (c callPurgeKeys) Dispatch(origin primitives.RuntimeOrigin, args sc.VaryingData) (primitives.PostDispatchInfo, error) {
	if !origin.IsSignedOrigin() {
		return primitives.PostDispatchInfo{}, primitives.NewDispatchErrorBadOrigin()
	}

	who, err := origin.AsSigned()
	if err != nil {
		return primitives.PostDispatchInfo{}, primitives.NewDispatchErrorOther(sc.Str(err.Error()))
	}

	return primitives.PostDispatchInfo{}, c.module.DoPurgeKeys(who)
}

func (_ callPurgeKeys) Docs() string {
	return "Removes any session key(s) of the function caller."
}
