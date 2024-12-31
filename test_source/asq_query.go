package test

import "github.com/stcredzero/asq/test_source/common"

func asq_end() {}

func asq_query() {
	e := &common.E{}
	e.Inst().Foo()
	asq_end()
}
