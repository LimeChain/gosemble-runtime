package system

import (
	"bytes"

	sc "github.com/LimeChain/goscale"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

// Provide the preimage (runtime binary) `code` for an upgrade that has been authorized.
//
// If the authorization required a version check, this call will ensure the spec name
// remains unchanged and that the spec version has increased.
//
// Depending on the runtime's `OnSetCode` configuration, this function may directly apply
// the new `code` in the same block or attempt to schedule the upgrade.
//
// All origins are allowed.
type callApplyAuthorizedUpgrade struct {
	primitives.Callable
	dbWeight     primitives.RuntimeDbWeight
	codeUpgrader CodeUpgrader
}

func newCallApplyAuthorizedUpgrade(moduleId sc.U8, functionId sc.U8, dbWeight primitives.RuntimeDbWeight, codeUpgrader CodeUpgrader) primitives.Call {
	call := callApplyAuthorizedUpgrade{
		Callable: primitives.Callable{
			ModuleId:   moduleId,
			FunctionId: functionId,
			Arguments:  sc.NewVaryingData(sc.Sequence[sc.U8]{}),
		},
		dbWeight:     dbWeight,
		codeUpgrader: codeUpgrader,
	}

	return call
}

func (c callApplyAuthorizedUpgrade) DecodeArgs(buffer *bytes.Buffer) (primitives.Call, error) {
	codeBlob, err := sc.DecodeSequence[sc.U8](buffer)
	if err != nil {
		return nil, err
	}
	c.Arguments = sc.NewVaryingData(codeBlob)
	return c, nil
}

func (c callApplyAuthorizedUpgrade) Encode(buffer *bytes.Buffer) error {
	return c.Callable.Encode(buffer)
}

func (c callApplyAuthorizedUpgrade) Bytes() []byte {
	return c.Callable.Bytes()
}

func (c callApplyAuthorizedUpgrade) ModuleIndex() sc.U8 {
	return c.Callable.ModuleIndex()
}

func (c callApplyAuthorizedUpgrade) FunctionIndex() sc.U8 {
	return c.Callable.FunctionIndex()
}

func (c callApplyAuthorizedUpgrade) Args() sc.VaryingData {
	return c.Callable.Args()
}

func (c callApplyAuthorizedUpgrade) BaseWeight() primitives.Weight {
	return callApplyAuthorizedUpgradeWeight(c.dbWeight)
}

func (_ callApplyAuthorizedUpgrade) WeighData(baseWeight primitives.Weight) primitives.Weight {
	return primitives.WeightFromParts(baseWeight.RefTime, 0)
}

func (_ callApplyAuthorizedUpgrade) ClassifyDispatch(baseWeight primitives.Weight) primitives.DispatchClass {
	return primitives.NewDispatchClassOperational()
}

func (_ callApplyAuthorizedUpgrade) PaysFee(baseWeight primitives.Weight) primitives.Pays {
	return primitives.PaysNo
}

func (c callApplyAuthorizedUpgrade) Dispatch(origin primitives.RuntimeOrigin, args sc.VaryingData) (primitives.PostDispatchInfo, error) {
	codeBlob := args[0].(sc.Sequence[sc.U8])

	return c.codeUpgrader.DoApplyAuthorizeUpgrade(codeBlob)
}

func (_ callApplyAuthorizedUpgrade) Docs() string {
	return "Provide new, already-authorized runtime code."
}
