package system

import (
	"bytes"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/primitives/io"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

// Make some on-chain remark and emit event.
// Can be executed by any origin.
type callRemarkWithEvent struct {
	primitives.Callable
	dbWeight       primitives.RuntimeDbWeight
	eventDepositor primitives.EventDepositor
	ioHashing      io.Hashing
}

func newCallRemarkWithEvent(
	moduleId sc.U8,
	functionId sc.U8,
	dbWeight primitives.RuntimeDbWeight,
	ioHashing io.Hashing,
	eventDepositor primitives.EventDepositor,
) primitives.Call {
	call := callRemarkWithEvent{
		Callable: primitives.Callable{
			ModuleId:   moduleId,
			FunctionId: functionId,
			Arguments:  sc.NewVaryingData(sc.Sequence[sc.U8]{}),
		},
		dbWeight:       dbWeight,
		eventDepositor: eventDepositor,
		ioHashing:      ioHashing,
	}

	return call
}

func (c callRemarkWithEvent) DecodeArgs(buffer *bytes.Buffer) (primitives.Call, error) {
	args, err := sc.DecodeSequence[sc.U8](buffer)
	if err != nil {
		return nil, err
	}
	c.Arguments = sc.NewVaryingData(args)
	return c, nil
}

func (c callRemarkWithEvent) Encode(buffer *bytes.Buffer) error {
	return c.Callable.Encode(buffer)
}

func (c callRemarkWithEvent) Bytes() []byte {
	return c.Callable.Bytes()
}

func (c callRemarkWithEvent) ModuleIndex() sc.U8 {
	return c.Callable.ModuleIndex()
}

func (c callRemarkWithEvent) FunctionIndex() sc.U8 {
	return c.Callable.FunctionIndex()
}

func (c callRemarkWithEvent) Args() sc.VaryingData {
	return c.Callable.Args()
}

func (c callRemarkWithEvent) BaseWeight() primitives.Weight {
	message := c.Arguments[0].(sc.Sequence[sc.U8])
	return callRemarkWithEventWeight(c.dbWeight, sc.U64(len(message)))
}

func (_ callRemarkWithEvent) WeighData(baseWeight primitives.Weight) primitives.Weight {
	return primitives.WeightFromParts(baseWeight.RefTime, 0)
}

func (_ callRemarkWithEvent) ClassifyDispatch(baseWeight primitives.Weight) primitives.DispatchClass {
	return primitives.NewDispatchClassNormal()
}

func (_ callRemarkWithEvent) PaysFee(baseWeight primitives.Weight) primitives.Pays {
	return primitives.PaysYes
}

func (c callRemarkWithEvent) Dispatch(origin primitives.RuntimeOrigin, args sc.VaryingData) (primitives.PostDispatchInfo, error) {
	who, err := EnsureSigned(origin)
	if err != nil {
		return primitives.PostDispatchInfo{}, err
	}

	message := args[0].(sc.Sequence[sc.U8])

	hash, err := primitives.NewH256(sc.BytesToFixedSequenceU8(c.ioHashing.Blake256(sc.SequenceU8ToBytes(message)))...)
	if err != nil {
		return primitives.PostDispatchInfo{}, err
	}

	event := newEventRemarked(c.Callable.ModuleId, who.Value, hash)
	c.eventDepositor.DepositEvent(event)

	return primitives.PostDispatchInfo{}, nil
}

func (_ callRemarkWithEvent) Docs() string {
	return "Make some on-chain remark and emit an event."
}

func (c callRemarkWithEvent) typedArgs(args sc.VaryingData) sc.Sequence[sc.U8] {
	message := args[0].(sc.Sequence[sc.U8])

	return message
}
