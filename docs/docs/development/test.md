---
layout: default
permalink: /development/test
---

# Testing 🧪

Currently, the project contains unit and integration tests. Integration tests use [Gossamer](https://github.com/LimeChain/gossamer), which
imports all the necessary Host functions and interacts with the Runtime.

Before running the tests, make sure to build the runtime with the benchmarking features enabled, as there are some tests that rely on them.

````bash
make build-benchmarking
````

And then run the tests with:

```bash
make test
```

or

```bash
make test-unit
make test-integration
```

### Debug 🐛

To aid the debugging process, there is a set of functions provided by the logger instance that can be called within the Runtime to log messages.

```go
logger := log.NewLogger()

logger.Critical(message string) // logs and aborts the execution
logger.Warn(message string)
logger.Info(message string)
logger.Debug(message string)
logger.Trace(message string)
```