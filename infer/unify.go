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
			rightType := u.i.Resolve(rightFunc.Ret).(FuncMeta).data.(int)
			leftType := u.i.Resolve(leftFunc.Ret).(FuncMeta).data.(int)
			rightElems := u.i.Resolve(rightFunc.Args[0]).(FuncMeta).data.(map[int]TypeRef)
			leftElems := u.i.Resolve(leftFunc.Args[0]).(FuncMeta).data.(map[int]TypeRef)
			if leftType == PartialTuple && rightType != PartialTuple {
				return u.unify(swap(con))
			}
			if rightType == PartialTuple && leftType == WholeTuple {
				for propName, propValue := range rightElems {
					sourceProp, ok := leftElems[propName]
					if !ok {
						panic("property doesn't belong to tuple")
					}
					u.i.AddCons(propValue, sourceProp)
				}
				u.i.SetRef(con.Right, con.Left)
				return nil
			}
			if leftType == PartialTuple && rightType == PartialTuple {
				for propName, propValue := range rightElems {
					leftPartial, ok := leftElems[propName]
					if !ok {
						continue
					}
					u.i.AddCons(propValue, leftPartial)
				}

				for propName, propValue := range leftElems {
					rightPartial, ok := rightElems[propName]
					if !ok {
						rightElems[propName] = propValue
						continue
					}
					u.i.AddCons(propValue, rightPartial)
				}
				u.i.SetRef(con.Left, con.Right)
				return nil
			}
			if rightType == WholeTuple && leftType == WholeTuple {
				return nil
			}
			panic(fmt.Sprintf("tuple instance case unhandled: %d %d", rightType, leftType))
		}

		if leftFunc.Kind == KindStructInstance && rightFunc.Kind == KindStructInstance {
			rightType := u.i.Resolve(rightFunc.Ret).(FuncMeta).data.(int)
			leftType := u.i.Resolve(leftFunc.Ret).(FuncMeta).data.(int)

			// If there is a partial, it will always be on the right
			if leftType == PartialStruct && rightType != PartialStruct {
				return u.unify(swap(con))
			}
			if (leftType == WholeStruct || leftType == ArrStruct || leftType == StrStruct) && rightType == PartialStruct {
				partialProps := u.i.Resolve(rightFunc.Args[0]).(FuncMeta).data.(map[string]TypeRef)
				wholeProps := u.i.Resolve(leftFunc.Args[0]).(FuncMeta).data.(map[string]TypeRef)
				for propName, propValue := range partialProps {
					wholeProp, ok := wholeProps[propName]
					if !ok {
						panic("property doesn't belong to struct")
					}
					u.i.AddCons(propValue, wholeProp)
				}
				u.i.SetRef(con.Right, con.Left)
				return nil
			}
			if leftType == PartialStruct && rightType == PartialStruct {
				// UNTESTED
				rightPartials := u.i.Resolve(rightFunc.Args[0]).(FuncMeta).data.(map[string]TypeRef)
				leftPartials := u.i.Resolve(leftFunc.Args[0]).(FuncMeta).data.(map[string]TypeRef)
				for propName, propValue := range rightPartials {
					leftPartial, ok := leftPartials[propName]
					if !ok {
						continue
					}
					u.i.AddCons(propValue, leftPartial)
				}

				for propName, propValue := range leftPartials {
					rightPartial, ok := rightPartials[propName]
					if !ok {
						rightPartials[propName] = propValue
						continue
					}
					u.i.AddCons(propValue, rightPartial)
				}
				u.i.SetRef(con.Left, con.Right)
				return nil
			}
			if leftType == ArrStruct && rightType == ArrStruct {
				leftSub := leftFunc.Args[1]
				rightSub := rightFunc.Args[1]
				u.i.AddCons(leftSub, rightSub)
				u.i.SetRef(con.Left, con.Right)
				return nil
			}
			if leftType == StrStruct && rightType == StrStruct {
				return nil
			}
			if leftType == WholeStruct && rightType == WholeStruct {
				return nil
			}
			panic(fmt.Sprintf("struct instance case unhandled: %d %d", leftType, rightType))
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
