package aura_ext

import (
	"errors"
	"fmt"
	"reflect"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/execution/types"
	"github.com/LimeChain/gosemble/frame/aura"
	"github.com/LimeChain/gosemble/frame/executive"
	"github.com/LimeChain/gosemble/primitives/io"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

type BlockExecutor struct {
	ioCrypto        io.Crypto
	ioHashing       io.Hashing
	module          Module
	executiveModule executive.Module
}

func NewBlockExecutor(module Module, executiveModule executive.Module) BlockExecutor {
	return BlockExecutor{
		ioCrypto:        io.NewCrypto(),
		ioHashing:       io.NewHashing(),
		module:          module,
		executiveModule: executiveModule,
	}
}

func (be BlockExecutor) ExecuteBlock(block primitives.Block) error {
	header := block.Header()

	authorities, err := be.module.storage.Authorities.Get()
	if err != nil {
		return err
	}

	var seal *primitives.DigestSeal
	digestItems := sc.Sequence[primitives.DigestItem]{}
	for _, digestItem := range header.Digest.Sequence {
		if !digestItem.IsSeal() {
			digestItems = append(digestItems, digestItem)
			continue
		}

		s, err := digestItem.AsSeal()
		if err != nil {
			return err
		}
		if reflect.DeepEqual(sc.FixedSequenceU8ToBytes(s.ConsensusEngineId), aura.EngineId[:]) {
			if seal != nil {
				return errors.New("found multiple AuRa seals digests")
			}
			seal = &s
			continue
		}

		digestItems = append(digestItems, digestItem)
	}

	if seal == nil {
		return errors.New("could not find an AuRa seal digest")
	}

	header.Digest = primitives.NewDigest(digestItems)

	preRuntimes, err := header.Digest.PreRuntimes()
	if err != nil {
		return err
	}

	authorIndex, err := be.module.auraModule.FindAuthor(preRuntimes)
	if err != nil {
		return err
	}
	if !authorIndex.HasValue {
		return errors.New("could not find AuRa author index")
	}

	preHash := be.ioHashing.Blake256(header.Bytes())

	// sanity check
	if int(authorIndex.Value) > len(authorities) {
		return fmt.Errorf("invalid AuRa author index [%d]", authorIndex.Value)
	}

	bytesAuthority := sc.FixedSequenceU8ToBytes(authorities[authorIndex.Value].FixedSequence)

	verified := be.ioCrypto.Sr25519Verify(sc.SequenceU8ToBytes(seal.Message), preHash, bytesAuthority)
	if !verified {
		return fmt.Errorf("invalid AuRa seal")
	}

	return be.executiveModule.ExecuteBlock(types.NewBlock(header, block.Extrinsics()))
}
