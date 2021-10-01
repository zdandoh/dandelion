package infer

import (
	"dandelion/ast"
	"dandelion/transform"
	"dandelion/types"
	"errors"
	"fmt"
)

type Resolver struct {
	i *Inferer
	ResolvedTypes map[ast.NodeHash]types.Type
}

func NewResolver(i *Inferer) *Resolver {
	r := &Resolver{}
	r.ResolvedTypes = make(map[ast.NodeHash]types.Type)
	r.i = i

	return r
}

func (r *Resolver) resolve(nodeRef TypeRef) (types.Type, error) {
	var retType types.Type
	var err error
	res := r.i.Resolve(nodeRef)

	switch ty := res.(type) {
	case TypeFunc:
		switch ty.Kind {
		case KindFunc:
			funType := types.FuncType{}
			for _, arg := range ty.Args {
				resArg, err := r.resolve(arg)
				if err != nil {
					return nil, fmt.Errorf("func arg: %w", err)
				}
				funType.ArgTypes = append(funType.ArgTypes, resArg)
			}

			funType.RetType, err = r.resolve(ty.Ret)
			if err != nil {
				return nil, fmt.Errorf("func ret: %w", err)
			}

			retType = funType
		case KindArray:
			arrType := types.ArrayType{}
			arrType.Subtype, err = r.resolve(ty.Args[0])
			if err != nil {
				return nil, fmt.Errorf("array: %w", err)
			}

			retType = arrType
		case KindTuple:
			tupType := types.TupleType{}
			wholeTup := r.i.Resolve(ty.Ret).(FuncMeta).data.(int)
			if wholeTup != WholeTuple {
				panic("partial tuple left after inference")
			}

			tupElems := r.i.Resolve(ty.Args[0]).(FuncMeta).data.(map[int]TypeRef)
			for k := 0; k < len(tupElems); k++ {
				tupElem := tupElems[k]
				resElem, err := r.resolve(tupElem)
				if err != nil {
					return nil, fmt.Errorf("tuple: %w", err)
				}
				tupType.Types = append(tupType.Types, resElem)
			}

			retType = tupType
		case KindStructInstance:
			structType := r.i.Resolve(ty.Ret).(FuncMeta).data.(int)
			if structType == PartialStruct {
				panic("partial struct left after inference")
			}
			if structType == WholeStruct {
				structType := r.i.Resolve(ty.Args[1]).(FuncMeta).data.(*ast.StructDef).Type
				retType = structType
			} else if structType == ArrStruct {
				arrSubtype, err := r.resolve(ty.Args[1])
				if err != nil {
					return nil, fmt.Errorf("unknown array subtype: %s", r.i.String(nodeRef))
				}
				retType = types.ArrayType{arrSubtype}
			} else {
				panic("unknown partial struct type")
			}
		case KindCoro:
			yields, err := r.resolve(ty.Args[0])
			if err != nil {
				return nil, fmt.Errorf("coro yield: %w", err)
			}
			gets, err := r.resolve(ty.Ret)
			if err != nil {
				// Can't detect the send type, just use int
				gets = types.IntType{}
			}
			coroType := types.CoroutineType{yields, gets}
			retType = coroType
		default:
			panic(fmt.Sprintf("unknown function kind during type resolution: %s | %s", ty.Kind, r.i.String(ty)))
		}
	case FuncMeta:
		return nil, errors.New("incorrectly trying to resolve func meta")
	default:
		base, ok := res.(TypeBase)
		if ok {
			retType = base.Type
		} else {
			return nil, fmt.Errorf("couldn't resolve type: %s", r.i.String(nodeRef))
		}
	}

	return retType, nil
}

func (r *Resolver) WalkNode(astNode ast.Node) ast.Node {
	hash := ast.HashNode(astNode)

	var nodeType types.Type
	var err error
	if !ast.IsVoid(astNode) {
		nodeRef := r.i.TypeRef(astNode)
		nodeType, err = r.resolve(nodeRef)
		if err != nil {
			panic(fmt.Sprintf("error resolving type during inference: %s | %s | %s", err, astNode, r.i.String(nodeRef)))
		}
	} else {
		nodeType = types.VoidType{}
	}

	_, isBegin := astNode.(*ast.BeginExp)
	_, isFunApp := astNode.(*ast.FunApp)
	if types.Equals(nodeType, types.VoidType{}) && !ast.Statement(astNode) && !isBegin && !isFunApp && !transform.IsCloArg(astNode) {
		panic("invalid void expression: " + astNode.String())
	}
	r.ResolvedTypes[hash] = nodeType
	fmt.Println(astNode, "|", nodeType.TypeString())

	return nil
}

func (r *Resolver) WalkBlock(block *ast.Block) *ast.Block {
	return nil
}

func Resolve(prog *ast.Program, i *Inferer) map[ast.NodeHash]types.Type {
	r := NewResolver(i)

	for _, fun := range prog.Funcs {
		ast.WalkAst(fun, r)
	}

	return r.ResolvedTypes
}