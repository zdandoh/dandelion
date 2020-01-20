package typecheck

import (
	"fmt"
	"reflect"
	"strings"
)

type Subs map[Constrainable]Constrainable

func LookupVar(v TypeVar, subs Subs) Constrainable {
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

func LookupFunc(fName string, subs Subs, i *TypeInferer) string {
	argStrs := make([]string, 0)
	baseFun := i.FunLookup[fName]
	for _, arg := range baseFun.Args {
		argStrs = append(argStrs, LookupVar(arg, subs).ConsString())
	}

	return fmt.Sprintf("(%s -> %s)", strings.Join(argStrs, " "), LookupVar(baseFun.Ret, subs).ConsString())
}

func ReplaceCons(check Constrainable, old Constrainable, new Constrainable) Constrainable {
	switch cons := check.(type) {
	case TypeVar:
		if cons == old {
			return new
		}
	case Container:
		if cons.Subtype == old {
			return Container{cons.Type, new, cons.Index}
		}
	}

	return check
}

func ReplaceAllCons(constraints []Constraint, old Constrainable, new Constrainable) {
	for i, con := range constraints {
		newLeft := ReplaceCons(con.Left, old, new)
		newRight := ReplaceCons(con.Right, old, new)

		constraints[i] = Constraint{newLeft, newRight}
	}
}

func Unify(constraints []Constraint, subs Subs, curr int) Subs {
	if curr == len(constraints) {
		return subs
	}
	currCons := constraints[curr]

	if currCons.Left == currCons.Right {
		constraints[curr] = Constraint{}
		return Unify(constraints, subs, curr+1)
	}

	leftBase, isLeftBase := currCons.Left.(BaseType)
	rightBase, isRightBase := currCons.Right.(BaseType)

	if isLeftBase && isRightBase && leftBase != rightBase {
		panic("Type inference failed, base types not equal")
	}
	if isLeftBase {
		subs[currCons.Right] = leftBase
		ReplaceAllCons(constraints, currCons.Right, leftBase)
		return Unify(constraints, subs, curr+1)
	}
	if isRightBase {
		subs[currCons.Left] = rightBase
		ReplaceAllCons(constraints, currCons.Left, rightBase)
		return Unify(constraints, subs, curr+1)
	}

	rightVar, rightIsVar := currCons.Right.(TypeVar)
	leftVar, leftIsVar := currCons.Left.(TypeVar)
	if rightIsVar && leftIsVar {
		subs[currCons.Left] = rightVar
		ReplaceAllCons(constraints, leftVar, rightVar)
		return Unify(constraints, subs, curr+1)
	}

	_, isContainer := currCons.Right.(Container)
	if leftIsVar && isContainer {
		// Don't need to do anything special to containers
		return Unify(constraints, subs, curr+1)
	}

	panic(fmt.Sprintf("Unable to unify '%s' and '%s'", reflect.TypeOf(currCons.Left), reflect.TypeOf(currCons.Right)))

	return Unify(constraints, subs, curr+1)
}
