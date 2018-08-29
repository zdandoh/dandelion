package main

func (f *Function) run() {
	for i := 0; i < len(f.statements); i++ {
		f.statements[i].Run()
	}
}

type AddIntStatement struct {
	a VarInt
	b VarInt
}

func (s *AddIntStatement) Run() Var {
	return s.a + s.b
}
