package session

import (
	"bytes"
	sc "github.com/LimeChain/goscale"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

// callSetKeys sets the session keys of thee function caller to `keys`.
// Can be executed by any origin.
type callSetKeys struct {
	primitives.Callable
	module Module
}

func newCallSetKeys(moduleId sc.U8, functionId sc.U8, module Module) primitives.Call {
	call := callSetKeys{
		Callable: primitives.Callable{
			ModuleId:   moduleId,
			FunctionId: functionId,
			Arguments:  sc.NewVaryingData(sc.Sequence[sc.U8]{}),
		},
		module: module,
	}

	return call
}

func (c callSetKeys) DecodeArgs(buffer *bytes.Buffer) (primitives.Call, error) {
	keys, err := c.module.handler.DecodeKeys(buffer)
	if err != nil {
		return nil, err
	}
	proof, err := sc.DecodeSequence[sc.U8](buffer)
	if err != nil {
		return nil, err
	}

	c.Arguments = sc.NewVaryingData(keys, proof)
	return c, nil
}

func (c callSetKeys) Encode(buffer *bytes.Buffer) error {
	return c.Callable.Encode(buffer)
}

func (c callSetKeys) Bytes() []byte {
	return c.Callable.Bytes()
}

func (c callSetKeys) ModuleIndex() sc.U8 {
	return c.Callable.ModuleIndex()
}

func (c callSetKeys) FunctionIndex() sc.U8 {
	return c.Callable.FunctionIndex()
}

func (c callSetKeys) Args() sc.VaryingData {
	return c.Callable.Args()
}

func (c callSetKeys) BaseWeight() primitives.Weight {
	return callSetKeysWeight(c.module.config.DbWeight)
}

func (_ callSetKeys) WeighData(baseWeight primitives.Weight) primitives.Weight {
	return primitives.WeightFromParts(baseWeight.RefTime, 0)
}

func (_ callSetKeys) ClassifyDispatch(baseWeight primitives.Weight) primitives.DispatchClass {
	return primitives.NewDispatchClassNormal()
}

func (_ callSetKeys) PaysFee(baseWeight primitives.Weight) primitives.Pays {
	return primitives.PaysYes
}

func (c callSetKeys) Dispatch(origin primitives.RuntimeOrigin, args sc.VaryingData) (primitives.PostDispatchInfo, error) {
	if !origin.IsSignedOrigin() {
		return primitives.PostDispatchInfo{}, primitives.NewDispatchErrorBadOrigin()
	}

	who, err := origin.AsSigned()
	if err != nil {
		return primitives.PostDispatchInfo{}, primitives.NewDispatchErrorOther(sc.Str(err.Error()))
	}
	publicKeys := args[0].(sc.FixedSequence[primitives.Sr25519PublicKey])

	keys, err := toSessionKeys(c.module.handler.KeyTypeIds(), publicKeys)
	if err != nil {
		return primitives.PostDispatchInfo{}, primitives.NewDispatchErrorOther(sc.Str(err.Error()))
	}

	return primitives.PostDispatchInfo{}, c.module.DoSetKeys(who, keys)
}

func (_ callSetKeys) Docs() string {
	return "Sets the session key(s) of the function caller to `keys`. " +
		"Allows an account to set its session key prior to becoming a validator." +
		"This doesn't take effect until the next session."
}
