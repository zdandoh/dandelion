package infer

import (
	"dandelion/ast"
	"dandelion/types"
	"fmt"
	"strings"
	"testing"
)

//func TestInfer(t *testing.T) {
//	src := `
//a = 1;
//b = a + 4;
//`
//
//	progAst := parser.ParseProgram(src)
//	InferTypes(progAst.Funcs["main"])
//}

func TestViralVarReplacement(t *testing.T) {
	i := NewInferer()
	root := i.NewVar()

	left := i.NewVar()
	right := i.NewVar()

	i.SetRef(left, right)
	i.SetRef(left, root)

	v := i.NewVar()
	i.SetRef(left, v)

	lastV := i.Resolve(v)
	if i.Resolve(root) == lastV && i.Resolve(left) == lastV && i.Resolve(right) == lastV {

	} else {
		t.Fail()
	}
}

func TestViralReplacementBaseAndVar(t *testing.T) {
	i := NewInferer()

	base := i.BaseRef(TypeBase{types.IntType{}})

	root := i.NewVar()

	left := i.NewVar()
	right := i.NewVar()

	i.SetRef(left, root)
	i.SetRef(right, left)
	i.SetRef(right, base)

	lastV := i.Resolve(base)
	if i.Resolve(root) == lastV && i.Resolve(left) == lastV && i.Resolve(right) == lastV {

	} else {
		t.Fail()
	}
}

func TestViralReplacementFunc(t *testing.T) {
	i := NewInferer()

	t1 := i.NewVar()
	t2 := i.NewVar()
	t3 := i.NewVar()
	base := i.BaseRef(TypeBase{types.IntType{}})
	fun := i.FuncRef(KindFunc, t1, t2, t3)
	fun2 := i.FuncRef(KindFunc, t1, t2)

	i.SetRef(t1, base)
	i.SetRef(t3, fun2)

	fmt.Println(i.String(fun))
	if i.String(fun) != "(t2, (t2) -> <int>) -> <int>" {
		t.Fail()
	}
}

func TestTypeFuncContainsSimple(t *testing.T) {
	i := NewInferer()

	t1 := i.NewVar()
	t2 := i.NewVar()
	t3 := i.NewVar()

	fun := i.FuncRef(KindFunc, t1, t2, t3)

	if !i.Contains(i.Resolve(fun).(TypeFunc), t3) || !i.Contains(i.Resolve(fun).(TypeFunc), t1) {
		t.Fail()
	}
}

func TestTypeFuncContainsComplex(t *testing.T) {
	i := NewInferer()

	t1 := i.NewVar()
	t2 := i.NewVar()
	t3 := i.NewVar()

	t4 := i.NewVar()
	t5 := i.NewVar()
	t6 := i.NewVar()

	fun := i.FuncRef(KindFunc, t1, t2, t3)
	fun2 := i.FuncRef(KindFunc, t4, t5, t6)

	i.SetRef(t3, fun2)
	if !i.Contains(i.Resolve(fun).(TypeFunc), t4) || !i.Contains(i.Resolve(fun).(TypeFunc), t6) {
		t.Fail()
	}
}

func TestViralReplaceFuncsPanic(t *testing.T) {
	i := NewInferer()

	t1 := i.NewVar()
	t2 := i.NewVar()
	t3 := i.NewVar()
	t4 := i.NewVar()

	fun := i.FuncRef(KindFunc, t1, t2)
	fun2 := i.FuncRef(KindFunc, t3, t4)
	unrelatedFun := i.FuncRef(KindFunc, t1, t2, t3, t4)
	i.SetRef(fun, fun2)

	defer func() {
		r := recover()
		if r != nil && strings.Contains(r.(string), "replacement would result in recursive overflow") {
			fmt.Println("recovered: ", r)
			return
		}
		t.Fail()
	}()

	i.SetRef(t3, unrelatedFun)
}

func TestSameReplacement(t *testing.T) {
	i := NewInferer()

	t1 := i.NewVar()
	t2 := i.NewVar()

	i.SetRef(t1, t1)
	i.SetRef(t1, t2)

	if i.String(t1) != "t2" {
		t.Fail()
	}
}

func TestReplacement(t *testing.T) {
	i := NewInferer()

	fun := &ast.FunDef{}

	genRef := i.TypeRef(fun)
	funRef := i.FuncRef(KindFunc, i.NewVar())

	fmt.Println(i.String(genRef))
	fmt.Println(i.String(funRef))
	fmt.Println(i.varLibrary)
}

func TestReplaceFunc(t *testing.T) {
	i := NewInferer()

	funRef := i.FuncRef(KindTupleAccess, i.NewVar())
	t3 := i.NewVar()
	t4 := i.NewVar()

	i.SetRef(t3, funRef)
	i.SetRef(funRef, t4)

	if !stringsEqual(i.String(t3), i.String(t4), i.String(funRef)) {
		t.Fail()
	}
}

func TestFuncMeta(t *testing.T) {
	i := NewInferer()

	t1 := i.NewVar()
	meta := i.FuncMeta(45)
	funRef := i.FuncRef(KindTupleAccess, meta, i.NewVar())

	i.SetRef(funRef, t1)
	fmt.Println(i.String(t1), i.String(funRef), i.String(meta))
}

func TestReplaceTwoFunc(t *testing.T) {
	i := NewInferer()

	t1 := i.NewVar()
	fun1 := i.FuncRef(KindTupleAccess, i.NewVar(), i.NewVar())
	fun2 := i.FuncRef(KindArray, i.NewVar(), i.NewVar())

	i.SetRef(t1, fun1)
	i.SetRef(t1, fun2)

	if !stringsEqual(i.String(t1), i.String(fun1), i.String(fun2)) {
		t.Fail()
	}
}

func stringsEqual(strs... string) bool {
	str1 := strs[0]
	for _, str := range strs[1:] {
		if str != str1 {
			return false
		}
	}

	return true
}