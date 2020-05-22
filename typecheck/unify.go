package typecheck

import (
	"dandelion/ast"
	"dandelion/errs"
	"dandelion/types"
	"reflect"
)

type Subs map[ConsHash]*SubPair
type SubPair struct {
	Old Constrainable
	New Constrainable
}

func (s Subs) Set(old Constrainable, new Constrainable) {
	hash := HashCons(old)
	s[hash] = &SubPair{old, new}
}

func (s Subs) Get(cons Constrainable) (Constrainable, bool) {
	pair, ok := s[HashCons(cons)]
	if !ok {
		return nil, ok
	}
	return pair.New, ok
}

func (s Subs) GetPair(cons Constrainable) (Constrainable, Constrainable) {
	pair := s[HashCons(cons)]
	return pair.Old, pair.New
}

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
		oldCont, isOldCont := old.(Container)
		if isOldCont && oldCont.ID == cons.ID {
			return new
		}
		return Container{cons.Type, ReplaceCons(cons.Subtype, old, new), cons.Index, cons.ID}
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

		u.cons[i] = Constraint{newLeft, newRight, con.Source}
	}
}

// TODO stop using this. Create a more advanced container for subs that allows holding unhashable types
func (u *Unifier) ReplaceAllSubs(old Constrainable, new Constrainable) {
	for _, pair := range u.subs {
		u.subs.Set(pair.Old, ReplaceCons(pair.New, old, new))
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

func (u *Unifier) resolveStructOpt(currCons Constraint, old Constrainable, newOpt StructOptions) (Constrainable, error) {
	var newItem Constrainable = newOpt
	if len(newOpt.Types) == 0 {
		errs.Error(errs.ErrorType, currCons.Source, "no type satisfies struct constraints")
		return nil, errs.ErrorType
	}
	if len(newOpt.Types) == 1 {
		// There is only one struct option, add base type constraints
		structType := newOpt.Types[0].(types.StructType)

		var structDef *ast.StructDef
		for i := 0; i < u.prog.StructCount(); i++ {
			s := u.prog.StructNo(i)
			if s.Type.Name == structType.Name {
				structDef = s
				break
			}
		}

		for depVar, dep := range newOpt.Dependants {
			DebugInfer(structType.Name, structDef)
			if structDef.HasMethod(dep) {
				// We need to treat struct methods differently than members
				method := structDef.Method(dep)
				u.cons = append(u.cons, Constraint{depVar, remFirstArg(u.funcs[method.TargetName]), currCons.Source})
				continue
			} else {
				// Member
				u.cons = append(u.cons, Constraint{depVar, BaseType{structDef.MemberType(dep)}, currCons.Source})
			}
		}
		newItem = BaseType{structType}
	}

	u.ReplaceAllCons(old, newItem)
	return newItem, nil
}

func (u *Unifier) resolvePrimitiveMethod(sourceCons Constraint, structOpt StructOptions, baseCons Constrainable) {
	for depVar, depName := range structOpt.Dependants {
		switch cons := baseCons.(type) {
		case Container:
			// Need to setup type inference for methods of containers
			switch depName {
			case "push":
				pushType := Fun{
					Args: []Constrainable{cons.Subtype},
					Ret:  BaseType{types.VoidType{}},
				}
				u.cons = append(u.cons, Constraint{depVar, pushType, sourceCons.Source})
			case "pop":
				popType := Fun{
					Args: []Constrainable{},
					Ret:  cons.Subtype,
				}
				u.cons = append(u.cons, Constraint{depVar, popType, sourceCons.Source})
			default:
				errs.Error(errs.ErrorValue, sourceCons.Source, "container doesn't have method: "+depName)
				errs.CheckExit()
			}
		}
	}

	u.ReplaceAllCons(structOpt, baseCons)
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
		errs.Error(errs.ErrorType, currCons.Source, "base types not equal %s != %s", leftBase.TypeString(), rightBase.TypeString())
		return errs.ErrorType
	}
	if isLeftBase {
		return u.Swap(currCons)
	}
	if isRightBase && isLeftStructOpt {
		newOpt := StructOptions{}
		newOpt.Dependants = leftStructOpt.Dependants
		newOpt.Types = []types.Type{rightBase.Type}
		_, err := u.resolveStructOpt(currCons, leftStructOpt, newOpt)
		if err != nil {
			return err
		}
		// I am not sure if this replacement needs to recorded in subs, but it seems like it doesn't
		return nil
	}
	if isRightBase {
		u.subs.Set(currCons.Left, rightBase)
		u.ReplaceAllCons(currCons.Left, rightBase)
		return nil
	}

	// Unify type vars
	rightVar, rightIsVar := currCons.Right.(TypeVar)
	leftVar, leftIsVar := currCons.Left.(TypeVar)
	if rightIsVar && leftIsVar {
		u.subs.Set(currCons.Left, rightVar)
		u.ReplaceAllCons(leftVar, rightVar)
		return nil
	}
	if rightIsVar && !leftIsVar {
		return u.Swap(currCons)
	}

	rightFun, rightIsFun := currCons.Right.(Fun)
	leftFun, leftIsFun := currCons.Left.(Fun)
	if rightIsFun && leftIsFun {
		if len(rightFun.Args) != len(leftFun.Args) {
			errs.Error(errs.ErrorValue, currCons.Source, "function application doesn't have correct argument count")
			return errs.ErrorValue
		}
		for k, arg := range leftFun.Args {
			u.cons = append(u.cons, Constraint{arg, rightFun.Args[k], currCons.Source})
		}
		u.cons = append(u.cons, Constraint{leftFun.Ret, rightFun.Ret, currCons.Source})
		return nil
	}
	if rightIsFun && !leftIsFun {
		u.subs.Set(currCons.Left, rightFun)
		u.ReplaceAllCons(currCons.Left, rightFun)
		return nil
	}

	// Unify containers
	rightContainer, isRightContainer := currCons.Right.(Container)
	leftContainer, isLeftContainer := currCons.Left.(Container)
	if leftIsVar && isRightContainer {
		u.subs.Set(currCons.Left, rightContainer)
		u.ReplaceAllCons(leftVar, rightContainer)
		return nil
	}
	if isLeftContainer && isRightContainer {
		nullT := types.VoidType{}

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
			u.subs.Set(old, newCon)
		}

		u.cons = append(u.cons, Constraint{leftContainer.Subtype, rightContainer.Subtype, currCons.Source})
		return nil
	}

	// Unify tuples
	rightTuple, isRightTuple := currCons.Right.(Tup)
	leftTuple, isLeftTuple := currCons.Left.(Tup)
	if leftIsVar && isRightTuple {
		u.subs.Set(currCons.Left, rightTuple)
		u.ReplaceAllCons(leftVar, rightTuple)
		return nil
	}
	if isLeftTuple && isRightContainer {
		return u.Swap(currCons)
	}
	if isRightTuple && isLeftContainer {
		if leftContainer.Index < 0 || leftContainer.Index >= len(rightTuple.Subtypes) {
			errs.Error(errs.ErrorValue, currCons.Source, "illegal index for tuple: %d", leftContainer.Index)
			return errs.ErrorValue
		}

		u.cons = append(u.cons, Constraint{leftContainer.Subtype, rightTuple.Subtypes[leftContainer.Index], currCons.Source})
		u.subs.Set(leftContainer, rightTuple)
		u.ReplaceAllCons(leftContainer, rightTuple)

		return nil
	}
	if isLeftTuple && isRightTuple {
		if len(leftTuple.Subtypes) != len(rightTuple.Subtypes) {
			errs.Error(errs.ErrorType, currCons.Source, "tuples have different subtype counts")
			return errs.ErrorType
		}
		for k, sub := range leftTuple.Subtypes {
			u.cons = append(u.cons, Constraint{sub, rightTuple.Subtypes[k], currCons.Source})
		}
		return nil
	}

	rightCoroutine, isRightCoroutine := currCons.Right.(Coroutine)
	leftCoroutine, isLeftCoroutine := currCons.Left.(Coroutine)
	if rightIsVar && isLeftCoroutine {
		return u.Swap(currCons)
	}
	if leftIsVar && isRightCoroutine {
		u.subs.Set(leftVar, rightCoroutine)
		u.ReplaceAllCons(leftVar, rightCoroutine)
		return nil
	}
	if isLeftCoroutine && isRightCoroutine {
		u.cons = append(u.cons, Constraint{leftCoroutine.Yields, rightCoroutine.Yields, currCons.Source})
		u.cons = append(u.cons, Constraint{leftCoroutine.Reads, rightCoroutine.Reads, currCons.Source})
		return nil
	}
	if isLeftCoroutine && isRightContainer {
		return u.Swap(currCons)
	}
	if isLeftContainer && isRightCoroutine {
		u.ReplaceAllCons(leftContainer, rightCoroutine)
		u.ReplaceAllSubs(leftContainer, rightCoroutine)
		u.cons = append(u.cons, Constraint{leftContainer.Subtype, rightCoroutine.Yields, currCons.Source})
		return nil
	}

	// Unify struct options
	if (rightIsVar || isRightContainer) && isLeftStructOpt {
		return u.Swap(currCons)
	}
	if leftIsVar && isRightStructOpt {
		u.subs.Set(leftVar, rightStructOpt)
		newItem, err := u.resolveStructOpt(currCons, leftVar, rightStructOpt)
		if err != nil {
			return err
		}
		u.subs.Set(leftVar, newItem)
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
		finalRepl, err := u.resolveStructOpt(currCons, leftStructOpt, newOpt)
		u.ReplaceAllSubs(leftStructOpt, finalRepl)
		if err != nil {
			return err
		}

		finalRepl, err = u.resolveStructOpt(currCons, rightStructOpt, newOpt)
		u.ReplaceAllSubs(rightStructOpt, finalRepl)

		return err
	}
	if isRightStructOpt && isLeftContainer {
		u.resolvePrimitiveMethod(currCons, rightStructOpt, leftContainer)
		return nil
	}

	errs.Error(errs.ErrorType, currCons.Source, "unable to infer type of expression - incompatible types '%s' and '%s'", reflect.TypeOf(currCons.Left).Name(), reflect.TypeOf(currCons.Right).Name())
	return errs.ErrorType
}

func (u *Unifier) Swap(cons Constraint) error {
	return u.Unify(Constraint{cons.Right, cons.Left, cons.Source})
}
