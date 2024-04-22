package babe

import (
	"encoding/json"

	sc "github.com/LimeChain/goscale"
	babetypes "github.com/LimeChain/gosemble/primitives/babe"
	"github.com/LimeChain/gosemble/primitives/types"
	"github.com/vedhavyas/go-subkey"
)

type GenesisConfig struct {
	Authorities sc.Sequence[babetypes.Authority]
	EpochConfig babetypes.EpochConfiguration
}

type genesisConfigJsonStruct struct {
	BabeGenesisConfig struct {
		Authorities []string `json:"authorities"`
		EpochConfig struct {
			C            [2]uint64 `json:"c"`
			AllowedSlots string    `json:"allowed_slots"`
		} `json:"epochConfig"`
	} `json:"babe"`
}

func (gc *GenesisConfig) UnmarshalJSON(data []byte) error {
	gcJson := genesisConfigJsonStruct{}

	if err := json.Unmarshal(data, &gcJson); err != nil {
		return err
	}

	addrExists := map[string]bool{}
	for _, addr := range gcJson.BabeGenesisConfig.Authorities {
		if addrExists[addr] {
			continue
		}

		_, pubKeyBytes, err := subkey.SS58Decode(addr)
		if err != nil {
			return err
		}

		pubKey, err := types.NewSr25519PublicKey(sc.BytesToSequenceU8(pubKeyBytes)...)
		if err != nil {
			return err
		}

		gc.Authorities = append(gc.Authorities, babetypes.Authority{Key: pubKey})
		addrExists[addr] = true
	}

	c := gcJson.BabeGenesisConfig.EpochConfig.C
	gc.EpochConfig.C = types.RationalValue{
		Numerator:   sc.U64(c[0]),
		Denominator: sc.U64(c[1]),
	}

	switch gcJson.BabeGenesisConfig.EpochConfig.AllowedSlots {
	case babetypes.NewPrimarySlots().String():
		gc.EpochConfig.AllowedSlots = babetypes.NewPrimarySlots()
	case babetypes.NewPrimaryAndSecondaryPlainSlots().String():
		gc.EpochConfig.AllowedSlots = babetypes.NewPrimaryAndSecondaryPlainSlots()
	case babetypes.NewPrimaryAndSecondaryVRFSlots().String():
		gc.EpochConfig.AllowedSlots = babetypes.NewPrimaryAndSecondaryVRFSlots()
	default:
		return babetypes.ErrInvalidAllowedSlots
	}

	return nil
}

func (m module) CreateDefaultConfig() ([]byte, error) {
	gcJson := genesisConfigJsonStruct{}

	gcJson.BabeGenesisConfig.Authorities = []string{}

	return json.Marshal(gcJson)
}

func (m module) BuildConfig(config []byte) error {
	gc := GenesisConfig{}
	if err := json.Unmarshal(config, &gc); err != nil {
		return err
	}

	m.StorageSegmentIndexSet(0)

	err := m.initializeGenesisAuthorities(gc.Authorities)
	if err != nil {
		return err
	}

	m.StorageEpochConfigSet(gc.EpochConfig)

	return nil
}
