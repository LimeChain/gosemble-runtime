package system

import (
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/hooks"
)

type defaultOnSetCode struct {
	module Module
}

func NewDefaultOnSetCode(module Module) hooks.OnSetCode {
	return defaultOnSetCode{module}
}

// What to do if the runtime wants to change the code to something new.
//
// The default implementation is responsible for setting the correct storage
// entry and emitting corresponding event and log item. (see
// It's unlikely that this needs to be customized, unless you are writing a parachain using
// `Cumulus`, where the actual code change is deferred.
func (d defaultOnSetCode) SetCode(codeBlob sc.Sequence[sc.U8]) error {
	d.module.UpdateCodeInStorage(codeBlob)
	return nil
}
