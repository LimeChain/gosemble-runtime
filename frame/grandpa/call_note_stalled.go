package grandpa

import (
	"bytes"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/frame/system"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

// This will trigger a forced authority set change at the beginning of the next session, to
// be enacted `delay` blocks after that. The `delay` should be high enough to safely assume
// that the block signalling the forced change will not be re-orged e.g. 1000 blocks.
// The block production rate (which may be slowed down because of finality lagging) should
// be taken into account when choosing the `delay`. The GRANDPA voters based on the new
// authority will start voting on top of `best_finalized_block_number` for new finalized
// blocks. `best_finalized_block_number` should be the highest of the latest finalized
// block of all validators of the new authority set.
//
// Only callable by root.
type callNoteStalled struct {
	primitives.Callable
	staleNotifier StaleNotifier
}

func newCallNoteStalled(moduleId sc.U8, functionId sc.U8, staleNotifier StaleNotifier) primitives.Call {
	call := callNoteStalled{
		Callable: primitives.Callable{
			ModuleId:   moduleId,
			FunctionId: functionId,
			Arguments:  sc.NewVaryingData(sc.U64(0), sc.U64(0)),
		},
		staleNotifier: staleNotifier,
	}

	return call
}

func (c callNoteStalled) DecodeArgs(buffer *bytes.Buffer) (primitives.Call, error) {
	delay, err := sc.DecodeU64(buffer)
	if err != nil {
		return c, err
	}

	bestFinalizedBlockNumber, err := sc.DecodeU64(buffer)
	if err != nil {
		return c, err
	}

	c.Arguments = sc.NewVaryingData(delay, bestFinalizedBlockNumber)
	return c, nil
}

func (c callNoteStalled) Encode(buffer *bytes.Buffer) error {
	return c.Callable.Encode(buffer)
}

func (c callNoteStalled) Bytes() []byte {
	return c.Callable.Bytes()
}

func (c callNoteStalled) ModuleIndex() sc.U8 {
	return c.Callable.ModuleIndex()
}

func (c callNoteStalled) FunctionIndex() sc.U8 {
	return c.Callable.FunctionIndex()
}

func (c callNoteStalled) Args() sc.VaryingData {
	return c.Callable.Args()
}

func (c callNoteStalled) BaseWeight() primitives.Weight {
	return callNoteStalledWeight(primitives.RuntimeDbWeight{})
}

func (_ callNoteStalled) WeighData(baseWeight primitives.Weight) primitives.Weight {
	return primitives.WeightFromParts(baseWeight.RefTime, 0)
}

func (_ callNoteStalled) ClassifyDispatch(baseWeight primitives.Weight) primitives.DispatchClass {
	return primitives.NewDispatchClassNormal()
}

func (_ callNoteStalled) PaysFee(baseWeight primitives.Weight) primitives.Pays {
	return primitives.PaysYes
}

func (c callNoteStalled) Dispatch(origin primitives.RuntimeOrigin, args sc.VaryingData) (primitives.PostDispatchInfo, error) {
	err := system.EnsureRoot(origin)
	if err != nil {
		return primitives.PostDispatchInfo{}, err
	}

	delay := args[0].(sc.U64)
	bestFinalizedBlockNumber := args[1].(sc.U64)

	c.staleNotifier.onStalled(delay, bestFinalizedBlockNumber)

	return primitives.PostDispatchInfo{}, nil
}

func (_ callNoteStalled) Docs() string {
	return "Note that the current authority set of the GRANDPA finality gadget has stalled."
}
