package main

type BulletPilot struct{}

func (e *BulletPilot) Inst() *BulletPilot {
	return e
}

func (e *BulletPilot) Foo() bool {
	return e == nil
}

func SourceTest1(e *BulletPilot) string {
	if e.Inst().Foo() {
		return "foo true"
	} else {
		return "foo false"
	}
}

func SourceTest2(e *BulletPilot) {
	e.Inst().Foo()
}

func SourceTest3(e *BulletPilot) {
	e.Inst().Foo()
	e.Foo()
}
