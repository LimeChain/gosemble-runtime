package babe

import (
	"errors"
	"testing"

	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/mocks"
)

var babeModule *mocks.BabeModule

func Test_ExternalTrigger_Trigger(t *testing.T) {
	target := ExternalTrigger{}

	target.Trigger(0)
}

func Test_SameAuthoritiesForever_Trigger(t *testing.T) {
	babeModule = new(mocks.BabeModule)

	target := SameAuthoritiesForever{Babe: babeModule}

	babeModule.On("ShouldEpochChange", sc.U64(0)).Return(true)
	babeModule.On("StorageAuthorities").Return(authorities, nil)
	babeModule.On("EnactEpochChange", authorities, nextAuthorities, sc.NewOption[sc.U32](nil)).Return(nil)

	target.Trigger(0)

	babeModule.AssertCalled(t, "ShouldEpochChange", sc.U64(0))
	babeModule.AssertCalled(t, "StorageAuthorities")
	babeModule.AssertCalled(t, "EnactEpochChange", authorities, nextAuthorities, sc.NewOption[sc.U32](nil))
}

func Test_SameAuthoritiesForever_Trigger_Fail(t *testing.T) {
	babeModule = new(mocks.BabeModule)

	someError := errors.New("error")

	target := SameAuthoritiesForever{Babe: babeModule}

	babeModule.On("ShouldEpochChange", sc.U64(0)).Return(true)
	babeModule.On("StorageAuthorities").Return(authorities, someError)

	target.Trigger(0)

	babeModule.AssertCalled(t, "ShouldEpochChange", sc.U64(0))
	babeModule.AssertCalled(t, "StorageAuthorities")
	babeModule.AssertNotCalled(t, "EnactEpochChange", authorities, nextAuthorities, sc.NewOption[sc.U32](nil))
}
