package balances

import (
	"bytes"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/frame/balances/types"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

type callTransferAll struct {
	primitives.Callable
	module Module
}

func newCallTransferAll(moduleId sc.U8, functionId sc.U8, module Module) primitives.Call {
	return callTransferAll{
		Callable: primitives.Callable{
			ModuleId:   moduleId,
			FunctionId: functionId,
			Arguments:  sc.NewVaryingData(primitives.MultiAddress{}, sc.Bool(true)),
		},
		module: module,
	}
}

func (c callTransferAll) DecodeArgs(buffer *bytes.Buffer) (primitives.Call, error) {
	dest, err := primitives.DecodeMultiAddress(buffer)
	if err != nil {
		return nil, err
	}

	keepAlive, err := sc.DecodeBool(buffer)
	if err != nil {
		return nil, err
	}

	c.Arguments = sc.NewVaryingData(
		dest,
		keepAlive,
	)

	return c, nil
}

func (c callTransferAll) Encode(buffer *bytes.Buffer) error {
	return c.Callable.Encode(buffer)
}

func (c callTransferAll) Bytes() []byte {
	return c.Callable.Bytes()
}

func (c callTransferAll) ModuleIndex() sc.U8 {
	return c.Callable.ModuleIndex()
}

func (c callTransferAll) FunctionIndex() sc.U8 {
	return c.Callable.FunctionIndex()
}

func (c callTransferAll) Args() sc.VaryingData {
	return c.Callable.Args()
}

func (c callTransferAll) BaseWeight() primitives.Weight {
	return callTransferAllWeight(c.module.constants.DbWeight)
}

func (_ callTransferAll) WeighData(baseWeight primitives.Weight) primitives.Weight {
	return primitives.WeightFromParts(baseWeight.RefTime, 0)
}

func (_ callTransferAll) ClassifyDispatch(baseWeight primitives.Weight) primitives.DispatchClass {
	return primitives.NewDispatchClassNormal()
}

func (_ callTransferAll) PaysFee(baseWeight primitives.Weight) primitives.Pays {
	return primitives.PaysYes
}

func (_ callTransferAll) Docs() string {
	return "Transfer the entire transferable balance from the caller account." +
		" NOTE: This function only attempts to transfer _transferable_ balances. This means that " +
		"any locked, reserved, or existential deposits (when `keep_alive` is `true`), will not be " +
		"transferred by this function. To ensure that this function results in a killed account," +
		" you might need to prepare the account by removing any reference counters, storage " +
		"deposits, etc... " +
		"The dispatch origin of this call must be Signed. " +
		"- `dest`: The recipient of the transfer. " +
		"- `keep_alive`: A boolean to determine if the `transfer_all` operation should send all " +
		"of the funds the account has, causing the sender account to be killed (false), or " +
		"transfer everything except at least the existential deposit, which will guarantee to keep the sender account alive (true)."
}

func (c callTransferAll) Dispatch(origin primitives.RuntimeOrigin, args sc.VaryingData) (primitives.PostDispatchInfo, error) {
	if !origin.IsSignedOrigin() {
		return primitives.PostDispatchInfo{}, primitives.NewDispatchErrorBadOrigin()
	}

	from, originErr := origin.AsSigned()
	if originErr != nil {
		return primitives.PostDispatchInfo{}, primitives.NewDispatchErrorOther(sc.Str(originErr.Error()))
	}

	dest, ok := args[0].(primitives.MultiAddress)
	if !ok {
		return primitives.PostDispatchInfo{}, primitives.NewDispatchErrorOther("invalid destination value in callTransferAll")
	}

	to, err := primitives.Lookup(dest)
	if err != nil {
		return primitives.PostDispatchInfo{}, primitives.NewDispatchErrorCannotLookup()
	}

	keepAlive, ok := args[1].(sc.Bool)
	if !ok {
		return primitives.PostDispatchInfo{}, primitives.NewDispatchErrorOther("invalid keepAlive value in callTransferAll")
	}

	preservation := types.PreservationExpendable
	if keepAlive {
		preservation = types.PreservationPreserve
	}

	reducibleBalance, err := c.module.reducibleBalance(from, preservation, types.FortitudePolite)
	if err != nil {
		return primitives.PostDispatchInfo{}, err
	}

	return primitives.PostDispatchInfo{}, c.module.transfer(from, to, reducibleBalance, preservation)
}
