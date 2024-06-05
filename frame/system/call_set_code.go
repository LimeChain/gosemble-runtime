package system

import (
	"bytes"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/hooks"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

// Set the new runtime code.
type callSetCode struct {
	primitives.Callable
	constants     consts
	dbWeight      primitives.RuntimeDbWeight
	hookOnSetCode hooks.OnSetCode
	codeUpgrader  CodeUpgrader
}

func newCallSetCode(
	moduleId sc.U8,
	functionId sc.U8,
	dbWeight primitives.RuntimeDbWeight,
	constants consts,
	hookOnSetCode hooks.OnSetCode,
	codeUpgrader CodeUpgrader,
) primitives.Call {
	call := callSetCode{
		Callable: primitives.Callable{
			ModuleId:   moduleId,
			FunctionId: functionId,
			Arguments:  sc.NewVaryingData(sc.Sequence[sc.U8]{}),
		},
		constants:     constants,
		dbWeight:      dbWeight,
		hookOnSetCode: hookOnSetCode,
		codeUpgrader:  codeUpgrader,
	}

	return call
}

func (c callSetCode) DecodeArgs(buffer *bytes.Buffer) (primitives.Call, error) {
	codeBlob, err := sc.DecodeSequence[sc.U8](buffer)
	if err != nil {
		return nil, err
	}
	c.Arguments = sc.NewVaryingData(codeBlob)
	return c, nil
}

func (c callSetCode) Encode(buffer *bytes.Buffer) error {
	return c.Callable.Encode(buffer)
}

func (c callSetCode) Bytes() []byte {
	return c.Callable.Bytes()
}

func (c callSetCode) ModuleIndex() sc.U8 {
	return c.Callable.ModuleIndex()
}

func (c callSetCode) FunctionIndex() sc.U8 {
	return c.Callable.FunctionIndex()
}

func (c callSetCode) Args() sc.VaryingData {
	return c.Callable.Args()
}

func (c callSetCode) BaseWeight() primitives.Weight {
	return callSetCodeWeight(c.dbWeight)
}

func (_ callSetCode) WeighData(baseWeight primitives.Weight) primitives.Weight {
	return primitives.WeightFromParts(baseWeight.RefTime, 0)
}

func (_ callSetCode) ClassifyDispatch(baseWeight primitives.Weight) primitives.DispatchClass {
	return primitives.NewDispatchClassOperational()
}

func (_ callSetCode) PaysFee(baseWeight primitives.Weight) primitives.Pays {
	return primitives.PaysNo
}

func (c callSetCode) Dispatch(origin primitives.RuntimeOrigin, args sc.VaryingData) (primitives.PostDispatchInfo, error) {
	// TODO: enable once 'sudo' module is implemented
	//
	// err := EnsureRoot(origin)
	// if err != nil {
	// 	return primitives.PostDispatchInfo{}, err
	// }

	codeBlob := args[0].(sc.Sequence[sc.U8])

	err := c.codeUpgrader.CanSetCode(codeBlob)
	if err != nil {
		return primitives.PostDispatchInfo{}, err
	}

	err = c.hookOnSetCode.SetCode(codeBlob)
	if err != nil {
		return primitives.PostDispatchInfo{}, err
	}

	// consume the rest of the block to prevent further transactions
	return primitives.PostDispatchInfo{
		ActualWeight: sc.NewOption[primitives.Weight](c.constants.BlockWeights.MaxBlock),
		PaysFee:      primitives.PaysNo,
	}, nil
}

func (_ callSetCode) Docs() string {
	return "Set the new runtime code."
}
