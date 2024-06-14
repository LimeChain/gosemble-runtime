package aura_ext

import (
	"errors"
)

var (
	errAuraAuthoritiesEmpty = errors.New("AuRa authorities empty, check genesis builder")
)

func (m Module) CreateDefaultConfig() ([]byte, error) {
	return nil, nil
}

func (m Module) BuildConfig(_ []byte) error {
	totalAuthorities, err := m.auraModule.StorageAuthorities()
	if err != nil {
		return err
	}

	if len(totalAuthorities) == 0 {
		return errAuraAuthoritiesEmpty
	}

	m.storage.Authorities.Put(totalAuthorities)

	return nil
}
