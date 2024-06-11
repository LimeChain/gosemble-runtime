package sudo

import (
	"encoding/json"
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/primitives/types"
	"github.com/vedhavyas/go-subkey"
)

type GenesisConfig struct {
	Key types.AccountId
}

type genesisConfigJsonStruct struct {
	SudoGenesisConfig struct {
		Key string `json:"key"`
	} `json:"sudo"`
}

func (gc *GenesisConfig) UnmarshalJSON(data []byte) error {
	gcJson := genesisConfigJsonStruct{}

	if err := json.Unmarshal(data, &gcJson); err != nil {
		return err
	}

	_, acc, err := subkey.SS58Decode(gcJson.SudoGenesisConfig.Key)
	if err != nil {
		return err
	}

	key, err := types.NewAccountId(sc.BytesToSequenceU8(acc)...)
	if err != nil {
		return err
	}

	gc.Key = key

	return nil
}

func (m Module) CreateDefaultConfig() ([]byte, error) {
	gc := genesisConfigJsonStruct{}

	return json.Marshal(gc)
}

func (m Module) BuildConfig(config []byte) error {
	gc := GenesisConfig{}
	if err := json.Unmarshal(config, &gc); err != nil {
		return err
	}

	m.storage.Key.Put(gc.Key)

	return nil
}
