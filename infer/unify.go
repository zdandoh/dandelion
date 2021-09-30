package infer

import (
	"dandelion/types"
	"errors"
	"fmt"
)

type Unifier struct {
	i *Inferer
}

func Unify(i *Inferer) {
	u := &Unifier{}
	u.i = i

	for k := 0; k < len(u.i.cons); k++ {
		err := u.unify(u.i.cons[k])
		fmt.Println("--- CONS ---")
		i.printCons()
		if err != nil {
			panic(err)
		}
	}
}

func swap(con *TCons) *TCons {
	return &TCons{con.Right, con.Left}
}

func (u *Unifier) unify(con *TCons) error {
	left := u.i.Resolve(con.Left)
	right := u.i.Resolve(con.Right)

	_, leftIsVar := left.(TypeVar)
	_, rightIsVar := right.(TypeVar)

	if rightIsVar && !leftIsVar {
		// If it exists, the type variable will always be on the left for the rest of the cases
		return u.unify(swap(con))
	}

	// Type var replacements
	if leftIsVar && rightIsVar {
		u.i.SetRef(con.Left, con.Right)
		return nil
	}

	leftFunc, leftIsFunc := left.(TypeFunc)
	rightFunc, rightIsFunc := right.(TypeFunc)

	if leftIsVar && rightIsFunc && rightFunc.Reducible() {
		// The function is reducible, replace the type var if possible
		// This means we will carry the information as many hops as possible
		// It must be this way for property accesses to get carried all the way
		// to the end
		u.i.SetRef(con.Left, con.Right)
		return nil
	}
	if leftIsVar && rightIsFunc {
		u.i.SetRef(con.Left, con.Right)
		return nil
	}

	if leftIsFunc && rightIsFunc {
		// If one function is reducible & the other isn't, the reducible one is on the left
		if rightFunc.Reducible() && !leftFunc.Reducible() {
			return u.unify(swap(con))
		}

		// Rule to equate container subtypes with their concrete type func
		if leftFunc.Kind == KindContainer && rightFunc.Kind == KindArray {
			u.i.AddCons(leftFunc.Ret, rightFunc.Ret)
		}
		if leftFunc.Kind == KindContainer && rightFunc.Kind == KindCoro {
			u.i.AddCons(leftFunc.Ret, rightFunc.Args[0])
		}

		// Unify partial tuples (aka tuple accesses) and normal tuples such that
		// tuple accesses are overwritten by tuples
		if leftFunc.Kind == KindTuple && rightFunc.Kind == KindTuple {
			rightIdx := u.i.Resolve(rightFunc.Ret).(FuncMeta).data.(int)
			leftIdx := u.i.Resolve(leftFunc.Ret).(FuncMeta).data.(int)
			if rightIdx == WholeTuple && leftIdx != WholeTuple {
				return u.unify(swap(con))
			}
			if rightIdx != WholeTuple && leftIdx == WholeTuple {
				u.i.AddCons(rightFunc.Args[0], leftFunc.Args[rightIdx])
				rightFunc.Args = leftFunc.Args[:]
				rightFunc.Ret = leftFunc.Ret
				return nil
			}
		}

		if leftFunc.Kind != rightFunc.Kind && !leftFunc.Reducible() && !rightFunc.Reducible() {
			return fmt.Errorf("incomparable non-reducible type functions with non-like kinds: %s != %s", u.i.String(leftFunc), u.i.String(rightFunc))
		}

		if leftFunc.Kind != rightFunc.Kind {
			// We have two functions of different kinds where at least one is reducible, but haven't defined
			// any specific special rules for this combo. Replace the reducible one.
			u.i.SetRef(con.Left, con.Right)
			return nil
		}

		err := u.unify(&TCons{leftFunc.Ret, rightFunc.Ret})
		if err != nil {
			return err
		}

		if len(rightFunc.Args) != len(leftFunc.Args) {
			return fmt.Errorf("function argument count mismatch: %s != %s", u.i.String(leftFunc), u.i.String(rightFunc))
		}

		for k := range rightFunc.Args {
			err = u.unify(&TCons{leftFunc.Args[k], rightFunc.Args[k]})
			if err != nil {
				return err
			}
		}

		return nil
	}
	leftBase, leftIsBase := left.(TypeBase)
	rightBase, rightIsBase := right.(TypeBase)
	if leftIsVar && rightIsBase {
		u.i.SetRef(con.Left, con.Right)
		return nil
	}

	// After this point, if one is a base type & one if a function, the base type will be on the left
	if leftIsFunc && rightIsBase {
		return u.unify(swap(con))
	}

	if leftIsBase && rightIsFunc && rightFunc.Reducible() {
		u.i.SetRef(con.Right, con.Left)
		return nil
	}
	if leftIsBase && rightIsFunc {
		return fmt.Errorf("can't assign base type to function: %s | %s", u.i.String(leftBase), u.i.String(rightFunc))
	}

	if leftIsBase && rightIsBase && !types.Equals(leftBase.Type, rightBase.Type) {
		return errors.New("assignment of non-equal base types")
	}

	return nil
}
