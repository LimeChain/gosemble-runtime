package grandpa

import (
	"bytes"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/frame/system"
	grandpatypes "github.com/LimeChain/gosemble/primitives/grandpa"
	"github.com/LimeChain/gosemble/primitives/session"
	staking "github.com/LimeChain/gosemble/primitives/staking"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

// Report voter equivocation/misbehavior. This method will verify the
// equivocation proof and validate the given key ownership proof
// against the extracted offender. If both are valid, the offence
// will be reported.
//
// This extrinsic must be called unsigned and it is expected that only
// block authors will call it (validated in `ValidateUnsigned`), as such
// if the block author is defined it will be defined as the equivocation
// reporter.
type callReportEquivocationUnsigned struct {
	primitives.Callable
	dbWeight            primitives.RuntimeDbWeight
	offenceReportSystem staking.OffenceReportSystem
}

func newCallReportEquivocationUnsigned(moduleId sc.U8, functionId sc.U8, dbWeight primitives.RuntimeDbWeight) primitives.Call {
	call := callReportEquivocationUnsigned{
		Callable: primitives.Callable{
			ModuleId:   moduleId,
			FunctionId: functionId,
			Arguments:  sc.NewVaryingData(),
		},
		dbWeight: dbWeight,
	}

	return call
}

func (c callReportEquivocationUnsigned) DecodeArgs(buffer *bytes.Buffer) (primitives.Call, error) {
	equivocationProof, err := grandpatypes.DecodeEquivocationProof(buffer)
	if err != nil {
		return nil, err
	}

	keyOwnerProof, err := session.DecodeMembershipProof(buffer)
	if err != nil {
		return nil, err
	}

	c.Arguments = sc.NewVaryingData(equivocationProof, keyOwnerProof)
	return c, nil
}

func (c callReportEquivocationUnsigned) Encode(buffer *bytes.Buffer) error {
	return c.Callable.Encode(buffer)
}

func (c callReportEquivocationUnsigned) Bytes() []byte {
	return c.Callable.Bytes()
}

func (c callReportEquivocationUnsigned) ModuleIndex() sc.U8 {
	return c.Callable.ModuleIndex()
}

func (c callReportEquivocationUnsigned) FunctionIndex() sc.U8 {
	return c.Callable.FunctionIndex()
}

func (c callReportEquivocationUnsigned) Args() sc.VaryingData {
	return c.Callable.Args()
}

func (c callReportEquivocationUnsigned) BaseWeight() primitives.Weight {
	return callReportEquivocationUnsignedWeight(c.dbWeight)
}

func (_ callReportEquivocationUnsigned) WeighData(baseWeight primitives.Weight) primitives.Weight {
	return primitives.WeightFromParts(baseWeight.RefTime, 0)
}

func (_ callReportEquivocationUnsigned) ClassifyDispatch(baseWeight primitives.Weight) primitives.DispatchClass {
	return primitives.NewDispatchClassNormal()
}

func (_ callReportEquivocationUnsigned) PaysFee(baseWeight primitives.Weight) primitives.Pays {
	return primitives.PaysYes
}

func (c callReportEquivocationUnsigned) Dispatch(origin primitives.RuntimeOrigin, args sc.VaryingData) (primitives.PostDispatchInfo, error) {
	_, err := system.EnsureNone(origin)
	if err != nil {
		return primitives.PostDispatchInfo{}, err
	}

	equivocationProof := args[0].(grandpatypes.EquivocationProof)
	keyOwnerProof := args[1].(grandpatypes.KeyOwnerProof)

	err = c.offenceReportSystem.ProcessEvidence(sc.NewOption[primitives.AccountId](nil), equivocationProof, keyOwnerProof)
	if err != nil {
		return primitives.PostDispatchInfo{}, err
	}

	// Waive the fee since the report is valid and beneficial
	return primitives.PostDispatchInfo{PaysFee: primitives.PaysNo}, nil
}

func (_ callReportEquivocationUnsigned) Docs() string {
	return ""
}
