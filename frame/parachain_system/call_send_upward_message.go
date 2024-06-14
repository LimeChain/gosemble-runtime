package parachain_system

import (
	"bytes"
	"errors"

	sc "github.com/LimeChain/goscale"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

type callSendUpwardMessage struct {
	primitives.Callable
	module module
}

func newCallSendUpwardMessage(moduleId sc.U8, functionId sc.U8, module module) primitives.Call {
	return callSendUpwardMessage{
		Callable: primitives.Callable{
			ModuleId:   moduleId,
			FunctionId: functionId,
			Arguments:  sc.NewVaryingData(),
		},
		module: module,
	}
}

func (c callSendUpwardMessage) DecodeArgs(buffer *bytes.Buffer) (primitives.Call, error) {
	upwardMessage, err := sc.DecodeSequence[sc.U8](buffer)
	if err != nil {
		return nil, err
	}
	c.Arguments = sc.NewVaryingData(upwardMessage)

	return c, nil
}

func (c callSendUpwardMessage) Encode(buffer *bytes.Buffer) error {
	return c.Callable.Encode(buffer)
}

func (c callSendUpwardMessage) Bytes() []byte {
	return c.Callable.Bytes()
}

func (c callSendUpwardMessage) ModuleIndex() sc.U8 {
	return c.Callable.ModuleIndex()
}

func (c callSendUpwardMessage) FunctionIndex() sc.U8 {
	return c.Callable.FunctionIndex()
}

func (c callSendUpwardMessage) Args() sc.VaryingData {
	return c.Callable.Args()
}

func (c callSendUpwardMessage) BaseWeight() primitives.Weight {
	return primitives.WeightFromParts(1000, 0)
}

func (_ callSendUpwardMessage) WeighData(baseWeight primitives.Weight) primitives.Weight {
	return primitives.WeightFromParts(baseWeight.RefTime, 0)
}

func (_ callSendUpwardMessage) ClassifyDispatch(baseWeight primitives.Weight) primitives.DispatchClass {
	return primitives.NewDispatchClassOperational()
}

func (_ callSendUpwardMessage) PaysFee(baseWeight primitives.Weight) primitives.Pays {
	return primitives.PaysYes
}

func (c callSendUpwardMessage) Dispatch(origin primitives.RuntimeOrigin, args sc.VaryingData) (primitives.PostDispatchInfo, error) {
	if !origin.IsRootOrigin() {
		return primitives.PostDispatchInfo{}, primitives.NewDispatchErrorBadOrigin()
	}

	data, ok := args[0].(sc.Sequence[sc.U8])
	if !ok {
		return primitives.PostDispatchInfo{}, errors.New("couldn't dispatch callSendUpwardMessage value")
	}

	return primitives.PostDispatchInfo{}, c.module.sendUpwardMessage(data)
}

func (_ callSendUpwardMessage) Docs() string {
	return "Sends an upward message."
}
