package grandpa

import (
	"encoding/json"
	"errors"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/primitives/types"
	"github.com/vedhavyas/go-subkey"
)

var (
	errAuthoritiesAlreadyInitialized = errors.New("Authorities are already initialized!")
	errInvalidAddrValue              = errors.New("invalid address in genesis config json")
	errInvalidWeightValue            = errors.New("invalid weight in genesis config json")
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

func (m Module) CreateDefaultConfig() ([]byte, error) {
	gc := &genesisConfigJsonStruct{}
	gc.GrandpaGenesisConfig.Authorities = [][2]interface{}{}

	return json.Marshal(gc)
}

func (m Module) BuildConfig(config []byte) error {
	gc := GenesisConfig{}
	if err := json.Unmarshal(config, &gc); err != nil {
		return err
	}

	// todo missing
	// CurrentSetId::<T>::put(SetId::default());

	if len(gc.Authorities) == 0 {
		return nil
	}

	authorities, err := m.storage.Authorities.Get()
	if err != nil {
		return err
	}

	if len(authorities) > 0 {
		return errAuthoritiesAlreadyInitialized
	}

	// todo missing
	// &BoundedAuthorityList::<T::MaxAuthorities>::try_from(authorities).expect(
	// 	"Grandpa: `Config::MaxAuthorities` is smaller than the number of genesis authorities!",
	// ),

	m.storage.Authorities.Put(gc.Authorities)

	// todo missing
	//// NOTE: initialize first session of first set. this is necessary for
	//// the genesis set and session since we only update the set -> session
	//// mapping whenever a new session starts, i.e. through `on_new_session`.
	// SetIdSession::<T>::insert(0, 0);

	return nil
}
