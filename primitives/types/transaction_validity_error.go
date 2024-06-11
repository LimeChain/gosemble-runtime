package types

import (
	"bytes"

	sc "github.com/LimeChain/goscale"
)

var (
	errInvalidTransactionValidityErrorType = newTypeError("TransactionValidityError")
)

const (
	TransactionValidityErrorInvalidTransaction sc.U8 = iota
	TransactionValidityErrorUnknownTransaction
)

// TransactionValidityError Errors that can occur while checking the validity of a transaction.
type TransactionValidityError struct {
	sc.VaryingData
}

func NewTransactionValidityError(value sc.Encodable) error {
	// InvalidTransaction = 0 - Transaction is invalid.
	// UnknownTransaction = 1 - Transaction validity can't be determined.
	switch value.(type) {
	case InvalidTransaction:
		return TransactionValidityError{sc.NewVaryingData(TransactionValidityErrorInvalidTransaction, value)}
	case UnknownTransaction:
		return TransactionValidityError{sc.NewVaryingData(TransactionValidityErrorUnknownTransaction, value)}
	default:
		return errInvalidTransactionValidityErrorType
	}
}

func (e TransactionValidityError) Encode(buffer *bytes.Buffer) error {
	if len(e.VaryingData) != 2 {
		return errInvalidTransactionValidityErrorType
	}
	switch e.VaryingData[0] {
	case TransactionValidityErrorUnknownTransaction, TransactionValidityErrorInvalidTransaction:
		return e.VaryingData.Encode(buffer)
	default:
		return errInvalidTransactionValidityErrorType
	}
}

func (err TransactionValidityError) Error() string {
	if len(err.VaryingData) != 2 {
		return errInvalidTransactionValidityErrorType.Error()
	}

	switch err.VaryingData[0] {
	case TransactionValidityErrorUnknownTransaction:
		return err.VaryingData[1].(UnknownTransaction).Error()
	case TransactionValidityErrorInvalidTransaction:
		return err.VaryingData[1].(InvalidTransaction).Error()
	default:
		return errInvalidTransactionValidityErrorType.Error()
	}
}

func DecodeTransactionValidityError(buffer *bytes.Buffer) (TransactionValidityError, error) {
	b, err := sc.DecodeU8(buffer)
	if err != nil {
		return TransactionValidityError{}, err
	}

	switch b {
	case TransactionValidityErrorInvalidTransaction:
		value, err := DecodeInvalidTransaction(buffer)
		if err != nil {
			return TransactionValidityError{}, err
		}
		err = NewTransactionValidityError(value)
		if txErr, ok := err.(TransactionValidityError); ok {
			return txErr, nil
		}
		return TransactionValidityError{}, err
	case TransactionValidityErrorUnknownTransaction:
		value, err := DecodeUnknownTransaction(buffer)
		if err != nil {
			return TransactionValidityError{}, err
		}
		err = NewTransactionValidityError(value)
		if txErr, ok := err.(TransactionValidityError); ok {
			return txErr, nil
		}
		return TransactionValidityError{}, err
	default:
		return TransactionValidityError{}, errInvalidTransactionValidityErrorType
	}
}

func (e TransactionValidityError) MetadataDefinition(typesInvalidTxId int, typesUnknownTxId int) *MetadataTypeDefinition {
	def := NewMetadataTypeDefinitionVariant(
		sc.Sequence[MetadataDefinitionVariant]{
			NewMetadataDefinitionVariant(
				"Invalid",
				sc.Sequence[MetadataTypeDefinitionField]{
					NewMetadataTypeDefinitionField(typesInvalidTxId),
				},
				TransactionValidityErrorInvalidTransaction,
				""),
			NewMetadataDefinitionVariant(
				"Unknown",
				sc.Sequence[MetadataTypeDefinitionField]{
					NewMetadataTypeDefinitionField(typesUnknownTxId),
				},
				TransactionValidityErrorUnknownTransaction,
				""),
		},
	)
	return &def
}
