package example1

type Thingy1 struct{}

func (t Thingy1) Inst() Thingy1 {
	return t
}
func (t Thingy1) Foo() bool {
	return true
}

var e = new(Thingy1)

func asq_query2() {
	//asq_start
	e.Inst().Foo()
	//asq_end
}
