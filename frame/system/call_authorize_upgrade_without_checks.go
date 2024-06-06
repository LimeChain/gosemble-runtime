package system

import (
	"bytes"

	sc "github.com/LimeChain/goscale"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

// Authorize new runtime code and an upgrade sans verification.
type callAuthorizeUpgradeWithoutChecks struct {
	primitives.Callable
	dbWeight     primitives.RuntimeDbWeight
	codeUpgrader CodeUpgrader
}

func newCallAuthorizeUpgradeWithoutChecks(moduleId sc.U8, functionId sc.U8, dbWeight primitives.RuntimeDbWeight, codeUpgrader CodeUpgrader) primitives.Call {
	call := callAuthorizeUpgradeWithoutChecks{
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

func (c callAuthorizeUpgradeWithoutChecks) DecodeArgs(buffer *bytes.Buffer) (primitives.Call, error) {
	codeHash, err := primitives.DecodeH256(buffer)
	if err != nil {
		return nil, err
	}
	c.Arguments = sc.NewVaryingData(codeHash)
	return c, nil
}

func (c callAuthorizeUpgradeWithoutChecks) Encode(buffer *bytes.Buffer) error {
	return c.Callable.Encode(buffer)
}

func (c callAuthorizeUpgradeWithoutChecks) Bytes() []byte {
	return c.Callable.Bytes()
}

func (c callAuthorizeUpgradeWithoutChecks) ModuleIndex() sc.U8 {
	return c.Callable.ModuleIndex()
}

func (c callAuthorizeUpgradeWithoutChecks) FunctionIndex() sc.U8 {
	return c.Callable.FunctionIndex()
}

func (c callAuthorizeUpgradeWithoutChecks) Args() sc.VaryingData {
	return c.Callable.Args()
}

func (c callAuthorizeUpgradeWithoutChecks) BaseWeight() primitives.Weight {
	return callAuthorizeUpgradeWithoutChecksWeight(c.dbWeight)
}

func (_ callAuthorizeUpgradeWithoutChecks) WeighData(baseWeight primitives.Weight) primitives.Weight {
	return primitives.WeightFromParts(baseWeight.RefTime, 0)
}

func (_ callAuthorizeUpgradeWithoutChecks) ClassifyDispatch(baseWeight primitives.Weight) primitives.DispatchClass {
	return primitives.NewDispatchClassOperational()
}

func (_ callAuthorizeUpgradeWithoutChecks) PaysFee(baseWeight primitives.Weight) primitives.Pays {
	return primitives.PaysNo
}

func (c callAuthorizeUpgradeWithoutChecks) Dispatch(origin primitives.RuntimeOrigin, args sc.VaryingData) (primitives.PostDispatchInfo, error) {
	// TODO: enable once 'sudo' module is implemented
	//
	// err := EnsureRoot(origin)
	// if err != nil {
	// 	return err
	// }

	codeHash := args[0].(primitives.H256)

	c.codeUpgrader.DoAuthorizeUpgrade(codeHash, false)

	return primitives.PostDispatchInfo{}, nil
}

func (_ callAuthorizeUpgradeWithoutChecks) Docs() string {
	return "Authorize new runtime code and an upgrade sans verification."
}
