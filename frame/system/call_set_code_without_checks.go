package system

import (
	"bytes"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/hooks"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

// Set the new runtime code without doing any checks of the given `code`.
type callSetCodeWithoutChecks struct {
	primitives.Callable
	constants     consts
	dbWeight      primitives.RuntimeDbWeight
	hookOnSetCode hooks.OnSetCode
}

func newCallSetCodeWithoutChecks(
	moduleId sc.U8,
	functionId sc.U8,
	dbWeight primitives.RuntimeDbWeight,
	constants consts,
	hookOnSetCode hooks.OnSetCode,
) primitives.Call {
	call := callSetCodeWithoutChecks{
		Callable: primitives.Callable{
			ModuleId:   moduleId,
			FunctionId: functionId,
			Arguments:  sc.NewVaryingData(sc.Sequence[sc.U8]{}),
		},
		constants:     constants,
		dbWeight:      dbWeight,
		hookOnSetCode: hookOnSetCode,
	}

	return call
}

func (c callSetCodeWithoutChecks) DecodeArgs(buffer *bytes.Buffer) (primitives.Call, error) {
	codeBlob, err := sc.DecodeSequence[sc.U8](buffer)
	if err != nil {
		return nil, err
	}
	c.Arguments = sc.NewVaryingData(codeBlob)
	return c, nil
}

func (c callSetCodeWithoutChecks) Encode(buffer *bytes.Buffer) error {
	return c.Callable.Encode(buffer)
}

func (c callSetCodeWithoutChecks) Bytes() []byte {
	return c.Callable.Bytes()
}

func (c callSetCodeWithoutChecks) ModuleIndex() sc.U8 {
	return c.Callable.ModuleIndex()
}

func (c callSetCodeWithoutChecks) FunctionIndex() sc.U8 {
	return c.Callable.FunctionIndex()
}

func (c callSetCodeWithoutChecks) Args() sc.VaryingData {
	return c.Callable.Args()
}

func (c callSetCodeWithoutChecks) BaseWeight() primitives.Weight {
	return callSetCodeWithoutChecksWeight(c.dbWeight)
}

func (_ callSetCodeWithoutChecks) WeighData(baseWeight primitives.Weight) primitives.Weight {
	return primitives.WeightFromParts(baseWeight.RefTime, 0)
}

func (_ callSetCodeWithoutChecks) ClassifyDispatch(baseWeight primitives.Weight) primitives.DispatchClass {
	return primitives.NewDispatchClassOperational()
}

func (_ callSetCodeWithoutChecks) PaysFee(baseWeight primitives.Weight) primitives.Pays {
	return primitives.PaysNo
}

func (c callSetCodeWithoutChecks) Dispatch(origin primitives.RuntimeOrigin, args sc.VaryingData) (primitives.PostDispatchInfo, error) {
	// TODO: enable once 'sudo' module is implemented
	//
	// err := EnsureRoot(origin)
	// if err != nil {
	// 	return primitives.PostDispatchInfo{}, err
	// }

	codeBlob := args[0].(sc.Sequence[sc.U8])

	err := c.hookOnSetCode.SetCode(codeBlob)
	if err != nil {
		return primitives.PostDispatchInfo{}, err
	}

	// consume the rest of the block to prevent further transactions
	return primitives.PostDispatchInfo{
		ActualWeight: sc.NewOption[primitives.Weight](c.constants.BlockWeights.MaxBlock),
		PaysFee:      primitives.PaysNo,
	}, nil
}

func (_ callSetCodeWithoutChecks) Docs() string {
	return "Set the new runtime code without any checks."
}
