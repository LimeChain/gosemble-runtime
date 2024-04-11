package sudo

import (
	"bytes"
	sc "github.com/LimeChain/goscale"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

// callRemoveKey authenticates the current sudo key and sets the given `AccountId` as the new sudo key.
type callRemoveKey struct {
	primitives.Callable
	dbWeight primitives.RuntimeDbWeight
	module   Module
}

func newCallRemoveKey(moduleId sc.U8, functionId sc.U8, dbWeight primitives.RuntimeDbWeight, module Module) primitives.Call {
	call := callRemoveKey{
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

func (c callRemoveKey) DecodeSudoArgs(_ *bytes.Buffer, _ func(buffer *bytes.Buffer) (primitives.Call, error)) (primitives.Call, error) {
	c.Arguments = sc.NewVaryingData()

	return c, nil
}

func (c callRemoveKey) DecodeArgs(_ *bytes.Buffer) (primitives.Call, error) {
	c.module.logger.Critical("not implemented")
	return nil, nil
}

func (c callRemoveKey) Encode(buffer *bytes.Buffer) error {
	return c.Callable.Encode(buffer)
}

func (c callRemoveKey) Bytes() []byte {
	return c.Callable.Bytes()
}

func (_ callRemoveKey) Docs() string {
	return "Permanently removes the sudo key. This cannot be undone."
}

func (c callRemoveKey) ModuleIndex() sc.U8 { return c.Callable.ModuleIndex() }

func (c callRemoveKey) FunctionIndex() sc.U8 { return c.Callable.FunctionIndex() }

func (c callRemoveKey) Args() sc.VaryingData { return c.Callable.Args() }

func (c callRemoveKey) BaseWeight() primitives.Weight {
	return callRemoveKeyWeight(c.dbWeight)
}

func (_ callRemoveKey) WeighData(baseWeight primitives.Weight) primitives.Weight {
	return primitives.WeightFromParts(baseWeight.RefTime, 0)
}

func (_ callRemoveKey) ClassifyDispatch(_ primitives.Weight) primitives.DispatchClass {
	return primitives.NewDispatchClassNormal()
}

func (_ callRemoveKey) PaysFee(_ primitives.Weight) primitives.Pays {
	return primitives.PaysNo
}

func (c callRemoveKey) Dispatch(origin primitives.RuntimeOrigin, _ sc.VaryingData) (primitives.PostDispatchInfo, error) {
	err := c.module.ensureSudo(origin)
	if err != nil {
		return primitives.PostDispatchInfo{}, err
	}

	c.module.eventDepositor.DepositEvent(newEventKeyRemoved(c.ModuleId))
	c.module.storage.Key.Clear()

	return primitives.PostDispatchInfo{
		ActualWeight: sc.Option[primitives.Weight]{},
		PaysFee:      primitives.PaysNo,
	}, nil
}
