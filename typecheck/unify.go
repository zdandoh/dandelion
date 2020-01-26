package typecheck

import (
	"ahead/types"
	"fmt"
	"github.com/pkg/errors"
	"reflect"
)

type Subs map[Constrainable]Constrainable

type Unifier struct {
	cons []Constraint
	subs Subs
}

func NewUnifier(cons []Constraint) *Unifier {
	u := &Unifier{}
	u.cons = cons
	u.subs = make(Subs)

	return u
}

func Equals(c1 Constrainable, c2 Constrainable) bool {
	switch cons := c1.(type) {
	case TypeVar:
		other, same := c2.(TypeVar)
		if same && cons == other {
			return true
		}
		return false
	case BaseType:
		other, same := c2.(BaseType)
		if same && types.Equals(cons.Type, other.Type) {
			return true
		}
		return false
	case Fun:
		other, same := c2.(Fun)
		if same && Equals(cons.Ret, other.Ret) && len(cons.Args) == len(other.Args) {
			for k, arg := range cons.Args {
				if !Equals(arg, other.Args[k]) {
					return false
				}
			}
			return true
		}
		return false
	case Container:
		other, same := c2.(Container)
		if same && cons.Index == other.Index && cons.Type == other.Type && Equals(cons.Subtype, other.Subtype) {
			return true
		}
		return false
	case Tup:
		other, same := c2.(Tup)
		if same && len(cons.Subtypes) == len(other.Subtypes) {
			for k, sub := range cons.Subtypes {
				if !Equals(sub, other.Subtypes[k]) {
					return false
				}
			}
			return true
		}
		return false
	default:
		panic("Unknown constrainable in Equals check")
	}
}

func ReplaceCons(check Constrainable, old Constrainable, new Constrainable) Constrainable {
	if Equals(check, old) {
		return new
	}

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
		return Container{cons.Type, ReplaceCons(cons.Subtype, old, new), cons.Index}
	case Tup:
		newSubs := make([]Constrainable, 0)
		for _, sub := range cons.Subtypes {
			newSubs = append(newSubs, ReplaceCons(sub, old, new))
		}

		return Tup{newSubs}
	}

	return check
}

func (u *Unifier) ReplaceAllCons(old Constrainable, new Constrainable) {
	for i, con := range u.cons {
		newLeft := ReplaceCons(con.Left, old, new)
		newRight := ReplaceCons(con.Right, old, new)

		u.cons[i] = Constraint{newLeft, newRight}
	}
}

func (u *Unifier) UnifyAll() (Subs, error) {
	for i := 0; i < len(u.cons); i++ {
		err := u.Unify(u.cons[i])
		if err != nil {
			return nil, err
		}
	}

	return u.subs, nil
}

func (u *Unifier) Unify(currCons Constraint) error {
	// Skip same sided vars
	if Equals(currCons.Left, currCons.Right) {
		DebugInfer("Skipping equal constraint", currCons.Left.ConsString(), "=", currCons.Right.ConsString())
		return nil
	}

	leftBase, isLeftBase := currCons.Left.(BaseType)
	rightBase, isRightBase := currCons.Right.(BaseType)

	// Unify base types
	if isLeftBase && isRightBase && leftBase != rightBase {
		return fmt.Errorf("type inference failed, base types not equal")
	}
	if isLeftBase {
		u.subs[currCons.Right] = leftBase
		u.ReplaceAllCons(currCons.Right, leftBase)
		return nil
	}
	if isRightBase {
		u.subs[currCons.Left] = rightBase
		u.ReplaceAllCons(currCons.Left, rightBase)
		return nil
	}

	// Unify type vars
	rightVar, rightIsVar := currCons.Right.(TypeVar)
	leftVar, leftIsVar := currCons.Left.(TypeVar)
	if rightIsVar && leftIsVar {
		u.subs[currCons.Left] = rightVar
		u.ReplaceAllCons(leftVar, rightVar)
		return nil
	}
	if rightIsVar && !leftIsVar {
		return u.Unify(Constraint{currCons.Right, currCons.Left})
	}

	rightFun, rightIsFun := currCons.Right.(Fun)
	leftFun, leftIsFun := currCons.Left.(Fun)
	if rightIsFun && leftIsFun {
		if len(rightFun.Args) != len(leftFun.Args) {
			panic("Unified functions don't have equal arg counts")
		}
		for k, arg := range leftFun.Args {
			u.cons = append(u.cons, Constraint{arg, rightFun.Args[k]})
		}
		u.cons = append(u.cons, Constraint{leftFun.Ret, rightFun.Ret})
		return nil
	}
	if rightIsFun && !leftIsFun {
		u.subs[currCons.Left] = rightFun
		u.ReplaceAllCons(currCons.Left, rightFun)
		return nil
	}

	// Unify containers
	rightContainer, isRightContainer := currCons.Right.(Container)
	leftContainer, isLeftContainer := currCons.Left.(Container)
	if leftIsVar && isRightContainer {
		u.subs[currCons.Left] = rightContainer
		u.ReplaceAllCons(leftVar, rightContainer)
		return nil
	}
	if isLeftContainer && isRightContainer {
		nullT := types.NullType{}

		var new Constrainable
		var old Constrainable
		if rightContainer.Type != nullT {
			u.ReplaceAllCons(leftContainer, rightContainer)
			new = rightContainer
			old = leftContainer
		} else if leftContainer.Type != nullT {
			u.ReplaceAllCons(rightContainer, leftContainer)
			new = leftContainer
			old = rightContainer
		}
		u.subs[old] = new

		u.cons = append(u.cons, Constraint{leftContainer.Subtype, rightContainer.Subtype})
		return nil
	}

	// Unify tuples
	rightTuple, isRightTuple := currCons.Right.(Tup)
	leftTuple, isLeftTuple := currCons.Left.(Tup)
	if leftIsVar && isRightTuple {
		u.subs[currCons.Left] = rightTuple
		u.ReplaceAllCons(leftVar, rightTuple)
		return nil
	}
	if isLeftTuple && isRightContainer {
		return u.Unify(Constraint{currCons.Right, currCons.Left})
	}
	if isRightTuple && isLeftContainer {
		if leftContainer.Index < 0 || leftContainer.Index >= len(rightTuple.Subtypes) {
			return fmt.Errorf("illegal index for tuple:", leftContainer.Index)
		}

		u.cons = append(u.cons, Constraint{leftContainer.Subtype, rightTuple.Subtypes[leftContainer.Index]})
		u.subs[leftContainer] = rightTuple
		u.ReplaceAllCons(leftContainer, rightTuple)

		return nil
	}
	if isLeftTuple && isRightTuple {
		if len(leftTuple.Subtypes) != len(rightTuple.Subtypes) {
			return fmt.Errorf("cannot unify, tuples have different subtype counts")
		}
		for k, sub := range leftTuple.Subtypes {
			u.cons = append(u.cons, Constraint{sub, rightTuple.Subtypes[k]})
		}
		return nil
	}

	fmt.Println(currCons)
	return errors.Errorf("unable to unify '%v' and '%v'", reflect.TypeOf(currCons.Left), reflect.TypeOf(currCons.Right))
}
