---
layout: default
title: Modules
permalink: /development/modules
---

# Core Modules 🧱

Contains core modules that can be used and configured for runtime development and business logic.

## Supported modules

### System modules

| Name                                                                            | Description                                                                                                                                 |
|---------------------------------------------------------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------|
| [executive](https://github.com/limechain/gosemble/tree/develop/frame/executive) | A non-typical module, which wraps around the system module and provides functionality for block production and offchain workers.            |
| [support](https://github.com/limechain/gosemble/tree/develop/frame/support)     | A non-typical module, defining types for storage variables and logic for transactional layered execution.                                   |
| [system](https://github.com/limechain/gosemble/tree/develop/frame/system)       | Manages the core storage items of the runtime, such as extrinsic data, indices, event records. Executes block production and deposits logs. |

### Functional modules

These modules provide features than can be useful for your blockchain and can be plugged to your runtime code.

| Name                                                                                                | Description                                                                   |
|-----------------------------------------------------------------------------------------------------|-------------------------------------------------------------------------------|
| [aura](https://github.com/limechain/gosemble/tree/develop/frame/aura)                               | Manages the AuRa (Authority Round) consensus mechanism.                       |
| [balances](https://github.com/limechain/gosemble/tree/develop/frame/balances)                       | Provides functionality for handling accounts and balances of native currency. |
| [grandpa](https://github.com/limechain/gosemble/tree/develop/frame/grandpa)                         | Manages the GRANDPA block finalization.                                       |
| [session](https://github.com/limechain/gosemble/tree/develop/frame/session)                         | Allows validators to manage their session keys, handles session rotation.     |
| [timestamp](https://github.com/limechain/gosemble/tree/develop/frame/timestamp)                     | Manages on-chain time.                                                        |
| [transaction payment](https://github.com/limechain/gosemble/tree/develop/frame/transaction_payment) | Manages pre-dispatch execution fees.                                          |       

### Parachain modules

In addition to functional modules, which are useful for any blockchain, there are modules that provide features specifically for blockchain integration with a relay chain.

| Name                                                                                      | Description              |
|-------------------------------------------------------------------------------------------|--------------------------|
| [parachain_info](https://github.com/limechain/gosemble/tree/develop/frame/parachain_info) | Stores the parachain id. |


## Structure

Each module has the following structure:

* `configuration` - types and parameters on which the module depends on.
* `constants` - Constants and immutable parameters from the configuration.
* `calls` - a set of extrinsic calls that define the module's functionality.
* `errors` - dispatched during extrinsic calls execution.
* `events` - declaration of events, emitted during extrinsic calls execution.
* `genesis builder` - genesis configuration definition.
* `storage` - lists all storage keys that can be modified by the given module during extrinsic call state transition.
* `types` - type definitions for the module.

### File structure

We recommend the following structure for a module:

```
├── frame
│   ├── <module_name> # Name of the module
│   │   ├── call_1_.go # extrinsic call functionality for call 1
│   │   ├── ...
│   │   ├── call_N_.go # extrinsic call functionality for call N
│   │   ├── config.go
│   │   ├── constants.go
│   │   ├── errors.go
│   │   ├── events.go
│   │   ├── genesis_builder.go
│   │   ├── module.go # unites all the components of the module
│   │   ├── storage.go
│   │   ├── types.go
│   ├── <module_name_2>
│   │   ├── ...
│   ├── ...
```

Unnecessary components of the module **can be omitted**.
