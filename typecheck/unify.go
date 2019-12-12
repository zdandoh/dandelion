package typecheck

import (
	"ahead/types"
	"fmt"
)

type Substitutions map[TypeVar]types.Type

func (s Substitutions) Copy() Substitutions {
	newSubs := make(Substitutions)
	for k, v := range s {
		newSubs[k] = v
	}

	return newSubs
}

func (s Substitutions) String() string {
	str := ""
	for k, v := range s {
		str += fmt.Sprintf("%s = %s | ", k, v.TypeString())
	}

	return str
}

func UnifyConstraints(cons []Constraint) {
	subs := make(Substitutions)

	for _, con := range cons {
		fmt.Println("Unifying:", con.Right.ConsString(), con.Left.ConsString())
		subs = Unify(con.Left, con.Right, subs)
		fmt.Println(subs)
	}
}

func Unify(left Constrainable, right Constrainable, subs Substitutions) Substitutions {
	if subs == nil {
		return nil
	} else if left == right {
		return subs
	}

	leftVar, ok := left.(TypeVar)
	if ok {
		return UnifyVar(leftVar, right, subs)
	}
	rightVar, ok := right.(TypeVar)
	if ok {
		return UnifyVar(rightVar, left, subs)
	}

	leftFun, isLeftFun := left.(Fun)
	rightFun, isRightFun := right.(Fun)
	if isLeftFun && isRightFun {
		if len(leftFun.Args) != len(rightFun.Args) {
			return nil
		}

		subs = Unify(leftFun.Ret, rightFun.Ret, subs)
		for i := 0; i < len(leftFun.Args); i++ {
			subs = Unify(leftFun.Args[i], rightFun.Args[i], subs)
		}
		return subs
	}

	return nil
}

func UnifyVar(tVar TypeVar, con Constrainable, subs Substitutions) Substitutions {
	_, ok := subs[tVar]
	if ok {
		return Unify(BaseType{subs[tVar]}, con, subs)
	}

	conVar, conIsVar := con.(TypeVar)
	if conIsVar {
		conType, conInSubs := subs[conVar]
		if conInSubs {
			return Unify(tVar, BaseType{conType}, subs)
		}
	}

	newSubs := subs.Copy()
	newSubs[tVar] = con.(BaseType).Type
	return newSubs
}

func Contains(tVar TypeVar, con Constrainable, subs Substitutions) bool {
	if tVar == con {
		return true
	}

	return false
}
