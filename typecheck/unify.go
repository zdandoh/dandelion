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
	subs := make(map[TypeVar]types.Type)

	for _, con := range cons {
		fmt.Println("Unifying:", con)
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