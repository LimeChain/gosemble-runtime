package genesisbuilder

import (
	"bytes"
	"fmt"
	"strings"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/primitives/hashing"
	"github.com/LimeChain/gosemble/primitives/log"
	primitives "github.com/LimeChain/gosemble/primitives/types"
	"github.com/LimeChain/gosemble/utils"
)

const (
	ApiModuleName = "GenesisBuilder"
	apiVersion    = 1
)

type GenesisBuilder interface {
	CreateDefaultConfig() ([]byte, error)
	BuildConfig(config []byte) error
}

// Module implements the GenesisBuilder Runtime API definition.
//
// For more information about API definition, see:
// https://github.com/paritytech/polkadot-sdk/blob/master/substrate/primitives/genesis-builder/src/lib.rs#L38
type Module struct {
	modules  []primitives.Module
	memUtils utils.WasmMemoryTranslator
	logger   log.RuntimeLogger
}

func New(modules []primitives.Module, logger log.RuntimeLogger) Module {
	return Module{
		modules:  modules,
		memUtils: utils.NewMemoryTranslator(),
		logger:   logger,
	}
}

// Name returns the name of the api module.
func (m Module) Name() string {
	return ApiModuleName
}

// Item returns the first 8 bytes of the Blake2b hash of the name and version of the api module.
func (m Module) Item() primitives.ApiItem {
	hash := hashing.MustBlake2b8([]byte(ApiModuleName))
	return primitives.NewApiItem(hash, apiVersion)
}

// CreateDefaultConfig returns the default genesis configuration of the runtime.
// Includes all modules' JSON default configuration.
// Returns a pointer-size of the serialised JSON representation of the default genesis configuration.
func (m Module) CreateDefaultConfig() int64 {
	gcs := []string{}
	for _, module := range m.modules {
		genesisBuilder, ok := module.(GenesisBuilder)
		if !ok {
			continue
		}

		gcJsonBytes, err := genesisBuilder.CreateDefaultConfig()
		if err != nil {
			m.logger.Critical(err.Error())
		}

		if len(gcJsonBytes) == 0 {
			continue
		}

		// gcJsonBytes[1:len(gcJsonBytes)-1] trims first and last characters which represent start and end of the json
		// CreateDefaultConfig returns a valid json (e.g. {"system":{}}), and here we need it as a json field
		gcs = append(gcs, string(gcJsonBytes[1:len(gcJsonBytes)-1]))
	}

	gcJson := []byte(fmt.Sprintf("{%s}", strings.Join(gcs, ",")))

	return m.memUtils.BytesToOffsetAndSize(sc.BytesToSequenceU8(gcJson).Bytes())
}

// BuildConfig validates the genesis configuration and stores it in the storage.
// It takes two arguments:
// - dataPtr: Pointer to the data in the Wasm memory.
// - dataLen: Length of the data.
// which represent the SCALE-encoded serialised JSON genesis configuration.
// The serialised bytes must contain the genesis configuration for each runtime module.
func (m Module) BuildConfig(dataPtr int32, dataLen int32) int64 {
	gcJsonBytes := m.memUtils.GetWasmMemorySlice(dataPtr, dataLen)
	gcDecoded, err := sc.DecodeSequence[sc.U8](bytes.NewBuffer(gcJsonBytes))
	if err != nil {
		m.logger.Critical(err.Error())
	}

	gcDecodedBytes := sc.SequenceU8ToBytes(gcDecoded)

	for _, module := range m.modules {
		genesisBuilder, ok := module.(GenesisBuilder)
		if !ok {
			continue
		}

		if err := genesisBuilder.BuildConfig(gcDecodedBytes); err != nil {
			m.logger.Critical(err.Error())
		}
	}

	return m.memUtils.BytesToOffsetAndSize([]byte{0})
}

// todo: metadata
// func (m Module) Metadata() primitives.RuntimeApiMetadata {
// 	return primitives.RuntimeApiMetadata{}
// }
