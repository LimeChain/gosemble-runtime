package session

import (
	"bytes"
	"encoding/json"
	"errors"
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/primitives/types"
	"github.com/vedhavyas/go-subkey"
)

var (
	errInvalidAddrValue          = errors.New("invalid address in genesis config json")
	errInvalidConsensusKeysValue = errors.New("invalid consensus keys in genesis config json")
	errInvalidValidatorValue     = errors.New("invalid validator in genesis config json")
)

type genesisConfigKeys struct {
	AccountId types.AccountId
	Validator types.AccountId
	Keys      sc.Sequence[types.SessionKey]
}

type GenesisConfig struct {
	Keys []genesisConfigKeys
}

type genesisConfigJsonStruct struct {
	SessionGenesisConfig struct {
		Keys [][3]interface{} `json:"keys"`
	} `json:"session"`
}

func (gc *GenesisConfig) UnmarshalJSON(data []byte) error {
	gcJson := genesisConfigJsonStruct{}

	if err := json.Unmarshal(data, &gcJson); err != nil {
		return err
	}

	for _, keyString := range gcJson.SessionGenesisConfig.Keys {
		accountString, ok := keyString[0].(string)
		if !ok {
			return errInvalidAddrValue
		}

		_, acc, err := subkey.SS58Decode(accountString)
		if err != nil {
			return err
		}

		accId, err := types.NewAccountId(sc.BytesToSequenceU8(acc)...)
		if err != nil {
			return err
		}

		validatorStrimg, ok := keyString[1].(string)
		if !ok {
			return errInvalidValidatorValue
		}

		_, validator, err := subkey.SS58Decode(validatorStrimg)
		if err != nil {
			return err
		}

		validatorId, err := types.NewAccountId(sc.BytesToSequenceU8(validator)...)
		if err != nil {
			return err
		}

		consensusKeys, ok := keyString[2].(map[string]string)
		if !ok {
			return errInvalidConsensusKeysValue
		}

		configKeys := genesisConfigKeys{
			AccountId: accId,
			Validator: validatorId,
		}

		for consensusKeyTypeId, consensusKey := range consensusKeys {
			keyTypeId, err := sc.DecodeFixedSequence[sc.U8](4, bytes.NewBuffer([]byte(consensusKeyTypeId)))
			if err != nil {
				return err
			}

			// TODO: check if actual consensus key type id is necessary

			_, key, err := subkey.SS58Decode(consensusKey)
			if err != nil {
				return err
			}

			sr25519PublicKey, err := types.NewSr25519PublicKey(sc.BytesToSequenceU8(key)...)
			if err != nil {
				return err
			}

			configKeys.Keys = append(configKeys.Keys, types.NewSessionKeyFromBytes(sr25519PublicKey.Bytes(), keyTypeId))
		}

		gc.Keys = append(gc.Keys, configKeys)
	}

	return nil
}

func (m Module) CreateDefaultConfig() ([]byte, error) {
	gc := genesisConfigJsonStruct{}
	gc.SessionGenesisConfig.Keys = [][3]interface{}{}

	return json.Marshal(gc)
}

func (m Module) BuildConfig(config []byte) error {
	gc := GenesisConfig{}
	if err := json.Unmarshal(config, &gc); err != nil {
		return err
	}

	for _, sessionKeysConfig := range gc.Keys {
		_, err := m.InnerSetKeys(sessionKeysConfig.Validator, sessionKeysConfig.Keys)
		if err != nil {
			return err
		}

		_, err = m.systemModule.IncConsumersWithoutLimit(sessionKeysConfig.AccountId)
		if errors.Is(err, types.NewDispatchErrorNoProviders()) {
			_, err := m.systemModule.IncProviders(sessionKeysConfig.AccountId)
			if err != nil {
				return err
			}
		}
	}

	var validators sc.Sequence[types.AccountId]
	initialValidators := m.manager.NewSessionGenesis(0)
	if !initialValidators.HasValue {
		m.logger.Warn("No initial validator provided by `SessionManager`, use session config keys to generate the initial validator set.")
		validators = getValidators(gc.Keys)
	} else {
		validators = initialValidators.Value
	}

	var nextSessionValidators sc.Sequence[types.AccountId]
	secondSession := m.manager.NewSessionGenesis(1)
	if !secondSession.HasValue {
		nextSessionValidators = validators
	} else {
		nextSessionValidators = secondSession.Value
	}

	queuedKeys, err := m.buildQueuedKeys(nextSessionValidators)
	if err != nil {
		return err
	}

	err = m.handler.OnGenesisSession(queuedKeys)
	if err != nil {
		return err
	}
	m.storage.Validators.Put(validators)
	m.storage.QueuedKeys.Put(queuedKeys)

	m.manager.StartSession(0)
	return nil
}

func (m Module) buildQueuedKeys(validators sc.Sequence[types.AccountId]) (sc.Sequence[queuedKey], error) {
	var result sc.Sequence[queuedKey]

	keyTypeIds := m.handler.KeyTypeIds()

	for _, validator := range validators {
		nextKeys, err := m.storage.NextKeys.Get(validator)
		if err != nil {
			return nil, err
		}

		sessionKeys, err := toSessionKeys(keyTypeIds, nextKeys)
		if err != nil {
			return nil, err
		}

		result = append(result, queuedKey{
			Validator: validator,
			Keys:      sessionKeys,
		})
	}

	return result, nil
}

func getValidators(genesisConfig []genesisConfigKeys) sc.Sequence[types.AccountId] {
	var result sc.Sequence[types.AccountId]
	for _, config := range genesisConfig {
		result = append(result, config.Validator)
	}

	return result
}

func toPublicKeys(sessionKeys sc.Sequence[types.SessionKey]) sc.FixedSequence[types.Sr25519PublicKey] {
	var result sc.FixedSequence[types.Sr25519PublicKey]

	for _, sessionKey := range sessionKeys {
		key := sc.SequenceU8ToBytes(sessionKey.Key)
		result = append(result, types.Sr25519PublicKey{FixedSequence: sc.BytesToFixedSequenceU8(key)})
	}

	return result
}

func toSessionKeys(keyTypeIds sc.Sequence[sc.FixedSequence[sc.U8]], keys sc.FixedSequence[types.Sr25519PublicKey]) (sc.Sequence[types.SessionKey], error) {
	if len(keys) > len(keyTypeIds) {
		return nil, errors.New("invalid length")
	}

	var sessionKeys sc.Sequence[types.SessionKey]
	for i, nextKey := range keys {
		sessionKeys = append(sessionKeys, types.SessionKey{
			Key:    sc.BytesToSequenceU8(sc.FixedSequenceU8ToBytes(nextKey.FixedSequence)),
			TypeId: keyTypeIds[i],
		})
	}

	return sessionKeys, nil
}
