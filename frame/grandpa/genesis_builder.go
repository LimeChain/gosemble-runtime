package grandpa

import (
	"encoding/json"
	"errors"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/primitives/types"
	"github.com/vedhavyas/go-subkey"
)

var (
	errInvalidAddrValue   = errors.New("Grandpa: invalid address in genesis config json")
	errInvalidWeightValue = errors.New("Grandpa: invalid weight in genesis config json")
)

type GenesisConfig struct {
	Authorities sc.Sequence[types.Authority]
}
type genesisConfigJsonStruct struct {
	GrandpaGenesisConfig struct {
		Authorities [][2]interface{} `json:"authorities"`
	} `json:"grandpa"`
}

func (gc *GenesisConfig) UnmarshalJSON(data []byte) error {
	gcJson := genesisConfigJsonStruct{}

	if err := json.Unmarshal(data, &gcJson); err != nil {
		return err
	}

	addrExists := map[string]bool{}
	for _, a := range gcJson.GrandpaGenesisConfig.Authorities {
		addrString, ok := a[0].(string)
		if !ok {
			return errInvalidAddrValue
		}

		if addrExists[addrString] {
			continue
		}

		_, publicKey, err := subkey.SS58Decode(addrString)
		if err != nil {
			return err
		}

		who, err := types.NewAccountId(sc.BytesToSequenceU8(publicKey)...)
		if err != nil {
			return err
		}

		weightFloat, ok := a[1].(float64)
		if !ok {
			return errInvalidWeightValue
		}

		weight := sc.U64(uint64(weightFloat))

		gc.Authorities = append(gc.Authorities, types.Authority{Id: who, Weight: weight})
		addrExists[addrString] = true
	}

	return nil
}

func (m module) CreateDefaultConfig() ([]byte, error) {
	gc := &genesisConfigJsonStruct{}
	gc.GrandpaGenesisConfig.Authorities = [][2]interface{}{}

	return json.Marshal(gc)
}

func (m module) BuildConfig(config []byte) error {
	gc := GenesisConfig{}
	if err := json.Unmarshal(config, &gc); err != nil {
		return err
	}

	m.storage.CurrentSetId.Put(0)
	return m.initialize(gc.Authorities)
}
