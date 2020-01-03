package typecheck

import (
	"fmt"
	"reflect"
	"strings"
)

func LookupVar(v TypeVar, subs map[Constrainable]Constrainable) Constrainable {
	cons, ok := subs[v]
	if !ok {
		return v
	}
	_, isBase := cons.(BaseType)
	if isBase {
		return cons
	}

	return LookupVar(cons.(TypeVar), subs)
}

func LookupFunc(fName string, subs map[Constrainable]Constrainable, i *TypeInferer) string {
	argStrs := make([]string, 0)
	baseFun := i.FunLookup[fName]
	for _, arg := range baseFun.Args {
		argStrs = append(argStrs, LookupVar(arg, subs).ConsString())
	}

	return fmt.Sprintf("(%s -> %s)", strings.Join(argStrs, " "), LookupVar(baseFun.Ret, subs).ConsString())
}

func ReplaceCons(constraints []Constraint, old Constrainable, new Constrainable) {
	for i, con := range constraints {
		if con.Left == old {
			con = Constraint{new, con.Right}
		}
		if con.Right == old {
			con = Constraint{con.Left, new}
		}

		constraints[i] = con
	}
}

func Unify2(constraints []Constraint, subs map[Constrainable]Constrainable, curr int) map[Constrainable]Constrainable {
	if curr == len(constraints) {
		return subs
	}
	currCons := constraints[curr]

	if currCons.Left == currCons.Right {
		constraints[curr] = Constraint{}
		return Unify2(constraints, subs, curr+1)
	}

	leftBase, isLeftBase := currCons.Left.(BaseType)
	rightBase, isRightBase := currCons.Right.(BaseType)

	if isLeftBase && isRightBase && leftBase != rightBase {
		panic("Type inference failed, base types not equal")
	}
	if isLeftBase {
		subs[currCons.Right] = leftBase
		ReplaceCons(constraints, currCons.Right, leftBase)
		return Unify2(constraints, subs, curr+1)
	}
	if isRightBase {
		subs[currCons.Left] = rightBase
		ReplaceCons(constraints, currCons.Left, rightBase)
		return Unify2(constraints, subs, curr+1)
	}

	rightVar, rightIsVar := currCons.Right.(TypeVar)
	leftVar, leftIsVar := currCons.Left.(TypeVar)
	if rightIsVar && leftIsVar {
		subs[currCons.Left] = rightVar
		ReplaceCons(constraints, leftVar, rightVar)
		return Unify2(constraints, subs, curr+1)
	}

	fmt.Println(reflect.TypeOf(currCons.Left), reflect.TypeOf(currCons.Right))
	panic("wut")

	return Unify2(constraints, subs, curr+1)
}
