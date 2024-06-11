---
layout: default
title: API
permalink: /development/api
---

# Runtime API modules

Contains modules of the [Runtime API Specification](https://spec.polkadot.network/chap-runtime-api).

## Supported modules

| Name                                                                                                         | Description                                                               |
|--------------------------------------------------------------------------------------------------------------|---------------------------------------------------------------------------|
| [AccountNonceApi](https://github.com/limechain/gosemble/tree/develop/api/account_nonce)                      | Provides logic to get an account's nonce.                                 |
| [AuraApi](https://github.com/limechain/gosemble/tree/develop/api/aura)                                       | Manages block authoring AuRa consensus mechanism.                         |
| [BabeApi](https://github.com/limechain/gosemble/tree/develop/api/babe)                                       | Manages block authoring BABE consensus mechanism.                         |
| [Benchmarking](https://github.com/limechain/gosemble/tree/develop/api/benchmarking)                          | Provides functionality for benchmarking extrinsic calls and system hooks. |
| [BlockBuilder](https://github.com/limechain/gosemble/tree/develop/api/block_builder)                         | Provides functionality for building and finalizing a block.               |
| [Core](https://github.com/limechain/gosemble/tree/develop/api/core)                                          | Provides functionality for initialising and executing a block.            |
| [GenesisBuilder](https://github.com/limechain/gosemble/tree/develop/api/genesis_builder)                     | Builds genesis configuration.                                             |
| [GrandpaApi](https://github.com/limechain/gosemble/tree/develop/api/grandpa)                                 | Manages the GRANDPA block finalization.                                   |
| [Metadata](https://github.com/limechain/gosemble/tree/develop/api/metadata)                                  | Returns the metadata of the runtime                                       |
| [OffchainWorkerApi](https://github.com/limechain/gosemble/tree/develop/api/offchain_worker)                  | Provides functionality to start offchain worker operations.               |
| [SessionKeys](https://github.com/limechain/gosemble/tree/develop/api/session_keys)                           | Generates and decodes session keys                                        |
| [TaggedTransactionQueue](https://github.com/limechain/gosemble/tree/develop/api/tagged_transaction_queue)    | Validates transactions in the transaction queue.                          |
| [TransactionPaymentApi](https://github.com/limechain/gosemble/tree/develop/api/transaction_payment)          | Queries the runtime for transaction fees.                                 |
| [TransactionPaymentCallApi](https://github.com/limechain/gosemble/tree/develop/api/transaction_payment_call) | Queries the runtime for transaction call fees.                            |

## Structure

Each module must have the following:

* `name`
* `version`
* `exported` runtime functions
    * Each runtime-exported function takes care of reading from and writing to the WASM memory.
* `metadata` definition
* Explanatory `inline documentation`

### Example

An `example` module, which takes care of returning account balances.

```go
package example

// imports

// Module implements the Example Runtime API.
type Module struct {
    memUtils utils.WasmMemoryTranslator
    logger   log.Logger
}

func New(logger log.Logger) Module {
    return Module{
        memUtils: utils.NewMemoryTranslator(),
        logger:   logger,
    }
}

// Name returns the name of the api module.
func (m Module) Name() string {
    return "ExampleModule"
}

// Item returns the first 8 bytes of the Blake2b hash of the name and version of the api module.
func (m Module) Item() types.ApiItem {
    hash := hashing.MustBlake2b8([]byte(ApiModuleName))
    return types.NewApiItem(hash, apiVersion)
}

// AccountBalance returns the balance of given AccountId.
// It takes two arguments:
// - dataPtr: Pointer to the data in the Wasm memory.
// - dataLen: Length of the data.
// which represent the SCALE-encoded AccountId.
// Returns a pointer-size of the SCALE-encoded balance of the account.
// <Link to Specification if found>
func (m Module) AccountBalance(dataPtr int32, dataLen int32) int64 {
    b := m.memUtils.GetWasmMemorySlice(dataPtr, dataLen)
    buffer := bytes.NewBuffer(b)
    
    accountId, err := types.DecodeAccountId(buffer)
    if err != nil {
        m.logger.Critical(err.Error())
    }
	
    // Logic to get account balance
	
    return m.memUtils.BytesToOffsetAndSize(accountBalance)
}

// Metadata returns the runtime api metadata of the module.
func (m Module) Metadata() types.RuntimeApiMetadata {
    // Metadata declaration of the module, which includes:
    // * Name of the module
    // * Definition of each runtime exported function (AccountBalance).
    methods := sc.Sequence[types.RuntimeApiMethodMetadata]{
        types.RuntimeApiMethodMetadata{
            Name: "account_balance",
            Inputs: sc.Sequence[types.RuntimeApiMethodParamMetadata]{
                types.RuntimeApiMethodParamMetadata{
                    Name: "account",
                    Type: sc.ToCompact(metadata.TypesAddress32),
                },
            },
            Output: sc.ToCompact(metadata.PrimitiveTypesU128),
            Docs:   sc.Sequence[sc.Str]{" Get current account balance of given `AccountId`."},
        },
    }
    
    return types.RuntimeApiMetadata{
        Name:    "ExampleModule",
        Methods: methods,
        Docs:    sc.Sequence[sc.Str]{" The Example API to query account balances."},
    }
}
```


