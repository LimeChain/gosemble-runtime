package aura

import (
	"encoding/json"
	"errors"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/primitives/types"
	"github.com/vedhavyas/go-subkey"
)

var (
	errAuthoritiesAlreadyInitialized   = errors.New("Authorities are already initialized!")
	errAuthoritiesExceedMaxAuthorities = errors.New("Initial authority set must be less than MaxAuthorities")
)

type GenesisConfig struct {
	Authorities sc.Sequence[types.Sr25519PublicKey]
}

type genesisConfigJsonStruct struct {
	AuraGenesisConfig struct {
		Authorities []string `json:"authorities"`
	} `json:"aura"`
}

func (gc *GenesisConfig) UnmarshalJSON(data []byte) error {
	gcJson := genesisConfigJsonStruct{}

	if err := json.Unmarshal(data, &gcJson); err != nil {
		return err
	}

	addrExists := map[string]bool{}
	for _, a := range gcJson.AuraGenesisConfig.Authorities {
		if addrExists[a] {
			continue
		}

		_, pubKeyBytes, err := subkey.SS58Decode(a)
		if err != nil {
			return err
		}

		pubKey, err := types.NewSr25519PublicKey(sc.BytesToSequenceU8(pubKeyBytes)...)
		if err != nil {
			return err
		}

		gc.Authorities = append(gc.Authorities, pubKey)
		addrExists[a] = true
	}

	return nil
}

func (m Module) CreateDefaultConfig() ([]byte, error) {
	gc := genesisConfigJsonStruct{}
	gc.AuraGenesisConfig.Authorities = []string{}

	return json.Marshal(gc)
}

func (m Module) BuildConfig(config []byte) error {
	gc := GenesisConfig{}
	if err := json.Unmarshal(config, &gc); err != nil {
		return err
	}

	if len(gc.Authorities) == 0 {
		return nil
	}

	return m.initializeAuthorities(gc.Authorities)
}
