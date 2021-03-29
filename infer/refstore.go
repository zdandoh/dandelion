package infer

import (
	"dandelion/ast"
	"dandelion/types"
	"fmt"
	"reflect"
	"strings"
)

func (i *Inferer) NewVar() TypeRef {
	i.currVar++
	i.varLibrary[i.currVar] = append(i.varLibrary[i.currVar], TypeRef(len(i.varList)))
	i.varList = append(i.varList, i.currVar)

	t := TypeRef(len(i.varList) - 1)
	return t
}

func (i *Inferer) TypeRef(astNode ast.Node) TypeRef {
	hash := ast.HashNode(astNode)
	ref, ok := i.refs[hash]
	if ok {
		return ref
	}

	v := i.NewVar()
	i.refs[hash] = v
	return v
}

func (i *Inferer) TypeRefs(astNodes []ast.Node) []TypeRef {
	refs := make([]TypeRef, len(astNodes))
	for k, node := range astNodes {
		refs[k] = i.TypeRef(node)
	}

	return refs
}

func (i *Inferer) BaseRef(base TypeBase) TypeRef {
	ref, ok := i.varLibrary[base]
	if ok {
		return ref[0]
	}

	i.currVar++
	i.varLibrary[base] = append(i.varLibrary[base], TypeRef(len(i.varList)))
	i.varList = append(i.varList, base)

	return TypeRef(len(i.varList) - 1)
}

func (i *Inferer) FuncMeta(data interface{}) TypeRef {
	i.currVar++
	i.currMeta++
	meta := FuncMeta{i.currMeta, data}
	i.varLibrary[meta.Key()] = append(i.varLibrary[meta.Key()], TypeRef(len(i.varList)))
	i.varList = append(i.varList, meta)

	return TypeRef(len(i.varList) - 1)
}

func (i *Inferer) FuncRef(kind FuncKind, ret TypeRef, args... TypeRef) TypeRef {
	i.currFuncID++
	typeFun := TypeFunc{args, ret, kind, i.currFuncID}
	hash := typeFun.Key()

	ref, ok := i.varLibrary[hash]
	if ok {
		return ref[0]
	}

	i.currVar++
	i.varLibrary[hash] = append(i.varLibrary[hash], TypeRef(len(i.varList)))
	i.varList = append(i.varList, typeFun)

	return TypeRef(len(i.varList) - 1)
}

func (i *Inferer) TupleRef(typs... TypeRef) TypeRef {
	// Use int type as a dummy return type since we can't return nothing
	tupRef := i.FuncRef(KindTuple, i.BaseRef(TypeBase{types.IntType{}}), typs...)
	return tupRef
}

func (i *Inferer) ArrRef(subtype TypeRef) TypeRef {
	arrRef := i.FuncRef(KindArray, subtype, subtype)
	return arrRef
}

func (i *Inferer) StructRef(def *ast.StructDef) TypeRef {
	structFun := i.FuncRef(KindStructInstance, i.FuncMeta(def))
	return structFun
}

func (i *Inferer) CoroRef(yields TypeRef, gets TypeRef) TypeRef {
	coroFun := i.FuncRef(KindCoro, gets, yields)
	return coroFun
}

func (i *Inferer) ContainerRef(subtype TypeRef) TypeRef {
	containerFun := i.FuncRef(KindContainer, subtype)
	return containerFun
}

func (i *Inferer) Contains(fun TypeFunc, ref TypeRef) bool {
	resRef := i.Resolve(ref)

	for _, arg := range fun.Args {
		res := i.Resolve(arg)
		argFun, isFun := res.(TypeFunc)
		if isFun {
			if i.Contains(argFun, ref) {
				return true
			}
			continue
		}

		if res == resRef {
			return true
		}
	}

	resRet := i.Resolve(fun.Ret)
	retFun, isFun := resRet.(TypeFunc)
	if isFun {
		if i.Contains(retFun, ref) {
			return true
		} else {
			return false
		}
	}

	return resRet == resRef
}

func (i *Inferer) Resolve(ref TypeRef) StorableType {
	return i.varList[ref]
}

func (i *Inferer) SetRef(old TypeRef, new TypeRef) {
	oldVal := i.Resolve(old)
	newVal := i.Resolve(new)
	if oldVal.Key() == newVal.Key() {
		// Don't do anything if they're the same, or we'll break the varLibrary
		return
	}

	oldFun, isOldFun := oldVal.(TypeFunc)
	newFun, isNewFun := newVal.(TypeFunc)

	if isOldFun && i.Contains(oldFun, new) {
		return
	}
	if isNewFun && i.Contains(newFun, old) {
		return
	}
	if isOldFun && isNewFun && !oldFun.Reducible() {
		panic("cannot replace irreducible function")
	}
	if isOldFun && !isNewFun && !oldFun.Reducible() {
		panic("trying to simplify non-reducible func kind")
	}

	locs := i.varLibrary[oldVal.Key()]
	for _, loc := range locs {
		i.varList[loc] = newVal
	}

	i.varLibrary[newVal.Key()] = append(i.varLibrary[newVal.Key()], locs...)
	delete(i.varLibrary, oldVal.Key())
}

func (i *Inferer) String(t TypeExpr) string {
	switch ty := t.(type) {
	case TypeVar:
		return ty.String()
	case TypeRef:
		return i.String(i.Resolve(ty).(TypeExpr))
	case TypeBase:
		return ty.String()
	case TypeFunc:
		args := make([]string, len(ty.Args))
		for k, arg := range ty.Args {
			args[k] = i.String(arg)
		}
		ret := i.String(ty.Ret)

		return fmt.Sprintf("%s(%s) -> %s", ty.Kind, strings.Join(args, ", "), ret)
	case FuncMeta:
		return fmt.Sprintf("meta(%v)", ty.data)
	default:
		panic("can't stringify unknown TypeExpr: " + reflect.TypeOf(t).String())
	}
}