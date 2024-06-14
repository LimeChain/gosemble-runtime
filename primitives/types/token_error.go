package types

import (
	"bytes"

	sc "github.com/LimeChain/goscale"
)

const (
	TokenErrorFundsUnavailable sc.U8 = iota
	TokenErrorOnlyProvider
	TokenErrorBelowMinimum
	TokenErrorCannotCreate
	TokenErrorUnknownAsset
	TokenErrorFrozen
	TokenErrorUnsupported
	TokenErrorCannotCreateHold
	TokenErrorNotExpendable
	TokenErrorBlocked
)

type TokenError sc.VaryingData

func NewTokenErrorFundsUnavailable() TokenError {
	return TokenError(sc.NewVaryingData(TokenErrorFundsUnavailable))
}

func NewTokenErrorOnlyProvider() TokenError {
	return TokenError(sc.NewVaryingData(TokenErrorOnlyProvider))
}

func NewTokenErrorBelowMinimum() TokenError {
	return TokenError(sc.NewVaryingData(TokenErrorBelowMinimum))
}

func NewTokenErrorCannotCreate() TokenError {
	return TokenError(sc.NewVaryingData(TokenErrorCannotCreate))
}

func NewTokenErrorUnknownAsset() TokenError {
	return TokenError(sc.NewVaryingData(TokenErrorUnknownAsset))
}

func NewTokenErrorFrozen() TokenError {
	return TokenError(sc.NewVaryingData(TokenErrorFrozen))
}

func NewTokenErrorUnsupported() TokenError {
	return TokenError(sc.NewVaryingData(TokenErrorUnsupported))
}

func NewTokenErrorCannotCreateHold() TokenError {
	return TokenError(sc.NewVaryingData(TokenErrorCannotCreateHold))
}

func NewTokenErrorNotExpendable() TokenError {
	return TokenError(sc.NewVaryingData(TokenErrorNotExpendable))
}

func NewTokenErrorBlocked() TokenError {
	return TokenError(sc.NewVaryingData(TokenErrorBlocked))
}

func (err TokenError) Encode(buffer *bytes.Buffer) error {
	return err[0].Encode(buffer)
}

func (err TokenError) Error() string {
	if len(err) == 0 {
		return newTypeError("TokenError").Error()
	}

	switch err[0] {
	case TokenErrorFundsUnavailable:
		return "Funds are unavailable"
	case TokenErrorOnlyProvider:
		return "Account that must exist would die"
	case TokenErrorBelowMinimum:
		return "Account cannot exist with the funds that would be given"
	case TokenErrorCannotCreate:
		return "Account cannot be created"
	case TokenErrorUnknownAsset:
		return "The asset in question is unknown"
	case TokenErrorFrozen:
		return "Funds exist but are frozen"
	case TokenErrorUnsupported:
		return "Operation is not supported by the asset"
	case TokenErrorCannotCreateHold:
		return "Account cannot be created for recording amount on hold"
	case TokenErrorNotExpendable:
		return "Account that is desired to remain would die"
	case TokenErrorBlocked:
		return "Account cannot receive the assets"
	default:
		return newTypeError("TokenError").Error()
	}
}

func DecodeTokenError(buffer *bytes.Buffer) (TokenError, error) {
	b, err := sc.DecodeU8(buffer)
	if err != nil {
		return TokenError{}, newTypeError("TokenError")
	}

	switch b {
	case TokenErrorFundsUnavailable:
		return NewTokenErrorFundsUnavailable(), nil
	case TokenErrorOnlyProvider:
		return NewTokenErrorOnlyProvider(), nil
	case TokenErrorBelowMinimum:
		return NewTokenErrorBelowMinimum(), nil
	case TokenErrorCannotCreate:
		return NewTokenErrorCannotCreate(), nil
	case TokenErrorUnknownAsset:
		return NewTokenErrorUnknownAsset(), nil
	case TokenErrorFrozen:
		return NewTokenErrorFrozen(), nil
	case TokenErrorUnsupported:
		return NewTokenErrorUnsupported(), nil
	case TokenErrorCannotCreateHold:
		return NewTokenErrorCannotCreateHold(), nil
	case TokenErrorNotExpendable:
		return NewTokenErrorNotExpendable(), nil
	case TokenErrorBlocked:
		return NewTokenErrorBlocked(), nil
	default:
		return TokenError{}, newTypeError("TokenError")
	}
}

func (err TokenError) Bytes() []byte {
	return sc.EncodedBytes(err)
}
