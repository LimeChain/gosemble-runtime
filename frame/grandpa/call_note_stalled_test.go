package grandpa

import (
	"bytes"
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/mocks"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
)

var (
	delay                    = sc.U64(1000)
	bestFinalizedBlockNumber = sc.U64(1001)
	baseWeight               = primitives.WeightFromParts(123, 456)
)

var (
	staleNotifier *mocks.StaleNotifier
)

func setupCallNoteStalled() primitives.Call {
	staleNotifier = new(mocks.StaleNotifier)

	return newCallNoteStalled(moduleId, functionNoteStalledIndex, dbWeight, staleNotifier)
}

func Test_Call_NoteStalled_New(t *testing.T) {
	target := setupCallNoteStalled()

	expect := callNoteStalled{
		Callable: primitives.Callable{
			ModuleId:   moduleId,
			FunctionId: functionNoteStalledIndex,
			Arguments:  sc.NewVaryingData(),
		},
		dbWeight:      dbWeight,
		staleNotifier: staleNotifier,
	}

	assert.Equal(t, expect, target)
}

func Test_Call_NoteStalled_DecodeArgs(t *testing.T) {
	buffer := &bytes.Buffer{}
	buffer.Write(delay.Bytes())
	buffer.Write(bestFinalizedBlockNumber.Bytes())

	target := setupCallNoteStalled()

	call, err := target.DecodeArgs(buffer)
	assert.Nil(t, err)

	assert.Equal(t, 0, buffer.Len())
	assert.Equal(t, sc.NewVaryingData(delay, bestFinalizedBlockNumber), call.Args())
}

func Test_Call_NoteStalled_Encode(t *testing.T) {
	target := setupCallNoteStalled()

	buf := &bytes.Buffer{}

	err := target.Encode(buf)

	assert.NoError(t, err)
	assert.Equal(t, bytes.NewBuffer(append(moduleId.Bytes(), byte(functionNoteStalledIndex))), buf)
}

func Test_Call_NoteStalled_Bytes(t *testing.T) {
	target := setupCallNoteStalled()

	assert.Equal(t, append(moduleId.Bytes(), byte(functionNoteStalledIndex)), target.Bytes())
}

func Test_Call_NoteStalled_ModuleIndex(t *testing.T) {
	target := setupCallNoteStalled()

	assert.Equal(t, sc.U8(moduleId), target.ModuleIndex())
}

func Test_Call_NoteStalled_FunctionIndex(t *testing.T) {
	target := setupCallNoteStalled()

	assert.Equal(t, sc.U8(functionNoteStalledIndex), target.FunctionIndex())
}

func Test_Call_NoteStalled_BaseWeight(t *testing.T) {
	target := setupCallNoteStalled()

	assert.Equal(t, callNoteStalledWeight(dbWeight), target.BaseWeight())
}

func Test_Call_NoteStalled_WeighData(t *testing.T) {
	target := setupCallNoteStalled()

	assert.Equal(t, primitives.WeightFromParts(123, 0), target.WeighData(baseWeight))
}

func Test_Call_NoteStalled_ClassifyDispatch(t *testing.T) {
	target := setupCallNoteStalled()

	assert.Equal(t, primitives.NewDispatchClassNormal(), target.ClassifyDispatch(baseWeight))
}

func Test_Call_NoteStalled_PaysFee(t *testing.T) {
	target := setupCallNoteStalled()

	assert.Equal(t, primitives.PaysYes, target.PaysFee(baseWeight))
}

func Test_Call_NoteStalled_Dispatch_Success(t *testing.T) {
	target := setupCallNoteStalled()

	staleNotifier.On("OnStalled", delay, bestFinalizedBlockNumber).Return(nil)

	res, err := target.Dispatch(primitives.NewRawOriginRoot(), sc.NewVaryingData(delay, bestFinalizedBlockNumber))

	assert.Nil(t, err)
	assert.Equal(t, primitives.PostDispatchInfo{}, res)

	staleNotifier.AssertCalled(t, "OnStalled", delay, bestFinalizedBlockNumber)
}

func Test_Call_NoteStalled_Dispatch_InvalidOrigin(t *testing.T) {
	target := setupCallNoteStalled()

	staleNotifier.On("OnStalled", delay, bestFinalizedBlockNumber).Return(nil)

	res, err := target.Dispatch(primitives.NewRawOriginNone(), sc.NewVaryingData(delay, bestFinalizedBlockNumber))

	assert.Equal(t, primitives.NewDispatchErrorBadOrigin(), err)
	assert.Equal(t, primitives.PostDispatchInfo{}, res)

	staleNotifier.AssertNotCalled(t, "OnStalled", delay, bestFinalizedBlockNumber)
}

func Test_Call_NoteStalled_Docs(t *testing.T) {
	target := setupCallNoteStalled()

	assert.Equal(t, "Note that the current authority set of the GRANDPA finality gadget has stalled.", target.Docs())
}
