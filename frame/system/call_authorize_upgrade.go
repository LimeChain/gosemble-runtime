package system

import (
	"bytes"

	sc "github.com/LimeChain/goscale"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

// Authorize new runtime code.
type callAuthorizeUpgrade struct {
	primitives.Callable
	dbWeight     primitives.RuntimeDbWeight
	codeUpgrader CodeUpgrader
}

func newCallAuthorizeUpgrade(moduleId sc.U8, functionId sc.U8, dbWeight primitives.RuntimeDbWeight, codeUpgrader CodeUpgrader) primitives.Call {
	call := callAuthorizeUpgrade{
		Callable: primitives.Callable{
			ModuleId:   moduleId,
			FunctionId: functionId,
			Arguments:  sc.NewVaryingData(primitives.H256{}),
		},
		dbWeight:     dbWeight,
		codeUpgrader: codeUpgrader,
	}

	return call
}

func (c callAuthorizeUpgrade) DecodeArgs(buffer *bytes.Buffer) (primitives.Call, error) {
	codeHash, err := primitives.DecodeH256(buffer)
	if err != nil {
		return nil, err
	}
	c.Arguments = sc.NewVaryingData(codeHash)
	return c, nil
}

func (c callAuthorizeUpgrade) Encode(buffer *bytes.Buffer) error {
	return c.Callable.Encode(buffer)
}

func (c callAuthorizeUpgrade) Bytes() []byte {
	return c.Callable.Bytes()
}

func (c callAuthorizeUpgrade) ModuleIndex() sc.U8 {
	return c.Callable.ModuleIndex()
}

func (c callAuthorizeUpgrade) FunctionIndex() sc.U8 {
	return c.Callable.FunctionIndex()
}

func (c callAuthorizeUpgrade) Args() sc.VaryingData {
	return c.Callable.Args()
}

func (c callAuthorizeUpgrade) BaseWeight() primitives.Weight {
	return callAuthorizeUpgradeWeight(c.dbWeight)
}

func (_ callAuthorizeUpgrade) WeighData(baseWeight primitives.Weight) primitives.Weight {
	return primitives.WeightFromParts(baseWeight.RefTime, 0)
}

func (_ callAuthorizeUpgrade) ClassifyDispatch(baseWeight primitives.Weight) primitives.DispatchClass {
	return primitives.NewDispatchClassOperational()
}

func (_ callAuthorizeUpgrade) PaysFee(baseWeight primitives.Weight) primitives.Pays {
	return primitives.PaysNo
}

func (c callAuthorizeUpgrade) Dispatch(origin primitives.RuntimeOrigin, args sc.VaryingData) (primitives.PostDispatchInfo, error) {
	// TODO: enable once 'sudo' module is implemented
	//
	// err := EnsureRoot(origin)
	// if err != nil {
	// 	return err
	// }

	codeHash := args[0].(primitives.H256)

	c.codeUpgrader.DoAuthorizeUpgrade(codeHash, true)

	return primitives.PostDispatchInfo{}, nil
}

func (_ callAuthorizeUpgrade) Docs() string {
	return "Authorize new runtime code."
}
