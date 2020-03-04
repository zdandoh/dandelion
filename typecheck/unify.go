package typecheck

import (
	"ahead/ast"
	"ahead/types"
	"fmt"
	"github.com/pkg/errors"
	"reflect"
)

type Subs map[Constrainable]Constrainable

type Unifier struct {
	cons  []Constraint
	subs  Subs
	prog  *ast.Program
	funcs map[string]Fun
}

func NewUnifier(cons []Constraint, prog *ast.Program, funcs map[string]Fun) *Unifier {
	u := &Unifier{}
	u.cons = cons
	u.subs = make(Subs)
	u.prog = prog
	u.funcs = funcs

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
	case StructOptions:
		other, same := c2.(StructOptions)
		if same && len(cons.Dependants) == len(other.Dependants) && len(cons.Types) == len(cons.Types) {
			for k, v := range cons.Dependants {
				v2, ok := other.Dependants[k]
				if !ok || v != v2 {
					return false
				}
			}
			for k, t := range cons.Types {
				if !types.Equals(t, other.Types[k]) {
					return false
				}
			}
			return true
		}
		return false
	case Coroutine:
		other, same := c2.(Coroutine)
		if same && Equals(cons.Reads, other.Reads) && Equals(cons.Yields, other.Yields) {
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
	case Coroutine:
		return Coroutine{ReplaceCons(cons.Yields, old, new), ReplaceCons(cons.Reads, old, new)}
	case BaseType:
	// Don't do anything with base types
	case StructOptions:
		// Don't do anything with struct options
	default:
		panic("Can't replace constraint: " + cons.ConsString())
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

// TODO stop using this. Create a more advanced container for subs that allows holding unhashable types
func (u *Unifier) ReplaceAllSubs(old Constrainable, new Constrainable) {
	for con1, con2 := range u.subs {
		u.subs[con1] = ReplaceCons(con2, old, new)
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

func (u *Unifier) resolveStructOpt(old Constrainable, newOpt StructOptions) (Constrainable, error) {
	var newItem Constrainable = newOpt
	if len(newOpt.Types) == 0 {
		return nil, fmt.Errorf("no type satisfies struct constraints")
	}
	if len(newOpt.Types) == 1 {
		// There is only one struct option, add base type constraints
		structType := newOpt.Types[0].(types.StructType)

		var structDef *ast.StructDef
		for _, s := range u.prog.Structs {
			if s.Type.Name == structType.Name {
				structDef = s
				break
			}
		}

		for depVar, dep := range newOpt.Dependants {
			fmt.Println(structType.Name, structDef)
			if structDef.HasMethod(dep) {
				// We need to treat struct methods differently than members
				method := structDef.Method(dep)
				u.cons = append(u.cons, Constraint{depVar, remFirstArg(u.funcs[method.TargetName])})
				continue
			} else {
				// Member
				u.cons = append(u.cons, Constraint{depVar, BaseType{structDef.MemberType(dep)}})
			}
		}
		newItem = BaseType{structType}
	}

	u.ReplaceAllCons(old, newItem)
	return newItem, nil
}

func (u *Unifier) Unify(currCons Constraint) error {
	// Skip same sided vars
	if Equals(currCons.Left, currCons.Right) {
		DebugInfer("Skipping equal constraint", currCons.Left.ConsString(), "=", currCons.Right.ConsString())
		return nil
	}

	leftBase, isLeftBase := currCons.Left.(BaseType)
	rightBase, isRightBase := currCons.Right.(BaseType)
	rightStructOpt, isRightStructOpt := currCons.Right.(StructOptions)
	leftStructOpt, isLeftStructOpt := currCons.Left.(StructOptions)

	// Unify base types
	if isLeftBase && isRightBase && leftBase != rightBase {
		return fmt.Errorf("type inference failed, base types not equal")
	}
	if isLeftBase {
		return u.Unify(Constraint{currCons.Right, currCons.Left})
	}
	if isRightBase && isLeftStructOpt {
		newOpt := StructOptions{}
		newOpt.Dependants = leftStructOpt.Dependants
		newOpt.Types = []types.Type{rightBase.Type}
		_, err := u.resolveStructOpt(leftStructOpt, newOpt)
		if err != nil {
			return err
		}
		// I am not sure if this replacement needs to recorded in subs, but it seems like it doesn't
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

		var newCon Constrainable
		var old Constrainable
		if rightContainer.Type != nullT {
			u.ReplaceAllCons(leftContainer, rightContainer)
			newCon = rightContainer
			old = leftContainer
		} else if leftContainer.Type != nullT {
			u.ReplaceAllCons(rightContainer, leftContainer)
			newCon = leftContainer
			old = rightContainer
		}

		// Only do a replacement if one of them had a more concrete type
		if old != nil && newCon != nil {
			u.subs[old] = newCon
		}

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

	rightCoroutine, isRightCoroutine := currCons.Right.(Coroutine)
	leftCoroutine, isLeftCoroutine := currCons.Left.(Coroutine)
	if rightIsVar && isLeftCoroutine {
		return u.Unify(Constraint{rightVar, leftCoroutine})
	}
	if leftIsVar && isRightCoroutine {
		u.subs[leftVar] = rightCoroutine
		u.ReplaceAllCons(leftVar, rightCoroutine)
		return nil
	}
	if isLeftCoroutine && isRightCoroutine {
		u.subs[leftCoroutine.Yields] = rightCoroutine.Yields
		u.subs[leftCoroutine.Reads] = rightCoroutine.Reads
		u.ReplaceAllCons(leftCoroutine, rightCoroutine)
		return nil
	}

	// Unify struct options
	if rightIsVar && isLeftStructOpt {
		return u.Unify(Constraint{rightVar, leftStructOpt})
	}
	if leftIsVar && isRightStructOpt {
		u.subs[leftVar] = rightStructOpt
		newItem, err := u.resolveStructOpt(leftVar, rightStructOpt)
		if err != nil {
			return err
		}
		u.subs[leftVar] = newItem
		return nil
	}
	if isLeftStructOpt && isRightStructOpt {
		intersectStructs := make([]types.Type, 0)
		for _, lType := range leftStructOpt.Types {
			for _, rType := range rightStructOpt.Types {
				if types.Equals(lType, rType) {
					intersectStructs = append(intersectStructs, lType)
				}
			}
		}

		allDeps := make(map[TypeVar]string)
		for k, v := range leftStructOpt.Dependants {
			allDeps[k] = v
		}
		for k, v := range rightStructOpt.Dependants {
			allDeps[k] = v
		}

		newOpt := StructOptions{intersectStructs, allDeps}
		finalRepl, err := u.resolveStructOpt(leftStructOpt, newOpt)
		u.ReplaceAllSubs(leftStructOpt, finalRepl)
		if err != nil {
			return err
		}

		finalRepl, err = u.resolveStructOpt(rightStructOpt, newOpt)
		u.ReplaceAllSubs(rightStructOpt, finalRepl)

		return err
	}

	return errors.Errorf("unable to unify '%v' and '%v'", reflect.TypeOf(currCons.Left), reflect.TypeOf(currCons.Right))
}
