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
type callReportEquivocation struct {
	primitives.Callable
	offenceReportSystem staking.OffenceReportSystem
}

func newCallReportEquivocation(moduleId sc.U8, functionId sc.U8) primitives.Call {
	call := callReportEquivocation{
		Callable: primitives.Callable{
			ModuleId:   moduleId,
			FunctionId: functionId,
			Arguments:  sc.NewVaryingData(),
		},
	}

	return call
}

func (c callReportEquivocation) DecodeArgs(buffer *bytes.Buffer) (primitives.Call, error) {
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

func (c callReportEquivocation) Encode(buffer *bytes.Buffer) error {
	return c.Callable.Encode(buffer)
}

func (c callReportEquivocation) Bytes() []byte {
	return c.Callable.Bytes()
}

func (c callReportEquivocation) ModuleIndex() sc.U8 {
	return c.Callable.ModuleIndex()
}

func (c callReportEquivocation) FunctionIndex() sc.U8 {
	return c.Callable.FunctionIndex()
}

func (c callReportEquivocation) Args() sc.VaryingData {
	return c.Callable.Args()
}

func (c callReportEquivocation) BaseWeight() primitives.Weight {
	return callReportEquivocationWeight(primitives.RuntimeDbWeight{})
}

func (_ callReportEquivocation) WeighData(baseWeight primitives.Weight) primitives.Weight {
	return primitives.WeightFromParts(baseWeight.RefTime, 0)
}

func (_ callReportEquivocation) ClassifyDispatch(baseWeight primitives.Weight) primitives.DispatchClass {
	return primitives.NewDispatchClassNormal()
}

func (_ callReportEquivocation) PaysFee(baseWeight primitives.Weight) primitives.Pays {
	return primitives.PaysYes
}

func (c callReportEquivocation) Dispatch(origin primitives.RuntimeOrigin, args sc.VaryingData) (primitives.PostDispatchInfo, error) {
	reporter, err := system.EnsureSigned(origin)
	if err != nil {
		return primitives.PostDispatchInfo{}, err
	}

	equivocationProof := args[0].(grandpatypes.EquivocationProof)
	keyOwnerProof := args[1].(grandpatypes.KeyOwnerProof)

	err = c.offenceReportSystem.ProcessEvidence(reporter, equivocationProof, keyOwnerProof)
	if err != nil {
		return primitives.PostDispatchInfo{}, err
	}

	// Waive the fee since the report is valid and beneficial
	return primitives.PostDispatchInfo{PaysFee: primitives.PaysNo}, nil
}

func (_ callReportEquivocation) Docs() string {
	return ""
}
