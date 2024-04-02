package types

import (
	"bytes"
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/stretchr/testify/assert"
)

func Test_TokenError(t *testing.T) {
	for _, tt := range []struct {
		name           string
		newErr         TokenError
		expectedErr    error
		expectedErrMsg string
	}{
		{
			name:           "TokenErrorNoFunds",
			newErr:         NewTokenErrorNoFunds(),
			expectedErr:    TokenError(sc.NewVaryingData(TokenErrorNoFunds)),
			expectedErrMsg: "Funds are unavailable.",
		},
		{
			name:           "TokenErrorWouldDie",
			newErr:         NewTokenErrorWouldDie(),
			expectedErr:    TokenError(sc.NewVaryingData(TokenErrorWouldDie)),
			expectedErrMsg: "Account that must exist would die.",
		},
		{
			name:           "TokenErrorBelowMinimum",
			newErr:         NewTokenErrorBelowMinimum(),
			expectedErr:    TokenError(sc.NewVaryingData(TokenErrorBelowMinimum)),
			expectedErrMsg: "Account cannot exist with the funds that would be given.",
		},
		{
			name:           "TokenErrorCannotCreate",
			newErr:         NewTokenErrorCannotCreate(),
			expectedErr:    TokenError(sc.NewVaryingData(TokenErrorCannotCreate)),
			expectedErrMsg: "Account cannot be created.",
		},
		{
			name:           "TokenErrorUnknownAsset",
			newErr:         NewTokenErrorUnknownAsset(),
			expectedErr:    TokenError(sc.NewVaryingData(TokenErrorUnknownAsset)),
			expectedErrMsg: "The asset in question is unknown.",
		},
		{
			name:           "TokenErrorFrozen",
			newErr:         NewTokenErrorFrozen(),
			expectedErr:    TokenError(sc.NewVaryingData(TokenErrorFrozen)),
			expectedErrMsg: "Funds exist but are frozen.",
		},
		{
			name:           "TokenErrorUnsupported",
			newErr:         NewTokenErrorUnsupported(),
			expectedErr:    TokenError(sc.NewVaryingData(TokenErrorUnsupported)),
			expectedErrMsg: "Operation is not supported by the asset.",
		},
		{
			name:           "TokenErrorCannotCreateHold",
			newErr:         NewTokenErrorCannotCreateHold(),
			expectedErr:    TokenError(sc.NewVaryingData(TokenErrorCannotCreateHold)),
			expectedErrMsg: "Account cannot be created for a held balance.",
		},
		{
			name:           "TokenErrorNotExpendable",
			newErr:         NewTokenErrorNotExpendable(),
			expectedErr:    TokenError(sc.NewVaryingData(TokenErrorNotExpendable)),
			expectedErrMsg: "Withdrawal would cause unwanted loss of account.",
		},
		{
			name:           "TokenErrorBlocked",
			newErr:         NewTokenErrorBlocked(),
			expectedErr:    TokenError(sc.NewVaryingData(TokenErrorBlocked)),
			expectedErrMsg: "Account cannot receive the assets.",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			buffer := &bytes.Buffer{}
			err := tt.newErr.Encode(buffer)
			assert.NoError(t, err)

			haveErr, err := DecodeTokenError(buffer)
			assert.NoError(t, err)

			assert.Equal(t, tt.expectedErr, haveErr)
			assert.Equal(t, tt.expectedErrMsg, haveErr.Error())
		})
	}
}

func Test_DecodeTokenError_TypeError(t *testing.T) {
	for _, tt := range []struct {
		name    string
		errType sc.Encodable
	}{
		{
			name:    "invalid type",
			errType: sc.U8(10),
		},
		{
			name:    "nil",
			errType: sc.Empty{},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			buffer := &bytes.Buffer{}
			err := tt.errType.Encode(buffer)
			assert.NoError(t, err)

			_, err = DecodeTokenError(buffer)
			assert.Error(t, err)
			assert.Equal(t, "not a valid 'TokenError' type", err.Error())
		})
	}
}
