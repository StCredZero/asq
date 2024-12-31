package test

import "github.com/stcredzero/asq/test_source/common"

func Test1() {
	e := &common.E{}
	e.Inst().Foo()
}

func Test2() {
	e := &common.E{}
	_ = 42 // Intentionally unused
	e.Inst().Foo()
	_ = 43 // Intentionally unused
}
