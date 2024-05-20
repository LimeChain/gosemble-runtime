package grandpa

import (
	"bytes"
	"errors"

	sc "github.com/LimeChain/goscale"
	grandpafinality "github.com/LimeChain/gosemble/primitives/grandpafinality"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

// Proof of voter misbehavior on a given set id. Misbehavior/equivocation in
// GRANDPA happens when a voter votes on the same round (either at prevote or
// precommit stage) for different blocks. Proving is achieved by collecting the
// signed messages of conflicting votes.
type EquivocationProof struct {
	SetId        sc.U64
	Equivocation Equivocation
}

func (e EquivocationProof) Encode(buffer *bytes.Buffer) error {
	return sc.EncodeEach(buffer,
		e.SetId,
		e.Equivocation,
	)
}

func (e EquivocationProof) Bytes() []byte {
	return sc.EncodedBytes(e)
}

func DecodeEquivocationProof(buffer *bytes.Buffer) (EquivocationProof, error) {
	setId, err := sc.DecodeU64(buffer)
	if err != nil {
		return EquivocationProof{}, err
	}

	equivocation, err := DecodeEquivocation(buffer)
	if err != nil {
		return EquivocationProof{}, err
	}

	return EquivocationProof{
		SetId:        setId,
		Equivocation: equivocation,
	}, nil
}

// Returns the round number at which the equivocation occurred.
func (e EquivocationProof) Round() (sc.U64, error) {
	switch e.Equivocation.VaryingData[0] {
	case Prevote:
		equivocation := e.Equivocation.VaryingData[1].(grandpafinality.Equivocation)
		return equivocation.RoundNumber, nil
	case Precommit:
		equivocation := e.Equivocation.VaryingData[1].(grandpafinality.Equivocation)
		return equivocation.RoundNumber, nil
	default:
		return 0, errors.New("invalid 'Equivocation' type")
	}
}

// Returns the authority id of the equivocator.
func (e EquivocationProof) Offender() (primitives.AccountId, error) {
	switch e.Equivocation.VaryingData[0] {
	case Prevote:
		equivocation := e.Equivocation.VaryingData[1].(grandpafinality.Equivocation)
		return primitives.AccountId(equivocation.Identity), nil
	case Precommit:
		equivocation := e.Equivocation.VaryingData[1].(grandpafinality.Equivocation)
		return primitives.AccountId(equivocation.Identity), nil
	default:
		return primitives.AccountId{}, errors.New("invalid 'Equivocation' type")
	}
}

func CheckEquivocationProof(report EquivocationProof) bool {
	// TODO:
	return true

	// NOTE: the bare `Prevote` and `Precommit` types don't share any trait,
	// this is implemented as a macro to avoid duplication.

	// macro_rules! check {
	// 	( $equivocation:expr, $message:expr ) => {
	// 		// if both votes have the same target the equivocation is invalid.
	// 		if $equivocation.first.0.target_hash == $equivocation.second.0.target_hash &&
	// 			$equivocation.first.0.target_number == $equivocation.second.0.target_number
	// 		{
	// 			return false
	// 		}

	// 		// check signatures on both votes are valid
	// 		let valid_first = check_message_signature(
	// 			&$message($equivocation.first.0),
	// 			&$equivocation.identity,
	// 			&$equivocation.first.1,
	// 			$equivocation.round_number,
	// 			report.set_id,
	// 		);

	// 		let valid_second = check_message_signature(
	// 			&$message($equivocation.second.0),
	// 			&$equivocation.identity,
	// 			&$equivocation.second.1,
	// 			$equivocation.round_number,
	// 			report.set_id,
	// 		);

	// 		return valid_first && valid_second
	// 	};
	// }

	// match report.equivocation {
	// 	Equivocation::Prevote(equivocation) => {
	// 		check!(equivocation, grandpa::Message::Prevote);
	// 	},
	// 	Equivocation::Precommit(equivocation) => {
	// 		check!(equivocation, grandpa::Message::Precommit);
	// 	},
	// }
}
