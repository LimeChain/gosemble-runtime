package sudo

import (
	"bytes"
	sc "github.com/LimeChain/goscale"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

// callSetKey authenticates the current sudo key and sets the given `AccountId` as the new sudo key.
type callSetKey struct {
	primitives.Callable
	dbWeight primitives.RuntimeDbWeight
	module   Module
}

func newCallSetKey(moduleId sc.U8, functionId sc.U8, dbWeight primitives.RuntimeDbWeight, module Module) primitives.Call {
	call := callSetKey{
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

func (c callSetKey) DecodeSudoArgs(buffer *bytes.Buffer, _ func(buffer *bytes.Buffer) (primitives.Call, error)) (primitives.Call, error) {
	accountId, err := primitives.DecodeMultiAddress(buffer)
	if err != nil {
		return nil, err
	}
	c.Arguments = sc.NewVaryingData(accountId)

	return c, nil
}

func (c callSetKey) DecodeArgs(_ *bytes.Buffer) (primitives.Call, error) {
	c.module.logger.Critical("not implemented")
	return nil, nil
}

func (c callSetKey) Encode(buffer *bytes.Buffer) error {
	return c.Callable.Encode(buffer)
}

func (c callSetKey) Bytes() []byte {
	return c.Callable.Bytes()
}

func (c callSetKey) ModuleIndex() sc.U8 { return c.Callable.ModuleIndex() }

func (c callSetKey) FunctionIndex() sc.U8 { return c.Callable.FunctionIndex() }

func (c callSetKey) Args() sc.VaryingData { return c.Callable.Args() }

func (c callSetKey) BaseWeight() primitives.Weight {
	return callSetKeyWeight(c.dbWeight)
}

func (_ callSetKey) WeighData(baseWeight primitives.Weight) primitives.Weight {
	return primitives.WeightFromParts(baseWeight.RefTime, 0)
}

func (_ callSetKey) ClassifyDispatch(_ primitives.Weight) primitives.DispatchClass {
	return primitives.NewDispatchClassNormal()
}

func (_ callSetKey) PaysFee(_ primitives.Weight) primitives.Pays {
	return primitives.PaysNo
}

func (c callSetKey) Dispatch(origin primitives.RuntimeOrigin, args sc.VaryingData) (primitives.PostDispatchInfo, error) {
	err := c.module.ensureSudo(origin)
	if err != nil {
		return primitives.PostDispatchInfo{}, err
	}

	multiAddress := args[0].(primitives.MultiAddress)

	newKey, err := primitives.Lookup(multiAddress)
	if err != nil {
		c.module.logger.Debugf("Failed to lookup [%s]", multiAddress.Bytes())
		return primitives.PostDispatchInfo{}, primitives.NewDispatchErrorCannotLookup()
	}
	oldKey, err := c.module.storage.Key.Get()
	if err != nil {
		return primitives.PostDispatchInfo{}, primitives.NewDispatchErrorOther(sc.Str(err.Error()))
	}

	c.module.eventDepositor.DepositEvent(newEventKeyChanged(c.ModuleId, sc.NewOption[primitives.AccountId](oldKey), newKey))
	c.module.storage.Key.Put(newKey)

	return primitives.PostDispatchInfo{
		PaysFee: primitives.PaysNo,
	}, nil
}

func (_ callSetKey) Docs() string {
	return "Authenticates the current sudo key and sets the given `AccountId` as the new sudo."
}
