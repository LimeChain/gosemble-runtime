package babe

import (
	"encoding/json"
	"errors"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/primitives/types"
	"github.com/vedhavyas/go-subkey"
)

type GenesisConfig struct {
	Authorities sc.Sequence[Authority]
	EpochConfig BabeEpochConfiguration
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

	// authorities
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

		gc.Authorities = append(gc.Authorities, Authority{Key: pubKey})
		addrExists[addr] = true
	}

	// c
	c := gcJson.BabeGenesisConfig.EpochConfig.C
	gc.EpochConfig.C = types.RationalValue{Numerator: sc.U64(c[0]), Denominator: sc.U64(c[1])}

	// allowed slots
	switch gcJson.BabeGenesisConfig.EpochConfig.AllowedSlots {
	case NewPrimarySlots().String():
		gc.EpochConfig.AllowedSlots = NewPrimarySlots()
	case NewPrimaryAndSecondaryPlainSlots().String():
		gc.EpochConfig.AllowedSlots = NewPrimaryAndSecondaryPlainSlots()
	case NewPrimaryAndSecondaryVRFSlots().String():
		gc.EpochConfig.AllowedSlots = NewPrimaryAndSecondaryVRFSlots()
	default:
		return errors.New("invalid 'AllowedSlots' type")
	}

	return nil
}

func (m module) CreateDefaultConfig() ([]byte, error) {
	gcJson := genesisConfigJsonStruct{}
	// TODO: check if this is the correct configuration
	gcJson.BabeGenesisConfig.Authorities = []string{"5GrwvaEF5zXb26Fz9rcQpDWS57CtERHpNehXCPcNoHGKutQY"}
	gcJson.BabeGenesisConfig.EpochConfig.C = [2]uint64{1, 4}
	gcJson.BabeGenesisConfig.EpochConfig.AllowedSlots = "PrimaryAndSecondaryVRFSlots"
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
