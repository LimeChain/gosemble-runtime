package balances

import (
	"bytes"
	"testing"

	sc "github.com/LimeChain/goscale"
	balancestypes "github.com/LimeChain/gosemble/frame/balances/types"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_Call_ForceAdjustTotalIssuance_New(t *testing.T) {
	target := setupCallForceAdjustTotalIssuance()
	expected := callForceAdjustTotalIssuance{
		Callable: primitives.Callable{
			ModuleId:   target.module.Index,
			FunctionId: functionForceAdjustTotalIssuanceIndex,
			Arguments:  sc.NewVaryingData(balancestypes.AdjustmentDirection(0), sc.Compact{Number: sc.U128{}}),
		},
		module: target.module,
	}

	assert.Equal(t, expected, target)
}

func Test_Call_ForceAdjustTotalIssuance_DecodeArgs(t *testing.T) {
	target := setupCallForceAdjustTotalIssuance()
	buf := &bytes.Buffer{}

	adjustmentDirection := balancestypes.AdjustmentDirectionDecrease
	amount := sc.Compact{Number: sc.NewU128(1)}

	err := adjustmentDirection.Encode(buf)
	assert.NoError(t, err)

	err = amount.Encode(buf)
	assert.NoError(t, err)

	call, err := target.DecodeArgs(buf)
	assert.NoError(t, err)
	assert.Equal(t, sc.NewVaryingData(adjustmentDirection, amount), call.Args())
}

func Test_Call_ForceAdjustTotalIssuance_DecodeArgs_Err(t *testing.T) {
	target := setupCallForceAdjustTotalIssuance()
	buf := &bytes.Buffer{}

	_, err := target.DecodeArgs(buf)
	assert.Error(t, err)

	adjustmentDirection := balancestypes.AdjustmentDirectionDecrease
	err = adjustmentDirection.Encode(buf)
	assert.NoError(t, err)

	_, err = target.DecodeArgs(buf)
	assert.Error(t, err)
}

func Test_Call_ForceAdjustTotalIssuance_BaseWeight(t *testing.T) {
	target := setupCallForceAdjustTotalIssuance()
	assert.Equal(t, callForceAdjustTotalIssuanceWeight(target.module.dbWeight()), target.BaseWeight())
}

func Test_Call_ForceAdjustTotalIssuance_WeighData(t *testing.T) {
	target := setupCallForceAdjustTotalIssuance()
	assert.Equal(t, primitives.WeightFromParts(124, 0), target.WeighData(baseWeight))
}

func Test_Call_ForceAdjustTotalIssuance_ClassifyDispatch(t *testing.T) {
	target := setupCallForceAdjustTotalIssuance()

	assert.Equal(t, primitives.NewDispatchClassNormal(), target.ClassifyDispatch(baseWeight))
}

func Test_Call_ForceAdjustTotalIssuance_PaysFee(t *testing.T) {
	target := setupCallForceAdjustTotalIssuance()

	assert.Equal(t, primitives.PaysYes, target.PaysFee(baseWeight))
}

func Test_Call_ForceAdjustTotalIssuance_Success(t *testing.T) {
	ti := sc.NewU128(1)
	delta := sc.NewU128(1)

	target := setupCallForceAdjustTotalIssuance()

	mockTotalIssuance.On("Get").Return(ti, nil)
	mockTotalIssuance.On("Put", ti.Add(delta))
	mockInactiveIssuance.On("Get").Return(sc.NewU128(0), nil)
	mockStoredMap.On("DepositEvent", newEventTotalIssuanceForced(moduleId, ti, ti.Add(delta)))

	_, err := target.Dispatch(primitives.NewRawOriginRoot(), sc.NewVaryingData(balancestypes.AdjustmentDirectionIncrease, sc.Compact{Number: delta}))
	assert.NoError(t, err)

	mockTotalIssuance.AssertCalled(t, "Put", ti.Add(delta))
	mockStoredMap.AssertCalled(t, "DepositEvent", newEventTotalIssuanceForced(moduleId, ti, ti.Add(delta)))

}
func Test_Call_ForceAdjustTotalIssuance_Dispatch(t *testing.T) {
	tests := []struct {
		name        string
		origin      primitives.RuntimeOrigin
		args        sc.VaryingData
		expectedErr error
	}{
		{
			name:        "error bad origin",
			origin:      primitives.NewRawOriginNone(),
			expectedErr: primitives.NewDispatchErrorBadOrigin(),
		},
		{
			name:        "error invalid arg adjustment direction",
			origin:      primitives.NewRawOriginRoot(),
			args:        sc.NewVaryingData(sc.Empty{}),
			expectedErr: errInvalidArgAdjustmentDirection,
		},
		{
			name:        "error invalid arg delta compact",
			origin:      primitives.NewRawOriginRoot(),
			args:        sc.NewVaryingData(balancestypes.AdjustmentDirectionDecrease, sc.Empty{}),
			expectedErr: errInvalidArgDeltaCompact,
		},
		{
			name:        "error invalid arg delta",
			origin:      primitives.NewRawOriginRoot(),
			args:        sc.NewVaryingData(balancestypes.AdjustmentDirectionDecrease, sc.Compact{Number: sc.U8(1)}),
			expectedErr: errInvalidArgDelta,
		},
		{
			name:   "error forceAdjustTotalIssuance",
			origin: primitives.NewRawOriginRoot(),
			args:   sc.NewVaryingData(balancestypes.AdjustmentDirectionDecrease, sc.Compact{Number: sc.NewU128(0)}),
			expectedErr: primitives.NewDispatchErrorModule(primitives.CustomModuleError{
				Index:   moduleId,
				Err:     sc.U32(ErrorDeltaZero),
				Message: sc.NewOption[sc.Str](nil),
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target := setupCallForceAdjustTotalIssuance()

			_, err := target.Dispatch(tt.origin, tt.args)
			assert.Equal(t, tt.expectedErr, err)
		})
	}
}

func Test_Call_ForceAdjustTotalIssuance_forceAdjustTotalIssuance(t *testing.T) {
	tests := []struct {
		name                           string
		delta                          sc.U128
		direction                      balancestypes.AdjustmentDirection
		inactiveIssuance               sc.U128
		expectedGetTotalIssuanceErr    error
		expectedGetInactiveIssuanceErr error
		expectedErr                    error
	}{
		{
			name: "error delta zero",
			expectedErr: primitives.NewDispatchErrorModule(primitives.CustomModuleError{
				Index:   moduleId,
				Err:     sc.U32(ErrorDeltaZero),
				Message: sc.NewOption[sc.Str](nil),
			}),
		},
		{
			name:                        "error getTotalIssuance",
			delta:                       sc.NewU128(1),
			expectedGetTotalIssuanceErr: expectedErr,
			expectedErr:                 expectedErr,
		},
		{
			name:                           "error getInactiveIssuance",
			delta:                          sc.NewU128(1),
			expectedGetInactiveIssuanceErr: expectedErr,
			expectedErr:                    expectedErr,
		},
		{
			name:             "error inactiveIssuance.Gt(totalIssuance)",
			inactiveIssuance: sc.NewU128(3),
			delta:            sc.NewU128(1),
			expectedErr: primitives.NewDispatchErrorModule(primitives.CustomModuleError{
				Index:   moduleId,
				Err:     sc.U32(IssuanceDeactivated),
				Message: sc.NewOption[sc.Str](nil),
			}),
		},
		{
			name:             "direction decrease",
			inactiveIssuance: sc.NewU128(2),
			delta:            sc.NewU128(1),
			direction:        balancestypes.AdjustmentDirectionDecrease,
			expectedErr: primitives.NewDispatchErrorModule(primitives.CustomModuleError{
				Index:   moduleId,
				Err:     sc.U32(IssuanceDeactivated),
				Message: sc.NewOption[sc.Str](nil),
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target := setupCallForceAdjustTotalIssuance()

			mockTotalIssuance.On("Get").Return(sc.NewU128(1), tt.expectedGetTotalIssuanceErr)
			mockTotalIssuance.On("Put", mock.Anything)
			mockInactiveIssuance.On("Get").Return(tt.inactiveIssuance, tt.expectedGetInactiveIssuanceErr)
			mockStoredMap.On("DepositEvent", mock.Anything)

			err := target.forceAdjustTotalIssuance(tt.delta, tt.direction)
			assert.Equal(t, tt.expectedErr, err)

			mockTotalIssuance.AssertNotCalled(t, "Put", mock.Anything)
		})
	}
}

func setupCallForceAdjustTotalIssuance() callForceAdjustTotalIssuance {
	return newCallForceAdjustTotalIssuance(functionForceAdjustTotalIssuanceIndex, setupModule())
}
