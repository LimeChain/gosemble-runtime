package grandpa

import (
	"bytes"
	"errors"

	sc "github.com/LimeChain/goscale"
	grandpafinality "github.com/LimeChain/gosemble/primitives/grandpafinality"
)

// Wrapper object for GRANDPA equivocation proofs, useful for unifying prevote
// and precommit equivocations under a common type.
const (
	// Proof of equivocation at prevote stage.
	Prevote sc.U8 = iota
	// Proof of equivocation at precommit stage.
	Precommit
)

type Equivocation struct {
	sc.VaryingData
}

func NewEquivocationPrevote(prevoteEquivocation grandpafinality.Equivocation) Equivocation {
	return Equivocation{sc.NewVaryingData(Prevote, prevoteEquivocation)}
}

func NewEquivocationPrecommit(precommitEquivocation grandpafinality.Equivocation) Equivocation {
	return Equivocation{sc.NewVaryingData(Precommit, precommitEquivocation)}
}

func DecodeEquivocation(buffer *bytes.Buffer) (Equivocation, error) {
	index, err := sc.DecodeU8(buffer)
	if err != nil {
		return Equivocation{}, err
	}

	switch index {
	case Prevote:
		equivocation, err := grandpafinality.DecodeEquivocation(buffer)
		if err != nil {
			return Equivocation{}, err
		}
		return NewEquivocationPrevote(equivocation), nil
	case Precommit:
		equivocation, err := grandpafinality.DecodeEquivocation(buffer)
		if err != nil {
			return Equivocation{}, err
		}
		return NewEquivocationPrecommit(equivocation), nil
	default:
		return Equivocation{}, errors.New("invalid 'Equivocation' index")
	}
}
