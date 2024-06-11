package types

import (
	"bytes"
	"errors"
	"reflect"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/primitives/io"
	"github.com/LimeChain/gosemble/primitives/log"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

const (
	// ExtrinsicFormatVersion is the current version of the [`UncheckedExtrinsic`] encoded format.
	//
	// This version needs to be bumped if the encoded representation changes.
	// It ensures that if the representation is changed and the format is not known,
	// the decoding fails.
	ExtrinsicFormatVersion = 4
	ExtrinsicBitSigned     = 0b1000_0000
	ExtrinsicUnmaskVersion = 0b0111_1111
)

var (
	errInvalidMultisigType = errors.New("invalid MultiSignature type in Verify")
)

type PayloadInitializer = func(call primitives.Call, extra primitives.SignedExtra) (
	primitives.SignedPayload, error,
)

type uncheckedExtrinsic struct {
	version sc.U8
	// The signature, address, number of extrinsics have come before from
	// the same signer and an era describing the longevity of this transaction,
	// if this is a signed extrinsic.
	signature         sc.Option[primitives.ExtrinsicSignature]
	function          primitives.Call
	extra             primitives.SignedExtra
	initializePayload PayloadInitializer
	crypto            io.Crypto
	hashing           io.Hashing
	logger            log.RuntimeLogger
}

// NewUncheckedExtrinsic returns a new instance of an unchecked extrinsic.
func NewUncheckedExtrinsic(version sc.U8, signature sc.Option[primitives.ExtrinsicSignature], function primitives.Call, extra primitives.SignedExtra, logger log.RuntimeLogger) primitives.UncheckedExtrinsic {
	return uncheckedExtrinsic{
		version:           version,
		signature:         signature,
		function:          function,
		extra:             extra,
		initializePayload: primitives.NewSignedPayload,
		crypto:            io.NewCrypto(),
		hashing:           io.NewHashing(),
		logger:            logger,
	}
}

// NewUnsignedUncheckedExtrinsic returns a new instance of an unsigned extrinsic.
func NewUnsignedUncheckedExtrinsic(function primitives.Call) primitives.UncheckedExtrinsic {
	return uncheckedExtrinsic{
		version:   ExtrinsicFormatVersion,
		signature: sc.NewOption[primitives.ExtrinsicSignature](nil),
		function:  function,
		crypto:    io.NewCrypto(),
	}
}

func (uxt uncheckedExtrinsic) Encode(buffer *bytes.Buffer) error {
	tempBuffer := &bytes.Buffer{}

	err := uxt.version.Encode(tempBuffer)
	if err != nil {
		return err
	}
	if uxt.signature.HasValue {
		err := uxt.signature.Value.Encode(tempBuffer)
		if err != nil {
			return err
		}
	}
	err = uxt.function.Encode(tempBuffer)
	if err != nil {
		return err
	}

	encodedLen := sc.ToCompact(uint64(tempBuffer.Len()))
	err = encodedLen.Encode(buffer)
	if err != nil {
		return err
	}
	_, err = buffer.Write(tempBuffer.Bytes())
	return err
}

func (uxt uncheckedExtrinsic) Bytes() []byte {
	return sc.EncodedBytes(uxt)
}

func (uxt uncheckedExtrinsic) Signature() sc.Option[primitives.ExtrinsicSignature] {
	return uxt.signature
}

func (uxt uncheckedExtrinsic) Function() primitives.Call {
	return uxt.function
}

func (uxt uncheckedExtrinsic) Extra() primitives.SignedExtra {
	return uxt.extra
}

func (uxt uncheckedExtrinsic) IsSigned() bool {
	return bool(uxt.signature.HasValue)
}

func (uxt uncheckedExtrinsic) Check() (primitives.CheckedExtrinsic, error) {
	if uxt.signature.HasValue {
		signer, signature, extra := uxt.signature.Value.Signer, uxt.signature.Value.Signature, uxt.signature.Value.Extra

		signerAddress, err := primitives.Lookup(signer)
		if err != nil {
			return nil, err
		}

		rawPayload, err := uxt.initializePayload(uxt.function, extra)
		if err != nil {
			return nil, err
		}

		verify, err := uxt.verify(signature, uxt.usingEncoded(rawPayload), signerAddress)
		if err != nil {
			return nil, err
		}

		if !verify {
			return nil, primitives.NewTransactionValidityError(primitives.NewInvalidTransactionBadProof())
		}

		return NewCheckedExtrinsic(sc.NewOption[primitives.AccountId](signerAddress), uxt.function, extra, uxt.logger), nil
	}

	return NewCheckedExtrinsic(sc.NewOption[primitives.AccountId](nil), uxt.function, uxt.extra, uxt.logger), nil
}

func (uxt uncheckedExtrinsic) verify(signature primitives.MultiSignature, msg sc.Sequence[sc.U8], signer primitives.AccountId) (bool, error) {
	msgBytes := sc.SequenceU8ToBytes(msg)
	signerBytes := signer.Bytes()

	if signature.IsEd25519() {
		sigEd25519, err := signature.AsEd25519()
		if err != nil {
			return false, err
		}
		sigBytes := sc.FixedSequenceU8ToBytes(sigEd25519.FixedSequence)
		return uxt.crypto.Ed25519Verify(sigBytes, msgBytes, signerBytes), nil
	} else if signature.IsSr25519() {
		sigSr25519, err := signature.AsSr25519()
		if err != nil {
			return false, err
		}
		sigBytes := sc.FixedSequenceU8ToBytes(sigSr25519.FixedSequence)
		return uxt.crypto.Sr25519Verify(sigBytes, msgBytes, signerBytes), nil
	} else if signature.IsEcdsa() {
		sigEcdsa, err := signature.AsEcdsa()
		if err != nil {
			return false, err
		}

		return uxt.verifyEcdsa(sigEcdsa, msgBytes, signerBytes)
	}

	return false, errInvalidMultisigType
}

func (uxt uncheckedExtrinsic) usingEncoded(sp primitives.SignedPayload) sc.Sequence[sc.U8] {
	enc := sp.Bytes()

	if len(enc) > 256 {
		hash := uxt.hashing.Blake256(enc)
		return sc.BytesToSequenceU8(hash)
	} else {
		return sc.BytesToSequenceU8(enc)
	}
}

func (uxt uncheckedExtrinsic) verifyEcdsa(signature primitives.SignatureEcdsa, msgBytes []byte, signer []byte) (bool, error) {
	sigBytes := sc.FixedSequenceU8ToBytes(signature.FixedSequence)
	msg := uxt.hashing.Blake256(msgBytes)

	// This returns either the 33-byte ECDSA Public Key or an error.
	recovered := uxt.crypto.EcdsaRecoverCompressed(sigBytes, msg)
	buffer := bytes.NewBuffer(recovered)

	result, err := sc.DecodeResult(buffer, primitives.DecodeEcdsaPublicKey, primitives.DecodeEcdsaVerifyError)
	if err != nil {
		return false, err
	}

	if result.HasError {
		uxt.logger.Debugf("Failed to verify signature. Error: [%s]", result.Value.(error).Error())
		return false, nil
	}

	// In order to match AccountId, ECDSA public keys are hashed to 32 bytes.
	hashPublicKey := uxt.hashing.Blake256(result.Value.Bytes())

	return reflect.DeepEqual(hashPublicKey, signer), nil
}
