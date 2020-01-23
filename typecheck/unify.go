package typecheck

import (
	"fmt"
	"reflect"
)

type Subs map[Constrainable]Constrainable

func ReplaceCons(check Constrainable, old Constrainable, new Constrainable) Constrainable {
	switch cons := check.(type) {
	case TypeVar:
		if cons == old {
			return new
		}
	case Fun:
		for i, arg := range cons.Args {
			cons.Args[i] = ReplaceCons(arg, old, new)
		}
		cons.Ret = ReplaceCons(cons.Ret, old, new)
		return cons
	case Container:
		_, isIndexer := new.(Indexer)
		if isIndexer {
			return cons.Subtype
		}
		return Container{cons.Type, ReplaceCons(cons.Subtype, old, new), cons.Index}
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

	leftBase, isLeftBase := currCons.Left.(BaseType)
	rightBase, isRightBase := currCons.Right.(BaseType)

	// Unify base types
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

	// Unify type vars
	rightVar, rightIsVar := currCons.Right.(TypeVar)
	leftVar, leftIsVar := currCons.Left.(TypeVar)
	if rightIsVar && leftIsVar && currCons.Left == currCons.Right {
		constraints[curr] = Constraint{}
		return Unify(constraints, subs, curr+1)
	}
	if rightIsVar && leftIsVar {
		subs[currCons.Left] = rightVar
		ReplaceAllCons(constraints, leftVar, rightVar)
		return Unify(constraints, subs, curr+1)
	}
	if rightIsVar && !leftIsVar {
		constraints[curr] = Constraint{currCons.Right, currCons.Left}
		return Unify(constraints, subs, curr)
	}

	rightFun, rightIsFun := currCons.Right.(Fun)
	leftFun, leftIsFun := currCons.Left.(Fun)
	if rightIsFun && leftIsFun {
		if len(rightFun.Args) != len(leftFun.Args) {
			panic("Unified functions don't have equal arg counts")
		}
		for k, arg := range leftFun.Args {
			constraints = append(constraints, Constraint{arg, rightFun.Args[k]})
		}
		constraints = append(constraints, Constraint{leftFun.Ret, rightFun.Ret})
		return Unify(constraints, subs, curr+1)
	}
	if rightIsFun && !leftIsFun {
		subs[currCons.Left] = rightFun
		ReplaceAllCons(constraints, currCons.Left, rightFun)
		return Unify(constraints, subs, curr+1)
	}

	// Unify containers
	rightContainer, isRightContainer := currCons.Right.(Container)
	if leftIsVar && isRightContainer {
		subs[currCons.Left] = rightContainer
		ReplaceAllCons(constraints, leftVar, rightContainer)
		return Unify(constraints, subs, curr+1)
	}

	rightIndexer, isRightIndexer := currCons.Right.(Indexer)
	if leftIsVar && isRightIndexer {
		ReplaceAllCons(constraints, leftVar, rightIndexer)
		return Unify(constraints, subs, curr+1)
	}

	panic(fmt.Sprintf("Unable to unify '%s' and '%s'", reflect.TypeOf(currCons.Left), reflect.TypeOf(currCons.Right)))

	return Unify(constraints, subs, curr+1)
}
