package common

type E struct{}

func (e *E) Inst() *E { return e }

func (e *E) Foo() {}
